# Solo Coder 填表数据

## 438-go-mailing-list — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 第一轮 |
| User Prompt | 运营做邮件推送触达用户。做一个邮件订阅管理API，用户可以订阅或退订不同邮件列表。退订立即生效，退订后不再收到该列表的邮件。群发支持定时发送，定时任务可以提前创建但未到时间不会执行。每封邮件有发送记录（收件人、时间、状态成功/失败/退回）。某个邮件列表退回率超过10%时自动暂停发送并生成告警通知运营团队。收件人邮箱格式明显不正确的不发送记为无效地址，同一邮箱连续3次无效自动退订。每个邮件列表最多10万个收件人，超过限制时拆分为多批次发送每批间隔30秒。支持A/B测试——同一封邮件可以有两个标题版本随机发送给各50%的收件人，统计打开率后自动选用更优版本补发另一半。 每个邮件列表最多10万个收件人，超过限制时拆分为多批次发送每批间隔30秒。支持A/B测试——同一封邮件可以有两个标题版本随机发送给各50%的收件人，统计打开率后自动选用更优版本补发另一半。邮件模板支持变量替换比如用户姓名和注册时间邮件打开率和点击率追踪统计。收件人列表支持从文件导入，格式不正确的自动过滤。同一封邮件不能在24小时内重复发送给同一收件人。邮件发送队列管理——高峰期自动排队控制发送速率。。 分两个独立程序。服务端启动后监听端口对外提供HTTP服务，负责所有业务逻辑，数据存在内存里。客户端是命令行工具，用户通过命令行发请求给服务端，服务端处理完返回结果，客户端展示给用户。两个程序各自独立main包，通信格式定义放在公共包。各程序内部按职责分文件。先go mod init再开发，go build ./...能编译通过。 本项目属于跨系统多模块架构。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：1) SendCampaign导致服务端panic崩溃——SendQueue.store和SendQueue.alertMgr从未初始化为nil，processItem调用q.store.UpdateTask时空指针解引用。2) /api/lists/{id}/subscribers路由被GetMailingList优先匹配，导致查询收件人列表永远返回404。3) 创建邮件列表时空名称返回"invalid email format"错误信息，空白名称(纯空格)可以通过校验。4) handlers.go中所有MethodNotAllowed错误都复用ErrInvalidEmailFormat。5) ListABTests方法错误定义在Store而非Service上。过程不满意：发邮件这个核心功能会导致服务崩溃，说明没有实际端到端测试过。 |
| github地址 | |
| 分支/文件夹 | 438-go-mailing-list |

---

## 438-go-mailing-list — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 第二轮 |
| User Prompt | 我刚跑了一下项目，发现发邮件的时候服务直接崩了，日志里是 nil pointer dereference panic，在 queue.go 的 processItem 里调 store.UpdateTask 的时候炸的。另外 GET /api/lists/{id}/subscribers 这个接口返回 404，但创建列表和订阅都能正常工作，订阅完查收件人列表查不到。还有创建列表的时候 name 传空字符串返回的错误信息是 "invalid email format"，有点奇怪。 |
| 任务类型 | Bug修复 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：handlers.go 的 GetSendRecords 方法用 parts[len(parts)-1] 从 URL 路径 /api/tasks/{taskID}/records 提取 taskID，实际取到的是 "records" 字符串而非 taskID，导致发送记录查询接口永远返回 null。过程不满意：R1 的 4 个 bug 均已修复（SendQueue 初始化、路由匹配顺序、空名称错误信息、MethodNotAllowed 复用），但修复过程中没有发现这个预先存在的 records 接口 bug，说明没有完整测试所有端点。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 438-go-mailing-list |

---
