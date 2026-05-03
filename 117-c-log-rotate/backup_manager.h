#ifndef BACKUP_MANAGER_H
#define BACKUP_MANAGER_H

#include <time.h>

#ifdef __cplusplus
extern "C" {
#endif

#define BACKUP_MAX_PATH 1024

typedef struct {
    char base_path[BACKUP_MAX_PATH];
    int max_backups;
} backup_config_t;

int backup_shift(const char *base_path, int max_backups);
int backup_create_with_date(const char *base_path, time_t *date);
int backup_cleanup(const char *base_path, int max_backups);
int backup_cleanup_temp_files(const char *base_path);
int backup_get_count(const char *base_path);
int backup_generate_filename(char *buf, size_t buf_size, const char *base_path, int index);
int backup_generate_date_filename(char *buf, size_t buf_size, const char *base_path, time_t date);

#ifdef __cplusplus
}
#endif

#endif
