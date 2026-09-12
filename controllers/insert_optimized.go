package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sensors/common"
	"sensors/initialize"
	"sort"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/core/logx"
)

// 固定字段定义（不需要动态添加）
var (
	// event表的固定字段
	fixedEventColumns = map[string]bool{
		"time":           true,
		"ds":             true,
		"event":          true,
		"distinct_id":    true,
		"$os":            true,
		"$os_version":    true,
		"$model":         true,
		"$brand":         true,
		"$screen_width":  true,
		"$screen_height": true,
		"$wifi":          true,
		"$network_type":  true,
		"$app_version":   true,
		"$app_name":      true,
		"$lib":           true,
		"$lib_version":   true,
		"$user_id":       true,
		"$is_first_day":  true,
	}

	// user表的固定字段
	fixedUserColumns = map[string]bool{
		"distinct_id":      true,
		"ds":               true,
		"user_id":          true,
		"first_visit_time": true,
		"last_visit_time":  true,
		"$name":            true,
		"$email":           true,
		"$phone":           true,
		"$avatar":          true,
	}
)

// InsertDataOptimized 优化版的数据插入函数
func InsertDataOptimized(ckpool *sql.DB, tableName string, jsonData []string) (retErr error) {
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

	// 解析所有JSON数据
	var allRecords []map[string]interface{}
	allColumns := make(map[string][]interface{})
	for _, jsonString := range jsonData {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
			logx.Errorf("failed to unmarshal JSON string: %v", err)
			continue
		}
		allRecords = append(allRecords, data)

		for col, val := range data {
			allColumns[col] = append(allColumns[col], val)
		}
	}

	if len(allRecords) == 0 {
		return fmt.Errorf("no valid records to insert")
	}

	// 获取固定字段列表
	var fixedColumns map[string]bool
	switch tableName {
	case "sensors.event":
		fixedColumns = fixedEventColumns
	case "sensors.user":
		fixedColumns = fixedUserColumns
	default:
		return fmt.Errorf("unknown table: %s", tableName)
	}

	// 处理动态字段
	for col, samples := range allColumns {
		if fixedColumns[col] {
			continue
		}

		detectedType := common.DetectClickHouseTypeFromSamples(samples)
		logx.Infof("Dynamic column detected: %s -> %s (from %d samples)", col, detectedType, len(samples))

		switch tableName {
		case "sensors.event":
			common.CachedTableSchema.Mutex.Lock()
			if !common.CachedTableSchema.Columns[col] {
				if err := addColumnToClickHouseTyped(col, detectedType, "sensors.event"); err != nil {
					logx.Errorf("failed to add column %s: %v", col, err)
				} else {
					common.CachedTableSchema.Columns[col] = true
				}
			}
			common.CachedTableSchema.Mutex.Unlock()
		case "sensors.user":
			common.CachedUsersSchema.Mutex.Lock()
			if !common.CachedUsersSchema.Columns[col] {
				if err := addColumnToClickHouseTyped(col, detectedType, "sensors.user"); err != nil {
					logx.Errorf("failed to add user column %s: %v", col, err)
				} else {
					common.CachedUsersSchema.Columns[col] = true
				}
			}
			common.CachedUsersSchema.Mutex.Unlock()
		}
	}

	// 构建列名列表
	columnSet := make(map[string]bool)
	for col := range allColumns {
		columnSet[col] = true
	}
	columnSet["ds"] = true

	sortedColumns := make([]string, 0, len(columnSet))
	for col := range columnSet {
		sortedColumns = append(sortedColumns, col)
	}
	sort.Strings(sortedColumns)

	// 准备INSERT语句
	placeholders := questionMarks(len(sortedColumns))
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(sortedColumns, ","), placeholders)

	stmt, err := tx.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	serverLoc := common.ServerLocation()

	// 插入每条记录
	for _, record := range allRecords {
		values := make([]interface{}, len(sortedColumns))
		for i, col := range sortedColumns {
			if col == "ds" {
				safeTs := safeEventTimestampFromMap(record)
				values[i] = time.Unix(safeTs/1000, 0).In(serverLoc)
			} else if col == "time" && tableName == "sensors.event" {
				safeTs := safeEventTimestampFromMap(record)
				values[i] = time.Unix(safeTs/1000, 0).In(serverLoc)
			} else {
				if val, ok := record[col]; ok {
					values[i] = val
				} else {
					values[i] = nil
				}
			}
		}

		_, err = stmt.Exec(values...)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("failed to execute insert: %v", err)
		}
	}

	return nil
}

// addColumnToClickHouseTyped 添加带类型的列
func addColumnToClickHouseTyped(columnName, ckType, tableName string) error {
	if len(columnName) == 0 || len(columnName) > 64 {
		return fmt.Errorf("invalid column name length: %d", len(columnName))
	}
	for _, c := range columnName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '$') {
			return fmt.Errorf("invalid column name format: contains illegal character")
		}
	}

	validTables := map[string]bool{
		"sensors.event": true,
		"sensors.user":  true,
	}
	if !validTables[tableName] {
		return fmt.Errorf("invalid table name: %s", tableName)
	}

	ckClient := initialize.Clickhouse

	parts := strings.Split(tableName, ".")
	dbName, tblName := parts[0], parts[1]
	checkQuery := `SELECT count() FROM system.columns WHERE database = ? AND table = ? AND name = ?`
	var count int
	err := ckClient.QueryRow(checkQuery, dbName, tblName, columnName).Scan(&count)
	if err != nil {
		logx.Errorf("check column exists failed: column=%s, err=%v", columnName, err)
		return err
	}

	if count > 0 {
		logx.Infof("column already exists: %s", columnName)
		return nil
	}

	columnDef := common.GetClickHouseColumnDefinition(columnName, ckType)
	alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tableName, columnDef)
	_, err = ckClient.Exec(alterQuery)
	if err != nil {
		if strings.Contains(err.Error(), "column with this name already exists") ||
			strings.Contains(err.Error(), "duplicate column") {
			logx.Infof("column already exists (concurrent add): %s", columnName)
			return nil
		}
		logx.Errorf("failed to add column: %v", err)
		return err
	}

	logx.Infof("successfully added typed column: %s %s to %s", columnName, ckType, tableName)
	return nil
}

func safeEventTimestampFromMap(data map[string]interface{}) int64 {
	if data == nil {
		return time.Now().UnixMilli()
	}

	timeVal, ok := data["time"]
	if !ok || timeVal == nil {
		return time.Now().UnixMilli()
	}

	switch v := timeVal.(type) {
	case int64:
		return normalizeTimestamp(v)
	case int:
		return normalizeTimestamp(int64(v))
	case float64:
		return normalizeTimestamp(int64(v))
	case string:
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			return ts.UnixMilli()
		}
		return time.Now().UnixMilli()
	case time.Time:
		return v.UnixMilli()
	default:
		return time.Now().UnixMilli()
	}
}

func normalizeTimestamp(ts int64) int64 {
	if ts <= 0 {
		return time.Now().UnixMilli()
	}
	if ts < 1e11 {
		return ts * 1000
	}
	if ts < 1e14 {
		return ts
	}
	if ts < 1e17 {
		return ts / 1000
	}
	return ts / 1e6
}
