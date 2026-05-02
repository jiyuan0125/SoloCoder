#include "regex.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

typedef struct {
    int *state_ids;
    int count;
    int capacity;
    Capture captures[REGEX_MAX_CAPTURES + 1];
} StateSet;

static StateSet *state_set_create(int max_states) {
    StateSet *set = (StateSet *)malloc(sizeof(StateSet));
    if (set == NULL) return NULL;
    set->state_ids = (int *)malloc(max_states * sizeof(int));
    if (set->state_ids == NULL) {
        free(set);
        return NULL;
    }
    set->count = 0;
    set->capacity = max_states;
    memset(set->captures, -1, sizeof(set->captures));
    return set;
}

static void state_set_free(StateSet *set) {
    if (set == NULL) return;
    if (set->state_ids != NULL) {
        free(set->state_ids);
    }
    free(set);
}

static void state_set_clear(StateSet *set) {
    set->count = 0;
    memset(set->captures, -1, sizeof(set->captures));
}

static int state_set_contains(StateSet *set, int state_id) {
    for (int i = 0; i < set->count; i++) {
        if (set->state_ids[i] == state_id) return 1;
    }
    return 0;
}

static void state_set_add(StateSet *set, int state_id, Capture captures[REGEX_MAX_CAPTURES + 1]) {
    if (state_set_contains(set, state_id)) {
        for (int i = 0; i <= REGEX_MAX_CAPTURES; i++) {
            if (set->captures[i].start == -1 && captures[i].start != -1) {
                set->captures[i].start = captures[i].start;
            }
            if (set->captures[i].end == -1 && captures[i].end != -1) {
                set->captures[i].end = captures[i].end;
            }
        }
        return;
    }
    
    if (set->count < set->capacity) {
        set->state_ids[set->count] = state_id;
        for (int i = 0; i <= REGEX_MAX_CAPTURES; i++) {
            if (captures[i].start != -1) {
                set->captures[i].start = captures[i].start;
            }
            if (captures[i].end != -1) {
                set->captures[i].end = captures[i].end;
            }
        }
        set->count++;
    }
}

static void state_set_copy(StateSet *dst, StateSet *src) {
    dst->count = src->count;
    memcpy(dst->state_ids, src->state_ids, src->count * sizeof(int));
    memcpy(dst->captures, src->captures, sizeof(src->captures));
}

static NFAState *find_state_by_id(Regex *regex, int id) {
    for (int i = 0; i < regex->state_count; i++) {
        if (regex->states[i]->id == id) {
            return regex->states[i];
        }
    }
    return NULL;
}

static int find_state_index(Regex *regex, int state_id) {
    for (int i = 0; i < regex->state_count; i++) {
        if (regex->states[i]->id == state_id) {
            return i;
        }
    }
    return -1;
}

static void epsilon_closure(Regex *regex, StateSet *set, int text_pos) {
    if (regex->state_count == 0) return;
    
    int *stack = (int *)malloc(regex->state_count * sizeof(int));
    Capture *stack_captures = (Capture *)malloc(regex->state_count * (REGEX_MAX_CAPTURES + 1) * sizeof(Capture));
    
    if (stack == NULL || stack_captures == NULL) {
        free(stack);
        free(stack_captures);
        return;
    }
    
    int stack_top = 0;
    
    for (int i = 0; i < set->count; i++) {
        if (stack_top >= regex->state_count) break;
        stack[stack_top] = set->state_ids[i];
        memcpy(&stack_captures[stack_top * (REGEX_MAX_CAPTURES + 1)], set->captures, sizeof(set->captures));
        stack_top++;
    }
    
    while (stack_top > 0) {
        stack_top--;
        int state_id = stack[stack_top];
        Capture current_captures[REGEX_MAX_CAPTURES + 1];
        memcpy(current_captures, &stack_captures[stack_top * (REGEX_MAX_CAPTURES + 1)], sizeof(current_captures));
        
        NFAState *state = find_state_by_id(regex, state_id);
        if (state == NULL) continue;
        
        if (state->capture_start_idx >= 0 && state->capture_start_idx <= REGEX_MAX_CAPTURES) {
            if (current_captures[state->capture_start_idx].start == -1) {
                current_captures[state->capture_start_idx].start = text_pos;
            }
        }
        
        if (state->capture_end_idx >= 0 && state->capture_end_idx <= REGEX_MAX_CAPTURES) {
            if (current_captures[state->capture_end_idx].end == -1) {
                current_captures[state->capture_end_idx].end = text_pos;
            }
        }
        
        NFAEdge *edge = state->edges;
        while (edge != NULL) {
            if (edge->type == NFA_EPSILON && edge->to != NULL) {
                int to_id = edge->to->id;
                if (!state_set_contains(set, to_id)) {
                    state_set_add(set, to_id, current_captures);
                    if (stack_top < regex->state_count) {
                        stack[stack_top] = to_id;
                        memcpy(&stack_captures[stack_top * (REGEX_MAX_CAPTURES + 1)], current_captures, sizeof(current_captures));
                        stack_top++;
                    }
                }
            }
            edge = edge->next;
        }
    }
    
    free(stack);
    free(stack_captures);
}

static void process_anchors(Regex *regex, StateSet *set, const char *text, int pos, int len) {
    if (regex->state_count == 0) return;
    
    int *stack = (int *)malloc(regex->state_count * sizeof(int));
    int *visited = (int *)malloc(regex->state_count * sizeof(int));
    
    if (stack == NULL || visited == NULL) {
        free(stack);
        free(visited);
        return;
    }
    
    int stack_top = 0;
    int visited_count = 0;
    
    for (int i = 0; i < set->count; i++) {
        int already_visited = 0;
        for (int j = 0; j < visited_count; j++) {
            if (visited[j] == set->state_ids[i]) {
                already_visited = 1;
                break;
            }
        }
        if (!already_visited && stack_top < regex->state_count) {
            stack[stack_top++] = set->state_ids[i];
            if (visited_count < regex->state_count) {
                visited[visited_count++] = set->state_ids[i];
            }
        }
    }
    
    while (stack_top > 0) {
        int state_id = stack[--stack_top];
        
        NFAState *state = find_state_by_id(regex, state_id);
        if (state == NULL) continue;
        
        NFAEdge *edge = state->edges;
        while (edge != NULL) {
            int matches = 0;
            
            if (edge->type == NFA_ASSERT_BOL) {
                if (pos == 0) {
                    matches = 1;
                } else if (pos > 0 && text[pos - 1] == '\n') {
                    matches = 1;
                }
            } else if (edge->type == NFA_ASSERT_EOL) {
                if (pos == len) {
                    matches = 1;
                } else if (pos < len && text[pos] == '\n') {
                    matches = 1;
                }
            }
            
            if (matches && edge->to != NULL) {
                int to_id = edge->to->id;
                if (!state_set_contains(set, to_id)) {
                    state_set_add(set, to_id, set->captures);
                    
                    int already_visited = 0;
                    for (int j = 0; j < visited_count; j++) {
                        if (visited[j] == to_id) {
                            already_visited = 1;
                            break;
                        }
                    }
                    if (!already_visited && stack_top < regex->state_count) {
                        stack[stack_top++] = to_id;
                        if (visited_count < regex->state_count) {
                            visited[visited_count++] = to_id;
                        }
                    }
                }
            }
            
            edge = edge->next;
        }
    }
    
    free(stack);
    free(visited);
}

static void move(Regex *regex, StateSet *current, char ch, StateSet *next) {
    state_set_clear(next);
    
    for (int i = 0; i < current->count; i++) {
        int state_id = current->state_ids[i];
        NFAState *state = find_state_by_id(regex, state_id);
        if (state == NULL) continue;
        
        NFAEdge *edge = state->edges;
        while (edge != NULL) {
            int matches = 0;
            
            if (edge->type == NFA_CHAR) {
                if (edge->ch == ch) matches = 1;
            } else if (edge->type == NFA_ANY) {
                if (ch != '\n') matches = 1;
            }
            
            if (matches && edge->to != NULL) {
                state_set_add(next, edge->to->id, current->captures);
            }
            
            edge = edge->next;
        }
    }
}

static int has_accept_state(Regex *regex, StateSet *set) {
    for (int i = 0; i < set->count; i++) {
        NFAState *state = find_state_by_id(regex, set->state_ids[i]);
        if (state != NULL && state->is_accept) {
            return 1;
        }
    }
    return 0;
}

static int get_accept_state_captures(Regex *regex, StateSet *set, Capture captures[REGEX_MAX_CAPTURES + 1]) {
    for (int i = 0; i < set->count; i++) {
        NFAState *state = find_state_by_id(regex, set->state_ids[i]);
        if (state != NULL && state->is_accept) {
            memcpy(captures, set->captures, sizeof(set->captures));
            return 1;
        }
    }
    return 0;
}

int regex_match(Regex *regex, const char *text, MatchResult *result) {
    if (regex == NULL || text == NULL || result == NULL) {
        return 0;
    }
    
    if (regex->start == NULL || regex->state_count == 0) {
        return 0;
    }
    
    int text_len = strlen(text);
    int steps = 0;
    
    memset(result, 0, sizeof(MatchResult));
    memset(result->captures, -1, sizeof(result->captures));
    
    StateSet *current = state_set_create(regex->state_count);
    StateSet *next = state_set_create(regex->state_count);
    StateSet *best_capture = state_set_create(regex->state_count);
    
    if (current == NULL || next == NULL || best_capture == NULL) {
        state_set_free(current);
        state_set_free(next);
        state_set_free(best_capture);
        return 0;
    }
    
    for (int start_pos = 0; start_pos <= text_len; start_pos++) {
        state_set_clear(current);
        
        Capture initial_captures[REGEX_MAX_CAPTURES + 1];
        memset(initial_captures, -1, sizeof(initial_captures));
        state_set_add(current, regex->start->id, initial_captures);
        
        epsilon_closure(regex, current, start_pos);
        process_anchors(regex, current, text, start_pos, text_len);
        epsilon_closure(regex, current, start_pos);
        
        int best_pos = -1;
        state_set_clear(best_capture);
        
        if (has_accept_state(regex, current)) {
            Capture captures[REGEX_MAX_CAPTURES + 1];
            if (get_accept_state_captures(regex, current, captures)) {
                best_pos = start_pos;
                state_set_copy(best_capture, current);
                memcpy(best_capture->captures, captures, sizeof(captures));
            }
        }
        
        for (int pos = start_pos; pos < text_len; pos++) {
            if (steps >= regex->max_steps) {
                state_set_free(current);
                state_set_free(next);
                state_set_free(best_capture);
                return -1;
            }
            steps++;
            
            char ch = text[pos];
            move(regex, current, ch, next);
            
            if (next->count == 0) {
                break;
            }
            
            epsilon_closure(regex, next, pos + 1);
            process_anchors(regex, next, text, pos + 1, text_len);
            epsilon_closure(regex, next, pos + 1);
            
            if (has_accept_state(regex, next)) {
                Capture captures[REGEX_MAX_CAPTURES + 1];
                if (get_accept_state_captures(regex, next, captures)) {
                    best_pos = pos + 1;
                    state_set_copy(best_capture, next);
                    memcpy(best_capture->captures, captures, sizeof(captures));
                }
            }
            
            StateSet *temp = current;
            current = next;
            next = temp;
            state_set_clear(next);
        }
        
        if (best_pos >= 0) {
            result->found = 1;
            result->start = start_pos;
            result->end = best_pos;
            result->count = 0;
            
            for (int i = 0; i <= REGEX_MAX_CAPTURES; i++) {
                result->captures[i] = best_capture->captures[i];
                if (best_capture->captures[i].start != -1) {
                    result->count = i + 1;
                }
            }
            
            result->captures[0].start = start_pos;
            result->captures[0].end = best_pos;
            if (result->count == 0) {
                result->count = 1;
            }
            
            state_set_free(current);
            state_set_free(next);
            state_set_free(best_capture);
            return 1;
        }
    }
    
    state_set_free(current);
    state_set_free(next);
    state_set_free(best_capture);
    return 0;
}

void regex_free(Regex *regex) {
    if (regex == NULL) return;
    
    if (regex->error_msg != NULL) {
        free(regex->error_msg);
    }
    
    if (regex->states != NULL) {
        int *visited = (int *)calloc(regex->state_count, sizeof(int));
        if (visited != NULL) {
            NFAState **stack = (NFAState **)malloc(regex->state_count * sizeof(NFAState *));
            if (stack != NULL) {
                int stack_top = 0;
                
                if (regex->start != NULL) {
                    stack[stack_top++] = regex->start;
                }
                
                while (stack_top > 0) {
                    NFAState *s = stack[--stack_top];
                    
                    if (s == NULL) continue;
                    
                    int already_visited = 0;
                    for (int i = 0; i < regex->state_count; i++) {
                        if (visited[i] == s->id) {
                            already_visited = 1;
                            break;
                        }
                    }
                    if (already_visited) continue;
                    
                    for (int i = 0; i < regex->state_count; i++) {
                        if (visited[i] == 0) {
                            visited[i] = s->id;
                            break;
                        }
                    }
                    
                    NFAEdge *edge = s->edges;
                    while (edge != NULL) {
                        if (edge->to != NULL && stack_top < regex->state_count) {
                            stack[stack_top++] = edge->to;
                        }
                        NFAEdge *next = edge->next;
                        free(edge);
                        edge = next;
                    }
                    s->edges = NULL;
                }
                
                for (int i = 0; i < regex->state_count; i++) {
                    if (regex->states[i] != NULL) {
                        free(regex->states[i]);
                    }
                }
                
                free(stack);
            }
            free(visited);
        }
        
        free(regex->states);
    }
    
    free(regex);
}
