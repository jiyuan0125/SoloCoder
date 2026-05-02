#ifndef REGEX_H
#define REGEX_H

#include <stddef.h>

#define REGEX_MAX_CAPTURES 10
#define REGEX_DEFAULT_MAX_STEPS 1000000

typedef enum {
    NFA_EPSILON = 0,
    NFA_CHAR,
    NFA_ANY,
    NFA_ASSERT_BOL,
    NFA_ASSERT_EOL
} NFAEdgeType;

typedef struct NFAState NFAState;

typedef struct NFAEdge {
    NFAEdgeType type;
    char ch;
    NFAState *to;
    struct NFAEdge *next;
} NFAEdge;

struct NFAState {
    int id;
    NFAEdge *edges;
    int is_accept;
    int capture_start_idx;
    int capture_end_idx;
};

typedef struct NFAFragment {
    NFAState *start;
    NFAState *out;
    int out_type;
} NFAFragment;

typedef struct Capture {
    int start;
    int end;
} Capture;

typedef struct MatchResult {
    int found;
    int start;
    int end;
    int count;
    Capture captures[REGEX_MAX_CAPTURES + 1];
} MatchResult;

typedef struct Regex {
    NFAState *start;
    NFAState **states;
    int state_count;
    int max_steps;
    int capture_count;
    char *error_msg;
    int error_pos;
} Regex;

NFAState *nfa_create_state(void);
void nfa_add_edge(NFAState *from, NFAEdgeType type, char ch, NFAState *to);
NFAFragment nfa_create_fragment(NFAState *start, NFAState *out);
NFAFragment nfa_char_fragment(char ch);
NFAFragment nfa_any_fragment(void);
NFAFragment nfa_epsilon_fragment(void);
NFAFragment nfa_bol_fragment(void);
NFAFragment nfa_eol_fragment(void);
NFAFragment nfa_concat(NFAFragment a, NFAFragment b);
NFAFragment nfa_alt(NFAFragment a, NFAFragment b);
NFAFragment nfa_star(NFAFragment frag);
NFAFragment nfa_optional(NFAFragment frag);
NFAFragment nfa_capture_start(NFAFragment frag, int group_idx);
NFAFragment nfa_capture_end(NFAFragment frag, int group_idx);
void nfa_reset_state_counter(void);
void nfa_free(NFAState *start);

int regex_compile(const char *pattern, Regex **out_regex);
int regex_match(Regex *regex, const char *text, MatchResult *result);
void regex_free(Regex *regex);

#endif
