#include "storage.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define STORAGE_MAGIC 0x494E5654
#define STORAGE_VERSION 1

#define MAX_FILENAME_LEN 512

typedef struct {
    uint32_t magic;
    uint16_t version;
    uint32_t reserved;
    uint64_t record_count;
} StorageHeader;

struct Storage {
    char filename[MAX_FILENAME_LEN];
    StorageFormat format;
    uint64_t record_count;
};

Storage* storage_create(const char* filename, StorageFormat format) {
    if (!filename || strlen(filename) >= MAX_FILENAME_LEN) {
        return NULL;
    }
    
    Storage* storage = (Storage*)malloc(sizeof(Storage));
    if (!storage) return NULL;
    
    strncpy(storage->filename, filename, MAX_FILENAME_LEN - 1);
    storage->filename[MAX_FILENAME_LEN - 1] = '\0';
    storage->format = format;
    storage->record_count = 0;
    
    return storage;
}

void storage_destroy(Storage* storage) {
    if (storage) {
        free(storage);
    }
}

static ReturnCode storage_save_binary(Storage* storage, const UniqueChecker* checker) {
    FILE* fp = fopen(storage->filename, "wb");
    if (!fp) {
        return RC_ERROR_STORAGE;
    }
    
    StorageHeader header = {0};
    header.magic = STORAGE_MAGIC;
    header.version = STORAGE_VERSION;
    header.reserved = 0;
    header.record_count = unique_checker_count(checker);
    
    if (fwrite(&header, sizeof(StorageHeader), 1, fp) != 1) {
        fclose(fp);
        return RC_ERROR_STORAGE;
    }
    
    typedef struct {
        FILE* fp;
        int error;
    } SaveContext;
    
    SaveContext ctx = {fp, 0};
    
    void save_iter(const InviteCodeRecord* record, void* user_data) {
        SaveContext* ctx_ptr = (SaveContext*)user_data;
        if (ctx_ptr->error) return;
        
        if (fwrite(record, sizeof(InviteCodeRecord), 1, ctx_ptr->fp) != 1) {
            ctx_ptr->error = 1;
        }
    }
    
    unique_checker_iterate(checker, save_iter, &ctx);
    
    if (ctx.error) {
        fclose(fp);
        return RC_ERROR_STORAGE;
    }
    
    if (fclose(fp) != 0) {
        return RC_ERROR_STORAGE;
    }
    
    storage->record_count = header.record_count;
    return RC_SUCCESS;
}

static ReturnCode storage_load_binary(Storage* storage, UniqueChecker* checker) {
    FILE* fp = fopen(storage->filename, "rb");
    if (!fp) {
        return RC_ERROR_STORAGE;
    }
    
    StorageHeader header = {0};
    if (fread(&header, sizeof(StorageHeader), 1, fp) != 1) {
        fclose(fp);
        return RC_ERROR_STORAGE;
    }
    
    if (header.magic != STORAGE_MAGIC) {
        fclose(fp);
        return RC_ERROR_STORAGE;
    }
    
    for (uint64_t i = 0; i < header.record_count; i++) {
        InviteCodeRecord record = {0};
        if (fread(&record, sizeof(InviteCodeRecord), 1, fp) != 1) {
            fclose(fp);
            return RC_ERROR_STORAGE;
        }
        
        ReturnCode rc = unique_checker_add(checker, &record);
        if (rc != RC_SUCCESS && rc != RC_ERROR_DUPLICATE_CODE) {
            fclose(fp);
            return rc;
        }
    }
    
    if (fclose(fp) != 0) {
        return RC_ERROR_STORAGE;
    }
    
    storage->record_count = header.record_count;
    return RC_SUCCESS;
}

ReturnCode storage_save(Storage* storage, const UniqueChecker* checker) {
    if (!storage || !checker) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    switch (storage->format) {
        case STORAGE_FORMAT_BINARY:
            return storage_save_binary(storage, checker);
        case STORAGE_FORMAT_TEXT:
        default:
            return RC_ERROR_INVALID_PARAM;
    }
}

ReturnCode storage_load(Storage* storage, UniqueChecker* checker) {
    if (!storage || !checker) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    if (!storage_file_exists(storage->filename)) {
        return RC_SUCCESS;
    }
    
    switch (storage->format) {
        case STORAGE_FORMAT_BINARY:
            return storage_load_binary(storage, checker);
        case STORAGE_FORMAT_TEXT:
        default:
            return RC_ERROR_INVALID_PARAM;
    }
}

ReturnCode storage_append_record(Storage* storage, const InviteCodeRecord* record) {
    (void)storage;
    (void)record;
    return RC_ERROR_STORAGE;
}

ReturnCode storage_update_record(Storage* storage, const InviteCodeRecord* record) {
    (void)storage;
    (void)record;
    return RC_ERROR_STORAGE;
}

const char* storage_get_filename(const Storage* storage) {
    if (!storage) return NULL;
    return storage->filename;
}

StorageFormat storage_get_format(const Storage* storage) {
    if (!storage) return STORAGE_FORMAT_BINARY;
    return storage->format;
}

bool storage_file_exists(const char* filename) {
    if (!filename) return false;
    FILE* fp = fopen(filename, "rb");
    if (fp) {
        fclose(fp);
        return true;
    }
    return false;
}

uint64_t storage_get_record_count(const Storage* storage) {
    if (!storage) return 0;
    return storage->record_count;
}
