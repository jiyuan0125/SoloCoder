# 中医馆管理系统 (CMMS)

一个完整的中医馆信息化管理系统，支持患者挂号、医生开处方、药房抓药和煎药管理等核心流程。

## 技术栈

- **后端**: Go + Gin + Gorilla WebSocket
- **前端**: React + Ant Design
- **数据存储**: 内存存储 (sync.Map)

## 项目结构

```
.
├── backend/                 # Go后端
│   ├── cmd/server/         # 服务入口
│   ├── internal/
│   │   ├── models/         # 数据模型
│   │   ├── storage/        # 内存存储
│   │   ├── handlers/       # HTTP处理器
│   │   └── ws/             # WebSocket管理
│   └── go.mod
├── frontend/               # React前端
│   ├── public/
│   ├── src/
│   │   ├── components/     # React组件
│   │   ├── api.js          # API封装
│   │   ├── App.js
│   │   └── index.js
│   └── package.json
└── README.md
```

## 功能特性

### 1. 患者挂号
- 填写姓名、手机号、主诉症状
- 自动生成格式为"日期+三位序号"的挂号序号（如 20260509001）
- 每天从1开始，跨天重置
- 高并发下保证唯一且连续

### 2. 医生处方
- 选择患者、填写诊断（病名+证型）
- 添加处方项（药材、剂量、单位、煎煮方式）
- 最多20味药，总剂量不超过500克
- 剂量自动调整为5的整数倍
- 状态流转：草稿 → 确认 → 已发药
- 确认后如需修改需作废重开

### 3. 药房管理
- 待发药处方列表
- 自煎/代煎两种模式
- 代煎支持：煎药机编号、锅数、预计完成时间
- 库存管理：扣减库存、缺货提示
- 效期管理：30天内临期(黄色)、过期(红色)

### 4. 实时叫号
- WebSocket实时推送
- 医生叫号后挂号大厅同步显示
- 候诊人数超过10人显示等待提示

### 5. 其他特性
- 价格变更：待处理处方自动更新，已完成的不变
- 服务端口：支持命令行参数(`--port`)或环境变量(`PORT`)

## 快速开始

### 启动后端

```bash
cd backend

# 下载依赖
go mod download

# 启动服务（默认端口8080）
go run cmd/server/main.go

# 或指定端口
go run cmd/server/main.go --port 9090
# 或
PORT=9090 go run cmd/server/main.go
```

### 启动前端

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm start
```

前端默认运行在 http://localhost:3000

## 页面说明

### 挂号大厅 (`/`)
- 显示当天所有挂号列表
- 可新增挂号
- 实时显示当前叫号
- 候诊人数>10时显示红色提示条

### 诊室工作台 (`/doctor`)
- 候诊队列列表
- 叫号功能
- 开处方、编辑处方、确认处方
- 处方查看、作废

### 药房管理台 (`/pharmacy`)
- 待发药处方
- 进行中的煎药任务
- 库存管理（临期/过期标识）
- 药材信息调整（库存、价格）

## API 接口

### 挂号相关
- `POST /api/register` - 新增挂号
- `GET /api/registrations` - 获取当天挂号列表
- `POST /api/call` - 叫号
- `POST /api/registrations/:id/finish` - 完成就诊

### 处方相关
- `POST /api/prescriptions` - 创建处方
- `PUT /api/prescriptions/:id` - 更新处方（仅草稿状态）
- `POST /api/prescriptions/:id/confirm` - 确认处方
- `POST /api/prescriptions/:id/void` - 作废处方
- `POST /api/prescriptions/:id/dispense` - 发药
- `GET /api/prescriptions` - 获取所有处方
- `GET /api/registrations/:registration_id/prescriptions` - 获取患者处方

### 药材相关
- `GET /api/herbs` - 获取药材列表
- `PUT /api/herbs` - 更新药材信息

### WebSocket
- `GET /ws` - WebSocket连接，接收叫号消息

## 预置药材数据

系统内置15味常用药材，包含：
- 正常效期：当归、黄芪、白术、茯苓、川芎、白芍、人参、枸杞、菊花、金银花、陈皮、半夏
- 临期（20天）：甘草、薄荷
- 过期：熟地

## 注意事项

1. 数据存储在内存中，服务重启后数据丢失
2. 处方中的药材不能使用过期药材
3. 发药时需检查所有药材库存，库存不足则整个处方不能发药
4. 价格变更仅影响未发药的处方
