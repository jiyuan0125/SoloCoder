#ifndef ASYNC_COMPRESS_H
#define ASYNC_COMPRESS_H

#include <pthread.h>

#ifdef __cplusplus
extern "C" {
#endif

#define COMPRESS_MAX_PATH 1024

typedef struct compress_task compress_task_t;

struct compress_task {
    char src_path[COMPRESS_MAX_PATH];
    char dst_path[COMPRESS_MAX_PATH];
    int completed;
    int failed;
    compress_task_t *next;
};

typedef struct {
    compress_task_t *head;
    compress_task_t *tail;
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    pthread_t thread;
    int running;
    int initialized;
} async_compress_t;

int async_compress_init(async_compress_t *ac);
void async_compress_destroy(async_compress_t *ac);
int async_compress_submit(async_compress_t *ac, const char *src_path, const char *dst_path);
int compress_sync(const char *src_path, const char *dst_path);

#ifdef __cplusplus
}
#endif

#endif
