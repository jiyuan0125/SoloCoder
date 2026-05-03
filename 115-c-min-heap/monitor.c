#include "monitor.h"
#include <stdio.h>
#include <string.h>
#include <stdarg.h>

void monitor_init(MonitorReport* report) {
    memset(report, 0, sizeof(MonitorReport));
}

static int report_append(MonitorReport* report, const char* format, ...) {
    va_list args;
    va_start(args, format);

    int remaining = MONITOR_BUFFER_SIZE - report->buffer_len - 1;
    if (remaining <= 0) {
        va_end(args);
        return -1;
    }

    int written = vsnprintf(
        report->buffer + report->buffer_len,
        remaining,
        format,
        args
    );

    va_end(args);

    if (written < 0 || written >= remaining) {
        return -1;
    }

    report->buffer_len += written;
    return 0;
}

int monitor_generate_stats_report(const TriageSystem* ts, MonitorReport* report, time_t current_time) {
    monitor_init(report);

    SystemStats stats;
    triage_get_stats(ts, &stats);

    report_append(report, "========== 系统统计报告 ==========\n");
    report_append(report, "当前时间: %ld\n", (long)current_time);
    report_append(report, "今日已处理病人数: %d\n", stats.total_processed);
    report_append(report, "今日总接诊数: %d\n", stats.total_patients);
    report_append(report, "\n");

    double overall_total = 0;
    double overall_max = 0;
    int overall_count = 0;

    for (int p = 1; p <= 5; p++) {
        int pri_idx = p - 1;
        PriorityStats ps = stats.priority_stats[pri_idx];
        int waiting = triage_get_priority_count(ts, p);

        double avg_wait = 0;
        if (ps.count > 0) {
            avg_wait = ps.total_wait_time / ps.count / 60.0;
        }

        report_append(report, "[%d级 - %s]\n", p, get_priority_description(p));
        report_append(report, "  时间限制: %d 分钟\n", get_priority_time_limit(p));
        report_append(report, "  当前排队: %d 人\n", waiting);
        report_append(report, "  已处理: %d 人\n", ps.count);
        report_append(report, "  平均等待: %.1f 分钟\n", avg_wait);
        report_append(report, "  最长等待: %.1f 分钟\n", ps.max_wait_time / 60.0);
        report_append(report, "\n");

        if (ps.count > 0) {
            overall_total += ps.total_wait_time;
            if (ps.max_wait_time > overall_max) {
                overall_max = ps.max_wait_time;
            }
            overall_count += ps.count;
        }
    }

    report_append(report, "---------- 整体统计 ----------\n");
    if (overall_count > 0) {
        report_append(report, "  整体平均等待: %.1f 分钟\n", overall_total / overall_count / 60.0);
    } else {
        report_append(report, "  整体平均等待: 0.0 分钟\n");
    }
    report_append(report, "  整体最长等待: %.1f 分钟\n", overall_max / 60.0);
    report_append(report, "=================================\n");

    return 0;
}

int monitor_generate_queue_report(const TriageSystem* ts, MonitorReport* report, time_t current_time) {
    monitor_init(report);

    Patient patients[MAX_PATIENTS];
    int count = pq_get_all_patients(&ts->pq, patients, MAX_PATIENTS);

    report_append(report, "========== 当前排队情况 ==========\n");
    report_append(report, "当前时间: %ld\n", (long)current_time);
    report_append(report, "排队总人数: %d\n\n", count);

    if (count == 0) {
        report_append(report, "  (暂无排队病人)\n");
    } else {
        for (int p = 1; p <= 5; p++) {
            int has_priority = 0;
            for (int i = 0; i < count; i++) {
                if (patients[i].priority == p) {
                    has_priority = 1;
                    break;
                }
            }
            if (!has_priority) continue;

            report_append(report, "[%d级 - %s] 排队:\n", p, get_priority_description(p));

            for (int i = 0; i < count; i++) {
                if (patients[i].priority == p) {
                    double wait_min = difftime(current_time, patients[i].arrival_time) / 60.0;
                    report_append(report,
                        "  病人 #%d: 到达时间 %ld, 已等待 %.1f 分钟, 排队时间 %ld\n",
                        patients[i].id,
                        (long)patients[i].arrival_time,
                        wait_min,
                        (long)patients[i].queue_time);
                }
            }
            report_append(report, "\n");
        }
    }

    report_append(report, "=================================\n");

    return 0;
}

int monitor_generate_timeout_report(const TriageSystem* ts, MonitorReport* report, time_t current_time) {
    monitor_init(report);

    Patient patients[MAX_PATIENTS];
    int count = pq_get_all_patients(&ts->pq, patients, MAX_PATIENTS);

    report_append(report, "========== 超时预警报告 ==========\n");
    report_append(report, "当前时间: %ld\n\n", (long)current_time);

    int has_timeout = 0;

    for (int i = 0; i < count; i++) {
        Patient p = patients[i];
        int limit = get_priority_time_limit(p.priority);

        if (limit == 0) continue;

        double wait_minutes = difftime(current_time, p.arrival_time) / 60.0;

        if (wait_minutes > limit) {
            has_timeout = 1;
            double overdue = wait_minutes - limit;
            report_append(report,
                "[警告] 病人 #%d 超时!\n"
                "  当前优先级: %d 级(%s)\n"
                "  时间限制: %d 分钟\n"
                "  已等待: %.1f 分钟\n"
                "  超时: %.1f 分钟\n\n",
                p.id, p.priority, get_priority_description(p.priority),
                limit, wait_minutes, overdue);
        }
    }

    if (!has_timeout) {
        report_append(report, "  (暂无超时病人)\n");
    }

    report_append(report, "=================================\n");

    return 0;
}

void monitor_print_report(const MonitorReport* report) {
    if (report->buffer_len > 0) {
        printf("%s", report->buffer);
    }
}
