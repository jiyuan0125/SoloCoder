#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <inttypes.h>

#include "common.h"
#include "counter.h"
#include "sliding_window.h"
#include "cleanup.h"

#define NUM_THREADS       8
#define REQUESTS_PER_IP   20
#define NUM_IPS           5

typedef struct {
    rate_limiter_t *limiter;
    int thread_id;
    const char **ips;
    int num_ips;
    int requests_per_ip;
    volatile uint64_t *local_allowed;
    volatile uint64_t *local_rejected;
} thread_args_t;

static void* worker_thread(void *arg) {
    thread_args_t *args = (thread_args_t*)arg;
    rate_limiter_t *limiter = args->limiter;
    
    uint64_t allowed = 0;
    uint64_t rejected = 0;
    
    for (int r = 0; r < args->requests_per_ip; r++) {
        for (int i = 0; i < args->num_ips; i++) {
            bool result = rate_limit_check_and_increment(limiter, args->ips[i]);
            if (result) {
                allowed++;
            } else {
                rejected++;
            }
        }
    }
    
    *args->local_allowed = allowed;
    *args->local_rejected = rejected;
    
    return NULL;
}

static void print_header(const char *title) {
    printf("\n");
    printf("============================================================\n");
    printf("  %s\n", title);
    printf("============================================================\n");
}

static void print_stats(rate_limiter_t *limiter, const char **ips, int num_ips) {
    global_stats_t g_stats;
    get_global_stats(limiter, &g_stats);
    
    printf("\n【全局统计】\n");
    printf("  活跃Key数:     %" PRIu32 "\n", g_stats.active_keys);
    printf("  总请求数:       %" PRIu64 "\n", g_stats.total_requests);
    printf("  被拒绝请求数:   %" PRIu64 "\n", g_stats.rejected_requests);
    printf("  总处理请求:     %" PRIu64 "\n", g_stats.total_requests + g_stats.rejected_requests);
    
    printf("\n【各IP统计】\n");
    for (int i = 0; i < num_ips; i++) {
        key_stats_t k_stats;
        if (get_key_stats(limiter, ips[i], &k_stats)) {
            printf("  [%s]\n", ips[i]);
            printf("    近1分钟请求:  %" PRIu64 " (允许:%" PRIu64 ", 拒绝:%" PRIu64 ")\n", 
                   k_stats.requests_last_minute + k_stats.rejected_last_minute,
                   k_stats.requests_last_minute, k_stats.rejected_last_minute);
            printf("    累计允许:     %" PRIu64 "\n", k_stats.total_allowed);
            printf("    累计拒绝:     %" PRIu64 "\n", k_stats.total_rejected);
        }
    }
    printf("\n");
}

static void run_test(rate_limit_mode_t mode, uint32_t threshold, const char *test_name) {
    const char *ips[NUM_IPS] = {
        "192.168.1.1",
        "192.168.1.2",
        "192.168.1.3",
        "192.168.1.4",
        "192.168.1.5"
    };
    
    print_header(test_name);
    printf("配置: 模式=%s, 阈值=%" PRIu32 " 次/分钟\n",
           mode == RATE_LIMIT_FIXED_WINDOW ? "固定窗口" : "滑动窗口",
           threshold);
    printf("线程数: %d, 每个线程发送: %d 轮请求, 每轮 %d 个IP\n",
           NUM_THREADS, REQUESTS_PER_IP, NUM_IPS);
    printf("期望: 每个IP最多允许 %" PRIu32 " 次请求通过\n", threshold);
    
    rate_limiter_t *limiter = rate_limiter_create(mode, threshold);
    if (!limiter) {
        printf("错误: 无法创建限流器\n");
        return;
    }
    
    pthread_t threads[NUM_THREADS];
    thread_args_t args[NUM_THREADS];
    volatile uint64_t allowed_counts[NUM_THREADS];
    volatile uint64_t rejected_counts[NUM_THREADS];
    
    for (int i = 0; i < NUM_THREADS; i++) {
        args[i].limiter = limiter;
        args[i].thread_id = i;
        args[i].ips = ips;
        args[i].num_ips = NUM_IPS;
        args[i].requests_per_ip = REQUESTS_PER_IP;
        args[i].local_allowed = &allowed_counts[i];
        args[i].local_rejected = &rejected_counts[i];
        allowed_counts[i] = 0;
        rejected_counts[i] = 0;
        
        int ret = pthread_create(&threads[i], NULL, worker_thread, &args[i]);
        if (ret != 0) {
            printf("错误: 无法创建线程 %d\n", i);
            rate_limiter_destroy(limiter);
            return;
        }
    }
    
    printf("\n【并发请求中...】\n");
    
    for (int i = 0; i < NUM_THREADS; i++) {
        pthread_join(threads[i], NULL);
    }
    
    uint64_t total_allowed = 0;
    uint64_t total_rejected = 0;
    printf("\n【线程局部统计】\n");
    for (int i = 0; i < NUM_THREADS; i++) {
        printf("  线程%d: 允许=%" PRIu64 ", 拒绝=%" PRIu64 "\n", 
               i, allowed_counts[i], rejected_counts[i]);
        total_allowed += allowed_counts[i];
        total_rejected += rejected_counts[i];
    }
    printf("\n  线程统计合计: 允许=%" PRIu64 ", 拒绝=%" PRIu64 ", 总计=%" PRIu64 "\n",
           total_allowed, total_rejected, total_allowed + total_rejected);
    
    print_stats(limiter, ips, NUM_IPS);
    
    uint64_t expected_total = (uint64_t)NUM_THREADS * REQUESTS_PER_IP * NUM_IPS;
    uint64_t max_expected_allowed = (uint64_t)NUM_IPS * threshold;
    
    printf("【验证结果】\n");
    printf("  期望总请求数:   %" PRIu64 "\n", expected_total);
    printf("  实际总请求数:   %" PRIu64 "\n", total_allowed + total_rejected);
    printf("  最大允许请求数: %" PRIu64 " (%" PRIu32 " IPs * %" PRIu32 " 阈值)\n", 
           max_expected_allowed, NUM_IPS, threshold);
    printf("  实际允许请求数: %" PRIu64 "\n", total_allowed);
    
    if (total_allowed + total_rejected == expected_total) {
        printf("  [OK] 所有请求都被正确计数\n");
    } else {
        printf("  [FAIL] 请求计数不一致!\n");
    }
    
    if (total_allowed <= max_expected_allowed) {
        printf("  [OK] 限流生效: 允许请求数不超过阈值\n");
    } else {
        printf("  [FAIL] 限流失败: 允许请求数超过阈值!\n");
    }
    
    rate_limiter_destroy(limiter);
}

int main() {
    printf("\n");
    printf("============================================================\n");
    printf("  高性能请求计数器 - 网关限流演示程序\n");
    printf("============================================================\n");
    
    printf("\n特性说明:\n");
    printf("  - 原子操作: 使用 GCC __sync 系列内置函数实现无锁计数\n");
    printf("  - 固定窗口: 每分钟的前N个请求通过，后续拒绝\n");
    printf("  - 滑动窗口: 6个10秒桶，精确对齐到系统时间边界(xx:x0秒)\n");
    printf("  - 过期清理: 5分钟无活动的key自动移除\n");
    printf("  - 线程安全: 读写锁保护哈希表，原子操作保护计数器\n");
    
    run_test(RATE_LIMIT_FIXED_WINDOW, 10, "测试1: 固定窗口限流");
    run_test(RATE_LIMIT_SLIDING_WINDOW, 10, "测试2: 滑动窗口限流");
    
    printf("\n");
    printf("============================================================\n");
    printf("  演示结束\n");
    printf("============================================================\n");
    printf("\n");
    
    return 0;
}
