#ifndef DESENSITIZE_RULE_H
#define DESENSITIZE_RULE_H

#include "sensitive_info.h"
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

#define MAX_MASK_BUFFER 256

size_t desensitize_phone(const char *original, size_t len, char *output, size_t out_size);
size_t desensitize_id_card(const char *original, size_t len, char *output, size_t out_size);
size_t desensitize_bank_card(const char *original, size_t len, char *output, size_t out_size);
size_t desensitize_email(const char *original, size_t len, char *output, size_t out_size);

typedef size_t (*desensitize_func_t)(const char *original, size_t len, char *output, size_t out_size);

desensitize_func_t get_desensitize_func(sensitive_type_t type);

#ifdef __cplusplus
}
#endif

#endif
