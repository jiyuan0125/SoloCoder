#ifndef PARSER_H
#define PARSER_H

#include "regex.h"

int regex_parser_init(Regex *re);
void regex_parser_free(Regex *re);
int regex_parser_add_token(Regex *re, const Token *tok);
int regex_parser_parse_charset(const char **pp, int *pos, Token *tok, int flags);
int regex_parser_validate_quantifier(TokenType prev_type, int pos, RegexError *err);

#endif
