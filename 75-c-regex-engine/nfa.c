#include "regex.h"
#include <stdlib.h>
#include <string.h>

static int state_counter = 0;

NFAState *nfa_create_state(void) {
    NFAState *state = (NFAState *)malloc(sizeof(NFAState));
    if (state == NULL) return NULL;
    state->id = state_counter++;
    state->edges = NULL;
    state->is_accept = 0;
    state->capture_start_idx = -1;
    state->capture_end_idx = -1;
    return state;
}

void nfa_add_edge(NFAState *from, NFAEdgeType type, char ch, NFAState *to) {
    NFAEdge *edge = (NFAEdge *)malloc(sizeof(NFAEdge));
    if (edge == NULL) return;
    edge->type = type;
    edge->ch = ch;
    edge->to = to;
    edge->next = from->edges;
    from->edges = edge;
}

NFAFragment nfa_create_fragment(NFAState *start, NFAState *out) {
    NFAFragment frag;
    frag.start = start;
    frag.out = out;
    frag.out_type = 0;
    return frag;
}

NFAFragment nfa_char_fragment(char ch) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_CHAR, ch, out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_any_fragment(void) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_ANY, '\0', out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_epsilon_fragment(void) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_EPSILON, '\0', out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_bol_fragment(void) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_ASSERT_BOL, '\0', out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_eol_fragment(void) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_ASSERT_EOL, '\0', out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_concat(NFAFragment a, NFAFragment b) {
    nfa_add_edge(a.out, NFA_EPSILON, '\0', b.start);
    return nfa_create_fragment(a.start, b.out);
}

NFAFragment nfa_alt(NFAFragment a, NFAFragment b) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_EPSILON, '\0', a.start);
    nfa_add_edge(start, NFA_EPSILON, '\0', b.start);
    nfa_add_edge(a.out, NFA_EPSILON, '\0', out);
    nfa_add_edge(b.out, NFA_EPSILON, '\0', out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_star(NFAFragment frag) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_EPSILON, '\0', frag.start);
    nfa_add_edge(start, NFA_EPSILON, '\0', out);
    nfa_add_edge(frag.out, NFA_EPSILON, '\0', frag.start);
    nfa_add_edge(frag.out, NFA_EPSILON, '\0', out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_optional(NFAFragment frag) {
    NFAState *start = nfa_create_state();
    NFAState *out = nfa_create_state();
    nfa_add_edge(start, NFA_EPSILON, '\0', frag.start);
    nfa_add_edge(start, NFA_EPSILON, '\0', out);
    nfa_add_edge(frag.out, NFA_EPSILON, '\0', out);
    return nfa_create_fragment(start, out);
}

NFAFragment nfa_capture_start(NFAFragment frag, int group_idx) {
    NFAState *capture_start = nfa_create_state();
    capture_start->capture_start_idx = group_idx;
    nfa_add_edge(capture_start, NFA_EPSILON, '\0', frag.start);
    return nfa_create_fragment(capture_start, frag.out);
}

NFAFragment nfa_capture_end(NFAFragment frag, int group_idx) {
    NFAState *capture_end = nfa_create_state();
    capture_end->capture_end_idx = group_idx;
    nfa_add_edge(frag.out, NFA_EPSILON, '\0', capture_end);
    return nfa_create_fragment(frag.start, capture_end);
}

void nfa_reset_state_counter(void) {
    state_counter = 0;
}

static void nfa_free_state_recursive(NFAState *state, int **visited, int *count) {
    if (state == NULL) return;
    
    for (int i = 0; i < *count; i++) {
        if ((*visited)[i] == state->id) return;
    }
    
    *visited = (int *)realloc(*visited, (*count + 1) * sizeof(int));
    (*visited)[*count] = state->id;
    (*count)++;
    
    NFAEdge *edge = state->edges;
    while (edge != NULL) {
        NFAEdge *next = edge->next;
        nfa_free_state_recursive(edge->to, visited, count);
        free(edge);
        edge = next;
    }
    
    free(state);
}

void nfa_free(NFAState *start) {
    int *visited = NULL;
    int count = 0;
    nfa_free_state_recursive(start, &visited, &count);
    free(visited);
}
