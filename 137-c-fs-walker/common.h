#ifndef COMMON_H
#define COMMON_H

#include <stdint.h>
#include <stdlib.h>
#include <stdbool.h>
#include <sys/types.h>

#define MAX_PATH_LEN 4096
#define HASH_LEN 64

typedef struct {
    char path[MAX_PATH_LEN];
    off_t size;
    ino_t inode;
    char light_hash[HASH_LEN];
    char full_hash[HASH_LEN];
} FileInfo;

typedef struct {
    FileInfo *files;
    size_t count;
    size_t capacity;
} FileList;

typedef struct {
    char **patterns;
    size_t count;
} PatternList;

typedef struct {
    PatternList exclude_dirs;
    PatternList exclude_patterns;
    off_t min_size;
    off_t max_size;
    bool follow_symlinks;
    int max_open_files;
} ScanConfig;

typedef struct {
    FileInfo **files;
    size_t count;
    size_t capacity;
    off_t total_wasted;
} DuplicateGroup;

typedef struct {
    DuplicateGroup **groups;
    size_t count;
    size_t capacity;
    off_t total_wasted;
} DuplicateReport;

FileList *file_list_create(size_t initial_capacity);
void file_list_append(FileList *list, const FileInfo *info);
void file_list_free(FileList *list);

PatternList *pattern_list_create(void);
void pattern_list_append(PatternList *list, const char *pattern);
void pattern_list_clear(PatternList *list);
void pattern_list_free(PatternList *list);
bool pattern_list_match(const PatternList *list, const char *str);

DuplicateGroup *duplicate_group_create(size_t initial_capacity);
void duplicate_group_append(DuplicateGroup *group, FileInfo *info);
void duplicate_group_free(DuplicateGroup *group);

DuplicateReport *duplicate_report_create(size_t initial_capacity);
void duplicate_report_append(DuplicateReport *report, DuplicateGroup *group);
void duplicate_report_free(DuplicateReport *report);

#endif
