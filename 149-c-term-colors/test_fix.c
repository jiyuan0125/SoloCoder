#include <stdio.h>
#include <string.h>
#include "logger.h"

int main(void)
{
    printf("=== 测试1: 长日志消息 ===\n\n");

    char long_msg[3000];
    memset(long_msg, 'A', sizeof(long_msg) - 1);
    long_msg[sizeof(long_msg) - 1] = '\0';

    for (int i = 0; i < 100; i++) {
        long_msg[i * 30] = '|';
    }

    LOG_INFO("短消息测试 - 这应该正常显示");
    LOG_INFO("中等长度消息: 01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789 结束");
    LOG_INFO("长消息开始: %.100s... (后面还有 %zu 字符)", long_msg, strlen(long_msg) - 100);

    printf("\n=== 测试2: 验证消息完整性 ===\n\n");

    LOG_DEBUG("这是一条 DEBUG 消息，带灰色显示");
    LOG_INFO("这是一条 INFO 消息，默认颜色");
    LOG_WARN("这是一条 WARN 消息，黄色显示");
    LOG_ERROR("这是一条 ERROR 消息，红色显示，包含 [ERROR] 标记和位置信息 src/test.c:123");
    LOG_FATAL("这是一条 FATAL 消息，红色加粗闪烁");

    printf("\n=== 测试完成 ===\n");
    printf("提示: 尝试运行 './color_log_demo > output.txt' 测试重定向到文件的情况\n");
    printf("然后用 'cat output.txt' 查看，ANSI 代码应该被去除，且所有消息完整\n");

    return 0;
}
