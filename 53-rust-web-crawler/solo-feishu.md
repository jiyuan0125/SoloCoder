# Solo Coder 填表数据

## 53-rust-web-crawler — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:5b30cacce0b77444aded732ca1b723c6_69f53ae69dbef3d8838bc0f2.69f53b1f9dbef3d8838bc129.69f53b1e160b3d7f9ace029d:Trae CN.T(2026/5/2 07:45:35) |
| 第一轮Session ID | .335769888099319:5b30cacce0b77444aded732ca1b723c6_69f53ae69dbef3d8838bc0f2.69f53b1f9dbef3d8838bc129.69f53b1e160b3d7f9ace029d:Trae CN.T(2026/5/2 07:45:35) |
| 轮次 | 1 |
| User Prompt | 用 Rust 写一个并发爬虫。从种子 URL 开始抓取网页，解析出页面里的链接继续爬。支持深度限制，比如指定最大深度 3 就只爬前 3 层链接。同一个域名最多同时 2 个请求在跑，超过的排队等着。URL 去重用 HashSet，已经爬过的不再爬。每个域名先请求 /robots.txt 看有没有 Disallow 规则，有的话遵守。收到 SIGINT 或者调用 shutdown 的时候，已经发出去的请求等它返回，但不再发新的请求。爬到的结果（URL、标题、状态码）输出到一个文件里。代码分爬虫调度、URL 队列、页面解析、robots.txt 解析几个模块，cargo build 能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 | |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 53-rust-web-crawler |

---
