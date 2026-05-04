# Solo Coder 填表数据

## 362-go-emoji-normalizer — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 写一个 Go Emoji标准化处理库。用户输入的文本里经常混入各种Emoji表情，不同平台对同一个Emoji的编码可能不同——比如有些平台用组合Unicode码点来表示不同肤色。存储和搜索之前需要把Emoji标准化。 支持的功能：提取文本中所有Emoji并返回每个Emoji的位置和字符、统计文本中Emoji的总数量、移除文本中所有Emoji、把Emoji替换为指定文字如"[表情]"、检查一段文本是否只包含Emoji（用于验证用户昵称输入）。有些Emoji由多个Unicode码点组成（肤色修饰、零宽连接符序列），要作为整体识别不能拆散。比如"👨‍👩‍👧‍👦"这个家庭Emoji由4个人物加3个零宽连接符组成，应该被识别为1个Emoji而不是7个。至少能识别50个常见的基本Emoji。非Emoji的特殊Unicode字符不要误识别——比如™️、©️这类符号不算Emoji。替换时保持原文中非Emoji部分的位置不变，移除后不留下多余空格。纯空格或纯标点的文本在"是否只包含Emoji"检查中应该返回false。 核心库代码独立为一个 package，服务端引用这个库对外提供 HTTP 接口做Emoji处理，客户端通过命令行调用服务端。服务端和客户端各自有独立的 main 包，共享的消息协议放在公共包里。各程序内部按职责分文件。先 go mod init 再开发，go build ./... 能编译通过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 | 无 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 362-go-emoji-normalizer |

---
