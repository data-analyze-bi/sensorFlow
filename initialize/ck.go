package initialize

import (
	"database/sql"
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/tal-tech/go-zero/core/logx"
	"os"
	"sensors/common"
	"sync"

	_ "github.com/ClickHouse/clickhouse-go"
)

var (
	clickOnce     sync.Once
	Clickhouse    *sql.DB
	clickhouseDSN = "tcp://%v?username=%v&password=%v&database=%v"
)

func CkInit() {
	initSchemaCache()
	clint, err := getClickhouseInstance()
	if err != nil {
		logx.Errorf("getClickhouseInstance err: %v", err)
		return
	}
	if err := loadTableSchema(clint, "sensors.event"); err != nil {
		logx.Errorf("loadTableSchema failed: %v", err)
	}
	if err := loadUsersSchema(clint, "sensors.user"); err != nil {
		logx.Errorf("loadUsersSchema failed: %v", err)
	}
}

func initSchemaCache() {
	common.CachedTableSchema.Mutex.Lock()
	if common.CachedTableSchema.Columns == nil {
		common.CachedTableSchema.Columns = make(map[string]bool)
	}
	common.CachedTableSchema.Mutex.Unlock()

	common.CachedUsersSchema.Mutex.Lock()
	if common.CachedUsersSchema.Columns == nil {
		common.CachedUsersSchema.Columns = make(map[string]bool)
	}
	common.CachedUsersSchema.Mutex.Unlock()
}

func loadTableSchema(db *sql.DB, tableName string) error {
	rows, err := db.Query("DESCRIBE TABLE " + tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	common.CachedTableSchema.Mutex.Lock()
	defer common.CachedTableSchema.Mutex.Unlock()
	common.CachedTableSchema.Columns = make(map[string]bool)

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			logx.Error(err.Error())
			return err
		}
		strVal, ok := values[0].(string)
		if ok {
			common.CachedTableSchema.Columns[strVal] = true
		}

	}
	return nil
}
func loadUsersSchema(db *sql.DB, tableName string) error {
	rows, err := db.Query("DESCRIBE TABLE " + tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	common.CachedUsersSchema.Mutex.Lock()
	defer common.CachedUsersSchema.Mutex.Unlock()
	common.CachedUsersSchema.Columns = make(map[string]bool)

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			logx.Error(err.Error())
			return err
		}
		strVal, ok := values[0].(string)
		if ok {
			common.CachedUsersSchema.Columns[strVal] = true
		}
	}
	return nil
}
func getClickhouseInstance() (*sql.DB, error) {
	host, _ := beego.AppConfig.String("ckHost")
	username, _ := beego.AppConfig.String("ckUsername")
	database, _ := beego.AppConfig.String("ckDatabase")
	pwd, _ := beego.AppConfig.String("ckPwd")
	host = configValue("CLICKHOUSE_HOST", host)
	username = configValue("CLICKHOUSE_USER", username)
	database = configValue("CLICKHOUSE_DB", database)
	if envPwd, exists := os.LookupEnv("CLICKHOUSE_PASSWORD"); exists {
		pwd = envPwd
	}
	clickhouseDSN = fmt.Sprintf(clickhouseDSN, host, username, pwd, database)
	clickOnce.Do(func() {
		var err error
		Clickhouse, err = sql.Open("clickhouse", clickhouseDSN)
		if err != nil {
			logx.Error(err.Error())
			return
		}
		// 检查连接是否成功
		if err = Clickhouse.Ping(); err != nil {
			logx.Error(err.Error())
			return
		}
	})

	return Clickhouse, nil
}

func configValue(name, fallback string) string {
	if value, exists := os.LookupEnv(name); exists && value != "" {
		return value
	}
	return fallback
}
