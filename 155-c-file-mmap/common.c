#include "common.h"

void init_match_result(MatchResult *result) {
    result->matches = NULL;
    result->match_count = 0;
    result->allocated_count = 0;
}

void free_match_result(MatchResult *result) {
    if (result->matches) {
        free(result->matches);
        result->matches = NULL;
    }
    result->match_count = 0;
    result->allocated_count = 0;
}

int add_match_to_result(MatchResult *result, const MatchEntry *entry) {
    if (result->match_count >= result->allocated_count) {
        int new_size = result->allocated_count == 0 ? 16 : result->allocated_count * 2;
        MatchEntry *new_matches = realloc(result->matches, new_size * sizeof(MatchEntry));
        if (new_matches == NULL) {
            return -1;
        }
        result->matches = new_matches;
        result->allocated_count = new_size;
    }
    memcpy(&result->matches[result->match_count], entry, sizeof(MatchEntry));
    result->match_count++;
    return 0;
}

char *safe_strdup(const char *str) {
    if (str == NULL) return NULL;
    size_t len = strlen(str);
    char *dup = malloc(len + 1);
    if (dup == NULL) return NULL;
    strcpy(dup, str);
    return dup;
}

char *trim_newline(char *str) {
    if (str == NULL) return NULL;
    size_t len = strlen(str);
    while (len > 0 && (str[len - 1] == '\n' || str[len - 1] == '\r')) {
        str[len - 1] = '\0';
        len--;
    }
    return str;
}

static int is_regex_special_char(char c) {
    const char *special = ".^$*+?()[{\\|";
    return strchr(special, c) != NULL;
}

char *escape_regex_chars(const char *str) {
    if (str == NULL) return NULL;
    
    size_t src_len = strlen(str);
    size_t max_dst_len = src_len * 2 + 1;
    char *result = malloc(max_dst_len);
    if (result == NULL) return NULL;
    
    char *dst = result;
    const char *src = str;
    
    while (*src) {
        if (is_regex_special_char(*src)) {
            *dst++ = '\\';
        }
        *dst++ = *src++;
    }
    *dst = '\0';
    
    return result;
}
