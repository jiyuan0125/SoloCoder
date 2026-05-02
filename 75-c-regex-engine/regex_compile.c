#include "regex.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

typedef struct {
    const char *pattern;
    int pos;
    int len;
    char error_msg[256];
    int error_pos;
    int capture_count;
} Parser;

static int parser_init(Parser *parser, const char *pattern) {
    parser->pattern = pattern;
    parser->pos = 0;
    parser->len = strlen(pattern);
    parser->error_msg[0] = '\0';
    parser->error_pos = -1;
    parser->capture_count = 0;
    return 0;
}

static int parser_peek(Parser *parser) {
    if (parser->pos >= parser->len) return -1;
    return parser->pattern[parser->pos];
}

static int parser_advance(Parser *parser) {
    if (parser->pos >= parser->len) return -1;
    return parser->pattern[parser->pos++];
}

static int parser_at_end(Parser *parser) {
    return parser->pos >= parser->len;
}

static void parser_set_error(Parser *parser, const char *msg, int pos) {
    strncpy(parser->error_msg, msg, sizeof(parser->error_msg) - 1);
    parser->error_msg[sizeof(parser->error_msg) - 1] = '\0';
    parser->error_pos = pos;
}

static NFAFragment parse_alternation(Parser *parser);
static NFAFragment parse_concatenation(Parser *parser);
static NFAFragment parse_quantifier(Parser *parser);
static NFAFragment parse_atom(Parser *parser);

static NFAFragment parse_atom(Parser *parser) {
    int ch = parser_peek(parser);
    
    if (ch == '(') {
        parser_advance(parser);
        
        if (parser->capture_count >= REGEX_MAX_CAPTURES) {
            parser_set_error(parser, "Too many capture groups", parser->pos - 1);
            return nfa_create_fragment(NULL, NULL);
        }
        
        int capture_idx = ++parser->capture_count;
        
        NFAFragment frag = parse_alternation(parser);
        if (frag.start == NULL) {
            return frag;
        }
        
        if (parser_peek(parser) != ')') {
            parser_set_error(parser, "Unmatched '('", parser->pos);
            return nfa_create_fragment(NULL, NULL);
        }
        parser_advance(parser);
        
        frag = nfa_capture_start(frag, capture_idx);
        frag = nfa_capture_end(frag, capture_idx);
        return frag;
    }
    
    if (ch == '.') {
        parser_advance(parser);
        return nfa_any_fragment();
    }
    
    if (ch == '^') {
        parser_advance(parser);
        return nfa_bol_fragment();
    }
    
    if (ch == '$') {
        parser_advance(parser);
        return nfa_eol_fragment();
    }
    
    if (ch == '|' || ch == ')' || ch == '*' || ch == '+' || ch == '?') {
        parser_set_error(parser, "Unexpected metacharacter", parser->pos);
        return nfa_create_fragment(NULL, NULL);
    }
    
    if (ch == -1) {
        parser_set_error(parser, "Unexpected end of pattern", parser->pos);
        return nfa_create_fragment(NULL, NULL);
    }
    
    parser_advance(parser);
    return nfa_char_fragment((char)ch);
}

static NFAFragment parse_quantifier(Parser *parser) {
    NFAFragment frag = parse_atom(parser);
    if (frag.start == NULL) {
        return frag;
    }
    
    int has_quantifier = 0;
    
    while (!parser_at_end(parser)) {
        int ch = parser_peek(parser);
        
        if (ch == '*' || ch == '+' || ch == '?') {
            if (has_quantifier) {
                parser_set_error(parser, "Quantifier has nothing to repeat", parser->pos);
                return nfa_create_fragment(NULL, NULL);
            }
            has_quantifier = 1;
        }
        
        if (ch == '*') {
            parser_advance(parser);
            frag = nfa_star(frag);
        } else if (ch == '+') {
            parser_advance(parser);
            NFAFragment frag_copy = frag;
            NFAFragment star_frag = nfa_star(frag_copy);
            frag = nfa_concat(frag, star_frag);
        } else if (ch == '?') {
            parser_advance(parser);
            frag = nfa_optional(frag);
        } else {
            break;
        }
    }
    
    return frag;
}

static NFAFragment parse_concatenation(Parser *parser) {
    NFAFragment result = parse_quantifier(parser);
    if (result.start == NULL) {
        return result;
    }
    
    while (!parser_at_end(parser)) {
        int ch = parser_peek(parser);
        
        if (ch == '|' || ch == ')') {
            break;
        }
        
        if (ch == '*' || ch == '+' || ch == '?') {
            parser_set_error(parser, "Quantifier has nothing to repeat", parser->pos);
            return nfa_create_fragment(NULL, NULL);
        }
        
        NFAFragment next = parse_quantifier(parser);
        if (next.start == NULL) {
            return next;
        }
        
        result = nfa_concat(result, next);
    }
    
    return result;
}

static NFAFragment parse_alternation(Parser *parser) {
    NFAFragment result = parse_concatenation(parser);
    if (result.start == NULL) {
        return result;
    }
    
    while (!parser_at_end(parser) && parser_peek(parser) == '|') {
        parser_advance(parser);
        
        if (parser_at_end(parser)) {
            parser_set_error(parser, "Alternation with no right-hand side", parser->pos);
            return nfa_create_fragment(NULL, NULL);
        }
        
        NFAFragment right = parse_concatenation(parser);
        if (right.start == NULL) {
            return right;
        }
        
        result = nfa_alt(result, right);
    }
    
    return result;
}

static void collect_states(NFAState *state, NFAState ***array, int *count, int **visited, int *visited_count) {
    if (state == NULL) return;
    
    for (int i = 0; i < *visited_count; i++) {
        if ((*visited)[i] == state->id) return;
    }
    
    *visited = (int *)realloc(*visited, (*visited_count + 1) * sizeof(int));
    (*visited)[*visited_count] = state->id;
    (*visited_count)++;
    
    *array = (NFAState **)realloc(*array, (*count + 1) * sizeof(NFAState *));
    (*array)[*count] = state;
    (*count)++;
    
    NFAEdge *edge = state->edges;
    while (edge != NULL) {
        collect_states(edge->to, array, count, visited, visited_count);
        edge = edge->next;
    }
}

int regex_compile(const char *pattern, Regex **out_regex) {
    if (pattern == NULL || out_regex == NULL) {
        return -1;
    }
    
    nfa_reset_state_counter();
    
    Parser parser;
    parser_init(&parser, pattern);
    
    if (parser.len == 0) {
        Regex *regex = (Regex *)malloc(sizeof(Regex));
        if (regex == NULL) return -1;
        
        NFAState *start = nfa_create_state();
        NFAState *accept = nfa_create_state();
        accept->is_accept = 1;
        nfa_add_edge(start, NFA_EPSILON, '\0', accept);
        
        regex->start = start;
        regex->states = (NFAState **)malloc(2 * sizeof(NFAState *));
        regex->states[0] = start;
        regex->states[1] = accept;
        regex->state_count = 2;
        regex->max_steps = REGEX_DEFAULT_MAX_STEPS;
        regex->capture_count = 0;
        regex->error_msg = NULL;
        regex->error_pos = -1;
        
        *out_regex = regex;
        return 0;
    }
    
    NFAFragment frag = parse_alternation(&parser);
    
    if (frag.start == NULL) {
        Regex *regex = (Regex *)malloc(sizeof(Regex));
        if (regex == NULL) return -1;
        regex->start = NULL;
        regex->states = NULL;
        regex->state_count = 0;
        regex->max_steps = REGEX_DEFAULT_MAX_STEPS;
        regex->capture_count = 0;
        regex->error_msg = strdup(parser.error_msg);
        regex->error_pos = parser.error_pos;
        *out_regex = regex;
        return -1;
    }
    
    if (!parser_at_end(&parser)) {
        int ch = parser_peek(&parser);
        if (ch == ')') {
            parser_set_error(&parser, "Unmatched ')'", parser.pos);
        } else {
            parser_set_error(&parser, "Unexpected character", parser.pos);
        }
        Regex *regex = (Regex *)malloc(sizeof(Regex));
        if (regex == NULL) return -1;
        regex->start = NULL;
        regex->states = NULL;
        regex->state_count = 0;
        regex->max_steps = REGEX_DEFAULT_MAX_STEPS;
        regex->capture_count = 0;
        regex->error_msg = strdup(parser.error_msg);
        regex->error_pos = parser.error_pos;
        *out_regex = regex;
        return -1;
    }
    
    NFAState *accept = nfa_create_state();
    accept->is_accept = 1;
    nfa_add_edge(frag.out, NFA_EPSILON, '\0', accept);
    
    Regex *regex = (Regex *)malloc(sizeof(Regex));
    if (regex == NULL) return -1;
    
    regex->start = frag.start;
    regex->state_count = 0;
    regex->states = NULL;
    
    int *visited = NULL;
    int visited_count = 0;
    collect_states(frag.start, &regex->states, &regex->state_count, &visited, &visited_count);
    free(visited);
    
    regex->max_steps = REGEX_DEFAULT_MAX_STEPS;
    regex->capture_count = parser.capture_count;
    regex->error_msg = NULL;
    regex->error_pos = -1;
    
    *out_regex = regex;
    return 0;
}
