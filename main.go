package main

import (
	"fmt"
	"k8s-scheduler-plugins-practice/pkg/scheduler"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"os"
)

func main() {
	// 接入插件
	cmd := app.NewSchedulerCommand(
		// 调度插件
		app.WithPlugin(scheduler.TestSchedulingName, scheduler.NewTestPodNumScheduling),
		app.WithPlugin(scheduler.TestScoreSchedulingName, scheduler.NewTestScoreScheduling),
		app.WithPlugin(scheduler.PluginName, scheduler.NewUniformResourcePlugin),
	)

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
