#include "json_formatter.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <math.h>

typedef struct {
    char *data;
    size_t size;
    size_t capacity;
} StringBuffer;

static void string_buffer_init(StringBuffer *buf) {
    buf->data = NULL;
    buf->size = 0;
    buf->capacity = 0;
}

static void string_buffer_free(StringBuffer *buf) {
    if (buf->data) {
        free(buf->data);
        buf->data = NULL;
    }
    buf->size = 0;
    buf->capacity = 0;
}

static int string_buffer_ensure_capacity(StringBuffer *buf, size_t needed) {
    if (buf->size + needed + 1 > buf->capacity) {
        size_t new_cap = buf->capacity;
        if (new_cap == 0) new_cap = 32;
        while (new_cap < buf->size + needed + 1) {
            new_cap *= 2;
        }
        char *new_data = (char *)realloc(buf->data, new_cap);
        if (new_data == NULL) return -1;
        buf->data = new_data;
        buf->capacity = new_cap;
    }
    return 0;
}

static int string_buffer_append(StringBuffer *buf, const char *str) {
    size_t len = strlen(str);
    if (string_buffer_ensure_capacity(buf, len) != 0) return -1;
    memcpy(buf->data + buf->size, str, len);
    buf->size += len;
    buf->data[buf->size] = '\0';
    return 0;
}

static int string_buffer_append_char(StringBuffer *buf, char c) {
    if (string_buffer_ensure_capacity(buf, 1) != 0) return -1;
    buf->data[buf->size++] = c;
    buf->data[buf->size] = '\0';
    return 0;
}

static int string_buffer_append_indent(StringBuffer *buf, int count) {
    for (int i = 0; i < count; i++) {
        if (string_buffer_append_char(buf, ' ') != 0) return -1;
    }
    return 0;
}

static size_t escaped_length(const char *str) {
    if (str == NULL) return 4;
    size_t len = 0;
    for (const char *p = str; *p != '\0'; p++) {
        switch (*p) {
            case '\"':
            case '\\':
            case '\b':
            case '\f':
            case '\n':
            case '\r':
            case '\t':
                len += 2;
                break;
            default:
                len += 1;
                break;
        }
    }
    return len;
}

static int append_escaped_string(StringBuffer *buf, const char *str) {
    if (str == NULL) {
        return string_buffer_append(buf, "null");
    }
    
    if (string_buffer_append_char(buf, '\"') != 0) return -1;
    
    for (const char *p = str; *p != '\0'; p++) {
        switch (*p) {
            case '\"':
                if (string_buffer_append(buf, "\\\"") != 0) return -1;
                break;
            case '\\':
                if (string_buffer_append(buf, "\\\\") != 0) return -1;
                break;
            case '\b':
                if (string_buffer_append(buf, "\\b") != 0) return -1;
                break;
            case '\f':
                if (string_buffer_append(buf, "\\f") != 0) return -1;
                break;
            case '\n':
                if (string_buffer_append(buf, "\\n") != 0) return -1;
                break;
            case '\r':
                if (string_buffer_append(buf, "\\r") != 0) return -1;
                break;
            case '\t':
                if (string_buffer_append(buf, "\\t") != 0) return -1;
                break;
            default:
                if (string_buffer_append_char(buf, *p) != 0) return -1;
                break;
        }
    }
    
    if (string_buffer_append_char(buf, '\"') != 0) return -1;
    return 0;
}

static int format_float(double value, char *buffer, size_t size) {
    if (isnan(value) || isinf(value)) {
        if (size < 5) return -1;
        strcpy(buffer, "null");
        return 0;
    }
    
    int written = snprintf(buffer, size, "%.15g", value);
    if (written < 0 || (size_t)written >= size) {
        return -1;
    }
    
    int has_point = 0;
    int has_e = 0;
    for (int i = 0; buffer[i] != '\0'; i++) {
        if (buffer[i] == '.') has_point = 1;
        if (buffer[i] == 'e' || buffer[i] == 'E') has_e = 1;
    }
    
    if (!has_point && !has_e) {
        if ((size_t)written + 2 >= size) return -1;
        strcat(buffer, ".0");
    }
    
    return 0;
}

static int json_value_to_buffer(JsonValue *value, StringBuffer *buf, JsonFormatter *formatter);

static int json_object_to_buffer(JsonValue *obj, StringBuffer *buf, JsonFormatter *formatter) {
    if (obj == NULL || obj->type != JSON_OBJECT) return -1;
    
    if (string_buffer_append_char(buf, '{') != 0) return -1;
    
    size_t count = obj->value.object.count;
    
    if (formatter->style == JSON_FORMAT_PRETTY && count > 0) {
        if (string_buffer_append_char(buf, '\n') != 0) return -1;
        formatter->current_indent += formatter->indent_size;
    }
    
    for (size_t i = 0; i < count; i++) {
        JsonKeyValue *kv = &obj->value.object.items[i];
        
        if (formatter->style == JSON_FORMAT_PRETTY) {
            if (string_buffer_append_indent(buf, formatter->current_indent) != 0) return -1;
        }
        
        if (append_escaped_string(buf, kv->key) != 0) return -1;
        
        if (formatter->style == JSON_FORMAT_PRETTY) {
            if (string_buffer_append(buf, ": ") != 0) return -1;
        } else {
            if (string_buffer_append_char(buf, ':') != 0) return -1;
        }
        
        if (json_value_to_buffer(kv->value, buf, formatter) != 0) return -1;
        
        if (i < count - 1) {
            if (string_buffer_append_char(buf, ',') != 0) return -1;
        }
        
        if (formatter->style == JSON_FORMAT_PRETTY) {
            if (string_buffer_append_char(buf, '\n') != 0) return -1;
        }
    }
    
    if (formatter->style == JSON_FORMAT_PRETTY && count > 0) {
        formatter->current_indent -= formatter->indent_size;
        if (string_buffer_append_indent(buf, formatter->current_indent) != 0) return -1;
    }
    
    if (string_buffer_append_char(buf, '}') != 0) return -1;
    
    return 0;
}

static int json_array_to_buffer(JsonValue *arr, StringBuffer *buf, JsonFormatter *formatter) {
    if (arr == NULL || arr->type != JSON_ARRAY) return -1;
    
    if (string_buffer_append_char(buf, '[') != 0) return -1;
    
    size_t count = arr->value.array.count;
    
    if (formatter->style == JSON_FORMAT_PRETTY && count > 0) {
        if (string_buffer_append_char(buf, '\n') != 0) return -1;
        formatter->current_indent += formatter->indent_size;
    }
    
    for (size_t i = 0; i < count; i++) {
        JsonValue *val = arr->value.array.items[i];
        
        if (formatter->style == JSON_FORMAT_PRETTY) {
            if (string_buffer_append_indent(buf, formatter->current_indent) != 0) return -1;
        }
        
        if (json_value_to_buffer(val, buf, formatter) != 0) return -1;
        
        if (i < count - 1) {
            if (string_buffer_append_char(buf, ',') != 0) return -1;
        }
        
        if (formatter->style == JSON_FORMAT_PRETTY) {
            if (string_buffer_append_char(buf, '\n') != 0) return -1;
        }
    }
    
    if (formatter->style == JSON_FORMAT_PRETTY && count > 0) {
        formatter->current_indent -= formatter->indent_size;
        if (string_buffer_append_indent(buf, formatter->current_indent) != 0) return -1;
    }
    
    if (string_buffer_append_char(buf, ']') != 0) return -1;
    
    return 0;
}

static int json_value_to_buffer(JsonValue *value, StringBuffer *buf, JsonFormatter *formatter) {
    if (value == NULL) {
        return string_buffer_append(buf, "null");
    }
    
    switch (value->type) {
        case JSON_NULL:
            return string_buffer_append(buf, "null");
            
        case JSON_BOOL:
            if (value->value.bool_val) {
                return string_buffer_append(buf, "true");
            } else {
                return string_buffer_append(buf, "false");
            }
            
        case JSON_INT: {
            char int_buf[64];
            snprintf(int_buf, sizeof(int_buf), "%lld", value->value.int_val);
            return string_buffer_append(buf, int_buf);
        }
            
        case JSON_FLOAT: {
            char float_buf[64];
            if (format_float(value->value.float_val, float_buf, sizeof(float_buf)) != 0) {
                return string_buffer_append(buf, "null");
            }
            return string_buffer_append(buf, float_buf);
        }
            
        case JSON_STRING:
            return append_escaped_string(buf, value->value.str_val);
            
        case JSON_OBJECT:
            return json_object_to_buffer(value, buf, formatter);
            
        case JSON_ARRAY:
            return json_array_to_buffer(value, buf, formatter);
            
        default:
            return string_buffer_append(buf, "null");
    }
}

void json_formatter_init(JsonFormatter *formatter, JsonFormatStyle style) {
    json_formatter_init_ex(formatter, style, 2);
}

void json_formatter_init_ex(JsonFormatter *formatter, JsonFormatStyle style, int indent_size) {
    if (formatter == NULL) return;
    formatter->style = style;
    formatter->indent_size = (indent_size > 0) ? indent_size : 2;
    formatter->current_indent = 0;
}

char *json_to_string(JsonValue *value) {
    JsonFormatter formatter;
    json_formatter_init(&formatter, JSON_FORMAT_COMPACT);
    return json_to_string_ex(value, &formatter);
}

char *json_to_string_pretty(JsonValue *value) {
    JsonFormatter formatter;
    json_formatter_init(&formatter, JSON_FORMAT_PRETTY);
    return json_to_string_ex(value, &formatter);
}

char *json_to_string_ex(JsonValue *value, JsonFormatter *formatter) {
    if (value == NULL || formatter == NULL) return NULL;
    
    StringBuffer buf;
    string_buffer_init(&buf);
    
    int saved_indent = formatter->current_indent;
    formatter->current_indent = 0;
    
    if (json_value_to_buffer(value, &buf, formatter) != 0) {
        string_buffer_free(&buf);
        formatter->current_indent = saved_indent;
        return NULL;
    }
    
    formatter->current_indent = saved_indent;
    return buf.data;
}
