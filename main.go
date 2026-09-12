package main

import (
	"github.com/beego/beego/v2/core/logs"
	"sensors/common/cron"
	"sensors/filters/permission"
	_ "sensors/initialize"
	_ "sensors/routers"

	beego "github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	beego.BConfig.WebConfig.DirectoryIndex = true
	beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"

	fb := &permission.FilterChainBuilder{}
	// 说明：不用全匹配的方式匹配路由，某些接口请求因不需要token校验所以需要排除
	beego.InsertFilterChain("/*", fb.Permssion)
	beego.SetStaticPath("/xlsx", "xlsx")

	cron.Cron()
	logs.SetLevel(logs.LevelWarning)
	beego.Run()
}
