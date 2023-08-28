package scheduler

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

/*
	自定义调度插件：自定义最大POD数量的调度插件
*/

const TestSchedulingName = "test-pod-maxnum-scheduler"

// TestPodNumScheduling 调度器对象
type TestPodNumScheduling struct {
	fact informers.SharedInformerFactory
	args *Args
}

// Args 配置文件参数
type Args struct {
	MaxPods int `json:"maxPods,omitempty"`
}

func (s *TestPodNumScheduling) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

func (s *TestPodNumScheduling) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

// 通过label过滤不能调度的节点
const (
	SchedulingLabelKeyState   = "scheduling"
	SchedulingLabelValueState = "true"
)

// Filter 过滤方法 (过滤node条件)
func (s *TestPodNumScheduling) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {

	for k, v := range nodeInfo.Node().Labels {
		if k == SchedulingLabelKeyState && v != SchedulingLabelValueState {
			klog.V(3).Infof("这个节点设置不可调度\n")
			return framework.NewStatus(framework.Unschedulable, "这个节点设置不可调度")
		}
	}
	return framework.NewStatus(framework.Success)

}

// PreFilter 前置过滤方法 (过滤pod条件)
func (s *TestPodNumScheduling) PreFilter(_ context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	klog.V(3).Infof("当前被prefilter 的POD名称是:%s\n", p.Name)
	// informer list pod
	podList, err := s.fact.Core().V1().Pods().Lister().Pods(p.Namespace).List(labels.Everything())
	if err != nil {
		klog.V(3).Infof("POD informer list 发生错误\n")
		return framework.NewStatus(framework.Error)
	}

	// 过滤
	if s.args.MaxPods > 0 && len(podList) > s.args.MaxPods {
		klog.V(3).Infof("POD数量超过可调度数量，不能调度\n", p.Name)
		return framework.NewStatus(framework.Unschedulable,
			fmt.Sprintf("POD数量超过，不能调度，最多只能调度%d", s.args.MaxPods))
	}
	klog.V(3).Infof("POD成功调度:%s\n", p.Name)
	return framework.NewStatus(framework.Success)
}

func (s *TestPodNumScheduling) PreFilterExtensions() framework.PreFilterExtensions {
	return s
}

func (*TestPodNumScheduling) Name() string {
	return TestSchedulingName
}

// 检查是否实现接口对象
var _ framework.PreFilterPlugin = &TestPodNumScheduling{}
var _ framework.FilterPlugin = &TestPodNumScheduling{}

func NewTestPodNumScheduling(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	// 注入配置文件参数
	args := &Args{}
	if err := frameworkruntime.DecodeInto(configuration, args); err != nil {
		return nil, err
	}

	return &TestPodNumScheduling{
		fact: f.SharedInformerFactory(),
		args: args,
	}, nil
}
