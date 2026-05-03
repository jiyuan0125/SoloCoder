#include "backup_manager.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <dirent.h>
#include <unistd.h>
#include <errno.h>

typedef enum {
    FILE_TYPE_UNCOMPRESSED = 0,
    FILE_TYPE_COMPRESSED = 1
} file_type_t;

int backup_generate_filename(char *buf, size_t buf_size, const char *base_path, int index)
{
    if (buf == NULL || base_path == NULL) {
        return -1;
    }
    if (index == 0) {
        if (strlen(base_path) >= buf_size) {
            return -1;
        }
        strncpy(buf, base_path, buf_size - 1);
        buf[buf_size - 1] = '\0';
        return 0;
    }
    int written = snprintf(buf, buf_size, "%s.%d", base_path, index);
    if (written < 0 || (size_t)written >= buf_size) {
        return -1;
    }
    return 0;
}

int backup_generate_date_filename(char *buf, size_t buf_size, const char *base_path, time_t date)
{
    if (buf == NULL || base_path == NULL) {
        return -1;
    }
    struct tm *tm_info = localtime(&date);
    if (tm_info == NULL) {
        return -1;
    }
    char date_str[32];
    strftime(date_str, sizeof(date_str), "%Y-%m-%d", tm_info);
    
    int written = snprintf(buf, buf_size, "%s.%s", base_path, date_str);
    if (written < 0 || (size_t)written >= buf_size) {
        return -1;
    }
    return 0;
}

static int file_exists(const char *path)
{
    struct stat st;
    return (stat(path, &st) == 0);
}

static int file_exists_any(const char *base_name)
{
    char path[BACKUP_MAX_PATH];
    
    snprintf(path, sizeof(path), "%s", base_name);
    if (file_exists(path)) {
        return 1;
    }
    
    snprintf(path, sizeof(path), "%s.gz", base_name);
    if (file_exists(path)) {
        return 1;
    }
    
    return 0;
}

static off_t file_size(const char *path)
{
    struct stat st;
    if (stat(path, &st) != 0) {
        return -1;
    }
    return st.st_size;
}

int backup_shift(const char *base_path, int max_backups)
{
    if (base_path == NULL || max_backups < 0) {
        return -1;
    }
    
    if (max_backups == 0) {
        return 0;
    }
    
    char src_base[BACKUP_MAX_PATH];
    char dst_base[BACKUP_MAX_PATH];
    char src_path[BACKUP_MAX_PATH];
    char dst_path[BACKUP_MAX_PATH];
    int i;
    
    for (i = max_backups; i >= 1; i--) {
        if (backup_generate_filename(src_base, sizeof(src_base), base_path, i) != 0) {
            continue;
        }
        if (backup_generate_filename(dst_base, sizeof(dst_base), base_path, i + 1) != 0) {
            continue;
        }
        
        snprintf(src_path, sizeof(src_path), "%s.gz", src_base);
        if (file_exists(src_path)) {
            if (i >= max_backups) {
                unlink(src_path);
            } else {
                snprintf(dst_path, sizeof(dst_path), "%s.gz", dst_base);
                if (file_exists(dst_path)) {
                    unlink(dst_path);
                }
                rename(src_path, dst_path);
            }
        }
        
        snprintf(src_path, sizeof(src_path), "%s", src_base);
        if (file_exists(src_path)) {
            if (i >= max_backups) {
                unlink(src_path);
            } else {
                snprintf(dst_path, sizeof(dst_path), "%s", dst_base);
                if (file_exists(dst_path)) {
                    unlink(dst_path);
                }
                rename(src_path, dst_path);
            }
        }
    }
    
    return 0;
}

int backup_create_with_date(const char *base_path, time_t *date)
{
    if (base_path == NULL) {
        return -1;
    }
    
    char date_path[BACKUP_MAX_PATH];
    time_t t = date ? *date : time(NULL);
    
    if (backup_generate_date_filename(date_path, sizeof(date_path), base_path, t) != 0) {
        return -1;
    }
    
    int suffix = 0;
    char final_path[BACKUP_MAX_PATH];
    
    while (1) {
        if (suffix == 0) {
            strncpy(final_path, date_path, sizeof(final_path) - 1);
            final_path[sizeof(final_path) - 1] = '\0';
        } else {
            int written = snprintf(final_path, sizeof(final_path), "%s.%d", date_path, suffix);
            if (written < 0 || (size_t)written >= sizeof(final_path)) {
                return -1;
            }
        }
        
        if (!file_exists_any(final_path)) {
            break;
        }
        suffix++;
    }
    
    if (rename(base_path, final_path) != 0) {
        return -1;
    }
    
    return 0;
}

static void parse_backup_index(const char *name, const char *base_name, int *is_compressed, int *index)
{
    size_t base_len = strlen(base_name);
    size_t name_len = strlen(name);
    
    *is_compressed = 0;
    *index = -1;
    
    if (name_len <= base_len + 1) {
        return;
    }
    
    if (strncmp(name, base_name, base_len) != 0) {
        return;
    }
    
    if (name[base_len] != '.') {
        return;
    }
    
    const char *suffix = name + base_len + 1;
    size_t suffix_len = strlen(suffix);
    
    if (suffix_len >= 3 && strcmp(suffix + suffix_len - 3, ".gz") == 0) {
        *is_compressed = 1;
        suffix_len -= 3;
    }
    
    if (suffix_len == 0) {
        return;
    }
    
    int is_numeric = 1;
    int value = 0;
    for (size_t i = 0; i < suffix_len; i++) {
        if (suffix[i] < '0' || suffix[i] > '9') {
            is_numeric = 0;
            break;
        }
        value = value * 10 + (suffix[i] - '0');
    }
    
    if (is_numeric) {
        *index = value;
    }
}

int backup_get_count(const char *base_path)
{
    if (base_path == NULL) {
        return -1;
    }
    
    char dir_path[BACKUP_MAX_PATH];
    char *last_slash = strrchr(base_path, '/');
    const char *file_name;
    
    if (last_slash != NULL) {
        size_t dir_len = last_slash - base_path;
        if (dir_len >= sizeof(dir_path) - 1) {
            return -1;
        }
        strncpy(dir_path, base_path, dir_len);
        dir_path[dir_len] = '\0';
        file_name = last_slash + 1;
    } else {
        strncpy(dir_path, ".", sizeof(dir_path) - 1);
        dir_path[sizeof(dir_path) - 1] = '\0';
        file_name = base_path;
    }
    
    DIR *dir = opendir(dir_path);
    if (dir == NULL) {
        return 0;
    }
    
    int max_index = -1;
    struct dirent *entry;
    
    while ((entry = readdir(dir)) != NULL) {
        const char *name = entry->d_name;
        int is_compressed, index;
        
        parse_backup_index(name, file_name, &is_compressed, &index);
        
        if (index > max_index) {
            max_index = index;
        }
    }
    
    closedir(dir);
    return (max_index >= 0) ? max_index : 0;
}

int backup_cleanup(const char *base_path, int max_backups)
{
    if (base_path == NULL || max_backups < 0) {
        return -1;
    }
    
    if (max_backups == 0) {
        return 0;
    }
    
    char path[BACKUP_MAX_PATH];
    int max_index = backup_get_count(base_path);
    
    while (max_index > max_backups) {
        if (backup_generate_filename(path, sizeof(path), base_path, max_index) != 0) {
            break;
        }
        
        char gz_path[BACKUP_MAX_PATH];
        snprintf(gz_path, sizeof(gz_path), "%s.gz", path);
        
        unlink(gz_path);
        unlink(path);
        
        max_index--;
    }
    
    return 0;
}
