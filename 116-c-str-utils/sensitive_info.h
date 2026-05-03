#ifndef SENSITIVE_INFO_H
#define SENSITIVE_INFO_H

#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    SENSITIVE_TYPE_PHONE,
    SENSITIVE_TYPE_ID_CARD,
    SENSITIVE_TYPE_BANK_CARD,
    SENSITIVE_TYPE_EMAIL,
    SENSITIVE_TYPE_MAX
} sensitive_type_t;

typedef struct {
    size_t start;
    size_t length;
    sensitive_type_t type;
} match_result_t;

int is_digit(char c);
int is_alpha(char c);
int is_alnum(char c);

int match_phone(const char *str, size_t start, match_result_t *result);
int match_id_card(const char *str, size_t start, match_result_t *result);
int match_bank_card(const char *str, size_t start, match_result_t *result);
int match_email(const char *str, size_t start, match_result_t *result);

#ifdef __cplusplus
}
#endif

#endif
