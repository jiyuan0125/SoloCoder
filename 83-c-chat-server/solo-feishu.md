# Solo Coder 填表数据

## 83-c-chat-server — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 用 C 写一个 TCP 聊天服务器。基于 socket + pthreads，每个客户端一个连接线程。 客户端管理： 1. 客户端连接后发送 "ENTER_ROOM room_name" 加入房间，发送 "LEAVE" 离开房间 2. 发送 "MSG room_name message" 向指定房间广播消息，消息格式："[{sender}] {message}" 3. 发送 "PRIVMSG username message" 私聊指定用户，只有该用户能收到 4. 心跳：客户端每 30 秒发送 "PING"，服务器回复 "PONG"。60 秒没收到任何数据则断开该客户端 房间管理： 5. 房间所有人离开后不立即销毁，保留最近 100 条消息历史记录。如果有新用户加入则恢复 6. 新用户加入房间后，收到该房间的历史消息（最多最近 50 条） 7. 房间消息历史超过 100 条时，新消息进来丢弃最旧的 管理命令： 8. "LIST" 列出所有房间和每个房间的人数 9. "WHO room_name" 列出房间里所有人的用户名 10. "KICK room_name username" 踢出指定用户（只有房间创建者可以踢人） 协议格式： 11. 所有命令和响应都是文本行，以 \n 结尾 12. 服务器主动推送的消息也以 \n 结尾 13. 错误响应格式："ERROR: {description}" 代码分 server.c、room.c、client.c 几个文件，gcc -lpthread 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：add_client_to_room 在客户端重复 ENTER_ROOM 同一房间时会死锁——函数先 lock room->mutex，然后调 remove_client_from_room 尝试再 lock 同一个 mutex，实测客户端 hung 住服务器线程阻塞。check_timeouts 踢掉超时客户端后调用 remove_client 释放了 Client 结构体，但该客户端的 handler 线程 recv 返回后还会再调 remove_client 造成 double-free。find_room 遍历 rooms_head 链表时未加锁，多线程并发调用存在数据竞争。过程不满意：多线程并发安全是 socket+pthreads 服务器的核心要求，写完后没有测试重复 ENTER_ROOM 和超时踢人后线程退出的场景 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 83-c-chat-server |

---

## 83-c-chat-server — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我跑了一下发现几个问题。同一个客户端对同一个房间连发两次 ENTER_ROOM 会卡死，客户端 hung 住服务器线程也跟着阻塞了。还有 check_timeouts 踢掉超时客户端之后，那个客户端的 handler 线程 recv 返回后还会再调一次 remove_client，之前那个已经 free 过了，这应该是 double-free。另外 find_room 遍历房间链表的时候没加锁，多个线程同时调的话可能有问题。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：handle_enter_room 在持有 rooms_mutex 的情况下调用 find_room，find_room 内部再次 lock 同一个非递归 mutex（PTHREAD_MUTEX_INITIALIZER），导致每个 ENTER_ROOM 命令都永久死锁，服务器完全不可用。过程不满意：修复 find_room 数据竞争时没有检查已有调用方是否已持有该锁，改完后没有实际测试 ENTER_ROOM 命令能否正常执行 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 83-c-chat-server |

---

## 83-c-chat-server — 第 3 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 3 |
| User Prompt | ENTER_ROOM 命令一发就卡死了，服务器线程 hang 住。我看了一下代码，handle_enter_room 里先 pthread_mutex_lock(&rooms_mutex)，然后调 find_room，但 find_room 里面又 pthread_mutex_lock(&rooms_mutex)。rooms_mutex 是非递归锁，同一线程重复加锁就死锁了。R1 修 find_room 加锁的时候没注意到 handle_enter_room 已经持有了这个锁。还有 add_client_to_room 里也调了 find_room，如果调用链上有 rooms_mutex 也会死锁。另外 remove_client 里面也调了 find_room，需要检查所有调用路径。请统一梳理所有 find_room 的调用方，确保不存在嵌套加 rooms_mutex 的情况。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 |  |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 83-c-chat-server |

---
