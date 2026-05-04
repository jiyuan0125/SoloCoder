# Solo Coder 填表数据

## 150-c-xml-writer — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 帮我用 C 写一个 RSS 订阅内容生成模块。我们的技术博客需要自动生成 RSS Feed，让读者通过RSS 阅读器订阅新文章。这个模块负责把文章数据转成标准的 RSS XML 格式。 RSS 基本结构：XML 声明 + rss 根元素 + channel 元素 + item 元素列表。channel 包含标题、描述、链接、语言、最后更新时间等元信息。每个 item 代表一篇文章，包含标题、链接、描述、发布时间、作者、分类标签。 XML 生成规则： - XML 声明：<?xml version="1.0" encoding="UTF-8"?> - 所有元素必须正确嵌套和关闭 - 文本内容中的特殊字符要转义：< 变成 &lt;，> 变成 &gt;，& 变成 &amp;，" 变成 &quot; - 属性值用双引号包裹 - 输出要格式化（缩进 2 空格），方便人工查看 时间格式：RSS 使用 RFC 822 时间格式（如 Mon, 03 May 2026 12:00:00 GMT）。需要把Unix 时间戳或 C 标准时间类型转成这种格式。时区要处理——传入的时间可能是本地时间，需要转成 GMT。 CDATA 段：文章描述中可能包含 HTML 内容（如 <p>段落</p>），这些内容不应该被 XML 转义，而是用 CDATA 包裹：<description><![CDATA[<p>内容</p>]]></description>。但要注意，CDATA 内容中不能出现 ]]>，否则需要拆分成多个 CDATA 段。 输出方式：可以输出到文件（传入文件路径），也可以输出到内存缓冲区（传入缓冲区指针和大小）。大 Feed（几千篇文章）不能一次性在内存中构建完整的 XML 字符串——要支持流式输出，写一个 item 就刷新一次。 数量限制：RSS Feed 通常只包含最近的 N 篇文章（比如 20 篇）。模块要支持设置最大 item 数量。如果文章列表超过限制，只输出最新的 N 篇（按发布时间降序）。 有个隐藏坑：文章标题或描述中可能包含非法的 UTF-8 字节序列，直接输出会导致 XML 格式错误。需要在输出前验证并替换非法字节。 用纯 C 写，gcc 在 Linux 上编译通过。写个 main 演示：传入几篇文章信息，生成 RSS XML 文件。 代码要拆成至少 4 个源文件，按功能模块分——XML 结构构建、特殊字符转义、时间格式转换与 CDATA 处理、主程序。头文件和实现文件分开。gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 库/SDK |
| 修改范围 | 全新项目 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：XML_Escape 函数边界检查有 off-by-one bug（xml_escape.c 第56行 `pos + rlen >= output_size - 1` 应为 `>` 而非 `>=`），导致转义后内容恰好填满缓冲区时最后一个特殊字符被丢弃，demo 输出 author 字段的 `>` 未被转义为 `&gt;`，违反 PROMPT 明确要求。XML_WriteCDataContent（xml_builder.c 第82行）存在栈缓冲区溢出，temp 仅 3 字节但 strncpy 可能写入更多。XML_Escape_Length 在原始输入上计算长度，但 XML_Escape 内部先做 UTF-8 修复可能产生更长字符串，导致缓冲区不足被截断。xml_builder.c 整个模块是死代码，RSS 模块完全内联实现未使用 XML_Builder API。过程不满意：XML 转义是核心功能，边界 bug 导致 demo 输出可见错误，说明没有仔细验证转义结果。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 150-c-xml-writer |

---

## 150-c-xml-writer — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我刚跑了一下生成的 feed.xml，发现 author 字段里 email 地址的右尖括号没有被转义，输出是 `张三 &lt;zhangsan@example.com>` 而不是 `张三 &lt;zhangsan@example.com&gt;`。另外 xml_builder.c 这个模块好像完全没有被用到，RSS 那边全是自己内联写的 XML 输出逻辑。 |
| 任务类型 | Bug 修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：XML_WriteCDataContent（xml_builder.c 第81行）仍然存在栈缓冲区溢出漏洞，temp 仅 3 字节但 strncpy 的第三个参数 `p - start + 2` 可以远大于 3，会导致栈破坏。XML_WriteEscaped（xml_builder.c 第56行）使用 XML_Escape_Length 计算分配大小，但 XML_Escape 内部先做 UTF-8 修复可能产生更长字符串（非法字节被替换为 3 字节 U+FFFD），导致堆缓冲区不足被截断。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 150-c-xml-writer |

---
