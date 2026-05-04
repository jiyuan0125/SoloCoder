#include "template_engine.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

#define INITIAL_BUFFER_SIZE 1024

typedef struct {
    char *data;
    size_t length;
    size_t capacity;
} RenderBuffer;

static void buffer_init(RenderBuffer *buf)
{
    buf->data = (char *)malloc(INITIAL_BUFFER_SIZE);
    buf->length = 0;
    buf->capacity = buf->data ? INITIAL_BUFFER_SIZE : 0;
}

static void buffer_ensure(RenderBuffer *buf, size_t needed)
{
    if (buf->length + needed >= buf->capacity) {
        size_t new_cap = buf->capacity * 2;
        while (new_cap <= buf->length + needed) {
            new_cap *= 2;
        }
        char *new_data = (char *)realloc(buf->data, new_cap);
        if (new_data) {
            buf->data = new_data;
            buf->capacity = new_cap;
        }
    }
}

static void buffer_append(RenderBuffer *buf, const char *str)
{
    if (!str || !buf->data) return;
    size_t len = strlen(str);
    buffer_ensure(buf, len + 1);
    if (buf->data) {
        strcpy(buf->data + buf->length, str);
        buf->length += len;
    }
}

static void buffer_append_char(RenderBuffer *buf, char c)
{
    buffer_ensure(buf, 2);
    if (buf->data) {
        buf->data[buf->length++] = c;
        buf->data[buf->length] = '\0';
    }
}

static void buffer_free(RenderBuffer *buf)
{
    free(buf->data);
    buf->data = NULL;
    buf->length = 0;
    buf->capacity = 0;
}

char *te_html_escape(const char *input)
{
    if (!input) return NULL;
    
    RenderBuffer buf;
    buffer_init(&buf);
    if (!buf.data) return NULL;
    
    for (const char *p = input; *p; p++) {
        switch (*p) {
        case '<':
            buffer_append(&buf, "&lt;");
            break;
        case '>':
            buffer_append(&buf, "&gt;");
            break;
        case '&':
            buffer_append(&buf, "&amp;");
            break;
        case '"':
            buffer_append(&buf, "&quot;");
            break;
        case '\'':
            buffer_append(&buf, "&#39;");
            break;
        default:
            buffer_append_char(&buf, *p);
            break;
        }
    }
    
    return buf.data;
}

static char *value_to_string(const TE_Value *value)
{
    if (!value) return NULL;
    
    switch (value->type) {
    case TE_TYPE_NULL:
        return NULL;
    case TE_TYPE_BOOL:
        return value->data.bool_val ? strdup("true") : strdup("false");
    case TE_TYPE_INT: {
        char buf[64];
        snprintf(buf, sizeof(buf), "%lld", value->data.int_val);
        return strdup(buf);
    }
    case TE_TYPE_DOUBLE: {
        char buf[64];
        snprintf(buf, sizeof(buf), "%g", value->data.double_val);
        return strdup(buf);
    }
    case TE_TYPE_STRING:
        return value->data.string_val ? strdup(value->data.string_val) : NULL;
    case TE_TYPE_OBJECT:
        return strdup("[object]");
    case TE_TYPE_ARRAY:
        return strdup("[array]");
    default:
        return NULL;
    }
}

static TE_Value *resolve_variable(const TE_Value *context, const char *path)
{
    if (!context || !path || !*path) return NULL;
    
    const TE_Value *current = context;
    char *path_copy = strdup(path);
    if (!path_copy) return NULL;
    
    char *saveptr = NULL;
    char *token = strtok_r(path_copy, ".", &saveptr);
    
    while (token && current) {
        if (strcmp(token, "this") == 0) {
            token = strtok_r(NULL, ".", &saveptr);
            continue;
        }
        
        if (current->type == TE_TYPE_OBJECT) {
            current = te_object_get(current, token);
        } else if (current->type == TE_TYPE_ARRAY) {
            if (isdigit((unsigned char)*token)) {
                size_t idx = (size_t)atoi(token);
                current = te_array_get(current, idx);
            } else {
                current = NULL;
            }
        } else {
            current = NULL;
        }
        
        token = strtok_r(NULL, ".", &saveptr);
    }
    
    free(path_copy);
    return (TE_Value *)current;
}

static TE_Value *lookup_special_var(const TE_Value *context, const char *name, size_t index)
{
    if (strcmp(name, "@index") == 0) {
        return te_value_int((long long)index);
    }
    if (strcmp(name, "this") == 0 || strcmp(name, ".") == 0) {
        return (TE_Value *)context;
    }
    return NULL;
}

static void render_tokens(TE_Engine *engine, TE_Token *tokens, const TE_Value *context,
                          size_t loop_index, const TE_Value *loop_item, RenderBuffer *buf);

static void render_variable(TE_Engine *engine, const char *var_name, const TE_Value *context,
                             size_t loop_index, const TE_Value *loop_item,
                             int escape, RenderBuffer *buf)
{
    (void)engine;
    
    TE_Value *value = NULL;
    
    if (loop_item) {
        value = lookup_special_var(loop_item, var_name, loop_index);
        if (value && strcmp(var_name, "@index") == 0) {
            char *str_val = value_to_string(value);
            te_value_free(value);
            if (str_val) {
                buffer_append(buf, str_val);
                free(str_val);
            }
            return;
        }
        if (value && (strcmp(var_name, "this") == 0 || strcmp(var_name, ".") == 0)) {
        } else {
            value = resolve_variable(loop_item, var_name);
        }
    }
    
    if (!value && context) {
        value = resolve_variable(context, var_name);
    }
    
    if (!value) {
        return;
    }
    
    char *str_val = value_to_string(value);
    if (!str_val) {
        return;
    }
    
    if (escape) {
        char *escaped = te_html_escape(str_val);
        if (escaped) {
            buffer_append(buf, escaped);
            free(escaped);
        }
    } else {
        buffer_append(buf, str_val);
    }
    
    free(str_val);
}

static int evaluate_condition(TE_Engine *engine, const char *condition,
                               const TE_Value *context, size_t loop_index,
                               const TE_Value *loop_item)
{
    (void)engine;
    
    if (!condition || !*condition) {
        return 0;
    }
    
    TE_Value *value = NULL;
    
    if (loop_item) {
        value = lookup_special_var(loop_item, condition, loop_index);
        if (value && (strcmp(condition, "@index") == 0)) {
            int result = te_value_is_truthy(value);
            te_value_free(value);
            return result;
        }
        if (!value || strcmp(condition, "this") == 0) {
            value = resolve_variable(loop_item, condition);
        }
    }
    
    if (!value && context) {
        value = resolve_variable(context, condition);
    }
    
    return te_value_is_truthy(value);
}

static const char *get_partial(TE_Engine *engine, const char *name)
{
    if (!engine || !name) return NULL;
    
    for (size_t i = 0; i < engine->partials.count; i++) {
        if (strcmp(engine->partials.names[i], name) == 0) {
            return engine->partials.contents[i];
        }
    }
    return NULL;
}

static void render_tokens(TE_Engine *engine, TE_Token *tokens, const TE_Value *context,
                          size_t loop_index, const TE_Value *loop_item, RenderBuffer *buf)
{
    TE_Token *token = tokens;
    
    while (token) {
        switch (token->type) {
        case TOKEN_TEXT:
            if (token->content) {
                buffer_append(buf, token->content);
            }
            break;
            
        case TOKEN_VARIABLE:
            if (token->content) {
                render_variable(engine, token->content, context, loop_index, loop_item, 1, buf);
            }
            break;
            
        case TOKEN_VARIABLE_RAW:
            if (token->content) {
                render_variable(engine, token->content, context, loop_index, loop_item, 0, buf);
            }
            break;
            
        case TOKEN_IF: {
            int cond_result = 0;
            if (token->content) {
                cond_result = evaluate_condition(engine, token->content, context, loop_index, loop_item);
            }
            if (cond_result) {
                if (token->children) {
                    render_tokens(engine, token->children, context, loop_index, loop_item, buf);
                }
            } else {
                if (token->else_children) {
                    render_tokens(engine, token->else_children, context, loop_index, loop_item, buf);
                }
            }
            break;
        }
            
        case TOKEN_UNLESS: {
            int cond_result = 0;
            if (token->content) {
                cond_result = evaluate_condition(engine, token->content, context, loop_index, loop_item);
            }
            if (!cond_result) {
                if (token->children) {
                    render_tokens(engine, token->children, context, loop_index, loop_item, buf);
                }
            } else {
                if (token->else_children) {
                    render_tokens(engine, token->else_children, context, loop_index, loop_item, buf);
                }
            }
            break;
        }
            
        case TOKEN_EACH: {
            TE_Value *array_val = NULL;
            if (token->content) {
                if (loop_item) {
                    array_val = resolve_variable(loop_item, token->content);
                }
                if (!array_val && context) {
                    array_val = resolve_variable(context, token->content);
                }
            }
            
            if (array_val && array_val->type == TE_TYPE_ARRAY) {
                size_t len = te_array_length(array_val);
                for (size_t i = 0; i < len; i++) {
                    TE_Value *item = te_array_get(array_val, i);
                    if (token->children) {
                        render_tokens(engine, token->children, context, i, item, buf);
                    }
                }
            }
            break;
        }
            
        case TOKEN_PARTIAL: {
            if (token->content) {
                const char *partial_content = get_partial(engine, token->content);
                if (partial_content) {
                    TE_Token *partial_tokens = te_tokenize(engine, partial_content);
                    if (partial_tokens) {
                        if (loop_item) {
                            render_tokens(engine, partial_tokens->children, context, loop_index, loop_item, buf);
                        } else {
                            render_tokens(engine, partial_tokens->children, context, 0, NULL, buf);
                        }
                        te_token_free(partial_tokens);
                    }
                }
            }
            break;
        }
            
        default:
            break;
        }
        
        token = token->next;
    }
}

char *te_render(TE_Engine *engine, const char *template_str, TE_Value *context)
{
    if (!engine || !template_str) {
        return NULL;
    }
    
    TE_Token *tokens = te_tokenize(engine, template_str);
    if (!tokens) {
        return NULL;
    }
    
    RenderBuffer buf;
    buffer_init(&buf);
    if (!buf.data) {
        te_token_free(tokens);
        return NULL;
    }
    
    render_tokens(engine, tokens->children, context, 0, NULL, &buf);
    
    te_token_free(tokens);
    
    if (!buf.data) {
        return NULL;
    }
    
    return buf.data;
}
