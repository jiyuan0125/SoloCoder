#include "history.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define INITIAL_CAPACITY 16

void history_init(History *h) {
    h->entries = (char **)malloc(sizeof(char *) * INITIAL_CAPACITY);
    h->capacity = INITIAL_CAPACITY;
    h->count = 0;
    h->current = 0;
}

void history_free(History *h) {
    for (int i = 0; i < h->count; i++) {
        free(h->entries[i]);
    }
    free(h->entries);
    h->entries = NULL;
    h->count = 0;
    h->capacity = 0;
    h->current = 0;
}

static void history_resize(History *h, int new_capacity) {
    char **new_entries = (char **)malloc(sizeof(char *) * new_capacity);
    int copy_count = (h->count < new_capacity) ? h->count : new_capacity;
    int start = (h->count > new_capacity) ? (h->count - new_capacity) : 0;
    
    for (int i = 0; i < copy_count; i++) {
        new_entries[i] = h->entries[start + i];
    }
    
    free(h->entries);
    h->entries = new_entries;
    h->capacity = new_capacity;
    h->count = copy_count;
}

int history_add(History *h, const char *line) {
    if (line == NULL || line[0] == '\0') {
        return 0;
    }
    
    if (h->count > 0) {
        const char *last = h->entries[h->count - 1];
        if (strcmp(last, line) == 0) {
            h->current = h->count;
            return 0;
        }
    }
    
    if (h->count >= MAX_HISTORY) {
        for (int i = 1; i < h->count; i++) {
            h->entries[i - 1] = h->entries[i];
        }
        h->count--;
    }
    
    if (h->count >= h->capacity) {
        int new_capacity = h->capacity * 2;
        if (new_capacity > MAX_HISTORY) new_capacity = MAX_HISTORY;
        history_resize(h, new_capacity);
    }
    
    h->entries[h->count] = strdup(line);
    h->count++;
    h->current = h->count;
    
    return 1;
}

const char *history_prev(History *h) {
    if (h->count == 0) {
        return NULL;
    }
    if (h->current > 0) {
        h->current--;
        return h->entries[h->current];
    }
    return NULL;
}

const char *history_next(History *h) {
    if (h->count == 0) {
        return NULL;
    }
    if (h->current < h->count - 1) {
        h->current++;
        return h->entries[h->current];
    }
    h->current = h->count;
    return NULL;
}

void history_reset_pos(History *h) {
    h->current = h->count;
}

int history_get_count(const History *h) {
    return h->count;
}

const char *history_get(const History *h, int index) {
    if (index < 0 || index >= h->count) {
        return NULL;
    }
    return h->entries[index];
}
