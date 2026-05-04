#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include "scheduler.h"

static pthread_mutex_t log_mutex = PTHREAD_MUTEX_INITIALIZER;

static void log_task(const char *task_name, const char *action, long delay_ms) {
    pthread_mutex_lock(&log_mutex);
    
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    
    static struct timespec start_time = {0, 0};
    if (start_time.tv_sec == 0 && start_time.tv_nsec == 0) {
        start_time = ts;
    }
    
    long elapsed_ms = timespec_diff_ms(ts, start_time);
    double elapsed_sec = elapsed_ms / 1000.0;
    
    if (delay_ms >= 0) {
        printf("[%6.2fs] %s: %s (expected delay: %ld ms)\n", 
               elapsed_sec, task_name, action, delay_ms);
    } else {
        printf("[%6.2fs] %s: %s\n", 
               elapsed_sec, task_name, action);
    }
    
    fflush(stdout);
    pthread_mutex_unlock(&log_mutex);
}

static void once_task_1(void *arg) {
    long expected_delay = (long)(size_t)arg;
    log_task("一次性任务-1秒", "已执行", expected_delay);
}

static void once_task_3(void *arg) {
    long expected_delay = (long)(size_t)arg;
    log_task("一次性任务-3秒", "已执行", expected_delay);
}

static void once_task_5(void *arg) {
    long expected_delay = (long)(size_t)arg;
    log_task("一次性任务-5秒", "已执行", expected_delay);
}

static void periodic_cpu_metric(void *arg) {
    static int count = 0;
    count++;
    log_task("CPU指标采集", "执行中", 0);
    printf("              -> 第 %d 次采集完成\n", count);
    fflush(stdout);
}

static void periodic_mem_metric(void *arg) {
    static int count = 0;
    count++;
    log_task("内存指标采集", "执行中", 0);
    printf("              -> 第 %d 次采集完成\n", count);
    fflush(stdout);
}

static void periodic_5_times(void *arg) {
    long *data = (long *)arg;
    static int count = 0;
    count++;
    log_task("限制5次任务", "执行中", 0);
    printf("              -> 第 %d/%ld 次执行\n", count, data[1]);
    fflush(stdout);
}

static void long_running_task(void *arg) {
    log_task("耗时任务", "开始执行", 0);
    printf("              -> 模拟耗时操作，等待 2 秒...\n");
    fflush(stdout);
    sleep(2);
    log_task("耗时任务", "执行完成", 0);
}

static void cleanup_task(void *arg) {
    log_task("清理任务", "执行中", 0);
}

int main(void) {
    printf("========================================\n");
    printf("定时任务调度引擎演示\n");
    printf("========================================\n\n");
    
    printf("创建调度器 (4个工作线程)...\n");
    scheduler_t *sched = scheduler_create(4, 128);
    if (sched == NULL) {
        fprintf(stderr, "错误：无法创建调度器\n");
        return 1;
    }
    printf("调度器创建成功！\n\n");
    
    printf("----------------------------------------\n");
    printf("阶段 1: 注册多个一次性任务\n");
    printf("----------------------------------------\n\n");
    
    task_id_t t1 = scheduler_add_once(sched, "一次性任务-1秒", 1000, once_task_1, (void *)(size_t)1000);
    task_id_t t2 = scheduler_add_once(sched, "一次性任务-3秒", 3000, once_task_3, (void *)(size_t)3000);
    task_id_t t3 = scheduler_add_once(sched, "一次性任务-5秒", 5000, once_task_5, (void *)(size_t)5000);
    
    printf("已注册一次性任务：\n");
    printf("  - ID=%llu: 1秒后执行\n", (unsigned long long)t1);
    printf("  - ID=%llu: 3秒后执行\n", (unsigned long long)t2);
    printf("  - ID=%llu: 5秒后执行\n", (unsigned long long)t3);
    printf("\n");
    
    printf("----------------------------------------\n");
    printf("阶段 2: 注册周期性任务\n");
    printf("----------------------------------------\n\n");
    
    long interval1 = 2000;
    task_id_t p1 = scheduler_add_periodic(sched, "CPU指标采集", 0, interval1, 0, periodic_cpu_metric, NULL);
    
    long interval2 = 3000;
    task_id_t p2 = scheduler_add_periodic(sched, "内存指标采集", 1000, interval2, 0, periodic_mem_metric, NULL);
    
    static long limited_data[] = {4000, 5};
    task_id_t p3 = scheduler_add_periodic(sched, "限制5次任务", 2000, limited_data[0], limited_data[1], 
                                           periodic_5_times, limited_data);
    
    printf("已注册周期性任务：\n");
    printf("  - ID=%llu: CPU指标采集，每2秒1次（无限）\n", (unsigned long long)p1);
    printf("  - ID=%llu: 内存指标采集，每3秒1次，初始延迟1秒\n", (unsigned long long)p2);
    printf("  - ID=%llu: 限制5次任务，每4秒1次，最多5次\n", (unsigned long long)p3);
    printf("\n");
    
    printf("----------------------------------------\n");
    printf("阶段 3: 演示任务取消\n");
    printf("----------------------------------------\n\n");
    
    task_id_t to_cancel = scheduler_add_once(sched, "被取消的任务", 20000, cleanup_task, NULL);
    printf("注册了一个20秒后执行的任务 (ID=%llu)\n", (unsigned long long)to_cancel);
    
    sleep(2);
    
    printf("\n现在取消该任务...\n");
    int cancel_result = scheduler_cancel(sched, to_cancel);
    if (cancel_result == 0) {
        printf("任务 %llu 已成功取消！\n", (unsigned long long)to_cancel);
    } else {
        printf("取消任务失败！\n");
    }
    printf("\n");
    
    printf("----------------------------------------\n");
    printf("阶段 4: 演示耗时任务与并发\n");
    printf("----------------------------------------\n\n");
    
    printf("注册一个耗时任务（模拟需要2秒完成的操作）...\n");
    task_id_t long_task = scheduler_add_once(sched, "耗时任务", 5000, long_running_task, NULL);
    printf("耗时任务已注册 (ID=%llu)\n", (unsigned long long)long_task);
    
    printf("\n同时，周期性任务会继续执行，不会被阻塞...\n\n");
    
    printf("========================================\n");
    printf("现在观察任务执行 20 秒...\n");
    printf("========================================\n\n");
    
    for (int i = 0; i < 20; i++) {
        sleep(1);
        size_t pending = scheduler_task_count(sched);
        size_t queue_len = thread_pool_pending_count(sched->pool);
        pthread_mutex_lock(&log_mutex);
        printf("[状态] 待调度任务: %zu, 队列中: %zu\n", pending, queue_len);
        fflush(stdout);
        pthread_mutex_unlock(&log_mutex);
    }
    
    printf("\n========================================\n");
    printf("取消周期性任务并结束演示\n");
    printf("========================================\n\n");
    
    scheduler_cancel(sched, p1);
    scheduler_cancel(sched, p2);
    printf("已取消无限执行的周期性任务...\n");
    printf("等待3秒让所有任务完成...\n\n");
    
    sleep(3);
    
    printf("销毁调度器...\n");
    scheduler_destroy(sched);
    
    printf("\n========================================\n");
    printf("演示结束！\n");
    printf("========================================\n\n");
    
    printf("总结：\n");
    printf("1. 一次性任务：按指定时间执行，执行后自动移除\n");
    printf("2. 周期性任务：按间隔反复执行，支持最大执行次数限制\n");
    printf("3. 任务取消：通过ID可取消未执行的任务\n");
    printf("4. 并发执行：耗时任务在线程池中执行，不阻塞其他任务\n");
    printf("5. 最小堆优化：最近到期的任务始终在堆顶，O(1)时间复杂度获取\n");
    printf("6. 动态精度：根据任务到期时间自动调整检查频率\n");
    
    return 0;
}
