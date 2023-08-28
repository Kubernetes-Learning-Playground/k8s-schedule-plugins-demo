package scheduler

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"path/filepath"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MySchedulerPlugin struct {

}

func (p *MySchedulerPlugin) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	// 过滤函数，根据节点的资源使用情况过滤节点
	// 在这里实现您的节点筛选逻辑
	// 如果节点满足资源需求，返回 nil；否则，返回一个非 nil 的 Status 对象，指示节点不可用

	// 获取 Pod 的资源需求
	cpuRequests := pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU]
	memoryRequests := pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory]

	// 获取节点的资源容量
	cpuCapacity := nodeInfo.Node().Status.Capacity[v1.ResourceCPU]
	memoryCapacity := nodeInfo.Node().Status.Capacity[v1.ResourceMemory]

	// 检查节点的资源容量是否满足 Pod 的资源需求
	if cpuCapacity.Cmp(cpuRequests) < 0 || memoryCapacity.Cmp(memoryRequests) < 0 {
		// 节点资源不足，返回不可用的状态
		return framework.NewStatus(framework.Error, "Insufficient resources on the node")
	}

	// 节点资源满足需求，返回 nil
	return nil
}

func (p *MySchedulerPlugin) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// 评分函数，根据节点的资源利用率计算节点的评分
	// 在这里实现您的评分算法逻辑

	// 获取节点的资源利用率
	cpuUtilization, memoryUtilization, _ := getNodeResourceUtilization(nodeName)

	// 实现资源利用率评分算法，根据 CPU 和内存利用率计算节点的评分
	score := calculateResourceScore(cpuUtilization, memoryUtilization)

	// 返回节点评分
	return score, nil
}

// 获取节点的资源利用率
func getNodeResourceUtilization(nodeName string) (float64, float64, error) {
	// 建立与 Kubernetes 集群的连接
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return 0, 0, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return 0, 0, err
	}

	// 获取节点的资源利用率
	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return 0, 0, err
	}

	cpuUtilization := calculateAverageUtilization(node.Status.Capacity.Cpu(), node.Status.Allocatable.Cpu())
	memoryUtilization := calculateAverageUtilization(node.Status.Capacity.Memory(), node.Status.Allocatable.Memory())

	return cpuUtilization, memoryUtilization, nil
}

// 计算资源利用率的平均利用率
func calculateAverageUtilization(capacity, allocatable resourceQuantity) float64 {
	capacityValue := float64(capacity.Value())
	allocatableValue := float64(allocatable.Value())

	if capacityValue == 0 {
		return 0
	}

	utilization := (1 - (allocatableValue / capacityValue)) * 100
	return utilization
}

// 实现资源利用率评分算法，根据 CPU 和内存利用率计算节点的评分
func calculateResourceScore(cpuUtilization, memoryUtilization float64) int64 {
	// 在这里实现您的评分算法逻辑
	// ...

	// 假设简单的评分算法：将 CPU 和内存利用率归一化到范围 [0, 100]，然后计算平均值作为节点的评分
	normalizedCPU := normalize(cpuUtilization)
	normalizedMemory := normalize(memoryUtilization)
	average := (normalizedCPU + normalizedMemory) / 2

	// 将评分转换为 int64 类型并返回
	return int64(average)
}

// 归一化资源利用率到范围 [0, 100]
func normalize(utilization float64) float64 {
	// 在这里实现归一化逻辑
	// ...

	// 这里假设资源利用率已经是在范围 [0, 100] 内，不需要进行归一化处理
	return utilization
}

// 辅助函数，用于获取资源的数量
type resourceQuantity v1.ResourceQuantity

// 辅助函数，用于从资源的数量中获取值
func (r resourceQuantity) Value() int64 {

	value, _ := r.AsInt64()
	return value
}

// 辅助函数，用于将资源的数量转换为 int64 类型
func (r resourceQuantity) AsInt64() (int64, error) {
	return r.AsScale(0)
}

// 辅助函数，用于将资源的数量按照给定的 scale 转换为 int64 类型
func (r resourceQuantity) AsScale(scale int32) (int64, error) {
	return r.AsDec().ScaledValue(scale).Round(0).AsInt64()
}
