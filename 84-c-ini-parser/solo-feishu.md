# Solo Coder 填表数据

## 84-c-ini-parser — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 用 C 写一个 INI 配置文件解析器。支持读取、查询、修改和写回。 解析功能： 1. 支持 section：[section_name]，section 名不区分大小写 2. 支持 key=value 对。value 两端的空白字符自动去除 3. 注释：; 和 # 开头的行为注释。行内注释不支持（; 和 # 在值中出现算作普通字符） 4. 支持 include = path/to/other.ini 指令递归包含其他配置文件，最大递归深度 10 层，超过报错并输出当前深度 5. 解析错误时报告行号（从 1 开始）和错误描述 环境变量展开： 6. 值中的 ${ENV_VAR} 展开为 getenv() 的结果。如果环境变量不存在，保持 ${ENV_VAR} 原样不替换（不报错） 7. 展开后再去除 value 两端的空白 类型推断： 8. ini_get_string() 返回原始字符串值 9. ini_get_int() 如果值全是数字（可选前导 +/- 和空格）则解析为 int 返回，否则返回默认值 10. ini_get_bool() 如果值是 true/TRUE/1/yes/YES 返回 1，false/FALSE/0/no/NO 返回 0，否则返回默认值 写回功能： 11. ini_set(section, key, value) 设置值。如果 section 或 key 不存在则创建 12. ini_save(parser, path) 将当前配置写回文件。只修改值的部分，保留原有的注释、空行和格式 API： 13. ini_parse(path) 从文件解析，返回 ini_parser_t* 14. ini_free(parser) 释放所有内存 15. ini_get(parser, section, key, default_value) 获取值（字符串） 代码分 ini_parser.c、ini_parser.h 两个文件，gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：ini_save 丢失 include 指令——加载含 include = path 的 INI 文件后调用 ini_save 写回，原始的 include 行被永久丢弃，被引用文件的内容被内联展开。PROMPT 需求 12 明确要求"只修改值的部分，保留原有的注释、空行和格式"，include 指令属于原始格式的一部分，不应在 save 时丢失。过程不满意：测试用例没有验证 save 后 include 指令是否保留，写完后没有人工检查保存结果与原始文件的差异 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 84-c-ini-parser |

---
