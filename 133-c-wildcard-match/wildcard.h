#ifndef WILDCARD_H
#define WILDCARD_H

#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

#define WC_MAX_PATTERN_LEN    4096
#define WC_MAX_GROUPS         256
#define WC_MAX_CAPTURES       32
#define WC_MAX_RULES          1024
#define WC_MAX_PREFIX_LEN     64

typedef enum {
    WC_TOKEN_LITERAL,
    WC_TOKEN_STAR,
    WC_TOKEN_STARSTAR,
    WC_TOKEN_QUESTION,
    WC_TOKEN_CHARCLASS
} wc_token_type_t;

typedef enum {
    WC_CASE_SENSITIVE,
    WC_CASE_INSENSITIVE
} wc_case_mode_t;

typedef struct {
    char ch;
    char end;
    bool is_range;
} wc_char_range_t;

typedef struct {
    wc_char_range_t *ranges;
    size_t count;
    size_t capacity;
    bool negated;
} wc_char_class_t;

typedef struct {
    wc_token_type_t type;
    union {
        char literal;
        wc_char_class_t char_class;
    } data;
} wc_token_t;

typedef struct {
    wc_token_t *tokens;
    size_t count;
    size_t capacity;
    char *original_pattern;
    size_t wildcard_count;
    size_t starstar_count;
} wc_pattern_t;

typedef struct {
    char *text;
    size_t start;
    size_t end;
} wc_capture_t;

typedef struct {
    wc_capture_t captures[WC_MAX_CAPTURES];
    size_t count;
} wc_match_result_t;

typedef struct {
    wc_pattern_t pattern;
    int rule_index;
    void *user_data;
} wc_rule_t;

typedef struct {
    char prefix[WC_MAX_PREFIX_LEN + 1];
    int *rule_indices;
    size_t count;
    size_t capacity;
} wc_prefix_group_t;

typedef struct {
    wc_rule_t rules[WC_MAX_RULES];
    size_t rule_count;
    
    wc_prefix_group_t *groups;
    size_t group_count;
    size_t group_capacity;
    
    wc_prefix_group_t no_prefix_group;
    
    wc_case_mode_t case_mode;
} wc_batch_matcher_t;

wc_pattern_t* wc_pattern_create(const char *pattern_str, wc_case_mode_t case_mode);
void wc_pattern_destroy(wc_pattern_t *pattern);
const char* wc_pattern_get_error(void);

bool wc_match(const wc_pattern_t *pattern, const char *str, 
              wc_case_mode_t case_mode, wc_match_result_t *result);
bool wc_match_with_capture(const wc_pattern_t *pattern, const char *str, 
                            wc_case_mode_t case_mode, wc_match_result_t *result);
void wc_match_result_free(wc_match_result_t *result);
const char* wc_capture_get_text(const wc_match_result_t *result, size_t index);

wc_batch_matcher_t* wc_batch_matcher_create(wc_case_mode_t case_mode);
void wc_batch_matcher_destroy(wc_batch_matcher_t *matcher);
int wc_batch_matcher_add_rule(wc_batch_matcher_t *matcher, 
                               const char *pattern_str, void *user_data);
void wc_batch_matcher_build(wc_batch_matcher_t *matcher);

int* wc_batch_match(const wc_batch_matcher_t *matcher, const char *str, 
                     size_t *match_count);
void wc_batch_match_free(int *indices);

bool wc_match_rule(const wc_batch_matcher_t *matcher, const char *str, 
                   int rule_index, wc_match_result_t *result);

#ifdef __cplusplus
}
#endif

#endif
