#ifndef JSON_H
#define JSON_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef enum {
    JSON_OK = 0,
    JSON_ERROR_MEMORY,
    JSON_ERROR_SYNTAX,
    JSON_ERROR_EOF,
    JSON_ERROR_INVALID_ESCAPE,
    JSON_ERROR_INVALID_UNICODE,
    JSON_ERROR_INVALID_UTF8,
    JSON_ERROR_INVALID_NUMBER,
    JSON_ERROR_TRAILING_CONTENT,
    JSON_ERROR_INVALID_TYPE,
    JSON_ERROR_KEY_NOT_FOUND,
    JSON_ERROR_INDEX_OUT_OF_BOUNDS
} json_error_t;

typedef struct {
    int line;
    int column;
} json_pos_t;

typedef enum {
    JSON_NULL,
    JSON_BOOL,
    JSON_INT,
    JSON_FLOAT,
    JSON_STRING,
    JSON_ARRAY,
    JSON_OBJECT
} json_type_t;

struct json_value;

typedef struct {
    char *key;
    struct json_value *value;
} json_object_member_t;

typedef struct json_value json_value_t;

typedef struct {
    json_object_member_t *members;
    size_t count;
    size_t capacity;
} json_object_t;

typedef struct {
    json_value_t **elements;
    size_t count;
    size_t capacity;
} json_array_t;

struct json_value {
    json_type_t type;
    union {
        bool bool_val;
        long long int_val;
        double float_val;
        char *str_val;
        json_array_t *array_val;
        json_object_t *object_val;
    } u;
};

typedef struct json_parser json_parser_t;

typedef void (*json_sax_callback_t)(json_parser_t *parser, void *userdata, json_pos_t pos);
typedef void (*json_sax_key_callback_t)(json_parser_t *parser, void *userdata, const char *key, json_pos_t pos);
typedef void (*json_sax_string_callback_t)(json_parser_t *parser, void *userdata, const char *value, json_pos_t pos);
typedef void (*json_sax_int_callback_t)(json_parser_t *parser, void *userdata, long long value, json_pos_t pos);
typedef void (*json_sax_float_callback_t)(json_parser_t *parser, void *userdata, double value, json_pos_t pos);
typedef void (*json_sax_bool_callback_t)(json_parser_t *parser, void *userdata, bool value, json_pos_t pos);

typedef struct {
    json_sax_callback_t object_start;
    json_sax_callback_t object_end;
    json_sax_callback_t array_start;
    json_sax_callback_t array_end;
    json_sax_key_callback_t key;
    json_sax_string_callback_t string;
    json_sax_int_callback_t number_int;
    json_sax_float_callback_t number_float;
    json_sax_bool_callback_t boolean;
    json_sax_callback_t null;
} json_sax_callbacks_t;

struct json_parser {
    FILE *input;
    unsigned char buffer[4096];
    size_t buffer_pos;
    size_t buffer_len;
    bool buffer_eof;
    int line;
    int column;
    unsigned char lookahead[4];
    int lookahead_count;
    json_error_t error;
    char error_message[256];
    json_pos_t error_pos;
    const json_sax_callbacks_t *callbacks;
    void *userdata;
};

void json_parser_init(json_parser_t *parser, FILE *input, const json_sax_callbacks_t *callbacks, void *userdata);
json_error_t json_parse_sax(json_parser_t *parser);
const char *json_error_message(json_parser_t *parser);
json_pos_t json_error_position(json_parser_t *parser);

json_value_t *json_parse_dom(FILE *input, json_error_t *error);
void json_free(json_value_t *value);

json_type_t json_type(const json_value_t *value);

bool json_get_bool(const json_value_t *value, bool *out);
bool json_get_int(const json_value_t *value, long long *out);
bool json_get_float(const json_value_t *value, double *out);
bool json_get_string(const json_value_t *value, const char **out);

size_t json_array_length(const json_value_t *array);
json_value_t *json_get_at(const json_value_t *array, size_t index);

size_t json_object_size(const json_value_t *obj);
json_value_t *json_get(const json_value_t *obj, const char *key);

#ifdef __cplusplus
}
#endif

#endif
