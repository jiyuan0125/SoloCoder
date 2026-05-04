#include "log_channel.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <stdarg.h>
#include <errno.h>

static void log_message_destructor(void* data) {
    if (data != NULL) {
        log_message_destroy((log_message_t*)data);
    }
}

log_message_t* log_message_create(log_level_t level, const char* format, ...) {
    log_message_t* msg = (log_message_t*)malloc(sizeof(log_message_t));
    if (msg == NULL) {
        return NULL;
    }
    
    gettimeofday(&msg->timestamp, NULL);
    msg->thread_id = pthread_self();
    msg->level = level;
    
    va_list args;
    va_start(args, format);
    int len = vsnprintf(msg->message, LOG_MESSAGE_MAX_LEN - 1, format, args);
    va_end(args);
    
    if (len < 0) {
        free(msg);
        return NULL;
    }
    
    msg->message_len = (size_t)len;
    if (msg->message_len >= LOG_MESSAGE_MAX_LEN) {
        msg->message_len = LOG_MESSAGE_MAX_LEN - 1;
    }
    msg->message[msg->message_len] = '\0';
    
    return msg;
}

void log_message_destroy(log_message_t* msg) {
    if (msg != NULL) {
        free(msg);
    }
}

void log_message_array_destroy(log_message_t** array, size_t count) {
    if (array == NULL) {
        return;
    }
    for (size_t i = 0; i < count; i++) {
        if (array[i] != NULL) {
            log_message_destroy(array[i]);
        }
    }
    free(array);
}

log_channel_t* log_channel_create(const log_channel_config_t* config) {
    if (config == NULL) {
        return NULL;
    }
    
    log_channel_t* channel = (log_channel_t*)malloc(sizeof(log_channel_t));
    if (channel == NULL) {
        return NULL;
    }
    
    memset(channel, 0, sizeof(log_channel_t));
    
    channel->config = *config;
    if (channel->config.max_threads > LOG_MAX_THREADS) {
        channel->config.max_threads = LOG_MAX_THREADS;
    }
    if (channel->config.max_thread_messages == 0) {
        channel->config.max_thread_messages = 1024;
    }
    if (channel->config.max_total_messages == 0) {
        channel->config.max_total_messages = 65536;
    }
    
    for (uint32_t i = 0; i < LOG_MAX_THREADS; i++) {
        channel->writers[i].queue = NULL;
        channel->writers[i].is_active = false;
    }
    
    if (pthread_rwlock_init(&channel->writers_lock, NULL) != 0) {
        free(channel);
        return NULL;
    }
    
    if (pthread_mutex_init(&channel->channel_mutex, NULL) != 0) {
        pthread_rwlock_destroy(&channel->writers_lock);
        free(channel);
        return NULL;
    }
    
    if (pthread_cond_init(&channel->space_available, NULL) != 0) {
        pthread_mutex_destroy(&channel->channel_mutex);
        pthread_rwlock_destroy(&channel->writers_lock);
        free(channel);
        return NULL;
    }
    
    if (pthread_spin_init(&channel->stats_lock, PTHREAD_PROCESS_PRIVATE) != 0) {
        pthread_cond_destroy(&channel->space_available);
        pthread_mutex_destroy(&channel->channel_mutex);
        pthread_rwlock_destroy(&channel->writers_lock);
        free(channel);
        return NULL;
    }
    
    channel->is_shutdown = false;
    channel->writer_count = 0;
    channel->total_dropped = 0;
    channel->total_written = 0;
    
    return channel;
}

void log_channel_destroy(log_channel_t* channel) {
    if (channel == NULL) {
        return;
    }
    
    log_channel_shutdown(channel);
    
    pthread_rwlock_wrlock(&channel->writers_lock);
    for (uint32_t i = 0; i < LOG_MAX_THREADS; i++) {
        if (channel->writers[i].queue != NULL) {
            lf_deque_destroy(channel->writers[i].queue, log_message_destructor);
            channel->writers[i].queue = NULL;
        }
    }
    pthread_rwlock_unlock(&channel->writers_lock);
    
    pthread_spin_destroy(&channel->stats_lock);
    pthread_cond_destroy(&channel->space_available);
    pthread_mutex_destroy(&channel->channel_mutex);
    pthread_rwlock_destroy(&channel->writers_lock);
    
    free(channel);
}

log_writer_t* log_channel_register_writer(log_channel_t* channel) {
    if (channel == NULL) {
        return NULL;
    }
    
    pthread_rwlock_wrlock(&channel->writers_lock);
    
    if (channel->writer_count >= channel->config.max_threads) {
        pthread_rwlock_unlock(&channel->writers_lock);
        return NULL;
    }
    
    uint32_t slot = (uint32_t)-1;
    for (uint32_t i = 0; i < LOG_MAX_THREADS; i++) {
        if (!channel->writers[i].is_active && channel->writers[i].queue == NULL) {
            slot = i;
            break;
        }
    }
    
    if (slot == (uint32_t)-1) {
        pthread_rwlock_unlock(&channel->writers_lock);
        return NULL;
    }
    
    lf_deque_t* queue = lf_deque_create(channel->config.max_thread_messages);
    if (queue == NULL) {
        pthread_rwlock_unlock(&channel->writers_lock);
        return NULL;
    }
    
    channel->writers[slot].queue = queue;
    channel->writers[slot].thread_id = pthread_self();
    channel->writers[slot].is_active = true;
    channel->writer_count++;
    
    log_writer_t* writer = &channel->writers[slot];
    pthread_rwlock_unlock(&channel->writers_lock);
    
    return writer;
}

void log_channel_unregister_writer(log_channel_t* channel, log_writer_t* writer) {
    if (channel == NULL || writer == NULL) {
        return;
    }
    
    pthread_rwlock_wrlock(&channel->writers_lock);
    
    for (uint32_t i = 0; i < LOG_MAX_THREADS; i++) {
        if (&channel->writers[i] == writer) {
            writer->is_active = false;
            writer->thread_id = 0;
            channel->writer_count--;
            break;
        }
    }
    
    pthread_rwlock_unlock(&channel->writers_lock);
    
    pthread_mutex_lock(&channel->channel_mutex);
    pthread_cond_broadcast(&channel->space_available);
    pthread_mutex_unlock(&channel->channel_mutex);
}

size_t log_channel_total_size(log_channel_t* channel) {
    if (channel == NULL) {
        return 0;
    }
    
    size_t total = 0;
    pthread_rwlock_rdlock(&channel->writers_lock);
    
    for (uint32_t i = 0; i < LOG_MAX_THREADS; i++) {
        if (channel->writers[i].queue != NULL) {
            total += lf_deque_size(channel->writers[i].queue);
        }
    }
    
    pthread_rwlock_unlock(&channel->writers_lock);
    return total;
}

static int log_channel_try_write(log_channel_t* channel, log_writer_t* writer,
                                  log_message_t* msg) {
    if (channel == NULL || writer == NULL || msg == NULL) {
        return -1;
    }
    
    if (channel->is_shutdown) {
        return -2;
    }
    
    size_t total_size = log_channel_total_size(channel);
    
    if (channel->config.max_total_messages > 0 && 
        total_size >= channel->config.max_total_messages) {
        
        switch (channel->config.overflow_policy) {
            case LOG_POLICY_BLOCK:
                return -3;
                
            case LOG_POLICY_DROP_NEWEST:
                pthread_spin_lock(&channel->stats_lock);
                channel->total_dropped++;
                pthread_spin_unlock(&channel->stats_lock);
                return -4;
                
            case LOG_POLICY_DROP_OLDEST:
                log_channel_drop_half_oldest(channel);
                break;
        }
    }
    
    lf_deque_status_t status = lf_deque_push_back(writer->queue, msg);
    
    if (status == LF_DEQUE_FULL) {
        switch (channel->config.overflow_policy) {
            case LOG_POLICY_BLOCK:
                return -3;
                
            case LOG_POLICY_DROP_NEWEST:
                pthread_spin_lock(&channel->stats_lock);
                channel->total_dropped++;
                pthread_spin_unlock(&channel->stats_lock);
                return -4;
                
            case LOG_POLICY_DROP_OLDEST:
                lf_deque_drop_oldest(writer->queue, 
                    (lf_deque_size(writer->queue) + 1) / 2, 
                    log_message_destructor);
                status = lf_deque_push_back(writer->queue, msg);
                if (status != LF_DEQUE_OK) {
                    pthread_spin_lock(&channel->stats_lock);
                    channel->total_dropped++;
                    pthread_spin_unlock(&channel->stats_lock);
                    return -5;
                }
                break;
        }
    }
    
    if (status == LF_DEQUE_OK) {
        pthread_spin_lock(&channel->stats_lock);
        channel->total_written++;
        pthread_spin_unlock(&channel->stats_lock);
        return 0;
    }
    
    return -6;
}

int log_channel_write(log_channel_t* channel, log_writer_t* writer,
                       log_level_t level, const char* format, ...) {
    if (channel == NULL || writer == NULL || format == NULL) {
        return -1;
    }
    
    va_list args;
    va_start(args, format);
    
    log_message_t* msg = (log_message_t*)malloc(sizeof(log_message_t));
    if (msg == NULL) {
        va_end(args);
        return -1;
    }
    
    gettimeofday(&msg->timestamp, NULL);
    msg->thread_id = pthread_self();
    msg->level = level;
    
    int len = vsnprintf(msg->message, LOG_MESSAGE_MAX_LEN - 1, format, args);
    va_end(args);
    
    if (len < 0) {
        free(msg);
        return -1;
    }
    
    msg->message_len = (size_t)len;
    if (msg->message_len >= LOG_MESSAGE_MAX_LEN) {
        msg->message_len = LOG_MESSAGE_MAX_LEN - 1;
    }
    msg->message[msg->message_len] = '\0';
    
    int ret;
    
    if (channel->config.overflow_policy == LOG_POLICY_BLOCK) {
        pthread_mutex_lock(&channel->channel_mutex);
        
        while (!channel->is_shutdown) {
            ret = log_channel_try_write(channel, writer, msg);
            if (ret != -3) {
                pthread_mutex_unlock(&channel->channel_mutex);
                if (ret != 0) {
                    free(msg);
                }
                return ret;
            }
            
            pthread_cond_wait(&channel->space_available, &channel->channel_mutex);
        }
        
        pthread_mutex_unlock(&channel->channel_mutex);
        free(msg);
        return -2;
    } else {
        ret = log_channel_try_write(channel, writer, msg);
        if (ret != 0) {
            free(msg);
        }
        return ret;
    }
}

int log_channel_write_ex(log_channel_t* channel, log_writer_t* writer,
                          log_message_t* msg) {
    if (channel == NULL || writer == NULL || msg == NULL) {
        return -1;
    }
    
    int ret;
    
    if (channel->config.overflow_policy == LOG_POLICY_BLOCK) {
        pthread_mutex_lock(&channel->channel_mutex);
        
        while (!channel->is_shutdown) {
            ret = log_channel_try_write(channel, writer, msg);
            if (ret != -3) {
                pthread_mutex_unlock(&channel->channel_mutex);
                if (ret != 0) {
                    free(msg);
                }
                return ret;
            }
            
            pthread_cond_wait(&channel->space_available, &channel->channel_mutex);
        }
        
        pthread_mutex_unlock(&channel->channel_mutex);
        free(msg);
        return -2;
    } else {
        ret = log_channel_try_write(channel, writer, msg);
        if (ret != 0) {
            free(msg);
        }
        return ret;
    }
}

size_t log_channel_collect_all(log_channel_t* channel, log_message_t*** out_messages,
                                size_t* out_count) {
    if (channel == NULL || out_messages == NULL || out_count == NULL) {
        return 0;
    }
    
    *out_messages = NULL;
    *out_count = 0;
    
    size_t total_count = 0;
    log_message_t** all_messages = NULL;
    size_t alloc_size = 0;
    
    pthread_rwlock_wrlock(&channel->writers_lock);
    
    for (uint32_t i = 0; i < LOG_MAX_THREADS; i++) {
        if (channel->writers[i].queue != NULL) {
            void** batch = NULL;
            size_t batch_count = 0;
            
            size_t drained = lf_deque_drain(channel->writers[i].queue, &batch, &batch_count);
            
            if (drained > 0 && batch != NULL) {
                size_t new_size = total_count + drained;
                if (new_size > alloc_size) {
                    size_t new_alloc = alloc_size == 0 ? 1024 : alloc_size * 2;
                    while (new_alloc < new_size) {
                        new_alloc *= 2;
                    }
                    log_message_t** new_array = (log_message_t**)realloc(
                        all_messages, new_alloc * sizeof(log_message_t*));
                    if (new_array == NULL) {
                        for (size_t j = 0; j < batch_count; j++) {
                            log_message_destroy((log_message_t*)batch[j]);
                        }
                        free(batch);
                        continue;
                    }
                    all_messages = new_array;
                    alloc_size = new_alloc;
                }
                
                for (size_t j = 0; j < batch_count; j++) {
                    all_messages[total_count++] = (log_message_t*)batch[j];
                }
                free(batch);
            }
            
            if (!channel->writers[i].is_active && 
                lf_deque_size(channel->writers[i].queue) == 0) {
                lf_deque_destroy(channel->writers[i].queue, log_message_destructor);
                channel->writers[i].queue = NULL;
            }
        }
    }
    
    pthread_rwlock_unlock(&channel->writers_lock);
    
    pthread_mutex_lock(&channel->channel_mutex);
    pthread_cond_broadcast(&channel->space_available);
    pthread_mutex_unlock(&channel->channel_mutex);
    
    if (total_count > 0) {
        *out_messages = all_messages;
        *out_count = total_count;
    } else if (all_messages != NULL) {
        free(all_messages);
    }
    
    return total_count;
}

void log_channel_drop_half_oldest(log_channel_t* channel) {
    if (channel == NULL) {
        return;
    }
    
    size_t total_size = log_channel_total_size(channel);
    if (total_size == 0) {
        return;
    }
    
    size_t to_drop_total = (total_size + 1) / 2;
    size_t dropped_so_far = 0;
    
    pthread_rwlock_rdlock(&channel->writers_lock);
    
    for (uint32_t i = 0; i < LOG_MAX_THREADS && dropped_so_far < to_drop_total; i++) {
        if (channel->writers[i].queue != NULL) {
            size_t queue_size = lf_deque_size(channel->writers[i].queue);
            if (queue_size > 0) {
                size_t queue_drop = (queue_size + 1) / 2;
                if (dropped_so_far + queue_drop > to_drop_total) {
                    queue_drop = to_drop_total - dropped_so_far;
                }
                
                lf_deque_drop_oldest(channel->writers[i].queue, queue_drop, 
                                     log_message_destructor);
                dropped_so_far += queue_drop;
                
                pthread_spin_lock(&channel->stats_lock);
                channel->total_dropped += queue_drop;
                pthread_spin_unlock(&channel->stats_lock);
            }
        }
    }
    
    pthread_rwlock_unlock(&channel->writers_lock);
    
    pthread_mutex_lock(&channel->channel_mutex);
    pthread_cond_broadcast(&channel->space_available);
    pthread_mutex_unlock(&channel->channel_mutex);
}

bool log_channel_is_shutdown(log_channel_t* channel) {
    if (channel == NULL) {
        return true;
    }
    return channel->is_shutdown;
}

void log_channel_shutdown(log_channel_t* channel) {
    if (channel == NULL) {
        return;
    }
    
    channel->is_shutdown = true;
    
    pthread_mutex_lock(&channel->channel_mutex);
    pthread_cond_broadcast(&channel->space_available);
    pthread_mutex_unlock(&channel->channel_mutex);
}

size_t log_channel_get_stats(log_channel_t* channel, size_t* dropped, size_t* written) {
    if (channel == NULL) {
        return 0;
    }
    
    pthread_spin_lock(&channel->stats_lock);
    if (dropped != NULL) {
        *dropped = channel->total_dropped;
    }
    if (written != NULL) {
        *written = channel->total_written;
    }
    pthread_spin_unlock(&channel->stats_lock);
    
    size_t current = log_channel_total_size(channel);
    return current;
}
