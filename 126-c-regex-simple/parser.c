#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "parser.h"

#define INITIAL_TOKEN_CAPACITY 16
#define INITIAL_CHARSET_CAPACITY 8

int regex_parser_init(Regex *re) {
    re->tokens = (Token *)malloc(INITIAL_TOKEN_CAPACITY * sizeof(Token));
    if (re->tokens == NULL) {
        return REGEX_ERR_MEMORY;
    }
    re->token_count = 0;
    re->token_capacity = INITIAL_TOKEN_CAPACITY;
    re->group_count = 0;
    return REGEX_OK;
}

void regex_parser_free(Regex *re) {
    if (re == NULL) return;
    
    for (int i = 0; i < re->token_count; i++) {
        if (re->tokens[i].charset != NULL) {
            free(re->tokens[i].charset);
            re->tokens[i].charset = NULL;
        }
    }
    
    if (re->tokens != NULL) {
        free(re->tokens);
        re->tokens = NULL;
    }
    
    re->token_count = 0;
    re->token_capacity = 0;
}

int regex_parser_add_token(Regex *re, const Token *tok) {
    if (re->token_count >= re->token_capacity) {
        int new_capacity = re->token_capacity * 2;
        Token *new_tokens = (Token *)realloc(re->tokens, new_capacity * sizeof(Token));
        if (new_tokens == NULL) {
            return REGEX_ERR_MEMORY;
        }
        re->tokens = new_tokens;
        re->token_capacity = new_capacity;
    }
    
    memcpy(&re->tokens[re->token_count], tok, sizeof(Token));
    re->token_count++;
    return REGEX_OK;
}

int regex_parser_parse_charset(const char **pp, int *pos, Token *tok, int flags) {
    const char *p = *pp;
    int charset_capacity = INITIAL_CHARSET_CAPACITY;
    int charset_len = 0;
    char *charset = (char *)malloc(charset_capacity * sizeof(char));
    
    if (charset == NULL) {
        return REGEX_ERR_MEMORY;
    }
    
    p++;
    (*pos)++;
    
    if (*p == '^') {
        tok->type = TOKEN_NEG_CHARSET;
        p++;
        (*pos)++;
    } else {
        tok->type = TOKEN_CHARSET;
    }
    
    while (*p != '\0' && *p != ']') {
        if (*p == '-' && charset_len > 0 && *(p+1) != '\0' && *(p+1) != ']') {
            char start = charset[charset_len - 1];
            char end = *(p + 1);
            
            if (start > end) {
                free(charset);
                return REGEX_ERR_INVALID_RANGE;
            }
            
            for (char c = start + 1; c <= end; c++) {
                if (charset_len >= charset_capacity - 1) {
                    charset_capacity *= 2;
                    char *new_charset = (char *)realloc(charset, charset_capacity * sizeof(char));
                    if (new_charset == NULL) {
                        free(charset);
                        return REGEX_ERR_MEMORY;
                    }
                    charset = new_charset;
                }
                charset[charset_len++] = c;
            }
            
            p += 2;
            (*pos) += 2;
        } else {
            if (charset_len >= charset_capacity - 1) {
                charset_capacity *= 2;
                char *new_charset = (char *)realloc(charset, charset_capacity * sizeof(char));
                if (new_charset == NULL) {
                    free(charset);
                    return REGEX_ERR_MEMORY;
                }
                charset = new_charset;
            }
            
            if (flags & REGEX_CASE_INSENSITIVE) {
                charset[charset_len++] = tolower(*p);
            } else {
                charset[charset_len++] = *p;
            }
            p++;
            (*pos)++;
        }
    }
    
    if (*p != ']') {
        free(charset);
        return REGEX_ERR_UNMATCHED_BRACKET;
    }
    p++;
    (*pos)++;
    
    charset[charset_len] = '\0';
    tok->charset = charset;
    tok->charset_len = charset_len;
    
    *pp = p;
    return REGEX_OK;
}

int regex_parser_validate_quantifier(TokenType prev_type, int pos, RegexError *err) {
    (void)pos;
    if (prev_type == TOKEN_END || 
        prev_type == TOKEN_CARET || 
        prev_type == TOKEN_DOLLAR ||
        prev_type == TOKEN_GROUP_START ||
        prev_type == TOKEN_STAR ||
        prev_type == TOKEN_PLUS ||
        prev_type == TOKEN_QUESTION) {
        *err = REGEX_ERR_MISSING_PREV;
        return 0;
    }
    return 1;
}

Regex *regex_compile(const char *pattern, int flags, RegexError *err, int *err_pos) {
    Regex *re = (Regex *)malloc(sizeof(Regex));
    if (re == NULL) {
        if (err != NULL) *err = REGEX_ERR_MEMORY;
        if (err_pos != NULL) *err_pos = 0;
        return NULL;
    }
    
    re->flags = flags;
    
    int ret = regex_parser_init(re);
    if (ret != REGEX_OK) {
        free(re);
        if (err != NULL) *err = ret;
        if (err_pos != NULL) *err_pos = 0;
        return NULL;
    }
    
    const char *p = pattern;
    int pos = 0;
    TokenType prev_type = TOKEN_END;
    int paren_count = 0;
    
    while (*p != '\0') {
        Token tok;
        memset(&tok, 0, sizeof(Token));
        
        switch (*p) {
            case '.':
                tok.type = TOKEN_DOT;
                prev_type = TOKEN_DOT;
                p++;
                pos++;
                break;
                
            case '*':
                if (!regex_parser_validate_quantifier(prev_type, pos, err)) {
                    regex_parser_free(re);
                    free(re);
                    if (err_pos != NULL) *err_pos = pos;
                    return NULL;
                }
                tok.type = TOKEN_STAR;
                prev_type = TOKEN_STAR;
                p++;
                pos++;
                break;
                
            case '+':
                if (!regex_parser_validate_quantifier(prev_type, pos, err)) {
                    regex_parser_free(re);
                    free(re);
                    if (err_pos != NULL) *err_pos = pos;
                    return NULL;
                }
                tok.type = TOKEN_PLUS;
                prev_type = TOKEN_PLUS;
                p++;
                pos++;
                break;
                
            case '?':
                if (!regex_parser_validate_quantifier(prev_type, pos, err)) {
                    regex_parser_free(re);
                    free(re);
                    if (err_pos != NULL) *err_pos = pos;
                    return NULL;
                }
                tok.type = TOKEN_QUESTION;
                prev_type = TOKEN_QUESTION;
                p++;
                pos++;
                break;
                
            case '[':
                ret = regex_parser_parse_charset(&p, &pos, &tok, flags);
                if (ret != REGEX_OK) {
                    regex_parser_free(re);
                    free(re);
                    if (err != NULL) *err = ret;
                    if (err_pos != NULL) *err_pos = pos;
                    return NULL;
                }
                if (tok.type == TOKEN_CHARSET) {
                    prev_type = TOKEN_CHARSET;
                } else {
                    prev_type = TOKEN_NEG_CHARSET;
                }
                break;
                
            case '^':
                tok.type = TOKEN_CARET;
                prev_type = TOKEN_CARET;
                p++;
                pos++;
                break;
                
            case '$':
                tok.type = TOKEN_DOLLAR;
                prev_type = TOKEN_DOLLAR;
                p++;
                pos++;
                break;
                
            case '(':
                tok.type = TOKEN_GROUP_START;
                prev_type = TOKEN_GROUP_START;
                paren_count++;
                re->group_count++;
                p++;
                pos++;
                break;
                
            case ')':
                if (paren_count == 0) {
                    regex_parser_free(re);
                    free(re);
                    if (err != NULL) *err = REGEX_ERR_UNMATCHED_PAREN;
                    if (err_pos != NULL) *err_pos = pos;
                    return NULL;
                }
                tok.type = TOKEN_GROUP_END;
                prev_type = TOKEN_GROUP_END;
                paren_count--;
                p++;
                pos++;
                break;
                
            case '\\':
                p++;
                pos++;
                if (*p == '\0') {
                    regex_parser_free(re);
                    free(re);
                    if (err != NULL) *err = REGEX_ERR_INVALID_ESCAPE;
                    if (err_pos != NULL) *err_pos = pos - 1;
                    return NULL;
                }
                tok.type = TOKEN_CHAR;
                if (flags & REGEX_CASE_INSENSITIVE) {
                    tok.ch = tolower(*p);
                } else {
                    tok.ch = *p;
                }
                prev_type = TOKEN_CHAR;
                p++;
                pos++;
                break;
                
            default:
                tok.type = TOKEN_CHAR;
                if (flags & REGEX_CASE_INSENSITIVE) {
                    tok.ch = tolower(*p);
                } else {
                    tok.ch = *p;
                }
                prev_type = TOKEN_CHAR;
                p++;
                pos++;
                break;
        }
        
        ret = regex_parser_add_token(re, &tok);
        if (ret != REGEX_OK) {
            regex_parser_free(re);
            free(re);
            if (err != NULL) *err = ret;
            if (err_pos != NULL) *err_pos = pos;
            return NULL;
        }
    }
    
    if (paren_count > 0) {
        regex_parser_free(re);
        free(re);
        if (err != NULL) *err = REGEX_ERR_UNMATCHED_PAREN;
        if (err_pos != NULL) *err_pos = pos;
        return NULL;
    }
    
    Token end_tok;
    memset(&end_tok, 0, sizeof(Token));
    end_tok.type = TOKEN_END;
    regex_parser_add_token(re, &end_tok);
    
    if (err != NULL) *err = REGEX_OK;
    return re;
}

void regex_free(Regex *re) {
    if (re == NULL) return;
    regex_parser_free(re);
    free(re);
}
