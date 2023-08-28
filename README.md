## k8s自定义调度插件
### 项目思路与功能


### 项目部署
1. 编译应用程序
```bigquery
docker run --rm -it -v /root/k8s-schedule-practice:/app -w /app -e GOPROXY=https://goproxy.cn -e CGO_ENABLED=0  golang:1.18.7-alpine3.15 go build -o ./test-pod-maxNum-scheduler .
```
2. 修改与适配yaml(请依照自己环境适配)
```bigquery
[root@VM-0-16-centos yaml]# cd ..
[root@VM-0-16-centos k8s-schedule-practice]# cd yaml
[root@VM-0-16-centos yaml]# kubectl apply -f .
clusterrole.rbac.authorization.k8s.io/test-scheduling-clusterrole unchanged
serviceaccount/test-scheduling-sa unchanged
clusterrolebinding.rbac.authorization.k8s.io/test-scheduling-clusterrolebinding unchanged
configmap/test-scheduling-config unchanged
deployment.apps/test-pod-maxnum-scheduler unchanged
deployment.apps/testngx unchanged
```

3. 查看是否部署成功
```bigquery
[root@VM-0-16-centos yaml]# kubectl get pods -nkube-system | grep test-pod-maxnum-scheduler-d4cdb7794-5jrkm
test-pod-maxnum-scheduler-d4cdb7794-5jrkm   1/1     Running   0          6h9m
```
4. 部署测试
```bigquery
[root@VM-0-16-centos yaml]# kubectl get pods
NAME                                   READY   STATUS    RESTARTS   AGE
appdeployer-sample-55d9544448-hp7f5    1/1     Running   0          2d2h
rediscluster-sample-6997f8d4d5-h2nj7   1/1     Running   0          20h
testngx-55bc49f85f-6qbnx               1/1     Running   0          6h9m
testngx-55bc49f85f-cfwhl               0/1     Pending   0          6h7m
testngx-55bc49f85f-hxjl2               0/1     Pending   0          6h7m
testngx-55bc49f85f-jzbqd               0/1     Pending   0          6h7m
testngx-55bc49f85f-mrg2m               0/1     Pending   0          6h7m
testngx-55bc49f85f-spt7p               0/1     Pending   0          6h7m
testngx-55bc49f85f-tngcw               0/1     Pending   0          6h7m
testngx-55bc49f85f-vkbcv               0/1     Pending   0          6h7m
testngx-55bc49f85f-wg2zk               1/1     Running   0          6h7m
testngx-55bc49f85f-wvp9f               0/1     Pending   0          6h7m
usermetrics-5ddb966ffd-gdtfj           1/1     Running   0          6d20h
```
5. 查看组件日志
```bigquery
[root@VM-0-16-centos yaml]# kubectl logs -f test-pod-maxnum-scheduler-d4cdb7794-5jrkm -nkube-system
I1202 10:00:32.807396       1 scheduler.go:516] "Attempting to schedule pod" pod="default/testngx-55bc49f85f-hxjl2"
I1202 10:00:32.807413       1 test-pod-maxNum-scheduler.go:35] 当前被prefilter 的POD名称是:testngx-55bc49f85f-hxjl2
W1202 10:00:32.807421       1 listers.go:79] can not retrieve list of objects using index : Index with name namespace does not exist
I1202 10:00:32.807436       1 test-pod-maxNum-scheduler.go:43] POD数量超过可调度数量，不能调度！%!(EXTRA string=testngx-55bc49f85f-hxjl2)
I1202 10:00:32.807509       1 factory.go:381] "Unable to schedule pod; no fit; waiting" pod="default/testngx-55bc49f85f-hxjl2" err="0/1 nodes are available: 1 POD数量超过，不能调度，最多只能调度6."
I1202 10:00:32.807546       1 scheduler.go:414] "Updating pod condition" pod="default/testngx-55bc49f85f-hxjl2" conditionType=PodScheduled conditionStatus=False conditionReason="Unschedulable"
I1202 10:00:32.807570       1 scheduler.go:516] "Attempting to schedule pod" pod="default/testngx-55bc49f85f-spt7p"
I1202 10:00:32.807585       1 test-pod-maxNum-scheduler.go:35] 当前被prefilter 的POD名称是:testngx-55bc49f85f-spt7p
W1202 10:00:32.807594       1 listers.go:79] can not retrieve list of objects using index : Index with name namespace does not exist
I1202 10:00:32.807607       1 test-pod-maxNum-scheduler.go:43] POD数量超过可调度数量，不能调度！%!(EXTRA string=testngx-55bc49f85f-spt7p)
I1202 10:00:32.807669       1 factory.go:381] "Unable to schedule pod; no fit; waiting" pod="default/testngx-55bc49f85f-spt7p" err="0/1 nodes are available: 1 POD数量超过，不能调度，"
I1202 10:00:32.807702       1 scheduler.go:414] "Updating pod condition" pod="default/testngx-55bc49f85f-spt7p" conditionType=PodScheduled conditionStatus=False conditionReason="Unschedulable"
```
6. 查看测试pod
```bigquery
[root@VM-0-16-centos yaml]# kubectl describe pods testngx-55bc49f85f-wvp9f
Events:
  Type     Reason            Age   From                       Message
  ----     ------            ----  ----                       -------
  Warning  FailedScheduling  6h9m  test-pod-maxnum-scheduler  0/1 nodes are available: 1 POD数量超过，不能调度，最多只能调度6.
```

### 项目目录
```bigquery
├── main.go
├── src     # 插件逻辑
│   ├── test-pod-maxNum-scheduler.go # 
│   └── test-pod-score-scheduler.go
├── test    # 测试pod yaml
│   └── testdep.yaml
└── yaml # 部署插件用
    ├── scheduling-deployment.yaml
    └── test-scheduling.yaml
```

### 参考：
```bigquery
https://github.com/kubernetes-sigs/scheduler-plugins
```