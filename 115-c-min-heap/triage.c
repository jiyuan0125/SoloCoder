#include "triage.h"
#include <stdio.h>
#include <string.h>

static char g_callback_buffer[512];

void triage_init(TriageSystem* ts, TriageCallback callback) {
    memset(ts, 0, sizeof(TriageSystem));
    pq_init(&ts->pq);
    ts->callback = callback;
    ts->start_time = time(NULL);
}

int triage_add_patient(TriageSystem* ts, int priority, time_t arrival_time) {
    if (priority < 1 || priority > 5) {
        return -1;
    }

    int patient_id = pq_add_patient(&ts->pq, priority, arrival_time);
    if (patient_id > 0) {
        ts->stats.total_patients++;
        if (ts->callback != NULL) {
            snprintf(g_callback_buffer, sizeof(g_callback_buffer),
                "[分诊] 病人 #%d 到达，优先级：%d 级(%s)，到达时间：%ld",
                patient_id, priority, get_priority_description(priority),
                (long)arrival_time);
            ts->callback(g_callback_buffer);
        }
    }
    return patient_id;
}

int triage_call_next(TriageSystem* ts, Patient* out_patient, time_t current_time) {
    Patient patient;
    int patient_id = pq_extract_next(&ts->pq, &patient);

    if (patient_id <= 0) {
        return -1;
    }

    double wait_time = difftime(current_time, patient.arrival_time);

    int pri_idx = patient.priority - 1;
    ts->stats.priority_stats[pri_idx].total_wait_time += wait_time;
    ts->stats.priority_stats[pri_idx].processed_count++;
    if (wait_time > ts->stats.priority_stats[pri_idx].max_wait_time) {
        ts->stats.priority_stats[pri_idx].max_wait_time = wait_time;
    }
    ts->stats.total_processed++;

    if (out_patient != NULL) {
        *out_patient = patient;
    }

    if (ts->callback != NULL) {
        snprintf(g_callback_buffer, sizeof(g_callback_buffer),
            "[叫号] 呼叫病人 #%d，优先级：%d 级(%s)，等待时间：%.0f 分钟",
            patient_id, patient.priority, get_priority_description(patient.priority),
            wait_time / 60.0);
        ts->callback(g_callback_buffer);
    }

    return patient_id;
}

int triage_check_priority_upgrade(TriageSystem* ts, time_t current_time) {
    Patient patients[MAX_PATIENTS];
    int count = pq_get_all_patients(&ts->pq, patients, MAX_PATIENTS);
    int upgraded = 0;

    for (int i = 0; i < count; i++) {
        Patient p = patients[i];
        double wait_minutes = difftime(current_time, p.arrival_time) / 60.0;
        int new_priority = p.priority;

        if (p.priority == 5 && wait_minutes >= 60) {
            new_priority = 4;
        } else if (p.priority == 4 && wait_minutes >= 40) {
            new_priority = 3;
        } else if (p.priority == 3 && wait_minutes >= 20) {
            new_priority = 2;
        } else if (p.priority == 2 && wait_minutes >= 5) {
            new_priority = 1;
        }

        if (new_priority != p.priority) {
            int result = pq_change_priority(&ts->pq, p.id, new_priority, current_time);
            if (result == 0) {
                upgraded++;
                if (ts->callback != NULL) {
                    snprintf(g_callback_buffer, sizeof(g_callback_buffer),
                        "[升级] 病人 #%d 等待 %.0f 分钟，从 %d 级升级到 %d 级",
                        p.id, wait_minutes, p.priority, new_priority);
                    ts->callback(g_callback_buffer);
                }
            }
        }
    }

    return upgraded;
}

int triage_check_timeouts(TriageSystem* ts, time_t current_time) {
    Patient patients[MAX_PATIENTS];
    int count = pq_get_all_patients(&ts->pq, patients, MAX_PATIENTS);
    int timeouts = 0;

    for (int i = 0; i < count; i++) {
        Patient p = patients[i];
        int limit = get_priority_time_limit(p.priority);

        if (limit == 0) continue;

        double wait_minutes = difftime(current_time, p.arrival_time) / 60.0;

        if (wait_minutes > limit) {
            timeouts++;
            if (ts->callback != NULL) {
                double overdue = wait_minutes - limit;
                snprintf(g_callback_buffer, sizeof(g_callback_buffer),
                    "[警告] 超时预警！病人 #%d，当前优先级：%d 级(%s)，已等待 %.1f 分钟，超时 %.1f 分钟",
                    p.id, p.priority, get_priority_description(p.priority),
                    wait_minutes, overdue);
                ts->callback(g_callback_buffer);
            }
        }
    }

    return timeouts;
}

void triage_get_stats(const TriageSystem* ts, SystemStats* out_stats) {
    if (out_stats != NULL) {
        *out_stats = ts->stats;
    }
}

int triage_get_waiting_count(const TriageSystem* ts) {
    return pq_get_size(&ts->pq);
}

int triage_get_priority_count(const TriageSystem* ts, int priority) {
    return pq_get_priority_count(&ts->pq, priority);
}
