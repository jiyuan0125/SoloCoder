#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "wildcard.h"

typedef struct {
    const char *str;
    size_t str_len;
    const wc_pattern_t *pattern;
    wc_case_mode_t case_mode;
    wc_match_result_t *result;
    const char *capture_text;
    size_t capture_stack[WC_MAX_CAPTURES];
    size_t capture_stack_top;
} match_context_t;

static int char_cmp(char a, char b, wc_case_mode_t case_mode) {
    if (case_mode == WC_CASE_INSENSITIVE) {
        return tolower((unsigned char)a) == tolower((unsigned char)b);
    }
    return a == b;
}

static bool char_class_matches(const wc_char_class_t *cc, char ch, wc_case_mode_t case_mode) {
    bool matched = false;
    
    for (size_t i = 0; i < cc->count; i++) {
        const wc_char_range_t *range = &cc->ranges[i];
        if (range->is_range) {
            char start = range->ch;
            char end = range->end;
            if (case_mode == WC_CASE_INSENSITIVE) {
                char lower_ch = tolower((unsigned char)ch);
                if (lower_ch >= tolower((unsigned char)start) && 
                    lower_ch <= tolower((unsigned char)end)) {
                    matched = true;
                    break;
                }
            } else {
                if (ch >= start && ch <= end) {
                    matched = true;
                    break;
                }
            }
        } else {
            if (char_cmp(ch, range->ch, case_mode)) {
                matched = true;
                break;
            }
        }
    }
    
    return cc->negated ? !matched : matched;
}

static void push_capture(match_context_t *ctx, size_t pos) {
    if (ctx->capture_stack_top < WC_MAX_CAPTURES) {
        ctx->capture_stack[ctx->capture_stack_top++] = pos;
    }
}

static size_t pop_capture(match_context_t *ctx) {
    if (ctx->capture_stack_top > 0) {
        return ctx->capture_stack[--ctx->capture_stack_top];
    }
    return 0;
}

static void add_capture_result(match_context_t *ctx, size_t start, size_t end) {
    if (ctx->result && ctx->result->count < WC_MAX_CAPTURES) {
        wc_capture_t *cap = &ctx->result->captures[ctx->result->count];
        cap->text = (char*)ctx->capture_text;
        cap->start = start;
        cap->end = end;
        ctx->result->count++;
    }
}

static bool match_recursive(match_context_t *ctx, size_t token_idx, size_t str_pos) {
    const wc_pattern_t *pattern = ctx->pattern;
    const char *str = ctx->str;
    
    if (token_idx >= pattern->count) {
        return str_pos >= ctx->str_len;
    }
    
    if (str_pos > ctx->str_len) {
        return false;
    }
    
    const wc_token_t *token = &pattern->tokens[token_idx];
    
    switch (token->type) {
        case WC_TOKEN_LITERAL: {
            if (str_pos >= ctx->str_len) return false;
            if (char_cmp(str[str_pos], token->data.literal, ctx->case_mode)) {
                return match_recursive(ctx, token_idx + 1, str_pos + 1);
            }
            return false;
        }
        
        case WC_TOKEN_QUESTION: {
            if (str_pos >= ctx->str_len) return false;
            return match_recursive(ctx, token_idx + 1, str_pos + 1);
        }
        
        case WC_TOKEN_CHARCLASS: {
            if (str_pos >= ctx->str_len) return false;
            if (char_class_matches(&token->data.char_class, str[str_pos], ctx->case_mode)) {
                return match_recursive(ctx, token_idx + 1, str_pos + 1);
            }
            return false;
        }
        
        case WC_TOKEN_STAR: {
            size_t next_token_idx = token_idx + 1;
            
            push_capture(ctx, str_pos);
            
            if (next_token_idx >= pattern->count) {
                for (size_t i = str_pos; i <= ctx->str_len; i++) {
                    if (str[i] == '/' && i < ctx->str_len) {
                        continue;
                    }
                    if (i == ctx->str_len || str[i] == '/') {
                        size_t cap_start = pop_capture(ctx);
                        add_capture_result(ctx, cap_start, i);
                        return true;
                    }
                }
                pop_capture(ctx);
                return false;
            }
            
            for (size_t i = str_pos; i <= ctx->str_len; i++) {
                if (i < ctx->str_len && str[i] == '/') {
                    break;
                }
                
                if (match_recursive(ctx, next_token_idx, i)) {
                    size_t cap_start = pop_capture(ctx);
                    add_capture_result(ctx, cap_start, i);
                    return true;
                }
            }
            
            pop_capture(ctx);
            return false;
        }
        
        case WC_TOKEN_STARSTAR: {
            size_t next_token_idx = token_idx + 1;
            bool next_is_slash = false;
            size_t skip_after_slash = next_token_idx;
            
            if (next_token_idx < pattern->count) {
                const wc_token_t *next_tok = &pattern->tokens[next_token_idx];
                if (next_tok->type == WC_TOKEN_LITERAL && next_tok->data.literal == '/') {
                    next_is_slash = true;
                    skip_after_slash = next_token_idx + 1;
                }
            }
            
            push_capture(ctx, str_pos);
            
            if (next_token_idx >= pattern->count) {
                size_t cap_start = pop_capture(ctx);
                add_capture_result(ctx, cap_start, ctx->str_len);
                return true;
            }
            
            if (next_is_slash) {
                if (match_recursive(ctx, skip_after_slash, str_pos)) {
                    size_t cap_start = pop_capture(ctx);
                    add_capture_result(ctx, cap_start, str_pos);
                    return true;
                }
            }
            
            for (size_t i = str_pos; i <= ctx->str_len; i++) {
                if (match_recursive(ctx, next_token_idx, i)) {
                    size_t cap_start = pop_capture(ctx);
                    add_capture_result(ctx, cap_start, i);
                    return true;
                }
            }
            
            pop_capture(ctx);
            return false;
        }
    }
    
    return false;
}

bool wc_match(const wc_pattern_t *pattern, const char *str, 
              wc_case_mode_t case_mode, wc_match_result_t *result) {
    if (!pattern || !str) return false;
    
    match_context_t ctx;
    ctx.str = str;
    ctx.str_len = strlen(str);
    ctx.pattern = pattern;
    ctx.case_mode = case_mode;
    ctx.result = result;
    ctx.capture_text = str;
    ctx.capture_stack_top = 0;
    
    if (result) {
        result->count = 0;
        memset(result->captures, 0, sizeof(result->captures));
    }
    
    return match_recursive(&ctx, 0, 0);
}

bool wc_match_with_capture(const wc_pattern_t *pattern, const char *str, 
                            wc_case_mode_t case_mode, wc_match_result_t *result) {
    if (!pattern || !str) return false;
    
    match_context_t ctx;
    ctx.str = str;
    ctx.str_len = strlen(str);
    ctx.pattern = pattern;
    ctx.case_mode = case_mode;
    ctx.result = result;
    ctx.capture_text = str;
    ctx.capture_stack_top = 0;
    
    if (result) {
        result->count = 0;
        memset(result->captures, 0, sizeof(result->captures));
    }
    
    bool matched = match_recursive(&ctx, 0, 0);
    
    if (!matched && result) {
        result->count = 0;
    }
    
    return matched;
}

void wc_match_result_free(wc_match_result_t *result) {
    (void)result;
}

const char* wc_capture_get_text(const wc_match_result_t *result, size_t index) {
    if (!result || index >= result->count) return NULL;
    
    const wc_capture_t *cap = &result->captures[index];
    if (!cap->text) return NULL;
    
    size_t len = cap->end - cap->start;
    if (len == 0) return "";
    
    char *buf = (char*)malloc(len + 1);
    if (!buf) return NULL;
    
    memcpy(buf, cap->text + cap->start, len);
    buf[len] = '\0';
    
    return buf;
}
