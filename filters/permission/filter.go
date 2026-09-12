package permission

import (
	"sensors/controllers"
	_ "sensors/routers"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	_ "github.com/go-sql-driver/mysql"
)

const bearerLength = len("Bearer ")

type FilterChainBuilder struct {
}

func (builder *FilterChainBuilder) Permssion(next beego.FilterFunc) beego.FilterFunc {
	return func(ctx *context.Context) {
		url := ctx.Input.URL()
		ignoreUrls, _ := beego.AppConfig.String("ignore_urls")
		allowed := false
		for _, ignoredURL := range strings.Split(ignoreUrls, ",") {
			if strings.TrimSpace(ignoredURL) == url {
				allowed = true
				break
			}
		}
		if allowed {
			next(ctx)
		} else {
			ctx.Output.JSON(controllers.ErrMsg("无权限", 40001), true, true)
			return

		}
	}
}
