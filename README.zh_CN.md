<div align="center">
  <h1>OpenDSN: A DAG-Based Multi-Version Decentralized Storage Network</h1>
  <p>
    <img src="https://img.shields.io/badge/Go-%E2%89%A51.18.1-00ADD8?style=flat&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Rust-rustup-black?style=flat&logo=rust&logoColor=white" alt="Rust">
    <img src="https://img.shields.io/badge/Sector-8MiB-blue" alt="Sector">
    <img src="https://img.shields.io/badge/Frontend-React%20%2B%20Vite-61DAFB?style=flat&logo=react&logoColor=white" alt="Frontend">
    <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
  </p>
  <p>
    <a href="./README.md">English</a> · <a href="./README.zh_CN.md">简体中文</a>
  </p>
</div>

OpenDSN 是一个面向多版本文件管理的去中心化存储网络系统，提供文件导入、初始版本上传、增量更新、版本链查询、指定版本检索、矿工状态采集、Proof 信息展示以及前后端一体化的演示与运维能力。

它既可以作为本地多节点实验环境使用，也可以部署到云服务器中，通过脚本快速拉起创世节点、普通存储节点、后端 API 与前端页面。

## 🎬 系统演示

OpenDSN 前端页面主要围绕三个核心功能展开：文件信息查询、文件上传和文件检索。

### 文件信息查询

<p align="center">
  <img src="./assets/readme/file_infomation.gif" alt="OpenDSN 文件信息查询演示" width="900">
</p>

查询系统当前记录的所有 root，并查看指定 root 对应的完整版本链信息。

### 文件上传

<p align="center">
  <img src="./assets/readme/file_upload.gif" alt="OpenDSN 文件上传演示" width="900">
</p>

先导入本地文件生成 root CID，再通过前端提交初始版本或更新版本的存储交易。

### 文件检索

<p align="center">
  <img src="./assets/readme/file_retrieval.gif" alt="OpenDSN 文件检索演示" width="900">
</p>

根据 root 或 head 发起版本检索请求，将指定版本的文件恢复到本地工作区。

## ✨ 系统特性

- **多版本文件管理**：支持文件 root、head 和完整版本链的维护与查询。
- **增量更新机制**：支持初始版本创建和后续版本更新，适合多版本文件演示场景。
- **可视化运维**：前端页面可展示存储节点信息、最新 Proof 信息以及版本相关操作结果。
- **脚本化管理**：提供启动脚本、退出脚本和 nginx 托管样板，便于本地实验与云部署。
- **演示友好**：支持前后端分离部署，适合实验演示、课程展示和系统验证。

## 目录
- [1. 项目概述](#1-项目概述)
- [2. 部署指南](#2-部署指南)
- [3. 运行示例](#3-运行示例)
- [4. 系统前端部署](#4-系统前端部署)
- [5. 退出系统](#5-退出系统)

## 1. 项目概述

OpenDSN 由以下几个核心部分组成：

- **链与存储节点**
  - `lotus daemon` 与 `lotus-miner` 负责链服务与存储节点运行。
- **节点启动脚本**
  - `scripts/genesis_node_start.sh` 用于启动创世节点。
  - `scripts/node_start.sh` 用于让普通节点接入现有网络。
- **后端 API**
  - `cmd/opendsn-api` 提供前端调用接口，负责聚合 miner 信息、proof 信息以及多版本文件操作。
- **前端页面**
  - `demo-web` 提供系统展示页面，可通过 `nginx` 托管。
- **退出脚本**
  - `scripts/exit.sh` 用于统一关闭仓库内启动的核心进程。

OpenDSN 当前默认围绕 `8MiB` sector、脚本化启动流程和演示型前端进行组织，便于在实验室内网、云服务器或课程展示环境中快速复现一套完整的多版本 DSN 系统。

Copyright (c) 2023-2025, Guo Hechuan, MIT License

## 2. 部署指南

### 2.1 系统要求

<table>
  <tr>
    <th align="left">类别</th>
    <th align="left">要求</th>
  </tr>
  <tr>
    <td>CPU</td>
    <td>4 核及以上</td>
  </tr>
  <tr>
    <td>内存</td>
    <td>8GB 及以上</td>
  </tr>
  <tr>
    <td>存储</td>
    <td>支持 8MiB sectors 的存储空间</td>
  </tr>
  <tr>
    <td>网络</td>
    <td>稳定的网络连接</td>
  </tr>
  <tr>
    <td>操作系统</td>
    <td>Linux 或 macOS</td>
  </tr>
  <tr>
    <td>Go</td>
    <td>1.18.1 或更高版本</td>
  </tr>
  <tr>
    <td>Rust</td>
    <td>建议通过 <code>rustup</code> 安装</td>
  </tr>
  <tr>
    <td>其他依赖</td>
    <td><code>git</code>、<code>jq</code>、<code>pkg-config</code>、<code>clang</code>、<code>hwloc</code>、<code>bsdiff</code> / <code>bspatch</code> 等</td>
  </tr>
</table>

> [!TIP]
> 如果你计划在云服务器上部署多节点网络，建议优先准备同一内网中的多台机器，并提前确认节点之间的私网互通和安全组规则。

### 2.2 环境准备与编译

#### 2.2.1 安装系统依赖

Ubuntu / Debian:

```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget bsdiff -y && sudo apt upgrade -y
```

macOS:

```bash
brew install go bzr jq pkg-config rustup hwloc coreutils bsdiff
```

#### 2.2.2 安装 Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

#### 2.2.3 安装 Go

```bash
wget -c https://golang.org/dl/go1.18.1.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
```

#### 2.2.4 配置环境变量

```bash
export LOTUS_PATH=~/.lotus-local-net
export LOTUS_MINER_PATH=~/.lotus-miner-local-net
export LOTUS_SKIP_GENESIS_CHECK=_yes_
export CGO_CFLAGS_ALLOW="-D__BLST_PORTABLE__"
export CGO_CFLAGS="-D__BLST_PORTABLE__"
export IPFS_GATEWAY=https://proof-parameters.s3.cn-south-1.jdcloud-oss.com/ipfs/
```

#### 2.2.5 获取源码并构建

```bash
git clone https://github.com/BDS-SDU/OpenDSN.git
cd OpenDSN
make debug
```

> [!IMPORTANT]
> 构建完成后，请确认仓库根目录下已经生成 `lotus`、`lotus-miner`、`lotus-seed` 等可执行文件，再继续后续步骤。

### 2.3 启动创世节点

OpenDSN 当前推荐通过脚本一键启动第一个节点：

```bash
cd /path/to/OpenDSN
bash scripts/genesis_node_start.sh
```

#### 脚本会自动完成的工作

- 清理旧的本地网络数据和日志
- 拉取 8MiB 参数文件
- 预密封 2 个创世扇区
- 生成 `localnet.json` 和 `devgen.car`
- 启动创世 `lotus daemon`
- 导入创世矿工密钥
- 初始化创世矿工 `t01000`
- 启动 `lotus-miner`
- 启动 `listen_and_send.sh`
- 启动 `opendsn-api`

#### 脚本执行后需要记录的信息

```bash
./lotus net listen
./lotus-miner net listen
```

这两条命令分别给出：

- 创世节点 daemon 的 multiaddr
- 创世节点 miner 的 multiaddr

后续添加更多节点时，需要把这两个地址记录下来。

> [!TIP]
> `genesis_node_start.sh` 已经集成了创世节点、监听脚本和后端 API 的启动逻辑。部署时推荐优先使用该脚本，而不是手动逐条执行命令。

### 2.4 添加更多节点

新增节点时，推荐统一使用：

```bash
bash scripts/node_start.sh <daemon_multiaddr> <miner_multiaddr> <genesis_ip> [local_ip]
```

在运行脚本前，还需要先把创世节点生成的：

```text
devgen.car
```

复制到新节点的仓库根目录。

#### 参数说明

- `daemon_multiaddr`
  - 来自创世节点执行 `./lotus net listen` 的输出
- `miner_multiaddr`
  - 来自创世节点执行 `./lotus-miner net listen` 的输出
- `genesis_ip`
  - 创世节点所在机器的 IP
  - 新节点会通过这个地址把钱包地址和本机 IP 发给创世节点上的 `listen_and_send.sh`
- `local_ip`
  - 可选参数
  - 如果不传，脚本会自动检测本机 IP
  - 如果本机同时有多个网卡或多个地址，建议显式指定

#### `node_start.sh` 自动完成的工作

`scripts/node_start.sh` 会自动完成“普通节点接入网络并成为存储提供者”的流程。脚本内部包括：

- 启动本地 `lotus daemon`
- 连接到创世节点 daemon 和 miner
- 创建钱包地址
- 将钱包地址和本机 IP 发给创世节点
- 等待创世节点打款
- 初始化 `lotus-miner`
- 启动 `lotus-miner`
- 配置 `sealed` 与 `unseal` 存储目录

一个典型示例如下：

```bash
bash scripts/node_start.sh \
  /ip4/192.168.1.10/tcp/1234/p2p/12D3KooW... \
  /ip4/192.168.1.10/tcp/2345/p2p/12D3KooW... \
  192.168.1.10 \
  192.168.1.11
```

如果最后一个 `local_ip` 不传，脚本会优先自动检测公网 IPv4；没有公网时，再回退到私网 IPv4。

> [!IMPORTANT]
> 新节点在运行前必须已经拿到创世节点生成的 `devgen.car`，并且能够访问创世节点的 `9999` 端口以及 lotus daemon 和 lotus miner 的通信地址。

### 2.5 常见问题

1. 创世节点启动失败
- 检查 `go`、`rustup` 和系统依赖是否安装完整
- 检查 `make debug` 是否成功完成
- 检查端口是否被旧进程占用

2. 新节点无法加入网络
- 确认新节点目录下已经复制了 `devgen.car`
- 确认 `daemon_multiaddr` 和 `miner_multiaddr` 复制完整
- 确认创世节点 `9999` 端口和节点通信端口在内网可达

3. 新节点没有收到创世节点打款
- 确认创世节点上的 `listen_and_send.sh` 已启动
- 确认 `genesis_ip` 填写正确
- 确认新节点向创世节点发送的钱包地址和本机 IP 没有被防火墙拦截

4. 多网卡环境 IP 不正确
- 创世节点和普通节点现在都优先选择公网 IPv4
- 如果你的部署环境更适合使用私网地址，建议在 `node_start.sh` 中显式传入 `local_ip`

## 3. 运行示例

### 3.1 创建存储交易：`client import + client deal`

OpenDSN 的文件上传通常分两步：

1. 先通过 `lotus client import` 获取文件 CID / root
2. 再通过 `lotus client deal` 将该 root 上传到网络

#### 导入本地文件

```bash
./lotus client import <inputPath>
```

例如：

```bash
./lotus client import v1
```

命令执行后会输出文件对应的 root CID。这个 root 会作为后续 `client deal`、`list-version` 和 `retrieve-version` 的关键输入。

#### 创建初始版本

```bash
./lotus client deal --create <rootCid> <miner> <price> <duration>
```

例如：

```bash
./lotus client deal --create bafykbzaceaanrd4jqhaaalcbsftfzzc7wfi6o7qk27hnjlhfmv72qeeokgtni t01000 0.026 518400
```

这会把当前文件标记为一个版本链的起始版本，并在本地生成：

- `fileroots.log`
- `<rootCid>_meta`

#### 上传新版本

```bash
./lotus client deal --update --previous <previousRoot> <newRoot> <miner> <price> <duration>
```

例如：

```bash
./lotus client deal --update --previous bafykbzaceaanrd4jqhaaalcbsftfzzc7wfi6o7qk27hnjlhfmv72qeeokgtni bafk2bzaceaox2bgqudb7o2wq7ra444fshsb7iedtic4htq2c63t4ysr77uzfk t01000 0.026 518400
```

这会把新版本接到已有版本链上，并更新对应的元数据文件。

### 3.2 文件检索：`retrieve-version`

OpenDSN 通过目标 root 或 head 恢复指定版本：

```bash
./lotus client retrieve-version <dataCid> <outputPath>
```

例如：

```bash
./lotus client retrieve-version bafk2bzacebjm5xaqy7sdklvzyhojz4swfgjqhh5lxzroa7zej4slm3bmgopzi retrieve_v3
```

该命令会：

- 从目标版本开始反向读取本地 `<cid>_meta`
- 找到整条版本链
- 依次取回 root 和 patch 文件
- 最后使用 `bspatch` 重建出指定版本

### 3.3 查看所有根版本：`list-root`

要查看当前 DSN 中记录的所有文件根版本：

```bash
./lotus client list-root
```

该命令会读取：

```text
fileroots.log
```

并输出类似：

```text
ROOT 0: <rootCid>
ROOT 1: <rootCid>
```

### 3.4 查看版本链：`list-version`

要查看某个文件从 root 开始的完整版本链：

```bash
./lotus client list-version <rootCid>
```

例如：

```bash
./lotus client list-version bafykbzaceaanrd4jqhaaalcbsftfzzc7wfi6o7qk27hnjlhfmv72qeeokgtni
```

命令会输出类似：

```text
ROOT(v1): <rootCid>
HEAD(v2): <headCid>
HEAD(v3): <headCid>
```

### 3.5 交易管理

1. 查看交易状态

```bash
./lotus client list-deals
```

2. 查看交易详情

```bash
./lotus client get-deal <DealCID>
```

### 3.6 常见问题

1. `client deal --create` 失败
- 检查 root CID 是否有效
- 检查 miner ID 是否正确
- 检查价格和时长是否符合链上约束

2. `client deal --update` 失败
- 检查 `--previous` 是否指向已有版本
- 检查 `<previousCid>_meta` 是否存在
- 检查新版本 root 是否已经通过 `client import` 获得

3. `retrieve-version` 失败
- 检查目标 CID 对应的 `<cid>_meta` 是否完整
- 检查 root / patch 文件是否都已成功上链
- 检查输出路径是否不存在

4. `list-root` 或 `list-version` 没有内容
- 说明当前目录下还没有形成完整的版本元数据
- 先执行 `client deal --create`，再执行 `client deal --update`

## 4. 系统前端部署

### 4.1 依赖安装

前端页面依赖：

- Node.js
- npm
- nginx

先检查当前机器是否已经安装：

```bash
node -v
npm -v
nginx -v
```

如果 `node -v` 和 `npm -v` 发现没有安装，或者版本过低导致前端无法编译，可以使用 `nvm` 安装较新的 Node.js 版本，例如：

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

nvm install 22
nvm use 22
nvm alias default 22

node -v
npm -v
which node
which npm
```

如果 `nginx` 没有安装，可以在 Ubuntu 上执行：

```bash
sudo apt update
sudo apt install -y nginx
```

> [!TIP]
> 如果你在云服务器上遇到前端构建失败，优先检查 `node -v`。较旧的 Node 版本很容易导致 `npm run build` 失败。

### 4.2 前端页面编译

在 `demo-web` 目录下执行：

```bash
cd /path/to/OpenDSN/demo-web
npm install
npm run build
```

构建完成后会生成：

```text
demo-web/dist
```

> [!IMPORTANT]
> 正式部署时，请使用 `npm run build` 生成静态页面并交给 `nginx` 托管，不要依赖 `npm run dev` 作为长期运行方案。

### 4.3 nginx 托管

仓库里已经提供了 nginx 配置样板：

```text
deploy/nginx/sites-available/opendsn
```

需要注意，这个样板文件中的 `root` 当前是按仓库默认路径写死的。例如：

```nginx
root /home/jiahao/go/src/OpenDSN/demo-web/dist;
```

如果你的 OpenDSN 实际部署路径不同，拷贝到系统 nginx 目录之前，需要先把它改成你自己的仓库绝对路径。

当前样板默认监听 `8081` 端口，这是为了方便在本地测试时与其他系统并行打开。如果你是单独正式部署 OpenDSN，可以把样板中的：

```nginx
listen 8081;
```

改成：

```nginx
listen 80;
```

#### Ubuntu 上的典型托管步骤

1. 拷贝配置

```bash
sudo cp /path/to/OpenDSN/deploy/nginx/sites-available/opendsn /etc/nginx/sites-available/opendsn
```

2. 启用站点

```bash
sudo ln -sf /etc/nginx/sites-available/opendsn /etc/nginx/sites-enabled/opendsn
```

3. 如有需要，移除默认站点

```bash
sudo rm -f /etc/nginx/sites-enabled/default
```

4. 检查配置

```bash
sudo nginx -t
```

5. 重载 nginx

```bash
sudo systemctl reload nginx
```

### 4.4 查看前端页面

1. 先启动后端和链服务：

```bash
cd /path/to/OpenDSN
bash scripts/genesis_node_start.sh
```

2. 再通过浏览器访问页面：

- 如果 nginx 样板保持 `8081`：

```text
http://<server-ip>:8081
```

- 如果你已经改成 `80`：

```text
http://<server-ip>
```

3. 可以用下面的命令检查后端和 nginx 是否正常：

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8081/api/miners
```

> [!TIP]
> 如果页面静态内容正常、但数据为空，通常优先检查：
> 1. 后端 API 是否已经启动
> 2. nginx 是否已重载
> 3. 前端是否重新执行过 `npm run build`
> 4. 浏览器是否缓存了旧资源

如果页面能打开但数据是空的，优先检查：

- 是否重新执行过 `npm run build`
- nginx 是否已经重载
- 浏览器是否缓存了旧资源
- 当前运行的是否真的是 `opendsn-api`

## 5. 退出系统

OpenDSN 仓库提供了统一的退出脚本：

```bash
bash scripts/exit.sh
```

该脚本会尝试停止：

- `listen_and_send.sh`
- 监听 `9999` 端口的 `nc` 进程
- `opendsn-api`
- `lotus-miner`
- `lotus daemon`

如果你仍然在开发模式下运行前端 `npm run dev`，该脚本也会尝试停止前端 dev server。

需要注意：

- 如果前端是通过 `nginx + dist` 托管的，`exit.sh` 不会关闭 nginx
- 如需停止 nginx，需要单独执行：

```bash
sudo systemctl stop nginx
```

脚本执行完成后，终端会输出：

```text
Exit script finished.
```

表示后端和仓库内管理的相关进程已经基本退出。
