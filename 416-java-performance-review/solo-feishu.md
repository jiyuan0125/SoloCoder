# Solo Coder 填表数据

## 416-java-performance-review — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 绩效考核需要线上化。做一个绩效评估API，每季度HR发起评估，流程是员工自评（1到5分加评语）→上级评分（1到5分加评语）→HR确认。上级评分和员工自评分差异不能恰好为0，至少0.5分调整，防止上级偷懒直接抄员工自评分。评估完成后员工看到最终得分和部门内排名百分位。员工季度中间调岗了，绩效由新部门上级来评，原部门评语保留在记录中。绩效等级根据最终得分划分：4.5以上S级、3.5到4.5A级、2.5到3.5B级、1.5到2.5C级、1.5以下D级。连续两个季度C级进入绩效改进计划，连续三个季度C级或任一季度D级启动辞退流程。绩效得分影响年终奖系数：S级1.5、A级1.2、B级1.0、C级0.5、D级0。历史评估按部门和季度筛选查询。 BigDecimal处理金额，Optional不能直接.get()。 Maven多模块（common/server/client），Spring Boot server 内存存储，纯Java CLI client HTTP调用。Java 17，禁止Lombok。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 纯后端API服务 |
| 修改范围 | 跨系统多模块 |
| 任务是否完成 | 未完成任务 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：maven-compiler-plugin缺少-parameters配置，Spring Boot 3.x下@PathVariable未显式指定name导致路径变量接口返回500错误（需手动修复编译配置后才可正常工作）。上级评分差值校验仅检查≠0未检查≥0.5，PROMPT要求至少0.5分调整。调岗后无业务联动更新departmentId/managerId。过程不满意：编译配置遗漏导致原始交付物核心接口不可用，说明构建后未实际测试启动 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 416-java-performance-review |

---
