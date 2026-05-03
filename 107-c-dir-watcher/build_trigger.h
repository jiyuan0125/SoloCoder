#ifndef BUILD_TRIGGER_H
#define BUILD_TRIGGER_H

#include "common.h"
#include <pthread.h>
#include <sys/wait.h>

typedef struct BuildTrigger BuildTrigger;

typedef enum {
    BUILD_IDLE,
    BUILD_RUNNING,
    BUILD_PENDING
} BuildStatus;

typedef struct {
    const char *command;
    char **argv;
    bool use_shell;
} BuildConfig;

BuildTrigger *build_trigger_create(const BuildConfig *config, BuildCallback callback, void *user_data);
void build_trigger_destroy(BuildTrigger *trigger);

int build_trigger_request(BuildTrigger *trigger, const PathList *changed_files);
BuildStatus build_trigger_get_status(const BuildTrigger *trigger);
bool build_trigger_is_running(const BuildTrigger *trigger);

void build_trigger_cancel(BuildTrigger *trigger);
int build_trigger_wait(BuildTrigger *trigger);

pid_t build_trigger_get_current_pid(const BuildTrigger *trigger);

#endif
