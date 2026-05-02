# Solo Coder 填表数据

## 79-c-json-parser — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:68780ce131d51f9d6f156729aac8817c_69f62cc557c8d33337f239a9.69f62cfa57c8d33337f239e3.69f62cf9b899bda2270951cd:Trae CN.T(2026/5/3 00:57:30) |
| 第一轮Session ID | .335769888099319:68780ce131d51f9d6f156729aac8817c_69f62cc557c8d33337f239a9.69f62cfa57c8d33337f239e3.69f62cf9b899bda2270951cd:Trae CN.T(2026/5/3 00:57:30) |
| 轮次 | 1 |
| User Prompt | 用 C 写一个 JSON 解析器。支持流式（SAX）解析和 DOM 解析两种模式。 SAX 流式解析： 1. json_parse_sax(input, callbacks) 从 FILE* 流式解析，通过回调通知：object_start, object_end, array_start, array_end, key, string, number, bool, null 2. 回调携带位置信息（行号:列号，都从 1 开始计数） 3. 支持从文件流读取大文件，不要一次性全部读入内存（每次读 4KB 缓冲区） 数值处理： 4. 整数（没有小数点和 e/E 的）解析为 long long，返回 json_int 类型 5. 浮点数（有小数点或 e/E 的）解析为 double，返回 json_float 类型 6. 不要把所有数字都当 double 处理，否则 9223372036854775807 这种大整数会丢精度 字符串处理： 7. 正确处理转义序列：\n \t \r \\ \" \/ 和 \uXXXX（4 位十六进制 Unicode，转为 UTF-8） 8. 遇到非法 UTF-8 字节序列报错并指出位置 错误处理： 9. 语法错误返回错误码并设置错误信息（包含位置：第几行第几列） 10. 输入末尾有多余内容（JSON 值结束后还有非空白字符）也视为错误 DOM 解析： 11. json_parse_dom(input) 将整个 JSON 解析成一棵树结构（json_value），支持 json_get(obj, key) 和 json_get_at(arr, index) 取值 12. json_free(value) 递归释放 DOM 树的所有内存 代码分 json_parser.c、json_dom.c、json.h 几个文件，gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 | |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 79-c-json-parser |

---
