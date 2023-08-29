package scheduler

import (
	"context"
	"fmt"
	"k8s-scheduler-plugins-practice/pkg/k8s_config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
)

/*
	自定义调度插件：查看资源cpu mem后打分的调度插件
*/

const (
	PluginName = "test-uniform-scheduler"
)

type UniformResourcePlugin struct {
	handle  framework.Handle
}

var ccc *versioned.Clientset
var err error

func NewUniformResourcePlugin(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	cfg := k8s_config.K8sRestConfig()
	ccc, err = versioned.NewForConfig(cfg)
	if err != nil {
		log.Fatal("versioned config error: ", err)
	}
	// 注入配置文件参数
	args := &Args{}
	if err := frameworkruntime.DecodeInto(configuration, args); err != nil {
		return nil, err
	}
	return &UniformResourcePlugin{handle: f}, nil
}


func (up *UniformResourcePlugin) Name() string {
	return PluginName
}

var _ framework.FilterPlugin = &UniformResourcePlugin{}
var _ framework.ScorePlugin = &UniformResourcePlugin{}

func (up *UniformResourcePlugin) Filter(_ context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.Info("uniform plugin filter step start...")
	cpuRequest := pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU]
	memRequest := pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory]

	n := nodeInfo.Node()
	availableCpu := n.Status.Allocatable[v1.ResourceCPU]
	availableMem := n.Status.Allocatable[v1.ResourceMemory]
	if availableCpu.Cmp(cpuRequest) < 0 || availableMem.Cmp(memRequest) < 0 {
		schedulerMessage := fmt.Sprintf("insufficient resources on node %s", n.Name)
		return framework.NewStatus(framework.Unschedulable, schedulerMessage)
	}

	return framework.NewStatus(framework.Success)
}

func (up *UniformResourcePlugin) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.Info("uniform plugin score step start...")
	allNodeInfos, err := up.handle.SnapshotSharedLister().NodeInfos().List()
	if err != nil {
		errMessage := fmt.Sprintf("snapshot node list error: %s", err)
		return 0, framework.NewStatus(framework.Unschedulable, errMessage)
	}


	for _, nodeInfo := range allNodeInfos {
		if nodeInfo.Node().Name == nodeName {
			n := nodeInfo.Node()

			// FIXME: 需要使用metrics-server 接口，不能使用node Allocatable 字段
			nodeMetric, err := ccc.MetricsV1beta1().NodeMetricses().Get(context.Background(), n.Name, metav1.GetOptions{})
			if err != nil {
				klog.Errorf("ccc.MetricsV1beta1().NodeMetricses() error: %v", err)
			}

			//availableCpu := n.Status.Allocatable[v1.ResourceCPU]
			//availableMem := n.Status.Allocatable[v1.ResourceMemory]

			availableCpu := nodeMetric.Usage.Cpu()
			availableMem := nodeMetric.Usage.Memory()

			capacityCpu := n.Status.Capacity[v1.ResourceCPU]
			capacityMem := n.Status.Capacity[v1.ResourceMemory]

			cpuUtilization := float64(availableCpu.MilliValue()) / float64(availableCpu.MilliValue() + capacityCpu.MilliValue())
			memUtilization := float64(availableMem.MilliValue()) / float64(availableMem.MilliValue() + capacityMem.MilliValue())
			klog.Infof("availableCpu:%v, availableMem:%v, capacityCpu:%v, capacityMem:%v, cpuUtilization:%v, " +
				"memUtilization: %v", availableCpu.MilliValue(), availableMem.MilliValue(), capacityCpu.MilliValue(),
				capacityMem.MilliValue(), cpuUtilization, memUtilization)

			score := int64((1-cpuUtilization) * (1-memUtilization) * 100)

			return score, framework.NewStatus(framework.Success)
		}
	}
	return 0, framework.NewStatus(framework.Error)
}

func (up *UniformResourcePlugin) ScoreExtensions() framework.ScoreExtensions {
	return up
}

func (up *UniformResourcePlugin) NormalizeScore(_ context.Context, state *framework.CycleState, p *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	fmt.Println("scores ", scores)

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
		klog.Infof("[Node]: %v, [Score]: %v, [Pod]:  %v", scores[i].Name, scores[i].Score, p.GetName())
	}
	return framework.NewStatus(framework.Success, "")
}







