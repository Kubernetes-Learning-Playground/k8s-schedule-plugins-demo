## k8s自定义调度插件



### 项目目录
```bigquery
.
├── main.go
├── src     # 插件逻辑
│   └── test-pod-maxNum-scheduler.go
├── test    # 测试pod yaml
│   └── testdep.yaml
└── yaml # 部署插件用
    ├── scheduling-deployment.yaml
    └── test-scheduling.yaml
```