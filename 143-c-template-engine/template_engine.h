#ifndef TEMPLATE_ENGINE_H
#define TEMPLATE_ENGINE_H

#include <stddef.h>

typedef enum {
    TE_TYPE_NULL,
    TE_TYPE_BOOL,
    TE_TYPE_INT,
    TE_TYPE_DOUBLE,
    TE_TYPE_STRING,
    TE_TYPE_OBJECT,
    TE_TYPE_ARRAY
} TE_ValueType;

typedef struct TE_Value TE_Value;
typedef struct TE_Object TE_Object;
typedef struct TE_Array TE_Array;

struct TE_Value {
    TE_ValueType type;
    union {
        int bool_val;
        long long int_val;
        double double_val;
        char *string_val;
        TE_Object *object_val;
        TE_Array *array_val;
    } data;
};

struct TE_Object {
    char **keys;
    TE_Value **values;
    size_t count;
    size_t capacity;
};

struct TE_Array {
    TE_Value **items;
    size_t count;
    size_t capacity;
};

typedef struct {
    char *open;
    char *close;
    char *open_triple;
    char *close_triple;
} TE_Delimiters;

typedef struct TE_Partials {
    char **names;
    char **contents;
    size_t count;
    size_t capacity;
} TE_Partials;

typedef struct {
    char *message;
    int has_error;
} TE_Error;

typedef enum {
    TOKEN_TEXT,
    TOKEN_VARIABLE,
    TOKEN_VARIABLE_RAW,
    TOKEN_IF,
    TOKEN_UNLESS,
    TOKEN_ELSE,
    TOKEN_END_IF,
    TOKEN_EACH,
    TOKEN_END_EACH,
    TOKEN_PARTIAL
} TE_TokenType;

typedef struct TE_Token TE_Token;

struct TE_Token {
    TE_TokenType type;
    char *content;
    size_t start;
    size_t end;
    TE_Token *next;
    TE_Token *children;
    TE_Token *else_children;
};

typedef struct {
    TE_Delimiters delimiters;
    TE_Partials partials;
    TE_Error error;
} TE_Engine;

TE_Engine *te_engine_create(void);
void te_engine_destroy(TE_Engine *engine);
void te_engine_set_delimiters(TE_Engine *engine, const char *open, const char *close);
void te_engine_add_partial(TE_Engine *engine, const char *name, const char *content);

TE_Value *te_value_null(void);
TE_Value *te_value_bool(int val);
TE_Value *te_value_int(long long val);
TE_Value *te_value_double(double val);
TE_Value *te_value_string(const char *val);
TE_Value *te_value_object(void);
TE_Value *te_value_array(void);
void te_value_free(TE_Value *value);
int te_value_is_truthy(const TE_Value *value);

void te_object_set(TE_Value *obj, const char *key, TE_Value *value);
TE_Value *te_object_get(const TE_Value *obj, const char *key);

void te_array_push(TE_Value *arr, TE_Value *value);
TE_Value *te_array_get(const TE_Value *arr, size_t index);
size_t te_array_length(const TE_Value *arr);

TE_Token *te_tokenize(TE_Engine *engine, const char *template_str);
void te_token_free(TE_Token *token);

char *te_render(TE_Engine *engine, const char *template_str, TE_Value *context);

char *te_html_escape(const char *input);

#endif
