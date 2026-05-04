#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "timezone.h"
#include "time_formatter.h"
#include "time_parser.h"

static void print_usage(const char *prog_name) {
    printf("时间戳格式化工具 v1.0\n\n");
    printf("用法:\n");
    printf("  %s format <时间戳> [选项]      - 将时间戳格式化为可读格式\n", prog_name);
    printf("  %s parse <时间字符串> [选项]    - 解析时间字符串为时间戳\n", prog_name);
    printf("  %s demo                          - 运行演示\n", prog_name);
    printf("  %s help                          - 显示帮助信息\n\n", prog_name);
    printf("选项:\n");
    printf("  --timezone <时区名>              - 指定时区 (默认: UTC)\n");
    printf("  --format <格式类型>               - 输出格式类型 (format命令)\n");
    printf("                                    absolute   - 绝对时间 (默认)\n");
    printf("                                    relative   - 相对时间\n");
    printf("                                    date       - 纯日期\n");
    printf("                                    time       - 纯时间\n");
    printf("                                    friendly   - 友好时间\n\n");
    printf("支持的时区:\n");
    printf("  UTC, Asia/Shanghai (北京), Asia/Tokyo (东京)\n");
    printf("  America/New_York (纽约), Europe/London (伦敦)\n\n");
    printf("示例:\n");
    printf("  %s format 1777871080 --timezone Asia/Shanghai --format friendly\n", prog_name);
    printf("  %s parse \"2026-05-03 14:30:00\" --timezone Asia/Shanghai\n", prog_name);
    printf("  %s parse \"May 3 2026 14:30\" --timezone America/New_York\n", prog_name);
}

static TimeFormatType parse_format_type(const char *str) {
    if (strcmp(str, "absolute") == 0) return FORMAT_ABSOLUTE;
    if (strcmp(str, "relative") == 0) return FORMAT_RELATIVE;
    if (strcmp(str, "date") == 0) return FORMAT_DATE_ONLY;
    if (strcmp(str, "time") == 0) return FORMAT_TIME_ONLY;
    if (strcmp(str, "friendly") == 0) return FORMAT_FRIENDLY;
    return FORMAT_ABSOLUTE;
}

static int cmd_format(int argc, char *argv[]) {
    time_t timestamp = 0;
    const TimeZoneInfo *tz = &TIMEZONE_UTC;
    TimeFormatType format_type = FORMAT_ABSOLUTE;
    int has_timestamp = 0;
    
    for (int i = 2; i < argc; i++) {
        if (strcmp(argv[i], "--timezone") == 0 || strcmp(argv[i], "-z") == 0) {
            if (i + 1 < argc) {
                tz = timezone_get_by_name(argv[++i]);
            } else {
                fprintf(stderr, "错误: --timezone 参数需要时区名\n");
                return 1;
            }
        } else if (strcmp(argv[i], "--format") == 0 || strcmp(argv[i], "-f") == 0) {
            if (i + 1 < argc) {
                format_type = parse_format_type(argv[++i]);
            } else {
                fprintf(stderr, "错误: --format 参数需要格式类型\n");
                return 1;
            }
        } else if (!has_timestamp) {
            timestamp = (time_t)atoll(argv[i]);
            has_timestamp = 1;
        }
    }
    
    if (!has_timestamp) {
        fprintf(stderr, "错误: 请提供时间戳\n");
        return 1;
    }
    
    char buffer[TIME_FORMAT_MAX_LEN];
    time_t now = time(NULL);
    
    if (format_time(timestamp, now, tz, format_type, buffer, sizeof(buffer)) != 0) {
        fprintf(stderr, "错误: 格式化时间失败\n");
        return 1;
    }
    
    printf("%s\n", buffer);
    return 0;
}

static int cmd_parse(int argc, char *argv[]) {
    const char *input = NULL;
    const TimeZoneInfo *tz = &TIMEZONE_UTC;
    char combined_input[1024] = "";
    int input_start = -1;
    
    for (int i = 2; i < argc; i++) {
        if (strcmp(argv[i], "--timezone") == 0 || strcmp(argv[i], "-z") == 0) {
            if (i + 1 < argc) {
                tz = timezone_get_by_name(argv[++i]);
            } else {
                fprintf(stderr, "错误: --timezone 参数需要时区名\n");
                return 1;
            }
        } else if (input_start == -1) {
            input_start = i;
        }
    }
    
    if (input_start == -1) {
        fprintf(stderr, "错误: 请提供时间字符串\n");
        return 1;
    }
    
    for (int i = input_start; i < argc; i++) {
        if (strlen(combined_input) > 0) {
            strcat(combined_input, " ");
        }
        strcat(combined_input, argv[i]);
    }
    input = combined_input;
    
    time_t result;
    ParseError error;
    
    if (parse_time_string(input, tz, &result, &error) != 0) {
        fprintf(stderr, "解析失败: %s\n", error.message);
        if (error.position >= 0) {
            fprintf(stderr, "位置: %d\n", error.position);
        }
        return 1;
    }
    
    printf("%lld\n", (long long)result);
    return 0;
}

static int cmd_demo(void) {
    printf("=== 时间戳格式化工具演示 ===\n\n");
    
    time_t now = time(NULL);
    const TimeZoneInfo *tz = &TIMEZONE_BEIJING;
    
    printf("当前时间戳: %lld\n\n", (long long)now);
    
    printf("--- 各种格式展示 ---\n");
    char buffer[TIME_FORMAT_MAX_LEN];
    TimeFormatType formats[] = {FORMAT_ABSOLUTE, FORMAT_RELATIVE, FORMAT_DATE_ONLY, FORMAT_TIME_ONLY, FORMAT_FRIENDLY};
    const char *format_names[] = {"绝对时间", "相对时间", "纯日期", "纯时间", "友好时间"};
    
    for (int i = 0; i < 5; i++) {
        format_time(now, now, tz, formats[i], buffer, sizeof(buffer));
        printf("%s: %s\n", format_names[i], buffer);
    }
    printf("\n");
    
    printf("--- 相对时间示例 (时区: 北京) ---\n");
    time_t test_times[] = {
        now - 30,
        now - 300,
        now - 3600,
        now - 86400,
        now - 2 * 86400,
        now - 5 * 86400,
        now - 15 * 86400,
        now - 45 * 86400
    };
    const char *test_labels[] = {
        "30秒前",
        "5分钟前",
        "1小时前",
        "1天前",
        "2天前",
        "5天前",
        "15天前",
        "45天前"
    };
    
    for (int i = 0; i < 8; i++) {
        format_time(test_times[i], now, tz, FORMAT_RELATIVE, buffer, sizeof(buffer));
        printf("%s -> %s\n", test_labels[i], buffer);
        
        format_time(test_times[i], now, tz, FORMAT_FRIENDLY, buffer, sizeof(buffer));
        printf("  友好格式: %s\n", buffer);
    }
    printf("\n");
    
    printf("--- 时区转换示例 ---\n");
    time_t test_ts = 1777871080;
    const TimeZoneInfo *tzs[] = {&TIMEZONE_UTC, &TIMEZONE_BEIJING, &TIMEZONE_TOKYO, &TIMEZONE_NEWYORK, &TIMEZONE_LONDON};
    const char *tz_names[] = {"UTC", "北京 (UTC+8)", "东京 (UTC+9)", "纽约 (UTC-5/-4)", "伦敦 (UTC+0/+1)"};
    
    printf("时间戳: %lld\n\n", (long long)test_ts);
    for (int i = 0; i < 5; i++) {
        format_time(test_ts, now, tzs[i], FORMAT_ABSOLUTE, buffer, sizeof(buffer));
        printf("%s: %s\n", tz_names[i], buffer);
    }
    printf("\n");
    
    printf("--- 时间字符串解析示例 ---\n");
    const char *test_strings[] = {
        "2026-05-03 14:30:00",
        "2026/05/03 14:30",
        "May 3 2026 14:30",
        "May 3, 2026",
        "2026-02-29"
    };
    
    for (int i = 0; i < 5; i++) {
        time_t parsed;
        ParseError err;
        
        printf("解析 \"%s\" (UTC): ", test_strings[i]);
        if (parse_time_string(test_strings[i], &TIMEZONE_UTC, &parsed, &err) == 0) {
            printf("%lld", (long long)parsed);
            format_time(parsed, now, &TIMEZONE_UTC, FORMAT_ABSOLUTE, buffer, sizeof(buffer));
            printf(" -> %s\n", buffer);
        } else {
            printf("失败: %s\n", err.message);
        }
    }
    
    printf("\n=== 演示结束 ===\n");
    return 0;
}

int main(int argc, char *argv[]) {
    if (argc < 2) {
        print_usage(argv[0]);
        return 1;
    }
    
    const char *cmd = argv[1];
    
    if (strcmp(cmd, "help") == 0 || strcmp(cmd, "-h") == 0 || strcmp(cmd, "--help") == 0) {
        print_usage(argv[0]);
        return 0;
    }
    
    if (strcmp(cmd, "format") == 0) {
        return cmd_format(argc, argv);
    }
    
    if (strcmp(cmd, "parse") == 0) {
        return cmd_parse(argc, argv);
    }
    
    if (strcmp(cmd, "demo") == 0) {
        return cmd_demo();
    }
    
    fprintf(stderr, "未知命令: %s\n", cmd);
    fprintf(stderr, "使用 '%s help' 查看帮助\n", argv[0]);
    return 1;
}
