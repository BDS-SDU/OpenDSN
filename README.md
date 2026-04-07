# OpenDSN

## 目录
- [1. 项目概述](#1-项目概述)
- [2. 部署指南](#2-部署指南)
- [3. 运行示例](#3-运行示例)
- [4. 系统前端部署](#4-系统前端部署)
- [5. 退出系统](#5-退出系统)

## 1. 项目概述

OpenDSN 是论文《FileDAG: A Multi-Version Decentralized Storage Network Built on DAG-Based Blockchain》(IEEE TC 2023) 的代码实现。该项目实现了一个基于 DAG 架构的去中心化存储网络，通过增量生成算法实现高效的数据去重和压缩，并结合 DAG-Rider 共识构建双层 DAG 区块链账本结构，实现多版本文件的存储、更新与检索。

Copyright (c) 2023-2025, Guo Hechuan, MIT License

## 2. 部署指南

### 2.1 系统要求

1. 硬件要求
- CPU：4 核及以上
- 内存：8GB 及以上
- 存储：支持 8MiB sectors 的存储空间
- 网络：稳定的网络连接

2. 软件要求
- 操作系统：Linux 或 macOS
- Go：1.18.1 或更高版本
- Rust：建议通过 `rustup` 安装
- 其他依赖：`git`、`jq`、`pkg-config`、`clang`、`hwloc` 等

### 2.2 环境准备与编译

1. 安装系统依赖

Ubuntu / Debian:

```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget bsdiff -y && sudo apt upgrade -y
```

macOS:

```bash
brew install go bzr jq pkg-config rustup hwloc coreutils
```

2. 安装 Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

3. 安装 Go

```bash
wget -c https://golang.org/dl/go1.18.1.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
```

4. 配置环境变量

```bash
export LOTUS_PATH=~/.lotus-local-net
export LOTUS_MINER_PATH=~/.lotus-miner-local-net
export LOTUS_SKIP_GENESIS_CHECK=_yes_
export CGO_CFLAGS_ALLOW="-D__BLST_PORTABLE__"
export CGO_CFLAGS="-D__BLST_PORTABLE__"
export IPFS_GATEWAY=https://proof-parameters.s3.cn-south-1.jdcloud-oss.com/ipfs/
```

5. 获取源码并构建

```bash
git clone https://github.com/BDS-SDU/OpenDSN.git
cd OpenDSN
make debug
```

### 2.3 启动创世节点

OpenDSN 当前推荐通过脚本一键启动第一个节点：

```bash
cd /path/to/OpenDSN
bash scripts/genesis_node_start.sh
```

这个脚本会自动完成以下工作：

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

脚本末尾会输出两条很重要的信息：

```bash
./lotus net listen
./lotus-miner net listen
```

这两条命令分别给出：

- 创世节点 daemon 的 multiaddr
- 创世节点 miner 的 multiaddr

后续添加更多节点时，需要把这两个地址记录下来。

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

其中：

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

常用的是下面两种 `deal` 方式。

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

当前样板默认监听8081端口，这是为了方便在本地测试时与其他系统并行打开。如果你是单独正式部署 OpenDSN，可以把样板中的：

```nginx
listen 8081;
```

改成：

```nginx
listen 80;
```

Ubuntu 上的典型托管步骤如下：

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
