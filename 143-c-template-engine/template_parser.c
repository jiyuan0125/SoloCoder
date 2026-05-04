#include "template_engine.h"
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

static char *str_trim(const char *str)
{
    if (!str) return NULL;
    
    while (isspace((unsigned char)*str)) str++;
    
    if (*str == '\0') {
        char *empty = (char *)malloc(1);
        if (empty) empty[0] = '\0';
        return empty;
    }
    
    const char *end = str + strlen(str) - 1;
    while (end > str && isspace((unsigned char)*end)) end--;
    
    size_t len = end - str + 1;
    char *result = (char *)malloc(len + 1);
    if (!result) return NULL;
    
    strncpy(result, str, len);
    result[len] = '\0';
    return result;
}

static char *str_dup(const char *str)
{
    if (!str) return NULL;
    size_t len = strlen(str);
    char *result = (char *)malloc(len + 1);
    if (result) {
        strcpy(result, str);
    }
    return result;
}

TE_Engine *te_engine_create(void)
{
    TE_Engine *engine = (TE_Engine *)malloc(sizeof(TE_Engine));
    if (!engine) return NULL;
    
    memset(engine, 0, sizeof(TE_Engine));
    
    engine->delimiters.open = str_dup("{{");
    engine->delimiters.close = str_dup("}}");
    engine->delimiters.open_triple = str_dup("{{{");
    engine->delimiters.close_triple = str_dup("}}}");
    
    if (!engine->delimiters.open || !engine->delimiters.close ||
        !engine->delimiters.open_triple || !engine->delimiters.close_triple) {
        te_engine_destroy(engine);
        return NULL;
    }
    
    engine->partials.names = NULL;
    engine->partials.contents = NULL;
    engine->partials.count = 0;
    engine->partials.capacity = 0;
    
    engine->error.message = NULL;
    engine->error.has_error = 0;
    
    return engine;
}

void te_engine_destroy(TE_Engine *engine)
{
    if (!engine) return;
    
    free(engine->delimiters.open);
    free(engine->delimiters.close);
    free(engine->delimiters.open_triple);
    free(engine->delimiters.close_triple);
    
    for (size_t i = 0; i < engine->partials.count; i++) {
        free(engine->partials.names[i]);
        free(engine->partials.contents[i]);
    }
    free(engine->partials.names);
    free(engine->partials.contents);
    
    free(engine->error.message);
    free(engine);
}

void te_engine_set_delimiters(TE_Engine *engine, const char *open, const char *close)
{
    if (!engine || !open || !close) return;
    
    free(engine->delimiters.open);
    free(engine->delimiters.close);
    free(engine->delimiters.open_triple);
    free(engine->delimiters.close_triple);
    
    engine->delimiters.open = str_dup(open);
    engine->delimiters.close = str_dup(close);
    
    size_t open_len = strlen(open);
    size_t close_len = strlen(close);
    engine->delimiters.open_triple = (char *)malloc(open_len + 2);
    engine->delimiters.close_triple = (char *)malloc(close_len + 2);
    
    if (engine->delimiters.open_triple) {
        strcpy(engine->delimiters.open_triple, open);
        engine->delimiters.open_triple[open_len] = open[open_len - 1];
        engine->delimiters.open_triple[open_len + 1] = '\0';
    }
    
    if (engine->delimiters.close_triple) {
        strcpy(engine->delimiters.close_triple, close);
        engine->delimiters.close_triple[close_len] = close[close_len - 1];
        engine->delimiters.close_triple[close_len + 1] = '\0';
    }
}

void te_engine_add_partial(TE_Engine *engine, const char *name, const char *content)
{
    if (!engine || !name || !content) return;
    
    for (size_t i = 0; i < engine->partials.count; i++) {
        if (strcmp(engine->partials.names[i], name) == 0) {
            free(engine->partials.contents[i]);
            engine->partials.contents[i] = str_dup(content);
            return;
        }
    }
    
    if (engine->partials.count >= engine->partials.capacity) {
        size_t new_cap = engine->partials.capacity == 0 ? 8 : engine->partials.capacity * 2;
        char **new_names = (char **)realloc(engine->partials.names, new_cap * sizeof(char *));
        char **new_contents = (char **)realloc(engine->partials.contents, new_cap * sizeof(char *));
        
        if (!new_names || !new_contents) {
            free(new_names);
            free(new_contents);
            return;
        }
        
        engine->partials.names = new_names;
        engine->partials.contents = new_contents;
        engine->partials.capacity = new_cap;
    }
    
    engine->partials.names[engine->partials.count] = str_dup(name);
    engine->partials.contents[engine->partials.count] = str_dup(content);
    engine->partials.count++;
}

static void te_set_error(TE_Engine *engine, const char *message)
{
    if (!engine) return;
    free(engine->error.message);
    engine->error.message = str_dup(message);
    engine->error.has_error = 1;
}

static TE_Token *te_token_create(TE_TokenType type, const char *content, size_t start, size_t end)
{
    TE_Token *token = (TE_Token *)malloc(sizeof(TE_Token));
    if (!token) return NULL;
    
    token->type = type;
    token->content = content ? str_dup(content) : NULL;
    token->start = start;
    token->end = end;
    token->next = NULL;
    token->children = NULL;
    token->else_children = NULL;
    
    return token;
}

void te_token_free(TE_Token *token)
{
    if (!token) return;
    
    free(token->content);
    te_token_free(token->children);
    te_token_free(token->else_children);
    te_token_free(token->next);
    free(token);
}

static const char *str_find(const char *haystack, const char *needle)
{
    if (!haystack || !needle) return NULL;
    size_t needle_len = strlen(needle);
    if (needle_len == 0) return haystack;
    
    for (; *haystack; haystack++) {
        if (strncmp(haystack, needle, needle_len) == 0) {
            return haystack;
        }
    }
    return NULL;
}

static int starts_with(const char *str, const char *prefix)
{
    if (!str || !prefix) return 0;
    size_t prefix_len = strlen(prefix);
    return strncmp(str, prefix, prefix_len) == 0;
}

static char *extract_between(const char *start, const char *open_delim, const char *close_delim, const char **end_ptr)
{
    size_t open_len = strlen(open_delim);
    size_t close_len = strlen(close_delim);
    
    if (!starts_with(start, open_delim)) {
        if (end_ptr) *end_ptr = start;
        return NULL;
    }
    
    const char *content_start = start + open_len;
    const char *close_pos = str_find(content_start, close_delim);
    
    if (!close_pos) {
        if (end_ptr) *end_ptr = start;
        return NULL;
    }
    
    size_t content_len = close_pos - content_start;
    char *content = (char *)malloc(content_len + 1);
    if (!content) {
        if (end_ptr) *end_ptr = start;
        return NULL;
    }
    
    strncpy(content, content_start, content_len);
    content[content_len] = '\0';
    
    if (end_ptr) *end_ptr = close_pos + close_len;
    return content;
}

static TE_Token *parse_tag(TE_Engine *engine, const char *tag_content, size_t start_pos, const char **end_ptr)
{
    (void)engine;
    
    char *trimmed = str_trim(tag_content);
    if (!trimmed) return NULL;
    
    TE_TokenType type;
    char *content = NULL;
    
    if (starts_with(trimmed, "#if ")) {
        type = TOKEN_IF;
        content = str_dup(trimmed + 4);
    } else if (strcmp(trimmed, "#if") == 0) {
        type = TOKEN_IF;
        content = str_dup("");
    } else if (starts_with(trimmed, "#unless ")) {
        type = TOKEN_UNLESS;
        content = str_dup(trimmed + 8);
    } else if (strcmp(trimmed, "#unless") == 0) {
        type = TOKEN_UNLESS;
        content = str_dup("");
    } else if (starts_with(trimmed, "#else")) {
        type = TOKEN_ELSE;
        content = str_dup("");
    } else if (strcmp(trimmed, "/if") == 0 || strcmp(trimmed, "/unless") == 0) {
        type = TOKEN_END_IF;
        content = str_dup("");
    } else if (starts_with(trimmed, "#each ")) {
        type = TOKEN_EACH;
        content = str_dup(trimmed + 6);
    } else if (strcmp(trimmed, "#each") == 0) {
        type = TOKEN_EACH;
        content = str_dup("");
    } else if (strcmp(trimmed, "/each") == 0) {
        type = TOKEN_END_EACH;
        content = str_dup("");
    } else if (starts_with(trimmed, "> ")) {
        type = TOKEN_PARTIAL;
        content = str_dup(trimmed + 2);
    } else if (starts_with(trimmed, ">")) {
        type = TOKEN_PARTIAL;
        content = str_dup(trimmed + 1);
    } else {
        type = TOKEN_VARIABLE;
        content = str_dup(trimmed);
    }
    
    free(trimmed);
    
    char *trimmed_content = str_trim(content);
    free(content);
    
    TE_Token *token = te_token_create(type, trimmed_content, start_pos, start_pos);
    free(trimmed_content);
    
    (void)end_ptr;
    return token;
}

typedef struct {
    TE_Token *token;
    int in_else;
} StackFrame;

typedef struct {
    StackFrame *stack;
    size_t size;
    size_t capacity;
} TokenStack;

static void stack_init(TokenStack *s)
{
    s->stack = NULL;
    s->size = 0;
    s->capacity = 0;
}

static void stack_push(TokenStack *s, TE_Token *token)
{
    if (s->size >= s->capacity) {
        size_t new_cap = s->capacity == 0 ? 16 : s->capacity * 2;
        StackFrame *new_stack = (StackFrame *)realloc(s->stack, new_cap * sizeof(StackFrame));
        if (!new_stack) return;
        s->stack = new_stack;
        s->capacity = new_cap;
    }
    s->stack[s->size].token = token;
    s->stack[s->size].in_else = 0;
    s->size++;
}

static TE_Token *stack_pop(TokenStack *s)
{
    if (s->size == 0) return NULL;
    s->size--;
    return s->stack[s->size].token;
}

static TE_Token *stack_top_token(TokenStack *s)
{
    if (s->size == 0) return NULL;
    return s->stack[s->size - 1].token;
}

static int stack_top_in_else(TokenStack *s)
{
    if (s->size == 0) return 0;
    return s->stack[s->size - 1].in_else;
}

static void stack_set_in_else(TokenStack *s, int in_else)
{
    if (s->size == 0) return;
    s->stack[s->size - 1].in_else = in_else;
}

static void stack_free(TokenStack *s)
{
    free(s->stack);
    s->stack = NULL;
    s->size = 0;
    s->capacity = 0;
}

static void add_child_token(TE_Token *parent, TE_Token *child)
{
    if (!parent || !child) return;
    
    if (!parent->children) {
        parent->children = child;
    } else {
        TE_Token *last = parent->children;
        while (last->next) {
            last = last->next;
        }
        last->next = child;
    }
}

static void add_else_child_token(TE_Token *parent, TE_Token *child)
{
    if (!parent || !child) return;
    
    if (!parent->else_children) {
        parent->else_children = child;
    } else {
        TE_Token *last = parent->else_children;
        while (last->next) {
            last = last->next;
        }
        last->next = child;
    }
}

static void add_child_by_flag(TokenStack *stack, TE_Token *child)
{
    if (!stack || !child) return;
    
    TE_Token *parent = stack_top_token(stack);
    if (!parent) return;
    
    if (stack_top_in_else(stack)) {
        add_else_child_token(parent, child);
    } else {
        add_child_token(parent, child);
    }
}

TE_Token *te_tokenize(TE_Engine *engine, const char *template_str)
{
    if (!engine || !template_str) {
        te_set_error(engine, "Invalid parameters");
        return NULL;
    }
    
    engine->error.has_error = 0;
    free(engine->error.message);
    engine->error.message = NULL;
    
    TE_Token *root = te_token_create(TOKEN_TEXT, "", 0, 0);
    if (!root) {
        te_set_error(engine, "Memory allocation failed");
        return NULL;
    }
    
    TokenStack stack;
    stack_init(&stack);
    stack_push(&stack, root);
    
    const char *pos = template_str;
    size_t idx = 0;
    size_t open_len = strlen(engine->delimiters.open);
    size_t open_triple_len = strlen(engine->delimiters.open_triple);
    size_t close_len = strlen(engine->delimiters.close);
    size_t close_triple_len = strlen(engine->delimiters.close_triple);
    
    while (*pos) {
        const char *next_open = str_find(pos, engine->delimiters.open);
        
        if (!next_open) {
            if (*pos) {
                TE_Token *text_token = te_token_create(TOKEN_TEXT, pos, idx, idx + strlen(pos));
                if (text_token) {
                    add_child_by_flag(&stack, text_token);
                }
            }
            break;
        }
        
        int is_triple = 0;
        const char *actual_open = engine->delimiters.open;
        size_t actual_open_len = open_len;
        const char *actual_close = engine->delimiters.close;
        size_t actual_close_len = close_len;
        
        if (starts_with(next_open, engine->delimiters.open_triple)) {
            is_triple = 1;
            actual_open = engine->delimiters.open_triple;
            actual_open_len = open_triple_len;
            actual_close = engine->delimiters.close_triple;
            actual_close_len = close_triple_len;
        }
        
        if (next_open > pos) {
            size_t text_len = next_open - pos;
            char *text_content = (char *)malloc(text_len + 1);
            if (text_content) {
                strncpy(text_content, pos, text_len);
                text_content[text_len] = '\0';
                TE_Token *text_token = te_token_create(TOKEN_TEXT, text_content, idx, idx + text_len);
                free(text_content);
                if (text_token) {
                    add_child_by_flag(&stack, text_token);
                }
            }
            idx += text_len;
        }
        
        const char *tag_end;
        char *tag_content = extract_between(next_open, actual_open, actual_close, &tag_end);
        
        if (!tag_content) {
            te_set_error(engine, "Unclosed tag");
            te_token_free(root);
            stack_free(&stack);
            return NULL;
        }
        
        size_t tag_start_idx = idx;
        idx += (tag_end - next_open);
        
        TE_Token *tag_token = NULL;
        
        if (is_triple) {
            char *trimmed = str_trim(tag_content);
            tag_token = te_token_create(TOKEN_VARIABLE_RAW, trimmed, tag_start_idx, idx);
            free(trimmed);
        } else {
            tag_token = parse_tag(engine, tag_content, tag_start_idx, NULL);
        }
        
        free(tag_content);
        
        if (!tag_token) {
            te_set_error(engine, "Failed to parse tag");
            te_token_free(root);
            stack_free(&stack);
            return NULL;
        }
        
        switch (tag_token->type) {
        case TOKEN_IF:
        case TOKEN_UNLESS:
        case TOKEN_EACH:
            add_child_by_flag(&stack, tag_token);
            stack_push(&stack, tag_token);
            break;
            
        case TOKEN_ELSE:
            te_token_free(tag_token);
            if (stack.size > 1) {
                TE_Token *top = stack_top_token(&stack);
                if (top->type == TOKEN_IF || top->type == TOKEN_UNLESS) {
                    stack_set_in_else(&stack, 1);
                }
            }
            break;
            
        case TOKEN_END_IF:
            te_token_free(tag_token);
            if (stack.size <= 1) {
                te_set_error(engine, "Mismatched /if");
                te_token_free(root);
                stack_free(&stack);
                return NULL;
            }
            stack_pop(&stack);
            break;
            
        case TOKEN_END_EACH:
            te_token_free(tag_token);
            if (stack.size <= 1) {
                te_set_error(engine, "Mismatched /each");
                te_token_free(root);
                stack_free(&stack);
                return NULL;
            }
            stack_pop(&stack);
            break;
            
        default:
            add_child_by_flag(&stack, tag_token);
            break;
        }
        
        pos = tag_end;
    }
    
    if (stack.size > 1) {
        te_set_error(engine, "Unclosed block");
        te_token_free(root);
        stack_free(&stack);
        return NULL;
    }
    
    stack_free(&stack);
    return root;
}
