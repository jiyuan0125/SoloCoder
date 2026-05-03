#ifndef TRIAGE_H
#define TRIAGE_H

#include "priority_queue.h"
#include "common.h"

typedef void (*TriageCallback)(const char* message);

typedef struct {
    PriorityQueue pq;
    SystemStats stats;
    TriageCallback callback;
    time_t start_time;
} TriageSystem;

void triage_init(TriageSystem* ts, TriageCallback callback);
int triage_add_patient(TriageSystem* ts, int priority, time_t arrival_time);
int triage_call_next(TriageSystem* ts, Patient* out_patient);
int triage_check_priority_upgrade(TriageSystem* ts, time_t current_time);
int triage_check_timeouts(TriageSystem* ts, time_t current_time);
void triage_update_stats(TriageSystem* ts, time_t current_time);
void triage_get_stats(const TriageSystem* ts, SystemStats* out_stats);
int triage_get_waiting_count(const TriageSystem* ts);
int triage_get_priority_count(const TriageSystem* ts, int priority);

#endif
