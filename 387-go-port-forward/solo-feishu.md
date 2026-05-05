# Solo Coder 填表数据

## 387-go-port-forward — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 写一个 Go 命令行工具做简单的TCP端口转发。开发环境在本机但需要访问远程服务器上的数据库或缓存服务。 用法：./port-forward --local 3306 --remote 192.168.1.100:3306。监听本地端口接受连接后建立到远程地址的TCP连接，双向转发数据。支持同时处理多个客户端连接。客户端断开时要正确关闭远程连接不能泄漏。远程服务器不可达时给客户端返回明确的连接错误。正确处理TCP半关闭的情况。加 --verbose 参数时打印每个连接的详细信息包括连接时间和转发字节数。监听端口被占用时给出明确提示。程序收到SIGINT或SIGTERM信号时优雅关闭——等待现有连接处理完毕再退出。设置连接空闲超时——默认5分钟内没有数据传输则关闭连接，可通过 --timeout 参数调整。支持同时转发多个端口——多次指定 --local 和 --remote 参数对。每个连接的转发字节数和连接时长在verbose模式下实时打印。连接关闭时统计总转发字节数并记录到日志。加 --buffer-size 参数设置转发缓冲区大小默认4KB。 拆成两个独立程序：一个是后台常驻的服务进程，负责监听端口和双向转发数据，维护连接状态；另一个是命令行客户端，用户通过它管理转发规则和查看连接统计。两个程序各自独立 main 包，通过 TCP 通信，共享协议定义放公共包。各程序内部按职责分文件。先 go mod init 再开发，go build ./... 能编译通过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：connID 数据竞争——forwarder.go handleConnection 发送 connInfo 指针到 channel 后立即读取 connInfo.ID（line 115），但 server.go handleNewConnection 在之后才设置 connInfo.ID（line 239），实际运行 connID 几乎总是 0，导致所有并发连接共享 ID 0，stats 统计完全错误。handleAddForward TOCTOU 竞争——检查端口是否存在和添加 forwarder 之间释放了锁（server.go:154-186），并发 AddForward 同一端口可能都成功。pfclient 使用 Go flag 包导致 flags 必须在命令之前（如 --local 3306 --remote host add），与标准 CLI 约定（add --local 3306 --remote host）相反。过程不满意：connID 的读写竞争是基本的并发安全错误，写完后没有测试多连接场景下 stats 的正确性 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 387-go-port-forward |

---
