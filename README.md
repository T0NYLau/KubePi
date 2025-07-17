<p align="center"><a href="https://kubeoperator.io"><img src="https://kubeoperator.oss-cn-beijing.aliyuncs.com/kubepi/img/logo-red.png" alt="kubepi" width="300" /></a></p>
<P align="center"><b>KubePi</b> [kubəˈpaɪ]，一个现代化的 K8s 面板。</P>
<p align="center">
  <a href="https://www.gnu.org/licenses/gpl-3.0.html"><img src="https://shields.io/github/license/1Panel-dev/KubePi?color=%231890FF" alt="License: GPL v3"></a>
  <a href="https://github.com/1Panel-dev/KubePi/releases"><img src="https://img.shields.io/github/v/release/1Panel-dev/KubePi" alt="GitHub release"></a>
  <a href="https://github.com/1Panel-dev/KubePi"><img src="https://img.shields.io/github/stars/1Panel-dev/KubePi?color=%231890FF&style=flat-square" alt="GitHub Stars"></a>
  <a href="https://hub.docker.com/r/1panel/kubepi"><img src="https://img.shields.io/docker/pulls/1panel/kubepi?label=downloads" alt="Docker Pulls"></a>
</p>
<hr />

## KubePi 是什么？

KubePi 是一个现代化的 K8s 面板。KubePi 允许管理员导入多个 Kubernetes 集群，并且通过权限控制，将不同 cluster、namespace 的权限分配给指定用户；允许开发人员管理 Kubernetes 集群中运行的应用程序并对其进行故障排查，供开发人员更好地处理 Kubernetes 集群中的复杂性。

## 快速开始

```
docker run --privileged -d --restart=unless-stopped -p 80:80 1panel/kubepi

# 用户名: admin
# 密码: kubepi
```

你也可以通过 [1Panel 应用商店](https://apps.fit2cloud.com/1panel) 快速部署 KubePi。

使用手册请参考: [https://github.com/1Panel-dev/KubePi/wiki](https://github.com/1Panel-dev/KubePi/wiki)。

## UI 展示

![UI展示](https://kubeoperator.oss-cn-beijing.aliyuncs.com/kubepi/img/02-dashboard.png)

## 飞致云的其他明星项目

- [1Panel](https://github.com/1panel-dev/1panel/) - 现代化、开源的 Linux 服务器运维管理面板
- [JumpServer](https://github.com/jumpserver/jumpserver/) - 广受欢迎的开源堡垒机
- [DataEase](https://github.com/dataease/dataease/) - 人人可用的开源数据可视化分析工具
- [MeterSphere](https://github.com/metersphere/metersphere/) - 开源持续测试工具
- [Halo](https://github.com/halo-dev/halo/) - 强大易用的开源建站工具
- [MaxKB](https://github.com/1Panel-dev/MaxKB/) - 基于 LLM 大语言模型的开源知识库问答系统

## License

Copyright (c) 2014-2025 [FIT2CLOUD 飞致云](https://fit2cloud.com/), All rights reserved.

Licensed under The GNU General Public License version 3 (GPLv3)  (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at

<https://www.gnu.org/licenses/gpl-3.0.html>

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

# KubePi LLM集成修复

## 问题描述
KubePi平台在与DeepSeek AI模型集成时出现问题，用户配置后测试时无法显示AI回答。请求似乎很快完成但没有返回内容，响应中的model和object字段为空。

## 根本原因
1. KubePi后端在构造LLM API请求时，会自动在BaseURI后面添加`/v1/chat/completions`路径，而用户在配置DeepSeek模型时已经提供了完整路径，导致最终请求URL变成了无效的地址。
2. 错误处理和提示信息不够明确，导致用户难以定位问题。

## 修复内容

### 1. 后端修复
- 修改了`internal/service/v1/llm/llm.go`中API端点的构造方式，使用用户提供的完整BaseURI，不再自动添加路径

### 2. 前端优化
- 在LLM模型创建和编辑页面更新了BaseURI输入框的placeholder提示
- 添加了明确的提示信息，说明需要输入完整的API端点URL，包括`/v1/chat/completions`路径
- 优化了DeepSeek特殊响应格式的处理

### 3. 错误处理增强
- 增强了API错误处理，提供了更详细和针对性的错误信息
- 特别针对404错误提供了明确的配置指导
- 改进了网络错误和超时情况的用户提示

## 使用方法
配置DeepSeek模型时，请在BaseURI字段中输入完整的API端点URL，例如：
```
http://10.110.104.201:33117/v1/chat/completions
```

## 注意事项
- 不同LLM模型的API端点格式可能不同，请根据具体模型的API文档配置
- DeepSeek模型通常包含`</think>`标记分隔思考过程和回答，系统已处理此特性
- 如遇到问题，请检查错误信息，确认API端点配置是否正确
