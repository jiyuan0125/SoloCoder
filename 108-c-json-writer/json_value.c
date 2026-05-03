#include "json_value.h"
#include "json_allocator.h"
#include <stdlib.h>
#include <string.h>

JsonValue *json_create_null(void) {
    JsonValue *val = (JsonValue *)json_malloc(sizeof(JsonValue));
    if (val == NULL) return NULL;
    val->type = JSON_NULL;
    return val;
}

JsonValue *json_create_bool(int value) {
    JsonValue *val = (JsonValue *)json_malloc(sizeof(JsonValue));
    if (val == NULL) return NULL;
    val->type = JSON_BOOL;
    val->value.bool_val = value ? 1 : 0;
    return val;
}

JsonValue *json_create_int(long long value) {
    JsonValue *val = (JsonValue *)json_malloc(sizeof(JsonValue));
    if (val == NULL) return NULL;
    val->type = JSON_INT;
    val->value.int_val = value;
    return val;
}

JsonValue *json_create_float(double value) {
    JsonValue *val = (JsonValue *)json_malloc(sizeof(JsonValue));
    if (val == NULL) return NULL;
    val->type = JSON_FLOAT;
    val->value.float_val = value;
    return val;
}

JsonValue *json_create_string(const char *value) {
    JsonValue *val = (JsonValue *)json_malloc(sizeof(JsonValue));
    if (val == NULL) return NULL;
    val->type = JSON_STRING;
    val->value.str_val = json_strdup(value);
    if (val->value.str_val == NULL) {
        json_free_ptr(val);
        return NULL;
    }
    return val;
}

JsonValue *json_create_object(void) {
    JsonValue *val = (JsonValue *)json_malloc(sizeof(JsonValue));
    if (val == NULL) return NULL;
    val->type = JSON_OBJECT;
    val->value.object.items = NULL;
    val->value.object.count = 0;
    val->value.object.capacity = 0;
    return val;
}

JsonValue *json_create_array(void) {
    JsonValue *val = (JsonValue *)json_malloc(sizeof(JsonValue));
    if (val == NULL) return NULL;
    val->type = JSON_ARRAY;
    val->value.array.items = NULL;
    val->value.array.count = 0;
    val->value.array.capacity = 0;
    return val;
}

int json_object_add(JsonValue *obj, const char *key, JsonValue *value) {
    if (obj == NULL || obj->type != JSON_OBJECT || key == NULL || value == NULL) {
        return -1;
    }
    
    if (obj->value.object.count >= obj->value.object.capacity) {
        size_t new_cap = (obj->value.object.capacity == 0) ? 4 : obj->value.object.capacity * 2;
        JsonKeyValue *new_items = (JsonKeyValue *)json_realloc(
            obj->value.object.items,
            new_cap * sizeof(JsonKeyValue)
        );
        if (new_items == NULL) return -1;
        obj->value.object.items = new_items;
        obj->value.object.capacity = new_cap;
    }
    
    size_t idx = obj->value.object.count++;
    obj->value.object.items[idx].key = json_strdup(key);
    if (obj->value.object.items[idx].key == NULL) {
        obj->value.object.count--;
        return -1;
    }
    obj->value.object.items[idx].value = value;
    return 0;
}

int json_array_add(JsonValue *arr, JsonValue *value) {
    if (arr == NULL || arr->type != JSON_ARRAY || value == NULL) {
        return -1;
    }
    
    if (arr->value.array.count >= arr->value.array.capacity) {
        size_t new_cap = (arr->value.array.capacity == 0) ? 4 : arr->value.array.capacity * 2;
        JsonValue **new_items = (JsonValue **)json_realloc(
            arr->value.array.items,
            new_cap * sizeof(JsonValue *)
        );
        if (new_items == NULL) return -1;
        arr->value.array.items = new_items;
        arr->value.array.capacity = new_cap;
    }
    
    arr->value.array.items[arr->value.array.count++] = value;
    return 0;
}

void json_free(JsonValue *value) {
    if (value == NULL) return;
    
    switch (value->type) {
        case JSON_STRING:
            json_free_ptr(value->value.str_val);
            break;
            
        case JSON_OBJECT:
            for (size_t i = 0; i < value->value.object.count; i++) {
                json_free_ptr(value->value.object.items[i].key);
                json_free(value->value.object.items[i].value);
            }
            json_free_ptr(value->value.object.items);
            break;
            
        case JSON_ARRAY:
            for (size_t i = 0; i < value->value.array.count; i++) {
                json_free(value->value.array.items[i]);
            }
            json_free_ptr(value->value.array.items);
            break;
            
        default:
            break;
    }
    
    json_free_ptr(value);
}
