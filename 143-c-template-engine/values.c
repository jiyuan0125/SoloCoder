#include "template_engine.h"
#include <stdlib.h>
#include <string.h>

TE_Value *te_value_null(void)
{
    TE_Value *v = (TE_Value *)malloc(sizeof(TE_Value));
    if (v) {
        v->type = TE_TYPE_NULL;
    }
    return v;
}

TE_Value *te_value_bool(int val)
{
    TE_Value *v = (TE_Value *)malloc(sizeof(TE_Value));
    if (v) {
        v->type = TE_TYPE_BOOL;
        v->data.bool_val = val ? 1 : 0;
    }
    return v;
}

TE_Value *te_value_int(long long val)
{
    TE_Value *v = (TE_Value *)malloc(sizeof(TE_Value));
    if (v) {
        v->type = TE_TYPE_INT;
        v->data.int_val = val;
    }
    return v;
}

TE_Value *te_value_double(double val)
{
    TE_Value *v = (TE_Value *)malloc(sizeof(TE_Value));
    if (v) {
        v->type = TE_TYPE_DOUBLE;
        v->data.double_val = val;
    }
    return v;
}

TE_Value *te_value_string(const char *val)
{
    TE_Value *v = (TE_Value *)malloc(sizeof(TE_Value));
    if (v) {
        v->type = TE_TYPE_STRING;
        if (val) {
            v->data.string_val = (char *)malloc(strlen(val) + 1);
            if (v->data.string_val) {
                strcpy(v->data.string_val, val);
            } else {
                free(v);
                return NULL;
            }
        } else {
            v->data.string_val = (char *)malloc(1);
            if (v->data.string_val) {
                v->data.string_val[0] = '\0';
            } else {
                free(v);
                return NULL;
            }
        }
    }
    return v;
}

TE_Value *te_value_object(void)
{
    TE_Value *v = (TE_Value *)malloc(sizeof(TE_Value));
    if (!v) return NULL;
    
    v->type = TE_TYPE_OBJECT;
    v->data.object_val = (TE_Object *)malloc(sizeof(TE_Object));
    if (!v->data.object_val) {
        free(v);
        return NULL;
    }
    
    v->data.object_val->keys = NULL;
    v->data.object_val->values = NULL;
    v->data.object_val->count = 0;
    v->data.object_val->capacity = 0;
    
    return v;
}

TE_Value *te_value_array(void)
{
    TE_Value *v = (TE_Value *)malloc(sizeof(TE_Value));
    if (!v) return NULL;
    
    v->type = TE_TYPE_ARRAY;
    v->data.array_val = (TE_Array *)malloc(sizeof(TE_Array));
    if (!v->data.array_val) {
        free(v);
        return NULL;
    }
    
    v->data.array_val->items = NULL;
    v->data.array_val->count = 0;
    v->data.array_val->capacity = 0;
    
    return v;
}

void te_value_free(TE_Value *value)
{
    if (!value) return;
    
    switch (value->type) {
    case TE_TYPE_STRING:
        free(value->data.string_val);
        break;
    case TE_TYPE_OBJECT:
        if (value->data.object_val) {
            for (size_t i = 0; i < value->data.object_val->count; i++) {
                free(value->data.object_val->keys[i]);
                te_value_free(value->data.object_val->values[i]);
            }
            free(value->data.object_val->keys);
            free(value->data.object_val->values);
            free(value->data.object_val);
        }
        break;
    case TE_TYPE_ARRAY:
        if (value->data.array_val) {
            for (size_t i = 0; i < value->data.array_val->count; i++) {
                te_value_free(value->data.array_val->items[i]);
            }
            free(value->data.array_val->items);
            free(value->data.array_val);
        }
        break;
    default:
        break;
    }
    free(value);
}

int te_value_is_truthy(const TE_Value *value)
{
    if (!value) return 0;
    
    switch (value->type) {
    case TE_TYPE_NULL:
        return 0;
    case TE_TYPE_BOOL:
        return value->data.bool_val;
    case TE_TYPE_INT:
        return value->data.int_val != 0;
    case TE_TYPE_DOUBLE:
        return value->data.double_val != 0.0;
    case TE_TYPE_STRING:
        return value->data.string_val && value->data.string_val[0] != '\0';
    case TE_TYPE_OBJECT:
        return value->data.object_val != NULL;
    case TE_TYPE_ARRAY:
        return value->data.array_val != NULL;
    default:
        return 0;
    }
}

void te_object_set(TE_Value *obj, const char *key, TE_Value *value)
{
    if (!obj || obj->type != TE_TYPE_OBJECT || !key || !value) return;
    
    TE_Object *o = obj->data.object_val;
    if (!o) return;
    
    for (size_t i = 0; i < o->count; i++) {
        if (strcmp(o->keys[i], key) == 0) {
            te_value_free(o->values[i]);
            o->values[i] = value;
            return;
        }
    }
    
    if (o->count >= o->capacity) {
        size_t new_cap = o->capacity == 0 ? 8 : o->capacity * 2;
        char **new_keys = (char **)realloc(o->keys, new_cap * sizeof(char *));
        TE_Value **new_vals = (TE_Value **)realloc(o->values, new_cap * sizeof(TE_Value *));
        
        if (!new_keys || !new_vals) {
            free(new_keys);
            free(new_vals);
            te_value_free(value);
            return;
        }
        
        o->keys = new_keys;
        o->values = new_vals;
        o->capacity = new_cap;
    }
    
    o->keys[o->count] = (char *)malloc(strlen(key) + 1);
    if (!o->keys[o->count]) {
        te_value_free(value);
        return;
    }
    strcpy(o->keys[o->count], key);
    o->values[o->count] = value;
    o->count++;
}

TE_Value *te_object_get(const TE_Value *obj, const char *key)
{
    if (!obj || obj->type != TE_TYPE_OBJECT || !key) return NULL;
    
    TE_Object *o = obj->data.object_val;
    if (!o) return NULL;
    
    for (size_t i = 0; i < o->count; i++) {
        if (strcmp(o->keys[i], key) == 0) {
            return o->values[i];
        }
    }
    return NULL;
}

void te_array_push(TE_Value *arr, TE_Value *value)
{
    if (!arr || arr->type != TE_TYPE_ARRAY || !value) return;
    
    TE_Array *a = arr->data.array_val;
    if (!a) return;
    
    if (a->count >= a->capacity) {
        size_t new_cap = a->capacity == 0 ? 8 : a->capacity * 2;
        TE_Value **new_items = (TE_Value **)realloc(a->items, new_cap * sizeof(TE_Value *));
        
        if (!new_items) {
            te_value_free(value);
            return;
        }
        
        a->items = new_items;
        a->capacity = new_cap;
    }
    
    a->items[a->count] = value;
    a->count++;
}

TE_Value *te_array_get(const TE_Value *arr, size_t index)
{
    if (!arr || arr->type != TE_TYPE_ARRAY) return NULL;
    
    TE_Array *a = arr->data.array_val;
    if (!a || index >= a->count) return NULL;
    
    return a->items[index];
}

size_t te_array_length(const TE_Value *arr)
{
    if (!arr || arr->type != TE_TYPE_ARRAY) return 0;
    
    TE_Array *a = arr->data.array_val;
    if (!a) return 0;
    
    return a->count;
}
