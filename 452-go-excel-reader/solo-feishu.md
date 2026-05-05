# Solo Coder 填表数据

## 452-go-excel-reader — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 第一轮 |
| User Prompt | 写一个Go库读取xlsx格式文件，这是从Excel报表中提取数据的基础能力。xlsx本质是ZIP压缩包里包含XML文件。能读取指定sheet名称的数据返回二维字符串数组。支持按行号范围读取不用全部加载。日期单元格正确解析为可读日期字符串不要输出一串数字。纯数字单元格不用科学计数法（1000000不要显示1E+06）。单元格有合并时，非左上角的返回空字符串只在左上角返回实际值。有些Excel文件sheet名称里包含空格或特殊字符，引用时要能正确处理。公式单元格（以=开头的内容）读取的是公式的计算结果缓存值而不是公式本身，这个行为要在文档里说明。如果指定的sheet名称不存在，返回包含所有可用sheet名称的明确错误信息方便调用方排查。xlsx文件如果被加密保护需要密码才能打开，检测到加密时返回明确错误提示。空行（所有单元格都为空）默认跳过不包含在返回结果中，可通过参数 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 库/SDK |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：openpyxl等主流xlsx生成工具默认使用inlineStr格式存储字符串单元格，当前Cell结构体只定义了V字段（映射&lt;v&gt;标签），未处理&lt;is&gt;&lt;t&gt;嵌套结构，导致所有inlineStr类型的字符串单元格返回空字符串，用openpyxl创建的xlsx文件读取后所有文本内容全部丢失。过程不满意：xlsx读取库的核心能力是正确解析各种格式的单元格，但只处理了shared strings（type "s"）和数字（type "n"），没有覆盖inlineStr（type "inlineStr"）这种常见的字符串存储方式，说明对xlsx格式规范理解不全面 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 452-go-excel-reader |

---

## 452-go-excel-reader — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 第二轮 |
| User Prompt | 我拿一个openpyxl生成的xlsx文件试了下，上传后读sheet发现所有文本内容都是空的，数字倒是能正常显示。我用Python重新创建了一个用sharedStrings格式的xlsx文件测试，文字就能读出来了，看起来是inlineStr那种字符串存储方式没处理好，你看看怎么回事。 |
| 任务类型 | Bug修复 |
| 业务领域 | 库/SDK |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 完成了任务 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 | |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 452-go-excel-reader |

---
