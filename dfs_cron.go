package main

import (
	"fmt"
	"github.com/robfig/cron"
)

func (this *Server) strategyCheckNode() {
	fmt.Println("strategyCheckNode start")
	fmt.Println("strategyCheckNode end")
}

func (this *Server) strategyMove() {

}

func (this *Server) checkStorage() {

	maxStorage := Config().MaxStorage * 1024 * 1024 * 1024
	useStorage, _ := this.db.GetSize()
	percent := float64(useStorage) / float64(maxStorage)

	if percent > 0.8 {

	}
	dPrint("checkStorage:", maxStorage, useStorage, percent)

}

func (this *Server) initCron() {

	c := cron.New()
	c.AddFunc("@every 3s", func() {
		this.checkStorage()
	})

	// _, e := c.AddFunc("0/1 * * * ?", func() {
	// 	dPrint("schedule every two seconds ...")
	// })
	// if e != nil {
	// 	dPrint("添加任务失败: " + e.Error())
	// }
	c.Start()
}
