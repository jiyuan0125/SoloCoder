#ifndef LOG_ROTATE_H
#define LOG_ROTATE_H

#include <stdio.h>
#include <pthread.h>
#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

#define LOG_ROTATE_MAX_PATH 1024

typedef enum {
    LOG_ROTATE_MODE_SIZE = 1 << 0,
    LOG_ROTATE_MODE_TIME = 1 << 1,
    LOG_ROTATE_MODE_BOTH = (1 << 0) | (1 << 1)
} log_rotate_mode_t;

typedef struct {
    char base_path[LOG_ROTATE_MAX_PATH];
    size_t max_size;
    int max_backups;
    int mode;
    int compress_enabled;
} log_rotate_config_t;

typedef struct {
    log_rotate_config_t config;
    FILE *fp;
    pthread_mutex_t mutex;
    time_t last_rotate_date;
    int initialized;
} log_rotate_t;

int log_rotate_init(log_rotate_t *lr, const log_rotate_config_t *config);
void log_rotate_destroy(log_rotate_t *lr);
int log_rotate_write(log_rotate_t *lr, const char *format, ...);
int log_rotate_flush(log_rotate_t *lr);
void log_rotate_cleanup_compressor(void);

#ifdef __cplusplus
}
#endif

#endif
