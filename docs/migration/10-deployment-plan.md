# 周期罗盘 CycleCompass 生产发布与平滑切换方案 (Phase 7)

## 一、发布目标与环境
- **生产目标**：将 CycleCompass HTTP API 从原 Python (FastAPI) 升级为轻量高性能 Go 独立服务，实现极端低资源开销与高吞吐。
- **目标服务器**：`110.40.230.69` (Debian 13, 2 vCPU, 2GB RAM)。
- **公网 API**：`https://cycle-api.tanranran.cn`。
- **反代接入**：1Panel OpenResty (`/opt/1panel/www/conf.d/cycle-api.tanranran.cn.conf`)。

## 二、部署拓扑与端口规划
| 服务名称 | 部署形态 | 监听端口 | 状态说明 |
| :--- | :--- | :--- | :--- |
| 旧 Python API | Docker (`cycle-compass-api`) | `127.0.0.1:8000` | 当前线上运行（切流前接管全部公网流量） |
| 新 Go API | Docker (`cycle-compass-go`) | `127.0.0.1:8002` | 新建独立容器（切流前完成全量离线验收） |
| 外部接入反代 | 1Panel OpenResty | `80/443 (SSL)` | 统一入口，按 upstream 指向实施零停机切换 |

## 三、部署实施步骤

### Step 1: 代码与资源同步
1. 将本地编译并通过跨平台验证的 Linux x86_64 纯静态二进制或 Go 源码及种子数据上传至服务器 `/opt/cycle-compass/`。
2. 确保 `/opt/cycle-compass/data/` 目录具备持久化读写权限。

### Step 2: 启动 Go API 服务 (端口 8002)
1. 启动容器 `cycle-compass-go`，映射端口 `127.0.0.1:8002:8002`，挂载数据目录。
2. 检查 Go 服务的就绪状态：
   `curl -s http://127.0.0.1:8002/api/v1/health`
3. 执行 11 个核心接口的本地回归测试，确保全部返回 HTTP 200 且数据与黄金样本 100% 吻合。

### Step 3: 平滑切流 (OpenResty 毫秒级重载)
1. 备份现有反代配置：
   `cp /opt/1panel/www/conf.d/cycle-api.tanranran.cn.conf /opt/1panel/www/conf.d/cycle-api.tanranran.cn.conf.bak`
2. 修改 upstream 配置指向 `127.0.0.1:8002`：
   ```nginx
   upstream cycle_backend {
       server 127.0.0.1:8002;
       keepalive 32;
   }
   ```
3. 检查 Nginx 配置合法性并热重载：
   `docker exec 1Panel-openresty-lRvy nginx -t`
   `docker exec 1Panel-openresty-lRvy nginx -s reload`

### Step 4: 全链路公网验证
1. 通过公网发起 HTTPS 请求验证：`curl -s https://cycle-api.tanranran.cn/api/v1/health`。
2. 微信小程序前端进行核心功能走查（首页行情、罗盘周期、双策略信号、模拟投资持仓、回测试算）。

### Step 5: 旧服务退役与资源释放
1. 停止原 Python 容器 `cycle-compass-api`（保留镜像与数据，供紧急回滚备用）：
   `docker stop cycle-compass-api`
2. 观察主机内存与 Swap 释放情况（物理内存占用下降 200MB+）。
