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
	dPrint("checkStorage start")

	fmt.Println(Config().MaxStorage * 1024 * 1024)

	dPrint("checkStorage end")
}

func (this *Server) initCron() {

	c := cron.New()
	// c.AddFunc("@every 3s", func() {
	// 	this.checkStorage()
	// })

	// _, e := c.AddFunc("0/1 * * * ?", func() {
	// 	dPrint("schedule every two seconds ...")
	// })
	// if e != nil {
	// 	dPrint("添加任务失败: " + e.Error())
	// }
	c.Start()
}
