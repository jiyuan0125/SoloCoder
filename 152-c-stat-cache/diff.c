#include "diff.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define INITIAL_CHANGES_CAPACITY 32

static int change_compare(const void *a, const void *b) {
    const Change *ca = (const Change *)a;
    const Change *cb = (const Change *)b;
    return strcmp(ca->path, cb->path);
}

ChangeList *changelist_create(void) {
    ChangeList *list = (ChangeList *)malloc(sizeof(ChangeList));
    if (!list) return NULL;
    
    list->changes = (Change *)malloc(INITIAL_CHANGES_CAPACITY * sizeof(Change));
    if (!list->changes) {
        free(list);
        return NULL;
    }
    
    list->count = 0;
    list->capacity = INITIAL_CHANGES_CAPACITY;
    return list;
}

void changelist_destroy(ChangeList *list) {
    if (!list) return;
    
    for (size_t i = 0; i < list->count; i++) {
        free(list->changes[i].path);
    }
    free(list->changes);
    free(list);
}

int changelist_add(ChangeList *list, const char *path, ChangeType type, EntryType entry_type) {
    if (!list || !path) return -1;
    
    if (list->count >= list->capacity) {
        size_t new_capacity = list->capacity * 2;
        Change *new_changes = (Change *)realloc(list->changes, new_capacity * sizeof(Change));
        if (!new_changes) return -1;
        list->changes = new_changes;
        list->capacity = new_capacity;
    }
    
    Change *change = &list->changes[list->count];
    change->path = strdup(path);
    if (!change->path) return -1;
    
    change->type = type;
    change->entry_type = entry_type;
    list->count++;
    
    return 0;
}

void changelist_sort(ChangeList *list) {
    if (!list || list->count < 2) return;
    qsort(list->changes, list->count, sizeof(Change), change_compare);
}

const char *change_type_to_string(ChangeType type) {
    switch (type) {
        case CHANGE_ADDED:
            return "新增";
        case CHANGE_MODIFIED:
            return "已修改";
        case CHANGE_DELETED:
            return "已删除";
        default:
            return "未知";
    }
}

static int entry_unchanged(FileEntry *old_entry, FileEntry *new_entry) {
    if (old_entry->type != new_entry->type) return 0;
    if (old_entry->size != new_entry->size) return 0;
    if (old_entry->mtime != new_entry->mtime) return 0;
    return 1;
}

ChangeList *diff_compare(Snapshot *old_snap, Snapshot *new_snap, DiffOptions *options) {
    (void)options;
    ChangeList *list = changelist_create();
    if (!list) return NULL;
    
    if (!old_snap) {
        if (new_snap) {
            for (size_t i = 0; i < new_snap->count; i++) {
                if (changelist_add(list, new_snap->entries[i].path, 
                                   CHANGE_ADDED, new_snap->entries[i].type) != 0) {
                    changelist_destroy(list);
                    return NULL;
                }
            }
        }
        changelist_sort(list);
        return list;
    }
    
    if (!new_snap) {
        for (size_t i = 0; i < old_snap->count; i++) {
            if (changelist_add(list, old_snap->entries[i].path, 
                               CHANGE_DELETED, old_snap->entries[i].type) != 0) {
                changelist_destroy(list);
                return NULL;
            }
        }
        changelist_sort(list);
        return list;
    }
    
    size_t old_idx = 0;
    size_t new_idx = 0;
    
    while (old_idx < old_snap->count && new_idx < new_snap->count) {
        FileEntry *old_entry = &old_snap->entries[old_idx];
        FileEntry *new_entry = &new_snap->entries[new_idx];
        
        int cmp = strcmp(old_entry->path, new_entry->path);
        
        if (cmp == 0) {
            if (!entry_unchanged(old_entry, new_entry)) {
                if (changelist_add(list, old_entry->path, 
                                   CHANGE_MODIFIED, old_entry->type) != 0) {
                    changelist_destroy(list);
                    return NULL;
                }
            }
            old_idx++;
            new_idx++;
        } else if (cmp < 0) {
            if (changelist_add(list, old_entry->path, 
                               CHANGE_DELETED, old_entry->type) != 0) {
                changelist_destroy(list);
                return NULL;
            }
            old_idx++;
        } else {
            if (changelist_add(list, new_entry->path, 
                               CHANGE_ADDED, new_entry->type) != 0) {
                changelist_destroy(list);
                return NULL;
            }
            new_idx++;
        }
    }
    
    while (old_idx < old_snap->count) {
        FileEntry *old_entry = &old_snap->entries[old_idx];
        if (changelist_add(list, old_entry->path, 
                           CHANGE_DELETED, old_entry->type) != 0) {
            changelist_destroy(list);
            return NULL;
        }
        old_idx++;
    }
    
    while (new_idx < new_snap->count) {
        FileEntry *new_entry = &new_snap->entries[new_idx];
        if (changelist_add(list, new_entry->path, 
                           CHANGE_ADDED, new_entry->type) != 0) {
            changelist_destroy(list);
            return NULL;
        }
        new_idx++;
    }
    
    changelist_sort(list);
    return list;
}
