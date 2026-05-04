#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

#include "result.h"
#include "common.h"

#define COLOR_RESET   "\033[0m"
#define COLOR_RED     "\033[31m"
#define COLOR_GREEN   "\033[32m"
#define COLOR_YELLOW  "\033[33m"
#define COLOR_CYAN    "\033[36m"
#define COLOR_BOLD    "\033[1m"

const char *status_to_string(cmd_status_t status) {
    switch (status) {
        case CMD_STATUS_PENDING:   return "PENDING";
        case CMD_STATUS_RUNNING:   return "RUNNING";
        case CMD_STATUS_SUCCESS:   return "SUCCESS";
        case CMD_STATUS_FAILED:    return "FAILED";
        case CMD_STATUS_TIMEOUT:   return "TIMEOUT";
        case CMD_STATUS_KILLED:    return "KILLED";
        default:                    return "UNKNOWN";
    }
}

const char *status_to_color(cmd_status_t status) {
    switch (status) {
        case CMD_STATUS_SUCCESS:   return COLOR_GREEN;
        case CMD_STATUS_FAILED:    return COLOR_RED;
        case CMD_STATUS_TIMEOUT:   return COLOR_YELLOW;
        case CMD_STATUS_KILLED:    return COLOR_RED;
        default:                    return COLOR_CYAN;
    }
}

void print_color(const char *text, const char *color) {
    printf("%s%s%s", color, text, COLOR_RESET);
}

int calculate_summary(cmd_result_t *results, size_t count, summary_stats_t *stats) {
    if (!results || !stats) return -1;
    
    memset(stats, 0, sizeof(summary_stats_t));
    stats->total = count;
    
    for (size_t i = 0; i < count; i++) {
        switch (results[i].status) {
            case CMD_STATUS_SUCCESS:
                stats->success++;
                break;
            case CMD_STATUS_FAILED:
                stats->failed++;
                break;
            case CMD_STATUS_TIMEOUT:
                stats->timeout++;
                break;
            case CMD_STATUS_KILLED:
                stats->killed++;
                break;
            default:
                break;
        }
    }
    
    return 0;
}

void print_summary(summary_stats_t *stats) {
    if (!stats) return;
    
    printf("\n");
    printf(COLOR_BOLD "========== 执行汇总 ==========\n" COLOR_RESET);
    printf("总命令数: %d\n", stats->total);
    printf(COLOR_GREEN "成功:     %d\n" COLOR_RESET, stats->success);
    printf(COLOR_RED "失败:     %d\n" COLOR_RESET, stats->failed);
    printf(COLOR_YELLOW "超时:     %d\n" COLOR_RESET, stats->timeout);
    printf(COLOR_RED "被杀掉:   %d\n" COLOR_RESET, stats->killed);
    printf("==============================\n\n");
}

char **split_lines(const char *buf, size_t len, size_t *line_count) {
    if (!buf || !line_count) return NULL;
    *line_count = 0;
    
    size_t count = 0;
    for (size_t i = 0; i < len; i++) {
        if (buf[i] == '\n') count++;
    }
    if (len > 0 && buf[len - 1] != '\n') count++;
    
    if (count == 0) return NULL;
    
    char **lines = malloc(count * sizeof(char *));
    if (!lines) return NULL;
    
    size_t line_idx = 0;
    const char *start = buf;
    
    for (size_t i = 0; i < len; i++) {
        if (buf[i] == '\n') {
            size_t line_len = &buf[i] - start;
            lines[line_idx] = malloc(line_len + 1);
            if (!lines[line_idx]) {
                free_lines(lines, line_idx);
                return NULL;
            }
            memcpy(lines[line_idx], start, line_len);
            lines[line_idx][line_len] = '\0';
            line_idx++;
            start = &buf[i + 1];
        }
    }
    
    if (start < buf + len) {
        size_t line_len = buf + len - start;
        lines[line_idx] = malloc(line_len + 1);
        if (!lines[line_idx]) {
            free_lines(lines, line_idx);
            return NULL;
        }
        memcpy(lines[line_idx], start, line_len);
        lines[line_idx][line_len] = '\0';
        line_idx++;
    }
    
    *line_count = line_idx;
    return lines;
}

void free_lines(char **lines, size_t count) {
    if (!lines) return;
    for (size_t i = 0; i < count; i++) {
        free(lines[i]);
    }
    free(lines);
}

static void print_output_summary(const char *label, const char *buf, size_t len, 
                                  bool truncated, int summary_lines) {
    if (len == 0) {
        printf("  %s: (无输出)\n", label);
        return;
    }
    
    size_t line_count;
    char **lines = split_lines(buf, len, &line_count);
    
    if (!lines || line_count == 0) {
        printf("  %s: [%zu bytes]\n", label, len);
        return;
    }
    
    printf("  %s:", label);
    if (truncated) {
        printf(COLOR_YELLOW " (已截断，已保存 %zu 字节)" COLOR_RESET, len);
    }
    printf("\n");
    
    if (line_count <= (size_t)(summary_lines * 2)) {
        for (size_t i = 0; i < line_count; i++) {
            printf("    %s\n", lines[i]);
        }
    } else {
        for (int i = 0; i < summary_lines; i++) {
            printf("    %s\n", lines[i]);
        }
        printf("    ... (省略 %zu 行) ...\n", line_count - (size_t)(summary_lines * 2));
        for (size_t i = line_count - summary_lines; i < line_count; i++) {
            printf("    %s\n", lines[i]);
        }
    }
    
    free_lines(lines, line_count);
}

void print_result_summary(cmd_result_t *result, int summary_lines) {
    if (!result || !result->config) return;
    
    const char *label = result->config->label ? result->config->label : "";
    const char *status_str = status_to_string(result->status);
    const char *status_color = status_to_color(result->status);
    
    printf(COLOR_BOLD "--- 命令 [%d] %s ---" COLOR_RESET "\n", 
           result->config->id, label);
    printf("  命令行: %s\n", result->config->command);
    printf("  状态: ");
    print_color(status_str, status_color);
    printf("\n");
    printf("  退出码: %d\n", result->exit_code);
    printf("  耗时: %.2f 秒\n", result->elapsed_sec);
    
    if (result->config->work_dir) {
        printf("  工作目录: %s\n", result->config->work_dir);
    }
    
    print_output_summary("标准输出", result->output.stdout_buf, 
                         result->output.stdout_len, 
                         result->output.stdout_truncated,
                         summary_lines);
    print_output_summary("标准错误", result->output.stderr_buf,
                         result->output.stderr_len,
                         result->output.stderr_truncated,
                         summary_lines);
    printf("\n");
}

void print_all_results(cmd_result_t *results, size_t count, int summary_lines) {
    if (!results) return;
    
    for (size_t i = 0; i < count; i++) {
        print_result_summary(&results[i], summary_lines);
    }
}

void print_complete_report(cmd_result_t *results, size_t count, int summary_lines) {
    if (!results || count == 0) {
        printf("无结果可显示\n");
        return;
    }
    
    printf(COLOR_BOLD COLOR_CYAN "\n"
           "========================================\n"
           "         命令批量执行报告\n"
           "========================================\n" COLOR_RESET);
    
    summary_stats_t stats;
    calculate_summary(results, count, &stats);
    
    printf("\n--- 失败/超时/被杀掉的命令 ---\n");
    bool has_errors = false;
    for (size_t i = 0; i < count; i++) {
        if (results[i].status != CMD_STATUS_SUCCESS) {
            has_errors = true;
            print_result_summary(&results[i], summary_lines);
        }
    }
    if (!has_errors) {
        printf(COLOR_GREEN "  所有命令执行成功！\n" COLOR_RESET);
    }
    
    printf("\n--- 所有命令摘要 ---\n");
    printf(COLOR_BOLD "ID    状态        退出码   耗时      标签\n" COLOR_RESET);
    for (size_t i = 0; i < count; i++) {
        cmd_result_t *r = &results[i];
        const char *status_str = status_to_string(r->status);
        const char *status_color = status_to_color(r->status);
        const char *label = r->config->label ? r->config->label : "";
        
        printf("%-4d  ", r->config->id);
        print_color(status_str, status_color);
        printf("%-*s", (int)(10 - strlen(status_str)), "");
        printf("%-6d   %-6.2f  %s\n", r->exit_code, r->elapsed_sec, label);
    }
    
    print_summary(&stats);
}
