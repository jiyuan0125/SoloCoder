#include "sensitive_info.h"
#include <string.h>

int is_digit(char c) {
    return (c >= '0' && c <= '9');
}

int is_alpha(char c) {
    return ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'));
}

int is_alnum(char c) {
    return (is_digit(c) || is_alpha(c));
}

static int has_preceding_digit(const char *str, size_t start) {
    if (start == 0) return 0;
    return is_digit(str[start - 1]);
}

static int has_following_digit(const char *str, size_t start, size_t length) {
    size_t end = start + length;
    if (str[end] == '\0') return 0;
    return is_digit(str[end]);
}

static size_t count_consecutive_digits(const char *str, size_t start) {
    size_t count = 0;
    while (is_digit(str[start + count])) {
        count++;
    }
    return count;
}

static int is_valid_phone_at(const char *str, size_t start) {
    if (str[start] != '1') {
        return 0;
    }
    for (size_t i = 0; i < 11; i++) {
        if (!is_digit(str[start + i])) {
            return 0;
        }
    }
    return 1;
}

static int can_split_phones(const char *str, size_t start, size_t total_digits) {
    if (total_digits % 11 != 0) {
        return 0;
    }
    for (size_t i = 0; i < total_digits; i += 11) {
        if (!is_valid_phone_at(str, start + i)) {
            return 0;
        }
    }
    return 1;
}

static size_t find_digit_start(const char *str, size_t current) {
    size_t start = current;
    while (start > 0 && is_digit(str[start - 1])) {
        start--;
    }
    return start;
}

int match_phone(const char *str, size_t start, match_result_t *result) {
    if (str[start] != '1') {
        return 0;
    }
    
    size_t digit_count = count_consecutive_digits(str, start);
    
    if (digit_count < 11) {
        return 0;
    }
    
    int is_valid = 0;
    
    if (digit_count == 11) {
        int preceding_digit = has_preceding_digit(str, start);
        int following_digit = has_following_digit(str, start, 11);
        
        if (!preceding_digit && !following_digit) {
            is_valid = 1;
        } else if (preceding_digit) {
            size_t actual_start = find_digit_start(str, start);
            size_t total_digits = (start - actual_start) + digit_count;
            
            if (can_split_phones(str, actual_start, total_digits)) {
                size_t offset = start - actual_start;
                if (offset % 11 == 0) {
                    is_valid = 1;
                }
            }
        }
    } else {
        if (can_split_phones(str, start, digit_count)) {
            is_valid = 1;
        }
    }
    
    if (!is_valid) {
        return 0;
    }
    
    if (result != NULL) {
        result->start = start;
        result->length = 11;
        result->type = SENSITIVE_TYPE_PHONE;
    }
    
    return 1;
}

static int is_valid_id_card_char(char c) {
    return is_digit(c) || c == 'X' || c == 'x';
}

static size_t count_id_card_chars(const char *str, size_t start) {
    size_t count = 0;
    while (is_valid_id_card_char(str[start + count])) {
        count++;
    }
    return count;
}

static int char_to_digit(char c) {
    return c - '0';
}

static int is_valid_birthday_18(const char *str, size_t start) {
    int year = char_to_digit(str[start + 6]) * 1000 +
               char_to_digit(str[start + 7]) * 100 +
               char_to_digit(str[start + 8]) * 10 +
               char_to_digit(str[start + 9]);
    
    int month = char_to_digit(str[start + 10]) * 10 +
                char_to_digit(str[start + 11]);
    
    int day = char_to_digit(str[start + 12]) * 10 +
              char_to_digit(str[start + 13]);
    
    if (year < 1900 || year > 2030) {
        return 0;
    }
    
    if (month < 1 || month > 12) {
        return 0;
    }
    
    int max_day;
    switch (month) {
        case 2:
            if ((year % 4 == 0 && year % 100 != 0) || (year % 400 == 0)) {
                max_day = 29;
            } else {
                max_day = 28;
            }
            break;
        case 4:
        case 6:
        case 9:
        case 11:
            max_day = 30;
            break;
        default:
            max_day = 31;
            break;
    }
    
    if (day < 1 || day > max_day) {
        return 0;
    }
    
    return 1;
}

static int is_valid_birthday_15(const char *str, size_t start) {
    int year = 1900 +
               char_to_digit(str[start + 6]) * 10 +
               char_to_digit(str[start + 7]);
    
    int month = char_to_digit(str[start + 8]) * 10 +
                char_to_digit(str[start + 9]);
    
    int day = char_to_digit(str[start + 10]) * 10 +
              char_to_digit(str[start + 11]);
    
    if (year < 1900 || year > 2030) {
        return 0;
    }
    
    if (month < 1 || month > 12) {
        return 0;
    }
    
    int max_day;
    switch (month) {
        case 2:
            if ((year % 4 == 0 && year % 100 != 0) || (year % 400 == 0)) {
                max_day = 29;
            } else {
                max_day = 28;
            }
            break;
        case 4:
        case 6:
        case 9:
        case 11:
            max_day = 30;
            break;
        default:
            max_day = 31;
            break;
    }
    
    if (day < 1 || day > max_day) {
        return 0;
    }
    
    return 1;
}

int match_id_card(const char *str, size_t start, match_result_t *result) {
    size_t char_count = count_id_card_chars(str, start);
    
    if (char_count != 15 && char_count != 18) {
        return 0;
    }
    
    if (char_count == 18) {
        for (size_t i = start; i < start + 17; i++) {
            if (!is_digit(str[i])) {
                return 0;
            }
        }
        if (!is_valid_id_card_char(str[start + 17])) {
            return 0;
        }
        if (!is_valid_birthday_18(str, start)) {
            return 0;
        }
    } else {
        for (size_t i = start; i < start + 15; i++) {
            if (!is_digit(str[i])) {
                return 0;
            }
        }
        if (!is_valid_birthday_15(str, start)) {
            return 0;
        }
    }
    
    if (has_preceding_digit(str, start)) {
        return 0;
    }
    
    size_t end = start + char_count;
    if (str[end] != '\0') {
        if (is_digit(str[end])) {
            return 0;
        }
        if (char_count == 18 && (str[end] == 'X' || str[end] == 'x')) {
            return 0;
        }
    }
    
    if (result != NULL) {
        result->start = start;
        result->length = char_count;
        result->type = SENSITIVE_TYPE_ID_CARD;
    }
    
    return 1;
}

static int luhn_check(const char *str, size_t start, size_t length) {
    int sum = 0;
    int double_next = 0;
    
    for (int i = (int)length - 1; i >= 0; i--) {
        int digit = char_to_digit(str[start + i]);
        
        if (double_next) {
            digit *= 2;
            if (digit > 9) {
                digit -= 9;
            }
        }
        
        sum += digit;
        double_next = !double_next;
    }
    
    return (sum % 10 == 0);
}

int match_bank_card(const char *str, size_t start, match_result_t *result) {
    size_t digit_count = count_consecutive_digits(str, start);
    
    if (digit_count < 13 || digit_count > 19) {
        return 0;
    }
    
    if (has_preceding_digit(str, start)) {
        return 0;
    }
    
    if (has_following_digit(str, start, digit_count)) {
        return 0;
    }
    
    if (!luhn_check(str, start, digit_count)) {
        return 0;
    }
    
    if (result != NULL) {
        result->start = start;
        result->length = digit_count;
        result->type = SENSITIVE_TYPE_BANK_CARD;
    }
    
    return 1;
}

static int is_valid_email_local_char(char c) {
    return is_alnum(c) || c == '.' || c == '_' || c == '-' || c == '+';
}

static int is_valid_email_domain_char(char c) {
    return is_alnum(c) || c == '.' || c == '-';
}

static int is_valid_email_start(char c) {
    return is_alnum(c);
}

int match_email(const char *str, size_t start, match_result_t *result) {
    if (!is_valid_email_start(str[start])) {
        return 0;
    }
    
    size_t pos = start;
    int has_at = 0;
    size_t at_pos = 0;
    int has_dot_after_at = 0;
    
    while (str[pos] != '\0') {
        if (str[pos] == '@') {
            if (has_at || pos == start) {
                return 0;
            }
            has_at = 1;
            at_pos = pos;
        } else if (has_at) {
            if (str[pos] == '.') {
                if (pos == at_pos + 1) {
                    return 0;
                }
                has_dot_after_at = 1;
            }
            if (!is_valid_email_domain_char(str[pos])) {
                break;
            }
        } else {
            if (!is_valid_email_local_char(str[pos])) {
                break;
            }
        }
        pos++;
    }
    
    if (!has_at || !has_dot_after_at) {
        return 0;
    }
    
    if (str[pos - 1] == '.') {
        pos--;
    }
    
    if (pos - at_pos < 3) {
        return 0;
    }
    
    if (result != NULL) {
        result->start = start;
        result->length = pos - start;
        result->type = SENSITIVE_TYPE_EMAIL;
    }
    
    return 1;
}
