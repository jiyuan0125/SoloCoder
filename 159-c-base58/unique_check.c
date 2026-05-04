#include "unique_check.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <math.h>

#define HASH_TABLE_SIZE 1000003

typedef struct HashNode {
    InviteCodeRecord record;
    struct HashNode* next;
} HashNode;

struct UniqueChecker {
    HashNode** buckets;
    uint64_t count;
};

static uint64_t calculate_total_possible(void) {
    uint64_t result = 1;
    for (int i = 0; i < INVITE_CODE_LENGTH; i++) {
        result *= BASE58_CHAR_COUNT;
    }
    return result;
}

static uint32_t hash_code(const char* code) {
    uint32_t hash = 5381;
    int c;
    
    while ((c = (unsigned char)*code++)) {
        hash = ((hash << 5) + hash) + c;
    }
    
    return hash % HASH_TABLE_SIZE;
}

UniqueChecker* unique_checker_create(void) {
    UniqueChecker* checker = (UniqueChecker*)malloc(sizeof(UniqueChecker));
    if (!checker) return NULL;
    
    checker->buckets = (HashNode**)calloc(HASH_TABLE_SIZE, sizeof(HashNode*));
    if (!checker->buckets) {
        free(checker);
        return NULL;
    }
    
    checker->count = 0;
    return checker;
}

void unique_checker_destroy(UniqueChecker* checker) {
    if (!checker) return;
    
    if (checker->buckets) {
        for (int i = 0; i < HASH_TABLE_SIZE; i++) {
            HashNode* node = checker->buckets[i];
            while (node) {
                HashNode* next = node->next;
                free(node);
                node = next;
            }
        }
        free(checker->buckets);
    }
    free(checker);
}

ReturnCode unique_checker_add(UniqueChecker* checker, const InviteCodeRecord* record) {
    if (!checker || !record || !record->code[0]) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    if (unique_checker_exists(checker, record->code)) {
        return RC_ERROR_DUPLICATE_CODE;
    }
    
    CapacityInfo cap = unique_checker_get_capacity(checker);
    if (cap.is_critical) {
        return RC_ERROR_CAPACITY_LIMIT;
    }
    
    uint32_t idx = hash_code(record->code);
    
    HashNode* new_node = (HashNode*)malloc(sizeof(HashNode));
    if (!new_node) {
        return RC_ERROR_STORAGE;
    }
    
    memcpy(&new_node->record, record, sizeof(InviteCodeRecord));
    new_node->next = checker->buckets[idx];
    checker->buckets[idx] = new_node;
    checker->count++;
    
    return RC_SUCCESS;
}

ReturnCode unique_checker_remove(UniqueChecker* checker, const char* code) {
    if (!checker || !code || !code[0]) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    uint32_t idx = hash_code(code);
    HashNode* prev = NULL;
    HashNode* node = checker->buckets[idx];
    
    while (node) {
        if (strcmp(node->record.code, code) == 0) {
            if (prev) {
                prev->next = node->next;
            } else {
                checker->buckets[idx] = node->next;
            }
            free(node);
            checker->count--;
            return RC_SUCCESS;
        }
        prev = node;
        node = node->next;
    }
    
    return RC_ERROR_NOT_FOUND;
}

const InviteCodeRecord* unique_checker_find(const UniqueChecker* checker, const char* code) {
    if (!checker || !code || !code[0]) {
        return NULL;
    }
    
    uint32_t idx = hash_code(code);
    const HashNode* node = checker->buckets[idx];
    
    while (node) {
        if (strcmp(node->record.code, code) == 0) {
            return &node->record;
        }
        node = node->next;
    }
    
    return NULL;
}

bool unique_checker_exists(const UniqueChecker* checker, const char* code) {
    return unique_checker_find(checker, code) != NULL;
}

ReturnCode unique_checker_update_status(UniqueChecker* checker, const char* code, 
                                         CodeStatus new_status, time_t use_time) {
    if (!checker || !code || !code[0]) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    uint32_t idx = hash_code(code);
    HashNode* node = checker->buckets[idx];
    
    while (node) {
        if (strcmp(node->record.code, code) == 0) {
            node->record.status = new_status;
            if (new_status == CODE_STATUS_USED) {
                node->record.use_time = use_time;
            }
            return RC_SUCCESS;
        }
        node = node->next;
    }
    
    return RC_ERROR_NOT_FOUND;
}

uint64_t unique_checker_count(const UniqueChecker* checker) {
    if (!checker) return 0;
    return checker->count;
}

CapacityInfo unique_checker_get_capacity(const UniqueChecker* checker) {
    CapacityInfo info = {0};
    info.total_possible = calculate_total_possible();
    
    if (checker) {
        info.used_count = checker->count;
    } else {
        info.used_count = 0;
    }
    
    if (info.total_possible > 0) {
        info.usage_percent = (double)info.used_count / (double)info.total_possible * 100.0;
    } else {
        info.usage_percent = 0.0;
    }
    
    info.is_warning = (info.usage_percent >= CAPACITY_WARNING_THRESHOLD * 100.0);
    info.is_critical = (info.used_count >= info.total_possible);
    
    return info;
}

void unique_checker_iterate(const UniqueChecker* checker, RecordIterator iter, void* user_data) {
    if (!checker || !iter) return;
    
    for (int i = 0; i < HASH_TABLE_SIZE; i++) {
        const HashNode* node = checker->buckets[i];
        while (node) {
            iter(&node->record, user_data);
            node = node->next;
        }
    }
}
