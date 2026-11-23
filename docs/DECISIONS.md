# 周期罗盘 CycleCompass 技术决策记录 (Architecture Decision Records)

## ADR-001：Go Web 框架与路由器选型
- **决策**：选用 Go 官方标准库 `net/http` + `github.com/go-chi/chi/v5` 路由器。
- **理由**：Chi 100% 兼容 Go 官方 `http.Handler` 标准，零过度抽象，支持精准的中间件编排（Gzip、Recoverer、RealIP），内存与 CPU 开销微乎其微。

## ADR-002：嵌入式数据库选型
- **决策**：选用纯 Go 实现的 `modernc.org/sqlite`，关闭 CGO 编译。
- **理由**：
  1. 纯 Go 实现无需安装 gcc/musl，跨平台编译（Mac -> Linux amd64）零外部 C 库依赖，完全消除动态链接库丢失风险。
  2. 启用 WAL 模式与 5000ms BusyTimeout，读写性能达纳秒级，彻底解决原 Python 进程内存持仓在重启时数据易失的问题。

## ADR-003：生产切流机制
- **决策**：采用独立端口 `8002` + 1Panel OpenResty 热重载切流。
- **理由**：
  1. 不直接覆写或停机旧服务，新旧服务在服务器内部并存（`8000` 与 `8002`）。
  2. 在本地通过完整 11 个接口回归后，修改 Nginx upstream 并执行 `nginx -s reload`，实现毫秒级零丢包热切流。
  3. 保留旧容器镜像在磁盘，具备 30 秒内秒级一键回滚能力。

## ADR-004：生产进程托管
- **决策**：采用 systemd 原生托管单文件纯静态 Linux 二进制。
- **理由**：进一步省去 Docker 守护进程与 bridge 虚拟网卡的额外内存开销，服务常驻物理内存仅 9.5MB，开机自启且具备 3 秒 Crash 自愈能力。
