package pkg

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

/*
	自定义调度插件：自定义最大POD数量的调度插件
*/

const TestScoreSchedulingName = "test-pod-score-scheduler"

// 调度器对象
type TestScoreScheduling struct {
}

const (
	NodeName = "vm-0-16-centos"
)

func (s *TestScoreScheduling) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	if nodeName == NodeName {
		return 20, framework.NewStatus(framework.Success)
	}
	return 10, framework.NewStatus(framework.Success)
}

func (s *TestScoreScheduling) NormalizeScore(ctx context.Context, state *framework.CycleState, p *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var min, max int64 = 0, 0
	//求出最小分数和最大分数
	for _, score := range scores {
		if score.Score < min {
			min = score.Score
		}
		if score.Score > max {
			max = score.Score
		}
	}
	// 特殊处理
	if max == min {
		min = min - 1
	}
	// 得分
	for i, score := range scores {
		scores[i].Score = (score.Score - min) * framework.MaxNodeScore / (max - min)
		klog.Infof("节点: %v, Score: %v   Pod:  %v", scores[i].Name, scores[i].Score, p.GetName())
	}
	return framework.NewStatus(framework.Success, "")

}

func (s *TestScoreScheduling) ScoreExtensions() framework.ScoreExtensions {
	return s
}

func (s *TestScoreScheduling) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

func (s *TestScoreScheduling) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

// Filter 过滤方法 (过滤node条件)
func (s *TestScoreScheduling) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.Info("开始过滤")

	return framework.NewStatus(framework.Success)

}

// PreFilter 前置过滤方法 (过滤pod条件)
func (s *TestScoreScheduling) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	klog.Info("开始预过滤")
	return framework.NewStatus(framework.Success)
}

func (s *TestScoreScheduling) PreFilterExtensions() framework.PreFilterExtensions {
	return s
}

func (*TestScoreScheduling) Name() string {
	return TestSchedulingName
}

// 检查是否实现接口对象
var _ framework.PreFilterPlugin = &TestScoreScheduling{}
var _ framework.FilterPlugin = &TestScoreScheduling{}
var _ framework.ScorePlugin = &TestScoreScheduling{}

func NewTestScoreScheduling(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	// 注入配置文件参数
	args := &Args{}
	if err := frameworkruntime.DecodeInto(configuration, args); err != nil {
		return nil, err
	}

	return &TestPodNumScheduling{}, nil
}
