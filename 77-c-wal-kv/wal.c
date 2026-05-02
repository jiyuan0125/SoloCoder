#include "wal.h"
#include "memtable.h"
#include <stdlib.h>
#include <string.h>

WAL* wal_open(const char* path) {
    WAL* wal = (WAL*)malloc(sizeof(WAL));
    if (!wal) return NULL;
    
    wal->path = strdup(path);
    if (!wal->path) {
        free(wal);
        return NULL;
    }
    
    wal->file = fopen(path, "a+b");
    if (!wal->file) {
        free(wal->path);
        free(wal);
        return NULL;
    }
    
    return wal;
}

void wal_close(WAL* wal) {
    if (!wal) return;
    if (wal->file) fclose(wal->file);
    free(wal->path);
    free(wal);
}

void wal_sync(WAL* wal) {
    if (wal && wal->file) {
        fflush(wal->file);
    }
}

int wal_append(WAL* wal, RecordType type, const char* key, size_t key_len, const char* value, size_t value_len) {
    if (!wal || !wal->file || !key || key_len == 0 || key_len > MAX_KEY_LEN) return -1;
    if (value_len > MAX_VALUE_LEN) return -1;
    
    size_t record_len = 1 + 2 + key_len + 4;
    if (type == RECORD_PUT && value) {
        record_len += value_len;
    }
    
    uint8_t* record = (uint8_t*)malloc(record_len);
    if (!record) return -1;
    
    size_t offset = 0;
    record[offset++] = (uint8_t)type;
    
    record[offset++] = (uint8_t)(key_len >> 8);
    record[offset++] = (uint8_t)(key_len & 0xff);
    
    memcpy(record + offset, key, key_len);
    offset += key_len;
    
    if (type == RECORD_PUT && value) {
        record[offset++] = (uint8_t)(value_len >> 24);
        record[offset++] = (uint8_t)((value_len >> 16) & 0xff);
        record[offset++] = (uint8_t)((value_len >> 8) & 0xff);
        record[offset++] = (uint8_t)(value_len & 0xff);
        
        memcpy(record + offset, value, value_len);
        offset += value_len;
    } else {
        record[offset++] = 0;
        record[offset++] = 0;
        record[offset++] = 0;
        record[offset++] = 0;
    }
    
    uint32_t crc = crc32(record, record_len);
    
    fseek(wal->file, 0, SEEK_END);
    fwrite(&crc, sizeof(uint32_t), 1, wal->file);
    fwrite(&record_len, sizeof(uint32_t), 1, wal->file);
    fwrite(record, 1, record_len, wal->file);
    
    free(record);
    return 0;
}

int wal_replay(WAL* wal, SkipList* memtable) {
    if (!wal || !wal->file || !memtable) return -1;
    
    fseek(wal->file, 0, SEEK_SET);
    
    while (1) {
        uint32_t crc_stored;
        uint32_t record_len;
        
        if (fread(&crc_stored, sizeof(uint32_t), 1, wal->file) != 1) break;
        if (fread(&record_len, sizeof(uint32_t), 1, wal->file) != 1) break;
        
        if (record_len == 0 || record_len > (1 + 2 + MAX_KEY_LEN + 4 + MAX_VALUE_LEN)) {
            fseek(wal->file, -(int)sizeof(uint32_t), SEEK_CUR);
            break;
        }
        
        uint8_t* record = (uint8_t*)malloc(record_len);
        if (!record) {
            fseek(wal->file, -(int)(sizeof(uint32_t) * 2), SEEK_CUR);
            break;
        }
        
        size_t nread = fread(record, 1, record_len, wal->file);
        if (nread != record_len) {
            free(record);
            fseek(wal->file, -(int)(sizeof(uint32_t) * 2 + nread), SEEK_CUR);
            break;
        }
        
        uint32_t crc_calculated = crc32(record, record_len);
        if (crc_calculated != crc_stored) {
            free(record);
            fseek(wal->file, -(int)(sizeof(uint32_t) * 2 + record_len), SEEK_CUR);
            break;
        }
        
        size_t offset = 0;
        RecordType type = (RecordType)record[offset++];
        
        uint16_t key_len = (record[offset] << 8) | record[offset + 1];
        offset += 2;
        
        char* key = (char*)malloc(key_len + 1);
        if (!key) {
            free(record);
            break;
        }
        memcpy(key, record + offset, key_len);
        key[key_len] = '\0';
        offset += key_len;
        
        uint32_t value_len = (record[offset] << 24) | 
                              (record[offset + 1] << 16) | 
                              (record[offset + 2] << 8) | 
                              record[offset + 3];
        offset += 4;
        
        char* value = NULL;
        if (type == RECORD_PUT && value_len > 0) {
            value = (char*)malloc(value_len + 1);
            if (!value) {
                free(key);
                free(record);
                break;
            }
            memcpy(value, record + offset, value_len);
            value[value_len] = '\0';
        }
        
        if (type == RECORD_PUT) {
            skiplist_put(memtable, key, key_len, value, value_len);
        } else if (type == RECORD_DELETE) {
            skiplist_delete(memtable, key, key_len);
        }
        
        free(key);
        free(value);
        free(record);
    }
    
    fseek(wal->file, 0, SEEK_END);
    return 0;
}
