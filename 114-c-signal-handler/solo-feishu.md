# Solo Coder 填表数据

## 114-c-signal-handler — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 帮我用 C 写一个后台服务的信号处理模块。我们的 C 服务以守护进程方式运行，需要能优雅地响应系统信号——收到停止信号时先完成正在处理的任务、关闭网络连接、写完日志，然后再退出，而不是直接被杀掉导致数据丢失。 要处理的信号：SIGTERM（正常停止）、SIGINT（Ctrl+C，调试用）、SIGHUP（重新加载配置）、SIGUSR1（切换日志文件，用于日志轮转）。每种信号对应不同的处理逻辑。 信号处理逻辑里不能做复杂操作（比如不能调用 malloc、不能写文件、不能用 printf），所以需要一个安全的机制：信号处理只设置标志位，主循环检测到标志位后在安全上下文中执行实际的清理逻辑。 优雅停机流程：收到 SIGTERM 后，设置停止标志。主循环检测到标志后：1）停止接受新请求；2）等待正在处理的请求完成（设置一个最大等待时间，比如 30 秒）；3）刷新所有缓冲区中的日志；4）关闭监听的套接字；5）释放资源；6）退出。如果等待超时了还有请求没处理完，强制退出并记录警告日志。 PID 文件管理：服务启动时创建 PID 文件（记录进程 ID），退出时删除。启动前检查 PID 文件是否已存在，如果存在且对应的进程还在运行，说明有另一个实例在跑，应该拒绝启动。 有个坑：多次快速发送 SIGTERM 可能导致清理逻辑被并发执行。需要确保清理逻辑只执行一次，即使收到多个信号也不会重复清理。 用纯 C 写，gcc 在 Linux 上编译通过。写个 main 演示：注册信号处理，模拟工作循环，发送信号观察停机流程。 代码要拆成至少 4 个源文件，按功能模块分——信号捕获与标志管理、优雅停机流程、PID 文件管理、主程序。头文件和实现文件分开。gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：main.c 主循环的 while 条件 `!signal_handler_is_stop_requested()` 导致收到 SIGTERM 后循环直接退出，循环体内的 shutdown_execute() 永远不会被调用，优雅停机流程（停止接受请求、等待请求完成、刷新日志、关闭套接字、释放资源）完全无法执行。实际测试 kill -TERM 后程序直接退出，没有任何 shutdown 步骤日志输出。main.c 中 g_request_mutex 声明了但从未使用，是死代码。编译时 chdir 返回值未检查有 warning。daemon_demo 编译产物被提交到仓库，.gitignore 未排除。过程不满意：优雅停机是整个模块的核心功能，写完后没有实际发送 SIGTERM 测试停机流程是否正常执行 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 114-c-signal-handler |

---

## 114-c-signal-handler — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 2 |
| User Prompt | 我刚跑了下这个信号处理模块，发现 kill -TERM 之后程序直接退出了，优雅停机的那些步骤（stop accepting、等请求完成、flush logs、close sockets、释放资源）一条都没执行，终端上也没看到 "Step 1: Stopping to accept" 这些日志。main.c 里 while 循环的条件 `!signal_handler_is_stop_requested()` 会在收到信号时直接退出循环，导致循环体内调用 shutdown_execute() 的代码根本跑不到。还有 main.c 里有个 g_request_mutex 声明了但没用到，编译的时候 chdir 那行有个 warning。 |
| 任务类型 | Bug修复 |
| 业务领域 | 库/SDK |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：shutdown_execute() 在 shutdown.c 中通过 pthread_mutex_trylock 获取 g_shutdown_mutex 后，在 while 循环中等待 g_active_request_count 变为 0（Step 2 等待活跃请求完成）。但 worker 线程完成请求后调用 shutdown_decrement_active_requests() 时需要 pthread_mutex_lock 获取同一把 g_shutdown_mutex 才能递减计数。shutdown_execute 持锁等待计数归零，worker 线程等锁才能递减计数，形成死锁。实际测试：无活跃请求时 kill -TERM 优雅停机 6 步全部正常执行；有活跃请求时 kill -TERM 程序永远卡在 Step 2，后续 Step 3-6 永远不会执行。过程不满意：R1 已反馈 kill -TERM 后停机步骤不执行，R2 修了主循环条件让 shutdown_execute 能被调用，但没有在有活跃请求的场景下实际测试，导致持锁等待与递减计数之间的死锁未被发现 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 114-c-signal-handler |

---
