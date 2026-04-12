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

OpenDSN is a decentralized storage network designed for multi-version file management. It provides file import, initial-version upload, incremental updates, version-chain lookup, targeted version retrieval, miner status collection, proof visualization, and an integrated backend/frontend demo workflow.

It can be used both as a local multi-node experimental environment and as a cloud deployment target, where the genesis node, storage nodes, backend API, and frontend page are launched through scripts.

## 🎬 Demo

OpenDSN's web interface focuses on three core workflows: file information lookup, file upload, and file retrieval.

### File Information

<p align="center">
  <img src="./assets/readme/file_infomation.gif" alt="OpenDSN file information demo" width="900">
</p>

Query all recorded roots in the system and inspect the complete version chain of a selected root.

### File Upload

<p align="center">
  <img src="./assets/readme/file_upload.gif" alt="OpenDSN file upload demo" width="900">
</p>

Import a local file to obtain its root CID, then submit either an initial-version deal or an update deal through the frontend.

### File Retrieval

<p align="center">
  <img src="./assets/readme/file_retrieval.gif" alt="OpenDSN file retrieval demo" width="900">
</p>

Use a root or head CID to retrieve the requested file version and write the reconstructed output back to the local workspace.

## ✨ Features

- **Multi-version file management**: maintain and inspect roots, heads, and complete version chains.
- **Incremental update workflow**: supports both initial version creation and subsequent version updates.
- **Visual observability**: the frontend displays storage node status, the latest proof information, and version-related operation results.
- **Script-oriented management**: includes startup scripts, exit scripts, and nginx deployment templates for both local and cloud environments.
- **Demo friendly**: suitable for experiments, classroom demonstrations, and system validation.

## Table of Contents
- [1. Overview](#1-overview)
- [2. Deployment Guide](#2-deployment-guide)
- [3. Usage Examples](#3-usage-examples)
- [4. Frontend Deployment](#4-frontend-deployment)
- [5. Exit the System](#5-exit-the-system)

## 1. Overview

OpenDSN is organized around the following core components:

- **Chain and storage nodes**
  - `lotus daemon` and `lotus-miner` provide the chain service and storage provider runtime.
- **Node startup scripts**
  - `scripts/genesis_node_start.sh` starts the genesis node.
  - `scripts/node_start.sh` connects regular nodes to an existing network.
- **Backend API**
  - `cmd/opendsn-api` serves the frontend and aggregates miner information, proof information, and multi-version file operations.
- **Frontend**
  - `demo-web` provides the web UI and can be served through `nginx`.
- **Exit script**
  - `scripts/exit.sh` shuts down the core processes started from this repository.

OpenDSN currently targets `8MiB` sectors, a script-driven startup flow, and a demo-oriented frontend, making it easy to reproduce a complete multi-version DSN system in a lab intranet, on cloud servers, or in a classroom presentation environment.

Copyright (c) 2023-2025, Guo Hechuan, MIT License

## 2. Deployment Guide

### 2.1 System Requirements

<table>
  <tr>
    <th align="left">Category</th>
    <th align="left">Requirement</th>
  </tr>
  <tr>
    <td>CPU</td>
    <td>At least 4 cores</td>
  </tr>
  <tr>
    <td>Memory</td>
    <td>At least 8GB</td>
  </tr>
  <tr>
    <td>Storage</td>
    <td>Enough storage space for 8MiB sectors</td>
  </tr>
  <tr>
    <td>Network</td>
    <td>Stable network connectivity</td>
  </tr>
  <tr>
    <td>Operating System</td>
    <td>Linux or macOS</td>
  </tr>
  <tr>
    <td>Go</td>
    <td>1.18.1 or later</td>
  </tr>
  <tr>
    <td>Rust</td>
    <td>Recommended to install through <code>rustup</code></td>
  </tr>
  <tr>
    <td>Other dependencies</td>
    <td><code>git</code>, <code>jq</code>, <code>pkg-config</code>, <code>clang</code>, <code>hwloc</code>, <code>bsdiff</code> / <code>bspatch</code>, etc.</td>
  </tr>
</table>

> [!TIP]
> If you plan to deploy a multi-node network on cloud servers, prepare multiple machines inside the same private network first, and verify private-network connectivity and security group rules in advance.

### 2.2 Environment Setup and Build

#### 2.2.1 Install system dependencies

Ubuntu / Debian:

```bash
sudo apt install mesa-opencl-icd ocl-icd-opencl-dev gcc git bzr jq pkg-config curl clang build-essential hwloc libhwloc-dev wget bsdiff -y && sudo apt upgrade -y
```

macOS:

```bash
brew install go bzr jq pkg-config rustup hwloc coreutils bsdiff
```

#### 2.2.2 Install Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

#### 2.2.3 Install Go

```bash
wget -c https://golang.org/dl/go1.18.1.linux-amd64.tar.gz -O - | sudo tar -xz -C /usr/local
```

#### 2.2.4 Configure environment variables

```bash
export LOTUS_PATH=~/.lotus-local-net
export LOTUS_MINER_PATH=~/.lotus-miner-local-net
export LOTUS_SKIP_GENESIS_CHECK=_yes_
export CGO_CFLAGS_ALLOW="-D__BLST_PORTABLE__"
export CGO_CFLAGS="-D__BLST_PORTABLE__"
export IPFS_GATEWAY=https://proof-parameters.s3.cn-south-1.jdcloud-oss.com/ipfs/
```

#### 2.2.5 Clone the source code and build

```bash
git clone https://github.com/BDS-SDU/OpenDSN.git
cd OpenDSN
make debug
```

> [!IMPORTANT]
> After the build finishes, make sure `lotus`, `lotus-miner`, `lotus-seed`, and the other required binaries are present in the repository root before continuing.

### 2.3 Start the Genesis Node

OpenDSN currently recommends starting the first node through the script:

```bash
cd /path/to/OpenDSN
bash scripts/genesis_node_start.sh
```

#### What the script does automatically

- Cleans old local network data and logs
- Fetches 8MiB parameter files
- Pre-seals 2 genesis sectors
- Generates `localnet.json` and `devgen.car`
- Starts the genesis `lotus daemon`
- Imports the genesis miner key
- Initializes the genesis miner `t01000`
- Starts `lotus-miner`
- Starts `listen_and_send.sh`
- Starts `opendsn-api`

#### Information you should record after the script finishes

```bash
./lotus net listen
./lotus-miner net listen
```

These two commands provide:

- the daemon multiaddr of the genesis node
- the miner multiaddr of the genesis node

You will need both addresses when adding more nodes later.

> [!TIP]
> `genesis_node_start.sh` already includes the startup logic for the genesis node, the funding listener, and the backend API. Prefer using this script instead of running the commands one by one.

### 2.4 Add More Nodes

To add a regular node, use:

```bash
bash scripts/node_start.sh <daemon_multiaddr> <miner_multiaddr> <genesis_ip> [local_ip]
```

Before running the script, copy the genesis node's:

```text
devgen.car
```

to the repository root of the new node.

#### Parameter description

- `daemon_multiaddr`
  - obtained from `./lotus net listen` on the genesis node
- `miner_multiaddr`
  - obtained from `./lotus-miner net listen` on the genesis node
- `genesis_ip`
  - the IP address of the machine where the genesis node is running
  - the new node sends its wallet address and local IP to `listen_and_send.sh` on this address
- `local_ip`
  - optional
  - if omitted, the script detects the local IP automatically
  - if the machine has multiple NICs or multiple addresses, it is safer to provide it explicitly

#### What `node_start.sh` does automatically

`node_start.sh` automates the full process of connecting a regular node and turning it into a storage provider. Internally, it will:

- start a local `lotus daemon`
- connect to the genesis node daemon and miner
- create a wallet address
- send the wallet address and local IP to the genesis node
- wait for initial funding from the genesis node
- initialize `lotus-miner`
- start `lotus-miner`
- configure the `sealed` and `unseal` storage paths

A typical example:

```bash
bash scripts/node_start.sh \
  /ip4/192.168.1.10/tcp/1234/p2p/12D3KooW... \
  /ip4/192.168.1.10/tcp/2345/p2p/12D3KooW... \
  192.168.1.10 \
  192.168.1.11
```

If the last `local_ip` argument is omitted, the script first tries to detect a public IPv4 address. If no public address is found, it falls back to a private IPv4 address.

> [!IMPORTANT]
> Before a new node starts, it must already have `devgen.car`, and it must be able to reach the genesis node's `9999` port as well as the lotus daemon and lotus miner communication addresses.

### 2.5 FAQ

1. Genesis node startup fails
- Check whether `go`, `rustup`, and all system dependencies are installed correctly
- Check whether `make debug` completed successfully
- Check whether any required ports are already occupied by old processes

2. A new node cannot join the network
- Make sure `devgen.car` has already been copied to the new node
- Make sure `daemon_multiaddr` and `miner_multiaddr` were copied completely
- Make sure the genesis node's `9999` port and node communication ports are reachable over the private network

3. A new node does not receive funding from the genesis node
- Confirm that `listen_and_send.sh` is running on the genesis node
- Confirm that `genesis_ip` is correct
- Confirm that the wallet address and local IP sent by the new node are not blocked by the firewall

4. Incorrect IP detected on multi-NIC machines
- Both genesis and regular nodes now prefer public IPv4 addresses
- If your deployment should use private addresses instead, pass `local_ip` explicitly to `node_start.sh`

## 3. Usage Examples

### 3.1 Create a storage deal: `client import + client deal`

OpenDSN file upload typically consists of two steps:

1. use `lotus client import` to get the file CID / root
2. use `lotus client deal` to upload that root to the network

#### Import a local file

```bash
./lotus client import <inputPath>
```

For example:

```bash
./lotus client import v1
```

The command prints the file's root CID. That root is the key input for later `client deal`, `list-version`, and `retrieve-version` operations.

#### Create the initial version

```bash
./lotus client deal --create <rootCid> <miner> <price> <duration>
```

For example:

```bash
./lotus client deal --create bafykbzaceaanrd4jqhaaalcbsftfzzc7wfi6o7qk27hnjlhfmv72qeeokgtni t01000 0.026 518400
```

This marks the file as the first version in a version chain and generates local metadata including:

- `fileroots.log`
- `<rootCid>_meta`

#### Upload a new version

```bash
./lotus client deal --update --previous <previousRoot> <newRoot> <miner> <price> <duration>
```

For example:

```bash
./lotus client deal --update --previous bafykbzaceaanrd4jqhaaalcbsftfzzc7wfi6o7qk27hnjlhfmv72qeeokgtni bafk2bzaceaox2bgqudb7o2wq7ra444fshsb7iedtic4htq2c63t4ysr77uzfk t01000 0.026 518400
```

This appends the new version to an existing version chain and updates the corresponding metadata file.

### 3.2 Retrieve a file: `retrieve-version`

OpenDSN restores a specific version through a target root or head:

```bash
./lotus client retrieve-version <dataCid> <outputPath>
```

For example:

```bash
./lotus client retrieve-version bafk2bzacebjm5xaqy7sdklvzyhojz4swfgjqhh5lxzroa7zej4slm3bmgopzi retrieve_v3
```

This command will:

- read the local `<cid>_meta` chain backwards from the target version
- reconstruct the full version chain
- fetch the root and patch data in order
- use `bspatch` to rebuild the requested version

### 3.3 List all roots: `list-root`

To list all file roots currently recorded in the DSN:

```bash
./lotus client list-root
```

This command reads:

```text
fileroots.log
```

and prints output similar to:

```text
ROOT 0: <rootCid>
ROOT 1: <rootCid>
```

### 3.4 View the version chain: `list-version`

To inspect the full version chain of a file starting from a root:

```bash
./lotus client list-version <rootCid>
```

For example:

```bash
./lotus client list-version bafykbzaceaanrd4jqhaaalcbsftfzzc7wfi6o7qk27hnjlhfmv72qeeokgtni
```

The command prints output similar to:

```text
ROOT(v1): <rootCid>
HEAD(v2): <headCid>
HEAD(v3): <headCid>
```

### 3.5 Deal management

1. View deal status

```bash
./lotus client list-deals
```

2. View deal details

```bash
./lotus client get-deal <DealCID>
```

### 3.6 FAQ

1. `client deal --create` fails
- Check whether the root CID is valid
- Check whether the miner ID is correct
- Check whether the price and duration satisfy on-chain constraints

2. `client deal --update` fails
- Check whether `--previous` points to an existing version
- Check whether `<previousCid>_meta` exists
- Check whether the new root was already obtained through `client import`

3. `retrieve-version` fails
- Check whether the target CID has a complete `<cid>_meta`
- Check whether the root and patch data were all uploaded successfully
- Check whether the output path already exists

4. `list-root` or `list-version` returns nothing
- This usually means a complete version metadata chain has not been formed in the current directory yet
- Run `client deal --create` first, then continue with `client deal --update`

## 4. Frontend Deployment

### 4.1 Install dependencies

The frontend depends on:

- Node.js
- npm
- nginx

First check whether they are already installed:

```bash
node -v
npm -v
nginx -v
```

If `node -v` and `npm -v` show that Node.js or npm is missing, or if the installed version is too old to build the frontend successfully, you can install a newer Node.js version with `nvm`:

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

If `nginx` is not installed, on Ubuntu you can run:

```bash
sudo apt update
sudo apt install -y nginx
```

> [!TIP]
> If the frontend build fails on a cloud server, check `node -v` first. Older Node.js versions are a common cause of `npm run build` failures.

### 4.2 Build the frontend

Run the following inside `demo-web`:

```bash
cd /path/to/OpenDSN/demo-web
npm install
npm run build
```

After the build completes, the output directory will be:

```text
demo-web/dist
```

> [!IMPORTANT]
> For production deployment, use `npm run build` and serve the generated static files with `nginx`. Do not rely on `npm run dev` as a long-running deployment method.

### 4.3 Serve with nginx

The repository already provides an nginx configuration template:

```text
deploy/nginx/sites-available/opendsn
```

Please note that the `root` path in this template is currently hard-coded according to the default repository location. For example:

```nginx
root /home/jiahao/go/src/OpenDSN/demo-web/dist;
```

If your actual OpenDSN deployment path is different, update this path to your own repository absolute path before copying the file into the system nginx directory.

The current template listens on port `8081` by default so that it can coexist with other systems during local testing. If you are deploying OpenDSN alone in production, you can change:

```nginx
listen 8081;
```

to:

```nginx
listen 80;
```

#### Typical Ubuntu hosting steps

1. Copy the configuration

```bash
sudo cp /path/to/OpenDSN/deploy/nginx/sites-available/opendsn /etc/nginx/sites-available/opendsn
```

2. Enable the site

```bash
sudo ln -sf /etc/nginx/sites-available/opendsn /etc/nginx/sites-enabled/opendsn
```

3. If needed, remove the default site

```bash
sudo rm -f /etc/nginx/sites-enabled/default
```

4. Test the configuration

```bash
sudo nginx -t
```

5. Reload nginx

```bash
sudo systemctl reload nginx
```

### 4.4 View the frontend page

1. Start the backend and chain services:

```bash
cd /path/to/OpenDSN
bash scripts/genesis_node_start.sh
```

2. Open the page in a browser:

- If the nginx template still uses `8081`:

```text
http://<server-ip>:8081
```

- If you changed it to `80`:

```text
http://<server-ip>
```

3. You can use the following commands to verify that both the backend and nginx are working:

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8081/api/miners
```

> [!TIP]
> If the page loads but no data appears, check these items first:
> 1. whether the backend API has started
> 2. whether nginx has been reloaded
> 3. whether `npm run build` was executed again after the latest frontend changes
> 4. whether the browser is using cached frontend assets

If the page opens but data is still empty, also check:

- whether `npm run build` was executed again
- whether nginx has already been reloaded
- whether the browser is caching old assets
- whether the running backend is actually `opendsn-api`

## 5. Exit the System

OpenDSN provides a unified exit script:

```bash
bash scripts/exit.sh
```

This script attempts to stop:

- `listen_and_send.sh`
- the `nc` process listening on port `9999`
- `opendsn-api`
- `lotus-miner`
- `lotus daemon`

If you are still running the frontend in development mode with `npm run dev`, the script will also try to stop the frontend dev server.

#### Notes

Please note:

- If the frontend is served through `nginx + dist`, `exit.sh` does not stop nginx
- To stop nginx manually, run:

```bash
sudo systemctl stop nginx
```

After the script finishes, the terminal will print:

```text
Exit script finished.
```

which indicates that the core processes managed by this repository have mostly been stopped.
