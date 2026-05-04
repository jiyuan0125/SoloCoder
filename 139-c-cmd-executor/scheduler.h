#ifndef SCHEDULER_H
#define SCHEDULER_H

#include "common.h"

typedef struct scheduler_s scheduler_t;

scheduler_t *scheduler_create(int max_concurrent);
void scheduler_destroy(scheduler_t *sched);

int scheduler_add_task(scheduler_t *sched, cmd_config_t *config);
int scheduler_run(scheduler_t *sched);
cmd_result_t *scheduler_get_results(scheduler_t *sched, size_t *count);
int scheduler_get_task_count(scheduler_t *sched);

#endif
