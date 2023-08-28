package scheduler

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

const CustomSchedulingName = "test-custom-scheduler"

var _ framework.FilterPlugin = &CustomPlugin{}
var _ framework.ScorePlugin = &CustomPlugin{}

// 自定义调度插件
type CustomPlugin struct{
	handle framework.Handle
}

func (p *CustomPlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
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
		klog.Infof("节点: %v, Score: %v   Pod:  %v", scores[i].Name, scores[i].Score, pod.GetName())
	}
	return framework.NewStatus(framework.Success, "")
}

func (p *CustomPlugin) ScoreExtensions() framework.ScoreExtensions {
	return p
}

// 实现scheduler.Plugin接口的函数，用于调度前的操作
func (p *CustomPlugin) Name() string {
	return CustomSchedulingName
}

func (p *CustomPlugin) Filter(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	fmt.Println("进入Filter")
	fmt.Println("filter node name: ", nodeInfo.Node().Name)
	fmt.Println("pod: ", pod.Name, pod.Namespace)
	//// 计算节点的资源得分
	//p.calculateNodeScores(nodeInfo)
	// Get node information from the framework handle.
	nodeInfoLister := p.handle.SnapshotSharedLister().NodeInfos()
	nodeInfo, err := nodeInfoLister.Get(nodeInfo.Node().Name)
	fmt.Println("nodeInfo: ", nodeInfo)
	for _, v := range nodeInfo.Pods {
		fmt.Println(v.Pod.Name)
	}
	if err != nil {
		return framework.NewStatus(framework.Unschedulable, "这个节点设置不可调度")
	}


	// Calculate the score based on node resources usage.
	cpuUsage := resource.NewQuantity(0, resource.DecimalSI)
	memUsage := resource.NewQuantity(0, resource.BinarySI)
	capacity := nodeInfo.Node().Status.Capacity

	// Iterate through all pods on the node and calculate CPU and memory usage.
	for _, pp := range nodeInfo.Pods {
		pod := pp.Pod
		for _, container := range pod.Spec.Containers {
			// Accumulate CPU requests and limits.
			if cpuReq := container.Resources.Requests[v1.ResourceCPU]; cpuReq.Value() != 0 {
				cpuUsage.Add(cpuReq)
			}
			//if cpuLim := container.Resources.Limits[v1.ResourceCPU]; cpuLim.Value() != 0 {
			//	cpuUsage.Add(cpuLim)
			//}

			// Accumulate memory requests and limits.
			if memReq := container.Resources.Requests[v1.ResourceMemory]; memReq.Value() != 0 {
				memUsage.Add(memReq)
			}
			//if memLim := container.Resources.Limits[v1.ResourceMemory]; memLim.Value() != 0 {
			//	memUsage.Add(memLim)
			//}
		}
	}


	// 假设评分逻辑为 CPU 使用量和内存使用量的加权和
	cpuWeight := int64(1)
	memWeight := int64(1)
	capWeight := int64(1)
	score := cpuWeight*cpuUsage.Value() + memWeight*memUsage.Value() + capWeight*capacity.Storage().Value()
	fmt.Println("score...: ", score)
	// Add additional scoring logic based on CPU, memory, and capacity usage.

	return nil
}

var NodeScoreMap = make(map[string]int)

// 计算节点的资源得分，存入全局map中
//func (p *CustomPlugin) calculateNodeScores(nodeInfo *framework.NodeInfo) map[string]int {
//	fmt.Println("node name: ", nodeInfo.Node().Name)
//	score := p.score(nodeInfo.Node())
//	NodeScoreMap[nodeInfo.Node().Name] = score
//	fmt.Println("score: ", score)
//	return NodeScoreMap
//}

// Score 方法用于评估节点的资源使用情况并返回得分
//func (p *CustomPlugin) score(node *v1.Node) int {
//	// 在这里根据节点的资源使用情况来计算得分
//	// 这里只是示例，你需要根据实际需求制定自己的得分计算逻辑
//
//	// 假设我们使用以下简单的逻辑来计算得分：
//	// CPU 使用率占总可用 CPU 的比例 * 100
//	cpuUsageRate := calculateCPUUsageRate(node)
//	// 内存使用率占总可用内存的比例 * 100
//	memoryUsageRate := calculateMemoryUsageRate(node)
//	// 存储使用率占总可用存储容量的比例 * 100
//	storageUsageRate := calculateStorageUsageRate(node)
//
//	// 综合各项得分计算总得分
//	score := int((cpuUsageRate + memoryUsageRate + storageUsageRate) / 3)
//	fmt.Println("有到这里～～～")
//	return score
//}

//
//func (p *CustomPlugin) Score(ctx context.Context, cycleState *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
//
//	// 选择得分最高的节点进行调度
//	selectedNode := p.selectNode(NodeScoreMap)
//
//	return selectedNode, nil
//}

func (p *CustomPlugin) Score(_ context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	// Get node information from the framework handle.
	nodeInfoLister := p.handle.SnapshotSharedLister().NodeInfos()
	nodeInfo, err := nodeInfoLister.Get(nodeName)
	fmt.Println("nodeInfo: ", nodeInfo)
	for _, v := range nodeInfo.Pods {
		fmt.Println(v.Pod.Name)
	}
	if err != nil {
		return 0, framework.NewStatus(framework.Error, err.Error())
	}


	// Calculate the score based on node resources usage.
	cpuUsage := resource.NewQuantity(0, resource.DecimalSI)
	memUsage := resource.NewQuantity(0, resource.BinarySI)
	capacity := nodeInfo.Node().Status.Capacity

	// Iterate through all pods on the node and calculate CPU and memory usage.
	for _, pp := range nodeInfo.Pods {
		pod := pp.Pod
		for _, container := range pod.Spec.Containers {
			// Accumulate CPU requests and limits.
			if cpuReq := container.Resources.Requests[v1.ResourceCPU]; cpuReq.Value() != 0 {
				cpuUsage.Add(cpuReq)
			}
			//if cpuLim := container.Resources.Limits[v1.ResourceCPU]; cpuLim.Value() != 0 {
			//	cpuUsage.Add(cpuLim)
			//}

			// Accumulate memory requests and limits.
			if memReq := container.Resources.Requests[v1.ResourceMemory]; memReq.Value() != 0 {
				memUsage.Add(memReq)
			}
			//if memLim := container.Resources.Limits[v1.ResourceMemory]; memLim.Value() != 0 {
			//	memUsage.Add(memLim)
			//}
		}
	}


	// 假设评分逻辑为 CPU 使用量和内存使用量的加权和
	cpuWeight := int64(1)
	memWeight := int64(1)
	capWeight := int64(1)
	score := cpuWeight*cpuUsage.Value() + memWeight*memUsage.Value() + capWeight*capacity.Storage().Value()
	fmt.Println("score...: ", score)
	// Add additional scoring logic based on CPU, memory, and capacity usage.

	return score, framework.NewStatus(framework.Success)
}

// 计算节点的 CPU 使用率
func calculateCPUUsageRate(node *v1.Node) float64 {
	// 在这里获取节点的 CPU 使用量和总可用 CPU，然后计算使用率
	// 这里只是示例，你需要根据实际情况获取节点的 CPU 使用量和总可用 CPU

	// 假设我们获取节点的 CPU 使用量和总可用 CPU
	usedCPU := getUsedCPU(node)
	availableCPU := node.Status.Allocatable[v1.ResourceCPU]
	// 计算 CPU 使用率
	cpuUsageRate := float64(usedCPU.MilliValue()) / float64(availableCPU.MilliValue()) * 100
	fmt.Printf("usedCPU: %v,availableCPU: %v,cpuUsageRate: %v\n", usedCPU, availableCPU, cpuUsageRate)

	return cpuUsageRate
}

// 计算节点的内存使用率
func calculateMemoryUsageRate(node *v1.Node) float64 {
	// 在这里获取节点的内存使用量和总可用内存，然后计算使用率
	// 这里只是示例，你需要根据实际情况获取节点的内存使用量和总可用内存

	// 假设我们获取节点的内存使用量和总可用内存
	usedMemory := getUsedMemory(node)
	availableMemory := node.Status.Allocatable[v1.ResourceMemory]

	// 计算内存使用率
	memoryUsageRate := float64(usedMemory.Value()) / float64(availableMemory.Value()) * 100
	fmt.Printf("usedMemory: %v,availableMemory: %v,memoryUsageRate: %v\n", usedMemory, availableMemory, memoryUsageRate)
	return memoryUsageRate
}

// 计算节点的存储使用率
func calculateStorageUsageRate(node *v1.Node) float64 {
	// 在这里获取节点的存储使用量和总可用存储容量，然后计算使用率
	// 这里只是示例，你需要根据实际情况获取节点的存储使用量和总可用存储容量

	// 假设我们获取节点的存储使用量和总可用存储容量
	usedStorage := getUsedStorage(node)
	availableStorage := node.Status.Allocatable[v1.ResourceStorage]

	// 计算存储使用率
	storageUsageRate := float64(usedStorage.Value()) / float64(availableStorage.Value()) * 100
	fmt.Printf("usedStorage: %v,availableStorage: %v,storageUsageRate: %v\n", usedStorage, availableStorage, storageUsageRate)

	return storageUsageRate
}

// 选择得分最高的节点进行调度
func (p *CustomPlugin) selectNode(scores map[string]int) int64 {
	//var selectedNode *v1.Node
	maxScore := -1

	// 遍历所有节点的得分，选择得分最高的节点
	for _, score := range scores {
		if score > maxScore {
			maxScore = score

			//selectedNode = &v1.Node{
			//	ObjectMeta: metav1.ObjectMeta{
			//		Name: nodeName,
			//	},
			//}
		}
	}

	return int64(maxScore)
}

// 获取节点的 CPU 使用量
func getUsedCPU(node *v1.Node) resource.Quantity {
	// 在这里根据你的调度器和集群环境获取节点的 CPU 使用量
	// 这里只是示例，你需要根据实际情况获取节点的 CPU 使用量

	// 假设我们获取节点的 CPU 使用量
	// 从节点的状态中获取已分配的 CPU 资源量
	usedCPU := node.Status.Allocatable[v1.ResourceCPU].DeepCopy()

	// 减去节点上未使用的 CPU 资源量，得到已使用的 CPU 资源量
	unusedCPU := node.Status.Capacity[v1.ResourceCPU]
	usedCPU.Sub(unusedCPU)

	return usedCPU
}

// 获取节点的内存使用量
func getUsedMemory(node *v1.Node) resource.Quantity {
	// 在这里根据你的调度器和集群环境获取节点的内存使用量
	// 这里只是示例，你需要根据实际情况获取节点的内存使用量

	// 假设我们获取节点的内存使用量
	// 从节点的状态中获取已分配的内存资源量
	usedMemory := node.Status.Allocatable[v1.ResourceMemory].DeepCopy()

	// 减去节点上未使用的内存资源量，得到已使用的内存资源量
	unusedMemory := node.Status.Capacity[v1.ResourceMemory]
	usedMemory.Sub(unusedMemory)

	return usedMemory
}

// 获取节点的存储使用量
func getUsedStorage(node *v1.Node) resource.Quantity {
	// 在这里根据你的调度器和集群环境获取节点的存储使用量
	// 这里只是示例，你需要根据实际情况获取节点的存储使用量

	// 假设我们获取节点的存储使用量
	// 从节点的状态中获取已分配的存储资源量
	usedStorage := node.Status.Allocatable[v1.ResourceStorage].DeepCopy()

	// 减去节点上未使用的存储资源量，得到已使用的存储资源量
	unusedStorage := node.Status.Capacity[v1.ResourceStorage]
	usedStorage.Sub(unusedStorage)

	return usedStorage
}


func NewCustomScheduling(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	// 注入配置文件参数
	args := &Args{}
	if err := frameworkruntime.DecodeInto(configuration, args); err != nil {
		return nil, err
	}

	return &CustomPlugin{handle: f}, nil
}
