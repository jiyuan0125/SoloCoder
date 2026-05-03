#include "priority_queue.h"
#include <string.h>
#include <stdio.h>

static void heap_swap(HeapNode* a, HeapNode* b) {
    HeapNode temp = *a;
    *a = *b;
    *b = temp;
}

static int heap_compare(const HeapNode* a, const HeapNode* b) {
    if (a->priority != b->priority) {
        return a->priority - b->priority;
    }
    if (a->queue_time < b->queue_time) {
        return -1;
    } else if (a->queue_time > b->queue_time) {
        return 1;
    }
    return 0;
}

static void heapify_up(PriorityQueue* pq, int index) {
    while (index > 0) {
        int parent = (index - 1) / 2;
        if (heap_compare(&pq->heap[index], &pq->heap[parent]) < 0) {
            heap_swap(&pq->heap[index], &pq->heap[parent]);
            index = parent;
        } else {
            break;
        }
    }
}

static void heapify_down(PriorityQueue* pq, int index) {
    int size = pq->size;
    while (1) {
        int left = 2 * index + 1;
        int right = 2 * index + 2;
        int smallest = index;

        if (left < size && heap_compare(&pq->heap[left], &pq->heap[smallest]) < 0) {
            smallest = left;
        }
        if (right < size && heap_compare(&pq->heap[right], &pq->heap[smallest]) < 0) {
            smallest = right;
        }

        if (smallest != index) {
            heap_swap(&pq->heap[index], &pq->heap[smallest]);
            index = smallest;
        } else {
            break;
        }
    }
}

static void cleanup_invalid_nodes(PriorityQueue* pq) {
    while (pq->size > 0 && !pq->heap[0].is_valid) {
        pq->heap[0] = pq->heap[pq->size - 1];
        pq->size--;
        heapify_down(pq, 0);
    }
}

void pq_init(PriorityQueue* pq) {
    memset(pq, 0, sizeof(PriorityQueue));
    pq->next_patient_id = 1;
}

int pq_add_patient(PriorityQueue* pq, int priority, time_t arrival_time) {
    if (pq->patient_count >= MAX_PATIENTS) {
        return -1;
    }
    if (pq->size >= HEAP_CAPACITY) {
        return -1;
    }

    int patient_id = pq->next_patient_id++;
    int idx = patient_id - 1;

    pq->patients[idx].id = patient_id;
    pq->patients[idx].priority = priority;
    pq->patients[idx].arrival_time = arrival_time;
    pq->patients[idx].queue_time = arrival_time;
    pq->patients[idx].is_active = 1;

    HeapNode node;
    node.patient_id = patient_id;
    node.priority = priority;
    node.queue_time = arrival_time;
    node.is_valid = 1;

    pq->heap[pq->size] = node;
    heapify_up(pq, pq->size);
    pq->size++;

    pq->patient_count++;

    return patient_id;
}

int pq_extract_next(PriorityQueue* pq, Patient* out_patient) {
    cleanup_invalid_nodes(pq);

    if (pq->size == 0) {
        return -1;
    }

    HeapNode root = pq->heap[0];
    int patient_id = root.patient_id;
    int idx = patient_id - 1;

    if (out_patient != NULL) {
        *out_patient = pq->patients[idx];
    }

    pq->heap[0] = pq->heap[pq->size - 1];
    pq->size--;
    heapify_down(pq, 0);

    pq->patients[idx].is_active = 0;

    return patient_id;
}

int pq_peek_next(const PriorityQueue* pq, Patient* out_patient) {
    PriorityQueue temp = *pq;
    cleanup_invalid_nodes(&temp);

    if (temp.size == 0) {
        return -1;
    }

    HeapNode root = temp.heap[0];
    int patient_id = root.patient_id;
    int idx = patient_id - 1;

    if (out_patient != NULL) {
        *out_patient = temp.patients[idx];
    }

    return patient_id;
}

int pq_change_priority(PriorityQueue* pq, int patient_id, int new_priority) {
    if (patient_id < 1 || patient_id >= pq->next_patient_id) {
        return -1;
    }

    int idx = patient_id - 1;
    Patient* patient = &pq->patients[idx];

    if (!patient->is_active) {
        return -1;
    }

    int old_priority = patient->priority;
    if (old_priority == new_priority) {
        return 0;
    }

    for (int i = 0; i < pq->size; i++) {
        if (pq->heap[i].patient_id == patient_id && pq->heap[i].is_valid) {
            pq->heap[i].is_valid = 0;
            break;
        }
    }

    patient->priority = new_priority;
    time_t now = time(NULL);
    patient->queue_time = now;

    HeapNode new_node;
    new_node.patient_id = patient_id;
    new_node.priority = new_priority;
    new_node.queue_time = now;
    new_node.is_valid = 1;

    if (pq->size >= HEAP_CAPACITY) {
        return -1;
    }

    pq->heap[pq->size] = new_node;
    heapify_up(pq, pq->size);
    pq->size++;

    return 0;
}

int pq_remove_patient(PriorityQueue* pq, int patient_id) {
    if (patient_id < 1 || patient_id >= pq->next_patient_id) {
        return -1;
    }

    int idx = patient_id - 1;
    Patient* patient = &pq->patients[idx];

    if (!patient->is_active) {
        return -1;
    }

    for (int i = 0; i < pq->size; i++) {
        if (pq->heap[i].patient_id == patient_id && pq->heap[i].is_valid) {
            pq->heap[i].is_valid = 0;
            break;
        }
    }

    patient->is_active = 0;

    return 0;
}

int pq_is_empty(const PriorityQueue* pq) {
    PriorityQueue temp = *pq;
    cleanup_invalid_nodes(&temp);
    return temp.size == 0;
}

int pq_get_size(const PriorityQueue* pq) {
    int count = 0;
    for (int i = 1; i < pq->next_patient_id; i++) {
        if (pq->patients[i - 1].is_active) {
            count++;
        }
    }
    return count;
}

int pq_get_patient(const PriorityQueue* pq, int patient_id, Patient* out_patient) {
    if (patient_id < 1 || patient_id >= pq->next_patient_id) {
        return -1;
    }

    int idx = patient_id - 1;
    if (!pq->patients[idx].is_active) {
        return -1;
    }

    if (out_patient != NULL) {
        *out_patient = pq->patients[idx];
    }

    return 0;
}

int pq_get_all_patients(const PriorityQueue* pq, Patient* out_array, int max_count) {
    int count = 0;
    for (int i = 1; i < pq->next_patient_id && count < max_count; i++) {
        if (pq->patients[i - 1].is_active) {
            out_array[count++] = pq->patients[i - 1];
        }
    }
    return count;
}

int pq_get_priority_count(const PriorityQueue* pq, int priority) {
    int count = 0;
    for (int i = 1; i < pq->next_patient_id; i++) {
        if (pq->patients[i - 1].is_active && pq->patients[i - 1].priority == priority) {
            count++;
        }
    }
    return count;
}
