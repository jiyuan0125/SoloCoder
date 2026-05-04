#include "lock_free_deque.h"
#include <stdlib.h>
#include <string.h>

static lf_node_t* lf_node_create(void* data) {
    lf_node_t* node = (lf_node_t*)malloc(sizeof(lf_node_t));
    if (node == NULL) {
        return NULL;
    }
    node->data = data;
    node->next = NULL;
    node->prev = NULL;
    return node;
}

lf_deque_t* lf_deque_create(size_t max_size) {
    lf_deque_t* deque = (lf_deque_t*)malloc(sizeof(lf_deque_t));
    if (deque == NULL) {
        return NULL;
    }
    deque->head = NULL;
    deque->tail = NULL;
    deque->size = 0;
    deque->max_size = max_size;
    if (pthread_spin_init(&deque->lock, PTHREAD_PROCESS_PRIVATE) != 0) {
        free(deque);
        return NULL;
    }
    return deque;
}

void lf_deque_destroy(lf_deque_t* deque, lf_data_destructor_t destructor) {
    if (deque == NULL) {
        return;
    }
    lf_deque_drop_oldest(deque, (size_t)-1, destructor);
    pthread_spin_destroy(&deque->lock);
    free(deque);
}

lf_deque_status_t lf_deque_push_back(lf_deque_t* deque, void* data) {
    if (deque == NULL) {
        return LF_DEQUE_ERROR;
    }
    
    pthread_spin_lock(&deque->lock);
    
    if (deque->max_size > 0 && deque->size >= deque->max_size) {
        pthread_spin_unlock(&deque->lock);
        return LF_DEQUE_FULL;
    }
    
    lf_node_t* node = lf_node_create(data);
    if (node == NULL) {
        pthread_spin_unlock(&deque->lock);
        return LF_DEQUE_ERROR;
    }
    
    if (deque->tail == NULL) {
        deque->head = node;
        deque->tail = node;
    } else {
        deque->tail->next = node;
        node->prev = deque->tail;
        deque->tail = node;
    }
    
    deque->size++;
    pthread_spin_unlock(&deque->lock);
    return LF_DEQUE_OK;
}

lf_deque_status_t lf_deque_push_front(lf_deque_t* deque, void* data) {
    if (deque == NULL) {
        return LF_DEQUE_ERROR;
    }
    
    pthread_spin_lock(&deque->lock);
    
    if (deque->max_size > 0 && deque->size >= deque->max_size) {
        pthread_spin_unlock(&deque->lock);
        return LF_DEQUE_FULL;
    }
    
    lf_node_t* node = lf_node_create(data);
    if (node == NULL) {
        pthread_spin_unlock(&deque->lock);
        return LF_DEQUE_ERROR;
    }
    
    if (deque->head == NULL) {
        deque->head = node;
        deque->tail = node;
    } else {
        deque->head->prev = node;
        node->next = deque->head;
        deque->head = node;
    }
    
    deque->size++;
    pthread_spin_unlock(&deque->lock);
    return LF_DEQUE_OK;
}

void* lf_deque_pop_back(lf_deque_t* deque) {
    if (deque == NULL) {
        return NULL;
    }
    
    pthread_spin_lock(&deque->lock);
    
    if (deque->tail == NULL) {
        pthread_spin_unlock(&deque->lock);
        return NULL;
    }
    
    lf_node_t* node = deque->tail;
    void* data = node->data;
    
    if (node->prev == NULL) {
        deque->head = NULL;
        deque->tail = NULL;
    } else {
        deque->tail = node->prev;
        deque->tail->next = NULL;
    }
    
    deque->size--;
    pthread_spin_unlock(&deque->lock);
    
    free(node);
    return data;
}

void* lf_deque_pop_front(lf_deque_t* deque) {
    if (deque == NULL) {
        return NULL;
    }
    
    pthread_spin_lock(&deque->lock);
    
    if (deque->head == NULL) {
        pthread_spin_unlock(&deque->lock);
        return NULL;
    }
    
    lf_node_t* node = deque->head;
    void* data = node->data;
    
    if (node->next == NULL) {
        deque->head = NULL;
        deque->tail = NULL;
    } else {
        deque->head = node->next;
        deque->head->prev = NULL;
    }
    
    deque->size--;
    pthread_spin_unlock(&deque->lock);
    
    free(node);
    return data;
}

void* lf_deque_peek_front(lf_deque_t* deque) {
    if (deque == NULL) {
        return NULL;
    }
    
    pthread_spin_lock(&deque->lock);
    void* data = (deque->head != NULL) ? deque->head->data : NULL;
    pthread_spin_unlock(&deque->lock);
    
    return data;
}

void* lf_deque_peek_back(lf_deque_t* deque) {
    if (deque == NULL) {
        return NULL;
    }
    
    pthread_spin_lock(&deque->lock);
    void* data = (deque->tail != NULL) ? deque->tail->data : NULL;
    pthread_spin_unlock(&deque->lock);
    
    return data;
}

size_t lf_deque_size(lf_deque_t* deque) {
    if (deque == NULL) {
        return 0;
    }
    
    pthread_spin_lock(&deque->lock);
    size_t size = deque->size;
    pthread_spin_unlock(&deque->lock);
    
    return size;
}

bool lf_deque_empty(lf_deque_t* deque) {
    return lf_deque_size(deque) == 0;
}

bool lf_deque_full(lf_deque_t* deque) {
    if (deque == NULL || deque->max_size == 0) {
        return false;
    }
    return lf_deque_size(deque) >= deque->max_size;
}

size_t lf_deque_drain(lf_deque_t* deque, void*** out_array, size_t* out_count) {
    if (deque == NULL || out_array == NULL || out_count == NULL) {
        return 0;
    }
    
    pthread_spin_lock(&deque->lock);
    
    size_t count = deque->size;
    if (count == 0) {
        pthread_spin_unlock(&deque->lock);
        *out_array = NULL;
        *out_count = 0;
        return 0;
    }
    
    void** array = (void**)malloc(count * sizeof(void*));
    if (array == NULL) {
        pthread_spin_unlock(&deque->lock);
        *out_array = NULL;
        *out_count = 0;
        return 0;
    }
    
    size_t i = 0;
    lf_node_t* node = deque->head;
    while (node != NULL && i < count) {
        array[i++] = node->data;
        lf_node_t* next = node->next;
        free(node);
        node = next;
    }
    
    deque->head = NULL;
    deque->tail = NULL;
    deque->size = 0;
    
    pthread_spin_unlock(&deque->lock);
    
    *out_array = array;
    *out_count = count;
    return count;
}

void lf_deque_drop_oldest(lf_deque_t* deque, size_t count, lf_data_destructor_t destructor) {
    if (deque == NULL || count == 0) {
        return;
    }
    
    pthread_spin_lock(&deque->lock);
    
    if (deque->size == 0) {
        pthread_spin_unlock(&deque->lock);
        return;
    }
    
    size_t to_drop = (count >= deque->size) ? deque->size : count;
    
    for (size_t i = 0; i < to_drop; i++) {
        if (deque->head == NULL) {
            break;
        }
        lf_node_t* node = deque->head;
        if (node->next == NULL) {
            deque->head = NULL;
            deque->tail = NULL;
        } else {
            deque->head = node->next;
            deque->head->prev = NULL;
        }
        deque->size--;
        
        if (destructor != NULL && node->data != NULL) {
            destructor(node->data);
        }
        free(node);
    }
    
    pthread_spin_unlock(&deque->lock);
}
