#ifndef DIFF_H
#define DIFF_H

#include "snapshot.h"

typedef enum {
    CHANGE_ADDED,
    CHANGE_MODIFIED,
    CHANGE_DELETED
} ChangeType;

typedef struct Change {
    char *path;
    ChangeType type;
    EntryType entry_type;
} Change;

typedef struct ChangeList {
    Change *changes;
    size_t count;
    size_t capacity;
} ChangeList;

typedef struct DiffOptions {
    time_t delete_add_merge_window;
} DiffOptions;

ChangeList *changelist_create(void);
void changelist_destroy(ChangeList *list);
int changelist_add(ChangeList *list, const char *path, ChangeType type, EntryType entry_type);
void changelist_sort(ChangeList *list);

ChangeList *diff_compare(Snapshot *old_snap, Snapshot *new_snap, DiffOptions *options);

const char *change_type_to_string(ChangeType type);

#endif
