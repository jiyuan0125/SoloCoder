#include "snapshot.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#define SNAPSHOT_MAGIC 0x534E4150
#define SNAPSHOT_VERSION 1
#define INITIAL_CAPACITY 64

static int entry_compare(const void *a, const void *b) {
    const FileEntry *ea = (const FileEntry *)a;
    const FileEntry *eb = (const FileEntry *)b;
    return strcmp(ea->path, eb->path);
}

Snapshot *snapshot_create(void) {
    Snapshot *snap = (Snapshot *)malloc(sizeof(Snapshot));
    if (!snap) return NULL;
    
    snap->entries = (FileEntry *)malloc(INITIAL_CAPACITY * sizeof(FileEntry));
    if (!snap->entries) {
        free(snap);
        return NULL;
    }
    
    snap->count = 0;
    snap->capacity = INITIAL_CAPACITY;
    snap->timestamp = time(NULL);
    snap->checksum = 0;
    
    return snap;
}

void snapshot_destroy(Snapshot *snap) {
    if (!snap) return;
    
    for (size_t i = 0; i < snap->count; i++) {
        free(snap->entries[i].path);
    }
    free(snap->entries);
    free(snap);
}

int snapshot_add_entry(Snapshot *snap, const char *path, off_t size, time_t mtime, EntryType type) {
    if (!snap || !path) return -1;
    
    if (snap->count >= snap->capacity) {
        size_t new_capacity = snap->capacity * 2;
        FileEntry *new_entries = (FileEntry *)realloc(snap->entries, new_capacity * sizeof(FileEntry));
        if (!new_entries) return -1;
        snap->entries = new_entries;
        snap->capacity = new_capacity;
    }
    
    FileEntry *entry = &snap->entries[snap->count];
    entry->path = strdup(path);
    if (!entry->path) return -1;
    
    entry->size = size;
    entry->mtime = mtime;
    entry->type = type;
    snap->count++;
    
    return 0;
}

FileEntry *snapshot_find_entry(Snapshot *snap, const char *path) {
    if (!snap || !path) return NULL;
    
    for (size_t i = 0; i < snap->count; i++) {
        if (strcmp(snap->entries[i].path, path) == 0) {
            return &snap->entries[i];
        }
    }
    return NULL;
}

void snapshot_sort(Snapshot *snap) {
    if (!snap || snap->count < 2) return;
    qsort(snap->entries, snap->count, sizeof(FileEntry), entry_compare);
}

unsigned int snapshot_calculate_checksum(Snapshot *snap) {
    if (!snap) return 0;
    
    unsigned int checksum = 0;
    for (size_t i = 0; i < snap->count; i++) {
        const char *path = snap->entries[i].path;
        while (*path) {
            checksum = (checksum << 5) + checksum + *path++;
        }
        checksum = (checksum << 5) + checksum + (unsigned int)snap->entries[i].size;
        checksum = (checksum << 5) + checksum + (unsigned int)snap->entries[i].mtime;
        checksum = (checksum << 5) + checksum + (unsigned int)snap->entries[i].type;
    }
    return checksum;
}

int snapshot_validate(Snapshot *snap) {
    if (!snap) return -1;
    
    unsigned int calculated = snapshot_calculate_checksum(snap);
    return (calculated == snap->checksum) ? 0 : -1;
}

int snapshot_save(Snapshot *snap, const char *filename) {
    if (!snap || !filename) return -1;
    
    snap->checksum = snapshot_calculate_checksum(snap);
    snapshot_sort(snap);
    
    FILE *fp = fopen(filename, "wb");
    if (!fp) return -1;
    
    uint32_t magic = SNAPSHOT_MAGIC;
    uint32_t version = SNAPSHOT_VERSION;
    uint64_t count = snap->count;
    uint64_t timestamp = (uint64_t)snap->timestamp;
    uint32_t checksum = snap->checksum;
    
    fwrite(&magic, sizeof(magic), 1, fp);
    fwrite(&version, sizeof(version), 1, fp);
    fwrite(&count, sizeof(count), 1, fp);
    fwrite(&timestamp, sizeof(timestamp), 1, fp);
    fwrite(&checksum, sizeof(checksum), 1, fp);
    
    for (size_t i = 0; i < snap->count; i++) {
        uint32_t path_len = (uint32_t)strlen(snap->entries[i].path) + 1;
        uint64_t size = (uint64_t)snap->entries[i].size;
        uint64_t mtime = (uint64_t)snap->entries[i].mtime;
        uint8_t type = (uint8_t)snap->entries[i].type;
        
        fwrite(&path_len, sizeof(path_len), 1, fp);
        fwrite(snap->entries[i].path, 1, path_len, fp);
        fwrite(&size, sizeof(size), 1, fp);
        fwrite(&mtime, sizeof(mtime), 1, fp);
        fwrite(&type, sizeof(type), 1, fp);
    }
    
    uint32_t end_magic = SNAPSHOT_MAGIC;
    fwrite(&end_magic, sizeof(end_magic), 1, fp);
    
    fflush(fp);
    if (fclose(fp) != 0) {
        return -1;
    }
    
    return 0;
}

Snapshot *snapshot_load(const char *filename) {
    if (!filename) return NULL;
    
    FILE *fp = fopen(filename, "rb");
    if (!fp) return NULL;
    
    uint32_t magic, version, checksum, end_magic;
    uint64_t count, timestamp;
    
    if (fread(&magic, sizeof(magic), 1, fp) != 1 || magic != SNAPSHOT_MAGIC) {
        fclose(fp);
        return NULL;
    }
    
    if (fread(&version, sizeof(version), 1, fp) != 1 || version != SNAPSHOT_VERSION) {
        fclose(fp);
        return NULL;
    }
    
    if (fread(&count, sizeof(count), 1, fp) != 1) {
        fclose(fp);
        return NULL;
    }
    
    if (fread(&timestamp, sizeof(timestamp), 1, fp) != 1) {
        fclose(fp);
        return NULL;
    }
    
    if (fread(&checksum, sizeof(checksum), 1, fp) != 1) {
        fclose(fp);
        return NULL;
    }
    
    Snapshot *snap = snapshot_create();
    if (!snap) {
        fclose(fp);
        return NULL;
    }
    
    snap->timestamp = (time_t)timestamp;
    snap->checksum = checksum;
    
    for (uint64_t i = 0; i < count; i++) {
        uint32_t path_len;
        uint64_t size, mtime;
        uint8_t type;
        char *path;
        
        if (fread(&path_len, sizeof(path_len), 1, fp) != 1) {
            snapshot_destroy(snap);
            fclose(fp);
            return NULL;
        }
        
        path = (char *)malloc(path_len);
        if (!path || fread(path, 1, (size_t)path_len, fp) != (size_t)path_len) {
            free(path);
            snapshot_destroy(snap);
            fclose(fp);
            return NULL;
        }
        
        if (fread(&size, sizeof(size), 1, fp) != 1 ||
            fread(&mtime, sizeof(mtime), 1, fp) != 1 ||
            fread(&type, sizeof(type), 1, fp) != 1) {
            free(path);
            snapshot_destroy(snap);
            fclose(fp);
            return NULL;
        }
        
        if (snapshot_add_entry(snap, path, (off_t)size, (time_t)mtime, (EntryType)type) != 0) {
            free(path);
            snapshot_destroy(snap);
            fclose(fp);
            return NULL;
        }
        free(path);
    }
    
    if (fread(&end_magic, sizeof(end_magic), 1, fp) != 1 || end_magic != SNAPSHOT_MAGIC) {
        snapshot_destroy(snap);
        fclose(fp);
        return NULL;
    }
    
    fclose(fp);
    
    if (snapshot_validate(snap) != 0) {
        snapshot_destroy(snap);
        return NULL;
    }
    
    return snap;
}
