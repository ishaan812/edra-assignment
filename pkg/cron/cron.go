package cronJobHandler

import "github.com/robfig/cron/v3"

var cronJobHandler *cron.Cron

func GetCronJobHandler() *cron.Cron {
	return cronJobHandler
}

func InitCron() {
	cronJobHandler = cron.New()
	cronJobHandler.Start()
}
