# Solo Coder 填表数据

## 389-go-db-dumper — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 写一个 Go 命令行工具把CSV文件数据转成SQL INSERT语句。不做真正的数据库连接，只是读CSV作为数据源生成SQL。 用法：./db-dumper --table users --input users.csv --output users.sql。CSV第一行是列名后面的行是数据。生成格式为INSERT INTO table_name (col1, col2) VALUES (val1, val2);。字符串值中的单引号转义为两个单引号。空字符串或"NULL"转为SQL的NULL。数字类型不加引号。每100条语句一个事务（BEGIN/COMMIT）。CSV中含逗号（双引号包裹的字段）和换行符要正确解析。只有表头没有数据时生成含注释说明的SQL文件。表名通过参数指定，如果未指定则使用CSV文件名（去掉扩展名）作为表名。列名如果包含SQL关键字或特殊字符自动用双引号包裹。CSV的BOM头（UTF-8文件开头的EF BB BF）要自动跳过。生成的SQL文件头部添加注释说明生成时间和源文件信息。 拆成两个独立程序：一个是后台常驻的服务进程，负责CSV解析和SQL语句生成，维护转换配置；另一个是命令行客户端，用户通过它指定输入输出文件和表名参数。两个程序各自独立 main 包，通过 TCP 通信，共享协议定义放公共包。各程序内部按职责分文件。先 go mod init 再开发，go build ./... 能编译通过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 完成了任务 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 |  |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 389-go-db-dumper |

---
