#include "duplicate.h"
#include "fingerprint.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <inttypes.h>

typedef struct {
    off_t size;
    FileInfo **files;
    size_t count;
    size_t capacity;
} SizeGroup;

typedef struct {
    char hash[HASH_LEN];
    FileInfo **files;
    size_t count;
    size_t capacity;
} HashGroup;

static SizeGroup *size_group_create(off_t size) {
    SizeGroup *group = malloc(sizeof(SizeGroup));
    if (!group) return NULL;
    group->size = size;
    group->count = 0;
    group->capacity = 8;
    group->files = malloc(group->capacity * sizeof(FileInfo *));
    if (!group->files) {
        free(group);
        return NULL;
    }
    return group;
}

static void size_group_append(SizeGroup *group, FileInfo *file) {
    if (!group || !file) return;
    if (group->count >= group->capacity) {
        size_t new_capacity = group->capacity * 2;
        FileInfo **new_files = realloc(group->files, new_capacity * sizeof(FileInfo *));
        if (new_files) {
            group->files = new_files;
            group->capacity = new_capacity;
        }
    }
    if (group->count < group->capacity) {
        group->files[group->count++] = file;
    }
}

static void size_group_free(SizeGroup *group) {
    if (group) {
        free(group->files);
        free(group);
    }
}

static HashGroup *hash_group_create(const char *hash) {
    HashGroup *group = malloc(sizeof(HashGroup));
    if (!group) return NULL;
    strncpy(group->hash, hash, HASH_LEN - 1);
    group->hash[HASH_LEN - 1] = '\0';
    group->count = 0;
    group->capacity = 8;
    group->files = malloc(group->capacity * sizeof(FileInfo *));
    if (!group->files) {
        free(group);
        return NULL;
    }
    return group;
}

static void hash_group_append(HashGroup *group, FileInfo *file) {
    if (!group || !file) return;
    if (group->count >= group->capacity) {
        size_t new_capacity = group->capacity * 2;
        FileInfo **new_files = realloc(group->files, new_capacity * sizeof(FileInfo *));
        if (new_files) {
            group->files = new_files;
            group->capacity = new_capacity;
        }
    }
    if (group->count < group->capacity) {
        group->files[group->count++] = file;
    }
}

static void hash_group_free(HashGroup *group) {
    if (group) {
        free(group->files);
        free(group);
    }
}

static SizeGroup **group_by_size(FileList *files, size_t *group_count) {
    if (!files || files->count < 2) {
        *group_count = 0;
        return NULL;
    }

    SizeGroup **groups = NULL;
    size_t capacity = 0;
    size_t count = 0;

    for (size_t i = 0; i < files->count; i++) {
        off_t size = files->files[i].size;
        bool found = false;
        for (size_t j = 0; j < count; j++) {
            if (groups[j]->size == size) {
                size_group_append(groups[j], &files->files[i]);
                found = true;
                break;
            }
        }
        if (!found) {
            if (count >= capacity) {
                size_t new_capacity = capacity == 0 ? 16 : capacity * 2;
                SizeGroup **new_groups = realloc(groups, new_capacity * sizeof(SizeGroup *));
                if (!new_groups) break;
                groups = new_groups;
                capacity = new_capacity;
            }
            groups[count] = size_group_create(size);
            if (groups[count]) {
                size_group_append(groups[count], &files->files[i]);
                count++;
            }
        }
    }

    SizeGroup **result = NULL;
    size_t result_count = 0;
    for (size_t i = 0; i < count; i++) {
        if (groups[i]->count >= 2) {
            result = realloc(result, (result_count + 1) * sizeof(SizeGroup *));
            result[result_count++] = groups[i];
        } else {
            size_group_free(groups[i]);
        }
    }
    free(groups);

    *group_count = result_count;
    return result;
}

static HashGroup **group_by_light_hash(SizeGroup *size_group, size_t *group_count) {
    if (!size_group || size_group->count < 2) {
        *group_count = 0;
        return NULL;
    }

    HashGroup **groups = NULL;
    size_t capacity = 0;
    size_t count = 0;

    for (size_t i = 0; i < size_group->count; i++) {
        FileInfo *file = size_group->files[i];
        bool found = false;
        for (size_t j = 0; j < count; j++) {
            if (strcmp(groups[j]->hash, file->light_hash) == 0) {
                hash_group_append(groups[j], file);
                found = true;
                break;
            }
        }
        if (!found) {
            if (count >= capacity) {
                size_t new_capacity = capacity == 0 ? 8 : capacity * 2;
                HashGroup **new_groups = realloc(groups, new_capacity * sizeof(HashGroup *));
                if (!new_groups) break;
                groups = new_groups;
                capacity = new_capacity;
            }
            groups[count] = hash_group_create(file->light_hash);
            if (groups[count]) {
                hash_group_append(groups[count], file);
                count++;
            }
        }
    }

    HashGroup **result = NULL;
    size_t result_count = 0;
    for (size_t i = 0; i < count; i++) {
        if (groups[i]->count >= 2) {
            result = realloc(result, (result_count + 1) * sizeof(HashGroup *));
            result[result_count++] = groups[i];
        } else {
            hash_group_free(groups[i]);
        }
    }
    free(groups);

    *group_count = result_count;
    return result;
}

static HashGroup **group_by_full_hash(HashGroup *light_group, size_t *group_count) {
    if (!light_group || light_group->count < 2) {
        *group_count = 0;
        return NULL;
    }

    for (size_t i = 0; i < light_group->count; i++) {
        FileInfo *file = light_group->files[i];
        if (strlen(file->full_hash) == 0) {
            compute_full_hash(file->path, file->full_hash, sizeof(file->full_hash));
        }
    }

    HashGroup **groups = NULL;
    size_t capacity = 0;
    size_t count = 0;

    for (size_t i = 0; i < light_group->count; i++) {
        FileInfo *file = light_group->files[i];
        bool found = false;
        for (size_t j = 0; j < count; j++) {
            if (strcmp(groups[j]->hash, file->full_hash) == 0) {
                hash_group_append(groups[j], file);
                found = true;
                break;
            }
        }
        if (!found) {
            if (count >= capacity) {
                size_t new_capacity = capacity == 0 ? 8 : capacity * 2;
                HashGroup **new_groups = realloc(groups, new_capacity * sizeof(HashGroup *));
                if (!new_groups) break;
                groups = new_groups;
                capacity = new_capacity;
            }
            groups[count] = hash_group_create(file->full_hash);
            if (groups[count]) {
                hash_group_append(groups[count], file);
                count++;
            }
        }
    }

    HashGroup **result = NULL;
    size_t result_count = 0;
    for (size_t i = 0; i < count; i++) {
        if (groups[i]->count >= 2) {
            result = realloc(result, (result_count + 1) * sizeof(HashGroup *));
            result[result_count++] = groups[i];
        } else {
            hash_group_free(groups[i]);
        }
    }
    free(groups);

    *group_count = result_count;
    return result;
}

DuplicateReport *find_duplicates(FileList *files) {
    if (!files || files->count < 2) {
        return duplicate_report_create(0);
    }

    compute_hashes_for_list(files, 32);

    size_t size_group_count = 0;
    SizeGroup **size_groups = group_by_size(files, &size_group_count);

    DuplicateReport *report = duplicate_report_create(size_group_count > 0 ? size_group_count : 0);

    for (size_t sg = 0; sg < size_group_count; sg++) {
        size_t light_group_count = 0;
        HashGroup **light_groups = group_by_light_hash(size_groups[sg], &light_group_count);

        for (size_t lg = 0; lg < light_group_count; lg++) {
            size_t full_group_count = 0;
            HashGroup **full_groups = group_by_full_hash(light_groups[lg], &full_group_count);

            for (size_t fg = 0; fg < full_group_count; fg++) {
                if (full_groups[fg]->count >= 2) {
                    DuplicateGroup *group = duplicate_group_create(full_groups[fg]->count);
                    for (size_t i = 0; i < full_groups[fg]->count; i++) {
                        duplicate_group_append(group, full_groups[fg]->files[i]);
                    }
                    group->total_wasted = (group->count - 1) * size_groups[sg]->size;
                    duplicate_report_append(report, group);
                }
                hash_group_free(full_groups[fg]);
            }
            free(full_groups);
            hash_group_free(light_groups[lg]);
        }
        free(light_groups);
        size_group_free(size_groups[sg]);
    }
    free(size_groups);

    return report;
}

static void format_size(off_t size, char *buffer, size_t buffer_len) {
    const char *units[] = {"B", "KB", "MB", "GB", "TB"};
    int unit_index = 0;
    double formatted = (double)size;

    while (formatted >= 1024.0 && unit_index < 4) {
        formatted /= 1024.0;
        unit_index++;
    }

    snprintf(buffer, buffer_len, "%.2f %s", formatted, units[unit_index]);
}

void print_duplicate_report(const DuplicateReport *report) {
    if (!report) {
        printf("No duplicate report available.\n");
        return;
    }

    if (report->count == 0) {
        printf("No duplicate files found.\n");
        return;
    }

    printf("============================================================\n");
    printf("Duplicate File Report\n");
    printf("============================================================\n\n");

    for (size_t g = 0; g < report->count; g++) {
        const DuplicateGroup *group = report->groups[g];
        if (!group || group->count < 2) continue;

        char size_str[32];
        char wasted_str[32];
        format_size(group->files[0]->size, size_str, sizeof(size_str));
        format_size(group->total_wasted, wasted_str, sizeof(wasted_str));

        printf("Group %zu:\n", g + 1);
        printf("  File size:        %s (%" PRId64 " bytes)\n", size_str, (int64_t)group->files[0]->size);
        printf("  Duplicate count:  %zu files\n", group->count);
        printf("  Wasted space:     %s\n", wasted_str);
        printf("  Files:\n");

        for (size_t i = 0; i < group->count; i++) {
            printf("    [%zu] %s\n", i + 1, group->files[i]->path);
        }
        printf("\n");
    }

    printf("============================================================\n");
    char total_wasted_str[32];
    format_size(report->total_wasted, total_wasted_str, sizeof(total_wasted_str));
    printf("Summary:\n");
    printf("  Total duplicate groups: %zu\n", report->count);
    printf("  Total wasted space:     %s (%" PRId64 " bytes)\n", 
           total_wasted_str, (int64_t)report->total_wasted);
    printf("============================================================\n");
}

void iterate_duplicates(const DuplicateReport *report, report_callback_t callback, void *user_data) {
    if (!report || !callback) return;
    for (size_t i = 0; i < report->count; i++) {
        if (report->groups[i]) {
            callback(report->groups[i], user_data);
        }
    }
}
