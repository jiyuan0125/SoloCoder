# Solo Coder 填表数据

## 327-electron-api-debugger — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 做一个 HTTP 请求调试工具，类似简化版 Postman。主界面上面是请求配置区（方法选择 GET/POST/PUT/DELETE、URL 输入框），下面分左右两栏：左边是请求参数编辑（Headers 键值对、Body 编辑器支持 JSON/表单/纯文本），右边是响应展示（状态码、响应头、响应体，JSON 自动格式化带语法高亮）。支持保存请求到历史记录，历史列表在左侧边栏，可以搜索、删除、收藏。环境变量功能：用户可以定义多组环境变量（比如 dev、staging、prod），在 URL 和 Headers 里用 {{variable}} 占位符引用，切换环境时自动替换。发送请求前先把占位符全部替换成实际值。Body 编辑器在 JSON 模式下提供基本的语法校验，格式不正确时标红提示。大响应体（超过 1MB）只显示前 100KB 并提示内容被截断。支持发送 FormData（文件上传），可以添加多个文件字段。请求超时默认 30 秒，可以在设置里调整。所有请求历史和环境变量配置持久化到本地 JSON 文件。界面用 HTML+CSS+JS 渲染，HTTP 请求发送走后端处理，前后端通过 Electron IPC 通信。代码分文件组织，package.json、main.js、前端页面齐全，npm install 后 npm start 能正常使用。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 桌面应用（含GUI） |
| 修改范围 | 模块内多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 | 无 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 327-electron-api-debugger |

---
