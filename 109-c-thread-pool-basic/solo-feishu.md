# Solo Coder 填表数据

## 109-c-thread-pool-basic — 第 1 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 1 |
| User Prompt | 帮我用 C 写一个通用的工作线程池模块，用于并发处理耗时的计算任务。比如一个影像处理服务，收到请求后把任务丢给线程池异步执行，主线程继续接收新请求不被阻塞。线程池在创建时指定线程数量和任务缓冲区容量。缓冲区满了的时候，新任务有两种策略可选：阻塞等待（直到缓冲区有空位）或者立即拒绝（丢弃任务并通知调用方）。策略在创建线程池时配置，运行中不能改。每个任务是一个回调指针加一个 void 参数。任务执行完毕后，调用方需要知道结果——支持设置完成回调，任务完成后自动调用。回调在哪个线程执行可以配置（工作线程直接调用或者由调用方自行处理）。优雅关闭：调用关闭后，线程池不再接受新任务，但会等待缓冲区中已有的任务全部执行完毕才真正退出所有线程。等待时间可以设一个超时，超时后强制退出并丢弃未执行的任务。另外支持一个紧急关闭模式，直接丢弃所有未执行任务并立即退出。状态查询：随时能拿到当前线程池的状态——活跃线程数、等待中的任务数、已完成的总任务数、被拒绝的任务数。有个隐藏坑：任务执行时如果崩溃了（比如段错误），不能让整个线程池挂掉，出问题的线程要能恢复或者被替换掉，线程池继续正常工作。用纯 C 写，不依赖第三方库，gcc -lpthread 在 Linux 上编译通过。写个 main 演示：创建4 个线程的池子，提交 20 个任务，观察执行顺序和并发情况。代码要拆成至少 4 个源文件，按功能模块分——任务队列管理、工作线程调度、状态统计、主程序。头文件和实现文件分开。gcc -lpthread 编译能过。 |
| 任务类型 | 0-1代码生成 |
| 业务领域 | 库/SDK |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 未完成 |
| 产物及过程是否满意 | 不满意 |
| 不满意原因 | 产物不满意：(1) TP_CALLBACK_DEFERRED 模式未实现——common.h 定义了该枚举值，thread_pool.c 的 worker_thread 只处理了 TP_CALLBACK_IN_WORKER，设置 TP_CALLBACK_DEFERRED 的任务完成后回调直接被丢弃，调用方永远收不到通知。(2) Crash recovery 后不调用 completion callback——PROMPT 要求"调用方需要知道结果"，worker_thread 中 siglongjmp 恢复后执行 continue 直接跳过了 completion callback 调用，任务崩溃时调用方无法通过回调得知失败。运行输出证实：Task #9 和 #16 崩溃后未出现 [Callback] Task completed with failure。(3) active_threads 语义不对——PROMPT 要求"活跃线程数"，代码实现为存活线程数（worker 创建到退出始终为 num_threads），所有任务执行完毕后 active_threads 仍为 4，更合理的语义应是目前正在执行任务的线程数。过程不满意：TP_CALLBACK_DEFERRED 和 crash recovery callback 是 PROMPT 中明确描述的功能需求，代码中定义了接口但未实现完整逻辑，说明编写时未逐条对照需求检查。 |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 109-c-thread-pool-basic |

---

## 109-c-thread-pool-basic — 第 2 轮

| 字段 | 值 |
|------|------|
| Trae Session ID |  |
| 第一轮Session ID |  |
| 轮次 | 2 |
| User Prompt | 我测了一下线程池，发现几个问题：1. 任务设了 TP_CALLBACK_DEFERRED 模式，但跑完之后回调根本没被调用，感觉这个模式压根没实现 2. 我故意让几个任务 crash 了（段错误），线程池确实没挂，但是 completion callback 也没被触发，调用方完全不知道任务失败了 3. 看了下 stats 输出，所有任务跑完了 active_threads 还是 4，这个应该是"正在执行任务的线程数"才对吧？不应该是一直等于线程池大小 |
| 任务类型 | Bug修复 |
| 业务领域 | 库/SDK |
| 修改范围 | 跨模块多文件 |
| 任务是否完成 | 已完成 |
| 产物及过程是否满意 | 满意 |
| 不满意原因 |  |
| github地址 | https://github.com/jiyuan0125/SoloCoder |
| 分支/文件夹 | 109-c-thread-pool-basic |

---
