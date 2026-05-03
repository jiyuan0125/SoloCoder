#include "sem_sync.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <errno.h>

void sem_sync_init(sem_sync_t *sync) {
    memset(sync, 0, sizeof(sem_sync_t));
    sync->sem = SEM_FAILED;
}

static int sem_do_open(sem_sync_t *sync, const char *name, int oflag, 
                        mode_t mode, unsigned int value) {
    sem_sync_init(sync);
    
    sync->name = strdup(name);
    if (sync->name == NULL) {
        return -1;
    }
    
    if (oflag & O_CREAT) {
        sync->sem = sem_open(name, oflag, mode, value);
        sync->owner = true;
    } else {
        sync->sem = sem_open(name, oflag);
    }
    
    if (sync->sem == SEM_FAILED) {
        free((void *)sync->name);
        sync->name = NULL;
        return -1;
    }
    
    return 0;
}

int sem_sync_create(sem_sync_t *sync, const char *name, unsigned int value) {
    return sem_do_open(sync, name, O_CREAT | O_EXCL | O_RDWR, 
                       S_IRUSR | S_IWUSR, value);
}

int sem_sync_open(sem_sync_t *sync, const char *name) {
    return sem_do_open(sync, name, O_RDWR, 0, 0);
}

int sem_sync_close(sem_sync_t *sync) {
    int ret = 0;
    
    if (sync->sem != SEM_FAILED) {
        if (sem_close(sync->sem) < 0) {
            ret = -1;
        }
        sync->sem = SEM_FAILED;
    }
    
    if (sync->name != NULL) {
        free((void *)sync->name);
        sync->name = NULL;
    }
    
    return ret;
}

int sem_sync_unlink(const char *name) {
    return sem_unlink(name);
}

bool sem_sync_exists(const char *name) {
    sem_t *sem = sem_open(name, O_RDONLY);
    if (sem != SEM_FAILED) {
        sem_close(sem);
        return true;
    }
    return (errno != ENOENT);
}

int sem_sync_wait(sem_sync_t *sync) {
    return sem_wait(sync->sem);
}

int sem_sync_trywait(sem_sync_t *sync) {
    return sem_trywait(sync->sem);
}

int sem_sync_timedwait(sem_sync_t *sync, int timeout_ms) {
    struct timespec ts;
    
    if (timeout_ms == SEM_SYNC_WAIT_NOWAIT) {
        return sem_trywait(sync->sem);
    }
    
    if (timeout_ms == SEM_SYNC_WAIT_INFINITE) {
        return sem_wait(sync->sem);
    }
    
    if (clock_gettime(CLOCK_REALTIME, &ts) < 0) {
        return -1;
    }
    
    ts.tv_sec += timeout_ms / 1000;
    ts.tv_nsec += (timeout_ms % 1000) * 1000000;
    
    while (ts.tv_nsec >= 1000000000) {
        ts.tv_nsec -= 1000000000;
        ts.tv_sec += 1;
    }
    
    return sem_timedwait(sync->sem, &ts);
}

int sem_sync_post(sem_sync_t *sync) {
    return sem_post(sync->sem);
}

int sem_sync_getvalue(sem_sync_t *sync, int *value) {
    return sem_getvalue(sync->sem, value);
}

void sem_set_init(sem_set_t *set) {
    memset(set, 0, sizeof(sem_set_t));
    sem_sync_init(&set->mutex);
    sem_sync_init(&set->empty);
    sem_sync_init(&set->full);
}

static void make_sem_names(const char *base, char *mutex_name, 
                            char *empty_name, char *full_name, size_t max_len) {
    snprintf(mutex_name, max_len, "%s_mtx", base);
    snprintf(empty_name, max_len, "%s_emp", base);
    snprintf(full_name, max_len, "%s_ful", base);
}

int sem_set_create(sem_set_t *set, const char *base_name, 
                   unsigned int mutex_val, unsigned int empty_val, unsigned int full_val) {
    char mutex_name[256];
    char empty_name[256];
    char full_name[256];
    
    sem_set_init(set);
    make_sem_names(base_name, mutex_name, empty_name, full_name, sizeof(mutex_name));
    
    set->base_name = strdup(base_name);
    if (set->base_name == NULL) {
        return -1;
    }
    
    if (sem_sync_create(&set->mutex, mutex_name, mutex_val) < 0) {
        free((void *)set->base_name);
        set->base_name = NULL;
        return -1;
    }
    
    if (sem_sync_create(&set->empty, empty_name, empty_val) < 0) {
        sem_sync_close(&set->mutex);
        sem_sync_unlink(mutex_name);
        free((void *)set->base_name);
        set->base_name = NULL;
        return -1;
    }
    
    if (sem_sync_create(&set->full, full_name, full_val) < 0) {
        sem_sync_close(&set->empty);
        sem_sync_unlink(empty_name);
        sem_sync_close(&set->mutex);
        sem_sync_unlink(mutex_name);
        free((void *)set->base_name);
        set->base_name = NULL;
        return -1;
    }
    
    return 0;
}

int sem_set_open(sem_set_t *set, const char *base_name) {
    char mutex_name[256];
    char empty_name[256];
    char full_name[256];
    
    sem_set_init(set);
    make_sem_names(base_name, mutex_name, empty_name, full_name, sizeof(mutex_name));
    
    set->base_name = strdup(base_name);
    if (set->base_name == NULL) {
        return -1;
    }
    
    if (sem_sync_open(&set->mutex, mutex_name) < 0) {
        free((void *)set->base_name);
        set->base_name = NULL;
        return -1;
    }
    
    if (sem_sync_open(&set->empty, empty_name) < 0) {
        sem_sync_close(&set->mutex);
        free((void *)set->base_name);
        set->base_name = NULL;
        return -1;
    }
    
    if (sem_sync_open(&set->full, full_name) < 0) {
        sem_sync_close(&set->empty);
        sem_sync_close(&set->mutex);
        free((void *)set->base_name);
        set->base_name = NULL;
        return -1;
    }
    
    return 0;
}

int sem_set_close(sem_set_t *set) {
    int ret = 0;
    
    if (sem_sync_close(&set->full) < 0) ret = -1;
    if (sem_sync_close(&set->empty) < 0) ret = -1;
    if (sem_sync_close(&set->mutex) < 0) ret = -1;
    
    if (set->base_name != NULL) {
        free((void *)set->base_name);
        set->base_name = NULL;
    }
    
    return ret;
}

int sem_set_unlink(const char *base_name) {
    char mutex_name[256];
    char empty_name[256];
    char full_name[256];
    int ret = 0;
    
    make_sem_names(base_name, mutex_name, empty_name, full_name, sizeof(mutex_name));
    
    if (sem_sync_unlink(full_name) < 0 && errno != ENOENT) ret = -1;
    if (sem_sync_unlink(empty_name) < 0 && errno != ENOENT) ret = -1;
    if (sem_sync_unlink(mutex_name) < 0 && errno != ENOENT) ret = -1;
    
    return ret;
}

bool sem_set_exists(const char *base_name) {
    char mutex_name[256];
    char empty_name[256];
    char full_name[256];
    
    make_sem_names(base_name, mutex_name, empty_name, full_name, sizeof(mutex_name));
    
    return sem_sync_exists(mutex_name) && 
           sem_sync_exists(empty_name) && 
           sem_sync_exists(full_name);
}
