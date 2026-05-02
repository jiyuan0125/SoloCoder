# Solo Coder 填表数据

## 53-rust-web-crawler — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | .335769888099319:937a10e542a8db3b7258376117032b5c_69f54b3c9dbef3d8838bc34d.69f54c529dbef3d8838bc428.69f54c4d160b3d7f9ace029e:Trae CN.T(2026/5/2 08:58:58) |
| 第一轮Session ID | .335769888099319:937a10e542a8db3b7258376117032b5c_69f54b3c9dbef3d8838bc34d.69f54c529dbef3d8838bc428.69f54c4d160b3d7f9ace029e:Trae CN.T(2026/5/2 08:58:58) |
| 轮次 | 1 |
| User Prompt | 用 Rust 写一个站点镜像工具，把目标网站下载到本地目录，保持原始的目录结构。从种子 URL 开始，只爬取同域名下的资源。 路径映射： 1. URL 路径映射到本地文件路径：https://example.com/about/team → about/team.html，如果 URL 末尾没有文件扩展名则自动加 .html 2. URL 中的 query string 和 fragment 在映射文件路径时要去掉：/search?q=test → search.html（不是 search?q=test.html） 3. 如果路径包含 .. 要先规范化再映射（/a/../b → /b → b.html） HTML 处理： 4. 下载的 HTML 中所有同域名的链接改写为相对路径：<a href="https://example.com/contact"> 改成 <a href="../contact.html">，外部链接保持原样不改动 5. 静态资源（CSS/JS/图片）也下载并改写路径 爬取控制： 6. 遵守 robots.txt 的 Disallow 规则 7. 并发限制：同一个域名最多 3 个并发连接 8. URL 去重，已下载的不重复请求 9. 支持深度限制，默认 3 层 输出： 10. 完成后打印统计：页面数、资源数（CSS/JS/图片）、失败数、总耗时秒数 代码分下载调度、路径映射、HTML 链接改写、robots 解析几个模块，cargo build 能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：根URL路径映射bug（种子URL为 https://example.com/ 时保存到 ./output.html 而非 ./output/index.html，文件在输出目录外）。Cargo.toml 中 tokio 和 lazy_static 两个依赖未在代码中使用 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 53-rust-web-crawler |

---
