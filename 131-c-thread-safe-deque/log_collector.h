#ifndef LOG_COLLECTOR_H
#define LOG_COLLECTOR_H

#include <stdio.h>
#include <pthread.h>
#include "log_channel.h"

typedef enum {
    LOG_OUTPUT_STDOUT = 0,
    LOG_OUTPUT_FILE,
    LOG_OUTPUT_BOTH
} log_output_type_t;

typedef struct log_collector_config_s {
    log_output_type_t  output_type;
    const char*        output_file;
    bool               append_to_file;
    uint32_t           flush_interval_ms;
    bool               sort_by_timestamp;
    size_t             batch_size_hint;
} log_collector_config_t;

typedef struct log_collector_s {
    log_channel_t*         channel;
    log_collector_config_t config;
    
    FILE*                  output_fp;
    volatile bool          is_running;
    volatile bool          is_stopped;
    pthread_t              collector_thread;
    
    pthread_mutex_t        collector_mutex;
    pthread_cond_t         wakeup_cond;
    
    uint64_t               total_processed;
    uint64_t               total_flushed;
} log_collector_t;

log_collector_t* log_collector_create(log_channel_t* channel,
                                        const log_collector_config_t* config);
void log_collector_destroy(log_collector_t* collector);

int log_collector_start(log_collector_t* collector);
int log_collector_stop(log_collector_t* collector, bool wait_for_drain);
int log_collector_flush(log_collector_t* collector);

size_t log_collector_collect_and_sort(log_collector_t* collector,
                                        log_message_t*** out_messages,
                                        size_t* out_count);

int log_collector_write_messages(log_collector_t* collector,
                                   log_message_t** messages,
                                   size_t count);

const char* log_level_to_string(log_level_t level);

void log_message_sort_by_timestamp(log_message_t** messages, size_t count);

uint64_t log_collector_get_stats(log_collector_t* collector,
                                   uint64_t* out_flushed);

#endif
