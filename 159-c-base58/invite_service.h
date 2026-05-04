#ifndef INVITE_SERVICE_H
#define INVITE_SERVICE_H

#include "common.h"
#include "code_generator.h"
#include "unique_check.h"
#include "storage.h"

typedef struct InviteService InviteService;

typedef struct {
    char code[INVITE_CODE_LENGTH + 1];
    char user_id[MAX_USER_ID_LEN];
    CodeStatus status;
    time_t create_time;
    time_t use_time;
} InviteCodeInfo;

InviteService* invite_service_create(const char* storage_path);
void invite_service_destroy(InviteService* service);

ReturnCode invite_service_generate(InviteService* service, const char* user_id, 
                                    char* out_code, size_t out_size);
ReturnCode invite_service_generate_batch(InviteService* service, const char* user_id,
                                          int count, char*** out_codes, int* out_actual_count);

ReturnCode invite_service_verify(InviteService* service, const char* code, 
                                  InviteCodeInfo* out_info);

ReturnCode invite_service_use(InviteService* service, const char* code);
ReturnCode invite_service_revoke(InviteService* service, const char* code);

ReturnCode invite_service_save(InviteService* service);
ReturnCode invite_service_load(InviteService* service);

CapacityInfo invite_service_get_capacity(const InviteService* service);
uint64_t invite_service_get_total_count(const InviteService* service);

void invite_service_free_codes(char** codes, int count);

#endif
