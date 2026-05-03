# Solo Coder 填表数据

## 81-c-smtp-client — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 用 C 写一个 SMTP 客户端库，用于发送邮件。基于 socket 连接 SMTP 服务器。 SMTP 协议交互： 1. 连接服务器后发送 EHLO，根据服务器响应判断是否支持 STARTTLS 和 AUTH 2. 支持 STARTTLS 升级到 TLS（用 OpenSSL）。如果服务器不支持 STARTTLS 就用明文继续 3. 支持 AUTH LOGIN 认证：发送用户名和密码的 base64 编码 4. 邮件格式：支持 From、To、Cc、Bcc、Subject、Date、Content-Type header。Bcc 收件人不出现在邮件 header 中 5. 邮件正文和附件用 MIME 编码，正文用 7bit 或 base64，附件用 base64 编码。Content-Transfer-Encoding: base64 BASE64 编码： 6. 自行实现 base64 编码函数（不用外部库），编码后每行 76 个字符（RFC 标准），最后一行可能不足 76 个 7. 自行实现 base64 解码函数（SMTP AUTH LOGIN 时需要解码响应） 连接管理： 8. 连接池：同一个域名的连接复用，空闲超过 60 秒自动关闭 9. 发送失败如果服务器返回 5xx 状态码，重试最多 2 次（共发 3 次），间隔 1 秒。4xx 状态码不重试直接报错 10. 所有 socket 操作设置 10 秒超时 API： 11. smtp_send邮件(to_list, subject, body, attachment_paths) 发送邮件，返回 0 成功 / 错误码 12. smtp_quit() 发送 QUIT 命令并关闭连接 代码分 smtp_client.c、base64.c、smtp_client.h 几个文件，gcc 编译能过（不需要 -lssl，TLS 部分可以模拟跳过）。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：smtp_send_mail_internal 中 smtp_data_end 失败后用 strstr(g_last_error, "5") 判断是否 5xx，这个方式太粗暴——错误信息 "DATA end failed with code 450: ..." 中的 450 包含字符 "5"，会被误判为 5xx 触发重试，但 450 实际是 4xx 不应该重试（PROMPT 明确要求 4xx 不重试）。同理 451、452 也会被误判。另外 build_mime_message 返回 NULL（内存分配失败）时也触发了重试，但 OOM 不是瞬态错误，重试没有意义。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 81-c-smtp-client |

---

## 81-c-smtp-client — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我看了下 smtp_data_end 失败后的错误码判断逻辑，用的 strstr(g_last_error, "5") 来区分 5xx 和 4xx。但错误信息里包含的是完整的状态码数字，比如 450 里面就有字符 "5"，会被当成 5xx 去重试了，450 明明是 4xx。PROMPT 说 4xx 不重试，这个判断方式不太对。还有 build_mime_message 返回 NULL 的时候也走了重试逻辑，但那是 malloc 失败，重试也没什么用吧 |
| 任务类型 | Bug修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 |  |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 81-c-smtp-client |

---
