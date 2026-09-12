package cron

import (
	"github.com/go-co-op/gocron"
	"sensors/common"
	"sensors/controllers"
)

func Cron() {
	loc := common.ServerLocation()
	s1 := gocron.NewScheduler(loc)
	s1.Every(6).Seconds().Do(controllers.CronInsertCk)
	s1.StartAsync()
	s2 := gocron.NewScheduler(loc)
	s2.Every(5).Seconds().Do(controllers.CronInsertUsersCk)
	s2.StartAsync()
}
