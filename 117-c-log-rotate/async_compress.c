#include "async_compress.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/wait.h>
#include <errno.h>

static int file_exists(const char *path)
{
    struct stat st;
    return (stat(path, &st) == 0);
}

int compress_sync(const char *src_path, const char *dst_path)
{
    if (src_path == NULL || dst_path == NULL) {
        return -1;
    }
    
    struct stat st;
    if (stat(src_path, &st) != 0) {
        return -1;
    }
    
    if (st.st_size == 0) {
        unlink(src_path);
        return 0;
    }
    
    char tmp_path[COMPRESS_MAX_PATH];
    snprintf(tmp_path, sizeof(tmp_path), "%s.XXXXXX", dst_path);
    
    int fd = mkstemp(tmp_path);
    if (fd < 0) {
        return -1;
    }
    close(fd);
    
    char gzip_cmd[2048];
    snprintf(gzip_cmd, sizeof(gzip_cmd), "gzip -c '%s' > '%s'", src_path, tmp_path);
    
    int ret = system(gzip_cmd);
    
    if (ret != 0) {
        unlink(tmp_path);
        unlink(src_path);
        return -1;
    }
    
    struct stat tmp_st;
    if (stat(tmp_path, &tmp_st) != 0 || tmp_st.st_size == 0) {
        unlink(tmp_path);
        unlink(src_path);
        return -1;
    }
    
    if (file_exists(dst_path)) {
        unlink(dst_path);
    }
    
    if (rename(tmp_path, dst_path) != 0) {
        unlink(tmp_path);
        unlink(src_path);
        return -1;
    }
    
    unlink(src_path);
    
    return 0;
}

static void *compress_worker(void *arg)
{
    async_compress_t *ac = (async_compress_t *)arg;
    
    while (1) {
        pthread_mutex_lock(&ac->mutex);
        
        while (ac->head == NULL && ac->running) {
            pthread_cond_wait(&ac->cond, &ac->mutex);
        }
        
        if (!ac->running && ac->head == NULL) {
            pthread_mutex_unlock(&ac->mutex);
            break;
        }
        
        compress_task_t *task = ac->head;
        if (task != NULL) {
            ac->head = task->next;
            if (ac->head == NULL) {
                ac->tail = NULL;
            }
        }
        
        pthread_mutex_unlock(&ac->mutex);
        
        if (task != NULL) {
            int result = compress_sync(task->src_path, task->dst_path);
            task->completed = 1;
            task->failed = (result != 0) ? 1 : 0;
            free(task);
        }
    }
    
    return NULL;
}

int async_compress_init(async_compress_t *ac)
{
    if (ac == NULL) {
        return -1;
    }
    
    memset(ac, 0, sizeof(async_compress_t));
    
    if (pthread_mutex_init(&ac->mutex, NULL) != 0) {
        return -1;
    }
    
    if (pthread_cond_init(&ac->cond, NULL) != 0) {
        pthread_mutex_destroy(&ac->mutex);
        return -1;
    }
    
    ac->running = 1;
    ac->initialized = 1;
    
    if (pthread_create(&ac->thread, NULL, compress_worker, ac) != 0) {
        pthread_cond_destroy(&ac->cond);
        pthread_mutex_destroy(&ac->mutex);
        memset(ac, 0, sizeof(async_compress_t));
        return -1;
    }
    
    return 0;
}

void async_compress_destroy(async_compress_t *ac)
{
    if (ac == NULL || !ac->initialized) {
        return;
    }
    
    pthread_mutex_lock(&ac->mutex);
    ac->running = 0;
    pthread_cond_signal(&ac->cond);
    pthread_mutex_unlock(&ac->mutex);
    
    pthread_join(ac->thread, NULL);
    
    compress_task_t *task = ac->head;
    while (task != NULL) {
        compress_task_t *next = task->next;
        free(task);
        task = next;
    }
    
    pthread_cond_destroy(&ac->cond);
    pthread_mutex_destroy(&ac->mutex);
    memset(ac, 0, sizeof(async_compress_t));
}

int async_compress_submit(async_compress_t *ac, const char *src_path, const char *dst_path)
{
    if (ac == NULL || src_path == NULL || dst_path == NULL) {
        return -1;
    }
    
    if (!ac->initialized) {
        return -1;
    }
    
    compress_task_t *task = (compress_task_t *)malloc(sizeof(compress_task_t));
    if (task == NULL) {
        return -1;
    }
    
    memset(task, 0, sizeof(compress_task_t));
    strncpy(task->src_path, src_path, COMPRESS_MAX_PATH - 1);
    task->src_path[COMPRESS_MAX_PATH - 1] = '\0';
    strncpy(task->dst_path, dst_path, COMPRESS_MAX_PATH - 1);
    task->dst_path[COMPRESS_MAX_PATH - 1] = '\0';
    task->next = NULL;
    
    pthread_mutex_lock(&ac->mutex);
    
    if (ac->tail == NULL) {
        ac->head = task;
        ac->tail = task;
    } else {
        ac->tail->next = task;
        ac->tail = task;
    }
    
    pthread_cond_signal(&ac->cond);
    pthread_mutex_unlock(&ac->mutex);
    
    return 0;
}
