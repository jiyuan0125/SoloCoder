#ifndef UNIQUE_CHECK_H
#define UNIQUE_CHECK_H

#include "common.h"

typedef struct UniqueChecker UniqueChecker;

typedef struct {
    uint64_t total_possible;
    uint64_t used_count;
    double usage_percent;
    bool is_warning;
    bool is_critical;
} CapacityInfo;

UniqueChecker* unique_checker_create(void);
void unique_checker_destroy(UniqueChecker* checker);

ReturnCode unique_checker_add(UniqueChecker* checker, const InviteCodeRecord* record);
ReturnCode unique_checker_remove(UniqueChecker* checker, const char* code);

const InviteCodeRecord* unique_checker_find(const UniqueChecker* checker, const char* code);
bool unique_checker_exists(const UniqueChecker* checker, const char* code);

ReturnCode unique_checker_update_status(UniqueChecker* checker, const char* code, 
                                         CodeStatus new_status, time_t use_time);

uint64_t unique_checker_count(const UniqueChecker* checker);
CapacityInfo unique_checker_get_capacity(const UniqueChecker* checker);

typedef void (*RecordIterator)(const InviteCodeRecord* record, void* user_data);
void unique_checker_iterate(const UniqueChecker* checker, RecordIterator iter, void* user_data);

#endif
