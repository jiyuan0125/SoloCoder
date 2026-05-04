#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "rss.h"
#include "time_utils.h"
#include "xml_escape.h"

static time_t create_time(int year, int month, int day, int hour, int min, int sec) {
    struct tm tm;
    memset(&tm, 0, sizeof(tm));
    tm.tm_year = year - 1900;
    tm.tm_mon = month - 1;
    tm.tm_mday = day;
    tm.tm_hour = hour;
    tm.tm_min = min;
    tm.tm_sec = sec;
    tm.tm_isdst = 0;
    return timegm(&tm);
}

int main(int argc, char *argv[]) {
    printf("=== RSS Feed 生成模块演示 ===\n\n");
    
    RSS_Channel channel;
    memset(&channel, 0, sizeof(channel));
    
    strncpy(channel.title, "技术博客 RSS Feed", RSS_MAX_TITLE_LEN - 1);
    strncpy(channel.link, "https://example.com/blog", RSS_MAX_LINK_LEN - 1);
    strncpy(channel.description, 
            "分享技术文章、编程教程和开发经验。关注 Web 开发、系统架构和最佳实践。",
            RSS_MAX_DESCRIPTION_LEN - 1);
    strncpy(channel.language, "zh-CN", RSS_MAX_LANGUAGE_LEN - 1);
    channel.last_build_date = create_time(2026, 5, 4, 14, 30, 0);
    channel.pub_date = create_time(2026, 5, 4, 12, 0, 0);
    
    RSS_Item items[5];
    memset(items, 0, sizeof(items));
    
    strncpy(items[0].title, "深入理解 C 语言中的内存管理", RSS_MAX_TITLE_LEN - 1);
    strncpy(items[0].link, "https://example.com/blog/c-memory-management", RSS_MAX_LINK_LEN - 1);
    strncpy(items[0].description, 
            "本文详细介绍了 C 语言中的动态内存分配机制，包括 malloc、calloc、realloc 和 free 的使用方法，以及常见的内存泄漏问题和调试技巧。",
            RSS_MAX_DESCRIPTION_LEN - 1);
    strncpy(items[0].author, "张三 <zhangsan@example.com>", RSS_MAX_AUTHOR_LEN - 1);
    strncpy(items[0].categories[0], "C语言", RSS_MAX_CATEGORY_LEN - 1);
    strncpy(items[0].categories[1], "内存管理", RSS_MAX_CATEGORY_LEN - 1);
    strncpy(items[0].categories[2], "编程基础", RSS_MAX_CATEGORY_LEN - 1);
    items[0].category_count = 3;
    items[0].pub_date = create_time(2026, 5, 3, 10, 0, 0);
    items[0].is_html_description = 0;
    
    strncpy(items[1].title, "现代 Web 开发中的 API 设计原则", RSS_MAX_TITLE_LEN - 1);
    strncpy(items[1].link, "https://example.com/blog/api-design-principles", RSS_MAX_LINK_LEN - 1);
    strncpy(items[1].description, 
            "<p>好的 API 设计是系统可维护性的关键。本文讨论了 RESTful API 设计的最佳实践，包括：</p>\n<ul>\n<li>资源命名规范</li>\n<li>HTTP 方法的正确使用</li>\n<li>状态码语义</li>\n<li>版本控制策略</li>\n</ul>",
            RSS_MAX_DESCRIPTION_LEN - 1);
    strncpy(items[1].author, "李四 <lisi@example.com>", RSS_MAX_AUTHOR_LEN - 1);
    strncpy(items[1].categories[0], "Web开发", RSS_MAX_CATEGORY_LEN - 1);
    strncpy(items[1].categories[1], "API", RSS_MAX_CATEGORY_LEN - 1);
    items[1].category_count = 2;
    items[1].pub_date = create_time(2026, 5, 2, 15, 30, 0);
    items[1].is_html_description = 1;
    
    strncpy(items[2].title, "测试与调试：从入门到精通", RSS_MAX_TITLE_LEN - 1);
    strncpy(items[2].link, "https://example.com/blog/testing-debugging", RSS_MAX_LINK_LEN - 1);
    strncpy(items[2].description, 
            "测试是软件开发中不可或缺的环节。本文涵盖了单元测试、集成测试、端到端测试的概念和工具推荐，以及实用的调试技巧。",
            RSS_MAX_DESCRIPTION_LEN - 1);
    strncpy(items[2].author, "王五 <wangwu@example.com>", RSS_MAX_AUTHOR_LEN - 1);
    strncpy(items[2].categories[0], "测试", RSS_MAX_CATEGORY_LEN - 1);
    strncpy(items[2].categories[1], "调试", RSS_MAX_CATEGORY_LEN - 1);
    items[2].category_count = 2;
    items[2].pub_date = create_time(2026, 5, 1, 9, 0, 0);
    items[2].is_html_description = 0;
    
    strncpy(items[3].title, "数据结构与算法：常见面试题解析", RSS_MAX_TITLE_LEN - 1);
    strncpy(items[3].link, "https://example.com/blog/dsa-interview", RSS_MAX_LINK_LEN - 1);
    strncpy(items[3].description, 
            "准备技术面试？本文解析了常见的数据结构和算法面试题，包括数组、链表、树、图的遍历，以及排序和搜索算法的优化。",
            RSS_MAX_DESCRIPTION_LEN - 1);
    strncpy(items[3].categories[0], "数据结构", RSS_MAX_CATEGORY_LEN - 1);
    strncpy(items[3].categories[1], "算法", RSS_MAX_CATEGORY_LEN - 1);
    strncpy(items[3].categories[2], "面试", RSS_MAX_CATEGORY_LEN - 1);
    items[3].category_count = 3;
    items[3].pub_date = create_time(2026, 4, 28, 14, 0, 0);
    items[3].is_html_description = 0;
    
    strncpy(items[4].title, "Docker 容器化最佳实践", RSS_MAX_TITLE_LEN - 1);
    strncpy(items[4].link, "https://example.com/blog/docker-best-practices", RSS_MAX_LINK_LEN - 1);
    strncpy(items[4].description, 
            "<p>Docker 已成为现代应用部署的标准。本文分享了 Dockerfile 编写的最佳实践，包括：</p>\n<ol>\n<li>多阶段构建优化镜像大小</li>\n<li>合理使用层缓存</li>\n<li>安全配置建议</li>\n<li>资源限制设置</li>\n</ol>",
            RSS_MAX_DESCRIPTION_LEN - 1);
    strncpy(items[4].author, "赵六 <zhaoliu@example.com>", RSS_MAX_AUTHOR_LEN - 1);
    strncpy(items[4].categories[0], "Docker", RSS_MAX_CATEGORY_LEN - 1);
    strncpy(items[4].categories[1], "DevOps", RSS_MAX_CATEGORY_LEN - 1);
    items[4].category_count = 2;
    items[4].pub_date = create_time(2026, 4, 25, 11, 30, 0);
    items[4].is_html_description = 1;
    
    const char *output_file = "feed.xml";
    if (argc > 1) {
        output_file = argv[1];
    }
    
    printf("1. 生成 RSS Feed 到文件: %s\n", output_file);
    int result = RSS_GenerateToFile(output_file, &channel, items, 5, 20);
    if (result == 0) {
        printf("   ✓ 文件生成成功\n\n");
    } else {
        printf("   ✗ 文件生成失败\n\n");
        return 1;
    }
    
    printf("2. 演示流式输出方式:\n");
    RSS_Writer *writer = RSS_Writer_ToFile("stream_feed.xml", 3);
    if (writer != NULL) {
        RSS_StartChannel(writer, &channel);
        
        for (int i = 0; i < 5; i++) {
            printf("   添加文章: %s\n", items[i].title);
            RSS_AddItem(writer, &items[i]);
        }
        
        RSS_EndChannel(writer);
        RSS_Writer_Free(writer);
        printf("   ✓ 流式输出完成（限制 3 篇）\n\n");
    }
    
    printf("3. 演示输出到缓冲区:\n");
    char buffer[4096];
    int buf_len = RSS_GenerateToBuffer(buffer, sizeof(buffer), &channel, items, 2, 10);
    if (buf_len > 0) {
        printf("   ✓ 缓冲区输出成功，大小: %d 字节\n", buf_len);
        printf("   前 200 字符:\n   ");
        for (int i = 0; i < 200 && buffer[i] != '\0'; i++) {
            if (buffer[i] == '\n') {
                printf("\n   ");
            } else {
                putchar(buffer[i]);
            }
        }
        printf("...\n\n");
    }
    
    printf("4. 演示特殊字符转义:\n");
    const char *test_str = "标题包含 < 符号 & \"引用\"";
    char escaped[256];
    XML_Escape(escaped, sizeof(escaped), test_str);
    printf("   原始: %s\n", test_str);
    printf("   转义后: %s\n\n", escaped);
    
    printf("5. 演示时间格式转换:\n");
    time_t now = time(NULL);
    char time_buf[RFC822_DATE_LEN];
    Time_ToRFC822(time_buf, sizeof(time_buf), now);
    printf("   当前时间 (RFC 822): %s\n\n", time_buf);
    
    printf("=== 演示完成 ===\n");
    printf("生成的文件:\n");
    printf("  - feed.xml (完整 RSS Feed，5 篇文章)\n");
    printf("  - stream_feed.xml (流式输出，限制 3 篇文章)\n");
    
    printf("\n查看 feed.xml 内容:\n");
    FILE *f = fopen("feed.xml", "r");
    if (f != NULL) {
        char line[1024];
        int line_count = 0;
        while (fgets(line, sizeof(line), f) != NULL && line_count < 50) {
            printf("%s", line);
            line_count++;
        }
        fclose(f);
        if (line_count >= 50) {
            printf("... (内容过长，已截断)\n");
        }
    }
    
    return 0;
}
