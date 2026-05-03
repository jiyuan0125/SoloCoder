# Solo Coder 填表数据

## 86-go-raft-node — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 用 Go 实现一个简化版 Raft 共识算法。5 个节点组成集群（端口 5001-5005），用 HTTP RPC 模拟网络通信。 角色和选举： 1. 三个角色：Follower、Candidate、Leader。初始全部是 Follower 2. Follower 在 election timeout 内没收到 Leader 心跳就变成 Candidate，发起选举 3. election timeout 范围 150-300ms，每个节点启动时随机固定一个值（不是每次选举都随机），节点间不能一样 4. Candidate 获得 majority（>= N/2+1）票数成为 Leader 日志复制： 5. Leader 接收客户端的 Set(key, value) 请求，追加到本地日志后向所有 Follower 发 AppendEntries RPC 6. 日志条目包含 term、index、command（SET key value）和已提交标记 7. 日志提交：Leader 的日志条目被 majority 确认后标记为 committed，然后 apply 到状态机 Leader 切换： 8. 新 Leader 上任后，旧 Leader 上未提交的日志条目要被截断（以新 Leader 的日志为准） 9. Leader 定期（心跳间隔 50ms）向所有 Follower 发 AppendEntries（即使没有新日志也发心跳） 快照： 10. 日志条目超过 100 条时可以创建快照（保存当前状态），快照后的日志可以截断 11. 快照创建过程中不能阻塞正常的 AppendEntries 处理 代码分 raft.go、rpc.go、state_machine.go 几个 package，go build 能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：raft.go 的 run() 函数（第154-163行）在 switch 中调用了 runCandidate() 和 runLeader() 两个未定义函数，go build 编译直接报 undefined 错误。产物不满意：项目缺少 go.mod 文件，Go modules 模式下无法编译。产物不满意：createSnapshot() 中 snapshotData 只通过 log.Printf 打印后丢弃（raft.go:491），从未持久化到磁盘，节点重启后快照数据不存在。产物不满意：定义了 InstallSnapshotArgs/InstallSnapshotReply 类型（raft.go:61-71）但从未实现 InstallSnapshot RPC 的 HTTP handler 和发送逻辑，follower 在 leader 截断日志后无法同步快照。产物不满意：createSnapshot() 第一阶段持有 rn.mu 期间调用 stateMachine.CreateSnapshot()（raft.go:430-448），会阻塞 AppendEntries 的锁获取，不符合需求11"快照创建过程中不能阻塞正常的 AppendEntries 处理"。过程不满意：代码存在明显的编译错误（调用未定义函数），说明写完后没有实际运行 go build 验证 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 86-go-raft-node |

---

## 86-go-raft-node — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我试着编译了一下，go build 直接报错说 runCandidate 和 runLeader 两个函数未定义，raft.go 里的 run() 方法在 switch 里调了这两个函数但代码里没实现。另外项目目录下也没有 go.mod 文件。快照那块我看了下代码，createSnapshot 里拿了 snapshotData 但只是 log.Printf 打了一下就丢弃了，没存到磁盘，InstallSnapshot RPC 也只有类型定义没有实际实现。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：createSnapshot() 仍然在持有 rn.mu 期间调用 stateMachine.CreateSnapshot()（raft.go:496），快照期间 JSON 序列化会阻塞其他 goroutine 获取 rn.mu，导致 AppendEntries RPC 被阻塞，不满足需求11"快照创建过程中不能阻塞正常的 AppendEntries 处理"。R1 已明确指出此问题，R2 未修复。产物不满意：Leader 的 sendAppendEntries() 没有 nextIndex <= lastSnapshotIndex 时发送 InstallSnapshot 的逻辑，快照后落后 follower 的 nextIndex 被设为 0（raft.go:528），Leader 只发 PrevLogIndex=-1 的 AppendEntries，follower 日志不匹配拒绝后 Leader 无法递减 nextIndex，形成死循环，follower 永远无法同步快照。产物不满意：go.mod 指定 Go 1.21 但使用 Go 1.20 已废弃的 rand.Seed()（raft.go:847）。过程不满意：R1 指出的快照锁问题未修复，且新增了 Leader 不发 InstallSnapshot 的关键 bug |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 86-go-raft-node |

---

## 86-go-raft-node — 第 3 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 3 |
| User Prompt | 上次提的快照锁问题还没修。createSnapshot() 里 stateMachine.CreateSnapshot()（raft.go:496）还是在 rn.mu 持有期间调用的，JSON 序列化期间 AppendEntries 的 rn.mu.Lock() 会被阻塞。正确的做法是先锁 rn.mu 拷贝元数据，然后释放锁，在无锁状态下调 CreateSnapshot() 和写磁盘，最后再重新加锁截断日志。另外 Leader 发心跳的时候有个关键 bug：sendAppendEntries 里没有判断 follower 的 nextIndex 是否已经落后于快照。快照后 nextIndex 被设为 0（raft.go:528），但 Leader 还是发 AppendEntries 给那个 follower，PrevLogIndex=-1，follower 的日志对不上就拒绝了，Leader 想 nextIdx-- 但已经在 0 了，这个 follower 就永远追不上了。应该在 sendAppendEntries 里加个判断：如果 nextIdx <= lastSnapshotIndex 就改发 InstallSnapshot。还有个小事，go.mod 用的是 Go 1.21，但 raft.go:847 的 rand.Seed() 在 1.20 就废弃了，改成 rand.New(rand.NewSource(...)) 吧。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：sendAppendEntries() 中 InstallSnapshot 路径（raft.go:386-401）在持有 rn.mu 期间调用 rn.stateMachine.CreateSnapshot() 做 JSON 序列化，与 R2 修复的 createSnapshot() 持锁问题是同一类错误——持锁期间做耗时 IO 会阻塞 AppendEntries 处理。正确的做法同 createSnapshot：先拷贝快照元数据释放锁，无锁状态调 CreateSnapshot() 和发送 RPC。产物不满意：rpc.go 使用 Go 1.16 已废弃的 io/ioutil 包（ioutil.ReadAll、ioutil.WriteFile），go.mod 指定 Go 1.21 应改用 io 和 os 包 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 86-go-raft-node |

---
