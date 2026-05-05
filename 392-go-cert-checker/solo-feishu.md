# Solo Coder 填表数据

## 392-go-cert-checker — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 写一个 Go 命令行工具检查TLS证书是否即将过期。管理了很多域名需要定期检查证书有效期提前告警。 用法：./cert-checker example.com:443 api.example.com:443 --warn-days 30。连接指定域名的443端口获取TLS证书信息，输出证书域名、颁发者、生效日期、过期日期、剩余天数。剩余天数少于告警阈值时醒目标记。支持从文件批量读取域名列表。证书可能有多个SAN要检查请求域名是否在列表中。域名无法连接或TLS握手失败时记录错误但继续检查其他域名。证书链中的中间证书不需要单独检查只关注叶子证书。支持自定义端口。连接超时设置为5秒避免某个域名响应慢阻塞整个检查。证书已经过期时标记为"已过期"而不是显示负数天数。加 --json 参数输出JSON格式方便脚本处理。如果证书的SAN列表为空则只检查Common Name。域名不带端口号时默认使用443端口。证书的Subject Alternative Names中包含通配符域名（如*.example.com）时标注为通配符证书。 拆成两个独立程序：一个是后台常驻的服务进程，负责域名列表管理和证书检查调度，维护历史检查结果供对比；另一个是命令行客户端，用户通过它提交域名和查看证书状态。两个程序各自独立 main 包，通过 TCP 通信，共享协议定义放公共包。各程序内部按职责分文件。先 go mod init 再开发，go build ./... 能编译通过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：命令行参数解析有 bug，PROMPT 中明确用法是 ./cert-checker example.com:443 --warn-days 30（flag 放在域名后面），但 Go 的 flag 包在遇到第一个非 flag 参数后就停止解析了，导致 --warn-days 30 被当成域名去连接，实际测试输出 "域名: --warn-days 错误: dial tcp: lookup --warn-days: no such host" 和 "域名: 30 错误: dial tcp: lookup 30: no such host"。这是一个基本的功能性 bug，文档规定的用法无法正常工作。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 392-go-cert-checker |

---
