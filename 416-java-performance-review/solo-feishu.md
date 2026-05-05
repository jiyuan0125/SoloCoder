# Solo Coder 填表数据

## 416-java-performance-review — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | |
| 第一轮Session ID | |
| 轮次 | 1 |
| User Prompt | 绩效考核需要线上化。做一个绩效评估API，每季度HR发起评估，流程是员工自评（1到5分加评语）→上级评分（1到5分加评语）→HR确认。上级评分和员工自评分差异不能恰好为0，至少0.5分调整，防止上级偷懒直接抄员工自评分。评估完成后员工看到最终得分和部门内排名百分位。员工季度中间调岗了，绩效由新部门上级来评，原部门评语保留在记录中。绩效等级根据最终得分划分：4.5以上S级、3.5到4.5A级、2.5到3.5B级、1.5到2.5C级、1.5以下D级。连续两个季度C级进入绩效改进计划，连续三个季度C级或任一季度D级启动辞退流程。绩效得分影响年终奖系数：S级1.5、A级1.2、B级1.0、C级0.5、D级0。历史评估按部门和季度筛选查询。 绩效等级划分：4.5以上S级、3.5到4.5A级、2.5到3.5B级、1.5到2.5C级、1.5以下D级。连续两个季度C级进入绩效改进计划，连续三个季度C级或任一季度D级启动辞退流程。绩效得分影响年终奖系数：S级1.5、A级1.2、B级1.0、C级0.5、D级0。员工只能查看自己的最终得分和排名，不能看同事的具体分数。金额用 BigDecimal 处理，不能用 double 或 float。Optional 不能直接 .get()，必须有 isPresent 检查或用 orElse。 做成 Maven 多模块工程。根 pom 统一管理所有依赖的版本。公共子模块（common）定义 DTO、请求响应模型和错误码。服务端子模块（server）基于 Spring Boot，监听端口对外提供 REST 接口，业务处理和数据管理全在内存中完成。客户端子模块（client）是纯 Java 实现的命令行程序，不使用 Spring，发起 HTTP 请求调用服务端接口并将结果展示在终端。每个子模块内部按功能分包。禁止 Lombok。Java 17，运行 mvn clean package 能正常构建。 本项目属于跨系统多模块架构。  |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：3个致命Bug——maven-compiler-plugin缺少-parameters编译参数导致所有@PathVariable/@RequestParam端点500（完整评估流程走不通）、CLI客户端help命令NPE崩溃（Cannot invoke String.hashCode()因为readLine返回null）、端口配置8080而非需求要求的8086。4个严重问题——上级评分差异校验只检查恰好为0（需求要求至少0.5分调整）、连续C级计数存在重复计算（currentGrade==C时先count=1再遍历所有记录）、调岗后PerformanceReview的departmentId和managerId不更新导致新部门上级无法评、历史查询findByDepartmentIdAndYearAndQuarter的year/quarter参数完全未使用。过程不满意：端口配置和-parameters是Spring Boot 3.x最基本的构建配置，应在开发阶段通过一次curl测试即可发现 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 416-java-performance-review |

---
