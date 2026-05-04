# Solo Coder 填表数据

## 143-c-template-engine — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 帮我用 C 写一个轻量级的模板渲染引擎，用于生成营销邮件的 HTML 内容。运营配置邮件模板，模板里有占位符，渲染时把占位符替换成实际数据。 变量替换：模板中用 {{变量名}} 表示占位符，渲染时替换成对应的值。变量名支持点号分隔的嵌套访问（如 {{user.name}} 表示 user 对象的 name 字段）。 条件渲染：支持 {{#if 条件}} ... {{/if}} 的条件块。条件为真时输出块内内容，为假时跳过。也支持 {{#unless 条件}}（取反）和 {{#else}}（else 分支）。 循环渲染：支持 {{#each 数组}} ... {{/each}} 的循环块。在循环体内可以用 {{this}} 引用当前元素，用 {{@index}} 引用当前索引（从 0 开始）。嵌套循环也要支持。 HTML 转义：变量值在替换到模板时自动进行 HTML 转义（< 变成 &lt;，> 变成 &gt; 等），防止 XSS 攻击。如果明确知道值是安全的 HTML，可以用 {{{三重花括号}}} 来跳过转义。 模板包含：支持在一个模板中引用另一个模板——{{> partial_name}} 会被替换成名为 partial_name的子模板的渲染结果。子模板可以访问外层模板的所有变量。 自定义分隔符：默认用 {{ }} 作为分隔符，但可以配置成其他符号（比如 [[ ]]），用于在同一个页面中混用模板引擎和其他模板语法（如 Vue 的 {{ }}）。 错误处理：变量未定义时输出空字符串而不是崩溃。循环遍历 null 时当作空数组处理。模板语法错误时（如未关闭的 if 块）给出错误信息。 有个容易忽略的点：条件判断中 {{#if user.active}} 中的 user.active 应该是布尔判断——如果值是 0、空字符串、null，应该视为 false；非零数字、非空字符串视为 true。不能简单判断指针是否非空。 用纯 C 写，gcc 在 Linux 上编译通过。写个 main 演示：定义一个模板和一组数据，渲染输出。 代码要拆成至少 4 个源文件，按功能模块分——模板解析、变量替换与条件渲染、循环渲染与 HTML 转义、主程序。头文件和实现文件分开。gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 库/SDK |
| 修改范围 | 全新项目 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：{{#else}} 分支完全未实现——template_parser.c 第 470 行将 TOKEN_ELSE 立即 free 丢弃，render.c 中无 TOKEN_ELSE 处理分支，导致 if/unless 的 else 分支完全无效。过程不满意：编译有 3 个 warning（unused function buffer_free、unused variable actual_open_len/actual_close_len），项目目录中有编译产物 template_demo 二进制文件未加入 .gitignore |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 143-c-template-engine |

---

## 143-c-template-engine — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我测了下，{{#else}} 分支不生效。写了个模板 `{{#if show}}显示A{{#else}}显示B{{/if}}`，show 为 false 的时候输出的是空的，应该输出"显示B"才对。你看下是不是 else 的处理逻辑漏了。 |
| 任务类型 | Bug 修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：R1 指出的 3 个编译器 warning 完全未修复（buffer_free 未使用、actual_open_len/actual_close_len 赋值后未使用），与 R1 完全一致。过程不满意：R1 指出的编译产物入库问题未修复——template_demo 二进制仍在 git 跟踪中，且新增了 test_else 二进制也被提交入库，.gitignore 仍缺少对无扩展名可执行文件的忽略规则 |
| github地址 |  |
| 分支/文件夹 | 143-c-template-engine |

---
