#ifndef COMPACTION_H
#define COMPACTION_H

#include "kv_store.h"
#include <pthread.h>

struct CompactionTask {
    pthread_t thread;
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    DB* db;
    bool running;
    bool stop_requested;
};

CompactionTask* compaction_task_create(DB* db);
void compaction_task_destroy(CompactionTask* task);
int compaction_task_start(CompactionTask* task);
void compaction_task_stop(CompactionTask* task);
void compaction_trigger(CompactionTask* task);

#endif
