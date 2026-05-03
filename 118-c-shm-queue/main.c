#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <signal.h>
#include <errno.h>
#include "shm_queue.h"

#define TEST_QUEUE_NAME   "/shmq_demo"
#define TEST_BUFFER_SIZE  (1024 * 1024)
#define TEST_MAX_MSG_LEN  65536
#define TEST_MSG_COUNT    10

static volatile sig_atomic_t g_sigint_received = 0;

static void sigint_handler(int sig) {
    (void)sig;
    g_sigint_received = 1;
}

static void setup_signal_handler(void) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = sigint_handler;
    sigaction(SIGINT, &sa, NULL);
    sigaction(SIGTERM, &sa, NULL);
}

static int writer_process(shm_queue_t *q) {
    int i;
    char msg[256];
    int ret;
    
    printf("[Writer] 开始执行，PID: %d\n", getpid());
    
    printf("[Writer] 开始发送 %d 条消息...\n", TEST_MSG_COUNT);
    
    for (i = 0; i < TEST_MSG_COUNT; i++) {
        if (g_sigint_received) {
            printf("[Writer] 收到中断信号，停止发送\n");
            break;
        }
        
        snprintf(msg, sizeof(msg), "Message #%d from writer (PID: %d)", i + 1, getpid());
        
        printf("[Writer] 发送消息 %d: \"%s\"\n", i + 1, msg);
        
        ret = shmq_send(q, msg, strlen(msg) + 1, SHMQ_WAIT_INFINITE);
        if (ret != SHMQ_OK) {
            fprintf(stderr, "[Writer] 发送消息失败: %d\n", ret);
            break;
        }
        
        usleep(100000);
    }
    
    const char *shutdown_msg = "SHUTDOWN";
    printf("[Writer] 发送终止消息: \"%s\"\n", shutdown_msg);
    shmq_send(q, shutdown_msg, strlen(shutdown_msg) + 1, SHMQ_WAIT_INFINITE);
    
    printf("[Writer] 等待读取者处理完成...\n");
    sleep(1);
    
    printf("[Writer] 进程结束\n");
    return 0;
}

static int reader_process(shm_queue_t *q) {
    char buf[1024];
    size_t len;
    int ret;
    int msg_count = 0;
    
    printf("[Reader] 开始执行，PID: %d\n", getpid());
    
    size_t max_len, buf_size;
    shmq_get_max_msg_len(q, &max_len);
    shmq_get_buffer_size(q, &buf_size);
    printf("[Reader] 队列信息: 缓冲区=%zu bytes, 最大消息=%zu bytes\n", 
           buf_size, max_len);
    
    printf("[Reader] 开始接收消息...\n");
    
    while (!g_sigint_received) {
        len = sizeof(buf);
        ret = shmq_recv(q, buf, &len, 5000);
        
        if (ret == SHMQ_TIMEOUT) {
            continue;
        }
        
        if (ret != SHMQ_OK) {
            fprintf(stderr, "[Reader] 接收消息失败: %d\n", ret);
            break;
        }
        
        msg_count++;
        printf("[Reader] 收到消息 %d: \"%s\" (长度=%zu bytes)\n", 
               msg_count, buf, len);
        
        if (strcmp(buf, "SHUTDOWN") == 0) {
            printf("[Reader] 收到终止信号，退出循环\n");
            break;
        }
    }
    
    printf("[Reader] 共收到 %d 条消息\n", msg_count);
    printf("[Reader] 进程结束\n");
    return 0;
}

int main(int argc, char *argv[]) {
    shm_queue_t q;
    shm_queue_config_t cfg;
    pid_t pid;
    int status;
    int ret;
    
    setup_signal_handler();
    
    printf("=== 共享内存消息队列演示程序 ===\n");
    printf("功能: 父进程写入消息，子进程读取消息\n");
    printf("按 Ctrl+C 可以提前终止\n");
    printf("================================\n\n");
    
    printf("清理可能残留的资源...\n");
    shmq_destroy(TEST_QUEUE_NAME);
    
    printf("创建队列...\n");
    shmq_config_init(&cfg);
    cfg.name = TEST_QUEUE_NAME;
    cfg.buffer_size = TEST_BUFFER_SIZE;
    cfg.max_msg_len = TEST_MAX_MSG_LEN;
    
    ret = shmq_create(&q, &cfg);
    if (ret != SHMQ_OK) {
        fprintf(stderr, "创建队列失败: %d\n", ret);
        return -1;
    }
    
    printf("队列创建成功，名称: %s, 缓冲区大小: %zu bytes\n", 
           cfg.name, cfg.buffer_size);
    
    printf("fork 子进程...\n");
    pid = fork();
    
    if (pid < 0) {
        perror("fork 失败");
        shmq_close(&q);
        shmq_destroy(TEST_QUEUE_NAME);
        return -1;
    }
    
    if (pid == 0) {
        return reader_process(&q);
    } else {
        printf("[Main] 父进程 PID: %d, 子进程 PID: %d\n", getpid(), pid);
        printf("\n");
        
        int writer_ret = writer_process(&q);
        
        printf("\n[Main] 等待子进程结束...\n");
        waitpid(pid, &status, 0);
        
        if (WIFEXITED(status)) {
            printf("[Main] 子进程退出码: %d\n", WEXITSTATUS(status));
        } else if (WIFSIGNALED(status)) {
            printf("[Main] 子进程被信号终止: %d\n", WTERMSIG(status));
        }
        
        printf("[Main] 关闭队列...\n");
        shmq_close(&q);
        
        printf("\n[Main] 清理资源...\n");
        shmq_destroy(TEST_QUEUE_NAME);
        printf("[Main] 资源已清理\n");
        
        printf("\n=== 演示程序结束 ===\n");
        printf("返回值: writer=%d, reader=%d\n", 
               writer_ret, WIFEXITED(status) ? WEXITSTATUS(status) : -1);
        
        return writer_ret;
    }
}
