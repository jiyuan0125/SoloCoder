#include "log_collector.h"
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <time.h>
#include <sys/time.h>

const char* log_level_to_string(log_level_t level) {
    switch (level) {
        case LOG_LEVEL_DEBUG: return "DEBUG";
        case LOG_LEVEL_INFO:  return "INFO ";
        case LOG_LEVEL_WARN:  return "WARN ";
        case LOG_LEVEL_ERROR: return "ERROR";
        default:              return "UNKWN";
    }
}

static int log_message_compare(const void* a, const void* b) {
    const log_message_t* const* ma = (const log_message_t* const*)a;
    const log_message_t* const* mb = (const log_message_t* const*)b;
    
    const log_message_t* msg_a = *ma;
    const log_message_t* msg_b = *mb;
    
    if (msg_a->timestamp.tv_sec != msg_b->timestamp.tv_sec) {
        return (msg_a->timestamp.tv_sec > msg_b->timestamp.tv_sec) ? 1 : -1;
    }
    if (msg_a->timestamp.tv_usec != msg_b->timestamp.tv_usec) {
        return (msg_a->timestamp.tv_usec > msg_b->timestamp.tv_usec) ? 1 : -1;
    }
    
    if (msg_a->thread_id != msg_b->thread_id) {
        return (msg_a->thread_id > msg_b->thread_id) ? 1 : -1;
    }
    
    return 0;
}

void log_message_sort_by_timestamp(log_message_t** messages, size_t count) {
    if (messages == NULL || count < 2) {
        return;
    }
    qsort(messages, count, sizeof(log_message_t*), log_message_compare);
}

log_collector_t* log_collector_create(log_channel_t* channel,
                                        const log_collector_config_t* config) {
    if (channel == NULL || config == NULL) {
        return NULL;
    }
    
    log_collector_t* collector = (log_collector_t*)malloc(sizeof(log_collector_t));
    if (collector == NULL) {
        return NULL;
    }
    
    memset(collector, 0, sizeof(log_collector_t));
    
    collector->channel = channel;
    collector->config = *config;
    
    if (collector->config.flush_interval_ms == 0) {
        collector->config.flush_interval_ms = 100;
    }
    
    if (collector->config.batch_size_hint == 0) {
        collector->config.batch_size_hint = 1024;
    }
    
    collector->output_fp = NULL;
    
    if (config->output_type == LOG_OUTPUT_FILE || 
        config->output_type == LOG_OUTPUT_BOTH) {
        if (config->output_file != NULL && config->output_file[0] != '\0') {
            const char* mode = config->append_to_file ? "a" : "w";
            collector->output_fp = fopen(config->output_file, mode);
            if (collector->output_fp == NULL) {
                free(collector);
                return NULL;
            }
            setvbuf(collector->output_fp, NULL, _IOFBF, 65536);
        }
    }
    
    if (pthread_mutex_init(&collector->collector_mutex, NULL) != 0) {
        if (collector->output_fp != NULL) {
            fclose(collector->output_fp);
        }
        free(collector);
        return NULL;
    }
    
    if (pthread_cond_init(&collector->wakeup_cond, NULL) != 0) {
        pthread_mutex_destroy(&collector->collector_mutex);
        if (collector->output_fp != NULL) {
            fclose(collector->output_fp);
        }
        free(collector);
        return NULL;
    }
    
    collector->is_running = false;
    collector->is_stopped = true;
    collector->total_processed = 0;
    collector->total_flushed = 0;
    
    return collector;
}

void log_collector_destroy(log_collector_t* collector) {
    if (collector == NULL) {
        return;
    }
    
    if (collector->is_running) {
        log_collector_stop(collector, true);
    }
    
    if (collector->output_fp != NULL) {
        fflush(collector->output_fp);
        fclose(collector->output_fp);
        collector->output_fp = NULL;
    }
    
    pthread_cond_destroy(&collector->wakeup_cond);
    pthread_mutex_destroy(&collector->collector_mutex);
    
    free(collector);
}

size_t log_collector_collect_and_sort(log_collector_t* collector,
                                        log_message_t*** out_messages,
                                        size_t* out_count) {
    if (collector == NULL || out_messages == NULL || out_count == NULL) {
        return 0;
    }
    
    *out_messages = NULL;
    *out_count = 0;
    
    log_message_t** messages = NULL;
    size_t count = 0;
    
    count = log_channel_collect_all(collector->channel, &messages, &count);
    
    if (count == 0 || messages == NULL) {
        return 0;
    }
    
    if (collector->config.sort_by_timestamp && count > 1) {
        log_message_sort_by_timestamp(messages, count);
    }
    
    *out_messages = messages;
    *out_count = count;
    
    return count;
}

int log_collector_write_messages(log_collector_t* collector,
                                   log_message_t** messages,
                                   size_t count) {
    if (collector == NULL || messages == NULL || count == 0) {
        return -1;
    }
    
    for (size_t i = 0; i < count; i++) {
        log_message_t* msg = messages[i];
        if (msg == NULL) {
            continue;
        }
        
        struct tm tm_info;
        char time_buf[64];
        time_t sec = msg->timestamp.tv_sec;
        localtime_r(&sec, &tm_info);
        strftime(time_buf, sizeof(time_buf), "%Y-%m-%d %H:%M:%S", &tm_info);
        
        char output_buf[LOG_MESSAGE_MAX_LEN + 256];
        int output_len = snprintf(output_buf, sizeof(output_buf),
            "%s.%06lu [%s] [%lu] %s\n",
            time_buf,
            (unsigned long)msg->timestamp.tv_usec,
            log_level_to_string(msg->level),
            (unsigned long)msg->thread_id,
            msg->message);
        
        if (output_len > 0) {
            if (collector->config.output_type == LOG_OUTPUT_STDOUT ||
                collector->config.output_type == LOG_OUTPUT_BOTH) {
                fwrite(output_buf, 1, (size_t)output_len, stdout);
            }
            if ((collector->config.output_type == LOG_OUTPUT_FILE ||
                 collector->config.output_type == LOG_OUTPUT_BOTH) &&
                collector->output_fp != NULL) {
                fwrite(output_buf, 1, (size_t)output_len, collector->output_fp);
            }
        }
    }
    
    collector->total_processed += count;
    
    return 0;
}

static void* collector_thread_func(void* arg) {
    log_collector_t* collector = (log_collector_t*)arg;
    if (collector == NULL) {
        return NULL;
    }
    
    collector->is_stopped = false;
    
    while (collector->is_running) {
        struct timespec timeout;
        clock_gettime(CLOCK_REALTIME, &timeout);
        timeout.tv_sec += collector->config.flush_interval_ms / 1000;
        timeout.tv_nsec += (collector->config.flush_interval_ms % 1000) * 1000000;
        if (timeout.tv_nsec >= 1000000000) {
            timeout.tv_sec++;
            timeout.tv_nsec -= 1000000000;
        }
        
        pthread_mutex_lock(&collector->collector_mutex);
        int wait_ret = 0;
        if (collector->is_running) {
            wait_ret = pthread_cond_timedwait(&collector->wakeup_cond,
                                                &collector->collector_mutex,
                                                &timeout);
        }
        pthread_mutex_unlock(&collector->collector_mutex);
        
        log_message_t** messages = NULL;
        size_t count = 0;
        
        size_t collected = log_collector_collect_and_sort(collector, &messages, &count);
        
        if (collected > 0 && messages != NULL) {
            log_collector_write_messages(collector, messages, count);
            log_message_array_destroy(messages, count);
        }
        
        if (wait_ret == 0 || log_channel_total_size(collector->channel) > 0) {
            if (collector->output_fp != NULL) {
                fflush(collector->output_fp);
                collector->total_flushed++;
            }
            fflush(stdout);
        }
    }
    
    log_message_t** messages = NULL;
    size_t count = 0;
    size_t collected = log_collector_collect_and_sort(collector, &messages, &count);
    
    if (collected > 0 && messages != NULL) {
        log_collector_write_messages(collector, messages, count);
        log_message_array_destroy(messages, count);
    }
    
    if (collector->output_fp != NULL) {
        fflush(collector->output_fp);
        collector->total_flushed++;
    }
    fflush(stdout);
    
    collector->is_stopped = true;
    return NULL;
}

int log_collector_start(log_collector_t* collector) {
    if (collector == NULL) {
        return -1;
    }
    
    if (collector->is_running) {
        return 0;
    }
    
    collector->is_running = true;
    collector->is_stopped = false;
    
    int ret = pthread_create(&collector->collector_thread, NULL,
                              collector_thread_func, collector);
    if (ret != 0) {
        collector->is_running = false;
        collector->is_stopped = true;
        return -1;
    }
    
    return 0;
}

int log_collector_stop(log_collector_t* collector, bool wait_for_drain) {
    if (collector == NULL) {
        return -1;
    }
    
    if (!collector->is_running) {
        return 0;
    }
    
    collector->is_running = false;
    
    pthread_mutex_lock(&collector->collector_mutex);
    pthread_cond_signal(&collector->wakeup_cond);
    pthread_mutex_unlock(&collector->collector_mutex);
    
    if (wait_for_drain) {
        pthread_join(collector->collector_thread, NULL);
    }
    
    return 0;
}

int log_collector_flush(log_collector_t* collector) {
    if (collector == NULL) {
        return -1;
    }
    
    log_message_t** messages = NULL;
    size_t count = 0;
    
    size_t collected = log_collector_collect_and_sort(collector, &messages, &count);
    
    if (collected > 0 && messages != NULL) {
        log_collector_write_messages(collector, messages, count);
        log_message_array_destroy(messages, count);
    }
    
    if (collector->output_fp != NULL) {
        fflush(collector->output_fp);
    }
    fflush(stdout);
    
    return 0;
}

uint64_t log_collector_get_stats(log_collector_t* collector,
                                   uint64_t* out_flushed) {
    if (collector == NULL) {
        return 0;
    }
    
    if (out_flushed != NULL) {
        *out_flushed = collector->total_flushed;
    }
    
    return collector->total_processed;
}
