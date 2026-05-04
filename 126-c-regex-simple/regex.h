#ifndef REGEX_H
#define REGEX_H

#include <stddef.h>

typedef enum {
    REGEX_OK = 0,
    REGEX_ERR_SYNTAX,
    REGEX_ERR_MEMORY,
    REGEX_ERR_UNMATCHED_BRACKET,
    REGEX_ERR_UNMATCHED_PAREN,
    REGEX_ERR_MISSING_PREV,
    REGEX_ERR_INVALID_RANGE,
    REGEX_ERR_INVALID_ESCAPE,
    REGEX_NO_MATCH
} RegexError;

typedef enum {
    TOKEN_CHAR,
    TOKEN_DOT,
    TOKEN_STAR,
    TOKEN_PLUS,
    TOKEN_QUESTION,
    TOKEN_CHARSET,
    TOKEN_NEG_CHARSET,
    TOKEN_CARET,
    TOKEN_DOLLAR,
    TOKEN_GROUP_START,
    TOKEN_GROUP_END,
    TOKEN_END
} TokenType;

typedef struct {
    TokenType type;
    char ch;
    int start;
    int end;
    char *charset;
    int charset_len;
} Token;

typedef struct {
    Token *tokens;
    int token_count;
    int token_capacity;
    int flags;
    int group_count;
} Regex;

typedef struct {
    int start;
    int length;
} RegexMatch;

#define REGEX_CASE_INSENSITIVE 0x01
#define REGEX_FULL_MATCH        0x02

Regex *regex_compile(const char *pattern, int flags, RegexError *err, int *err_pos);
void regex_free(Regex *re);
int regex_match(const Regex *re, const char *str, RegexMatch *match, int match_count);
int regex_search(const Regex *re, const char *str, RegexMatch *match, int match_count);
const char *regex_strerror(RegexError err);

#endif
