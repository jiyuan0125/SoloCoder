#define _DEFAULT_SOURCE
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "wildcard.h"

static char last_error[256] = {0};

const char* wc_pattern_get_error(void) {
    return last_error;
}

static void set_error(const char *msg) {
    strncpy(last_error, msg, sizeof(last_error) - 1);
    last_error[sizeof(last_error) - 1] = '\0';
}

static wc_char_class_t* char_class_create(void) {
    wc_char_class_t *cc = (wc_char_class_t*)malloc(sizeof(wc_char_class_t));
    if (!cc) return NULL;
    cc->ranges = NULL;
    cc->count = 0;
    cc->capacity = 0;
    cc->negated = false;
    return cc;
}

static void char_class_destroy(wc_char_class_t *cc) {
    if (cc) {
        free(cc->ranges);
        cc->ranges = NULL;
        cc->count = 0;
        cc->capacity = 0;
    }
}

static void char_class_full_destroy(wc_char_class_t *cc) {
    if (cc) {
        char_class_destroy(cc);
        free(cc);
    }
}

static int char_class_add_range(wc_char_class_t *cc, char start, char end, bool is_range) {
    if (cc->count >= cc->capacity) {
        size_t new_cap = cc->capacity == 0 ? 8 : cc->capacity * 2;
        wc_char_range_t *new_ranges = (wc_char_range_t*)realloc(
            cc->ranges, new_cap * sizeof(wc_char_range_t));
        if (!new_ranges) return -1;
        cc->ranges = new_ranges;
        cc->capacity = new_cap;
    }
    cc->ranges[cc->count].ch = start;
    cc->ranges[cc->count].end = end;
    cc->ranges[cc->count].is_range = is_range;
    cc->count++;
    return 0;
}

static wc_pattern_t* pattern_create_empty(void) {
    wc_pattern_t *p = (wc_pattern_t*)malloc(sizeof(wc_pattern_t));
    if (!p) return NULL;
    p->tokens = NULL;
    p->count = 0;
    p->capacity = 0;
    p->original_pattern = NULL;
    p->wildcard_count = 0;
    p->starstar_count = 0;
    return p;
}

void wc_pattern_destroy(wc_pattern_t *pattern) {
    if (!pattern) return;
    for (size_t i = 0; i < pattern->count; i++) {
        if (pattern->tokens[i].type == WC_TOKEN_CHARCLASS) {
            char_class_destroy(&pattern->tokens[i].data.char_class);
        }
    }
    free(pattern->tokens);
    free(pattern->original_pattern);
    free(pattern);
}

static int pattern_add_token(wc_pattern_t *p, wc_token_type_t type, void *data) {
    if (p->count >= p->capacity) {
        size_t new_cap = p->capacity == 0 ? 16 : p->capacity * 2;
        wc_token_t *new_tokens = (wc_token_t*)realloc(
            p->tokens, new_cap * sizeof(wc_token_t));
        if (!new_tokens) return -1;
        p->tokens = new_tokens;
        p->capacity = new_cap;
    }
    p->tokens[p->count].type = type;
    if (type == WC_TOKEN_LITERAL) {
        p->tokens[p->count].data.literal = *(char*)data;
    } else if (type == WC_TOKEN_CHARCLASS) {
        memcpy(&p->tokens[p->count].data.char_class, data, sizeof(wc_char_class_t));
    }
    if (type == WC_TOKEN_STAR || type == WC_TOKEN_STARSTAR || 
        type == WC_TOKEN_QUESTION || type == WC_TOKEN_CHARCLASS) {
        p->wildcard_count++;
    }
    if (type == WC_TOKEN_STARSTAR) {
        p->starstar_count++;
    }
    p->count++;
    return 0;
}

static int pattern_add_literal(wc_pattern_t *p, char ch) {
    return pattern_add_token(p, WC_TOKEN_LITERAL, &ch);
}

static int pattern_add_star(wc_pattern_t *p) {
    if (p->count > 0) {
        wc_token_type_t last_type = p->tokens[p->count - 1].type;
        if (last_type == WC_TOKEN_STAR) {
            return 0;
        }
        if (last_type == WC_TOKEN_STARSTAR) {
            return 0;
        }
    }
    char dummy = 0;
    return pattern_add_token(p, WC_TOKEN_STAR, &dummy);
}

static int pattern_add_starstar(wc_pattern_t *p) {
    if (p->count > 0) {
        wc_token_type_t last_type = p->tokens[p->count - 1].type;
        if (last_type == WC_TOKEN_STARSTAR) {
            return 0;
        }
        if (last_type == WC_TOKEN_STAR) {
            p->tokens[p->count - 1].type = WC_TOKEN_STARSTAR;
            p->starstar_count++;
            return 0;
        }
    }
    char dummy = 0;
    return pattern_add_token(p, WC_TOKEN_STARSTAR, &dummy);
}

static int pattern_add_question(wc_pattern_t *p) {
    char dummy = 0;
    return pattern_add_token(p, WC_TOKEN_QUESTION, &dummy);
}

static int pattern_add_charclass(wc_pattern_t *p, wc_char_class_t *cc) {
    return pattern_add_token(p, WC_TOKEN_CHARCLASS, cc);
}

static char parse_escaped_char(const char **p) {
    const char *ptr = *p;
    if (*ptr != '\\') return 0;
    ptr++;
    if (*ptr == '\0') {
        set_error("Incomplete escape sequence");
        return 0;
    }
    char result;
    switch (*ptr) {
        case '\\': result = '\\'; break;
        case '*':  result = '*';  break;
        case '?':  result = '?';  break;
        case '[':  result = '[';  break;
        case ']':  result = ']';  break;
        case '-':  result = '-';  break;
        case '^':  result = '^';  break;
        case '!':  result = '!';  break;
        default:
            result = *ptr;
            break;
    }
    *p = ptr;
    return result;
}

static wc_char_class_t* parse_char_class(const char **p) {
    const char *ptr = *p;
    if (*ptr != '[') return NULL;
    ptr++;
    
    wc_char_class_t *cc = char_class_create();
    if (!cc) {
        set_error("Out of memory");
        return NULL;
    }
    
    if (*ptr == '^' || *ptr == '!') {
        cc->negated = true;
        ptr++;
    }
    
    bool first = true;
    while (*ptr != ']' && *ptr != '\0') {
        if (*ptr == '\\') {
            char escaped = parse_escaped_char(&ptr);
            if (escaped == 0 && last_error[0] != '\0') {
                char_class_full_destroy(cc);
                return NULL;
            }
            if (char_class_add_range(cc, escaped, escaped, false) < 0) {
                set_error("Out of memory");
                char_class_full_destroy(cc);
                return NULL;
            }
            first = false;
            ptr++;
        } else if (*ptr == '-' && !first && *(ptr + 1) != ']' && *(ptr + 1) != '\0') {
            ptr++;
            char end;
            if (*ptr == '\\') {
                end = parse_escaped_char(&ptr);
                if (end == 0 && last_error[0] != '\0') {
                    char_class_full_destroy(cc);
                    return NULL;
                }
            } else {
                end = *ptr;
            }
            cc->ranges[cc->count - 1].end = end;
            cc->ranges[cc->count - 1].is_range = true;
            first = false;
            ptr++;
        } else {
            char ch = *ptr;
            if (char_class_add_range(cc, ch, ch, false) < 0) {
                set_error("Out of memory");
                char_class_full_destroy(cc);
                return NULL;
            }
            first = false;
            ptr++;
        }
    }
    
    if (*ptr != ']') {
        set_error("Unterminated character class");
        char_class_full_destroy(cc);
        return NULL;
    }
    ptr++;
    
    *p = ptr;
    return cc;
}

wc_pattern_t* wc_pattern_create(const char *pattern_str, wc_case_mode_t case_mode) {
    (void)case_mode;
    
    if (!pattern_str) {
        set_error("NULL pattern");
        return NULL;
    }
    
    size_t len = strlen(pattern_str);
    if (len >= WC_MAX_PATTERN_LEN) {
        set_error("Pattern too long");
        return NULL;
    }
    
    wc_pattern_t *p = pattern_create_empty();
    if (!p) {
        set_error("Out of memory");
        return NULL;
    }
    
    p->original_pattern = strdup(pattern_str);
    if (!p->original_pattern) {
        set_error("Out of memory");
        wc_pattern_destroy(p);
        return NULL;
    }
    
    const char *ptr = pattern_str;
    while (*ptr != '\0') {
        if (*ptr == '\\') {
            char escaped = parse_escaped_char(&ptr);
            if (escaped == 0 && last_error[0] != '\0') {
                wc_pattern_destroy(p);
                return NULL;
            }
            if (pattern_add_literal(p, escaped) < 0) {
                set_error("Out of memory");
                wc_pattern_destroy(p);
                return NULL;
            }
            ptr++;
        } else if (*ptr == '*') {
            int star_count = 0;
            while (*ptr == '*') {
                star_count++;
                ptr++;
            }
            if (star_count >= 2) {
                if (pattern_add_starstar(p) < 0) {
                    set_error("Out of memory");
                    wc_pattern_destroy(p);
                    return NULL;
                }
            } else {
                if (pattern_add_star(p) < 0) {
                    set_error("Out of memory");
                    wc_pattern_destroy(p);
                    return NULL;
                }
            }
        } else if (*ptr == '?') {
            if (pattern_add_question(p) < 0) {
                set_error("Out of memory");
                wc_pattern_destroy(p);
                return NULL;
            }
            ptr++;
        } else if (*ptr == '[') {
            wc_char_class_t *cc = parse_char_class(&ptr);
            if (!cc) {
                wc_pattern_destroy(p);
                return NULL;
            }
            if (pattern_add_charclass(p, cc) < 0) {
                set_error("Out of memory");
                char_class_full_destroy(cc);
                wc_pattern_destroy(p);
                return NULL;
            }
            free(cc);
        } else {
            if (pattern_add_literal(p, *ptr) < 0) {
                set_error("Out of memory");
                wc_pattern_destroy(p);
                return NULL;
            }
            ptr++;
        }
    }
    
    last_error[0] = '\0';
    return p;
}
