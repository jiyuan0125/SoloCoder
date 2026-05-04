#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "wildcard.h"

static int prefix_cmp(const char *prefix, const char *str, wc_case_mode_t case_mode) {
    if (case_mode == WC_CASE_INSENSITIVE) {
        while (*prefix != '\0') {
            if (tolower((unsigned char)*prefix) != tolower((unsigned char)*str)) {
                return 0;
            }
            prefix++;
            str++;
        }
        return 1;
    } else {
        while (*prefix != '\0') {
            if (*prefix != *str) {
                return 0;
            }
            prefix++;
            str++;
        }
        return 1;
    }
}

static int extract_pattern_prefix(const wc_pattern_t *pattern, char *prefix_buf, size_t buf_size) {
    if (pattern->count == 0) {
        if (buf_size > 0) {
            prefix_buf[0] = '\0';
        }
        return 0;
    }
    
    size_t prefix_len = 0;
    size_t i;
    
    for (i = 0; i < pattern->count && prefix_len < buf_size - 1; i++) {
        const wc_token_t *token = &pattern->tokens[i];
        
        if (token->type == WC_TOKEN_LITERAL) {
            prefix_buf[prefix_len++] = token->data.literal;
        } else if (token->type == WC_TOKEN_STAR || token->type == WC_TOKEN_STARSTAR ||
                   token->type == WC_TOKEN_QUESTION || token->type == WC_TOKEN_CHARCLASS) {
            break;
        }
    }
    
    prefix_buf[prefix_len] = '\0';
    return (int)prefix_len;
}

wc_batch_matcher_t* wc_batch_matcher_create(wc_case_mode_t case_mode) {
    wc_batch_matcher_t *matcher = (wc_batch_matcher_t*)malloc(sizeof(wc_batch_matcher_t));
    if (!matcher) return NULL;
    
    memset(matcher, 0, sizeof(wc_batch_matcher_t));
    matcher->case_mode = case_mode;
    matcher->no_prefix_group.prefix[0] = '\0';
    
    return matcher;
}

void wc_batch_matcher_destroy(wc_batch_matcher_t *matcher) {
    if (!matcher) return;
    
    for (size_t i = 0; i < matcher->rule_count; i++) {
        wc_pattern_destroy(&matcher->rules[i].pattern);
    }
    
    for (size_t i = 0; i < matcher->group_count; i++) {
        free(matcher->groups[i].rule_indices);
    }
    free(matcher->groups);
    free(matcher->no_prefix_group.rule_indices);
    
    free(matcher);
}

int wc_batch_matcher_add_rule(wc_batch_matcher_t *matcher, 
                               const char *pattern_str, void *user_data) {
    if (!matcher || !pattern_str) return -1;
    
    if (matcher->rule_count >= WC_MAX_RULES) {
        return -1;
    }
    
    wc_pattern_t *pattern = wc_pattern_create(pattern_str, matcher->case_mode);
    if (!pattern) {
        return -1;
    }
    
    int rule_idx = (int)matcher->rule_count;
    wc_rule_t *rule = &matcher->rules[matcher->rule_count++];
    
    memcpy(&rule->pattern, pattern, sizeof(wc_pattern_t));
    rule->rule_index = rule_idx;
    rule->user_data = user_data;
    
    free(pattern);
    
    return rule_idx;
}

static int find_or_create_group(wc_batch_matcher_t *matcher, const char *prefix) {
    if (prefix[0] == '\0') {
        return -1;
    }
    
    for (size_t i = 0; i < matcher->group_count; i++) {
        if (strcmp(matcher->groups[i].prefix, prefix) == 0) {
            return (int)i;
        }
    }
    
    if (matcher->group_count >= matcher->group_capacity) {
        size_t new_cap = matcher->group_capacity == 0 ? 16 : matcher->group_capacity * 2;
        wc_prefix_group_t *new_groups = (wc_prefix_group_t*)realloc(
            matcher->groups, new_cap * sizeof(wc_prefix_group_t));
        if (!new_groups) return -1;
        matcher->groups = new_groups;
        matcher->group_capacity = new_cap;
    }
    
    wc_prefix_group_t *group = &matcher->groups[matcher->group_count];
    memset(group, 0, sizeof(wc_prefix_group_t));
    strncpy(group->prefix, prefix, WC_MAX_PREFIX_LEN);
    group->prefix[WC_MAX_PREFIX_LEN] = '\0';
    
    return (int)matcher->group_count++;
}

static int add_rule_to_group(wc_prefix_group_t *group, int rule_index) {
    if (group->count >= group->capacity) {
        size_t new_cap = group->capacity == 0 ? 8 : group->capacity * 2;
        int *new_indices = (int*)realloc(group->rule_indices, new_cap * sizeof(int));
        if (!new_indices) return -1;
        group->rule_indices = new_indices;
        group->capacity = new_cap;
    }
    group->rule_indices[group->count++] = rule_index;
    return 0;
}

void wc_batch_matcher_build(wc_batch_matcher_t *matcher) {
    if (!matcher) return;
    
    for (size_t i = 0; i < matcher->group_count; i++) {
        free(matcher->groups[i].rule_indices);
    }
    free(matcher->groups);
    matcher->groups = NULL;
    matcher->group_count = 0;
    matcher->group_capacity = 0;
    
    free(matcher->no_prefix_group.rule_indices);
    matcher->no_prefix_group.rule_indices = NULL;
    matcher->no_prefix_group.count = 0;
    matcher->no_prefix_group.capacity = 0;
    
    for (size_t i = 0; i < matcher->rule_count; i++) {
        const wc_rule_t *rule = &matcher->rules[i];
        char prefix[WC_MAX_PREFIX_LEN + 1];
        
        int prefix_len = extract_pattern_prefix(&rule->pattern, prefix, sizeof(prefix));
        
        if (prefix_len > 0) {
            int group_idx = find_or_create_group(matcher, prefix);
            if (group_idx >= 0) {
                add_rule_to_group(&matcher->groups[group_idx], (int)i);
            } else {
                add_rule_to_group(&matcher->no_prefix_group, (int)i);
            }
        } else {
            add_rule_to_group(&matcher->no_prefix_group, (int)i);
        }
    }
}

static bool match_rule_internal(const wc_batch_matcher_t *matcher, 
                                 const wc_rule_t *rule, const char *str,
                                 wc_match_result_t *result) {
    return wc_match(&rule->pattern, str, matcher->case_mode, result);
}

int* wc_batch_match(const wc_batch_matcher_t *matcher, const char *str, 
                     size_t *match_count) {
    if (!matcher || !str || !match_count) {
        if (match_count) *match_count = 0;
        return NULL;
    }
    
    *match_count = 0;
    
    int *matches = (int*)malloc(matcher->rule_count * sizeof(int));
    if (!matches) return NULL;
    
    for (size_t i = 0; i < matcher->group_count; i++) {
        const wc_prefix_group_t *group = &matcher->groups[i];
        
        if (!prefix_cmp(group->prefix, str, matcher->case_mode)) {
            continue;
        }
        
        for (size_t j = 0; j < group->count; j++) {
            int rule_idx = group->rule_indices[j];
            const wc_rule_t *rule = &matcher->rules[rule_idx];
            
            if (match_rule_internal(matcher, rule, str, NULL)) {
                matches[*match_count] = rule_idx;
                (*match_count)++;
            }
        }
    }
    
    for (size_t j = 0; j < matcher->no_prefix_group.count; j++) {
        int rule_idx = matcher->no_prefix_group.rule_indices[j];
        const wc_rule_t *rule = &matcher->rules[rule_idx];
        
        if (match_rule_internal(matcher, rule, str, NULL)) {
            matches[*match_count] = rule_idx;
            (*match_count)++;
        }
    }
    
    if (*match_count == 0) {
        free(matches);
        return NULL;
    }
    
    int *result = (int*)realloc(matches, *match_count * sizeof(int));
    return result ? result : matches;
}

void wc_batch_match_free(int *indices) {
    free(indices);
}

bool wc_match_rule(const wc_batch_matcher_t *matcher, const char *str, 
                   int rule_index, wc_match_result_t *result) {
    if (!matcher || !str || rule_index < 0 || 
        (size_t)rule_index >= matcher->rule_count) {
        return false;
    }
    
    const wc_rule_t *rule = &matcher->rules[rule_index];
    return match_rule_internal(matcher, rule, str, result);
}
