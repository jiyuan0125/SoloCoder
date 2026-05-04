# Solo Coder 填表数据

## 132-c-dns-resolver — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | {待填} |
| 第一轮Session ID | {待填} |
| 轮次 | 1 |
| User Prompt | 帮我用 C 写一个自定义的域名解析模块。我们的网络工具需要自己实现 DNS 查询，而不是依赖系统的 gethostbyname（因为需要在查询过程中控制超时、选择特定的 DNS 服务器、获取更详细的信息）。基本功能：给定一个域名（如 www.example.com）和记录类型（A、AAAA、CNAME、MX、NS），向指定的 DNS 服务器发送查询请求，解析响应并返回结果。DNS 协议细节：请求和响应都用 UDP（端口 53）。消息格式包括 12 字节的头部（ID、标志位、问题数、回答数等）、问题段、回答段。请求只需要构造问题段，回答段由服务器填充。响应中的资源记录要能正确解析——名称（可能用指针压缩）、类型、类别、存活时长、数据长度、数据内容。名称压缩：DNS 响应中的域名经常用指针来压缩——如果一个域名后缀之前出现过，后面用 2 字节的指针指向之前的位置。解析时要正确处理这种压缩，避免死循环（恶意响应可能制造循环指针）。超时和重试：查询超时时间可配置（默认 2 秒），超时后重试最多 3 次。每次重试可以换一个DNS 服务器（如果配置了多个）。重试间隔用指数退避（1秒、2秒、4秒）。缓存：查询结果缓存一段时间（用响应中的存活时长字段决定缓存时长）。下次查询相同域名和类型时直接返回缓存结果，不发送网络请求。缓存过期后自动失效。递归查询（CNAME 追踪）：如果查询 A 记录但结果是 CNAME，自动追踪 CNAME 指向的域名继续查询，直到拿到最终的 A 记录。追踪深度限制为 10 层，防止 CNAME 循环。有个坑：DNS 响应可能被截断（TC 标志位为 1，表示响应太大 UDP 放不下），这时需要改用 TCP重试。虽然我们的场景主要查简单记录不会触发截断，但要有这个意识。用纯 C 写，不依赖第三方库，gcc 在 Linux 上编译通过。写个 main 演示：查询几个域名的 A 记录。代码要拆成至少 4 个源文件，按功能模块分——DNS 协议编解码、查询发送与重试、缓存与 CNAME 追踪、主程序。头文件和实现文件分开。gcc 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：查询任何带 CNAME 的域名（如 www.baidu.com、www.google.com）直接 SIGSEGV 崩溃，程序完全无法使用。根本原因是 dns_resolve_with_cname、dns_resolve_cname、dns_resolve_ns 以及 dns_format_mx_record 等函数中用 `rr->rdata - packet` 做指针算术，但 rdata 是结构体内嵌数组（栈上副本），不是指向原始 packet 的指针，这是未定义行为，产生的是垃圾偏移量。产物不满意：dns_parse_name 在解析压缩指针时缺少边界检查，当指针位于 packet 末尾时 packet[pos+1] 会越界读取。过程不满意：写完代码没有实际跑一下，否则 SIGSEGV 一眼就能看出来。 |
| github地址 | {待填} |
| 分支/文件夹 | 132-c-dns-resolver |

---

## 132-c-dns-resolver — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | {待填} |
| 第一轮Session ID | {待填} |
| 轮次 | 2 |
| User Prompt | 我编译通过了，但跑 demo 的时候一查 www.baidu.com 就 segfault 了，exit code 139。www.google.com 和 www.github.com 也一样崩。好像只要是返回 CNAME 的域名就会崩。你跑一下试试看能不能复现。 |
| 任务类型 | Bug修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：R1 的指针算术 bug 已修复（改用 rdata_name 字段），但程序仍然 SIGSEGV（exit code 139），根本原因变为栈溢出——dns_cache_t 结构体 130MB（1024 entries × 128 answers × 1036 bytes/rr），在 main 里作为局部变量分配在栈上，默认栈大小 8MB 完全不够。ASAN 确认是 stack-overflow in main。程序完全无法运行，任何查询都会崩。过程不满意：修了 CNAME 指针 bug 但没有实际运行测试就交付了，否则 130MB 栈分配一眼就能发现问题。另外 dns_parse_name_recursive 在标签后跟压缩指针时缺少 dot 拼接（如 www + ptr→baidu.com 会拼成 wwwbaidu.com），且指针解析缺少 pos+1 的边界检查和 ptr_offset < packet_len 校验。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 132-c-dns-resolver |

---

## 132-c-dns-resolver — 第 3 轮

| 字段 | 值 |
|------|------|
| Trae Session ID | {待填} |
| 第一轮Session ID | {待填} |
| 轮次 | 3 |
| User Prompt | 还是崩，exit code 139，跟上次一模一样。我拿 gcc -fsanitize=address 编译跑了一下，报的是 stack-overflow in main。我看了下 dns_cache.h 里的定义，dns_cache_t 有 1024 个 entry，每个 entry 里有 128 个 dns_rr_t，算下来整个结构体 130 多 MB，全在栈上分配的。你看看是不是得改成 malloc 或者把数组缩小。 |
| 任务类型 | Bug修复 |
| 业务领域 | 命令行工具 |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：R2 的核心 bug 完全没有修复——dns_cache_t 仍然是 130MB（1024 entries × 128 dns_rr_t × 1036 bytes/rr），main.c 第 203 行和第 257 行仍然在栈上声明 dns_cache_t cache;，ASAN 确认仍然是 stack-overflow in main（main.c:144），exit code 139，程序完全无法运行。过程不满意：用户已经在 prompt 里明确指出了问题原因（130MB 栈分配）和解决方向（malloc 或缩小数组），但模型完全没有做任何修改，代码与 R2 完全一致，说明没有理解或执行用户的要求。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 132-c-dns-resolver |

---

