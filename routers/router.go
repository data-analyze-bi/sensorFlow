package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"sensors/controllers"
)

func init() {

	//埋点数据接收
	ns1 := beego.NewNamespace("/sensors",
		beego.NSNamespace("/send",
			beego.NSInclude(
				&controllers.BaseController{},
			),
		),
	)
	beego.AddNamespace(ns1)
	//beego.SetStaticPath("/swagger/", "swagger")
}
