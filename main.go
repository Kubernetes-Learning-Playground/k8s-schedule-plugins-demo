package main

import (
	"fmt"
	"k8s-scheduler-plugins-practice/src"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"os"
)
func main() {
	//
	command := app.NewSchedulerCommand(
		app.WithPlugin(src.TestSchedulingName, src.NewTestPodNumScheduling), // 调度插件
	)
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
