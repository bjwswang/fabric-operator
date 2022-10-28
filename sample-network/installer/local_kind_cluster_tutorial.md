本文档介绍如何在`kind`集群上通过fabric operator 创建一个示例的 fabirc 区块链网络。
**准备工作参考 how-to-deploy-fabric-using-operator**

## 本地`kind`集群下的部署流程
1. 部署`kind`  
`./network kind`

**NOTE: 默认使用1.24.4 可通过`export KIND_CLUSTER_IMAGE=xxx修改`**

2. 配置环境变量  
`export TEST_NETWORK_INGRESS_DOMAIN=localho.st` 

3. 初始化集群  
```
./network cluster init
```
**NOTE:如果非首次初始化，可通过`export TEST_NETWORK_STAGE_DOCKER_IMAGES=false`跳过镜像pull/load**
4. 启动 operator，通过 kustomize 来部署网络所需的 CAs, peers, orderers
开始部署
```
./network up
```
4. 检查相关资源及状态，包括 pod、deployment、service、ingress等等
```
kubectl -n test-network get all
```
5. 确认服务状态正常
```
$ kubectl get pod -n test-network
NAME                                       READY   STATUS    RESTARTS   AGE
fabric-operator-5f675c54cc-hkqbq           1/1     Running   0          59m
hlf-console-5456764dd7-2dm7g               4/4     Running   1          3m47s
ingress-nginx-controller-f68f5f945-7k48m   1/1     Running   0          59m
org0-ca-7dd5c4c88b-mmn8b                   1/1     Running   0          58m
org0-orderersnode1-6dcd858cf8-r6gtx        2/2     Running   1          51m
org0-orderersnode2-69d5f47c88-rbt9h        2/2     Running   0          51m
org0-orderersnode3-6987c6548b-j4vhf        2/2     Running   1          51m
org1-ca-b9975f4c4-pqhgj                    1/1     Running   0          58m
org1-peer1-5ccdf7f99b-bfk8c                2/2     Running   0          49m
org1-peer2-69cc68d7bd-lqprn                2/2     Running   0          48m
org2-ca-556c6f4c7d-ngpw4                   1/1     Running   0          58m
org2-peer1-775696ccb7-qdc8w                2/2     Running   0          48m
org2-peer2-5c8f857677-9cw9b                2/2     Running   0          48m
```
6. 创建新的channel(网络测试) 
`./network channel create` 
**默认创建 mychannel** 

7. 部署智能合约(网络测试) 
1) 下载`fabric-samples` 

`git clone https://github.com/hyperledger/fabric-samples.git /tmp/fabric-samples`
2) 部署`asset-transfer-basic`

`network cc deploy   asset-transfer-basic basic_1.0 /tmp/fabric-samples/asset-transfer-basic/chaincode-java`
默认使用ccaas（External chaincode builder. 有关`external builder`可参考
- [External Builders](https://hyperledger-fabric.readthedocs.io/en/latest/cc_launcher.html) 
- [Running Chaincode as an external service](https://hyperledger-fabric.readthedocs.io/en/latest/cc_service.html)


## 可能碰到的问题
1. 支持私有镜像仓库
    a. 如果使用docker，则需配置`insecure-registries`。 Macos Desktop，需在`Preferences-DockerEngin`中
```
{
  "builder": {
    "gc": {
      "defaultKeepStorage": "20GB",
      "enabled": true
    }
  },
  "experimental": false,
  "features": {
    "buildkit": true
  },
  "insecure-registries": [
    "172.22.50.223"
  ]
}
```

    b. 如果是在`kind`单节点k8s服务，需配置`/etc/containerd/config.toml`。
需配置：

```
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."172.22.50.223"]
          endpoint = ["http://172.22.50.223"]
```

完整配置如下：
```
version = 2

[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    restrict_oom_score_adj = false
    sandbox_image = "registry.k8s.io/pause:3.7"
    tolerate_missing_hugepages_controller = true
    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      discard_unpacked_layers = true
      snapshotter = "overlayfs"
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          base_runtime_spec = "/etc/containerd/cri-base.json"
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
            SystemdCgroup = true
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler]
          base_runtime_spec = "/etc/containerd/cri-base.json"
          runtime_type = "io.containerd.runc.v2"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler.options]
            SystemdCgroup = true
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."k8s.gcr.io"]
          endpoint = ["https://registry.k8s.io", "https://k8s.gcr.io"]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
          endpoint = ["http://kind-registry:5000"]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."172.22.50.223"]
          endpoint = ["http://172.22.50.223"]

[proxy_plugins]
  [proxy_plugins.fuse-overlayfs]
    address = "/run/containerd-fuse-overlayfs.sock"
    type = "snapshot"
```

    c. docker配置私有镜像仓库身份认证
        `docker login 172.22.50.223 -u xxx -p xxx` 
    d. kind | k8s 配置私有镜像仓库身份认证
    i. 创建dockerconfigjson secret
```
kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```
    ii. Pod中配置`imagePullSecrets`
```
apiVersion: v1
kind: Pod
metadata:
  name: private-reg
spec:
  containers:
  - name: private-reg-container
    image: <your-private-image>
  imagePullSecrets:
  - name: regcred
```

2. network up问题
    a. `PodSecurityPolicy`问题(已在代码中解决)
    根源: k8s 1.24 -> k8s 1.25 升级后PodSecurityPolicy 
    社区讨论： `https://github.com/hyperledger-labs/fabric-operator/issues/63`
