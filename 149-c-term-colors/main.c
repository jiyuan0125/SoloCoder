#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <math.h>

#include "ansi_color.h"
#include "terminal.h"
#include "logger.h"

static void demo_basic_log_levels(void)
{
    printf("\n=== 基本日志级别演示 ===\n\n");

    LOG_DEBUG("这是一条 DEBUG 级别的消息 - 用于详细调试信息");
    LOG_INFO("这是一条 INFO 级别的消息 - 用于常规程序运行信息");
    LOG_WARN("这是一条 WARN 级别的消息 - 用于警告信息");
    LOG_ERROR("这是一条 ERROR 级别的消息 - 用于错误信息");
    LOG_FATAL("这是一条 FATAL 级别的消息 - 用于致命错误(红色加粗闪烁)");
}

static void demo_message_highlight(void)
{
    printf("\n=== 消息内容特殊标记高亮演示 ===\n\n");

    LOG_INFO("处理请求时发现 [ERROR] 连接超时");
    LOG_INFO("配置文件加载成功，但 [WARN] 某些参数使用默认值");
    LOG_INFO("初始化完成，[DEBUG] 详细信息已写入日志");
    LOG_INFO("系统状态：[INFO] 所有服务正常运行");

    LOG_ERROR("异常发生在 src/main.c:128 位置");
    LOG_WARN("注意：检查 config.json:45 的配置项");
    LOG_INFO("堆栈跟踪: handler.c:234 -> processor.c:567");
}

static void demo_ansi_16_colors(void)
{
    printf("\n=== 16 色(8色+亮色)演示 ===\n\n");

    const char* colors[] = {
        ANSI_FG_BLACK,   ANSI_FG_RED,     ANSI_FG_GREEN,   ANSI_FG_YELLOW,
        ANSI_FG_BLUE,    ANSI_FG_MAGENTA, ANSI_FG_CYAN,    ANSI_FG_WHITE,
        "\033[90m", "\033[91m", "\033[92m", "\033[93m",
        "\033[94m", "\033[95m", "\033[96m", "\033[97m"
    };
    const char* names[] = {
        "BLACK",   "RED",     "GREEN",   "YELLOW",
        "BLUE",    "MAGENTA", "CYAN",    "WHITE",
        "GRAY(br BLACK)", "BR RED", "BR GREEN", "BR YELLOW",
        "BR BLUE", "BR MAGENTA", "BR CYAN", "BR WHITE"
    };

    printf("前景色:\n");
    for (int i = 0; i < 16; i++) {
        printf("  %s%s%-16s%s", colors[i], (i < 8 ? "  " : ""), names[i], ANSI_RESET);
        if ((i + 1) % 4 == 0) printf("\n");
    }

    printf("\n背景色:\n");
    const char* bg_colors[] = {
        ANSI_BG_BLACK,   ANSI_BG_RED,     ANSI_BG_GREEN,   ANSI_BG_YELLOW,
        ANSI_BG_BLUE,    ANSI_BG_MAGENTA, ANSI_BG_CYAN,    ANSI_BG_WHITE
    };
    for (int i = 0; i < 8; i++) {
        printf("  %s %s BG %s", bg_colors[i], names[i], ANSI_RESET);
    }
    printf("\n");
}

static void demo_256_colors(void)
{
    printf("\n=== 256 色演示 ===\n\n");

    printf("标准 16 色:\n");
    for (int i = 0; i < 16; i++) {
        char code[32];
        ansi_format_fg_256(code, sizeof(code), (uint8_t)i);
        printf("%s%3d%s ", code, i, ANSI_RESET);
        if ((i + 1) % 8 == 0) printf("\n");
    }

    printf("\n216 颜色立方体 (6x6x6):\n");
    for (int r = 0; r < 6; r++) {
        for (int g = 0; g < 6; g++) {
            for (int b = 0; b < 6; b++) {
                int idx = 16 + r * 36 + g * 6 + b;
                char code[32];
                ansi_format_bg_256(code, sizeof(code), (uint8_t)idx);
                printf("%s   %s", code, ANSI_RESET);
            }
            printf(" ");
        }
        printf("\n");
    }

    printf("\n灰度色阶 (24色):\n");
    for (int i = 232; i < 256; i++) {
        char code[32];
        ansi_format_bg_256(code, sizeof(code), (uint8_t)i);
        printf("%s  %s", code, ANSI_RESET);
        if ((i - 232 + 1) % 12 == 0) printf("\n");
    }
    printf("\n");
}

static void demo_truecolor_rgb(void)
{
    printf("\n=== 24位真彩色(RGB)演示 ===\n\n");

    printf("红色渐变:\n");
    for (int i = 0; i < 32; i++) {
        int r = 128 + i * 4;
        char code[64];
        ansi_format_bg_rgb(code, sizeof(code), (uint8_t)r, 0, 0);
        printf("%s  %s", code, ANSI_RESET);
    }
    printf("\n");

    printf("绿色渐变:\n");
    for (int i = 0; i < 32; i++) {
        int g = 128 + i * 4;
        char code[64];
        ansi_format_bg_rgb(code, sizeof(code), 0, (uint8_t)g, 0);
        printf("%s  %s", code, ANSI_RESET);
    }
    printf("\n");

    printf("蓝色渐变:\n");
    for (int i = 0; i < 32; i++) {
        int b = 128 + i * 4;
        char code[64];
        ansi_format_bg_rgb(code, sizeof(code), 0, 0, (uint8_t)b);
        printf("%s  %s", code, ANSI_RESET);
    }
    printf("\n");

    printf("\n彩虹色条:\n");
    for (int i = 0; i < 64; i++) {
        int h = i * 360 / 64;
        float s = 1.0f, v = 1.0f;
        float c = v * s;
        float x = c * (1 - fabs(fmod(h / 60.0f, 2) - 1));
        float m = v - c;
        float r, g, b;

        if (h < 60) { r = c; g = x; b = 0; }
        else if (h < 120) { r = x; g = c; b = 0; }
        else if (h < 180) { r = 0; g = c; b = x; }
        else if (h < 240) { r = 0; g = x; b = c; }
        else if (h < 300) { r = x; g = 0; b = c; }
        else { r = c; g = 0; b = x; }

        uint8_t ri = (uint8_t)((r + m) * 255);
        uint8_t gi = (uint8_t)((g + m) * 255);
        uint8_t bi = (uint8_t)((b + m) * 255);

        char code[64];
        ansi_format_bg_rgb(code, sizeof(code), ri, gi, bi);
        printf("%s  %s", code, ANSI_RESET);
    }
    printf("\n");
}

static void demo_text_attributes(void)
{
    printf("\n=== 文本属性演示 ===\n\n");

    printf("  %s普通文本%s\n", ANSI_RESET, ANSI_RESET);
    printf("  %s粗体文本%s (BOLD)\n", ANSI_ATTR_STR_BOLD, ANSI_RESET);
    printf("  %s暗色文本%s (DIM)\n", ANSI_ATTR_STR_DIM, ANSI_RESET);
    printf("  %s斜体文本%s (ITALIC)\n", ANSI_ATTR_STR_ITALIC, ANSI_RESET);
    printf("  %s下划线文本%s (UNDERLINE)\n", ANSI_ATTR_STR_UNDERLINE, ANSI_RESET);
    printf("  %s闪烁文本%s (BLINK)\n", ANSI_ATTR_STR_BLINK, ANSI_RESET);
    printf("  %s反色文本%s (REVERSE)\n", "\033[7m", ANSI_RESET);

    printf("\n组合效果:\n");
    printf("  %s%s红色粗体%s\n", ANSI_FG_RED, ANSI_ATTR_STR_BOLD, ANSI_RESET);
    printf("  %s%s%s黄色下划线闪烁%s\n",
           ANSI_FG_YELLOW, ANSI_ATTR_STR_UNDERLINE, ANSI_ATTR_STR_BLINK, ANSI_RESET);
    printf("  %s%s%s绿色背景蓝色粗体%s\n",
           ANSI_BG_GREEN, ANSI_FG_BLUE, ANSI_ATTR_STR_BOLD, ANSI_RESET);
}

static void demo_terminal_detection(void)
{
    printf("\n=== 终端检测信息 ===\n\n");

    terminal_info_t info = terminal_detect(STDOUT_FILENO);

    printf("  stdout isatty:     %s\n", info.is_tty ? "是 (终端)" : "否 (管道/文件)");
    printf("  颜色能力:          ");
    switch (info.colors) {
        case TERM_COLORS_UNKNOWN: printf("未知"); break;
        case TERM_COLORS_8:       printf("8色"); break;
        case TERM_COLORS_16:      printf("16色"); break;
        case TERM_COLORS_256:     printf("256色"); break;
        case TERM_COLORS_TRUE:    printf("24位真彩色"); break;
        default:                   printf("%d色", (int)info.colors); break;
    }
    printf("\n");

    printf("  支持真彩色:        %s\n", info.supports_truecolor ? "是" : "否");
    printf("  支持256色:         %s\n", info.supports_256color ? "是" : "否");
    printf("  当前颜色模式:      ");
    switch (ansi_get_color_mode()) {
        case ANSI_COLOR_MODE_16:        printf("16色"); break;
        case ANSI_COLOR_MODE_256:       printf("256色"); break;
        case ANSI_COLOR_MODE_TRUECOLOR: printf("真彩色"); break;
    }
    printf("\n");

    printf("  环境变量 NO_COLOR: %s\n", terminal_check_no_color() ? "已设置" : "未设置");
    printf("  环境变量 FORCE_COLOR: %s\n", terminal_check_force_color() ? "已设置" : "未设置");
    printf("  当前输出模式:      ");
    switch (terminal_get_mode()) {
        case TERM_OUTPUT_AUTO:         printf("自动检测"); break;
        case TERM_OUTPUT_FORCE_COLOR:  printf("强制彩色"); break;
        case TERM_OUTPUT_FORCE_NO_COLOR: printf("强制无彩色"); break;
        case TERM_OUTPUT_STRIP:        printf("去除ANSI代码"); break;
    }
    printf("\n");

    const char* term = getenv("TERM");
    printf("  TERM 环境变量:     %s\n", term ? term : "(未设置)");

    const char* colorterm = getenv("COLORTERM");
    printf("  COLORTERM 环境变量: %s\n", colorterm ? colorterm : "(未设置)");
}

static void demo_strip_ansi_codes(void)
{
    printf("\n=== ANSI 代码去除演示 ===\n\n");

    const char* colored = "\033[31m\033[1m错误:\033[0m \033[33m连接超时\033[0m";
    printf("  原始带颜色字符串: %s\n", colored);

    char stripped[256];
    size_t len = terminal_strip_ansi_codes(stripped, sizeof(stripped),
                                             colored, strlen(colored));
    printf("  去除ANSI代码后:   %s (%zu字符)\n", stripped, len);

    printf("\n  测试: 程序输出到文件时会自动去除颜色代码\n");
    printf("  试试运行: ./color_log_demo > output.txt\n");
    printf("  然后查看: cat output.txt\n");
}

static void* thread_log_func(void* arg)
{
    (void)arg;

    for (int i = 0; i < 3; i++) {
        LOG_INFO("来自线程 %lu 的消息 #%d", (unsigned long)pthread_self(), i + 1);
        usleep(10000);
    }

    return NULL;
}

static void demo_multi_thread(void)
{
    printf("\n=== 多线程日志演示 ===\n\n");

    pthread_t threads[3];

    for (int i = 0; i < 3; i++) {
        pthread_create(&threads[i], NULL, thread_log_func, NULL);
    }

    for (int i = 0; i < 3; i++) {
        pthread_join(threads[i], NULL);
    }

    printf("\n  每个线程使用独立的线程局部缓冲区，无锁竞争\n");
}

static void demo_log_level_filter(void)
{
    printf("\n=== 日志级别过滤演示 ===\n\n");

    logger_config_t* cfg = log_get_default_config();

    printf("当前最低日志级别: DEBUG\n");
    printf("--------------------------------\n");
    LOG_DEBUG("这条会显示");
    LOG_INFO("这条会显示");
    LOG_WARN("这条会显示");
    LOG_ERROR("这条会显示");

    logger_config_t new_cfg = *cfg;
    new_cfg.min_level = LOG_LEVEL_WARN;
    log_set_default_config(&new_cfg);

    printf("\n当前最低日志级别: WARN\n");
    printf("--------------------------------\n");
    LOG_DEBUG("这条不会显示");
    LOG_INFO("这条不会显示");
    LOG_WARN("这条会显示");
    LOG_ERROR("这条会显示");

    log_set_default_config(cfg);
}

static void print_usage(void)
{
    printf("\n使用方法:\n");
    printf("  ./color_log_demo [选项]\n\n");
    printf("选项:\n");
    printf("  --help, -h     显示此帮助\n");
    printf("  --all          运行所有演示\n");
    printf("  --levels       基本日志级别演示\n");
    printf("  --highlight    消息高亮演示\n");
    printf("  --16color      16色演示\n");
    printf("  --256color     256色演示\n");
    printf("  --truecolor    真彩色演示\n");
    printf("  --attrs        文本属性演示\n");
    printf("  --term         终端检测\n");
    printf("  --strip        ANSI去除演示\n");
    printf("  --thread       多线程演示\n");
    printf("  --filter       级别过滤演示\n\n");
    printf("环境变量:\n");
    printf("  NO_COLOR=1     禁用颜色\n");
    printf("  FORCE_COLOR=1  强制启用颜色\n\n");
}

int main(int argc, char* argv[])
{
    printf("\n");
    printf("========================================\n");
    printf("    终端彩色日志模块演示程序\n");
    printf("========================================\n");

    if (argc > 1) {
        if (strcmp(argv[1], "--help") == 0 || strcmp(argv[1], "-h") == 0) {
            print_usage();
            return 0;
        }

        int run_all = 0;

        for (int i = 1; i < argc; i++) {
            if (strcmp(argv[i], "--all") == 0) {
                run_all = 1;
            }
        }

        for (int i = 1; i < argc; i++) {
            const char* arg = argv[i];
            if (strcmp(arg, "--all") == 0) continue;

            if (run_all || strcmp(arg, "--levels") == 0)
                demo_basic_log_levels();
            if (run_all || strcmp(arg, "--highlight") == 0)
                demo_message_highlight();
            if (run_all || strcmp(arg, "--16color") == 0)
                demo_ansi_16_colors();
            if (run_all || strcmp(arg, "--256color") == 0)
                demo_256_colors();
            if (run_all || strcmp(arg, "--truecolor") == 0)
                demo_truecolor_rgb();
            if (run_all || strcmp(arg, "--attrs") == 0)
                demo_text_attributes();
            if (run_all || strcmp(arg, "--term") == 0)
                demo_terminal_detection();
            if (run_all || strcmp(arg, "--strip") == 0)
                demo_strip_ansi_codes();
            if (run_all || strcmp(arg, "--thread") == 0)
                demo_multi_thread();
            if (run_all || strcmp(arg, "--filter") == 0)
                demo_log_level_filter();

            if (!run_all) break;
        }

        printf("\n");
        return 0;
    }

    demo_basic_log_levels();
    demo_message_highlight();
    demo_ansi_16_colors();
    demo_256_colors();
    demo_truecolor_rgb();
    demo_text_attributes();
    demo_terminal_detection();
    demo_strip_ansi_codes();
    demo_multi_thread();
    demo_log_level_filter();

    printf("\n========================================\n");
    printf("    演示结束。使用 --help 查看更多选项\n");
    printf("========================================\n\n");

    return 0;
}
