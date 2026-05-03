#include "build_trigger.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <errno.h>

struct BuildTrigger {
    BuildConfig config;
    BuildCallback callback;
    void *user_data;
    
    BuildStatus status;
    pid_t current_pid;
    PathList *pending_files;
    
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    pthread_t worker_thread;
    bool running;
    bool cancel_requested;
};

static PathList *copy_path_list(const PathList *src) {
    if (!src) return NULL;
    
    PathList *copy = path_list_create(src->count > 0 ? src->count : 16);
    if (!copy) return NULL;
    
    for (size_t i = 0; i < src->count; i++) {
        if (!path_list_append(copy, src->paths[i])) {
            path_list_destroy(copy);
            return NULL;
        }
    }
    
    return copy;
}

static int execute_command(const BuildConfig *config) {
    if (!config || !config->command) {
        return -1;
    }
    
    pid_t pid = fork();
    if (pid == -1) {
        perror("fork");
        return -1;
    }
    
    if (pid == 0) {
        if (config->use_shell) {
            execl("/bin/sh", "sh", "-c", config->command, (char *)NULL);
        } else {
            execvp(config->command, config->argv);
        }
        perror("exec");
        _exit(127);
    }
    
    int status;
    if (waitpid(pid, &status, 0) == -1) {
        perror("waitpid");
        return -1;
    }
    
    if (WIFEXITED(status)) {
        return WEXITSTATUS(status);
    }
    
    return -1;
}

static void *worker_thread_func(void *arg) {
    BuildTrigger *trigger = (BuildTrigger *)arg;
    
    while (1) {
        pthread_mutex_lock(&trigger->mutex);
        
        while (trigger->status == BUILD_IDLE && trigger->running && !trigger->cancel_requested) {
            pthread_cond_wait(&trigger->cond, &trigger->mutex);
        }
        
        if (!trigger->running || trigger->cancel_requested) {
            pthread_mutex_unlock(&trigger->mutex);
            break;
        }
        
        PathList *files_to_build = NULL;
        if (trigger->pending_files) {
            files_to_build = trigger->pending_files;
            trigger->pending_files = NULL;
        }
        
        trigger->status = BUILD_RUNNING;
        pthread_mutex_unlock(&trigger->mutex);
        
        if (trigger->callback) {
            trigger->callback(files_to_build, trigger->user_data);
        }
        
        if (files_to_build) {
            execute_command(&trigger->config);
            path_list_destroy(files_to_build);
        }
        
        pthread_mutex_lock(&trigger->mutex);
        
        if (trigger->pending_files) {
            trigger->status = BUILD_PENDING;
        } else {
            trigger->status = BUILD_IDLE;
        }
        
        if (trigger->status == BUILD_PENDING) {
            pthread_cond_signal(&trigger->cond);
        }
        
        pthread_mutex_unlock(&trigger->mutex);
    }
    
    return NULL;
}

BuildTrigger *build_trigger_create(const BuildConfig *config, BuildCallback callback, void *user_data) {
    if (!config) return NULL;
    
    BuildTrigger *trigger = malloc(sizeof(BuildTrigger));
    if (!trigger) return NULL;
    
    memset(trigger, 0, sizeof(BuildTrigger));
    
    memcpy(&trigger->config, config, sizeof(BuildConfig));
    trigger->callback = callback;
    trigger->user_data = user_data;
    
    trigger->status = BUILD_IDLE;
    trigger->current_pid = -1;
    trigger->pending_files = NULL;
    trigger->running = false;
    trigger->cancel_requested = false;
    
    if (pthread_mutex_init(&trigger->mutex, NULL) != 0) {
        free(trigger);
        return NULL;
    }
    
    if (pthread_cond_init(&trigger->cond, NULL) != 0) {
        pthread_mutex_destroy(&trigger->mutex);
        free(trigger);
        return NULL;
    }
    
    trigger->running = true;
    if (pthread_create(&trigger->worker_thread, NULL, worker_thread_func, trigger) != 0) {
        pthread_cond_destroy(&trigger->cond);
        pthread_mutex_destroy(&trigger->mutex);
        free(trigger);
        return NULL;
    }
    
    return trigger;
}

void build_trigger_destroy(BuildTrigger *trigger) {
    if (!trigger) return;
    
    build_trigger_cancel(trigger);
    
    pthread_mutex_lock(&trigger->mutex);
    trigger->running = false;
    pthread_cond_signal(&trigger->cond);
    pthread_mutex_unlock(&trigger->mutex);
    
    pthread_join(trigger->worker_thread, NULL);
    
    if (trigger->pending_files) {
        path_list_destroy(trigger->pending_files);
    }
    
    pthread_cond_destroy(&trigger->cond);
    pthread_mutex_destroy(&trigger->mutex);
    
    free(trigger);
}

int build_trigger_request(BuildTrigger *trigger, const PathList *changed_files) {
    if (!trigger) return -1;
    
    pthread_mutex_lock(&trigger->mutex);
    
    if (trigger->cancel_requested) {
        pthread_mutex_unlock(&trigger->mutex);
        return -1;
    }
    
    PathList *new_pending = copy_path_list(changed_files);
    if (!new_pending) {
        pthread_mutex_unlock(&trigger->mutex);
        return -1;
    }
    
    if (trigger->pending_files) {
        for (size_t i = 0; i < new_pending->count; i++) {
            if (!path_list_contains(trigger->pending_files, new_pending->paths[i])) {
                path_list_append(trigger->pending_files, new_pending->paths[i]);
            }
        }
        path_list_destroy(new_pending);
    } else {
        trigger->pending_files = new_pending;
    }
    
    if (trigger->status == BUILD_IDLE) {
        trigger->status = BUILD_PENDING;
        pthread_cond_signal(&trigger->cond);
    }
    
    pthread_mutex_unlock(&trigger->mutex);
    
    return 0;
}

BuildStatus build_trigger_get_status(const BuildTrigger *trigger) {
    if (!trigger) return BUILD_IDLE;
    return trigger->status;
}

bool build_trigger_is_running(const BuildTrigger *trigger) {
    if (!trigger) return false;
    
    pthread_mutex_lock((pthread_mutex_t *)&trigger->mutex);
    bool running = (trigger->status == BUILD_RUNNING);
    pthread_mutex_unlock((pthread_mutex_t *)&trigger->mutex);
    
    return running;
}

void build_trigger_cancel(BuildTrigger *trigger) {
    if (!trigger) return;
    
    pthread_mutex_lock(&trigger->mutex);
    trigger->cancel_requested = true;
    
    if (trigger->pending_files) {
        path_list_destroy(trigger->pending_files);
        trigger->pending_files = NULL;
    }
    
    pthread_cond_signal(&trigger->cond);
    pthread_mutex_unlock(&trigger->mutex);
}

int build_trigger_wait(BuildTrigger *trigger) {
    if (!trigger) return -1;
    
    pthread_mutex_lock(&trigger->mutex);
    while (trigger->status != BUILD_IDLE && trigger->running) {
        pthread_cond_wait(&trigger->cond, &trigger->mutex);
    }
    pthread_mutex_unlock(&trigger->mutex);
    
    return 0;
}

pid_t build_trigger_get_current_pid(const BuildTrigger *trigger) {
    if (!trigger) return -1;
    return trigger->current_pid;
}
