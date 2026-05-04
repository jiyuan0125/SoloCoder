#ifndef STORAGE_H
#define STORAGE_H

#include "common.h"
#include "unique_check.h"

typedef struct Storage Storage;

typedef enum {
    STORAGE_FORMAT_BINARY = 0,
    STORAGE_FORMAT_TEXT = 1
} StorageFormat;

Storage* storage_create(const char* filename, StorageFormat format);
void storage_destroy(Storage* storage);

ReturnCode storage_save(Storage* storage, const UniqueChecker* checker);
ReturnCode storage_load(Storage* storage, UniqueChecker* checker);

ReturnCode storage_append_record(Storage* storage, const InviteCodeRecord* record);
ReturnCode storage_update_record(Storage* storage, const InviteCodeRecord* record);

const char* storage_get_filename(const Storage* storage);
StorageFormat storage_get_format(const Storage* storage);

bool storage_file_exists(const char* filename);
uint64_t storage_get_record_count(const Storage* storage);

#endif
