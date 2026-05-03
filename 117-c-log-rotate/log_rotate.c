#include "log_rotate.h"
#include "backup_manager.h"
#include "async_compress.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <sys/stat.h>
#include <unistd.h>
#include <errno.h>
#include <time.h>

static async_compress_t g_compressor;
static int g_compressor_initialized = 0;
static pthread_mutex_t g_compressor_mutex = PTHREAD_MUTEX_INITIALIZER;

static void ensure_compressor_init(void)
{
    pthread_mutex_lock(&g_compressor_mutex);
    if (!g_compressor_initialized) {
        async_compress_init(&g_compressor);
        g_compressor_initialized = 1;
    }
    pthread_mutex_unlock(&g_compressor_mutex);
}

static off_t get_file_size(FILE *fp)
{
    if (fp == NULL) {
        return -1;
    }
    int fd = fileno(fp);
    if (fd < 0) {
        return -1;
    }
    struct stat st;
    if (fstat(fd, &st) != 0) {
        return -1;
    }
    return st.st_size;
}

static time_t get_current_date(void)
{
    time_t now = time(NULL);
    struct tm *tm_info = localtime(&now);
    if (tm_info == NULL) {
        return now;
    }
    tm_info->tm_hour = 0;
    tm_info->tm_min = 0;
    tm_info->tm_sec = 0;
    return mktime(tm_info);
}

static int needs_rotate_by_size(log_rotate_t *lr)
{
    if (!(lr->config.mode & LOG_ROTATE_MODE_SIZE)) {
        return 0;
    }
    off_t size = get_file_size(lr->fp);
    if (size < 0) {
        return 0;
    }
    return (size_t)size >= lr->config.max_size;
}

static int needs_rotate_by_time(log_rotate_t *lr)
{
    if (!(lr->config.mode & LOG_ROTATE_MODE_TIME)) {
        return 0;
    }
    time_t current_date = get_current_date();
    if (current_date != lr->last_rotate_date) {
        return 1;
    }
    return 0;
}

static int perform_rotate(log_rotate_t *lr)
{
    if (lr->fp == NULL) {
        return -1;
    }
    
    fflush(lr->fp);
    fclose(lr->fp);
    lr->fp = NULL;
    
    int rotate_by_time = 0;
    if (lr->config.mode & LOG_ROTATE_MODE_TIME) {
        time_t current_date = get_current_date();
        if (current_date != lr->last_rotate_date) {
            rotate_by_time = 1;
            lr->last_rotate_date = current_date;
        }
    }
    
    char backup_path[LOG_ROTATE_MAX_PATH];
    int backup_created = 0;
    
    if (lr->config.max_backups > 0) {
        if (rotate_by_time && (lr->config.mode & LOG_ROTATE_MODE_TIME)) {
            if (backup_create_with_date(lr->config.base_path, NULL) == 0) {
                backup_created = 1;
            }
        } else {
            backup_shift(lr->config.base_path, lr->config.max_backups);
            
            if (backup_generate_filename(backup_path, sizeof(backup_path), 
                                         lr->config.base_path, 1) == 0) {
                if (rename(lr->config.base_path, backup_path) == 0) {
                    backup_created = 1;
                }
            }
        }
        
        if (backup_created && lr->config.compress_enabled) {
            ensure_compressor_init();
            
            char compress_src[LOG_ROTATE_MAX_PATH];
            char compress_dst[LOG_ROTATE_MAX_PATH];
            
            if (rotate_by_time && (lr->config.mode & LOG_ROTATE_MODE_TIME)) {
                struct tm *tm_info = localtime(&lr->last_rotate_date);
                if (tm_info != NULL) {
                    char date_str[32];
                    strftime(date_str, sizeof(date_str), "%Y-%m-%d", tm_info);
                    snprintf(compress_src, sizeof(compress_src), "%s.%s", 
                             lr->config.base_path, date_str);
                    snprintf(compress_dst, sizeof(compress_dst), "%s.%s.gz", 
                             lr->config.base_path, date_str);
                } else {
                    strncpy(compress_src, backup_path, sizeof(compress_src) - 1);
                    snprintf(compress_dst, sizeof(compress_dst), "%s.gz", backup_path);
                }
            } else {
                strncpy(compress_src, backup_path, sizeof(compress_src) - 1);
                snprintf(compress_dst, sizeof(compress_dst), "%s.gz", backup_path);
            }
            
            async_compress_submit(&g_compressor, compress_src, compress_dst);
        }
        
        backup_cleanup(lr->config.base_path, lr->config.max_backups);
    }
    
    lr->fp = fopen(lr->config.base_path, "a");
    if (lr->fp == NULL) {
        return -1;
    }
    
    return 0;
}

int log_rotate_init(log_rotate_t *lr, const log_rotate_config_t *config)
{
    if (lr == NULL || config == NULL) {
        return -1;
    }
    
    memset(lr, 0, sizeof(log_rotate_t));
    
    if (strlen(config->base_path) >= LOG_ROTATE_MAX_PATH) {
        return -1;
    }
    
    lr->config = *config;
    
    if (pthread_mutex_init(&lr->mutex, NULL) != 0) {
        return -1;
    }
    
    lr->fp = fopen(config->base_path, "a");
    if (lr->fp == NULL) {
        pthread_mutex_destroy(&lr->mutex);
        return -1;
    }
    
    lr->last_rotate_date = get_current_date();
    lr->initialized = 1;
    
    return 0;
}

void log_rotate_destroy(log_rotate_t *lr)
{
    if (lr == NULL || !lr->initialized) {
        return;
    }
    
    pthread_mutex_lock(&lr->mutex);
    
    if (lr->fp != NULL) {
        fflush(lr->fp);
        fclose(lr->fp);
        lr->fp = NULL;
    }
    
    pthread_mutex_unlock(&lr->mutex);
    pthread_mutex_destroy(&lr->mutex);
    
    lr->initialized = 0;
}

int log_rotate_write(log_rotate_t *lr, const char *format, ...)
{
    if (lr == NULL || format == NULL) {
        return -1;
    }
    
    if (!lr->initialized || lr->fp == NULL) {
        return -1;
    }
    
    pthread_mutex_lock(&lr->mutex);
    
    if (needs_rotate_by_size(lr) || needs_rotate_by_time(lr)) {
        if (perform_rotate(lr) != 0) {
            pthread_mutex_unlock(&lr->mutex);
            return -1;
        }
    }
    
    if (lr->fp == NULL) {
        pthread_mutex_unlock(&lr->mutex);
        return -1;
    }
    
    va_list args;
    va_start(args, format);
    int written = vfprintf(lr->fp, format, args);
    va_end(args);
    
    if (written >= 0) {
        fflush(lr->fp);
    }
    
    pthread_mutex_unlock(&lr->mutex);
    
    return written;
}

int log_rotate_flush(log_rotate_t *lr)
{
    if (lr == NULL || !lr->initialized) {
        return -1;
    }
    
    pthread_mutex_lock(&lr->mutex);
    
    if (lr->fp != NULL) {
        fflush(lr->fp);
    }
    
    pthread_mutex_unlock(&lr->mutex);
    
    return 0;
}

void log_rotate_cleanup_compressor(void)
{
    pthread_mutex_lock(&g_compressor_mutex);
    if (g_compressor_initialized) {
        async_compress_destroy(&g_compressor);
        g_compressor_initialized = 0;
    }
    pthread_mutex_unlock(&g_compressor_mutex);
}
