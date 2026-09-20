package controllers

import (
	"encoding/json"
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/tal-tech/go-zero/core/logx"
	"sensors/common"
	"sensors/initialize"
	"strconv"
	"strings"
)

type BaseController struct {
	beego.Controller
}

func (c *BaseController) URLMapping() {
	c.Mapping("Post", c.Post)
}

// @Title 埋点
// @Success 200 {object} controllers.Result
// @router / [post]
func (c *BaseController) Post() {
	data := c.GetString("data_list")
	if data == "" {
		data = c.GetString("data")
	}
	if data == "" {
		c.Fail("payload missing", 400)
		return
	}
	if len(data) > common.MaxEncodedPayloadBytes {
		c.Fail("payload too large", 413)
		return
	}

	// 埋点服务通过外部二进制还原 payload，避免明文解密逻辑暴露在服务源码内。
	logx.Infof("payload license processing start, encoded_len=%d", len(data))
	uncompressedData, err := common.ProcessPayloadByLicense(data)
	if err != nil {
		logx.Errorf("payload license processing failed:%v", err.Error())
		c.Fail("payload decode failed", 500)
		return
	}
	logx.Infof("payload license processing done, restored_len=%d", len(uncompressedData))

	var (
		events   []map[string]interface{}
		goEvents map[string]interface{}
		jsonData []string
		users    map[string]interface{}
	)

	// 尝试解析为事件数组
	err = json.Unmarshal(uncompressedData, &events)
	if err != nil {
		logx.Errorf("failed to unmarshal as events array: %v", err)
		events = nil
	}

	// 如果不是数组，尝试解析为单个事件对象
	if len(events) == 0 {
		err = json.Unmarshal(uncompressedData, &goEvents)
		if err != nil {
			logx.Errorf("failed to unmarshal as single event: %v", err)
		} else if eventType, ok := safeGetString(goEvents, "type"); ok && eventType == "track" {
			events = append(events, goEvents)
		}
	}

	// 如果都不是事件，尝试解析为用户属性
	if len(events) == 0 {
		err = json.Unmarshal(uncompressedData, &users)
		if err != nil {
			logx.Errorf("failed to unmarshal as user properties: %v", err)
			c.Fail("invalid data format", 400)
			return
		}
	}
	//事件入redis
	pool := initialize.PoolDef
	for _, v := range events {
		v = common.RemoveDollarSign(v)
		v = common.FlattenJSON(v)
		if v["event"] == nil {
			//
			continue
		}
		insertDataIntoClickHouse(v) //自动新增字段
		tmp, _ := json.Marshal(v)
		pool.Hset(common.CacheEventsName, v["event"].(string), common.Int64ToStr(common.GetUnixTime()))
		_, err = pool.Sadd(fmt.Sprintf("%v%v", common.CacheEventsData, v["event"]), string(tmp))
		if err != nil {
			jsonData = append(jsonData, string(tmp))
			logx.Errorf("CacheEventsData err:%v", err.Error())

		}
	}
	//用户属性
	if users != nil {
		users = common.RemoveDollarSign(users)
		users = common.FlattenJSON(users)
		method, _ := safeGetString(users, "type")
		for _, key := range keysToRemove {
			delete(users, key)
		}
		insertUsersDataIntoClickHouse(users) //用户属性添加
		usersTmp, err := common.ValuesToStrings(users)
		if err != nil {
			logx.Errorf("ValuesToStrings users err:%v", err.Error())
		}
		distinctId := usersTmp["distinct_id"]
		switch method {
		case common.TrackSignup: //注册
			pool.Hmset(fmt.Sprintf("%v%v", common.CacheUserInfo, distinctId), usersTmp)

		case common.TrackProfileSet: //用户属性
			pool.Hmset(fmt.Sprintf("%v%v", common.CacheUserInfo, distinctId), usersTmp)
		}
		_, err = pool.Sadd(common.CacheUsersLog, distinctId)
		if err != nil {
			DeleteData(initialize.Clickhouse, "sensors.user", distinctId)
			usersInfoJson, _ := json.Marshal(users)
			InsertData(initialize.Clickhouse, "sensors.user", []string{string(usersInfoJson)})
		}

	}
	if len(jsonData) > 0 { //没入库redis直接入库ck
		InsertData(initialize.Clickhouse, "sensors.event", jsonData)
	}

	logx.Info("Data processed successfully")
	c.Ok(map[string]interface{}{"content": "", "totalElements": 0})
}

func insertDataIntoClickHouse(data map[string]interface{}) {
	logx.Info("insertDataIntoClickHouse start", data, common.CachedTableSchema)
	common.CachedTableSchema.Mutex.Lock()
	defer common.CachedTableSchema.Mutex.Unlock()
	// 检查是否有新的key
	for k := range data {
		if !common.CachedTableSchema.Columns[k] {
			logx.Infof("insertDataIntoClickHouse add column start, column=%s", k)
			err := addColumnToClickHouse(k)
			if err != nil {
				logx.Errorf("insertDataIntoClickHouse add column failed, column=%s err=%v", k, err)
				continue
			}
			common.CachedTableSchema.Columns[k] = true

		}
	}
	logx.Info("insertDataIntoClickHouse end")
	return
}
func insertUsersDataIntoClickHouse(data map[string]interface{}) {
	common.CachedUsersSchema.Mutex.Lock()
	defer common.CachedUsersSchema.Mutex.Unlock()
	// 检查是否有新的key
	for k := range data {
		if !common.CachedUsersSchema.Columns[k] {
			// 锁定写入，添加新列
			addUsersColumnToClickHouse(k)
			common.CachedUsersSchema.Columns[k] = true
		}
	}
	return
}

func addColumnToClickHouse(newColumnName string) error {
	// 验证列名，防止SQL注入
	if len(newColumnName) == 0 || len(newColumnName) > 64 {
		return fmt.Errorf("invalid column name length: %d", len(newColumnName))
	}
	for _, c := range newColumnName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			return fmt.Errorf("invalid column name format: contains illegal character")
		}
	}

	ckClient := initialize.Clickhouse

	// 先检查列是否已存在，防止并发重复添加
	checkQuery := `SELECT count() FROM system.columns
		WHERE database = 'sensors' AND table = 'event' AND name = ?`
	var count int
	err := ckClient.QueryRow(checkQuery, newColumnName).Scan(&count)
	if err != nil {
		logx.Errorf("check column exists failed: column=%s, err=%v", newColumnName, err)
		return err
	}

	if count > 0 {
		logx.Infof("column already exists: %s", newColumnName)
		return nil
	}

	// 列不存在才添加
	alterQuery := fmt.Sprintf("ALTER TABLE sensors.event ADD COLUMN `%s` String", newColumnName)
	_, err = ckClient.Exec(alterQuery)
	if err != nil {
		// 检查是否是因为列已存在导致的错误（并发情况）
		if strings.Contains(err.Error(), "column with this name already exists") ||
			strings.Contains(err.Error(), "duplicate column") {
			logx.Infof("column already exists (concurrent add): %s", newColumnName)
			return nil
		}
		logx.Errorf("addColumnToClickHouse err:%v", err.Error())
		return err
	}
	logx.Infof("successfully added column: %s", newColumnName)
	return nil
}
func addUsersColumnToClickHouse(newColumnName string) error {
	// 验证列名，防止SQL注入
	if len(newColumnName) == 0 || len(newColumnName) > 64 {
		logx.Errorf("invalid column name length: %d", len(newColumnName))
		return fmt.Errorf("invalid column name length: %d", len(newColumnName))
	}
	for _, c := range newColumnName {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_') {
			logx.Errorf("invalid column name format: contains illegal character")
			return fmt.Errorf("invalid column name format: contains illegal character")
		}
	}

	ckClient := initialize.Clickhouse

	// 先检查列是否已存在，防止并发重复添加
	checkQuery := `SELECT count() FROM system.columns
		WHERE database = 'sensors' AND table = 'user' AND name = ?`
	var count int
	err := ckClient.QueryRow(checkQuery, newColumnName).Scan(&count)
	if err != nil {
		logx.Errorf("check column exists failed: column=%s, err=%v", newColumnName, err)
		return err
	}

	if count > 0 {
		logx.Infof("column already exists: %s", newColumnName)
		return nil
	}

	// 列不存在才添加
	alterQuery := fmt.Sprintf("ALTER TABLE sensors.user ADD COLUMN `%s` String", newColumnName)
	_, err = ckClient.Exec(alterQuery)
	if err != nil {
		// 检查是否是因为列已存在导致的错误（并发情况）
		if strings.Contains(err.Error(), "column with this name already exists") ||
			strings.Contains(err.Error(), "duplicate column") {
			logx.Infof("column already exists (concurrent add): %s", newColumnName)
			return nil
		}
		logx.Errorf("addUsersColumnToClickHouse err:%v", err.Error())
		return err
	}
	logx.Infof("successfully added user column: %s", newColumnName)
	return nil
}

func safeGetString(data map[string]interface{}, key string) (string, bool) {
	if data == nil {
		return "", false
	}
	v, ok := data[key]
	if !ok || v == nil {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), true
	case int64:
		return strconv.FormatInt(val, 10), true
	case int:
		return strconv.Itoa(val), true
	case bool:
		if val {
			return "true", true
		}
		return "false", true
	default:
		bytes, err := json.Marshal(val)
		if err != nil {
			return "", false
		}
		return string(bytes), true
	}
}

type any = interface{}

type Result struct {
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
	Status int         `json:"status"`
}

func (c *BaseController) Ok(data any) {
	c.Data["json"] = SuccessData(data)
	c.ServeJSON()
}

func (c *BaseController) Fail(msg string, status int) {
	c.Data["json"] = ErrMsg(msg, status)
	c.ServeJSON()
}

func ErrMsg(msg string, status ...int) Result {
	var r Result
	if len(status) > 0 {
		r.Status = status[0]
	} else {
		r.Status = 500000
	}
	r.Msg = msg
	r.Data = nil

	return r
}

func SuccessData(data any) Result {
	var r Result

	r.Status = 200
	r.Msg = "ok"
	r.Data = data

	return r
}

var keysToRemove = []string{"type", "event"}

// min 返回两个整数的最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
