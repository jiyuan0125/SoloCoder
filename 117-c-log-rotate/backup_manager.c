#include "backup_manager.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <dirent.h>
#include <unistd.h>
#include <errno.h>
#include <time.h>

typedef enum {
    BACKUP_TYPE_UNKNOWN = 0,
    BACKUP_TYPE_NUMERIC,
    BACKUP_TYPE_DATE,
    BACKUP_TYPE_TEMP,
    BACKUP_TYPE_HARDLINK
} backup_type_t;

typedef struct {
    char name[BACKUP_MAX_PATH];
    backup_type_t type;
    int numeric_index;
    time_t date_value;
    time_t mtime;
    int is_compressed;
} backup_file_t;

static int g_file_count = 0;
static backup_file_t *g_files = NULL;
static int g_files_capacity = 0;

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

static int is_date_format(const char *str)
{
    if (strlen(str) != 10) return 0;
    if (str[4] != '-' || str[7] != '-') return 0;
    for (int i = 0; i < 4; i++) if (str[i] < '0' || str[i] > '9') return 0;
    for (int i = 5; i < 7; i++) if (str[i] < '0' || str[i] > '9') return 0;
    for (int i = 8; i < 10; i++) if (str[i] < '0' || str[i] > '9') return 0;
    return 1;
}

static time_t parse_date(const char *str)
{
    struct tm tm;
    memset(&tm, 0, sizeof(tm));
    if (sscanf(str, "%d-%d-%d", &tm.tm_year, &tm.tm_mon, &tm.tm_mday) != 3) {
        return 0;
    }
    tm.tm_year -= 1900;
    tm.tm_mon -= 1;
    tm.tm_hour = 0;
    tm.tm_min = 0;
    tm.tm_sec = 0;
    return mktime(&tm);
}

static int is_numeric_suffix(const char *str, size_t len)
{
    if (len == 0) return 0;
    for (size_t i = 0; i < len; i++) {
        if (str[i] < '0' || str[i] > '9') return 0;
    }
    return 1;
}

static int parse_backup_filename(const char *name, const char *base_name, 
                                 backup_type_t *type, int *num_idx, time_t *date_val, int *is_compressed)
{
    size_t base_len = strlen(base_name);
    size_t name_len = strlen(name);
    
    *type = BACKUP_TYPE_UNKNOWN;
    *num_idx = -1;
    *date_val = 0;
    *is_compressed = 0;
    
    if (name_len <= base_len + 1) return 0;
    if (strncmp(name, base_name, base_len) != 0) return 0;
    if (name[base_len] != '.') return 0;
    
    const char *suffix = name + base_len + 1;
    size_t suffix_len = strlen(suffix);
    
    if (suffix_len >= 3 && strcmp(suffix + suffix_len - 3, ".gz") == 0) {
        *is_compressed = 1;
        suffix_len -= 3;
    }
    
    if (suffix_len == 0) return 0;
    
    const char *underscore = strchr(suffix, '_');
    const char *dot = strchr(suffix, '.');
    
    if (suffix_len >= 4 && strncmp(suffix, "tmp.", 4) == 0) {
        *type = BACKUP_TYPE_TEMP;
        return 1;
    }
    
    if (underscore != NULL && (size_t)(underscore - suffix) < suffix_len) {
        if (!(*is_compressed)) {
            *type = BACKUP_TYPE_HARDLINK;
            return 1;
        }
    }
    
    if (dot != NULL && !(*is_compressed)) {
        size_t first_part_len = dot - suffix;
        if (first_part_len > 0 && is_numeric_suffix(suffix, first_part_len)) {
            const char *second_part = dot + 1;
            size_t second_part_len = suffix_len - first_part_len - 1;
            if (second_part_len > 0 && !is_numeric_suffix(second_part, second_part_len)) {
                *type = BACKUP_TYPE_HARDLINK;
                return 1;
            }
        }
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
        *type = BACKUP_TYPE_NUMERIC;
        *num_idx = value;
        return 1;
    }
    
    if (is_date_format(suffix)) {
        *type = BACKUP_TYPE_DATE;
        *date_val = parse_date(suffix);
        return 1;
    }
    
    return 0;
}

static void ensure_files_capacity(int needed)
{
    if (g_files_capacity < needed) {
        int new_cap = (needed > 16) ? needed : 16;
        g_files = (backup_file_t *)realloc(g_files, new_cap * sizeof(backup_file_t));
        g_files_capacity = new_cap;
    }
}

static void collect_backup_files(const char *base_path)
{
    g_file_count = 0;
    
    char dir_path[BACKUP_MAX_PATH];
    const char *file_name;
    char *last_slash = strrchr(base_path, '/');
    
    if (last_slash != NULL) {
        size_t dir_len = last_slash - base_path;
        if (dir_len >= sizeof(dir_path) - 1) return;
        strncpy(dir_path, base_path, dir_len);
        dir_path[dir_len] = '\0';
        file_name = last_slash + 1;
    } else {
        strncpy(dir_path, ".", sizeof(dir_path) - 1);
        dir_path[sizeof(dir_path) - 1] = '\0';
        file_name = base_path;
    }
    
    DIR *dir = opendir(dir_path);
    if (dir == NULL) return;
    
    struct dirent *entry;
    while ((entry = readdir(dir)) != NULL) {
        const char *name = entry->d_name;
        backup_type_t type;
        int num_idx;
        time_t date_val;
        int is_compressed;
        
        if (parse_backup_filename(name, file_name, &type, &num_idx, &date_val, &is_compressed)) {
            if (type == BACKUP_TYPE_TEMP || type == BACKUP_TYPE_HARDLINK) continue;
            
            ensure_files_capacity(g_file_count + 1);
            
            strncpy(g_files[g_file_count].name, name, BACKUP_MAX_PATH - 1);
            g_files[g_file_count].name[BACKUP_MAX_PATH - 1] = '\0';
            g_files[g_file_count].type = type;
            g_files[g_file_count].numeric_index = num_idx;
            g_files[g_file_count].date_value = date_val;
            g_files[g_file_count].is_compressed = is_compressed;
            
            char full_path[BACKUP_MAX_PATH * 2];
            snprintf(full_path, sizeof(full_path), "%s/%s", dir_path, name);
            struct stat st;
            if (stat(full_path, &st) == 0) {
                g_files[g_file_count].mtime = st.st_mtime;
            } else {
                g_files[g_file_count].mtime = 0;
            }
            
            g_file_count++;
        }
    }
    
    closedir(dir);
}

int backup_cleanup_temp_files(const char *base_path)
{
    if (base_path == NULL) return -1;
    
    char dir_path[BACKUP_MAX_PATH];
    const char *file_name;
    char *last_slash = strrchr(base_path, '/');
    
    if (last_slash != NULL) {
        size_t dir_len = last_slash - base_path;
        if (dir_len >= sizeof(dir_path) - 1) return -1;
        strncpy(dir_path, base_path, dir_len);
        dir_path[dir_len] = '\0';
        file_name = last_slash + 1;
    } else {
        strncpy(dir_path, ".", sizeof(dir_path) - 1);
        dir_path[sizeof(dir_path) - 1] = '\0';
        file_name = base_path;
    }
    
    DIR *dir = opendir(dir_path);
    if (dir == NULL) return -1;
    
    struct dirent *entry;
    while ((entry = readdir(dir)) != NULL) {
        const char *name = entry->d_name;
        backup_type_t type;
        int num_idx;
        time_t date_val;
        int is_compressed;
        
        if (parse_backup_filename(name, file_name, &type, &num_idx, &date_val, &is_compressed)) {
            if (type == BACKUP_TYPE_TEMP || type == BACKUP_TYPE_HARDLINK) {
                char full_path[BACKUP_MAX_PATH * 2];
                snprintf(full_path, sizeof(full_path), "%s/%s", dir_path, name);
                unlink(full_path);
            }
        }
    }
    
    closedir(dir);
    return 0;
}

static int compare_backups_by_age(const void *a, const void *b)
{
    const backup_file_t *f1 = (const backup_file_t *)a;
    const backup_file_t *f2 = (const backup_file_t *)b;
    
    if (f1->type == BACKUP_TYPE_NUMERIC && f2->type == BACKUP_TYPE_NUMERIC) {
        if (f1->numeric_index != f2->numeric_index) {
            return f1->numeric_index - f2->numeric_index;
        }
        return f2->is_compressed - f1->is_compressed;
    }
    
    if (f1->type == BACKUP_TYPE_DATE && f2->type == BACKUP_TYPE_DATE) {
        if (f1->date_value < f2->date_value) return -1;
        if (f1->date_value > f2->date_value) return 1;
        return f2->is_compressed - f1->is_compressed;
    }
    
    if (f1->mtime < f2->mtime) return -1;
    if (f1->mtime > f2->mtime) return 1;
    return 0;
}

static void remove_uncompressed_duplicates(const char *base_path, const char *dir_path)
{
    for (int i = 0; i < g_file_count; i++) {
        if (g_files[i].is_compressed) continue;
        
        for (int j = 0; j < g_file_count; j++) {
            if (i == j) continue;
            if (!g_files[j].is_compressed) continue;
            
            int is_duplicate = 0;
            
            if (g_files[i].type == g_files[j].type) {
                if (g_files[i].type == BACKUP_TYPE_NUMERIC) {
                    if (g_files[i].numeric_index == g_files[j].numeric_index) {
                        is_duplicate = 1;
                    }
                } else if (g_files[i].type == BACKUP_TYPE_DATE) {
                    if (g_files[i].date_value == g_files[j].date_value) {
                        is_duplicate = 1;
                    }
                }
            }
            
            if (!is_duplicate && g_files[i].mtime == g_files[j].mtime && g_files[i].mtime > 0) {
                is_duplicate = 1;
            }
            
            if (is_duplicate) {
                char full_path[BACKUP_MAX_PATH * 2];
                snprintf(full_path, sizeof(full_path), "%s/%s", dir_path, g_files[i].name);
                unlink(full_path);
                g_files[i].type = BACKUP_TYPE_UNKNOWN;
                break;
            }
        }
    }
    
    int new_count = 0;
    for (int i = 0; i < g_file_count; i++) {
        if (g_files[i].type != BACKUP_TYPE_UNKNOWN) {
            if (i != new_count) {
                memcpy(&g_files[new_count], &g_files[i], sizeof(backup_file_t));
            }
            new_count++;
        }
    }
    g_file_count = new_count;
}

int backup_shift(const char *base_path, int max_backups)
{
    if (base_path == NULL || max_backups < 0) return -1;
    if (max_backups == 0) return 0;
    
    char dir_path[BACKUP_MAX_PATH];
    const char *file_name;
    char *last_slash = strrchr(base_path, '/');
    
    if (last_slash != NULL) {
        size_t dir_len = last_slash - base_path;
        if (dir_len >= sizeof(dir_path) - 1) return -1;
        strncpy(dir_path, base_path, dir_len);
        dir_path[dir_len] = '\0';
        file_name = last_slash + 1;
    } else {
        strncpy(dir_path, ".", sizeof(dir_path) - 1);
        dir_path[sizeof(dir_path) - 1] = '\0';
        file_name = base_path;
    }
    
    collect_backup_files(base_path);
    
    remove_uncompressed_duplicates(base_path, dir_path);
    
    for (int i = 0; i < g_file_count; i++) {
        if (g_files[i].type != BACKUP_TYPE_NUMERIC) continue;
        
        char gz_suffix[4] = "";
        if (g_files[i].is_compressed) {
            strcpy(gz_suffix, ".gz");
        }
        
        int idx = g_files[i].numeric_index;
        char src_path[BACKUP_MAX_PATH * 2];
        char dst_path[BACKUP_MAX_PATH * 2];
        
        snprintf(src_path, sizeof(src_path), "%s/%s", dir_path, g_files[i].name);
        
        if (idx >= max_backups) {
            unlink(src_path);
        } else {
            snprintf(dst_path, sizeof(dst_path), "%s/%s.%d%s", 
                     dir_path, file_name, idx + 1, gz_suffix);
            if (file_exists(dst_path)) unlink(dst_path);
            rename(src_path, dst_path);
        }
    }
    
    return 0;
}

int backup_create_with_date(const char *base_path, time_t *date)
{
    if (base_path == NULL) return -1;
    
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
            if (written < 0 || (size_t)written >= sizeof(final_path)) return -1;
        }
        
        struct stat st;
        int exists = (stat(final_path, &st) == 0);
        if (!exists) {
            char gz_path[BACKUP_MAX_PATH];
            snprintf(gz_path, sizeof(gz_path), "%s.gz", final_path);
            exists = (stat(gz_path, &st) == 0);
        }
        
        if (!exists) break;
        suffix++;
    }
    
    if (rename(base_path, final_path) != 0) return -1;
    return 0;
}

int backup_get_count(const char *base_path)
{
    if (base_path == NULL) return -1;
    
    collect_backup_files(base_path);
    return g_file_count;
}

int backup_cleanup(const char *base_path, int max_backups)
{
    if (base_path == NULL || max_backups < 0) return -1;
    if (max_backups == 0) return 0;
    
    char dir_path[BACKUP_MAX_PATH];
    char *last_slash = strrchr(base_path, '/');
    
    if (last_slash != NULL) {
        size_t dir_len = last_slash - base_path;
        if (dir_len >= sizeof(dir_path) - 1) return -1;
        strncpy(dir_path, base_path, dir_len);
        dir_path[dir_len] = '\0';
    } else {
        strncpy(dir_path, ".", sizeof(dir_path) - 1);
        dir_path[sizeof(dir_path) - 1] = '\0';
    }
    
    collect_backup_files(base_path);
    
    remove_uncompressed_duplicates(base_path, dir_path);
    
    if (g_file_count <= max_backups) return 0;
    
    qsort(g_files, g_file_count, sizeof(backup_file_t), compare_backups_by_age);
    
    int to_delete = g_file_count - max_backups;
    for (int i = 0; i < to_delete; i++) {
        char full_path[BACKUP_MAX_PATH * 2];
        snprintf(full_path, sizeof(full_path), "%s/%s", dir_path, g_files[i].name);
        unlink(full_path);
    }
    
    return 0;
}
