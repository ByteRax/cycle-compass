# 周期罗盘 CycleCompass 迁移问题追踪与闭环清单 (Issue Tracker)

| Issue ID | 发现阶段 | 问题现象与根因 | 修复方案与闭环状态 |
| :---: | :---: | :--- | :--- |
| **ISSUE-001** | Phase 0 | Python 镜像构建因 `orjson` 缺少预编译 wheel 而编译失败 | 移除 `orjson`，FastAPI 0.165+ 走 Pydantic v2 高性能直出字节流。(已闭环) |
| **ISSUE-002** | Phase 1 | `curl -I` 请求 FastAPI 接口返回 `405 Method Not Allowed` | Go Router 对所有 GET 路由同步注册 HEAD 处理支持。(已闭环) |
| **ISSUE-003** | Phase 4 | 腾讯行情接口返回 GBK 编码导致中文标的名称在 Go 中乱码 | 引入 `golang.org/x/text` 进行 GB18030 自动转码，结合 `KnownAssets` 字典精准兜底校正。(已闭环) |
| **ISSUE-004** | Phase 4 | 模拟投资流水 SQLite 查询排序与黄金样本逆序 | 修改 SQL 排序规则为 `ORDER BY id ASC`，与黄金样本严格对齐。(已闭环) |
| **ISSUE-005** | Phase 4 | 回测周定投事件描述与黄金样本存在前缀差异 | 去除定投事件描述中的多余前缀，统一格式为 `本周定投 {amount} 元买入 {target}`。(已闭环) |
| **ISSUE-006** | Phase 5 | 模拟投资账本流水在服务重启时丢失 | 引入 SQLite WAL 数据库持久化存储，首次启动自动 Seed。(已闭环) |
| **ISSUE-007** | Phase 6 | 2GB 服务器物理内存吃紧，常驻内存高达 225.8MB | 迁移 Go 纯静态二进制，内存降至 9.5MB，释放 220MB 物理内存。(已闭环) |
