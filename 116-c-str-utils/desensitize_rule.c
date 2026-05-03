#include "desensitize_rule.h"
#include <string.h>

static size_t calculate_masked_length(size_t original_len, size_t keep_before, size_t keep_after) {
    if (original_len <= keep_before + keep_after) {
        return original_len;
    }
    return keep_before + (original_len - keep_before - keep_after) + keep_after;
}

static size_t apply_mask(const char *original, size_t len, char *output, size_t out_size,
                          size_t keep_before, size_t keep_after) {
    size_t needed = calculate_masked_length(len, keep_before, keep_after);
    
    if (output == NULL || out_size == 0) {
        return needed;
    }
    
    if (out_size <= needed) {
        if (out_size > 0) {
            output[0] = '\0';
        }
        return needed;
    }
    
    size_t i = 0;
    size_t out_idx = 0;
    
    for (i = 0; i < keep_before && i < len; i++) {
        output[out_idx++] = original[i];
    }
    
    size_t mask_count = len - keep_before - keep_after;
    if (mask_count > 0) {
        for (i = 0; i < mask_count; i++) {
            output[out_idx++] = '*';
        }
    }
    
    size_t start_after = len > keep_after ? len - keep_after : 0;
    for (i = start_after; i < len; i++) {
        output[out_idx++] = original[i];
    }
    
    output[out_idx] = '\0';
    return out_idx;
}

size_t desensitize_phone(const char *original, size_t len, char *output, size_t out_size) {
    return apply_mask(original, len, output, out_size, 3, 4);
}

size_t desensitize_id_card(const char *original, size_t len, char *output, size_t out_size) {
    return apply_mask(original, len, output, out_size, 3, 4);
}

size_t desensitize_bank_card(const char *original, size_t len, char *output, size_t out_size) {
    return apply_mask(original, len, output, out_size, 0, 4);
}

static size_t find_at_position(const char *str, size_t len) {
    for (size_t i = 0; i < len; i++) {
        if (str[i] == '@') {
            return i;
        }
    }
    return len;
}

size_t desensitize_email(const char *original, size_t len, char *output, size_t out_size) {
    size_t at_pos = find_at_position(original, len);
    
    if (at_pos == len || at_pos == 0) {
        if (output != NULL && out_size > len) {
            memcpy(output, original, len);
            output[len] = '\0';
        }
        return len;
    }
    
    size_t needed;
    if (at_pos <= 1) {
        needed = len;
    } else {
        needed = 1 + (at_pos - 1) + (len - at_pos);
    }
    
    if (output == NULL || out_size == 0) {
        return needed;
    }
    
    if (out_size <= needed) {
        if (out_size > 0) {
            output[0] = '\0';
        }
        return needed;
    }
    
    size_t out_idx = 0;
    
    output[out_idx++] = original[0];
    
    if (at_pos > 1) {
        for (size_t i = 1; i < at_pos; i++) {
            output[out_idx++] = '*';
        }
    }
    
    for (size_t i = at_pos; i < len; i++) {
        output[out_idx++] = original[i];
    }
    
    output[out_idx] = '\0';
    return out_idx;
}

desensitize_func_t get_desensitize_func(sensitive_type_t type) {
    switch (type) {
        case SENSITIVE_TYPE_PHONE:
            return desensitize_phone;
        case SENSITIVE_TYPE_ID_CARD:
            return desensitize_id_card;
        case SENSITIVE_TYPE_BANK_CARD:
            return desensitize_bank_card;
        case SENSITIVE_TYPE_EMAIL:
            return desensitize_email;
        default:
            return NULL;
    }
}
