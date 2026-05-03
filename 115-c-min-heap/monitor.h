#ifndef MONITOR_H
#define MONITOR_H

#include "triage.h"

#define MONITOR_BUFFER_SIZE 2048

typedef struct {
    char buffer[MONITOR_BUFFER_SIZE];
    int buffer_len;
} MonitorReport;

void monitor_init(MonitorReport* report);
int monitor_generate_stats_report(const TriageSystem* ts, MonitorReport* report, time_t current_time);
int monitor_generate_queue_report(const TriageSystem* ts, MonitorReport* report, time_t current_time);
int monitor_generate_timeout_report(const TriageSystem* ts, MonitorReport* report, time_t current_time);
void monitor_print_report(const MonitorReport* report);

#endif
