#ifndef LOG_CHANNEL_H
#define LOG_CHANNEL_H

#include <stdint.h>
#include <stdbool.h>
#include <pthread.h>
#include <sys/time.h>
#include "lock_free_deque.h"

#define LOG_MESSAGE_MAX_LEN  1024
#define LOG_MAX_THREADS      64

typedef enum {
    LOG_POLICY_BLOCK = 0,
    LOG_POLICY_DROP_NEWEST,
    LOG_POLICY_DROP_OLDEST
} log_policy_t;

typedef enum {
    LOG_LEVEL_DEBUG = 0,
    LOG_LEVEL_INFO,
    LOG_LEVEL_WARN,
    LOG_LEVEL_ERROR
} log_level_t;

typedef struct log_message_s {
    struct timeval timestamp;
    pthread_t      thread_id;
    log_level_t    level;
    size_t         message_len;
    char           message[LOG_MESSAGE_MAX_LEN];
} log_message_t;

typedef struct log_channel_config_s {
    size_t         max_total_messages;
    size_t         max_thread_messages;
    log_policy_t   overflow_policy;
    uint32_t       max_threads;
} log_channel_config_t;

typedef struct log_writer_s {
    lf_deque_t*         queue;
    pthread_t           thread_id;
    bool                is_active;
} log_writer_t;

typedef struct log_channel_s {
    log_writer_t        writers[LOG_MAX_THREADS];
    uint32_t            writer_count;
    pthread_rwlock_t    writers_lock;
    
    log_channel_config_t config;
    
    volatile bool       is_shutdown;
    pthread_cond_t      space_available;
    pthread_mutex_t     channel_mutex;
    
    pthread_spinlock_t  stats_lock;
    size_t              total_dropped;
    size_t              total_written;
} log_channel_t;

log_channel_t* log_channel_create(const log_channel_config_t* config);
void log_channel_destroy(log_channel_t* channel);

log_writer_t* log_channel_register_writer(log_channel_t* channel);
void log_channel_unregister_writer(log_channel_t* channel, log_writer_t* writer);

int log_channel_write(log_channel_t* channel, log_writer_t* writer,
                       log_level_t level, const char* format, ...);

int log_channel_write_ex(log_channel_t* channel, log_writer_t* writer,
                          log_message_t* msg);

size_t log_channel_collect_all(log_channel_t* channel, log_message_t*** out_messages,
                                size_t* out_count);

size_t log_channel_total_size(log_channel_t* channel);
bool log_channel_is_shutdown(log_channel_t* channel);
void log_channel_shutdown(log_channel_t* channel);

void log_channel_drop_half_oldest(log_channel_t* channel);
size_t log_channel_get_stats(log_channel_t* channel, size_t* dropped, size_t* written);

log_message_t* log_message_create(log_level_t level, const char* format, ...);
void log_message_destroy(log_message_t* msg);

void log_message_array_destroy(log_message_t** array, size_t count);

#endif
