package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	"github.com/tal-tech/go-zero/core/logx"
	"sensors/common"
	"sensors/initialize"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 页面 API
type CronController struct {
	BaseController
}

var (
	// worker池，限制并发goroutine数量，防止goroutine泄漏
	workerPoolSize  = 100
	workerSemaphore = make(chan struct{}, workerPoolSize)
)

func CronInsertCk() {
	pool := initialize.PoolDef
	key, err := pool.Hkeys(common.CacheEventsName)
	if err != nil {
		logx.Errorf("CacheEventsName err:%v", err.Error())
		return
	}

	// 使用worker池控制并发，防止goroutine泄漏
	for _, s := range key {
		// 获取信号量，限制并发数
		workerSemaphore <- struct{}{}
		go func(name string) {
			defer func() {
				// 释放信号量
				<-workerSemaphore
				// 捕获panic，防止单个worker崩溃影响其他worker
				if r := recover(); r != nil {
					logx.Errorf("worker %s panic recovered: %v", name, r)
				}
			}()
			worker(name)
		}(s)
	}
}

func worker(name string) {
	// 这里放置每个协程要执行的任务
	pool := initialize.PoolDef
	key := fmt.Sprintf("%v%v", common.CacheEventsData, name)
	data, err := pool.SpopN(key, 500)
	if err != nil {
		logx.Errorf("insert SpopN key:%v", fmt.Sprintf("%v%v", common.CacheEventsData, name))
		return
	}
	if len(data) == 0 {
		return
	}
	ckpool := initialize.Clickhouse
	if err = InsertData(ckpool, "sensors.event", data); err != nil {
		logx.Errorf("InsertData failed for key=%s: %v", key, err)
		_, err = pool.Sadd(key, data)
		if err != nil {
			logx.Errorf("restore redis event cache failed for key=%s: %v", key, err)
		}
	}
}

func CronInsertUsersCk() {
	table := "sensors.user"
	pool := initialize.PoolDef
	key, err := pool.SpopN(common.CacheUsersLog, 100)
	if err != nil {
		logx.Errorf("CacheUsersLog err:%v", err.Error())
		return
	}
	ckpool := initialize.Clickhouse
	for _, v := range key {
		tmp, _ := pool.Hgetall(fmt.Sprintf("%v%v", common.CacheUserInfo, v))
		tmpJson, _ := json.Marshal(tmp)
		DeleteData(ckpool, table, v)
		if err = InsertData(ckpool, table, []string{string(tmpJson)}); err != nil {
			logx.Errorf("InsertData failed for user=%s: %v", v, err)
			_, err = pool.Sadd(common.CacheUsersLog, key)
			if err != nil {
				logx.Errorf("restore redis user cache failed: %v", err)
			}
		}

	}

}

func DeleteData(ckpool *sql.DB, tableName, distinctId string) error {
	// 验证表名白名单，防止SQL注入
	validTables := map[string]bool{
		"sensors.event": true,
		"sensors.user":  true,
	}
	if !validTables[tableName] {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	// 验证distinctId格式（只允许字母数字下划线短横线，最大128字符）
	if len(distinctId) == 0 || len(distinctId) > 128 {
		return fmt.Errorf("invalid distinct_id length: %d", len(distinctId))
	}
	for _, c := range distinctId {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '@' || c == '.') {
			return fmt.Errorf("invalid distinct_id format: contains illegal character")
		}
	}

	// 使用参数化查询防止SQL注入
	// ClickHouse的ALTER TABLE DELETE语法需要特殊处理
	// 使用占位符虽然ClickHouse ALTER不完全支持，但我们已经验证了输入
	deleteQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE distinct_id = ?", tableName)
	result, err := ckpool.ExecContext(context.Background(), deleteQuery, distinctId)
	if err != nil {
		logx.Errorf("DeleteData failed for table=%s, distinct_id=%s, err=%v", tableName, distinctId, err)
		return err
	}
	logx.Infof("DeleteData success for table=%s, distinct_id=%s, result=%v", tableName, distinctId, result)
	return nil
}

func InsertData(ckpool *sql.DB, tableName string, jsonData []string) (retErr error) {
	//return errors.New("2222")
	tx, err := ckpool.Begin()
	if err != nil {
		logx.Errorf("Begin err:%v", err.Error())
		return err
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = tx.Rollback()
			retErr = fmt.Errorf("insert panic: %v", recovered)
		} else if err := tx.Commit(); err != nil {
			retErr = fmt.Errorf("commit failed: %w", err)
			logx.Infof("Commit err:%v", err.Error())
		}
	}()
	// 获取所有可能出现的列名集合
	allColumns := make(map[string]bool)
	for _, jsonString := range jsonData {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
			logx.Errorf("failed to unmarshal JSON string: %v", err)
			continue
		}
		record, err := common.ValuesToStrings(data)
		if err != nil {
			logx.Errorf("ValuesToStrings err:%v", err.Error())
		}
		for col := range record {
			allColumns[col] = true
		}
	}
	// 从列名集合构建实际的列名列表
	sortedColumns := make([]string, 0, len(allColumns))
	for col := range allColumns {
		sortedColumns = append(sortedColumns, col)
		switch tableName {
		case "sensors.event":
			common.CachedTableSchema.Mutex.Lock()
			if !common.CachedTableSchema.Columns[col] {
				addColumnToClickHouse(col)
				common.CachedTableSchema.Columns[col] = true
			}
			common.CachedTableSchema.Mutex.Unlock()
		case "sensors.user":
			common.CachedUsersSchema.Mutex.Lock()
			if !common.CachedUsersSchema.Columns[col] {
				addUsersColumnToClickHouse(col)
				common.CachedUsersSchema.Columns[col] = true
			}
			common.CachedUsersSchema.Mutex.Unlock()
		}

	}
	sortedColumns = append(sortedColumns, "ds")
	sort.Strings(sortedColumns)
	// 构建通用的INSERT SQL模板
	placeholders := questionMarks(len(sortedColumns))
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(sortedColumns, ","), placeholders)
	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		fmt.Errorf(err.Error())
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close() // 确保stmt在函数结束时关闭
	serverLoc := common.ServerLocation()
	for _, jsonString := range jsonData {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
			logx.Errorf("failed to unmarshal JSON string: %v", err)
			continue
		}
		records, err := common.ValuesToStrings(data)
		if err != nil {
			logx.Errorf("ValuesToStrings err:%v", err.Error())
		}
		record := make([]interface{}, len(sortedColumns)) // 初始化值切片
		for i, col := range sortedColumns {
			record[i] = clickhouseValue(tableName, col, records, serverLoc)
		}
		_, err = stmt.Exec(record...)
		if err != nil {
			stmt.Close() // 在出错的情况下也要关闭stmt
			return fmt.Errorf("failed to execute insert: %v", err)
		}
	}

	return nil

}

// 辅助函数：生成 (?, ?, ...) 形式的占位符
func questionMarks(n int) string {
	qmarks := make([]string, n)
	for i := range qmarks {
		qmarks[i] = "?"
	}
	return strings.Join(qmarks, ", ")
}

func clickhouseValue(tableName, col string, records map[string]string, loc *time.Location) interface{} {
	switch col {
	case "ds":
		t := eventTime(records, loc)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	case "time":
		if tableName == "sensors.event" {
			return eventTime(records, loc)
		}
	case "first_visit_time", "last_visit_time":
		if parsed, ok := parseMilliseconds(records[col], loc); ok {
			return parsed
		}
	case "_flush_time":
		return parseUInt64(records[col])
	case "screen_width", "screen_height":
		return parseUInt16(records[col])
	case "timezone_offset":
		return parseInt16(records[col])
	case "wifi", "is_first_day":
		return parseBool(records[col])
	}
	return records[col]
}

func eventTime(records map[string]string, loc *time.Location) time.Time {
	if parsed, ok := parseMilliseconds(records["time"], loc); ok {
		return parsed
	}
	return time.Now().In(loc)
}

func parseMilliseconds(raw string, loc *time.Location) (time.Time, bool) {
	ts, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || ts <= 0 {
		return time.Time{}, false
	}
	if ts < 1e11 {
		ts *= 1000
	}
	return time.Unix(ts/1000, (ts%1000)*int64(time.Millisecond)).In(loc), true
}

func parseUInt64(raw string) uint64 {
	v, _ := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	return v
}

func parseUInt16(raw string) uint16 {
	v, _ := strconv.ParseUint(strings.TrimSpace(raw), 10, 16)
	return uint16(v)
}

func parseInt16(raw string) int16 {
	v, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 16)
	return int16(v)
}

func parseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
