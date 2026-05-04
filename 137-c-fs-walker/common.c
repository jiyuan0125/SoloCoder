#include "common.h"
#include <string.h>
#include <fnmatch.h>
#include <stdio.h>

static void ensure_capacity(void **array, size_t *capacity, size_t current_count, size_t element_size) {
    if (current_count >= *capacity) {
        size_t new_capacity = *capacity == 0 ? 8 : (*capacity * 2);
        void *new_array = realloc(*array, new_capacity * element_size);
        if (new_array) {
            *array = new_array;
            *capacity = new_capacity;
        }
    }
}

FileList *file_list_create(size_t initial_capacity) {
    FileList *list = malloc(sizeof(FileList));
    if (!list) return NULL;
    list->count = 0;
    list->capacity = initial_capacity > 0 ? initial_capacity : 32;
    list->files = malloc(list->capacity * sizeof(FileInfo));
    if (!list->files) {
        free(list);
        return NULL;
    }
    return list;
}

void file_list_append(FileList *list, const FileInfo *info) {
    if (!list || !info) return;
    ensure_capacity((void **)&list->files, &list->capacity, list->count, sizeof(FileInfo));
    if (list->count < list->capacity) {
        list->files[list->count] = *info;
        list->count++;
    }
}

void file_list_free(FileList *list) {
    if (list) {
        free(list->files);
        free(list);
    }
}

PatternList *pattern_list_create(void) {
    PatternList *list = malloc(sizeof(PatternList));
    if (!list) return NULL;
    list->count = 0;
    list->patterns = NULL;
    return list;
}

void pattern_list_append(PatternList *list, const char *pattern) {
    if (!list || !pattern) return;
    size_t new_count = list->count + 1;
    char **new_patterns = realloc(list->patterns, new_count * sizeof(char *));
    if (!new_patterns) return;
    
    list->patterns = new_patterns;
    list->patterns[list->count] = strdup(pattern);
    list->count = new_count;
}

void pattern_list_clear(PatternList *list) {
    if (!list) return;
    for (size_t i = 0; i < list->count; i++) {
        free(list->patterns[i]);
    }
    free(list->patterns);
    list->patterns = NULL;
    list->count = 0;
}

void pattern_list_free(PatternList *list) {
    if (list) {
        pattern_list_clear(list);
        free(list);
    }
}

bool pattern_list_match(const PatternList *list, const char *str) {
    if (!list || !str) return false;
    for (size_t i = 0; i < list->count; i++) {
        if (fnmatch(list->patterns[i], str, 0) == 0) {
            return true;
        }
    }
    return false;
}

DuplicateGroup *duplicate_group_create(size_t initial_capacity) {
    DuplicateGroup *group = malloc(sizeof(DuplicateGroup));
    if (!group) return NULL;
    group->count = 0;
    group->total_wasted = 0;
    group->capacity = initial_capacity > 0 ? initial_capacity : 8;
    group->files = malloc(group->capacity * sizeof(FileInfo *));
    if (!group->files) {
        free(group);
        return NULL;
    }
    return group;
}

void duplicate_group_append(DuplicateGroup *group, FileInfo *info) {
    if (!group || !info) return;
    ensure_capacity((void **)&group->files, &group->capacity, group->count, sizeof(FileInfo *));
    if (group->count < group->capacity) {
        group->files[group->count] = info;
        group->count++;
    }
}

void duplicate_group_free(DuplicateGroup *group) {
    if (group) {
        free(group->files);
        free(group);
    }
}

DuplicateReport *duplicate_report_create(size_t initial_capacity) {
    DuplicateReport *report = malloc(sizeof(DuplicateReport));
    if (!report) return NULL;
    report->count = 0;
    report->total_wasted = 0;
    report->capacity = initial_capacity > 0 ? initial_capacity : 16;
    report->groups = malloc(report->capacity * sizeof(DuplicateGroup *));
    if (!report->groups) {
        free(report);
        return NULL;
    }
    return report;
}

void duplicate_report_append(DuplicateReport *report, DuplicateGroup *group) {
    if (!report || !group) return;
    ensure_capacity((void **)&report->groups, &report->capacity, report->count, sizeof(DuplicateGroup *));
    if (report->count < report->capacity) {
        report->groups[report->count] = group;
        report->total_wasted += group->total_wasted;
        report->count++;
    }
}

void duplicate_report_free(DuplicateReport *report) {
    if (report) {
        for (size_t i = 0; i < report->count; i++) {
            duplicate_group_free(report->groups[i]);
        }
        free(report->groups);
        free(report);
    }
}
