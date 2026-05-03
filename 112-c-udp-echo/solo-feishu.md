# Solo Coder 填表数据

## 112-c-udp-echo — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 帮我用 C 写一个心跳检测服务。游戏客户端每隔一段时间给服务端发一条心跳包，服务端收到后原样回传。如果服务端超过一定时间没收到某个客户端的心跳，就判定它掉线了。 心跳协议用 UDP，消息格式很简单：4 字节魔数（0xDEADBEEF）+ 4 字节序列号 + 8 字节时间戳+ 20 字节客户端标识字符串。服务端收到后把整个包原封不动发回去，不改任何字节。 客户端超时判定：每个客户端有一个最后活跃时间。后台每隔 1 秒扫描一次，如果某个客户端超过 10 秒没有新的心跳包，标记为离线并从记录中移除。离线的客户端如果又发来了心跳包，当作新客户端重新注册。 服务端要有统计功能：当前在线客户端数量、累计连接过的客户端总数、收到的总包数、发送的总包数。还要记录每个客户端的 IP 和端口。 UDP 是不可靠的，客户端可能发包但服务端没收到（或者反过来）。所以客户端发送心跳时应该带序列号，如果客户端发现连续 3 个序列号的回复都没收到，才认为服务端有问题。不过这个逻辑主要在客户端，服务端这边只要确保尽量回复就行。 有个坑：UDP 包最大 65507 字节，但实际以太网上超过 1472 字节就可能被分片丢失。心跳包虽然很小，但服务端回复时要确保不会因为网络 MTU 问题丢包。 支持多客户端并发——多个客户端同时发心跳包要都能正确处理和回复。 用纯 C 写，不依赖第三方库，gcc 在 Linux 上编译通过。写个 main 启动服务，再写一个模拟客户端发心跳包验证。 代码要拆成至少 4 个源文件，按功能模块分——心跳包收发、客户端连接管理、超时检测与清理、主程序。头文件和实现文件分开。gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 网络服务 |
| 修改范围 | 全新项目 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：客户端模拟器(client_sim.c)心跳间隔机制存在严重bug——receive_ack函数通过setsockopt设置SO_RCVTIMEO为3秒，主循环中wait_loops=interval_sec*10次调用receive_ack，每次recvfrom阻塞最多3秒后才超时返回，导致实际心跳间隔远超配置值（配置1秒实际约30秒，配置2秒实际约62秒），客户端在服务端10秒超时窗口内无法发出第二个心跳包，基本功能不成立。过程不满意：client_sim.c重复定义了heartbeat_protocol.h中已有的协议结构体heartbeat_packet_t、htonll/ntohll宏、HEARTBEAT_MAGIC等常量，没有复用服务端头文件违反DRY原则。.gitignore未排除编译产物heartbeat_server和heartbeat_client。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 112-c-udp-echo |

---

## 112-c-udp-echo — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 跑了一下模拟客户端，配置1秒发一次心跳但实际30多秒才发第二个，服务端10秒就把它踢掉了。看了下代码，client_sim.c里wait_loops每次调receive_ack，receive_ack里面recvfrom的超时设了3秒，循环10次就是30秒，跟配置的间隔完全对不上。 |
| 任务类型 | Bug修复 |
| 业务领域 | 网络服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：client_sim.c仍然重复定义了heartbeat_protocol.h中已有的heartbeat_packet_t结构体、htonll/ntohll宏和HEARTBEAT_MAGIC等常量，未复用服务端头文件违反DRY原则。.gitignore未排除编译产物heartbeat_server和heartbeat_client二进制文件。过程不满意：R1已明确指出DRY违反和.gitignore缺失两个问题，本轮仅修复了心跳间隔核心bug，未一并处理这两个已知的代码质量问题。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 112-c-udp-echo |

---

## 112-c-udp-echo — 第 3 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 3 |
| User Prompt | 看了下client_sim.c，发现里面把heartbeat_protocol.h里的结构体、宏什么的又重新定义了一遍，比如heartbeat_packet_t、htonll这些，两边的定义完全一样。另外.gitignore里也没加上编译出来的heartbeat_server和heartbeat_client这两个二进制，git status的时候会看到它们。 |
| 任务类型 | Bug修复 |
| 业务领域 | 网络服务 |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 |  |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 112-c-udp-echo |

---
