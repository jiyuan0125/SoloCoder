#include "shm_queue.h"
#include "shm_manager.h"
#include "sem_sync.h"
#include "msg_buffer.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <time.h>

#define SHMQ_ABNORMAL_TIMEOUT_MS  5000

typedef struct shm_queue_private {
    shm_manager_t shm;
    sem_set_t sems;
    msg_buffer_t msgbuf;
} shm_queue_private_t;

void shmq_config_init(shm_queue_config_t *cfg) {
    memset(cfg, 0, sizeof(shm_queue_config_t));
    cfg->name = SHMQ_DEFAULT_NAME;
    cfg->buffer_size = 1024 * 1024;
    cfg->max_msg_len = 65536;
    cfg->default_timeout_ms = SHMQ_WAIT_INFINITE;
    cfg->reuse_existing = false;
    cfg->cleanup_on_destroy = true;
    cfg->max_concurrent_writers = SHMQ_DEFAULT_WRITERS;
}

size_t shmq_calculate_total_size(size_t buffer_size) {
    return msg_buffer_calculate_size(buffer_size);
}

static void shmq_init(shm_queue_t *q) {
    memset(q, 0, sizeof(shm_queue_t));
    q->role = SHMQ_ROLE_UNKNOWN;
}

static int do_create(shm_queue_t *q, const shm_queue_config_t *cfg) {
    shm_queue_private_t *priv = NULL;
    size_t total_size;
    
    total_size = shmq_calculate_total_size(cfg->buffer_size);
    
    priv = (shm_queue_private_t *)malloc(sizeof(shm_queue_private_t));
    if (priv == NULL) {
        return SHMQ_ERROR;
    }
    memset(priv, 0, sizeof(shm_queue_private_t));
    
    q->config = *cfg;
    q->config.name = strdup(cfg->name);
    if (q->config.name == NULL) {
        free(priv);
        return SHMQ_ERROR;
    }
    
    if (shm_manager_create(&priv->shm, cfg->name, total_size) < 0) {
        free((void *)q->config.name);
        free(priv);
        return SHMQ_ERROR;
    }
    
    unsigned int max_writers = cfg->max_concurrent_writers;
    if (max_writers == 0) {
        max_writers = SHMQ_DEFAULT_WRITERS;
    }
    if (sem_set_create(&priv->sems, cfg->name, 1, max_writers, 0) < 0) {
        shm_manager_close(&priv->shm);
        shm_manager_unlink(cfg->name);
        free((void *)q->config.name);
        free(priv);
        return SHMQ_ERROR;
    }
    
    if (msg_buffer_init(&priv->msgbuf, priv->shm.addr, total_size,
                         cfg->max_msg_len, cfg->buffer_size) < 0) {
        sem_set_close(&priv->sems);
        sem_set_unlink(cfg->name);
        shm_manager_close(&priv->shm);
        shm_manager_unlink(cfg->name);
        free((void *)q->config.name);
        free(priv);
        return SHMQ_ERROR;
    }
    
    q->private_data = priv;
    q->role = SHMQ_ROLE_CREATOR;
    return SHMQ_OK;
}

static int try_recover(shm_queue_t *q, const char *name) {
    shm_manager_t shm;
    msg_buffer_t msgbuf;
    
    (void)q;
    
    if (shm_manager_open(&shm, name, 4096) < 0) {
        return SHMQ_ERROR;
    }
    
    if (msg_buffer_attach(&msgbuf, shm.addr, 4096) < 0) {
        shm_manager_close(&shm);
        return SHMQ_ERROR;
    }
    
    msg_buffer_detach(&msgbuf);
    shm_manager_close(&shm);
    
    shm_manager_unlink(name);
    sem_set_unlink(name);
    
    return SHMQ_OK;
}

int shmq_create(shm_queue_t *q, const shm_queue_config_t *cfg) {
    shm_queue_config_t default_cfg;
    
    shmq_init(q);
    
    if (cfg == NULL) {
        shmq_config_init(&default_cfg);
        cfg = &default_cfg;
    }
    
    if (cfg->buffer_size < SHMQ_MIN_BUFFER_SIZE) {
        return SHMQ_ERROR;
    }
    
    if (cfg->max_msg_len > SHMQ_MAX_MSG_LEN) {
        return SHMQ_ERROR;
    }
    
    if (shmq_exists(cfg->name)) {
        if (cfg->reuse_existing) {
            return shmq_open(q, cfg->name);
        } else {
            if (try_recover(q, cfg->name) < 0) {
                return SHMQ_EXISTS;
            }
        }
    }
    
    return do_create(q, cfg);
}

int shmq_open(shm_queue_t *q, const char *name) {
    shm_queue_private_t *priv = NULL;
    shm_manager_t shm;
    msg_buffer_t msgbuf;
    size_t total_size;
    
    shmq_init(q);
    
    if (name == NULL) {
        name = SHMQ_DEFAULT_NAME;
    }
    
    if (shm_manager_open(&shm, name, 4096) < 0) {
        return SHMQ_NOT_EXISTS;
    }
    
    if (msg_buffer_attach(&msgbuf, shm.addr, 4096) < 0) {
        shm_manager_close(&shm);
        return SHMQ_ERROR;
    }
    
    total_size = shmq_calculate_total_size(msgbuf.header->buffer_size);
    msg_buffer_detach(&msgbuf);
    shm_manager_close(&shm);
    
    priv = (shm_queue_private_t *)malloc(sizeof(shm_queue_private_t));
    if (priv == NULL) {
        return SHMQ_ERROR;
    }
    memset(priv, 0, sizeof(shm_queue_private_t));
    
    q->config.name = strdup(name);
    if (q->config.name == NULL) {
        free(priv);
        return SHMQ_ERROR;
    }
    q->config.cleanup_on_destroy = false;
    
    if (shm_manager_open(&priv->shm, name, total_size) < 0) {
        free((void *)q->config.name);
        free(priv);
        return SHMQ_NOT_EXISTS;
    }
    
    if (sem_set_open(&priv->sems, name) < 0) {
        shm_manager_close(&priv->shm);
        free((void *)q->config.name);
        free(priv);
        return SHMQ_NOT_EXISTS;
    }
    
    if (msg_buffer_attach(&priv->msgbuf, priv->shm.addr, total_size) < 0) {
        sem_set_close(&priv->sems);
        shm_manager_close(&priv->shm);
        free((void *)q->config.name);
        free(priv);
        return SHMQ_ERROR;
    }
    
    q->config.buffer_size = priv->msgbuf.header->buffer_size;
    q->config.max_msg_len = priv->msgbuf.header->max_msg_len;
    q->config.default_timeout_ms = SHMQ_WAIT_INFINITE;
    q->private_data = priv;
    q->role = SHMQ_ROLE_ATTACHER;
    
    return SHMQ_OK;
}

int shmq_close(shm_queue_t *q) {
    shm_queue_private_t *priv;
    
    if (q == NULL || q->private_data == NULL) {
        return SHMQ_ERROR;
    }
    
    priv = (shm_queue_private_t *)q->private_data;
    
    msg_buffer_detach(&priv->msgbuf);
    sem_set_close(&priv->sems);
    shm_manager_close(&priv->shm);
    
    if (q->config.name != NULL) {
        free((void *)q->config.name);
    }
    free(priv);
    
    shmq_init(q);
    return SHMQ_OK;
}

int shmq_destroy(const char *name) {
    if (name == NULL) {
        name = SHMQ_DEFAULT_NAME;
    }
    
    shm_manager_unlink(name);
    sem_set_unlink(name);
    
    return SHMQ_OK;
}

int shmq_destroy_queue(shm_queue_t *q) {
    const char *name;
    
    if (q == NULL) {
        return SHMQ_ERROR;
    }
    
    name = q->config.name;
    if (name == NULL) {
        name = SHMQ_DEFAULT_NAME;
    }
    
    shmq_close(q);
    shmq_destroy(name);
    
    return SHMQ_OK;
}

bool shmq_exists(const char *name) {
    if (name == NULL) {
        name = SHMQ_DEFAULT_NAME;
    }
    return shm_manager_exists(name);
}

static int wait_with_recovery(sem_sync_t *sem, int timeout_ms, int max_retries) {
    int ret;
    int retries = 0;
    
    while (retries < max_retries) {
        if (timeout_ms == SHMQ_WAIT_INFINITE) {
            ret = sem_sync_timedwait(sem, SHMQ_ABNORMAL_TIMEOUT_MS);
        } else {
            ret = sem_sync_timedwait(sem, timeout_ms);
        }
        
        if (ret == 0) {
            return 0;
        }
        
        if (errno != ETIMEDOUT || timeout_ms != SHMQ_WAIT_INFINITE) {
            return ret;
        }
        
        retries++;
    }
    
    errno = ETIMEDOUT;
    return -1;
}

int shmq_send(shm_queue_t *q, const void *data, size_t len, int timeout_ms) {
    shm_queue_private_t *priv;
    int ret;
    
    if (q == NULL || q->private_data == NULL) {
        return SHMQ_ERROR;
    }
    
    priv = (shm_queue_private_t *)q->private_data;
    
    if (len > q->config.max_msg_len) {
        return SHMQ_ERROR;
    }
    
    if (timeout_ms == SHMQ_WAIT_NOWAIT) {
        if (sem_sync_trywait(&priv->sems.empty) < 0) {
            return SHMQ_FULL;
        }
    } else {
        ret = wait_with_recovery(&priv->sems.empty, timeout_ms, 3);
        if (ret < 0) {
            if (errno == EAGAIN || errno == ETIMEDOUT) {
                return SHMQ_TIMEOUT;
            }
            return SHMQ_ERROR;
        }
    }
    
    if (sem_sync_wait(&priv->sems.mutex) < 0) {
        sem_sync_post(&priv->sems.empty);
        return SHMQ_ERROR;
    }
    
    if (msg_buffer_write(&priv->msgbuf, data, len) < 0) {
        sem_sync_post(&priv->sems.mutex);
        sem_sync_post(&priv->sems.empty);
        return SHMQ_ERROR;
    }
    
    sem_sync_post(&priv->sems.mutex);
    sem_sync_post(&priv->sems.full);
    sem_sync_post(&priv->sems.empty);
    
    return SHMQ_OK;
}

int shmq_recv(shm_queue_t *q, void *data, size_t *len, int timeout_ms) {
    shm_queue_private_t *priv;
    int ret;
    
    if (q == NULL || q->private_data == NULL || len == NULL) {
        return SHMQ_ERROR;
    }
    
    priv = (shm_queue_private_t *)q->private_data;
    
    if (timeout_ms == SHMQ_WAIT_NOWAIT) {
        if (sem_sync_trywait(&priv->sems.full) < 0) {
            return SHMQ_EMPTY;
        }
    } else {
        ret = wait_with_recovery(&priv->sems.full, timeout_ms, 3);
        if (ret < 0) {
            if (errno == EAGAIN || errno == ETIMEDOUT) {
                return SHMQ_TIMEOUT;
            }
            return SHMQ_ERROR;
        }
    }
    
    if (sem_sync_wait(&priv->sems.mutex) < 0) {
        sem_sync_post(&priv->sems.full);
        return SHMQ_ERROR;
    }
    
    if (msg_buffer_read(&priv->msgbuf, data, len) < 0) {
        sem_sync_post(&priv->sems.mutex);
        sem_sync_post(&priv->sems.full);
        return SHMQ_ERROR;
    }
    
    sem_sync_post(&priv->sems.mutex);
    
    return SHMQ_OK;
}

int shmq_try_send(shm_queue_t *q, const void *data, size_t len) {
    return shmq_send(q, data, len, SHMQ_WAIT_NOWAIT);
}

int shmq_try_recv(shm_queue_t *q, void *data, size_t *len) {
    return shmq_recv(q, data, len, SHMQ_WAIT_NOWAIT);
}

size_t shmq_available_space(shm_queue_t *q) {
    shm_queue_private_t *priv;
    
    if (q == NULL || q->private_data == NULL) {
        return 0;
    }
    
    priv = (shm_queue_private_t *)q->private_data;
    return msg_buffer_available_space(&priv->msgbuf);
}

size_t shmq_available_data(shm_queue_t *q) {
    shm_queue_private_t *priv;
    
    if (q == NULL || q->private_data == NULL) {
        return 0;
    }
    
    priv = (shm_queue_private_t *)q->private_data;
    return msg_buffer_available_data(&priv->msgbuf);
}

bool shmq_is_full(shm_queue_t *q) {
    shm_queue_private_t *priv;
    
    if (q == NULL || q->private_data == NULL) {
        return true;
    }
    
    priv = (shm_queue_private_t *)q->private_data;
    return !msg_buffer_can_write(&priv->msgbuf, 1);
}

bool shmq_is_empty(shm_queue_t *q) {
    shm_queue_private_t *priv;
    
    if (q == NULL || q->private_data == NULL) {
        return true;
    }
    
    priv = (shm_queue_private_t *)q->private_data;
    return !msg_buffer_can_read(&priv->msgbuf);
}

int shmq_get_max_msg_len(shm_queue_t *q, size_t *out_len) {
    if (q == NULL || out_len == NULL) {
        return SHMQ_ERROR;
    }
    *out_len = q->config.max_msg_len;
    return SHMQ_OK;
}

int shmq_get_buffer_size(shm_queue_t *q, size_t *out_size) {
    if (q == NULL || out_size == NULL) {
        return SHMQ_ERROR;
    }
    *out_size = q->config.buffer_size;
    return SHMQ_OK;
}
