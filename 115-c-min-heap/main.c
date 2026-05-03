#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <time.h>
#include <stdarg.h>

#include "common.h"
#include "priority_queue.h"
#include "triage.h"
#include "monitor.h"

static time_t g_simulated_time;

static void log_message(const char* message) {
    printf("[模拟时间 %ld] %s\n", (long)g_simulated_time, message);
}

static void advance_simulated_time(int seconds) {
    g_simulated_time += seconds;
}

static void print_separator(void) {
    printf("------------------------------------------------------------\n");
}

typedef struct {
    int id;
    int new_priority;
    int trigger_time;
} PriorityChangeEvent;

typedef struct {
    int priority;
    int arrival_offset;
} PatientArrival;

int main(void) {
    printf("\n");
    printf("============================================\n");
    printf("    医院急诊科分诊调度系统模拟\n");
    printf("============================================\n");
    printf("\n");

    g_simulated_time = 1000000;

    TriageSystem ts;
    triage_init(&ts, log_message);

    MonitorReport report;
    monitor_init(&report);

    PatientArrival arrivals[] = {
        {5, 0},
        {4, 2},
        {3, 5},
        {2, 8},
        {1, 10},
        {5, 15},
        {4, 20},
        {3, 25},
        {4, 30},
        {5, 35},
        {3, 40},
        {2, 45},
    };
    int num_arrivals = sizeof(arrivals) / sizeof(arrivals[0]);

    PriorityChangeEvent changes[] = {
        {7, 2, 40},
        {2, 5, 50},
    };
    int num_changes = sizeof(changes) / sizeof(changes[0]);

    int next_arrival_idx = 0;
    int simulation_step = 0;
    int doctor_available = 1;

    printf("开始模拟...\n");
    printf("时间说明: 模拟时间单位为秒，但实际意义为分钟级模拟\n");
    printf("每步模拟推进约 30 秒 (0.5 分钟)\n");
    printf("\n");

    while (simulation_step < 200) {
        simulation_step++;

        while (next_arrival_idx < num_arrivals) {
            PatientArrival arr = arrivals[next_arrival_idx];
            time_t target_time = 1000000 + arr.arrival_offset * 60;

            if (g_simulated_time >= target_time) {
                triage_add_patient(&ts, arr.priority, target_time);
                next_arrival_idx++;
            } else {
                break;
            }
        }

        triage_check_priority_upgrade(&ts, g_simulated_time);

        triage_check_timeouts(&ts, g_simulated_time);

        if (doctor_available && !pq_is_empty(&ts.pq)) {
            Patient p;
            int called = triage_call_next(&ts, &p);
            if (called > 0) {
                doctor_available = 0;
            }
        }

        if (simulation_step % 10 == 0) {
            doctor_available = 1;
        }

        if (simulation_step % 5 == 3) {
            print_separator();
            monitor_generate_queue_report(&ts, &report, g_simulated_time);
            monitor_print_report(&report);
        }

        if (simulation_step % 20 == 0) {
            print_separator();
            monitor_generate_stats_report(&ts, &report, g_simulated_time);
            monitor_print_report(&report);

            monitor_generate_timeout_report(&ts, &report, g_simulated_time);
            monitor_print_report(&report);
        }

        if (simulation_step == 50) {
            printf("\n>>> 演示场景: 病人优先级动态调整 <<<\n");
            printf("将手动调整以下病人的优先级:\n");
            printf("  - 病人 #7: 从 4 级 升级到 2 级 (病情恶化)\n");
            printf("  - 病人 #2: 从 4 级 降级到 5 级 (病情好转)\n");
            printf("\n");

            for (int i = 0; i < num_changes; i++) {
                PriorityChangeEvent evt = changes[i];
                Patient p;
                if (pq_get_patient(&ts.pq, evt.id, &p) == 0) {
                    int old_pri = p.priority;
                    int result = pq_change_priority(&ts.pq, evt.id, evt.new_priority);
                    if (result == 0) {
                        if (evt.new_priority < old_pri) {
                            printf("[手动调整] 病人 #%d 从 %d 级 升级到 %d 级\n",
                                   evt.id, old_pri, evt.new_priority);
                        } else {
                            printf("[手动调整] 病人 #%d 从 %d 级 降级到 %d 级\n",
                                   evt.id, old_pri, evt.new_priority);
                        }
                    }
                } else {
                    printf("[手动调整] 病人 #%d 已不在队列中\n", evt.id);
                }
            }
            printf("\n");
        }

        advance_simulated_time(30);

        usleep(100000);
    }

    printf("\n");
    print_separator();
    printf("模拟结束 - 最终统计\n");
    print_separator();

    triage_update_stats(&ts, g_simulated_time);
    monitor_generate_stats_report(&ts, &report, g_simulated_time);
    monitor_print_report(&report);

    printf("\n============================================\n");
    printf("    模拟结束\n");
    printf("============================================\n");
    printf("\n");

    return 0;
}
