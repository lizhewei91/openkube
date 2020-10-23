# openKube
自定义CRD资源demo，kubebuilder v2.3.1

## 1、初始化项目目录

```go
export GO111MODULE=on   // 开启gomod

mkdir openkube    // 在gopath下面创建项目目录

kubebuilder init --domain openkube.com   // domain指定CRD资源group的域名,

注意：domain需要全部小写
```

## 2、API创建

```go
kubebuilder create api --group apps --version v1beta1 --kind UnitedSet --namespaced true 
// 实际上不仅会创建 API，也就是 CRD，还会生成 Controller 的框架

Create Resource [y/n]
y
Create Controller [y/n]
y
```

**参数解读**：- group 加上之前的 domian 即此 CRD 的 Group: apps.kruise.io；

- version 一般分三种，按社区标准：
  - v1alpha1: 此 api 不稳定，CRD 可能废弃、字段可能随时调整，不要依赖；
  - v1beta1: api 已稳定，会保证向后兼容，特性可能会调整；
  - v1: api 和特性都已稳定；
- kind: 此 CRD 的类型，类似于社区原生的 Service 的概念；
- namespaced: 此 CRD 是k8s集群级别资源还是 namespace隔离的，类似 node 和 Pod

## 3、创建webhook（可选）

https://www.qikqiak.com/post/k8s-admission-webhook/

```go
// 创建pod的 MutatingAdmissionWebhook
kubebuilder create webhook --group core --version v1 --kind Pod --defaulting

// 创建UnitedSet的 MutatingAdmissionWebhook 和 ValidatingAdmissionWebhook
kubebuilder create webhook --group apps --version v1beta1 --kind UnitedSet --defaulting --programmatic-validation
```
## 4、定义 CRD

在图 1中对应的文件定义 Spec 和 Status。

## 5、编写 Controller 和wenbook逻辑

在图 1 中对应的文件实现 Reconcile 以及webhook逻辑。

## 6、本地调试运行
### 6.1、CRD安装

```go
开启go mod模式
export GO111MODULE=on
export GOPROXY="http://mirrors.aliyun.com/goproxy"
export GOPROXY="https://goproxy.cn"

第一种方式：
将代码拷贝至k8s集群机器
make install    // 先生成config/crd/bases文件，里面包含crd的YAML文件，然后kubectl apply

第二种方式: 
在本地机器执行，make install，在config/crd/bases目录生成crd.yaml，拷贝至k8s集群机器，执行

kubectl create -f apps.openkube.com_unitedsets.yaml
```

然后我们就可以看到创建的CRD了

```
# kubectl get crd
NAME                           CREATED AT
unitedsets.apps.openkube.com   2020-10-15T03:35:45Z
```

创建一个unitedSet资源

```
# cd openKube/config/samples

# kubectl create -f apps_v1beta1_unitedset.yaml

# kubectl get unitedsets.apps.openkube.com
NAME               AGE
unitedset-sample   67s
```

看一眼yaml文件

```
[root@k8s-master samples]# cat apps_v1beta1_unitedset.yaml 
apiVersion: apps.openkube.com/v1beta1
kind: UnitedSet
metadata:
  name: unitedset-sample
spec:
  # Add fields here
  foo: bar
```

这里仅仅是把yaml存到etcd里了，我们controller监听到创建事件时啥事也没干。
### 6.2、修改配置文件

因为启用了webhook，所以要对默认的配置文件进行一些修改，来到config目录，config的核心是config/default目录。

#### 6.2.1、修改config/default/kustomization.yaml
```
# Adds namespace to all resources.
namespace: openkube-system

# Value of this field is prepended to the
# names of all resources, e.g. a deployment named
# "wordpress" becomes "alices-wordpress".
# Note that it should also match with the prefix (text before '-') of the namespace
# field above.
namePrefix: openkube-

# Labels to add to all resources and selectors.
#commonLabels:
#  someName: someValue

bases:
- ../crd
- ../rbac
- ../manager
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in 
# crd/kustomization.yaml
- ../webhook
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'. 'WEBHOOK' components are required.
#- ../certmanager
# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'. 
#- ../prometheus

patchesStrategicMerge:
  # Protect the /metrics endpoint by putting it behind auth.
  # If you want your controller-manager to expose the /metrics
  # endpoint w/o any authn/z, please comment the following line.
- manager_auth_proxy_patch.yaml

# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in 
# crd/kustomization.yaml
#- manager_webhook_patch.yaml

# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'.
# Uncomment 'CERTMANAGER' sections in crd/kustomization.yaml to enable the CA injection in the admission webhooks.
# 'CERTMANAGER' needs to be enabled to use ca injection
#- webhookcainjection_patch.yaml

# the following config is for teaching kustomize how to do var substitution
vars:
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER' prefix.
#- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR
#  objref:
#    kind: Certificate
#    group: cert-manager.io
#    version: v1alpha2
#    name: serving-cert # this name should match the one in certificate.yaml
#  fieldref:
#    fieldpath: metadata.namespace
#- name: CERTIFICATE_NAME
#  objref:
#    kind: Certificate
#    group: cert-manager.io
#    version: v1alpha2
#    name: serving-cert # this name should match the one in certificate.yaml
- name: SERVICE_NAMESPACE # namespace of the service
  objref:
    kind: Service
    version: v1
    name: webhook-service
  fieldref:
    fieldpath: metadata.namespace
- name: SERVICE_NAME
  objref:
    kind: Service
    version: v1
    name: webhook-service
```

#### 6.2.2、修改config/crd/kustomization.yaml文件
```
# This kustomization.yaml is not intended to be run by itself,
# since it depends on service name and namespace that are out of this kustomize package.
# It should be run by config/default
resources:
- bases/apps.openkube.com_unitedsets.yaml
# +kubebuilder:scaffold:crdkustomizeresource

patchesStrategicMerge:
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix.
# patches here are for enabling the conversion webhook for each CRD
- patches/webhook_in_unitedsets.yaml
# +kubebuilder:scaffold:crdkustomizewebhookpatch

# [CERTMANAGER] To enable webhook, uncomment all the sections with [CERTMANAGER] prefix.
# patches here are for enabling the CA injection for each CRD
#- patches/cainjection_in_unitedsets.yaml
# +kubebuilder:scaffold:crdkustomizecainjectionpatch

# the following config is for teaching kustomize how to do kustomization for CRDs.
configurations:
- kustomizeconfig.yaml
```
### 7、修改Makefile
```
# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
        cd config/manager && kustomize edit set image controller=${IMG}
#       kustomize build config/default | kubectl apply -f -
        kustomize build config/default > all_in_one.yaml
```
可以看到，此命令会使用kustomize订制整个config/default目录下的配置文件，生成所有的资源文件，再使用kubectl apply命令部署，但直接apply在部分版本的K8s中可能会出错。为了更清晰地了解kustomize生成的资源有哪些，我将它做了一些小修改，不直接apply，转而将资源重定向到all_in_one.yaml文件内。
### all_in_one
#### 分析
仔细分析一番生成的all_in_one.yaml文件，有6000多行，其中的CustomResourceDefinition资源占据绝大部分的内容，总共可大概有这几种类型的资源:
```
# CRD的资源描述,涉及到Unit的每一个字段，因此非常冗长.
kind: CustomResourceDefinition

# admission webhook
kind: MutatingWebhookConfiguration
kind: ValidatingWebhookConfiguration

# RBAC授权
kind: Role
kind: ClusterRole
kind: RoleBinding
kind: ClusterRoleBinding

# prometheus metric service
kind: Service

# openkube-webhook-service，接收APIServer的回调
kind: Service

# openkube controller deployment
kind: Deployment
```
### 8、修改all_in_one.yaml文件
#### 8.1、需要把all_in_one.yaml文件中CustomResourceDefinition.spec下新增一个字段：`preserveUnknownFields: false`
否则不加此字段kubectl apply会报错，bug已知存在于1.15-1.17以下的版本中，参考: Generated Metadata breaks crd
```
apiVersion: v1
kind: Namespace
metadata:
  labels:
    control-plane: controller-manager
  name: openkube-system
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.5
  name: unitedsets.apps.openkube.com
spec:
  preserveUnknownFields: false                           // 增加此参数
  conversion:
    strategy: Webhook
    webhookClientConfig:
...
```
#### 8.2、修改MutatingWebhookConfiguration 和 ValidatingWebhookConfiguration
这两个webhook配置需要修改什么呢？来看看下载的配置，以为例：MutatingWebhookConfiguration
```
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  creationTimestamp: null
  name: openkube-mutating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: Cg=
    service:
      name: openkube-webhook-service
      namespace: openkube-system
      path: /mutate-pod
  failurePolicy: Fail
  name: mpod.kb.io
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
```
这里面有两个地方要修改：

- caBundle现在是空的，需要补上
- clientConfig现在的配置是ca授权给的是Service unit-webhook-service，也即是会转发到deployment的pod，但我们现在是要本地调试，这里就要改成本地环境。

下面来讲述如何配置这两个点。
##### 8.2.1、CA证书签发
这里要分为多个步骤：

**1.ca.cert**

首先获取K8s CA的CA.cert文件：

```
kubectl config view --raw -o json | jq -r '.clusters[0].cluster."certificate-authority-data"' | tr -d '"' > ca.cert

```

ca.cert的内容，即可复制替换到上面的MutatingWebhookConfiguration和ValidatingWebhookConfigurationd的`webhooks.clientConfig.caBundle`里。(原来的`Cg==`要删掉.)

**2.csr**

创建证书签署请求json配置文件：

注意，hosts里面填写两种内容：

- openkube controller的service 在K8s中的域名，最后openkube controller是要放在K8s里运行的。
- 本地开发机的某个网卡IP地址，这个地址用来连接K8s集群进行调试。因此必须保证这个IP与K8s集群可以互通

```shell
cat > openkube-csr.json << EOF
{
  "hosts": [
    "openkube-webhook-service.default.svc",
    "openkube-webhook-service.default.svc.cluster.local",
    "192.168.254.1"
  ],
  "CN": "openkube-webhook-service",
  "key": {
    "algo": "rsa",
    "size": 2048
  }
}
EOF
```

**3.生成csr和pem私钥文件:**

```shell
[root@k8s-master deploy]# cat openkube-csr.json | cfssl genkey - | cfssljson -bare openkube
2020/05/23 17:44:39 [INFO] generate received request
2020/05/23 17:44:39 [INFO] received CSR
2020/05/23 17:44:39 [INFO] generating key: rsa-2048
2020/05/23 17:44:39 [INFO] encoded CSR
[root@k8s-master deploy]# ls
openkube.csr  openkube-csr.json  openkube-key.pem
```

**4.创建CertificateSigningRequest资源**

```shell
cat > csr.yaml << EOF 
apiVersion: certificates.k8s.io/v1beta1
kind: CertificateSigningRequest
metadata:
  name: openkube
spec:
  request: $(cat openkube.csr | base64 | tr -d '\n')
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

# apply
kubectl apply -f csr.yaml
```

**5.向集群提交此CertificateSigningRequest.**

查看状态：

```shell
[root@k8s-master deploy]# kubectl apply -f csr.yaml 
certificatesigningrequest.certificates.k8s.io/openkube created
[root@k8s-master deploy]# kubectl describe csr openkube 
Name:         openkube
Labels:       <none>
...

CreationTimestamp:  Tue, 20 Oct 2020 23:37:47 -0400
Requesting User:    kubernetes-admin
Signer:             kubernetes.io/legacy-unknown
Status:             Pending
Subject:
  Common Name:    openkube-webhook-service
  Serial Number:  
Subject Alternative Names:
         DNS Names:     openkube-webhook-service.default.svc
                        openkube-webhook-service.default.svc.cluster.local
         IP Addresses:  10.200.224.94
Events:  <none>
```

可以看到它还是pending的状态，需要同意一下请求:

```shell
[root@k8s-master deploy]#  kubectl certificate approve openkube
certificatesigningrequest.certificates.k8s.io/openkube approved
[root@k8s-master deploy]# kubectl get csr openkube 
NAME       AGE     SIGNERNAME                     REQUESTOR          CONDITION
openkube   2m15s   kubernetes.io/legacy-unknown   kubernetes-admin   Approved,Issued
# 保存客户端crt文件
[root@k8s-master deploy]# kubectl get csr openkube -o jsonpath='{.status.certificate}' | base64 --decode > openkube.crt
```

可以看到，现在已经签署完毕了。

汇总一下：

- 第1步生成的ca.cert文件给caBundle字段使用
- 第3步生成的unit-key.pem私钥文件和第5步生成的unit.crt文件，提供给客户端(unit controller)https服务使用

##### 8.2.2、更新WebhookConfiguration
根据上面生成的证书相关内容，对all_in_one.yaml 中的WebhookConfiguration进行替换，替换之后：
```
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  creationTimestamp: null
  name: openkube-mutating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01EY3hNekF5TURBMU4xb1hEVE13TURjeE1UQXlNREExTjFvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBUEZRCkhlbDd1SW4vVGVGbmFHRVlvOElOQWlZeEFrbmF1UnY2TGhYVXYwV25lRHFONVJ1M0FoRzM5ZlJvSFhzQURRUWIKWWRXRmtjU1FJUGNzWmdJMy9nNFdGVkhvd2s3Rk5qdUlWdzBDc0dsS05XZldkeG9NV0Y4YkNsVEpoRU9Td3ZkbQprU1NYdllpdzlkVVhDa3kweXgvdzE2WkN4ZXBPUzJsTUFqc3ZjenhWd3dZbUZ4RDBQUEdrbkIwOUdNUjVXNzdPCjVQdFhRMUo2T1pNcmVLdDB0YllOZ25LTHdqSk5GZVFxZzZCR2ZYQVFwUkhNUHpIdHlNV3MxWkg2TTQwZ0FoVEgKQjkwbzJ1QWQwRXpWM1p0YWp3dnBHSmN0UFJvQXNmYWpjTDRqOFU5MUlpTmMyLzRtK3RQZ0lCVUJZT0ZJWDBWbApQVG8yMU1tcWdBSXREeGNJRDZjQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFPclFRMXZoMlk0WE5mWC9BWEluNEtuVlArdVIKaW15Y0dQdmJiZTJqYXZOVHRGMGJ5Ui9QdGxNNDlzSmxZWEhwSStyYmZxYzd3SGx5RTdmVFA3TWRseTJFTDVTUwpCQW94d3p6eVUzeFpBTUptbHRGWitWMDZ4Zi94aE1TY2FvTW5EQ0Q2TmdBbmZvMkNmRGxaR2NHeXJNNUVZeDFICkg5dXpNWmxTc3NSQTE4Y1c3UlNZYW9oazhHenZJc2grUitKd1l5bnoxZTBzOXVZYzFEM20xbWl4dFpsY2dGNjAKOW5leFlVQlBqVlIxc1ZLZGxsdmcxclRHZHdTQ3ZaV28vL1NBS0RmWEZHRnFoM3QyWEVzcUNWY3pkWVozMFdpVQpQMGdNbnBzWW8yUmpGbHN6SitoUUVRNlBQei9HZHQ4RlEwODlzTTBtM2ZWdnRMRXc3NVpzczgrRUphdz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    url: https://10.200.224.94:9877/mutate-pod
#    service:
#      name: openkube-webhook-service
#      namespace: openkube-system
#      path: /mutate-pod
  failurePolicy: Fail
  name: mpod.kb.io
  rules:
  - apiGroups:
    - ""
    apiVersions:
    - v1
    operations:
    - CREATE
    resources:
    - pods
- clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01EY3hNekF5TURBMU4xb1hEVE13TURjeE1UQXlNREExTjFvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBUEZRCkhlbDd1SW4vVGVGbmFHRVlvOElOQWlZeEFrbmF1UnY2TGhYVXYwV25lRHFONVJ1M0FoRzM5ZlJvSFhzQURRUWIKWWRXRmtjU1FJUGNzWmdJMy9nNFdGVkhvd2s3Rk5qdUlWdzBDc0dsS05XZldkeG9NV0Y4YkNsVEpoRU9Td3ZkbQprU1NYdllpdzlkVVhDa3kweXgvdzE2WkN4ZXBPUzJsTUFqc3ZjenhWd3dZbUZ4RDBQUEdrbkIwOUdNUjVXNzdPCjVQdFhRMUo2T1pNcmVLdDB0YllOZ25LTHdqSk5GZVFxZzZCR2ZYQVFwUkhNUHpIdHlNV3MxWkg2TTQwZ0FoVEgKQjkwbzJ1QWQwRXpWM1p0YWp3dnBHSmN0UFJvQXNmYWpjTDRqOFU5MUlpTmMyLzRtK3RQZ0lCVUJZT0ZJWDBWbApQVG8yMU1tcWdBSXREeGNJRDZjQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFPclFRMXZoMlk0WE5mWC9BWEluNEtuVlArdVIKaW15Y0dQdmJiZTJqYXZOVHRGMGJ5Ui9QdGxNNDlzSmxZWEhwSStyYmZxYzd3SGx5RTdmVFA3TWRseTJFTDVTUwpCQW94d3p6eVUzeFpBTUptbHRGWitWMDZ4Zi94aE1TY2FvTW5EQ0Q2TmdBbmZvMkNmRGxaR2NHeXJNNUVZeDFICkg5dXpNWmxTc3NSQTE4Y1c3UlNZYW9oazhHenZJc2grUitKd1l5bnoxZTBzOXVZYzFEM20xbWl4dFpsY2dGNjAKOW5leFlVQlBqVlIxc1ZLZGxsdmcxclRHZHdTQ3ZaV28vL1NBS0RmWEZHRnFoM3QyWEVzcUNWY3pkWVozMFdpVQpQMGdNbnBzWW8yUmpGbHN6SitoUUVRNlBQei9HZHQ4RlEwODlzTTBtM2ZWdnRMRXc3NVpzczgrRUphdz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    url: https://10.200.224.94:9877/mutate-apps-openKube-com-v1beta1-unitedset
#    service:
#      name: openkube-webhook-service
#      namespace: openkube-system
#      path: /mutate-apps-openKube-com-v1beta1-unitedset
  failurePolicy: Fail
  name: munitedset.kb.io
  rules:
  - apiGroups:
    - apps.openKube.com
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - unitedsets
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: ValidatingWebhookConfiguration
metadata:
  creationTimestamp: null
  name: openkube-validating-webhook-configuration
webhooks:
- clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01EY3hNekF5TURBMU4xb1hEVE13TURjeE1UQXlNREExTjFvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBUEZRCkhlbDd1SW4vVGVGbmFHRVlvOElOQWlZeEFrbmF1UnY2TGhYVXYwV25lRHFONVJ1M0FoRzM5ZlJvSFhzQURRUWIKWWRXRmtjU1FJUGNzWmdJMy9nNFdGVkhvd2s3Rk5qdUlWdzBDc0dsS05XZldkeG9NV0Y4YkNsVEpoRU9Td3ZkbQprU1NYdllpdzlkVVhDa3kweXgvdzE2WkN4ZXBPUzJsTUFqc3ZjenhWd3dZbUZ4RDBQUEdrbkIwOUdNUjVXNzdPCjVQdFhRMUo2T1pNcmVLdDB0YllOZ25LTHdqSk5GZVFxZzZCR2ZYQVFwUkhNUHpIdHlNV3MxWkg2TTQwZ0FoVEgKQjkwbzJ1QWQwRXpWM1p0YWp3dnBHSmN0UFJvQXNmYWpjTDRqOFU5MUlpTmMyLzRtK3RQZ0lCVUJZT0ZJWDBWbApQVG8yMU1tcWdBSXREeGNJRDZjQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFPclFRMXZoMlk0WE5mWC9BWEluNEtuVlArdVIKaW15Y0dQdmJiZTJqYXZOVHRGMGJ5Ui9QdGxNNDlzSmxZWEhwSStyYmZxYzd3SGx5RTdmVFA3TWRseTJFTDVTUwpCQW94d3p6eVUzeFpBTUptbHRGWitWMDZ4Zi94aE1TY2FvTW5EQ0Q2TmdBbmZvMkNmRGxaR2NHeXJNNUVZeDFICkg5dXpNWmxTc3NSQTE4Y1c3UlNZYW9oazhHenZJc2grUitKd1l5bnoxZTBzOXVZYzFEM20xbWl4dFpsY2dGNjAKOW5leFlVQlBqVlIxc1ZLZGxsdmcxclRHZHdTQ3ZaV28vL1NBS0RmWEZHRnFoM3QyWEVzcUNWY3pkWVozMFdpVQpQMGdNbnBzWW8yUmpGbHN6SitoUUVRNlBQei9HZHQ4RlEwODlzTTBtM2ZWdnRMRXc3NVpzczgrRUphdz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    url: https://10.200.224.94:9877/validate-apps-openKube-com-v1beta1-unitedset
#    service:
#      name: openkube-webhook-service
#      namespace: openkube-system
#      path: /validate-apps-openKube-com-v1beta1-unitedset
  failurePolicy: Fail
  name: vunitedset.kb.io
  rules:
  - apiGroups:
    - apps.openKube.com
    apiVersions:
    - v1beta1
    operations:
    - CREATE
    - UPDATE
    resources:
    - unitedsets
 ```
注意，url中的ip地址需要是**本地开发机的ip地址**，同时此ip需要能与K8s集群正常通信，uri为service.path.

修改完两个WebhookConfiguration之后，下一步就可以去部署all_in_one.yaml文件了，由于现在controller要在本地运行调试，因此，这个阶段，要记得把all_in_one_local.yaml中的Deployment资源部分注释掉。
```
[root@k8s-master deploy]# kubectl apply -f all_in_one_local.yaml --validate=false
namespace/openkube-system created
customresourcedefinition.apiextensions.k8s.io/unitedsets.apps.openkube.com created
role.rbac.authorization.k8s.io/openkube-leader-election-role created
clusterrole.rbac.authorization.k8s.io/openkube-manager-role created
clusterrole.rbac.authorization.k8s.io/openkube-proxy-role created
clusterrole.rbac.authorization.k8s.io/openkube-metrics-reader created
rolebinding.rbac.authorization.k8s.io/openkube-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/openkube-manager-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/openkube-proxy-rolebinding created
service/openkube-controller-manager-metrics-service created
service/openkube-webhook-service created
mutatingwebhookconfiguration.admissionregistration.k8s.io/openkube-mutating-webhook-configuration created
```
K8s这边的CRD资源、webhook资源、RBAC授权都已经搞定了，下一步就是启动本地的controller进行调试了。
### 9、本地启动controller

启动之前要把上面准备好的证书、私钥，放在指定的目录内，默认指定目录是：`/tmp/k8s-webhook-server/serving-certs/`

```shell
windows系统的Temp目录，C:\Users\suning\AppData\Local\Temp

进入Temp目录，创建

# mkdir -r /$TMPDIR/k8s-webhook-server/serving-certs/
# cp openkube-key.pem $TMPDIR/k8s-webhook-server/serving-certs/tls.key
# cp openkube.crt $TMPDIR/k8s-webhook-server/serving-certs/tls.crt
```

证书准备好之后，就可以在IDE内启动controller了：
```
go run main.go
```
可以开始愉快的调试了~
### 10、部署sample

假设调试已经完毕，可以开始测试部署一个Unit实例了。

默认的sample在这里：`config/samples/apps_v1beta1_unitedset.yaml`，里面的Group、version、kind等已经填好了，补充下内容即可，例如sample.yaml：

```yaml
apiVersion: apps.openkube.com/v1beta1
kind: UnitedSet
metadata:
  name: unitedset-sample
#  namespace: tc
spec:
  # Add fields here
  foo: bar
```

```shell
[root@k8s-master samples]# kubectl apply -f apps_v1beta1_unitedset.yaml 
unitedset.apps.openkube.com/unitedset-sample created
[root@k8s-master samples]# kubectl get unitedsets.apps.openkube.com 
NAME               AGE
unitedset-sample   16s
[root@k8s-master samples]# kubectl describe unitedsets.apps.openkube.com unitedset-sample 
Name:         unitedset-sample
Namespace:    default
Labels:       <none>
Annotations:  openkube.com/unitedSet-hash: 7vzdzzfdf44w88b92wb9fb767dczxx647848bb5vx2wxw968d4bv2x24v425w9x4
API Version:  apps.openkube.com/v1beta1
Kind:         UnitedSet
Metadata:
  Creation Timestamp:  2020-10-22T07:45:05Z
```

可以看到，unitedSet资源已经创建成功了，hash值已经注入到`Annotations`中

### 11、发布

如果已经调试和测试完毕，可以进入正式发布了。弄清了上面的步骤，发布比较简单了

#### 11.1、打包push docker镜像

修改Makefile中的IMG，`IMG ?= 192.168.87.134:5000/controller-manage/openkube-controller:v1.0.0`

```shell
# make docker-build docker-push 
```

#### 11.2、 Make deploy

```shell
# make deploy
```

#### 11.3、 修改deployment

为什么要修改deployment呢？还是因为证书的问题，deployment运行同样也需要证书，那就将证书做成Secret资源，以Secret的形式挂载进pod里面把。

**生成secret**

```shell
# kubectl create secret generic openkube-cert --from-file=./tls.crt --from-file=./tls.key
```

**修改Deployment，添加Secret挂载**
```
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    metadata:
      labels:
        control-plane: controller-manager
    spec:
      containers:
      - args:
        - --metrics-addr=127.0.0.1:8080
        - --enable-leader-election
        command:
        - /manager
        image: 192.168.87.134:5000/controller-manage/openkube-controller:v1.0.0
        name: manager
        resources:
          limits:
            cpu: 100m
            memory: 30Mi
          requests:
            cpu: 100m
            memory: 20Mi
      - args:
        - --secure-listen-address=0.0.0.0:8443
        - --upstream=http://127.0.0.1:8080/
        - --logtostderr=true
        - --v=10
        image: gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0
        name: kube-rbac-proxy
        ports:
        - containerPort: 8443
          name: https
        volumeMounts:
        - mountPath: /tmp/k8s-webhook-server/serving-certs
          name: cert
          readOnly: true
      volumes:
      - name: cert
        secret:
          defaultMode: 420
          secretName: openkube-cert
      terminationGracePeriodSeconds: 10
```
另外，这两点不要忘记：

添加`CustomResourceDefinition.spec.preserveUnknownFields: false`

`webhooks.clientConfig.caBundleca`值配置

修改完毕，清理前面的apply的资源和sample，再次执行`kubectl apply -f all_in_one.yaml --validate=false`命令，可以看到，部署成功！

# 参考资料
https://book.kubebuilder.io/

https://blog.upweto.top/gitbooks/kubebuilder/

https://blog.csdn.net/qq_24271853/article/details/107085126



