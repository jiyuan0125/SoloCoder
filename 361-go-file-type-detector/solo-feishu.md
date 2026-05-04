# Solo Coder 填表数据

## 361-go-file-type-detector — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 写一个 Go 文件类型识别库。用户上传文件时不能只看文件扩展名来判断类型（扩展名容易被伪造），要通过读取文件头部的字节来判断真实的文件类型。  支持的文件类型：JPEG（开头FFD8FF）、PNG（开头89504E47）、GIF（开头47494638）、PDF（开头25504446）、ZIP（开头504B0304）、RAR（开头52617221）。提供两个入口函数：通过文件路径检测类型、通过字节切片（文件内容前512字节）检测类型。返回标准MIME类型字符串如"image/jpeg"。文件内容可能不足512字节甚至不足需要的最小字节数，这种情况要正确处理而不是panic。不匹配任何已知类型返回"unknown"而不是报错。有些文件类型共享相同的文件头标记——比如Office文档的DOCX和XLSX底层都是ZIP格式，此时返回"application/zip"是正确行为，调用方可以结合扩展名进一步判断。空文件返回"unknown"。通过文件路径检测时如果文件不存在或没有读取权限要返回明确的错误。  核心库代码独立为一个 package，服务端引用这个库对外提供 HTTP 接口做文件类型检测，客户端通过命令行传入文件路径或字节调用服务端。服务端和客户端各自有独立的 main 包，共享的消息协议放在公共包里。各程序内部按职责分文件。先 go mod init 再开发，go build ./... 能编译通过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：DetectFromPath 对空文件返回 ("", io.EOF) 而非 ("unknown", nil)，PROMPT 明确要求"空文件返回unknown"。Handler 层检查 len(req.Data) > 0 导致客户端发送空文件数据时被拒绝为"No data or path provided"。过程不满意：ioutil.ReadAll 自 Go 1.16 起已废弃应改用 io.ReadAll，客户端用 err.Error() != "EOF" 字符串比较而非 errors.Is(err, io.EOF)，写完后没有实际测试空文件场景 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 361-go-file-type-detector |

---
