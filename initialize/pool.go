package initialize

import (
	beego "github.com/beego/beego/v2/server/web"
	"os"
	"os/signal"
	"sensors/common/redis"
	"syscall"
)

var (
	PoolDef *redis.Redis
)

func poolInit() {

	// default admin
	hostAdmin, _ := beego.AppConfig.String("redisHostAdmin")
	typeAdmin, _ := beego.AppConfig.String("redisAdminType")
	pwdAdmin, _ := beego.AppConfig.String("redisPwdAdmin")
	hostAdmin = configValue("REDIS_HOST", hostAdmin)
	typeAdmin = configValue("REDIS_TYPE", typeAdmin)
	if envPwd, exists := os.LookupEnv("REDIS_PASSWORD"); exists {
		pwdAdmin = envPwd
	}
	defaultRedisConn := redis.NewRedis(hostAdmin, typeAdmin, pwdAdmin)

	//
	PoolDef = NewPool(defaultRedisConn)

	cleanupHook()
}

func NewPool(store *redis.Redis) *redis.Redis {
	return store
}

func cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		os.Exit(0)
	}()
}
