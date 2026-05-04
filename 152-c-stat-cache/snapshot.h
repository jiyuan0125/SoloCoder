#ifndef SNAPSHOT_H
#define SNAPSHOT_H

#include <sys/types.h>
#include <sys/stat.h>
#include <time.h>

typedef enum {
    ENTRY_FILE,
    ENTRY_DIR
} EntryType;

typedef struct FileEntry {
    char *path;
    off_t size;
    time_t mtime;
    EntryType type;
} FileEntry;

typedef struct Snapshot {
    FileEntry *entries;
    size_t count;
    size_t capacity;
    time_t timestamp;
    unsigned int checksum;
} Snapshot;

Snapshot *snapshot_create(void);
void snapshot_destroy(Snapshot *snap);
int snapshot_add_entry(Snapshot *snap, const char *path, off_t size, time_t mtime, EntryType type);
FileEntry *snapshot_find_entry(Snapshot *snap, const char *path);
int snapshot_save(Snapshot *snap, const char *filename);
Snapshot *snapshot_load(const char *filename);
int snapshot_validate(Snapshot *snap);
void snapshot_sort(Snapshot *snap);
unsigned int snapshot_calculate_checksum(Snapshot *snap);

#endif
