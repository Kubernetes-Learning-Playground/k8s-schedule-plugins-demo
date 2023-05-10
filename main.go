package main

import (
	"fmt"
	"k8s-scheduler-plugins-practice/pkg"
	"k8s-scheduler-plugins-practice/pkg/noderesource"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"os"
)

func main() {
	// 接入插件
	cmd := app.NewSchedulerCommand(
		app.WithPlugin(pkg.TestSchedulingName, pkg.NewTestPodNumScheduling), // 调度插件
		app.WithPlugin(pkg.TestScoreSchedulingName, pkg.NewTestScoreScheduling),
		app.WithPlugin(noderesource.TestNodeResourceSchedulingName, noderesource.NewAllocatable),
	)

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
