#include "sstable.h"
#include "memtable.h"
#include "kv_store.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

static int key_compare(const char* key1, size_t len1, const char* key2, size_t len2) {
    size_t min_len = len1 < len2 ? len1 : len2;
    int cmp = memcmp(key1, key2, min_len);
    if (cmp != 0) return cmp;
    if (len1 < len2) return -1;
    if (len1 > len2) return 1;
    return 0;
}

SSTableList* sstable_list_create(void) {
    SSTableList* list = (SSTableList*)malloc(sizeof(SSTableList));
    if (!list) return NULL;
    
    list->count = 0;
    list->capacity = 8;
    list->tables = (SSTable**)malloc(list->capacity * sizeof(SSTable*));
    if (!list->tables) {
        free(list);
        return NULL;
    }
    
    return list;
}

void sstable_list_destroy(SSTableList* list) {
    if (!list) return;
    for (size_t i = 0; i < list->count; i++) {
        sstable_close(list->tables[i]);
    }
    free(list->tables);
    free(list);
}

int sstable_list_add(SSTableList* list, SSTable* table) {
    if (!list || !table) return -1;
    
    if (list->count >= list->capacity) {
        size_t new_capacity = list->capacity * 2;
        SSTable** new_tables = (SSTable**)realloc(list->tables, new_capacity * sizeof(SSTable*));
        if (!new_tables) return -1;
        list->tables = new_tables;
        list->capacity = new_capacity;
    }
    
    list->tables[list->count++] = table;
    return 0;
}

void sstable_list_remove(SSTableList* list, size_t index) {
    if (!list || index >= list->count) return;
    
    sstable_close(list->tables[index]);
    
    for (size_t i = index; i < list->count - 1; i++) {
        list->tables[i] = list->tables[i + 1];
    }
    list->count--;
}

SSTable* sstable_list_get(SSTableList* list, size_t index) {
    if (!list || index >= list->count) return NULL;
    return list->tables[index];
}

size_t sstable_list_count(SSTableList* list) {
    if (!list) return 0;
    return list->count;
}

SSTable* sstable_create_from_memtable(const char* dir, uint64_t id, SkipList* memtable) {
    if (!dir || !memtable) return NULL;
    
    char path[512];
    snprintf(path, sizeof(path), "%s/sst_%lu.sst", dir, (unsigned long)id);
    
    FILE* file = fopen(path, "wb");
    if (!file) return NULL;
    
    size_t entry_count = 0;
    SkipListNode* node = skiplist_first(memtable);
    
    while (node) {
        uint32_t key_len = (uint32_t)node->key_len;
        
        fwrite(&key_len, sizeof(uint32_t), 1, file);
        fwrite(node->key, 1, key_len, file);
        
        if (node->deleted) {
            uint32_t tombstone = SSTABLE_TOMBSTONE;
            fwrite(&tombstone, sizeof(uint32_t), 1, file);
        } else {
            uint32_t value_len = (uint32_t)node->value_len;
            fwrite(&value_len, sizeof(uint32_t), 1, file);
            if (value_len > 0 && node->value) {
                fwrite(node->value, 1, value_len, file);
            }
        }
        
        entry_count++;
        node = skiplist_next(node);
    }
    
    fflush(file);
    fclose(file);
    
    return sstable_open(path);
}

SSTable* sstable_open(const char* path) {
    if (!path) return NULL;
    
    SSTable* table = (SSTable*)malloc(sizeof(SSTable));
    if (!table) return NULL;
    
    table->path = strdup(path);
    if (!table->path) {
        free(table);
        return NULL;
    }
    
    table->file = fopen(path, "rb");
    if (!table->file) {
        free(table->path);
        free(table);
        return NULL;
    }
    
    const char* name = strrchr(path, '/');
    if (name) name++;
    else name = path;
    
    table->id = 0;
    if (sscanf(name, "sst_%lu.sst", (unsigned long*)&table->id) != 1) {
        table->id = 0;
    }
    
    table->entry_count = 0;
    fseek(table->file, 0, SEEK_SET);
    while (1) {
        uint32_t key_len;
        if (fread(&key_len, sizeof(uint32_t), 1, table->file) != 1) break;
        if (fseek(table->file, key_len, SEEK_CUR) != 0) break;
        
        uint32_t value_len;
        if (fread(&value_len, sizeof(uint32_t), 1, table->file) != 1) break;
        if (fseek(table->file, value_len, SEEK_CUR) != 0) break;
        
        table->entry_count++;
    }
    fseek(table->file, 0, SEEK_SET);
    
    return table;
}

void sstable_close(SSTable* table) {
    if (!table) return;
    if (table->file) fclose(table->file);
    free(table->path);
    free(table);
}

KVEntry* sstable_get(SSTable* table, const char* key, size_t key_len) {
    if (!table || !table->file || !key || key_len == 0) return NULL;
    
    fseek(table->file, 0, SEEK_SET);
    
    while (1) {
        uint32_t current_key_len;
        if (fread(&current_key_len, sizeof(uint32_t), 1, table->file) != 1) break;
        
        char* current_key = (char*)malloc(current_key_len + 1);
        if (!current_key) break;
        
        if (fread(current_key, 1, current_key_len, table->file) != current_key_len) {
            free(current_key);
            break;
        }
        current_key[current_key_len] = '\0';
        
        int cmp = key_compare(key, key_len, current_key, current_key_len);
        
        if (cmp == 0) {
            uint32_t value_len;
            if (fread(&value_len, sizeof(uint32_t), 1, table->file) != 1) {
                free(current_key);
                break;
            }
            
            if (value_len == SSTABLE_TOMBSTONE) {
                free(current_key);
                return NULL;
            }
            
            char* value = NULL;
            if (value_len > 0) {
                value = (char*)malloc(value_len + 1);
                if (!value) {
                    free(current_key);
                    break;
                }
                if (fread(value, 1, value_len, table->file) != value_len) {
                    free(value);
                    free(current_key);
                    break;
                }
                value[value_len] = '\0';
            }
            
            KVEntry* entry = kv_entry_create(current_key, current_key_len, value, value_len, false);
            free(current_key);
            free(value);
            return entry;
        } else if (cmp < 0) {
            free(current_key);
            break;
        } else {
            uint32_t value_len;
            if (fread(&value_len, sizeof(uint32_t), 1, table->file) != 1) {
                free(current_key);
                break;
            }
            if (value_len != SSTABLE_TOMBSTONE && fseek(table->file, value_len, SEEK_CUR) != 0) {
                free(current_key);
                break;
            }
            free(current_key);
        }
    }
    
    return NULL;
}

uint64_t sstable_get_id(SSTable* table) {
    if (!table) return 0;
    return table->id;
}
