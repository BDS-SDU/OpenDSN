# OpenDSN

## 目录
- [1. 项目概述](#1-项目概述)
- [2. 部署指南](#2-部署指南)
- [3. 运行示例](#3-运行示例)

## 1. 项目概述
OpenDSN 是论文《FileDAG: A Multi-Version Decentralized Storage Network Built on DAG-Based Blockchain》(IEEE TC 2023) 的代码实现。该项目实现了一个基于 DAG 架构的去中心化存储网络，通过结合增量生成算法实现高效的数据去重和压缩，并结合DAG-Rider共识算法构建双层 DAG 区块链账本结构，实现高效的文件版本管理，支持版本回滚、分支管理和版本合并等高级功能。本项目在确保数据一致性、压缩优化与扩展性的同时，解决了传统去中心化存储网络在版本管理和效率方面的核心难题。

Copyright (c) 2023-2025, Guo Hechuan, MIT License

## 2. 部署指南

### 2.1 系统要求
OpenDSN 系统对运行环境有特定的要求：

1. 硬件要求
- CPU：4 核及以上
- 内存：8GB 及以上
- 存储：支持 8MiB sectors 的存储空间
- 网络：稳定的网络连接

2. 软件环境
- 操作系统：支持 Linux 或 MacOS
- Go 版本：1.18.1 或更高版本
- Rust 环境：需要安装 rustup
- 其他依赖：根据操作系统不同，需要安装相应的系统依赖包

### 2.2 安装步骤

1. 环境准备
根据操作系统安装必要的系统依赖：

Linux (Ubuntu/Debian):
```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget -y && sudo apt upgrade -y
```

MacOS:
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

### 2.3 启动节点

1. 启动第一个节点
```bash
./lotus daemon --bootstrap=false
```

2. 创建钱包
```bash
./lotus wallet new
```

3. 设置存储提供者
```bash
./lotus-miner init --owner=<address> --worker=<address> --sector-size=8MiB
```

4. 启动存储提供者
```bash
./lotus-miner run
```
### 2.4 常见问题

1. 系统依赖问题
- 确保已安装所有必要的系统依赖
- 检查 Go 和 Rust 版本是否符合要求
- 验证环境变量是否正确设置

2. 节点启动问题
- 检查端口是否被占用
- 验证创世文件是否正确
- 确保有足够的系统资源

3. 存储提供者问题
- 确保钱包地址有足够的资金
- 验证存储路径权限是否正确
- 检查存储空间是否充足

4. 网络连接问题
- 检查防火墙设置
- 验证节点地址是否正确
- 确保网络连接稳定

## 3. 运行示例

### 3.1 导入本地文件
要将本地文件导入到 OpenDSN 网络，使用以下命令：
```bash
./lotus client import <文件名>
```
执行后，系统会返回文件的 CID（内容标识符）。

### 3.2 文件版本管理

1. 创建新版本
```bash
./lotus client version create <文件CID>
```

2. 切换版本
```bash
./lotus client version switch <版本CID>
```

3. 查看版本历史
```bash
./lotus client version history <文件CID>
```
### 3.3 创建存储交易
OpenDSN 支持两种方式创建存储交易：交互式和非交互式。

1. 交互式创建交易
使用以下命令启动交互式交易创建：
```bash
./lotus client deal
```
按照提示依次输入：
1. 要存储的文件 CID
2. 存储期限（天数），例如输入 60 表示存储 60 天
3. 是否为 Filecoin Plus 交易（通常选择否）
4. 矿工 ID（多个 ID 用空格分隔）
5. 确认交易（输入 yes）

完成后，系统会返回交易 CID。

2. 非交互式创建交易（推荐）
使用以下命令直接创建交易：
```bash
./lotus client deal [dataCid miner price duration]
```
参数说明：
- dataCid：文件 CID
- miner：矿工 ID
- price：存储价格
- duration：存储期限（秒）

示例：
```bash
./lotus client deal bafylfkjaldfkjasldjflas t01000 0.0026 518400
```

### 3.4 交易管理

1. 查看交易状态
```bash
./lotus client list-deals
```

2. 查看交易详情
```bash
./lotus client get-deal <DealCID>
```

### 3.5 文件检索

1. 检索文件
```bash
./lotus client retrieve <DealCID> <输出路径>
```

2. 查看检索状态
```bash
./lotus client list-retrievals
```

### 3.6 常见问题

1. 交易创建失败
- 检查文件 CID 是否正确
- 确认矿工 ID 是否有效
- 验证存储价格是否合理
- 确保存储期限在有效范围内

2. 文件检索问题
- 确认交易是否处于活跃状态
- 检查网络连接是否正常
- 验证输出路径是否有写入权限

3. 交易状态异常
- 检查矿工节点是否在线
- 确认存储空间是否充足
- 验证交易参数是否正确
