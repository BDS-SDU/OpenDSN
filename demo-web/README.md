# OpenDSN Frontend Deployment Guide

这个目录包含 OpenDSN 的前端页面代码。前端基于 `React + Vite` 构建，支持两种运行方式：

- 开发调试：`npm run dev`
- 正式部署：`npm run build` 后交给 `nginx` 托管

当前仓库已经按正式部署方式做了适配：

- 生产环境 API 地址使用相对路径 `/api`
- `scripts/genesis_node_start.sh` 只启动后端 API，不再默认启动前端 dev server
- 仓库内提供了 nginx 站点配置样板：
  - [`../deploy/nginx/sites-available/opendsn`](../deploy/nginx/sites-available/opendsn)

## 1. 前端依赖

先检查当前机器是否已经安装了前端运行环境：

```bash
node -v
npm -v
```

如果这两条命令都提示未安装，可以在 Ubuntu 上执行：

```bash
sudo apt update
sudo apt install -y nodejs npm
```

安装完成后再次检查：

```bash
node -v
npm -v
```

确认都能正常输出版本号之后，再在 `demo-web` 目录执行：

```bash
npm install
```

说明：

- `npm install` 只需要在这台机器、这个前端目录里执行一次
- 如果以后删除了 `node_modules`，或者换了新机器，需要重新执行

## 2. 环境变量说明

本目录下有两个环境变量文件：

- [`.env`](./.env)
- [`.env.production`](./.env.production)

用途如下：

### 开发环境

`.env` 用于 `npm run dev`，例如：

```env
VITE_API_BASE_URL=http://100.64.126.151:8080
```

这表示开发模式下前端直接请求指定地址的后端 API。

### 生产环境

`.env.production` 用于 `npm run build`，当前内容为：

```env
VITE_API_BASE_URL=/api
```

这表示正式部署后，前端通过 nginx 的 `/api` 反向代理访问后端，而不是写死某个 IP。

## 3. 开发模式启动

如果只是本地调试页面效果，可以这样启动：

```bash
cd /path/to/OpenDSN/demo-web
npm install
npm run dev
```

启动成功后，Vite 会输出类似：

```text
Local:   http://localhost:5173/
Network: http://<server-ip>:5173/
```

说明：

- 开发模式适合调试
- 不推荐作为正式展示或长期部署方式

## 4. 正式部署方式

正式部署建议使用：

- `npm run build` 构建前端
- `nginx` 托管 `dist/`
- 后端 API 继续由 `scripts/genesis_node_start.sh` 启动

### 第一步：构建前端

```bash
cd /path/to/OpenDSN/demo-web
npm install
npm run build
```

构建完成后会生成：

```text
demo-web/dist/
```

### 第二步：启动后端 API

在仓库根目录执行：

```bash
cd /path/to/OpenDSN
bash scripts/genesis_node_start.sh
```

当前脚本会以本机监听方式启动后端：

```text
OPENDSN_API_ADDR=127.0.0.1:8080
```

同时脚本会动态计算：

```text
OPENDSN_REPO_ROOT
```

所以仓库不需要固定部署在某个写死的路径下。

这样外部用户不会直接访问后端端口，而是通过 nginx 的 `/api` 转发访问。

## 5. nginx 配置

仓库里已经提供了站点配置样板：

- [`../deploy/nginx/sites-available/opendsn`](../deploy/nginx/sites-available/opendsn)

当前这份样板默认：

- 前端页面入口：`80`
- 后端代理目标：`127.0.0.1:8080`

### 5.1 拷贝配置

在 Ubuntu 上执行：

```bash
sudo cp /path/to/OpenDSN/deploy/nginx/sites-available/opendsn /etc/nginx/sites-available/opendsn
```

### 5.2 启用站点

```bash
sudo ln -sf /etc/nginx/sites-available/opendsn /etc/nginx/sites-enabled/opendsn
```

### 5.3 如有需要，移除默认站点

如果默认站点冲突，可以执行：

```bash
sudo rm -f /etc/nginx/sites-enabled/default
```

### 5.4 检查配置

```bash
sudo nginx -t
```

如果输出包含：

```text
syntax is ok
test is successful
```

则说明配置可用。

### 5.5 重载 nginx

```bash
sudo systemctl reload nginx
```

## 6. 访问地址

按当前仓库样板配置，OpenDSN 页面默认访问地址为：

```text
http://<server-ip>
```

例如：

```text
http://100.64.126.151
```

如果以后你希望和另一个系统在同一台机器上并行托管，也可以把 nginx 配置中的：

```nginx
listen 80;
```

改成另一个端口，例如：

```nginx
listen 8082;
```

然后通过：

```text
http://<server-ip>:8082
```

访问。

## 7. 验证方法

### 验证前端构建是否成功

```bash
cd /path/to/OpenDSN/demo-web
npm run build
```

### 验证后端 API 是否已启动

```bash
curl http://127.0.0.1:8080/healthz
```

正常应返回：

```json
{"status":"ok"}
```

### 验证 nginx 是否正确代理 API

如果当前 nginx 监听 `80`，执行：

```bash
curl http://127.0.0.1/api/miners
```

如果能返回 JSON，说明：

- nginx 静态页面入口正常
- `/api` 转发正常

## 8. 常见问题

### 1. 页面能打开，但数据全是 404

请重点检查：

- 是否已经执行了 `npm run build`
- nginx 是否已重载
- 浏览器是否缓存了旧前端资源
- 当前前端是否是生产构建版本

当前仓库的前端代码已经改成“生产环境不重复拼 `/api`”，正常情况下不会再出现 `/api/api/...` 的问题。

### 2. 页面能打开，但拿不到 API 数据

请检查：

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1/api/miners
```

如果第一条失败，说明后端没有启动。  
如果第一条成功、第二条失败，说明 nginx 代理配置有问题。

### 3. 前端自己挂掉了

开发模式下 `npm run dev` 依赖当前进程/会话，关闭终端或会话中断时可能退出。  
正式部署时建议始终使用：

- `npm run build`
- `nginx`

而不是长期依赖 `vite dev server`。
