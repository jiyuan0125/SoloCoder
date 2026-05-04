#ifndef COMMON_H
#define COMMON_H

#include <stdint.h>
#include <stdbool.h>
#include <time.h>

#define INVITE_CODE_LENGTH 6
#define BASE58_CHAR_COUNT 58
#define MAX_USER_ID_LEN 64
#define INVALID_CODE_ERROR "邀请码无效"
#define CODE_USED_ERROR "邀请码已被使用"
#define CAPACITY_WARNING_THRESHOLD 0.8

typedef enum {
    CODE_STATUS_VALID = 0,
    CODE_STATUS_USED = 1,
    CODE_STATUS_REVOKED = 2
} CodeStatus;

typedef struct {
    char code[INVITE_CODE_LENGTH + 1];
    char user_id[MAX_USER_ID_LEN];
    CodeStatus status;
    time_t create_time;
    time_t use_time;
} InviteCodeRecord;

typedef enum {
    RC_SUCCESS = 0,
    RC_ERROR_INVALID_PARAM = -1,
    RC_ERROR_DUPLICATE_CODE = -2,
    RC_ERROR_NOT_FOUND = -3,
    RC_ERROR_STORAGE = -4,
    RC_ERROR_CAPACITY_LIMIT = -5,
    RC_ERROR_INVALID_CODE = -6
} ReturnCode;

extern const char BASE58_ALPHABET[BASE58_CHAR_COUNT + 1];

#endif
