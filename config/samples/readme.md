# 功能演示

## 准备工作

创立集群，部署 operator，以下内容在代码根目录操作。

<detail>
一些操作可能包括

```bash
# 创建kind集群
$ make kind
# 部署CRD
$ kubectl kustomize config/crd | kubectl apply -f -
# 创建admin-cluster-role
$ kubectl apply -f config/rbac/admin_cluster_role.yaml
```

</detail>

## 用户注册

### 1. 创建 5 个用户

```bash
kubectl apply -f config/samples/users
```

分别为:

- org1admin
- org2admin
- org3admin
- client
- client2

## 组织用户管理

### 1. 创建 3 个组织

```bash
kubectl apply -f config/samples/orgs
```

每个组织在 kubernetes 中对应一个 ns。

<details>

<summary>详细yaml为:</summary>

```bash
$ kubectl get ns
NAME                 STATUS   AGE
default              Active   16m
ingress-nginx        Active   16m
kube-node-lease      Active   16m
kube-public          Active   16m
kube-system          Active   16m
local-path-storage   Active   16m
org1                 Active   16m
org2                 Active   16m
org3                 Active   16m
```

</details>

同时对应一个 Namespaced 纬度的 CRD：Organization:

<details>

<summary>status 中会显示当前状态，如下:</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Organization
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"ibp.com/v1beta1","kind":"Organization","metadata":{"annotations":{},"name":"org1"},"spec":{"admin":"org1admin","caSpec":{"images":{"caImage":"hyperledgerk8s/fabric-ca","caInitImage":"hyperledgerk8s/ubi-minimal","caInitTag":"latest","caTag":"1.5.5-iam"},"license":{"accept":true},"resources":{"ca":{"limits":{"cpu":"100m","memory":"200M"},"requests":{"cpu":"10m","memory":"10M"}},"init":{"limits":{"cpu":"100m","memory":"200M"},"requests":{"cpu":"10m","memory":"10M"}}},"storage":{"ca":{"class":"standard","size":"100M"}},"version":"1.5.5"},"description":"test org1","displayName":"test organization","license":{"accept":true}}}
    creationTimestamp: "2022-12-20T07:01:32Z"
    generation: 1
    name: org1
    resourceVersion: "988707"
    uid: 4fe5089a-a03e-4133-a997-f9078fdcde5b
  spec:
    admin: org1admin
    clients:
    - client
    caSpec:
      images:
        caImage: hyperledgerk8s/fabric-ca
        caInitImage: hyperledgerk8s/ubi-minimal
        caInitTag: latest
        caTag: 1.5.5-iam
      license:
        accept: true
      resources:
        ca:
          limits:
            cpu: 100m
            memory: 200M
          requests:
            cpu: 10m
            memory: 10M
        init:
          limits:
            cpu: 100m
            memory: 200M
          requests:
            cpu: 10m
            memory: 10M
      storage:
        ca:
          class: standard
          size: 100M
      version: 1.5.5
    description: test org1
    displayName: test organization
    license:
      accept: true
  status:
    lastHeartbeatTime: 2022-12-20 15:02:13.497795 +0800 CST m=+455.849288746
    reason: allPodsDeployed
    status: "True"
    type: Deployed
kind: List
metadata:
  resourceVersion: ""
```

</details>

<details>
<summary>Organization对应IBPCA如下:</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: IBPCA
  metadata:
    creationTimestamp: "2022-12-20T07:01:32Z"
    generation: 2
    labels:
      app: org1
      app.kubernetes.io/instance: fabricorganization
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: org1
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Organization
      name: org1
      uid: 4fe5089a-a03e-4133-a997-f9078fdcde5b
    resourceVersion: "988705"
    uid: 0f2969e5-9734-4b36-b229-7d5a32fd8bb3
  spec:
    action:
      renew: {}
    configoverride:
      ca:
        ca: {}
        cfg:
          affiliations: {}
          identities: {}
        cors:
          enabled: null
          origins: null
        crl:
          expiry: 0s
        csr:
          cn: ""
        iam:
          enabled: true
          url: oidc.localho.st
        intermediate:
          enrollment:
            Type: ""
            name: ""
          parentserver: {}
          tls:
            client: {}
        ldap:
          attribute: {}
          tls:
            client: {}
        metrics: {}
        operations:
          metrics: {}
          tls: {}
        organization: org1
        registry: {}
        signing:
          default: null
          profiles: null
        tls:
          clientauth: {}
      tlsca:
        ca: {}
        cfg:
          affiliations: {}
          identities: {}
        cors:
          enabled: null
          origins: null
        crl:
          expiry: 0s
        csr:
          cn: ""
        iam:
          enabled: true
          url: oidc.localho.st
        intermediate:
          enrollment:
            Type: ""
            name: ""
          parentserver: {}
          tls:
            client: {}
        ldap:
          attribute: {}
          tls:
            client: {}
        metrics: {}
        operations:
          metrics: {}
          tls: {}
        organization: org1
        registry: {}
        signing:
          default: null
          profiles: null
        tls:
          clientauth: {}
    customNames:
      pvc: {}
    domain: localho.st
    images:
      caImage: hyperledgerk8s/fabric-ca
      caInitImage: hyperledgerk8s/ubi-minimal
      caInitTag: latest
      caTag: 1.5.5-iam
    ingress: {}
    license:
      accept: true
    replicas: 1
    resources:
      ca:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
      init:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
    storage:
      ca:
        class: standard
        size: 100M
    version: 1.5.5
  status:
    lastHeartbeatTime: 2022-12-20 15:02:13.453243 +0800 CST m=+455.804737690
    reason: allPodsDeployed
    status: "True"
    type: Deployed
    version: 1.0.0
    versions:
      reconciled: 1.5.5
kind: List
metadata:
  resourceVersion: ""
```

</details>

<details>

<summary>组织的Admin User:</summary>
总结:
- annotaion增加`bestchains.list.org1`
- label增加 `bestchains.organization.org1:admin`

```yaml
apiVersion: iam.tenxcloud.com/v1alpha1
kind: User
metadata:
  annotations:
    bestchains: '{"list":{"org1":{"organization":"org1","ids":{"org1admin":{"name":"org1admin","type":"client","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"org1admin","hf.Type":"client"},"creationTimestamp":"2022-12-30T03:31:30Z","lastAppliedTimestamp":"2022-12-30T03:31:30Z"}},"creationTimestamp":"2022-12-30T03:31:30Z","lastAppliedTimestamp":"2022-12-30T03:31:30Z","lastDeletionTimestamp":null}},"creationTimestamp":"2022-12-30T03:31:30Z","lastAppliedTimestamp":"2022-12-30T03:31:30Z","lastDeletetionTimestamp":null}'
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"iam.tenxcloud.com/v1alpha1","kind":"User","metadata":{"annotations":{},"labels":{"t7d.io.username":"org1admin"},"name":"org1admin"},"spec":{"description":"org1admin 用户信息的描述","email":"org1admin@tenxcloud.com","groups":["observability","system:nodes","system:masters","resource-reader","iam.tenxcloud.com","observability"],"name":"org1admin","password":"$2a$10$693K.zP98yCs1qVwEp//DuWYOtLIE1doihtGhcCyYh3IpgSdGGba2","phone":"18890901212","role":"admin"}}
  creationTimestamp: "2022-12-30T03:31:26Z"
  generation: 1
  labels:
    bestchains.organizaiton.org1: admin
    t7d.io.username: org1admin
  name: org1admin
  resourceVersion: "1254605"
  uid: 53b666a7-b6c1-42e1-8adb-5c98c3347f7e
spec:
  description: org1admin 用户信息的描述
  email: org1admin@tenxcloud.com
  groups:
  - observability
  - system:nodes
  - system:masters
  - resource-reader
  - iam.tenxcloud.com
  - observability
  name: org1admin
  password: $2a$10$693K.zP98yCs1qVwEp//DuWYOtLIE1doihtGhcCyYh3IpgSdGGba2
  phone: "18890901212"
  role: admin
```

</details>

<details>

<summary>组织的Client User:</summary>
总结:
- annotaion增加`bestchains.list.org1`
- label增加 `bestchains.organization.org1:client`

```yaml
apiVersion: iam.tenxcloud.com/v1alpha1
kind: User
metadata:
  annotations:
    bestchains: '{"list":{"org1":{"organization":"org1","ids":{"client":{"name":"client","type":"admin","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"client","hf.GenCRL":"true","hf.IntermediateCA":"true","hf.Registrar.Roles":"*","hf.RegistrarDelegateRoles":"*","hf.Revoker":"*","hf.Type":"admin","hf.hf.Registrar.Attributes":"*"},"creationTimestamp":"2022-12-30T03:40:17Z","lastAppliedTimestamp":"2022-12-30T03:40:17Z"}},"creationTimestamp":"2022-12-30T03:40:17Z","lastAppliedTimestamp":"2022-12-30T03:40:17Z","lastDeletionTimestamp":null},"org2":{"organization":"org2","ids":{"client":{"name":"client","type":"admin","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"client","hf.GenCRL":"true","hf.IntermediateCA":"true","hf.Registrar.Roles":"*","hf.RegistrarDelegateRoles":"*","hf.Revoker":"*","hf.Type":"admin","hf.hf.Registrar.Attributes":"*"},"creationTimestamp":"2022-12-30T03:40:17Z","lastAppliedTimestamp":"2022-12-30T03:40:17Z"}},"creationTimestamp":"2022-12-30T03:40:17Z","lastAppliedTimestamp":"2022-12-30T03:40:17Z","lastDeletionTimestamp":null},"org3":{"organization":"org3","ids":{"client":{"name":"client","type":"admin","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"client","hf.GenCRL":"true","hf.IntermediateCA":"true","hf.Registrar.Roles":"*","hf.RegistrarDelegateRoles":"*","hf.Revoker":"*","hf.Type":"admin","hf.hf.Registrar.Attributes":"*"},"creationTimestamp":"2022-12-30T03:40:22Z","lastAppliedTimestamp":"2022-12-30T03:40:22Z"}},"creationTimestamp":"2022-12-30T03:40:22Z","lastAppliedTimestamp":"2022-12-30T03:40:22Z","lastDeletionTimestamp":null}},"creationTimestamp":"2022-12-30T03:40:17Z","lastAppliedTimestamp":"2022-12-30T03:40:22Z","lastDeletetionTimestamp":null}'
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"iam.tenxcloud.com/v1alpha1","kind":"User","metadata":{"annotations":{},"labels":{"t7d.io.username":"client"},"name":"client"},"spec":{"description":"client 用户信息的描述","email":"client@tenxcloud.com","groups":["observability","system:nodes","system:masters","resource-reader","iam.tenxcloud.com","observability"],"name":"client","phone":"18890901212","role":"admin"}}
  creationTimestamp: "2022-12-30T03:39:54Z"
  generation: 1
  labels:
    bestchains.organizaiton.org1: client
    bestchains.organizaiton.org2: client
    bestchains.organizaiton.org3: client
    t7d.io.username: client
  name: client
  resourceVersion: "1257278"
  uid: db142b11-d5ed-463c-a2cc-348eaa77b889
spec:
  description: client 用户信息的描述
  email: client@tenxcloud.com
  groups:
  - observability
  - system:nodes
  - system:masters
  - resource-reader
  - iam.tenxcloud.com
  - observability
  name: client
  phone: "18890901212"
  role: admin
```

</details>

### 2. 更新组织的 client 用户

1. 新增 client 用户

通过更新 `spec.clients` 字段

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_organization_add_client.yaml
```

<details>

<summary>详细yaml</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Organization
metadata:
  name: org1
spec:
  license:
    accept: true
  displayName: "test organization"
  admin: org1admin
  clients:
    - client
    - client2
  description: "test org1"
  caSpec:
    license:
      accept: true
    images:
      caImage: hyperledgerk8s/fabric-ca
      caTag: "1.5.5-iam"
      caInitImage: hyperledgerk8s/ubi-minimal
      caInitTag: latest
    resources:
      ca:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
      init:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
    storage:
      ca:
        class: "standard"
        size: 100M
    version: 1.5.5
```

</details>

<details>
<summary>新client用户详情</summary>
总结:
- annotation增加`bestchains.list.org1`
- labels增加`bestchains.organization.org1:client`

```yaml
apiVersion: iam.tenxcloud.com/v1alpha1
kind: User
metadata:
  annotations:
    bestchains: '{"list":{"org1":{"organization":"org1","ids":{"client2":{"name":"client2","type":"admin","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"client2","hf.GenCRL":"true","hf.IntermediateCA":"true","hf.Registrar.Roles":"*","hf.RegistrarDelegateRoles":"*","hf.Revoker":"*","hf.Type":"admin","hf.hf.Registrar.Attributes":"*"},"creationTimestamp":"2022-12-30T03:32:16Z","lastAppliedTimestamp":"2022-12-30T03:32:16Z"}},"creationTimestamp":"2022-12-30T03:32:16Z","lastAppliedTimestamp":"2022-12-30T03:32:16Z","lastDeletionTimestamp":null}},"creationTimestamp":"2022-12-30T03:32:16Z","lastAppliedTimestamp":"2022-12-30T03:32:16Z","lastDeletetionTimestamp":null}'
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"iam.tenxcloud.com/v1alpha1","kind":"User","metadata":{"annotations":{},"labels":{"t7d.io.username":"client2"},"name":"client2"},"spec":{"description":"client2 用户信息的描述","email":"client2@tenxcloud.com","groups":["observability","system:nodes","system:masters","resource-reader","iam.tenxcloud.com","observability"],"name":"client2","phone":"18890901212","role":"admin"}}
  creationTimestamp: "2022-12-30T03:31:25Z"
  generation: 1
  labels:
    bestchains.organizaiton.org1: client
    t7d.io.username: client2
  name: client2
  resourceVersion: "1255034"
  uid: c2d78fd6-b2f3-4802-86ba-680587232169
spec:
  description: client2 用户信息的描述
  email: client2@tenxcloud.com
  groups:
  - observability
  - system:nodes
  - system:masters
  - resource-reader
  - iam.tenxcloud.com
  - observability
  name: client2
  phone: "18890901212"
  role: admin
```

</details>

2. 从组织中删除 client 用户

通过更新 `spec.clients` 字段

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_organization_remove_client.yaml
```

<details>

<summary>详细yaml</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Organization
metadata:
  name: org1
spec:
  license:
    accept: true
  displayName: "test organization"
  admin: org1admin
  clients:
    - client2
  description: "test org1"
  caSpec:
    license:
      accept: true
    images:
      caImage: hyperledgerk8s/fabric-ca
      caTag: "1.5.5-iam"
      caInitImage: hyperledgerk8s/ubi-minimal
      caInitTag: latest
    resources:
      ca:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
      init:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
    storage:
      ca:
        class: "standard"
        size: 100M
    version: 1.5.5
```

</details>

<details>
<summary>删除的client用户的详情</summary>
总结:
- annotation删除了`bestchains.list.org1`
- labels删除了`bestchains.organization.org1:client`

```yaml
apiVersion: iam.tenxcloud.com/v1alpha1
kind: User
metadata:
  annotations:
    bestchains: '{"list":{"org2":{"organization":"org2","ids":{"client":{"name":"client","type":"admin","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"client","hf.GenCRL":"true","hf.IntermediateCA":"true","hf.Registrar.Roles":"*","hf.RegistrarDelegateRoles":"*","hf.Revoker":"*","hf.Type":"admin","hf.hf.Registrar.Attributes":"*"},"creationTimestamp":"2022-12-30T03:31:35Z","lastAppliedTimestamp":"2022-12-30T03:31:35Z"}},"creationTimestamp":"2022-12-30T03:31:35Z","lastAppliedTimestamp":"2022-12-30T03:31:35Z","lastDeletionTimestamp":null},"org3":{"organization":"org3","ids":{"client":{"name":"client","type":"admin","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"client","hf.GenCRL":"true","hf.IntermediateCA":"true","hf.Registrar.Roles":"*","hf.RegistrarDelegateRoles":"*","hf.Revoker":"*","hf.Type":"admin","hf.hf.Registrar.Attributes":"*"},"creationTimestamp":"2022-12-30T03:31:39Z","lastAppliedTimestamp":"2022-12-30T03:31:39Z"}},"creationTimestamp":"2022-12-30T03:31:39Z","lastAppliedTimestamp":"2022-12-30T03:31:39Z","lastDeletionTimestamp":null}},"creationTimestamp":"2022-12-30T03:31:30Z","lastAppliedTimestamp":"2022-12-30T03:32:16Z","lastDeletetionTimestamp":"2022-12-30T03:33:21Z"}'
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"iam.tenxcloud.com/v1alpha1","kind":"User","metadata":{"annotations":{},"labels":{"t7d.io.username":"client"},"name":"client"},"spec":{"description":"client 用户信息的描述","email":"client@tenxcloud.com","groups":["observability","system:nodes","system:masters","resource-reader","iam.tenxcloud.com","observability"],"name":"client","phone":"18890901212","role":"admin"}}
  creationTimestamp: "2022-12-30T03:31:25Z"
  generation: 1
  labels:
    bestchains.organizaiton.org2: client
    bestchains.organizaiton.org3: client
    t7d.io.username: client
  name: client
  resourceVersion: "1255296"
  uid: f094a5f3-a350-49d9-8a12-d3bfae566682
spec:
  description: client 用户信息的描述
  email: client@tenxcloud.com
  groups:
  - observability
  - system:nodes
  - system:masters
  - resource-reader
  - iam.tenxcloud.com
  - observability
  name: client
  phone: "18890901212"
  role: admin
```

</details>

### 3.  转移组织 Admin 权限

通过更新 `spec.admin` 字段

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_organization_transfer_admin.yaml
```

<details>

<summary>详细yaml</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Organization
metadata:
  name: org1
spec:
  license:
    accept: true
  displayName: "test organization"
  admin: org2admin
  clients:
    - client
  description: "test org1"
  caSpec:
    license:
      accept: true
    images:
      caImage: hyperledgerk8s/fabric-ca
      caTag: "1.5.5-iam"
      caInitImage: hyperledgerk8s/ubi-minimal
      caInitTag: latest
    resources:
      ca:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
      init:
        limits:
          cpu: 100m
          memory: 200M
        requests:
          cpu: 10m
          memory: 10M
    storage:
      ca:
        class: "standard"
        size: 100M
    version: 1.5.5

```

</details>

<details>
<summary>原Admin用户</summary>

总结：

- annotations 去除了 `bestchains.list.org1`
- labels 去除了 `bestchains.organization.org1:admin`

```yaml
apiVersion: iam.tenxcloud.com/v1alpha1
kind: User
metadata:
  annotations:
    bestchains: '{"creationTimestamp":"2022-12-30T03:56:47Z","lastAppliedTimestamp":"2022-12-30T03:56:47Z","lastDeletetionTimestamp":"2022-12-30T04:03:50Z"}'
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"iam.tenxcloud.com/v1alpha1","kind":"User","metadata":{"annotations":{},"labels":{"t7d.io.username":"org1admin"},"name":"org1admin"},"spec":{"description":"org1admin 用户信息的描述","email":"org1admin@tenxcloud.com","groups":["observability","system:nodes","system:masters","resource-reader","iam.tenxcloud.com","observability"],"name":"org1admin","password":"$2a$10$693K.zP98yCs1qVwEp//DuWYOtLIE1doihtGhcCyYh3IpgSdGGba2","phone":"18890901212","role":"admin"}}
  creationTimestamp: "2022-12-30T03:56:36Z"
  generation: 1
  labels:
    t7d.io.username: org1admin
  name: org1admin
  resourceVersion: "1264655"
  uid: fb955b00-494b-4545-bd53-01ead52520f3
spec:
  description: org1admin 用户信息的描述
  email: org1admin@tenxcloud.com
  groups:
  - observability
  - system:nodes
  - system:masters
  - resource-reader
  - iam.tenxcloud.com
  - observability
  name: org1admin
  password: $2a$10$693K.zP98yCs1qVwEp//DuWYOtLIE1doihtGhcCyYh3IpgSdGGba2
  phone: "18890901212"
  role: admin
```

</details>

<details>
<summary>新Admin用户</summary>

总结：

- annotations 增加了 `bestchains.list.org1`
- labels 增加了 `bestchains.organization.org1:admin`

```yaml
apiVersion: iam.tenxcloud.com/v1alpha1
kind: User
metadata:
  annotations:
    bestchains: '{"list":{"org1":{"organization":"org1","ids":{"org2admin":{"name":"org2admin","type":"admin","attributes":{"hf.Affiliation":"","hf.EnrollmentID":"org2admin","hf.GenCRL":"true","hf.IntermediateCA":"true","hf.Registrar.Roles":"*","hf.RegistrarDelegateRoles":"*","hf.Revoker":"*","hf.Type":"admin","hf.hf.Registrar.Attributes":"*"},"creationTimestamp":"2022-12-30T04:03:50Z","lastAppliedTimestamp":"2022-12-30T04:03:50Z"}},"creationTimestamp":"2022-12-30T03:56:47Z","lastAppliedTimestamp":"2022-12-30T04:03:50Z","lastDeletionTimestamp":"2022-12-30T04:03:50Z"}},"creationTimestamp":"2022-12-30T04:03:50Z","lastAppliedTimestamp":"2022-12-30T04:03:50Z","lastDeletetionTimestamp":null}'
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"iam.tenxcloud.com/v1alpha1","kind":"User","metadata":{"annotations":{},"labels":{"t7d.io.username":"org2admin"},"name":"org2admin"},"spec":{"description":"org2admin 用户信息的描述","email":"org2admin@tenxcloud.com","groups":["observability","system:nodes","system:masters","resource-reader","iam.tenxcloud.com","observability"],"name":"org2admin","password":"$2a$10$693K.zP98yCs1qVwEp//DuWYOtLIE1doihtGhcCyYh3IpgSdGGba2","phone":"18890901212","role":"admin"}}
  creationTimestamp: "2022-12-30T03:56:36Z"
  generation: 1
  labels:
    bestchains.organizaiton.org1: admin
    bestchains.organizaiton.org2: admin
    t7d.io.username: org2admin
  name: org2admin
  resourceVersion: "1264656"
  uid: e80da140-bb9c-4a8d-a271-edd81971fbdf
spec:
  description: org2admin 用户信息的描述
  email: org2admin@tenxcloud.com
  groups:
  - observability
  - system:nodes
  - system:masters
  - resource-reader
  - iam.tenxcloud.com
  - observability
  name: org2admin
  password: $2a$10$693K.zP98yCs1qVwEp//DuWYOtLIE1doihtGhcCyYh3IpgSdGGba2
  phone: "18890901212"
  role: admin
```

</details>

## 组织和联盟管理

### 1. 创建联盟

某个组织要想创建一个联盟 (Federation)，需要按以下步骤：

1. 创建一个联盟 (Federation) 资源。
2. 创建一个提案（Proposal），类型为创建联盟（createFederation），这会触发在每个可以投票的组织中自动创建 选票（vote）资源。
3. 等待每个可以投票的组织手动修改选票（Vote），即投票，表明自己的意愿，同意或者拒绝。
4. 投票符合提案（Proposal）中规定的策略（spec.Policy）后，成功创建联盟。

#### 1.1 创建联盟 federation 的 CRD

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_federation.yaml
```

<details>

<summary>详细yaml为:</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Federation
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Federation","metadata":{"annotations":{},"name":"federation-sample","namespace":"org1"},"spec":{"description":"federation with org0 \u0026 org1","license":{"accept":true},"members":[{"initiator":true,"name":"org1","namespace":"org1"},{"name":"org2","namespace":"org2"}],"policy":"ALL"}}
  creationTimestamp: "2022-12-08T04:14:40Z"
  generation: 1
  name: federation-sample
  namespace: org1
  resourceVersion: "1908"
  uid: 534fb03b-e11f-4f5c-be48-b57c4d71cf04
spec:
  description: federation with org0 & org1
  license:
    accept: true
  members:
  - initiator: true
    name: org1
    namespace: org1
  - name: org2
    namespace: org2
  policy: ALL
status:
  lastHeartbeatTime: 2022-12-08 12:14:41.010557 +0800 CST m=+101.483887881
  status: "True"
  type: FederationPending
```

</details>

#### 1.2 发起提案 proposal，内容为创建联盟

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_proposal_create_federation.yaml
```

<details>

<summary>一个正在投票中的proposal如下：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Proposal
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Proposal","metadata":{"annotations":{},"name":"create-federation-sample"},"spec":{"createFederation":{},"federation":{"name":"federation-sample","namespace":"org1"},"initiator":{"name":"org1","namespace":"org1"},"policy":"All"}}
  creationTimestamp: "2022-12-08T04:15:20Z"
  generation: 1
  name: create-federation-sample
  resourceVersion: "1972"
  uid: dced6beb-5632-459b-8d2d-6ef67a66b71b
spec:
  createFederation: {}
  deprecated: false
  federation:
    name: federation-sample
    namespace: org1
  initiator:
    name: org1
    namespace: org1
  policy: All
status:
  phase: Voting
  votes:
  - description: ""
    organization:
      name: org2
      namespace: org2
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2022-12-08T04:15:20Z"
```

</details>

#### 1.3 涉及到的组织投票

投票 Vote 是一个 Namespaced 纬度的 CRD。

controller 会在每个有投票权的 Organization 下创建对应的 Vote，本示例中，会在 org1 和 org2 的 ns 下创建 Vote，因为 org1 是本提案的发起人，因此会自动将 org1 中的 Vote 置为 同意，等待 org2 投票:

<details>

<summary>此时全部的vote如下:</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T04:15:20Z"
    generation: 2
    labels:
      app: create-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-create-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: create-federation-sample
      uid: dced6beb-5632-459b-8d2d-6ef67a66b71b
    resourceVersion: "1971"
    uid: 08e1a28a-d83a-4f2e-9dd1-6b79434c7719
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: create-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T04:15:20Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T04:15:20Z"
    generation: 1
    labels:
      app: create-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org2-create-federation-sample
    namespace: org2
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: create-federation-sample
      uid: dced6beb-5632-459b-8d2d-6ef67a66b71b
    resourceVersion: "1966"
    uid: 45f1a332-1e9e-4ce3-a95a-139e3a302ef8
  spec:
    description: ""
    organizationName: org2
    proposalName: create-federation-sample
  status:
    phase: Created
kind: List
metadata:
  resourceVersion: ""
```

</details>

org2 投同意票可以用以下命令：

```bash
kubectl patch vote -n org2 vote-org2-create-federation-sample --type='json' -p='[{"op": "replace", "path": "/spec/decision", "value": true}]'
```

<details>

<summary>此时vote更新为：</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T04:15:20Z"
    generation: 2
    labels:
      app: create-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-create-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: create-federation-sample
      uid: dced6beb-5632-459b-8d2d-6ef67a66b71b
    resourceVersion: "1971"
    uid: 08e1a28a-d83a-4f2e-9dd1-6b79434c7719
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: create-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T04:15:20Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T04:15:20Z"
    generation: 2
    labels:
      app: create-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org2-create-federation-sample
    namespace: org2
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: create-federation-sample
      uid: dced6beb-5632-459b-8d2d-6ef67a66b71b
    resourceVersion: "2194"
    uid: 45f1a332-1e9e-4ce3-a95a-139e3a302ef8
  spec:
    decision: true
    description: ""
    organizationName: org2
    proposalName: create-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T04:18:25Z"
kind: List
metadata:
  resourceVersion: ""
```

</details>

#### 1.4 投票成功，联盟创立

因为 所有组织都同意了创建联盟的提案，proposal 的状态更新为 `Finished`:

```yaml
status:
  conditions:
  - lastTransitionTime: "2022-12-08T04:18:25Z"
    message: Success
    reason: Success
    status: "True"
    type: Succeeded
  phase: Finished
  votes:
  - decision: true
    description: ""
    organization:
      name: org2
      namespace: org2
    phase: Voted
    startTime: "2022-12-08T04:18:25Z"
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2022-12-08T04:15:20Z"
```

<details>

<summary>完整proposal yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Proposal
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Proposal","metadata":{"annotations":{},"name":"create-federation-sample"},"spec":{"createFederation":{},"federation":{"name":"federation-sample","namespace":"org1"},"initiator":{"name":"org1","namespace":"org1"},"policy":"All"}}
  creationTimestamp: "2022-12-08T04:15:20Z"
  generation: 1
  name: create-federation-sample
  resourceVersion: "2196"
  uid: dced6beb-5632-459b-8d2d-6ef67a66b71b
spec:
  createFederation: {}
  deprecated: false
  federation:
    name: federation-sample
    namespace: org1
  initiator:
    name: org1
    namespace: org1
  policy: All
status:
  conditions:
  - lastTransitionTime: "2022-12-08T04:18:25Z"
    message: Success
    reason: Success
    status: "True"
    type: Succeeded
  phase: Finished
  votes:
  - decision: true
    description: ""
    organization:
      name: org2
      namespace: org2
    phase: Voted
    startTime: "2022-12-08T04:18:25Z"
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2022-12-08T04:15:20Z"
```

</details>

同时，联盟 federation 的状态也会同步更新为 `FederationActivated`：

```yaml
status:
  lastHeartbeatTime: 2022-12-08 12:35:44.826345 +0800 CST m=+33.724520930
  status: "True"
  type: FederationActivated
```

<details>

<summary>完整proposal yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Federation
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Federation","metadata":{"annotations":{},"name":"federation-sample","namespace":"org1"},"spec":{"description":"federation with org0 \u0026 org1","license":{"accept":true},"members":[{"initiator":true,"name":"org1","namespace":"org1"},{"name":"org2","namespace":"org2"}],"policy":"ALL"}}
  creationTimestamp: "2022-12-08T04:14:40Z"
  generation: 1
  name: federation-sample
  namespace: org1
  resourceVersion: "3510"
  uid: 534fb03b-e11f-4f5c-be48-b57c4d71cf04
spec:
  description: federation with org0 & org1
  license:
    accept: true
  members:
    - initiator: true
      name: org1
      namespace: org1
    - name: org2
      namespace: org2
  policy: ALL
status:
  lastHeartbeatTime: 2022-12-08 12:35:44.826345 +0800 CST m=+33.724520930
  status: "True"
  type: FederationActivated
```

</details>

至此，联盟创立成功。

### 2. 向一个联盟中添加组织

某个组织要在现有的联盟中添加一个组织作为成员，需要按以下步骤：

1. 创建一个提案（Proposal），类型为添加成员（AddMember），这会触发在每个可以投票的组织中自动创建 选票（vote）资源。
2. 等待每个可以投票的组织手动修改选票（Vote），即投票，表明自己的意愿，同意或者拒绝。
3. 投票符合提案（Proposal）中规定的策略（spec.Policy）后，成功添加该组织作为成员。

#### 2.1 发起提案 proposal，内容为添加成员

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_proposal_add_member.yaml
```

<details>

<summary>完整proposal yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Proposal
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Proposal","metadata":{"annotations":{},"name":"add-member-federation-sample"},"spec":{"addMember":{"members":[{"name":"org3","namespace":"org3"}]},"federation":{"name":"federation-sample","namespace":"org1"},"initiator":{"name":"org1","namespace":"org1"},"policy":"All"}}
  creationTimestamp: "2022-12-08T05:21:33Z"
  generation: 2
  name: add-member-federation-sample
  resourceVersion: "6911"
  uid: 51ccea30-88f4-4860-8008-9f4d3ec306df
spec:
  addMember:
    members:
    - name: org3
      namespace: org3
  deprecated: false
  endAt: "2022-12-09T05:21:33Z"
  federation:
    name: federation-sample
    namespace: org1
  initiator:
    name: org1
    namespace: org1
  policy: All
  startAt: "2022-12-08T05:21:33Z"
status:
  phase: Voting
  votes:
  - description: ""
    organization:
      name: org3
      namespace: org3
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2022-12-08T05:21:34Z"
  - description: ""
    organization:
      name: org2
      namespace: org2
```

</details>

<details>

<summary>此时选票结果为：</summary>

```bash
$kubectl get vote -A
NAMESPACE   NAME                                     AGE
org1        vote-org1-add-member-federation-sample   67s
org1        vote-org1-create-federation-sample       16m
org2        vote-org2-add-member-federation-sample   67s
org2        vote-org2-create-federation-sample       16m
org3        vote-org3-add-member-federation-sample   67s
```

</details>

#### 2.2 涉及到的组织投票

```bash
kubectl patch vote -n org2 vote-org2-add-member-federation-sample --type='json' -p='[{"op": "replace", "path": "/spec/decision", "value": true}]'
kubectl patch vote -n org3 vote-org3-add-member-federation-sample --type='json' -p='[{"op": "replace", "path": "/spec/decision", "value": true}]'
```

<details>
<summary>此时选票结果为：</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:21:33Z"
    generation: 2
    labels:
      app: add-member-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-add-member-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: add-member-federation-sample
      uid: 51ccea30-88f4-4860-8008-9f4d3ec306df
    resourceVersion: "6910"
    uid: 1cf0499c-235a-4193-abfb-3054a41c62f0
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: add-member-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:21:34Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:06:10Z"
    generation: 2
    labels:
      app: create-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-create-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: create-federation-sample
      uid: 76060189-0827-4143-96a2-6c988b1a100d
    resourceVersion: "5775"
    uid: a6bcb00f-12ef-4380-8fcd-3e085306bc24
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: create-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:06:10Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:21:33Z"
    generation: 2
    labels:
      app: add-member-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org2-add-member-federation-sample
    namespace: org2
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: add-member-federation-sample
      uid: 51ccea30-88f4-4860-8008-9f4d3ec306df
    resourceVersion: "7164"
    uid: 437f9bd6-5f6c-4079-878f-d271a1a65e9d
  spec:
    decision: true
    description: ""
    organizationName: org2
    proposalName: add-member-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:25:02Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:06:10Z"
    generation: 2
    labels:
      app: create-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org2-create-federation-sample
    namespace: org2
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: create-federation-sample
      uid: 76060189-0827-4143-96a2-6c988b1a100d
    resourceVersion: "5793"
    uid: ba193749-2ba3-4d89-9bfb-6fcd466c7e9c
  spec:
    decision: true
    description: ""
    organizationName: org2
    proposalName: create-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:06:21Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:21:33Z"
    generation: 2
    labels:
      app: add-member-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org3-add-member-federation-sample
    namespace: org3
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: add-member-federation-sample
      uid: 51ccea30-88f4-4860-8008-9f4d3ec306df
    resourceVersion: "7167"
    uid: 0eda44c9-57cd-4b3c-ace5-c274366b77d0
  spec:
    decision: true
    description: ""
    organizationName: org3
    proposalName: add-member-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:25:03Z"
kind: List
metadata:
  resourceVersion: ""
```

</details>

<details>
<summary>此时提案结果为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Proposal
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Proposal","metadata":{"annotations":{},"name":"add-member-federation-sample"},"spec":{"addMember":{"members":[{"name":"org3","namespace":"org3"}]},"federation":{"name":"federation-sample","namespace":"org1"},"initiator":{"name":"org1","namespace":"org1"},"policy":"All"}}
  creationTimestamp: "2022-12-08T05:21:33Z"
  generation: 2
  name: add-member-federation-sample
  resourceVersion: "7169"
  uid: 51ccea30-88f4-4860-8008-9f4d3ec306df
spec:
  addMember:
    members:
    - name: org3
      namespace: org3
  deprecated: false
  endAt: "2022-12-09T05:21:33Z"
  federation:
    name: federation-sample
    namespace: org1
  initiator:
    name: org1
    namespace: org1
  policy: All
  startAt: "2022-12-08T05:21:33Z"
status:
  conditions:
  - lastTransitionTime: "2022-12-08T05:25:03Z"
    message: Success
    reason: Success
    status: "True"
    type: Succeeded
  phase: Finished
  votes:
  - decision: true
    description: ""
    organization:
      name: org3
      namespace: org3
    phase: Voted
    startTime: "2022-12-08T05:25:03Z"
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2022-12-08T05:21:34Z"
  - decision: true
    description: ""
    organization:
      name: org2
      namespace: org2
    phase: Voted
    startTime: "2022-12-08T05:25:02Z"
```

</details>

#### 2.3 投票成功，成员成功添加

<details>

<summary>联盟yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Federation
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Federation","metadata":{"annotations":{},"name":"federation-sample","namespace":"org1"},"spec":{"description":"federation with org0 \u0026 org1","license":{"accept":true},"members":[{"initiator":true,"name":"org1","namespace":"org1"},{"name":"org2","namespace":"org2"}],"policy":"ALL"}}
  creationTimestamp: "2022-12-08T04:14:40Z"
  generation: 2
  name: federation-sample
  namespace: org1
  resourceVersion: "8794"
  uid: 534fb03b-e11f-4f5c-be48-b57c4d71cf04
spec:
  description: federation with org0 & org1
  license:
    accept: true
  members:
  - initiator: true
    name: org1
    namespace: org1
  - name: org2
    namespace: org2
  - name: org3
    namespace: org3
  policy: ALL
status:
  lastHeartbeatTime: 2022-12-08 12:35:44.826345 +0800 CST m=+33.724520930
  status: "True"
  type: FederationActivated
```

</details>

### 3. 从一个联盟中驱逐一个组织

某个组织要在现有的联盟中驱逐一个组织，需要按以下步骤：

1. 创建一个提案（Proposal），类型为驱逐成员（deleteMember），这会触发在每个可以投票的组织中自动创建 选票（vote）资源。
2. 等待每个可以投票的组织手动修改选票（Vote），即投票，表明自己的意愿，同意或者拒绝。
3. 投票符合提案（Proposal）中规定的策略（spec.Policy）后，成功添加该组织作为成员。

#### 3.1 发起提案 proposal，内容为驱逐成员

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_proposal_delete_member.yaml
```

<details>

<summary>完整proposal yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Proposal
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Proposal","metadata":{"annotations":{},"name":"delete-member-federation-sample"},"spec":{"deleteMember":{"member":{"name":"org2","namespace":"org2"}},"federation":{"name":"federation-sample","namespace":"org1"},"initiator":{"name":"org1","namespace":"org1"},"policy":"All"}}
  creationTimestamp: "2022-12-08T05:49:08Z"
  generation: 2
  name: delete-member-federation-sample
  resourceVersion: "8967"
  uid: 02ef7287-0857-471e-b2d5-6ce4326c35f4
spec:
  deleteMember:
    member:
      name: org2
      namespace: org2
  deprecated: false
  endAt: "2022-12-09T05:49:08Z"
  federation:
    name: federation-sample
    namespace: org1
  initiator:
    name: org1
    namespace: org1
  policy: All
  startAt: "2022-12-08T05:49:08Z"
status:
  phase: Voting
  votes:
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2022-12-08T05:49:08Z"
  - description: ""
    organization:
      name: org3
      namespace: org3
```

</details>

当驱逐 org2 时，org2 不会拥有选票 Vote:

<details>

<summary>此时选票结果为：</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:49:08Z"
    generation: 2
    labels:
      app: delete-member-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-delete-member-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: delete-member-federation-sample
      uid: 02ef7287-0857-471e-b2d5-6ce4326c35f4
    resourceVersion: "8966"
    uid: 03235fda-d032-4125-8f99-6ea94e92b93c
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: delete-member-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:49:08Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:49:08Z"
    generation: 1
    labels:
      app: delete-member-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org3-delete-member-federation-sample
    namespace: org3
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: delete-member-federation-sample
      uid: 02ef7287-0857-471e-b2d5-6ce4326c35f4
    resourceVersion: "8965"
    uid: f1a05522-7ba2-4363-b992-12c29c9dbda1
  spec:
    description: ""
    organizationName: org3
    proposalName: delete-member-federation-sample
  status:
    phase: Created
kind: List
metadata:
  resourceVersion: ""
```

</details>

#### 3.2 涉及到的组织投票

```bash
kubectl patch vote -n org3 vote-org3-delete-member-federation-sample --type='json' -p='[{"op": "replace", "path": "/spec/decision", "value": true}]'
```

<details>

<summary>此时选票结果为：</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:49:08Z"
    generation: 2
    labels:
      app: delete-member-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-delete-member-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: delete-member-federation-sample
      uid: 02ef7287-0857-471e-b2d5-6ce4326c35f4
    resourceVersion: "8966"
    uid: 03235fda-d032-4125-8f99-6ea94e92b93c
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: delete-member-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:49:08Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:49:08Z"
    generation: 2
    labels:
      app: delete-member-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org3-delete-member-federation-sample
    namespace: org3
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: delete-member-federation-sample
      uid: 02ef7287-0857-471e-b2d5-6ce4326c35f4
    resourceVersion: "9170"
    uid: f1a05522-7ba2-4363-b992-12c29c9dbda1
  spec:
    decision: true
    description: ""
    organizationName: org3
    proposalName: delete-member-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:51:45Z"
kind: List
metadata:
  resourceVersion: ""
```

</details>

#### 3.3 投票成功，成员成功驱逐

<details>

<summary>联盟yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Federation
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Federation","metadata":{"annotations":{},"name":"federation-sample","namespace":"org1"},"spec":{"description":"federation with org0 \u0026 org1","license":{"accept":true},"members":[{"initiator":true,"name":"org1","namespace":"org1"},{"name":"org2","namespace":"org2"}],"policy":"ALL"}}
  creationTimestamp: "2022-12-08T04:14:40Z"
  generation: 2
  name: federation-sample
  namespace: org1
  resourceVersion: "8794"
  uid: 534fb03b-e11f-4f5c-be48-b57c4d71cf04
spec:
  description: federation with org0 & org1
  license:
    accept: true
  members:
  - initiator: true
    name: org1
    namespace: org1 
  - name: org3
    namespace: org3
  policy: ALL
status:
  lastHeartbeatTime: 2022-12-08 12:35:44.826345 +0800 CST m=+33.724520930
  status: "True"
  type: FederationActivated
```

</details>

### 4. 解散联盟

某个组织想要解散现有联盟，需要按以下步骤：

1. 创建一个提案（Proposal），类型为解散联盟（dissolveFederation），这会触发在每个可以投票的组织中自动创建 选票（vote）资源。
2. 等待每个可以投票的组织手动修改选票（Vote），即投票，表明自己的意愿，同意或者拒绝。
3. 投票符合提案（Proposal）中规定的策略（spec.Policy）后，成功添加该组织作为成员。

#### 4.1 发起提案 proposal，内容为解散联盟

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_proposal_dissolve_federation.yaml
```

<details>

<summary>完整proposal yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Proposal
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Proposal","metadata":{"annotations":{},"name":"dissolve-federation-sample"},"spec":{"dissolveFederation":{},"federation":{"name":"federation-sample","namespace":"org1"},"initiator":{"name":"org1","namespace":"org1"},"policy":"All"}}
  creationTimestamp: "2022-12-08T05:58:32Z"
  generation: 2
  name: dissolve-federation-sample
  resourceVersion: "9710"
  uid: 9cd71ce9-22d4-4b4f-b3d4-947e57fdd10d
spec:
  deprecated: false
  dissolveFederation: {}
  endAt: "2022-12-09T05:58:32Z"
  federation:
    name: federation-sample
    namespace: org1
  initiator:
    name: org1
    namespace: org1
  policy: All
  startAt: "2022-12-08T05:58:32Z"
status:
  phase: Voting
  votes:
  - description: ""
    organization:
      name: org3
      namespace: org3
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2022-12-08T05:58:32Z"
```

</details>

<details>

<summary>此时选票结果为：</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:58:32Z"
    generation: 2
    labels:
      app: dissolve-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-dissolve-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: dissolve-federation-sample
      uid: 9cd71ce9-22d4-4b4f-b3d4-947e57fdd10d
    resourceVersion: "9709"
    uid: 96c880a1-71c7-4af3-92b9-aba2a4904f6c
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: dissolve-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:58:32Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:58:32Z"
    generation: 1
    labels:
      app: dissolve-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org3-dissolve-federation-sample
    namespace: org3
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: dissolve-federation-sample
      uid: 9cd71ce9-22d4-4b4f-b3d4-947e57fdd10d
    resourceVersion: "9704"
    uid: 5a537d77-167f-41ac-8ac0-c23946b820fa
  spec:
    description: ""
    organizationName: org3
    proposalName: dissolve-federation-sample
  status:
    phase: Created
kind: List
metadata:
  resourceVersion: ""
```

</details>

#### 4.2 涉及到的组织投票

```bash
kubectl patch vote -n org3 vote-org3-dissolve-federation-sample --type='json' -p='[{"op": "replace", "path": "/spec/decision", "value": true}]'
```

<details>

<summary>此时选票结果为：</summary>

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:58:32Z"
    generation: 2
    labels:
      app: dissolve-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org1-dissolve-federation-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: dissolve-federation-sample
      uid: 9cd71ce9-22d4-4b4f-b3d4-947e57fdd10d
    resourceVersion: "9709"
    uid: 96c880a1-71c7-4af3-92b9-aba2a4904f6c
  spec:
    decision: true
    description: ""
    organizationName: org1
    proposalName: dissolve-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T05:58:32Z"
- apiVersion: ibp.com/v1beta1
  kind: Vote
  metadata:
    creationTimestamp: "2022-12-08T05:58:32Z"
    generation: 2
    labels:
      app: dissolve-federation-sample
      app.kubernetes.io/instance: fabricproposal
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: vote-org3-dissolve-federation-sample
    namespace: org3
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Proposal
      name: dissolve-federation-sample
      uid: 9cd71ce9-22d4-4b4f-b3d4-947e57fdd10d
    resourceVersion: "9832"
    uid: 5a537d77-167f-41ac-8ac0-c23946b820fa
  spec:
    decision: true
    description: ""
    organizationName: org3
    proposalName: dissolve-federation-sample
  status:
    phase: Voted
    startTime: "2022-12-08T06:00:04Z"
kind: List
metadata:
  resourceVersion: ""

```

</details>

#### 4.3 投票成功，联盟成功解散

联盟状态变为：

```yaml
status:
  lastHeartbeatTime: 2022-12-08 14:00:04.395943 +0800 CST m=+302.039719659
  status: "True"
  type: FederationDissolved
```

<details>

<summary>联盟yaml为：</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Federation
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Federation","metadata":{"annotations":{},"name":"federation-sample","namespace":"org1"},"spec":{"description":"federation with org0 \u0026 org1","license":{"accept":true},"members":[{"initiator":true,"name":"org1","namespace":"org1"},{"name":"org2","namespace":"org2"}],"policy":"ALL"}}
  creationTimestamp: "2022-12-08T04:14:40Z"
  generation: 4
  name: federation-sample
  namespace: org1
  resourceVersion: "9841"
  uid: 534fb03b-e11f-4f5c-be48-b57c4d71cf04
spec:
  description: federation with org0 & org1
  license:
    accept: true
  members:
  - initiator: true
    name: org1
    namespace: org1
  policy: ALL
status:
  lastHeartbeatTime: 2022-12-08 14:00:04.395943 +0800 CST m=+302.039719659
  status: "True"
  type: FederationDissolved

```

</details>

## 网络管理

### 1. 创建网络

要想创建一个网络 (Network)，需要按以下步骤：

1. 创建一个网络 (Network) 资源。

#### 1.1 创建单 orderer 网络 Network 的 CR

需要先将 yaml 中的 `<org1AdminToken>` 替换为 发起者 initiator 组织 org1 管理员用户的 token。

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_network.yaml
```

<details>

<summary>详细yaml为:</summary>

```bash
$ kubectl get network  -oyaml
```

```yaml
apiVersion: v1
items:
  - apiVersion: ibp.com/v1beta1
    kind: Network
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: |
          {"apiVersion":"ibp.com/v1beta1","kind":"Network","metadata":{"annotations":{},"name":"network-sample"},"spec":{"federation":"federation-sample","initialToken":"eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA","license":{"accept":true},"members":[{"initiator":true,"name":"org1","namespace":"org1"},{"name":"org2","namespace":"org2"}]}}
      creationTimestamp: "2023-01-10T06:18:09Z"
      generation: 1
      name: network-sample
      resourceVersion: "7533"
      uid: 24b0acee-f1cf-466d-b42f-f7d704fa8c0f
    spec:
      federation: federation-sample
      initialToken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
      license:
        accept: true
      members:
        - initiator: true
          name: org1
          namespace: org1
        - name: org2
          namespace: org2
      orderSpec:
        action:
          enroll: {}
          reenroll: {}
        clusterSize: 1
        customNames:
          pvc: {}
        ingress: {}
        license:
          accept: true
        ordererType: etcdraft
    status:
      lastHeartbeatTime: 2023-01-10 06:35:10.090086501 +0000 UTC m=+1.405550566
      status: "True"
      type: Created
kind: List
metadata:
  resourceVersion: ""

```

</details>

<details>

<summary>orderer 详情为:</summary>

```bash
$ kubectl get ibporderer -n org1 -o yaml
```

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: IBPOrderer
  metadata:
    creationTimestamp: "2023-01-10T06:35:10Z"
    generation: 1
    labels:
      app: network-sample
      app.kubernetes.io/instance: fabricnetwork
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: network-sample
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Network
      name: network-sample
      uid: 24b0acee-f1cf-466d-b42f-f7d704fa8c0f
    resourceVersion: "9784"
    uid: dcf13d87-305d-45b3-a535-fa20ec1ebdee
  spec:
    action:
      enroll: {}
      reenroll: {}
    clusterSize: 1
    clustersecret:
    - enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample0
          enrollsecret: network-sample0
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample0
          enrollsecret: network-sample0
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    customNames:
      pvc: {}
    domain: 172.22.0.2.nip.io
    ingress: {}
    license:
      accept: true
    mspID: network-sample
    ordererType: etcdraft
    orgName: org1
    systemChannelName: network-sample
    useChannelLess: true
    version: 2.4.7
  status:
    lastHeartbeatTime: 2023-01-10 06:45:35.369558997 +0000 UTC m=+626.685023050
    reason: All nodes are deployed
    status: "True"
    type: Deployed
    version: 1.0.0
    versions:
      reconciled: ""
- apiVersion: ibp.com/v1beta1
  kind: IBPOrderer
  metadata:
    creationTimestamp: "2023-01-10T06:35:10Z"
    generation: 2
    labels:
      app: network-samplenode1
      app.kubernetes.io/instance: fabricorderer
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      parent: network-sample
    name: network-samplenode1
    namespace: org1
    resourceVersion: "9783"
    uid: 20ab9cf7-4e5e-4357-81ca-8fd04f77048e
  spec:
    action:
      enroll: {}
      reenroll: {}
    clusterSize: 1
    customNames:
      pvc: {}
    domain: 172.22.0.2.nip.io
    externalAddress: org1-network-samplenode1-orderer.172.22.0.2.nip.io:443
    images:
      grpcwebImage: hyperledgerk8s/grpc-web
      grpcwebTag: latest
      ordererImage: hyperledgerk8s/fabric-orderer
      ordererInitImage: hyperledgerk8s/ubi-minimal
      ordererInitTag: latest
      ordererTag: 2.4.7
    ingress: {}
    license:
      accept: true
    mspID: network-sample
    number: 1
    ordererType: etcdraft
    orgName: org1
    replicas: 1
    secret:
      enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample0
          enrollsecret: network-sample0
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1-network-samplenode1-orderer.172.22.0.2.nip.io
            - org1-network-samplenode1-operations.172.22.0.2.nip.io
            - org1-network-samplenode1-grpcweb.172.22.0.2.nip.io
            - org1-network-samplenode1-admin.172.22.0.2.nip.io
            - 127.0.0.1
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample0
          enrollsecret: network-sample0
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    systemChannelName: network-sample
    useChannelLess: true
    version: 2.4.7-1
  status:
    lastHeartbeatTime: 2023-01-10 06:45:35.358822706 +0000 UTC m=+626.674286766
    message: allPodsRunning
    reason: allPodsRunning
    status: "True"
    type: Deployed
    version: 1.0.0
    versions:
      reconciled: 2.4.7-1
kind: List
metadata:
  resourceVersion: ""

```

</details>

#### 1.2 创建三 orderer 网络 Network 的 CR

需要先将 yaml 中的 `<org1AdminToken>` 替换为 发起者 initiator 组织 org1 管理员用户的 token。

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_network_size_3.yaml
```

<details>

<summary>详细yaml为:</summary>

```bash
$ kubectl get network -o yaml
```

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: Network
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"ibp.com/v1beta1","kind":"Network","metadata":{"annotations":{},"name":"network-sample3"},"spec":{"federation":"federation-sample","initialToken":"eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA","license":{"accept":true},"members":[{"initiator":true,"name":"org1","namespace":"org1"},{"name":"org2","namespace":"org2"}],"orderSpec":{"clusterSize":3,"license":{"accept":true}}}}
    creationTimestamp: "2023-01-10T06:57:06Z"
    generation: 1
    name: network-sample3
    resourceVersion: "11866"
    uid: 05539cd0-c92d-4633-8755-4444919bde9d
  spec:
    federation: federation-sample
    initialToken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
    license:
      accept: true
    members:
    - initiator: true
      name: org1
      namespace: org1
    - name: org2
      namespace: org2
    orderSpec:
      action:
        enroll: {}
        reenroll: {}
      clusterSize: 3
      customNames:
        pvc: {}
      ingress: {}
      license:
        accept: true
      ordererType: etcdraft
  status:
    lastHeartbeatTime: 2023-01-10 06:57:06.14278582 +0000 UTC m=+1317.458249887
    status: "True"
    type: Created
kind: List
metadata:
  resourceVersion: ""

```

</details>

<details>

<summary>orderer 详情为:</summary>

```bash
$ kubectl get ibporderer -norg1 -o yaml
```

```yaml
apiVersion: v1
items:
- apiVersion: ibp.com/v1beta1
  kind: IBPOrderer
  metadata:
    creationTimestamp: "2023-01-10T06:57:06Z"
    generation: 1
    labels:
      app: network-sample3
      app.kubernetes.io/instance: fabricnetwork
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      helm.sh/chart: ibm-fabric
      release: operator
    name: network-sample3
    namespace: org1
    ownerReferences:
    - apiVersion: ibp.com/v1beta1
      blockOwnerDeletion: true
      controller: true
      kind: Network
      name: network-sample3
      uid: 05539cd0-c92d-4633-8755-4444919bde9d
    resourceVersion: "12329"
    uid: eda4aa96-6eca-4d05-9fec-8836e324baff
  spec:
    action:
      enroll: {}
      reenroll: {}
    clusterSize: 3
    clustersecret:
    - enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample30
          enrollsecret: network-sample30
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample30
          enrollsecret: network-sample30
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    - enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample31
          enrollsecret: network-sample31
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample31
          enrollsecret: network-sample31
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    - enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample32
          enrollsecret: network-sample32
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample32
          enrollsecret: network-sample32
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    customNames:
      pvc: {}
    domain: 172.22.0.2.nip.io
    ingress: {}
    license:
      accept: true
    mspID: network-sample3
    ordererType: etcdraft
    orgName: org1
    systemChannelName: network-sample3
    useChannelLess: true
    version: 2.4.7
  status:
    lastHeartbeatTime: 2023-01-10 06:58:04.886683429 +0000 UTC m=+1376.202147482
    reason: All nodes are deployed
    status: "True"
    type: Deployed
    version: 1.0.0
    versions:
      reconciled: ""
- apiVersion: ibp.com/v1beta1
  kind: IBPOrderer
  metadata:
    creationTimestamp: "2023-01-10T06:57:06Z"
    generation: 2
    labels:
      app: network-sample3node1
      app.kubernetes.io/instance: fabricorderer
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      parent: network-sample3
    name: network-sample3node1
    namespace: org1
    resourceVersion: "12282"
    uid: 866677ee-3b40-4a9f-a9f7-417422463df3
  spec:
    action:
      enroll: {}
      reenroll: {}
    clusterSize: 1
    customNames:
      pvc: {}
    domain: 172.22.0.2.nip.io
    externalAddress: org1-network-sample3node1-orderer.172.22.0.2.nip.io:443
    images:
      grpcwebImage: hyperledgerk8s/grpc-web
      grpcwebTag: latest
      ordererImage: hyperledgerk8s/fabric-orderer
      ordererInitImage: hyperledgerk8s/ubi-minimal
      ordererInitTag: latest
      ordererTag: 2.4.7
    ingress: {}
    license:
      accept: true
    mspID: network-sample3
    number: 1
    ordererType: etcdraft
    orgName: org1
    replicas: 1
    secret:
      enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample30
          enrollsecret: network-sample30
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1-network-sample3node1-orderer.172.22.0.2.nip.io
            - org1-network-sample3node1-operations.172.22.0.2.nip.io
            - org1-network-sample3node1-grpcweb.172.22.0.2.nip.io
            - org1-network-sample3node1-admin.172.22.0.2.nip.io
            - 127.0.0.1
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample30
          enrollsecret: network-sample30
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    systemChannelName: network-sample3
    useChannelLess: true
    version: 2.4.7-1
  status:
    lastHeartbeatTime: 2023-01-10 06:57:55.598828339 +0000 UTC m=+1366.914292397
    message: allPodsRunning
    reason: allPodsRunning
    status: "True"
    type: Deployed
    version: 1.0.0
    versions:
      reconciled: 2.4.7-1
- apiVersion: ibp.com/v1beta1
  kind: IBPOrderer
  metadata:
    creationTimestamp: "2023-01-10T06:57:06Z"
    generation: 2
    labels:
      app: network-sample3node2
      app.kubernetes.io/instance: fabricorderer
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      parent: network-sample3
    name: network-sample3node2
    namespace: org1
    resourceVersion: "12328"
    uid: ced9cdee-1f44-4c8c-86fc-6be1f447f61b
  spec:
    action:
      enroll: {}
      reenroll: {}
    clusterSize: 1
    customNames:
      pvc: {}
    domain: 172.22.0.2.nip.io
    externalAddress: org1-network-sample3node2-orderer.172.22.0.2.nip.io:443
    images:
      grpcwebImage: hyperledgerk8s/grpc-web
      grpcwebTag: latest
      ordererImage: hyperledgerk8s/fabric-orderer
      ordererInitImage: hyperledgerk8s/ubi-minimal
      ordererInitTag: latest
      ordererTag: 2.4.7
    ingress: {}
    license:
      accept: true
    mspID: network-sample3
    number: 2
    ordererType: etcdraft
    orgName: org1
    replicas: 1
    secret:
      enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample31
          enrollsecret: network-sample31
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1-network-sample3node2-orderer.172.22.0.2.nip.io
            - org1-network-sample3node2-operations.172.22.0.2.nip.io
            - org1-network-sample3node2-grpcweb.172.22.0.2.nip.io
            - org1-network-sample3node2-admin.172.22.0.2.nip.io
            - 127.0.0.1
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample31
          enrollsecret: network-sample31
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    systemChannelName: network-sample3
    useChannelLess: true
    version: 2.4.7-1
  status:
    lastHeartbeatTime: 2023-01-10 06:58:04.875963798 +0000 UTC m=+1376.191427856
    message: allPodsRunning
    reason: allPodsRunning
    status: "True"
    type: Deployed
    version: 1.0.0
    versions:
      reconciled: 2.4.7-1
- apiVersion: ibp.com/v1beta1
  kind: IBPOrderer
  metadata:
    creationTimestamp: "2023-01-10T06:57:06Z"
    generation: 2
    labels:
      app: network-sample3node3
      app.kubernetes.io/instance: fabricorderer
      app.kubernetes.io/managed-by: fabric-operator
      app.kubernetes.io/name: fabric
      creator: fabric
      parent: network-sample3
    name: network-sample3node3
    namespace: org1
    resourceVersion: "12319"
    uid: 4caa1b14-ba5d-4324-97d6-2290d9293c9c
  spec:
    action:
      enroll: {}
      reenroll: {}
    clusterSize: 1
    customNames:
      pvc: {}
    domain: 172.22.0.2.nip.io
    externalAddress: org1-network-sample3node3-orderer.172.22.0.2.nip.io:443
    images:
      grpcwebImage: hyperledgerk8s/grpc-web
      grpcwebTag: latest
      ordererImage: hyperledgerk8s/fabric-orderer
      ordererInitImage: hyperledgerk8s/ubi-minimal
      ordererInitTag: latest
      ordererTag: 2.4.7
    ingress: {}
    license:
      accept: true
    mspID: network-sample3
    number: 3
    ordererType: etcdraft
    orgName: org1
    replicas: 1
    secret:
      enrollment:
        component:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          enrollid: network-sample32
          enrollsecret: network-sample32
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
        tls:
          cahost: org1-org1-ca.172.22.0.2.nip.io
          caname: ca
          caport: "443"
          catls:
            cacert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNYRENDQWdLZ0F3SUJBZ0lSQU9XVFIwVGRteXFzRFE4Y3VqRnlBSTh3Q2dZSUtvWkl6ajBFQXdJd2dZTXgKQ3pBSkJnTlZCQVlUQWxWVE1SY3dGUVlEVlFRSUV3NU9iM0owYUNCRFlYSnZiR2x1WVRFUE1BMEdBMVVFQnhNRwpSSFZ5YUdGdE1Rd3dDZ1lEVlFRS0V3TkpRazB4RXpBUkJnTlZCQXNUQ2tKc2IyTnJZMmhoYVc0eEp6QWxCZ05WCkJBTVRIbTl5WnpFdGIzSm5NUzFqWVM0eE56SXVNakl1TUM0eUxtNXBjQzVwYnpBZUZ3MHlNekF4TVRBd05qQTUKTXpoYUZ3MHpNekF4TURjd05qQTVNemhhTUlHRE1Rc3dDUVlEVlFRR0V3SlZVekVYTUJVR0ExVUVDQk1PVG05eQpkR2dnUTJGeWIyeHBibUV4RHpBTkJnTlZCQWNUQmtSMWNtaGhiVEVNTUFvR0ExVUVDaE1EU1VKTk1STXdFUVlEClZRUUxFd3BDYkc5amEyTm9ZV2x1TVNjd0pRWURWUVFERXg1dmNtY3hMVzl5WnpFdFkyRXVNVGN5TGpJeUxqQXUKTWk1dWFYQXVhVzh3V1RBVEJnY3Foa2pPUFFJQkJnZ3Foa2pPUFFNQkJ3TkNBQVIzNzJNeWhUOERuRW1Kb1ZHNwpwcHpBdzJTLzBsZTFvbEpENlUvc01idDhDT3NkbEt1OXdmL29Iai9FMjVyak9HdHFWWHllNmJyQU5TTHk2UUNhCjBCeFNvMVV3VXpCUkJnTlZIUkVFU2pCSWdoNXZjbWN4TFc5eVp6RXRZMkV1TVRjeUxqSXlMakF1TWk1dWFYQXUKYVcrQ0ptOXlaekV0YjNKbk1TMXZjR1Z5WVhScGIyNXpMakUzTWk0eU1pNHdMakl1Ym1sd0xtbHZNQW9HQ0NxRwpTTTQ5QkFNQ0EwZ0FNRVVDSVFDNE1EU1d6bTVabEZ2MDNvcHJNZkxIMjl4VWxKQTNYK1VacTMxVzRNRVV5QUlnClJSYUY4cnhUWXJibzJoUkoxandjUDBBQWh3VSs4b2F0VU9vdDQ0ZUJFQ3M9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
          csr:
            hosts:
            - org1-network-sample3node3-orderer.172.22.0.2.nip.io
            - org1-network-sample3node3-operations.172.22.0.2.nip.io
            - org1-network-sample3node3-grpcweb.172.22.0.2.nip.io
            - org1-network-sample3node3-admin.172.22.0.2.nip.io
            - 127.0.0.1
            - org1
            - org1.org1
            - org1.org1.svc.cluster.local
          enrollid: network-sample32
          enrollsecret: network-sample32
          enrolltoken: eyJhbGciOiJSUzI1NiIsImtpZCI6ImE5N2MyZGEzZDYzZDY2MTA0NmRlZTk3ZjM3MjM0ZWE4ZjU0Yzg4ZjYifQ.eyJpc3MiOiJodHRwczovL3BvcnRhbC4xNzIuMjIuMC4yLm5pcC5pby9vaWRjIiwic3ViIjoiQ2dsdmNtY3hZV1J0YVc0U0JtczRjMk55WkEiLCJhdWQiOiJiZmYtY2xpZW50IiwiZXhwIjoxNjczNDE3NzgxLCJpYXQiOjE2NzMzMzEzODEsImF0X2hhc2giOiJPa0h1cFJIcDQ3N2ZaUHZsNHRwd0lBIiwiY19oYXNoIjoibzJzQUhDWEk3WGdmaldQcGI4ZDROZyIsImVtYWlsIjoib3JnMWFkbWluQHRlbnhjbG91ZC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiZ3JvdXBzIjpbIm9ic2VydmFiaWxpdHkiLCJzeXN0ZW06bm9kZXMiLCJzeXN0ZW06bWFzdGVycyIsInJlc291cmNlLXJlYWRlciIsImlhbS50ZW54Y2xvdWQuY29tIiwib2JzZXJ2YWJpbGl0eSJdLCJuYW1lIjoib3JnMWFkbWluIiwicHJlZmVycmVkX3VzZXJuYW1lIjoib3JnMWFkbWluIiwicGhvbmUiOiIiLCJ1c2VyaWQiOiJvcmcxYWRtaW4ifQ.BWhqxQy_p9Ws6gd5HNwltchLzOfFldXW6MaanC3qZmon3Nyj31Kcw4gnv7kTeRTB1RPyfZXmK46eOtm6ADvjOl5X0vN7tLvq0A-BK-5bf9D-hU_BNGShyGD5XMNloDDp3IQwkukFK7HF2-VHv7tioeGVLquaOn4-2vf18jKNtZ1dxeR3lnfeO9tukqwuInSXJN7mvikueJxYtbHwFR98ED7Rb0uMdctS3wQH2m3PxoZ1bt2JaQOVR_Br6bcRuw3H3akXUuPaj38EgfJnVC6DCQwI3Pu7jEZo689BJ3g4dBaRWCepoDRN2Pd6I3bvZE8xCk8UV1Lxz2CUbC6y2PcrtA
          enrolluser: org1admin
    systemChannelName: network-sample3
    useChannelLess: true
    version: 2.4.7-1
  status:
    lastHeartbeatTime: 2023-01-10 06:58:03.635121235 +0000 UTC m=+1374.950585299
    message: allPodsRunning
    reason: allPodsRunning
    status: "True"
    type: Deployed
    version: 1.0.0
    versions:
      reconciled: 2.4.7-1
kind: List
metadata:
  resourceVersion: ""

```

</details>

#### 1.3 删除网络

删除网络需要创建提案 Proposal, 根据提案的 policy, 各参与组织同意后，才会自动删除。

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_proposal_dissolve_network.yaml
```

<details>

<summary>详细yaml为:</summary>

```yaml
apiVersion: ibp.com/v1beta1
kind: Proposal
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ibp.com/v1beta1","kind":"Proposal","metadata":{"annotations":{},"name":"dissolve-network-sample"},"spec":{"dissolveNetwork":{"name":"network-sample"},"federation":"federation-sample","initiator":{"name":"org1","namespace":"org1"},"policy":"All"}}
  creationTimestamp: "2023-01-10T08:39:09Z"
  generation: 1
  labels:
    bestchains.proposal.type: DissolveNetworkProposal
  name: dissolve-network-sample
  resourceVersion: "15337"
  uid: edc94532-9101-4170-be3c-23b9ba738d39
spec:
  deprecated: false
  dissolveNetwork:
    name: network-sample
  endAt: "2023-01-11T08:39:09Z"
  federation: federation-sample
  initiator:
    name: org1
    namespace: org1
  policy: All
  startAt: "2023-01-10T08:39:09Z"
status:
  phase: Voting
  votes:
  - description: ""
    organization:
      name: org2
      namespace: org2
  - decision: true
    description: ""
    organization:
      name: org1
      namespace: org1
    phase: Voted
    startTime: "2023-01-10T08:39:09Z"
```

</details>

org2 投同意票可以用以下命令：

```bash
kubectl patch vote -n org2 vote-org2-dissolve-network-sample --type='json' -p='[{"op": "replace", "path": "/spec/decision", "value": true}]'
```

稍后，operator 会自动删除 network。网络删除完成。

### 节点管理

#### 1. 创建org1peer1节点
创建一个peer节点，需要按照一下步骤：

##### 1.1 获取组织org1的CA连接信息

```bash
 kubectl get cm -norg1 org1-connection-profile -ojson | jq -r '.binaryData."profile.json"' | base64 --decode
```

<summary>详细json为:</summary>
<details>

```json
{
  "endpoints": {
    "api": "https://org1-org1-ca.172.18.0.2.nip.io:443",
    "operations": "https://org1-org1-operations.172.18.0.2.nip.io:443"
  },
  "tls": {
    "cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNXekNDQWdHZ0F3SUJBZ0lRZnJGV01iNXE2MFRueitYZUwrVEIwREFLQmdncWhrak9QUVFEQWpDQmd6RUwKTUFrR0ExVUVCaE1DVlZNeEZ6QVZCZ05WQkFnVERrNXZjblJvSUVOaGNtOXNhVzVoTVE4d0RRWURWUVFIRXdaRQpkWEpvWVcweEREQUtCZ05WQkFvVEEwbENUVEVUTUJFR0ExVUVDeE1LUW14dlkydGphR0ZwYmpFbk1DVUdBMVVFCkF4TWViM0puTVMxdmNtY3hMV05oTGpFM01pNHhPQzR3TGpJdWJtbHdMbWx2TUI0WERUSXpNREV6TVRBek1qZzEKT1ZvWERUTXpNREV5T0RBek1qZzFPVm93Z1lNeEN6QUpCZ05WQkFZVEFsVlRNUmN3RlFZRFZRUUlFdzVPYjNKMAphQ0JEWVhKdmJHbHVZVEVQTUEwR0ExVUVCeE1HUkhWeWFHRnRNUXd3Q2dZRFZRUUtFd05KUWsweEV6QVJCZ05WCkJBc1RDa0pzYjJOclkyaGhhVzR4SnpBbEJnTlZCQU1USG05eVp6RXRiM0puTVMxallTNHhOekl1TVRndU1DNHkKTG01cGNDNXBiekJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCSVJOUllMRUlnRXBTWTI2ZzMzRAozWElmN1d6YVJITURLTGNqd292a3ZBaHNyVDh2T1VUUEoxMWVIaC9seDRjWlkxUC82WE9pL2wzeTFudXdrNndOCkh2T2pWVEJUTUZFR0ExVWRFUVJLTUVpQ0htOXlaekV0YjNKbk1TMWpZUzR4TnpJdU1UZ3VNQzR5TG01cGNDNXAKYjRJbWIzSm5NUzF2Y21jeExXOXdaWEpoZEdsdmJuTXVNVGN5TGpFNExqQXVNaTV1YVhBdWFXOHdDZ1lJS29aSQp6ajBFQXdJRFNBQXdSUUloQUt5ZlFKTXpZV1lOSEdOMTlQZE5YSlIycjBhYjgwVE1KKzlrT0M2N0VKRFhBaUJGCjhTN3JHZmJmcC91N2hLNlRZb0NFRjF1RGR2OWxkYm5SeGg5WFdnNFhydz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
  },
  "ca": {
    "signcerts": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNFakNDQWJtZ0F3SUJBZ0lVSC9lNjEwWW1Wc0QyR2swa1MvWFBEN0pSZXc0d0NnWUlLb1pJemowRUF3SXcKWHpFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVJBd0RnWURWUVFERXdkdmNtY3hMV05oCk1CNFhEVEl6TURFek1UQXpNalF3TUZvWERUTTRNREV5TnpBek1qUXdNRm93WHpFTE1Ba0dBMVVFQmhNQ1ZWTXgKRnpBVkJnTlZCQWdURGs1dmNuUm9JRU5oY205c2FXNWhNUlF3RWdZRFZRUUtFd3RJZVhCbGNteGxaR2RsY2pFUApNQTBHQTFVRUN4TUdSbUZpY21sak1SQXdEZ1lEVlFRREV3ZHZjbWN4TFdOaE1Ga3dFd1lIS29aSXpqMENBUVlJCktvWkl6ajBEQVFjRFFnQUVIdFhNdGh2ckdmUkNFczhCdElPVnpvL2c2c0tDVUpTOWx5QXNIVE1iMW80ZDhwUmkKRjRZcUtlejZ2WXNOM2s4TEZhNytXenRuYkU2YW1MSHNKTEZ5ZWFOVE1GRXdEZ1lEVlIwUEFRSC9CQVFEQWdFRwpNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHdIUVlEVlIwT0JCWUVGQ09pbjNjOFFOUGEvS2JwSklFSVJsNk9vYTZkCk1BOEdBMVVkRVFRSU1BYUhCSDhBQUFFd0NnWUlLb1pJemowRUF3SURSd0F3UkFJZ1Y2Z2dzMkM0cGpPeUJJVlEKdm5KdGR3TGZTZER5bkJSaXpjVGJ4SC9NNlF3Q0lIbDZPNGRmUURGYSt5ditWODlkV2hvY3RLSDYxQWVBZ01aNwpldXdFNVdpagotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
  },
  "tlsca": {
    "signcerts": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNCekNDQWE2Z0F3SUJBZ0lVQ0ptV2NLQXVYaklvKzd6V2MxczVwQjZwK2tBd0NnWUlLb1pJemowRUF3SXcKWWpFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVJNd0VRWURWUVFERXdwdmNtY3hMWFJzCmMyTmhNQjRYRFRJek1ERXpNVEF6TWpRd01Gb1hEVE00TURFeU56QXpNalF3TUZvd1lqRUxNQWtHQTFVRUJoTUMKVlZNeEZ6QVZCZ05WQkFnVERrNXZjblJvSUVOaGNtOXNhVzVoTVJRd0VnWURWUVFLRXd0SWVYQmxjbXhsWkdkbApjakVQTUEwR0ExVUVDeE1HUm1GaWNtbGpNUk13RVFZRFZRUURFd3B2Y21jeExYUnNjMk5oTUZrd0V3WUhLb1pJCnpqMENBUVlJS29aSXpqMERBUWNEUWdBRWthZFhoTExBWFlRWkRkcktUNXNpdkV1clRQb0FrWU5JTDZpeFBMTWoKcDF3NGVUMWhqOVQ4U1psMTR2R0VqMjlnZjBvaDd4MU1ILzRmeDQ0VVlSZlRzNk5DTUVBd0RnWURWUjBQQVFILwpCQVFEQWdFR01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZNbU85WWF5WWV0RFg0cEREQnEwCnJkTjFpZkNWTUFvR0NDcUdTTTQ5QkFNQ0EwY0FNRVFDSUYweFQvVDFFaXB5ZmRxWEh0dXMydk4yNHprZzJxT1UKNG96ekRwdHlINXZPQWlCZzcyQ2pSNlQyRGV6a2ZiTTYwekdNbzBtclFXRzVFSTFFR09OOEE5aTVVZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
  }
}
```

</details>

需要：
- `endpoints.api` : ca地址
- `tls.cert` : ca tls 证书

##### 1.2 更新peer的创建yaml

yaml位于 `config/samples/peers/ibp.com_v1beta1_peer_org1peer1.yaml` ,需要:

- 替换`<org1-ca-cert>` 为 上一步骤得到的`tls.cert`
- 替换`<org1AdminToken>` 为用户org1admin的token

##### 1.3 创建org1peer1

```bash
kubectl apply -f config/samples/peers/ibp.com_v1beta1_peer_org1peer1.yaml
```

### 2. 创建org2peer1节点


##### 2.1 获取组织org2的CA连接信息

```bash
 kubectl get cm -norg2 org2-connection-profile -ojson | jq -r '.binaryData."profile.json"' | base64 --decode
```

<summary>详细json为:</summary>
<details>

```json
{
  "endpoints": {
    "api": "https://org2-org2-ca.172.18.0.2.nip.io:443",
    "operations": "https://org2-org2-operations.172.18.0.2.nip.io:443"
  },
  "tls": {
    "cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNXekNDQWdHZ0F3SUJBZ0lRZnJGV01iNXE2MFRueitYZUwrVEIwREFLQmdncWhrak9QUVFEQWpDQmd6RUwKTUFrR0ExVUVCaE1DVlZNeEZ6QVZCZ05WQkFnVERrNXZjblJvSUVOaGNtOXNhVzVoTVE4d0RRWURWUVFIRXdaRQpkWEpvWVcweEREQUtCZ05WQkFvVEEwbENUVEVUTUJFR0ExVUVDeE1LUW14dlkydGphR0ZwYmpFbk1DVUdBMVVFCkF4TWViM0puTVMxdmNtY3hMV05oTGpFM01pNHhPQzR3TGpJdWJtbHdMbWx2TUI0WERUSXpNREV6TVRBek1qZzEKT1ZvWERUTXpNREV5T0RBek1qZzFPVm93Z1lNeEN6QUpCZ05WQkFZVEFsVlRNUmN3RlFZRFZRUUlFdzVPYjNKMAphQ0JEWVhKdmJHbHVZVEVQTUEwR0ExVUVCeE1HUkhWeWFHRnRNUXd3Q2dZRFZRUUtFd05KUWsweEV6QVJCZ05WCkJBc1RDa0pzYjJOclkyaGhhVzR4SnpBbEJnTlZCQU1USG05eVp6RXRiM0puTVMxallTNHhOekl1TVRndU1DNHkKTG01cGNDNXBiekJaTUJNR0J5cUdTTTQ5QWdFR0NDcUdTTTQ5QXdFSEEwSUFCSVJOUllMRUlnRXBTWTI2ZzMzRAozWElmN1d6YVJITURLTGNqd292a3ZBaHNyVDh2T1VUUEoxMWVIaC9seDRjWlkxUC82WE9pL2wzeTFudXdrNndOCkh2T2pWVEJUTUZFR0ExVWRFUVJLTUVpQ0htOXlaekV0YjNKbk1TMWpZUzR4TnpJdU1UZ3VNQzR5TG01cGNDNXAKYjRJbWIzSm5NUzF2Y21jeExXOXdaWEpoZEdsdmJuTXVNVGN5TGpFNExqQXVNaTV1YVhBdWFXOHdDZ1lJS29aSQp6ajBFQXdJRFNBQXdSUUloQUt5ZlFKTXpZV1lOSEdOMTlQZE5YSlIycjBhYjgwVE1KKzlrT0M2N0VKRFhBaUJGCjhTN3JHZmJmcC91N2hLNlRZb0NFRjF1RGR2OWxkYm5SeGg5WFdnNFhydz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
  },
  "ca": {
    "signcerts": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNFakNDQWJtZ0F3SUJBZ0lVSC9lNjEwWW1Wc0QyR2swa1MvWFBEN0pSZXc0d0NnWUlLb1pJemowRUF3SXcKWHpFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVJBd0RnWURWUVFERXdkdmNtY3hMV05oCk1CNFhEVEl6TURFek1UQXpNalF3TUZvWERUTTRNREV5TnpBek1qUXdNRm93WHpFTE1Ba0dBMVVFQmhNQ1ZWTXgKRnpBVkJnTlZCQWdURGs1dmNuUm9JRU5oY205c2FXNWhNUlF3RWdZRFZRUUtFd3RJZVhCbGNteGxaR2RsY2pFUApNQTBHQTFVRUN4TUdSbUZpY21sak1SQXdEZ1lEVlFRREV3ZHZjbWN4TFdOaE1Ga3dFd1lIS29aSXpqMENBUVlJCktvWkl6ajBEQVFjRFFnQUVIdFhNdGh2ckdmUkNFczhCdElPVnpvL2c2c0tDVUpTOWx5QXNIVE1iMW80ZDhwUmkKRjRZcUtlejZ2WXNOM2s4TEZhNytXenRuYkU2YW1MSHNKTEZ5ZWFOVE1GRXdEZ1lEVlIwUEFRSC9CQVFEQWdFRwpNQThHQTFVZEV3RUIvd1FGTUFNQkFmOHdIUVlEVlIwT0JCWUVGQ09pbjNjOFFOUGEvS2JwSklFSVJsNk9vYTZkCk1BOEdBMVVkRVFRSU1BYUhCSDhBQUFFd0NnWUlLb1pJemowRUF3SURSd0F3UkFJZ1Y2Z2dzMkM0cGpPeUJJVlEKdm5KdGR3TGZTZER5bkJSaXpjVGJ4SC9NNlF3Q0lIbDZPNGRmUURGYSt5ditWODlkV2hvY3RLSDYxQWVBZ01aNwpldXdFNVdpagotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
  },
  "tlsca": {
    "signcerts": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNCekNDQWE2Z0F3SUJBZ0lVQ0ptV2NLQXVYaklvKzd6V2MxczVwQjZwK2tBd0NnWUlLb1pJemowRUF3SXcKWWpFTE1Ba0dBMVVFQmhNQ1ZWTXhGekFWQmdOVkJBZ1REazV2Y25Sb0lFTmhjbTlzYVc1aE1SUXdFZ1lEVlFRSwpFd3RJZVhCbGNteGxaR2RsY2pFUE1BMEdBMVVFQ3hNR1JtRmljbWxqTVJNd0VRWURWUVFERXdwdmNtY3hMWFJzCmMyTmhNQjRYRFRJek1ERXpNVEF6TWpRd01Gb1hEVE00TURFeU56QXpNalF3TUZvd1lqRUxNQWtHQTFVRUJoTUMKVlZNeEZ6QVZCZ05WQkFnVERrNXZjblJvSUVOaGNtOXNhVzVoTVJRd0VnWURWUVFLRXd0SWVYQmxjbXhsWkdkbApjakVQTUEwR0ExVUVDeE1HUm1GaWNtbGpNUk13RVFZRFZRUURFd3B2Y21jeExYUnNjMk5oTUZrd0V3WUhLb1pJCnpqMENBUVlJS29aSXpqMERBUWNEUWdBRWthZFhoTExBWFlRWkRkcktUNXNpdkV1clRQb0FrWU5JTDZpeFBMTWoKcDF3NGVUMWhqOVQ4U1psMTR2R0VqMjlnZjBvaDd4MU1ILzRmeDQ0VVlSZlRzNk5DTUVBd0RnWURWUjBQQVFILwpCQVFEQWdFR01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZNbU85WWF5WWV0RFg0cEREQnEwCnJkTjFpZkNWTUFvR0NDcUdTTTQ5QkFNQ0EwY0FNRVFDSUYweFQvVDFFaXB5ZmRxWEh0dXMydk4yNHprZzJxT1UKNG96ekRwdHlINXZPQWlCZzcyQ2pSNlQyRGV6a2ZiTTYwekdNbzBtclFXRzVFSTFFR09OOEE5aTVVZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
  }
}
```

</details>

需要：
- `endpoints.api` : ca地址
- `tls.cert` : ca tls 证书

##### 1.2 更新peer的创建yaml

yaml位于 `config/samples/peers/ibp.com_v1beta1_peer_org2peer1.yaml` ,需要:

- 替换`<org2-ca-cert>` 为 上一步骤得到的`tls.cert`
- 替换`<org2AdminToken>` 为用户org2admin的token

##### 1.3 创建org2peer1

```bash
kubectl apply -f config/samples/peers/ibp.com_v1beta1_peer_org2peer1.yaml
```

### 通道管理

#### 1. 创建通道

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_channel_create.yaml
```

#### 2. peer节点加入通道

```bash
kubectl apply -f config/samples/ibp.com_v1beta1_channel_join.yaml
```
