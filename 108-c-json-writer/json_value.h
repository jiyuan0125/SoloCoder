#ifndef JSON_VALUE_H
#define JSON_VALUE_H

#include <stddef.h>

typedef enum {
    JSON_NULL,
    JSON_BOOL,
    JSON_INT,
    JSON_FLOAT,
    JSON_STRING,
    JSON_OBJECT,
    JSON_ARRAY
} JsonValueType;

typedef struct JsonValue JsonValue;
typedef struct JsonKeyValue JsonKeyValue;

struct JsonValue {
    JsonValueType type;
    union {
        int bool_val;
        long long int_val;
        double float_val;
        char *str_val;
        struct {
            JsonKeyValue *items;
            size_t count;
            size_t capacity;
        } object;
        struct {
            JsonValue **items;
            size_t count;
            size_t capacity;
        } array;
    } value;
};

struct JsonKeyValue {
    char *key;
    JsonValue *value;
};

JsonValue *json_create_null(void);
JsonValue *json_create_bool(int value);
JsonValue *json_create_int(long long value);
JsonValue *json_create_float(double value);
JsonValue *json_create_string(const char *value);
JsonValue *json_create_object(void);
JsonValue *json_create_array(void);

int json_object_add(JsonValue *obj, const char *key, JsonValue *value);
int json_array_add(JsonValue *arr, JsonValue *value);

void json_free(JsonValue *value);

#endif
