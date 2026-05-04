# 122-c-atomic-counter Solo Check Report

## 项目信息

| 字段 | 值 |
|------|-----|
| 项目ID | 122 |
| 项目名称 | c-atomic-counter |
| 技术栈 | C (gcc -lpthread) |
| 业务领域 | 库/SDK |
| 评测轮次 | 1 |
| 任务类型 | 0-1代码生成 |
| 修改范围 | 跨模块多文件 |
| Trae Session ID | |
| User Prompt | 见 PROMPT.txt 原文 |

## 评测结果总览

| 字段 | 值 |
|------|-----|
| 任务是否完成 | ❌ 未完成 |
| 产物及过程是否满意 | ❌ 不满意 |
| 不满意原因 | 见下方详细分析 |

## 编译与运行

| 项目 | 结果 |
|------|------|
| 编译命令 | `gcc -Wall -Wextra -o /tmp/solo-check-122 *.c -lpthread` |
| 编译结果 | ✅ 通过（1 warning: unused function `collect_expired_callback`） |
| 运行结果 | ✅ 正常运行，两个测试均通过 |
| 固定窗口测试 | ✅ 800 请求全部正确计数，50 允许 ≤ 50 阈值上限 |
| 滑动窗口测试 | ✅ 800 请求全部正确计数，50 允许 ≤ 50 阈值上限 |

## 需求对照表

| # | 需求 | 状态 | 备注 |
|---|------|------|------|
| 1 | 基本计数（原子加1、多线程安全） | ✅ 通过 | `__sync_fetch_and_add` + spinlock 双重保护 |
| 2 | 固定窗口限流（每分钟前N个通过） | ✅ 通过 | `fixed_window_check_and_inc` 实现正确 |
| 3 | 滑动窗口限流（6个10秒桶） | ✅ 通过 | `sliding_window_check_and_inc` 综合最近6桶计数 |
| 4 | key 自动清理（5分钟过期） | ✅ 通过 | 后台线程 + `cleanup_expired_keys` |
| 5 | 统计：活跃key数 | ✅ 通过 | `active_keys_count` 原子计数 |
| 6 | 统计：总请求量 | ❌ 未通过 | `global_total_requests` 只计数 allowed 请求，不是总请求量 |
| 7 | 统计：被拒绝请求量 | ✅ 通过 | `global_rejected_requests` 正确 |
| 8 | 统计：key 近1分钟请求次数 | ✅ 通过 | `get_key_stats` 滑动窗口正确 |
| 9 | 统计：key 近1分钟被拒绝次数 | ⚠️ 部分 | 滑动窗口正确，固定窗口始终返回 0 |
| 10 | 性能（CPU原子指令，非每次加锁） | ✅ 通过 | `__sync_fetch_and_add` 无锁全局计数 |
| 11 | 隐藏坑：窗口时间对齐 xx:x0 | ✅ 通过 | `get_aligned_bucket_time(now - now % 10)` |
| 12 | 纯C，gcc 在 Linux 编译通过 | ✅ 通过 | 有 1 个 warning |
| 13 | main 多线程演示 | ✅ 通过 | 8线程×5IP，限流验证通过 |
| 14 | ≥4个源文件，按功能模块分 | ✅ 通过 | counter.c / sliding_window.c / cleanup.c / main.c |
| 15 | 头文件和实现文件分开 | ✅ 通过 | 4组 .h/.c + common.h |

## Bug 清单

### Bug 1：`global_total_requests` 语义错误（严重）

- **文件**：`sliding_window.c` (L55-58, L95-98), `counter.h` (L44)
- **问题**：PROMPT 要求"总请求量"统计，但 `global_total_requests` 只在请求被允许时递增（`atomic_inc(&limiter->global_total_requests)` 仅在 `allowed=true` 分支）。`get_global_stats` 将该值赋给 `stats->total_requests`，返回的不是总请求量而是允许通过的请求数。
- **影响**：调用方获取的"总请求量"数据错误。demo 中 `print_stats` 需要手动相加 `total_requests + rejected_requests` 才能得到真实总量，说明开发者自己也意识到了字段语义不一致。
- **修复**：每次请求（无论允许或拒绝）都递增 `global_total_requests`，或者将字段改名为 `global_allowed_requests`。

### Bug 2：固定窗口模式 `rejected_last_minute` 始终为 0（中等）

- **文件**：`counter.h` (L24-27), `sliding_window.c` (L142-150)
- **问题**：`fixed` 窗口结构体只有 `requests`（计数允许通过的请求）和 `window_start`，没有 rejected 计数器。`get_key_stats` 对固定窗口硬编码 `stats->rejected_last_minute = 0`。
- **影响**：固定窗口模式下，per-key 的"最近1分钟被拒绝次数"永远为 0，PROMPT 要求提供该统计。
- **修复**：在 `fixed` 结构体中增加 `volatile uint32_t rejected` 字段，在 `fixed_window_check_and_inc` 拒绝时递增。

### Bug 3：死代码 `collect_expired_callback`（低）

- **文件**：`cleanup.c` (L77-83)
- **问题**：函数定义但从未被调用。实际清理逻辑直接遍历哈希表而非通过 `counter_traverse` + callback。
- **影响**：编译产生 `-Wunused-function` warning，代码整洁度差。
- **修复**：删除该函数。

### Bug 4：`counter_get_or_create` 竞态路径资源泄漏（低）

- **文件**：`counter.c` (L114-117)
- **问题**：两个线程同时为同一个 key 创建 counter 时，后到的线程发现 key 已存在，执行 `free(new_counter)` 但未先调用 `pthread_spin_destroy(&new_counter->spinlock)`。
- **影响**：spinlock 资源泄漏（仅在高并发竞态时触发，概率低）。
- **修复**：在 `free(new_counter)` 前加 `pthread_spin_destroy(&new_counter->spinlock)`。

### Bug 5：编译产物未加入 `.gitignore`（低）

- **文件**：`.gitignore`
- **问题**：目录下存在编译生成的 `rate_limiter` 二进制文件，但 `.gitignore` 未包含该文件名。
- **修复**：在 `.gitignore` 中添加 `rate_limiter`。

## 代码质量评价

| 维度 | 评分 | 说明 |
|------|------|------|
| 架构设计 | 8/10 | 哈希表分桶 + 读写锁 + per-key spinlock，层次清晰 |
| 并发安全 | 8/10 | 核心路径正确，双检锁模式正确使用，竞态路径有 minor 泄漏 |
| 原子操作 | 9/10 | `__sync_fetch_and_add` 用于全局计数，spinlock 用于 per-key 窗口操作 |
| 代码整洁 | 6/10 | 死代码、命名不一致（total_requests 实为 allowed）、编译 warning 未处理 |
| PROMPT 覆盖 | 7/10 | 核心功能完整，统计模块有两处缺陷 |
| 综合 | 7.6/10 | 核心限流功能扎实，统计和代码细节有缺陷 |

## 不满意原因

### 产物不满意
1. `global_total_requests` 只计数允许的请求，`get_global_stats` 返回的 `total_requests` 不是 PROMPT 要求的"总请求量"，属于语义错误
2. 固定窗口 `fixed` 结构体缺少 rejected 计数器，导致 `get_key_stats` 的 `rejected_last_minute` 始终为 0
3. `collect_expired_callback` 死代码未清理
4. `counter_get_or_create` 竞态路径 free 前未调 `pthread_spin_destroy`
5. 编译产物 `rate_limiter` 二进制未加入 `.gitignore`

### 过程不满意
编译有 `-Wunused-function` warning 但未处理，说明没有检查编译输出或选择忽略。作为底层库/SDK 代码，编译零 warning 是基本要求。
