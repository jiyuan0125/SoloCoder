#ifndef PRIORITY_QUEUE_H
#define PRIORITY_QUEUE_H

#include "common.h"

#define HEAP_CAPACITY (MAX_PATIENTS * 2)

typedef struct {
    int patient_id;
    int priority;
    time_t queue_time;
    int is_valid;
} HeapNode;

typedef struct {
    HeapNode heap[HEAP_CAPACITY];
    int size;
    Patient patients[MAX_PATIENTS];
    int patient_count;
    int next_patient_id;
} PriorityQueue;

void pq_init(PriorityQueue* pq);
int pq_add_patient(PriorityQueue* pq, int priority, time_t arrival_time);
int pq_extract_next(PriorityQueue* pq, Patient* out_patient);
int pq_peek_next(const PriorityQueue* pq, Patient* out_patient);
int pq_change_priority(PriorityQueue* pq, int patient_id, int new_priority);
int pq_remove_patient(PriorityQueue* pq, int patient_id);
int pq_is_empty(const PriorityQueue* pq);
int pq_get_size(const PriorityQueue* pq);
int pq_get_patient(const PriorityQueue* pq, int patient_id, Patient* out_patient);
int pq_get_all_patients(const PriorityQueue* pq, Patient* out_array, int max_count);
int pq_get_priority_count(const PriorityQueue* pq, int priority);

#endif
