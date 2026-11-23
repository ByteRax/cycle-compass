# 周期罗盘 CycleCompass 紧急回滚预案 (Phase 7)

## 一、触发回滚的判定条件
在切换至 Go API 后，若发生以下任意一种情况，立即触发 30 秒紧急回滚机制：
1. 公网接口请求出现大面积 500、502 或不可恢复异常。
2. 微信小程序核心业务流程受阻（如行情无法加载、策略信号计算不一致）。
3. 发现数据持久化异常或模拟账户账本数据错乱。
4. Go API 进程发生 Panic 并导致容器频繁 Crash 重启。

## 二、30 秒快速回滚流程

### 步骤 1: 确保 Python 旧容器恢复运行（若已停止）
```bash
docker start cycle-compass-api
# 检查 8000 端口存活
curl -s http://127.0.0.1:8000/api/v1/health
```

### 步骤 2: 恢复 OpenResty 反向代理配置
```bash
# 直接从备份覆盖，将 upstream 恢复至 8000
cp /opt/1panel/www/conf.d/cycle-api.tanranran.cn.conf.bak /opt/1panel/www/conf.d/cycle-api.tanranran.cn.conf
# 验证配置并执行毫秒级重载
docker exec 1Panel-openresty-lRvy nginx -t
docker exec 1Panel-openresty-lRvy nginx -s reload
```

### 步骤 3: 验证线上公网服务已回滚
```bash
curl -I https://cycle-api.tanranran.cn/api/v1/health
```

### 步骤 4: 保留现场并进行离线复盘
- 导出 Go API 崩溃日志：`docker logs cycle-compass-go > /tmp/go-crash.log`。
- 保留 SQLite 数据库现场文件进行只读离线分析。
