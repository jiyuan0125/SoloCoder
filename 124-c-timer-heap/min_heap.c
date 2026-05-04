#include "min_heap.h"
#include <stdlib.h>
#include <string.h>

#define INITIAL_CAPACITY 16

static void swap(timer_task_t **a, timer_task_t **b) {
    timer_task_t *temp = *a;
    *a = *b;
    *b = temp;
}

static void heapify_up(min_heap_t *heap, size_t index) {
    while (index > 0) {
        size_t parent = (index - 1) / 2;
        if (timespec_compare(&heap->data[index]->expire_time, 
                            &heap->data[parent]->expire_time) >= 0) {
            break;
        }
        swap(&heap->data[index], &heap->data[parent]);
        index = parent;
    }
}

static void heapify_down(min_heap_t *heap, size_t index) {
    while (1) {
        size_t left = 2 * index + 1;
        size_t right = 2 * index + 2;
        size_t smallest = index;

        if (left < heap->size && 
            timespec_compare(&heap->data[left]->expire_time, 
                           &heap->data[smallest]->expire_time) < 0) {
            smallest = left;
        }
        if (right < heap->size && 
            timespec_compare(&heap->data[right]->expire_time, 
                           &heap->data[smallest]->expire_time) < 0) {
            smallest = right;
        }
        if (smallest == index) {
            break;
        }
        swap(&heap->data[index], &heap->data[smallest]);
        index = smallest;
    }
}

static int heap_resize(min_heap_t *heap, size_t new_capacity) {
    timer_task_t **new_data = (timer_task_t **)realloc(heap->data, 
                                                        new_capacity * sizeof(timer_task_t *));
    if (new_data == NULL) {
        return -1;
    }
    heap->data = new_data;
    heap->capacity = new_capacity;
    return 0;
}

min_heap_t* min_heap_create(size_t initial_capacity) {
    if (initial_capacity == 0) {
        initial_capacity = INITIAL_CAPACITY;
    }
    min_heap_t *heap = (min_heap_t *)malloc(sizeof(min_heap_t));
    if (heap == NULL) {
        return NULL;
    }
    heap->data = (timer_task_t **)malloc(initial_capacity * sizeof(timer_task_t *));
    if (heap->data == NULL) {
        free(heap);
        return NULL;
    }
    heap->capacity = initial_capacity;
    heap->size = 0;
    return heap;
}

void min_heap_destroy(min_heap_t *heap) {
    if (heap == NULL) {
        return;
    }
    for (size_t i = 0; i < heap->size; i++) {
        free(heap->data[i]);
    }
    free(heap->data);
    free(heap);
}

int min_heap_insert(min_heap_t *heap, timer_task_t *task) {
    if (heap == NULL || task == NULL) {
        return -1;
    }
    if (heap->size >= heap->capacity) {
        size_t new_capacity = heap->capacity * 2;
        if (heap_resize(heap, new_capacity) != 0) {
            return -1;
        }
    }
    heap->data[heap->size] = task;
    heapify_up(heap, heap->size);
    heap->size++;
    return 0;
}

timer_task_t* min_heap_extract_min(min_heap_t *heap) {
    if (heap == NULL || heap->size == 0) {
        return NULL;
    }
    timer_task_t *min = heap->data[0];
    heap->data[0] = heap->data[heap->size - 1];
    heap->size--;
    if (heap->size > 0) {
        heapify_down(heap, 0);
    }
    if (heap->size < heap->capacity / 4 && heap->capacity > INITIAL_CAPACITY) {
        heap_resize(heap, heap->capacity / 2);
    }
    return min;
}

timer_task_t* min_heap_peek(min_heap_t *heap) {
    if (heap == NULL || heap->size == 0) {
        return NULL;
    }
    return heap->data[0];
}

int min_heap_remove(min_heap_t *heap, task_id_t id) {
    if (heap == NULL || heap->size == 0) {
        return -1;
    }
    size_t found_index = (size_t)-1;
    for (size_t i = 0; i < heap->size; i++) {
        if (heap->data[i]->id == id) {
            found_index = i;
            break;
        }
    }
    if (found_index == (size_t)-1) {
        return -1;
    }
    free(heap->data[found_index]);
    heap->data[found_index] = heap->data[heap->size - 1];
    heap->size--;
    if (found_index < heap->size) {
        if (found_index == 0 || 
            timespec_compare(&heap->data[found_index]->expire_time,
                           &heap->data[(found_index - 1) / 2]->expire_time) < 0) {
            heapify_up(heap, found_index);
        } else {
            heapify_down(heap, found_index);
        }
    }
    return 0;
}

size_t min_heap_size(min_heap_t *heap) {
    if (heap == NULL) {
        return 0;
    }
    return heap->size;
}

int min_heap_is_empty(min_heap_t *heap) {
    if (heap == NULL) {
        return 1;
    }
    return heap->size == 0;
}

int timespec_compare(const struct timespec *a, const struct timespec *b) {
    if (a->tv_sec != b->tv_sec) {
        return (a->tv_sec > b->tv_sec) ? 1 : -1;
    }
    if (a->tv_nsec != b->tv_nsec) {
        return (a->tv_nsec > b->tv_nsec) ? 1 : -1;
    }
    return 0;
}

struct timespec timespec_add_ms(struct timespec ts, long ms) {
    long add_sec = ms / 1000;
    long add_nsec = (ms % 1000) * 1000000L;
    
    ts.tv_sec += add_sec;
    ts.tv_nsec += add_nsec;
    
    while (ts.tv_nsec >= 1000000000L) {
        ts.tv_sec++;
        ts.tv_nsec -= 1000000000L;
    }
    while (ts.tv_nsec < 0) {
        ts.tv_sec--;
        ts.tv_nsec += 1000000000L;
    }
    
    return ts;
}

long timespec_diff_ms(struct timespec a, struct timespec b) {
    long sec_diff = a.tv_sec - b.tv_sec;
    long nsec_diff = a.tv_nsec - b.tv_nsec;
    return sec_diff * 1000 + nsec_diff / 1000000;
}
