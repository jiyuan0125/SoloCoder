# Solo Coder 填表数据

## 314-go-luggage-tracking — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 航空公司的行李追踪系统。旅客托运行李后，行李经过值机、安检、分拣、装机、到达、传送带等环节，每个环节工作人员扫码记录。旅客通过行李牌号查询行李走到哪一步。 各环节上报经常乱序到达，展示时必须按业务流程顺序排列不能按请求到达时间排。行李在某个环节停留超30分钟无新节点上报，自动标记"异常"状态。工作人员能查某航班当天所有行李进度。行李牌号10位数字加字母，提交时做格式校验。同一牌号同一时间只能有一个在途记录。航班取消时所有未完成扫描的行李标记"航班取消"。工作人员能补录遗漏的扫码节点。 操作记录要可追溯，管理员能查看谁在什么时间做了什么操作。敏感操作（比如删除数据、修改金额）要单独记录审计日志，不能被普通用户删除或修改。 所有行李追踪和操作记录在服务重启后不能丢失。 拆成服务端和客户端两个独立程序。服务端监听HTTP端口，接收各环节扫码上报，按流程顺序整理行李轨迹，数据存在内存中。客户端是命令行工具，工作人员用它扫码上报和查询航班行李，旅客用它通过行李牌号查进度。两个程序各自有独立的main包，共享的消息格式和错误码放在公共包里。各程序内部按职责分文件。先 go mod init 再开发，go build ./... 能编译通过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：store.go 中 CreateLuggage、UpdateLuggage、AddOperationLog、AddAuditLog 四个方法都在持有 s.mu.Lock()（写锁）的情况下调用了 saveLuggages()/saveOperations()/saveAudits()，这些 save 方法内部又尝试获取 s.mu.RLock()（读锁）。Go 的 sync.RWMutex 不支持同一 goroutine 在持有写锁时再获取读锁，会永久死锁。导致所有 POST 接口（扫码上报、补录、取消航班）第一次调用就卡死，之后整个服务所有请求全部挂起，服务完全不可用。过程不满意：RWMutex 的 Lock/RLock 嵌套调用是 Go 并发编程的基础知识，写完后没有做最基本的启动测试，否则一调用扫码接口就会发现服务挂了 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 314-go-luggage-tracking |

---

## 314-go-luggage-tracking — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 2 |
| User Prompt | 服务启动没问题，health 接口也正常返回。但是调一下 POST /api/scan 上报行李扫码，curl 就卡住了一直没响应，然后试别的接口也全卡住了，服务直接不可用。你看看 store.go 里写操作加锁的逻辑，我觉得是锁嵌套的问题 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 单文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：service.go 的 Scan() 方法中，对 luggage.Completed 或 luggage.Status == common.StatusFlightCancel 的行李直接返回 ErrLuggageAlreadyExists，导致已完成或航班取消的行李无法重新开始新的追踪流程。PROMPT 要求"同一牌号同一时间只能有一个在途记录"，已完成/取消的行李不在途，应该允许重新创建。store.go 的 CreateLuggage 已正确处理此逻辑（允许覆盖已完成记录），但 Scan() 在调用 CreateLuggage 之前就提前返回了错误。过程不满意：R1 死锁问题修复了锁嵌套，但修改 Scan 逻辑时没有重新审视业务条件，写完后没有完整测试行李全生命周期（创建→完成→重新创建） |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 314-go-luggage-tracking |

---
