#include "invite_service.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define MAX_GENERATE_ATTEMPTS 100

struct InviteService {
    CodeGenerator* generator;
    UniqueChecker* checker;
    Storage* storage;
};

InviteService* invite_service_create(const char* storage_path) {
    InviteService* service = (InviteService*)malloc(sizeof(InviteService));
    if (!service) return NULL;
    
    service->generator = code_generator_create();
    if (!service->generator) {
        free(service);
        return NULL;
    }
    
    service->checker = unique_checker_create();
    if (!service->checker) {
        code_generator_destroy(service->generator);
        free(service);
        return NULL;
    }
    
    service->storage = storage_create(storage_path, STORAGE_FORMAT_BINARY);
    if (!service->storage) {
        unique_checker_destroy(service->checker);
        code_generator_destroy(service->generator);
        free(service);
        return NULL;
    }
    
    return service;
}

void invite_service_destroy(InviteService* service) {
    if (!service) return;
    
    if (service->storage) {
        storage_destroy(service->storage);
    }
    if (service->checker) {
        unique_checker_destroy(service->checker);
    }
    if (service->generator) {
        code_generator_destroy(service->generator);
    }
    free(service);
}

static bool is_code_format_valid(const char* code) {
    if (!code || strlen(code) != INVITE_CODE_LENGTH) {
        return false;
    }
    
    for (int i = 0; code[i]; i++) {
        if (!code_generator_is_valid_char(code[i])) {
            return false;
        }
    }
    
    return true;
}

ReturnCode invite_service_generate(InviteService* service, const char* user_id, 
                                    char* out_code, size_t out_size) {
    if (!service || !user_id || !out_code || out_size < INVITE_CODE_LENGTH + 1) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    if (strlen(user_id) >= MAX_USER_ID_LEN) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    for (int attempt = 0; attempt < MAX_GENERATE_ATTEMPTS; attempt++) {
        char temp_code[INVITE_CODE_LENGTH + 1];
        ReturnCode rc = code_generator_generate_single(service->generator, temp_code, sizeof(temp_code));
        if (rc != RC_SUCCESS) {
            return rc;
        }
        
        if (!unique_checker_exists(service->checker, temp_code)) {
            InviteCodeRecord record = {0};
            memcpy(record.code, temp_code, INVITE_CODE_LENGTH);
            record.code[INVITE_CODE_LENGTH] = '\0';
            strncpy(record.user_id, user_id, MAX_USER_ID_LEN - 1);
            record.user_id[MAX_USER_ID_LEN - 1] = '\0';
            record.status = CODE_STATUS_VALID;
            record.create_time = time(NULL);
            record.use_time = 0;
            
            rc = unique_checker_add(service->checker, &record);
            if (rc == RC_SUCCESS) {
                strncpy(out_code, temp_code, out_size - 1);
                out_code[out_size - 1] = '\0';
                return RC_SUCCESS;
            }
        }
    }
    
    return RC_ERROR_CAPACITY_LIMIT;
}

ReturnCode invite_service_generate_batch(InviteService* service, const char* user_id,
                                          int count, char*** out_codes, int* out_actual_count) {
    if (!service || !user_id || count <= 0 || !out_codes || !out_actual_count) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    *out_codes = NULL;
    *out_actual_count = 0;
    
    char** codes = (char**)calloc(count, sizeof(char*));
    if (!codes) {
        return RC_ERROR_STORAGE;
    }
    
    int generated = 0;
    for (int i = 0; i < count; i++) {
        codes[i] = (char*)malloc(INVITE_CODE_LENGTH + 1);
        if (!codes[i]) {
            invite_service_free_codes(codes, generated);
            free(codes);
            return RC_ERROR_STORAGE;
        }
        
        ReturnCode rc = invite_service_generate(service, user_id, codes[i], INVITE_CODE_LENGTH + 1);
        if (rc == RC_SUCCESS) {
            generated++;
        } else if (rc == RC_ERROR_CAPACITY_LIMIT) {
            free(codes[i]);
            codes[i] = NULL;
            break;
        } else {
            free(codes[i]);
            codes[i] = NULL;
            invite_service_free_codes(codes, generated);
            free(codes);
            return rc;
        }
    }
    
    if (generated == 0) {
        free(codes);
        return RC_ERROR_CAPACITY_LIMIT;
    }
    
    *out_codes = codes;
    *out_actual_count = generated;
    return RC_SUCCESS;
}

ReturnCode invite_service_verify(InviteService* service, const char* code, 
                                  InviteCodeInfo* out_info) {
    if (!service || !code) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    if (!is_code_format_valid(code)) {
        return RC_ERROR_INVALID_CODE;
    }
    
    const InviteCodeRecord* record = unique_checker_find(service->checker, code);
    if (!record) {
        return RC_ERROR_NOT_FOUND;
    }
    
    if (record->status != CODE_STATUS_VALID) {
        return RC_ERROR_INVALID_CODE;
    }
    
    if (out_info) {
        strncpy(out_info->code, record->code, INVITE_CODE_LENGTH);
        out_info->code[INVITE_CODE_LENGTH] = '\0';
        strncpy(out_info->user_id, record->user_id, MAX_USER_ID_LEN - 1);
        out_info->user_id[MAX_USER_ID_LEN - 1] = '\0';
        out_info->status = record->status;
        out_info->create_time = record->create_time;
        out_info->use_time = record->use_time;
    }
    
    return RC_SUCCESS;
}

ReturnCode invite_service_use(InviteService* service, const char* code) {
    if (!service || !code) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    const InviteCodeRecord* record = unique_checker_find(service->checker, code);
    if (!record) {
        return RC_ERROR_NOT_FOUND;
    }
    
    if (record->status != CODE_STATUS_VALID) {
        return RC_ERROR_INVALID_CODE;
    }
    
    return unique_checker_update_status(service->checker, code, CODE_STATUS_USED, time(NULL));
}

ReturnCode invite_service_revoke(InviteService* service, const char* code) {
    if (!service || !code) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    const InviteCodeRecord* record = unique_checker_find(service->checker, code);
    if (!record) {
        return RC_ERROR_NOT_FOUND;
    }
    
    if (record->status == CODE_STATUS_USED) {
        return RC_ERROR_INVALID_CODE;
    }
    
    return unique_checker_update_status(service->checker, code, CODE_STATUS_REVOKED, 0);
}

ReturnCode invite_service_save(InviteService* service) {
    if (!service) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    return storage_save(service->storage, service->checker);
}

ReturnCode invite_service_load(InviteService* service) {
    if (!service) {
        return RC_ERROR_INVALID_PARAM;
    }
    
    return storage_load(service->storage, service->checker);
}

CapacityInfo invite_service_get_capacity(const InviteService* service) {
    if (!service || !service->checker) {
        CapacityInfo empty = {0};
        return empty;
    }
    
    return unique_checker_get_capacity(service->checker);
}

uint64_t invite_service_get_total_count(const InviteService* service) {
    if (!service || !service->checker) {
        return 0;
    }
    
    return unique_checker_count(service->checker);
}

void invite_service_free_codes(char** codes, int count) {
    if (!codes || count <= 0) return;
    
    for (int i = 0; i < count; i++) {
        if (codes[i]) {
            free(codes[i]);
            codes[i] = NULL;
        }
    }
    free(codes);
}
