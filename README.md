# web-operator

WebServer Operator 是一个基于 Kubernetes Operator 模式构建的控制器，用于自动化管理 Web 服务的部署和生命周期。

## 项目简介

WebServer Operator 允许用户通过 Kubernetes Custom Resource Definition (CRD) 来定义和管理 Web 服务。只需创建一个 `WebServer` 资源，Operator 会自动创建对应的 Deployment 和 Service 资源，并持续监控和维护其状态。

### 主要功能

- **自动化部署**: 创建 WebServer CR 后自动创建 Deployment 和 Service
- **动态更新**: 支持动态调整副本数、镜像、端口等配置
- **状态监控**: 实时更新 WebServer 的运行状态
- **命名空间支持**: 支持将资源部署到指定的 namespace
- **自动清理**: 删除 WebServer CR 时自动清理关联资源

## 快速开始

### 前置条件

- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- 访问 Kubernetes v1.11.3+ 集群的权限

### 部署到集群

**构建并推送镜像到 `IMG` 指定的位置:**

```sh
make docker-build docker-push IMG=<some-registry>/web-operator:tag
```

**注意:** 镜像需要发布到您指定的私有 registry。
确保您的工作环境有权限拉取该镜像。
如果上述命令无法工作，请检查您对 registry 的权限。

**安装 CRD 到集群:**

```sh
make install
```

**使用 `IMG` 指定的镜像部署 Manager 到集群:**

```sh
make deploy IMG=<some-registry>/web-operator:tag
```

> **注意**: 如果遇到 RBAC 错误，您可能需要授予自己 cluster-admin 权限或以管理员身份登录。

**创建解决方案实例**

您可以从 config/sample 应用示例：

```sh
kubectl apply -k config/samples/
```

>**注意**: 确保示例有默认值以便测试。

### 卸载

**从集群删除实例 (CRs):**

```sh
kubectl delete -k config/samples/
```

**从集群删除 APIs(CRDs):**

```sh
make uninstall
```

**从集群卸载控制器:**

```sh
make undeploy
```

## 使用示例

创建一个 WebServer 资源：

```yaml
apiVersion: web.xm.web/v1
kind: WebServer
metadata:
  name: my-webserver
  namespace: default
spec:
  replicas: 3
  image: nginx:1.30.0
  port: 80
  serviceType: ClusterIP
  # targetNamespace: other-namespace  # 可选：部署到指定 namespace
```

## 项目分发

以下是向用户发布和提供此解决方案的选项。

### 提供包含所有 YAML 文件的 bundle

1. 为构建并发布到 registry 的镜像构建安装程序：

```sh
make build-installer IMG=<some-registry>/web-operator:tag
```

**注意:** 上面提到的 makefile 目标会在 dist 目录中生成一个 'install.yaml' 文件。此文件包含使用 Kustomize 构建的所有资源，这些资源是在没有依赖项的情况下安装此项目所必需的。

2. 使用安装程序

用户只需运行 'kubectl apply -f <YAML BUNDLE 的 URL>' 即可安装项目，例如：

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/web-operator/<tag or branch>/dist/install.yaml
```

### 提供 Helm Chart

1. 使用可选的 helm 插件构建 chart

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. 查看在 'dist/chart' 下生成的 chart，用户可以从那里获取此解决方案。

**注意:** 如果您更改了项目，需要使用上面相同的命令更新 Helm Chart 以同步最新更改。此外，如果您创建 webhooks，需要使用上面带有 '--force' 标志的命令，并手动确保之前添加到 'dist/chart/values.yaml' 或 'dist/chart/manager/manager.yaml' 的任何自定义配置之后手动重新应用。


## 许可证

Copyright 2026 xiongming.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.