#include <stdlib.h>
#include <string.h>
#include "event_loop_internal.h"

static void heap_swap(struct timer_heap *heap, int i, int j) {
    timer_node_t *temp = heap->nodes[i];
    heap->nodes[i] = heap->nodes[j];
    heap->nodes[j] = temp;
    
    heap->nodes[i]->heap_index = i;
    heap->nodes[j]->heap_index = j;
}

static void heap_ify_up(struct timer_heap *heap, int index) {
    while (index > 0) {
        int parent = (index - 1) / 2;
        if (heap->nodes[parent]->expire_ms <= heap->nodes[index]->expire_ms) {
            break;
        }
        heap_swap(heap, parent, index);
        index = parent;
    }
}

static void heap_ify_down(struct timer_heap *heap, int index) {
    int min_index = index;
    int left = 2 * index + 1;
    int right = 2 * index + 2;
    
    if (left < heap->size && heap->nodes[left]->expire_ms < heap->nodes[min_index]->expire_ms) {
        min_index = left;
    }
    if (right < heap->size && heap->nodes[right]->expire_ms < heap->nodes[min_index]->expire_ms) {
        min_index = right;
    }
    
    if (min_index != index) {
        heap_swap(heap, index, min_index);
        heap_ify_down(heap, min_index);
    }
}

struct timer_heap* timer_heap_create(void) {
    struct timer_heap *heap = (struct timer_heap *)malloc(sizeof(struct timer_heap));
    if (!heap) return NULL;
    
    heap->nodes = (timer_node_t **)malloc(sizeof(timer_node_t *) * INITIAL_HEAP_CAPACITY);
    if (!heap->nodes) {
        free(heap);
        return NULL;
    }
    
    heap->size = 0;
    heap->capacity = INITIAL_HEAP_CAPACITY;
    return heap;
}

void timer_heap_destroy(struct timer_heap *heap) {
    if (!heap) return;
    
    for (int i = 0; i < heap->size; i++) {
        free(heap->nodes[i]);
    }
    free(heap->nodes);
    free(heap);
}

int timer_heap_insert(struct timer_heap *heap, timer_node_t *node) {
    if (!heap || !node) return -1;
    
    if (heap->size >= heap->capacity) {
        int new_capacity = heap->capacity * 2;
        timer_node_t **new_nodes = (timer_node_t **)realloc(heap->nodes, sizeof(timer_node_t *) * new_capacity);
        if (!new_nodes) return -1;
        
        heap->nodes = new_nodes;
        heap->capacity = new_capacity;
    }
    
    node->heap_index = heap->size;
    node->is_pending = 0;
    heap->nodes[heap->size] = node;
    heap->size++;
    
    heap_ify_up(heap, heap->size - 1);
    return 0;
}

timer_node_t* timer_heap_peek(struct timer_heap *heap) {
    if (!heap || heap->size == 0) return NULL;
    return heap->nodes[0];
}

timer_node_t* timer_heap_pop(struct timer_heap *heap) {
    if (!heap || heap->size == 0) return NULL;
    
    timer_node_t *min_node = heap->nodes[0];
    heap->nodes[0] = heap->nodes[heap->size - 1];
    heap->nodes[0]->heap_index = 0;
    heap->size--;
    
    if (heap->size > 0) {
        heap_ify_down(heap, 0);
    }
    
    min_node->heap_index = -1;
    return min_node;
}

int timer_heap_remove(struct timer_heap *heap, timer_node_t *node) {
    if (!heap || !node || node->heap_index < 0 || node->heap_index >= heap->size) {
        return -1;
    }
    
    int index = node->heap_index;
    
    if (index == heap->size - 1) {
        heap->size--;
        node->heap_index = -1;
        return 0;
    }
    
    heap_swap(heap, index, heap->size - 1);
    heap->size--;
    node->heap_index = -1;
    
    if (index > 0 && heap->nodes[index]->expire_ms < heap->nodes[(index - 1) / 2]->expire_ms) {
        heap_ify_up(heap, index);
    } else {
        heap_ify_down(heap, index);
    }
    
    return 0;
}
