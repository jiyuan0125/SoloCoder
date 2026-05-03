#ifndef JSON_FORMATTER_H
#define JSON_FORMATTER_H

#include "json_value.h"

typedef enum {
    JSON_FORMAT_COMPACT,
    JSON_FORMAT_PRETTY
} JsonFormatStyle;

typedef struct {
    JsonFormatStyle style;
    int indent_size;
    int current_indent;
} JsonFormatter;

void json_formatter_init(JsonFormatter *formatter, JsonFormatStyle style);
void json_formatter_init_ex(JsonFormatter *formatter, JsonFormatStyle style, int indent_size);

char *json_to_string(JsonValue *value);
char *json_to_string_pretty(JsonValue *value);
char *json_to_string_ex(JsonValue *value, JsonFormatter *formatter);

#endif
