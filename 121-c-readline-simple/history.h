#ifndef HISTORY_H
#define HISTORY_H

#include <stddef.h>

#define MAX_HISTORY 1000

typedef struct {
    char **entries;
    int count;
    int capacity;
    int current;
} History;

void history_init(History *h);
void history_free(History *h);
int history_add(History *h, const char *line);
const char *history_prev(History *h);
const char *history_next(History *h);
void history_reset_pos(History *h);
int history_get_count(const History *h);
const char *history_get(const History *h, int index);

#endif
