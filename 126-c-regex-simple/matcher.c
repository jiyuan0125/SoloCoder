#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "matcher.h"
#include "regex.h"

int regex_matcher_is_quantifier(TokenType type) {
    return (type == TOKEN_STAR || type == TOKEN_PLUS || type == TOKEN_QUESTION);
}

int regex_matcher_get_min_repeats(TokenType type) {
    switch (type) {
        case TOKEN_STAR: return 0;
        case TOKEN_PLUS: return 1;
        case TOKEN_QUESTION: return 0;
        default: return 1;
    }
}

int regex_matcher_get_max_repeats(TokenType type) {
    switch (type) {
        case TOKEN_STAR: return 1000000;
        case TOKEN_PLUS: return 1000000;
        case TOKEN_QUESTION: return 1;
        default: return 1;
    }
}

int regex_matcher_char_match(const Regex *re, const Token *tok, char ch) {
    char c = ch;
    if (re->flags & REGEX_CASE_INSENSITIVE) {
        c = tolower(ch);
    }
    
    switch (tok->type) {
        case TOKEN_CHAR:
            return (c == tok->ch);
            
        case TOKEN_DOT:
            return (ch != '\0');
            
        case TOKEN_CHARSET:
            for (int i = 0; i < tok->charset_len; i++) {
                if (c == tok->charset[i]) {
                    return 1;
                }
            }
            return 0;
            
        case TOKEN_NEG_CHARSET:
            for (int i = 0; i < tok->charset_len; i++) {
                if (c == tok->charset[i]) {
                    return 0;
                }
            }
            return 1;
            
        default:
            return 0;
    }
}

static int get_quantifier_after(const Regex *re, int token_pos, TokenType *quant_type) {
    if (token_pos + 1 < re->token_count) {
        TokenType next_type = re->tokens[token_pos + 1].type;
        if (regex_matcher_is_quantifier(next_type)) {
            *quant_type = next_type;
            return 1;
        }
    }
    *quant_type = TOKEN_END;
    return 0;
}

int regex_matcher_match_at(const MatchContext *ctx, int str_pos, int token_pos,
                            int *match_len, int *groups, int max_groups) {
    const Regex *re = ctx->re;
    const char *str = ctx->str;
    int str_len = ctx->str_len;
    
    int total_matched = 0;
    int current_group = -1;
    
    while (token_pos < re->token_count) {
        Token *tok = &re->tokens[token_pos];
        
        if (tok->type == TOKEN_GROUP_START) {
            current_group++;
            if (current_group < max_groups) {
                groups[current_group * 2] = str_pos;
            }
            token_pos++;
            continue;
        }
        
        if (tok->type == TOKEN_GROUP_END) {
            if (current_group >= 0 && current_group < max_groups) {
                groups[current_group * 2 + 1] = str_pos;
            }
            current_group--;
            token_pos++;
            continue;
        }
        
        if (tok->type == TOKEN_CARET) {
            if (str_pos != 0) {
                return 0;
            }
            token_pos++;
            continue;
        }
        
        if (tok->type == TOKEN_DOLLAR) {
            if (str_pos != str_len) {
                return 0;
            }
            token_pos++;
            continue;
        }
        
        if (tok->type == TOKEN_END) {
            break;
        }
        
        TokenType quant_type;
        int has_quantifier = get_quantifier_after(re, token_pos, &quant_type);
        int min_repeats = has_quantifier ? regex_matcher_get_min_repeats(quant_type) : 1;
        int max_repeats = has_quantifier ? regex_matcher_get_max_repeats(quant_type) : 1;
        int next_token_pos = has_quantifier ? token_pos + 2 : token_pos + 1;
        
        int max_possible = 0;
        for (int i = 0; i < max_repeats; i++) {
            if (str_pos + i >= str_len) break;
            if (regex_matcher_char_match(re, tok, str[str_pos + i])) {
                max_possible++;
            } else {
                break;
            }
        }
        
        if (max_possible < min_repeats) {
            return 0;
        }
        
        if (!has_quantifier) {
            if (max_possible >= 1) {
                str_pos++;
                total_matched++;
                token_pos = next_token_pos;
            } else {
                return 0;
            }
            continue;
        }
        
        int matched = 0;
        int found = 0;
        
        for (int try_count = max_possible; try_count >= min_repeats; try_count--) {
            int new_str_pos = str_pos + try_count;
            int temp_groups[20];
            int i;
            
            for (i = 0; i < max_groups * 2; i++) {
                temp_groups[i] = groups[i];
            }
            
            MatchContext sub_ctx = *ctx;
            int sub_match_len = 0;
            
            if (regex_matcher_match_at(&sub_ctx, new_str_pos, next_token_pos, &sub_match_len, temp_groups, max_groups)) {
                matched = try_count;
                found = 1;
                
                for (i = 0; i < max_groups * 2; i++) {
                    groups[i] = temp_groups[i];
                }
                break;
            }
        }
        
        if (!found) {
            return 0;
        }
        
        str_pos += matched;
        total_matched += matched;
        token_pos = next_token_pos;
    }
    
    if (re->flags & REGEX_FULL_MATCH) {
        if (str_pos < str_len && re->tokens[token_pos - 1].type != TOKEN_DOLLAR) {
            return 0;
        }
    }
    
    if (match_len != NULL) {
        *match_len = total_matched;
    }
    
    return 1;
}

int regex_match(const Regex *re, const char *str, RegexMatch *match, int match_count) {
    int str_len = strlen(str);
    int groups[20] = {0};
    int match_len = 0;
    
    Regex full_match_re = *re;
    full_match_re.flags |= REGEX_FULL_MATCH;
    
    MatchContext ctx;
    ctx.re = &full_match_re;
    ctx.str = str;
    ctx.str_len = str_len;
    ctx.matches = match;
    ctx.max_matches = match_count;
    ctx.groups = NULL;
    ctx.group_count = re->group_count;
    
    int result = regex_matcher_match_at(&ctx, 0, 0, &match_len, groups, re->group_count);
    
    if (result && match != NULL && match_count > 0) {
        match[0].start = 0;
        match[0].length = match_len;
        
        for (int i = 0; i < re->group_count && i < match_count - 1; i++) {
            match[i + 1].start = groups[i * 2];
            match[i + 1].length = groups[i * 2 + 1] - groups[i * 2];
        }
    }
    
    return result;
}

int regex_search(const Regex *re, const char *str, RegexMatch *match, int match_count) {
    int str_len = strlen(str);
    
    for (int start_pos = 0; start_pos <= str_len; start_pos++) {
        int groups[20] = {0};
        int match_len = 0;
        
        MatchContext ctx;
        ctx.re = re;
        ctx.str = str;
        ctx.str_len = str_len;
        ctx.matches = match;
        ctx.max_matches = match_count;
        ctx.groups = NULL;
        ctx.group_count = re->group_count;
        
        int result = regex_matcher_match_at(&ctx, start_pos, 0, &match_len, groups, re->group_count);
        
        if (result) {
            if (match != NULL && match_count > 0) {
                match[0].start = start_pos;
                match[0].length = match_len;
                
                for (int i = 0; i < re->group_count && i < match_count - 1; i++) {
                    match[i + 1].start = groups[i * 2];
                    match[i + 1].length = groups[i * 2 + 1] - groups[i * 2];
                }
            }
            return 1;
        }
    }
    
    return 0;
}
