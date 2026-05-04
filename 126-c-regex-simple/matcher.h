#ifndef MATCHER_H
#define MATCHER_H

#include "regex.h"

typedef struct {
    const Regex *re;
    const char *str;
    int str_len;
    RegexMatch *matches;
    int max_matches;
    int *groups;
    int group_count;
} MatchContext;

int regex_matcher_match_at(const MatchContext *ctx, int str_pos, int token_pos,
                            int *match_len, int *groups, int max_groups);
int regex_matcher_char_match(const Regex *re, const Token *tok, char ch);
int regex_matcher_is_quantifier(TokenType type);
int regex_matcher_get_min_repeats(TokenType type);
int regex_matcher_get_max_repeats(TokenType type);

#endif
