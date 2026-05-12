# 医院感染监测系统

一套完整的医院感染监测系统，用于管理感染病例监测、防控措施和报告。

## 技术栈

- **后端**: Go (Gin + GORM + SQLite)
- **前端**: React 18 + React Router + Axios
- **数据库**: SQLite

## 项目结构

```
├── backend/           # Go 后端
│   ├── main.go        # 服务入口
│   ├── models/        # 数据模型
│   ├── database/      # 数据库初始化
│   ├── handlers/      # API 处理器
│   └── cmd/           # CLI 客户端
└── frontend/          # React 前端
    ├── src/
    │   ├── pages/     # 页面组件
    │   ├── components/ # 通用组件
    │   └── api.js     # API 封装
    └── package.json
```

## 功能特性

### 1. 感染病例管理
- 录入和查看感染病例
- 系统自动判断感染类型（社区感染/院内感染）
- 入院48小时为分界线
- 支持按科室和类型筛选

### 2. 感染率统计
- 按科室统计月度感染率
- 保留两位小数百分比
- 预警阈值默认2%（可配置）
- 超标自动产生预警
- 连续两个月超标产生"需要干预"高优先级待办

### 3. 防控措施管理
- 记录防控措施执行情况
- 根据感染部位自动推荐防控方案
- 支持措施类型：隔离、手卫生、环境消毒、抗生素调整、器械消毒整改

### 4. 目标性监测
- 重点科室：ICU、新生儿科、烧伤科、血液科
- 呼吸机相关肺炎（VAP）
- 导管相关血流感染（CLABSI）
- 导尿管相关尿路感染（CAUTI）
- 器械使用率和相关感染发病率计算

### 5. 报告管理
- 月报和年报生成
- 审批流程：提交-一审-二审-终审-通过
- 小额跳过二审直接终审
- 驳回回到提交状态
- 终审后不可修改

### 6. 统计面板
- 近12个月全院感染率趋势（SVG折线图）
- 各科室感染率排名（SVG柱状图）
- 感染部位分布
- 病原体分布
- 预警通知展示

## 后端运行

```bash
cd backend

# 安装依赖
go mod tidy

# 运行服务 (默认端口8080)
go run .

# 指定端口
PORT=8081 go run .
# 或
go run . --port 8081
```

服务启动后，访问 http://localhost:8080

## 前端运行

```bash
cd frontend

# 安装依赖
npm install

# 开发模式
npm start

# 构建
npm run build
```

前端默认访问 http://localhost:3000

## CLI 客户端

```bash
cd backend/cmd

# 安装依赖
go mod tidy

# 构建
go build -o hicli .

# 报告感染病例
./hicli report-infection \
  --patient-id P20260001 \
  --name "张三" \
  --dept 1 \
  --admission 2026-05-01 \
  --infection 2026-05-05 \
  --site respiratory \
  --pathogen "肺炎克雷伯菌"

# 添加防控措施
./hicli add-measure \
  --case-id 1 \
  --type isolation \
  --dept 1 \
  --executor "李医生"

# 生成报告
./hicli generate-report --type monthly --month 2026-05

# 查看感染率
./hicli show-rates
./hicli show-rates --month 2026-05
```

## API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/departments | 获取科室列表 |
| POST | /api/infections | 创建感染病例 |
| GET | /api/infections | 获取感染病例列表 |
| GET | /api/infections/:id | 获取单个感染病例 |
| PUT | /api/infections/:id | 更新感染病例 |
| DELETE | /api/infections/:id | 删除感染病例（含关联记录） |
| POST | /api/measures | 添加防控措施 |
| GET | /api/measures | 获取防控措施列表 |
| GET | /api/measures/recommended | 获取推荐防控措施 |
| DELETE | /api/measures/:id | 删除防控措施 |
| POST | /api/monitoring | 创建目标性监测记录 |
| GET | /api/monitoring | 获取目标性监测列表 |
| GET | /api/monitoring/departments | 获取重点科室 |
| DELETE | /api/monitoring/:id | 删除目标性监测 |
| GET | /api/statistics/rates | 获取月度感染率统计 |
| GET | /api/statistics/trend | 获取12个月趋势数据 |
| GET | /api/statistics/alerts | 获取预警列表 |
| POST | /api/reports/generate | 生成报告 |
| GET | /api/reports | 获取报告列表 |
| POST | /api/reports/:id/approve/:action | 审批报告 |
| DELETE | /api/reports/:id | 删除报告 |

## 业务规则

1. **感染病例校验**
   - 患者住院号为空 → 返回 404
   - 感染日期早于入院日期 → 返回 400
   - 感染部位/病原体为空 → 返回 400
   - 手动指定感染类型 → 返回 400 及自动判断结果

2. **感染率统计**
   - 出院人数为零 → 感染率按零算
   - 感染率 = 院内感染例数 / 出院人数 × 100%
   - 保留两位小数

3. **防控措施**
   - 执行科室不存在 → 返回 404

4. **目标性监测**
   - 器械使用天数为负 → 返回 400

5. **删除清理**
   - 删除时自动清理关联记录
   - 无关联记录不产生效果

6. **价格调整**
   - 未执行的按新价格
   - 已执行的不受影响
   - 记录变更历史

## 数据模型

### 感染病例 (InfectionCase)
- 患者住院号、姓名、性别、年龄
- 入院科室、入院日期、感染日期
- 感染部位、感染类型（自动判断）
- 病原体、药敏试验结果

### 防控措施 (PreventionMeasure)
- 关联感染病例
- 措施类型、执行科室、执行人
- 执行日期、状态

### 目标性监测 (TargetMonitoring)
- 科室、月份
- 住院天数、各器械使用天数
- 各类型感染例数

### 预警 (Alert)
- 科室、月份
- 预警类型（超标/需要干预）
- 感染率、阈值
- 通知状态

### 报告 (Report)
- 报告类型（月报/年报）
- 月份/年份
- 全院感染率
- 各科室感染率、部位分布、病原体分布
- 审批状态
