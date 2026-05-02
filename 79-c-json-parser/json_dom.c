#include "json.h"
#include <string.h>

typedef struct {
    json_value_t *root;
    struct stack_frame *stack;
    int stack_size;
    int stack_capacity;
    char *pending_key;
    bool has_error;
} dom_builder_t;

typedef struct stack_frame {
    json_value_t *container;
    enum {
        FRAME_OBJECT,
        FRAME_ARRAY
    } type;
} stack_frame_t;

static json_value_t *json_new_value(json_type_t type)
{
    json_value_t *v = (json_value_t *)calloc(1, sizeof(json_value_t));
    if (!v) return NULL;
    v->type = type;
    return v;
}

static json_value_t *json_new_int(long long val)
{
    json_value_t *v = json_new_value(JSON_INT);
    if (v) v->u.int_val = val;
    return v;
}

static json_value_t *json_new_float(double val)
{
    json_value_t *v = json_new_value(JSON_FLOAT);
    if (v) v->u.float_val = val;
    return v;
}

static json_value_t *json_new_bool(bool val)
{
    json_value_t *v = json_new_value(JSON_BOOL);
    if (v) v->u.bool_val = val;
    return v;
}

static json_value_t *json_new_null(void)
{
    return json_new_value(JSON_NULL);
}

static json_value_t *json_new_string(const char *s)
{
    json_value_t *v = json_new_value(JSON_STRING);
    if (!v) return NULL;
    v->u.str_val = strdup(s);
    if (!v->u.str_val) {
        free(v);
        return NULL;
    }
    return v;
}

static json_array_t *json_array_new(void)
{
    json_array_t *arr = (json_array_t *)calloc(1, sizeof(json_array_t));
    if (!arr) return NULL;
    arr->capacity = 8;
    arr->elements = (json_value_t **)malloc(arr->capacity * sizeof(json_value_t *));
    if (!arr->elements) {
        free(arr);
        return NULL;
    }
    arr->count = 0;
    return arr;
}

static void json_array_add(json_array_t *arr, json_value_t *val)
{
    if (arr->count >= arr->capacity) {
        size_t new_cap = arr->capacity * 2;
        json_value_t **new_elems = (json_value_t **)realloc(arr->elements, new_cap * sizeof(json_value_t *));
        if (!new_elems) return;
        arr->elements = new_elems;
        arr->capacity = new_cap;
    }
    arr->elements[arr->count++] = val;
}

static json_object_t *json_object_new(void)
{
    json_object_t *obj = (json_object_t *)calloc(1, sizeof(json_object_t));
    if (!obj) return NULL;
    obj->capacity = 8;
    obj->members = (json_object_member_t *)malloc(obj->capacity * sizeof(json_object_member_t));
    if (!obj->members) {
        free(obj);
        return NULL;
    }
    obj->count = 0;
    return obj;
}

static void json_object_add(json_object_t *obj, const char *key, json_value_t *val)
{
    if (obj->count >= obj->capacity) {
        size_t new_cap = obj->capacity * 2;
        json_object_member_t *new_mems = (json_object_member_t *)realloc(obj->members, new_cap * sizeof(json_object_member_t));
        if (!new_mems) return;
        obj->members = new_mems;
        obj->capacity = new_cap;
    }
    obj->members[obj->count].key = strdup(key);
    obj->members[obj->count].value = val;
    obj->count++;
}

static void dom_builder_init(dom_builder_t *builder)
{
    memset(builder, 0, sizeof(*builder));
    builder->stack_capacity = 16;
    builder->stack = (stack_frame_t *)malloc(builder->stack_capacity * sizeof(stack_frame_t));
    builder->stack_size = 0;
}

static void dom_builder_push(dom_builder_t *builder, json_value_t *container, int frame_type)
{
    if (builder->stack_size >= builder->stack_capacity) {
        int new_cap = builder->stack_capacity * 2;
        stack_frame_t *new_stack = (stack_frame_t *)realloc(builder->stack, new_cap * sizeof(stack_frame_t));
        if (!new_stack) {
            builder->has_error = true;
            return;
        }
        builder->stack = new_stack;
        builder->stack_capacity = new_cap;
    }
    builder->stack[builder->stack_size].container = container;
    builder->stack[builder->stack_size].type = frame_type;
    builder->stack_size++;
}

static stack_frame_t *dom_builder_top(dom_builder_t *builder)
{
    if (builder->stack_size == 0) return NULL;
    return &builder->stack[builder->stack_size - 1];
}

static void dom_builder_pop(dom_builder_t *builder)
{
    if (builder->stack_size > 0)
        builder->stack_size--;
}

static void dom_builder_add_value(dom_builder_t *builder, json_value_t *val)
{
    if (builder->stack_size == 0) {
        builder->root = val;
    } else {
        stack_frame_t *top = dom_builder_top(builder);
        if (top->type == FRAME_ARRAY) {
            json_array_add(top->container->u.array_val, val);
        } else {
            if (builder->pending_key) {
                json_object_add(top->container->u.object_val, builder->pending_key, val);
                free(builder->pending_key);
                builder->pending_key = NULL;
            }
        }
    }
}

static void sax_object_start(json_parser_t *parser, void *userdata, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    json_value_t *obj = json_new_value(JSON_OBJECT);
    if (!obj) { builder->has_error = true; return; }
    obj->u.object_val = json_object_new();
    if (!obj->u.object_val) { free(obj); builder->has_error = true; return; }
    
    dom_builder_add_value(builder, obj);
    dom_builder_push(builder, obj, FRAME_OBJECT);
}

static void sax_object_end(json_parser_t *parser, void *userdata, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    dom_builder_pop(builder);
}

static void sax_array_start(json_parser_t *parser, void *userdata, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    json_value_t *arr = json_new_value(JSON_ARRAY);
    if (!arr) { builder->has_error = true; return; }
    arr->u.array_val = json_array_new();
    if (!arr->u.array_val) { free(arr); builder->has_error = true; return; }
    
    dom_builder_add_value(builder, arr);
    dom_builder_push(builder, arr, FRAME_ARRAY);
}

static void sax_array_end(json_parser_t *parser, void *userdata, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    dom_builder_pop(builder);
}

static void sax_key(json_parser_t *parser, void *userdata, const char *key, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    if (builder->pending_key) free(builder->pending_key);
    builder->pending_key = strdup(key);
}

static void sax_string(json_parser_t *parser, void *userdata, const char *value, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    json_value_t *v = json_new_string(value);
    if (!v) { builder->has_error = true; return; }
    dom_builder_add_value(builder, v);
}

static void sax_number_int(json_parser_t *parser, void *userdata, long long value, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    json_value_t *v = json_new_int(value);
    if (!v) { builder->has_error = true; return; }
    dom_builder_add_value(builder, v);
}

static void sax_number_float(json_parser_t *parser, void *userdata, double value, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    json_value_t *v = json_new_float(value);
    if (!v) { builder->has_error = true; return; }
    dom_builder_add_value(builder, v);
}

static void sax_boolean(json_parser_t *parser, void *userdata, bool value, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    json_value_t *v = json_new_bool(value);
    if (!v) { builder->has_error = true; return; }
    dom_builder_add_value(builder, v);
}

static void sax_null(json_parser_t *parser, void *userdata, json_pos_t pos)
{
    (void)parser; (void)pos;
    dom_builder_t *builder = (dom_builder_t *)userdata;
    json_value_t *v = json_new_null();
    if (!v) { builder->has_error = true; return; }
    dom_builder_add_value(builder, v);
}

json_value_t *json_parse_dom(FILE *input, json_error_t *error)
{
    dom_builder_t builder;
    dom_builder_init(&builder);
    
    json_sax_callbacks_t callbacks = {0};
    callbacks.object_start = sax_object_start;
    callbacks.object_end = sax_object_end;
    callbacks.array_start = sax_array_start;
    callbacks.array_end = sax_array_end;
    callbacks.key = sax_key;
    callbacks.string = sax_string;
    callbacks.number_int = sax_number_int;
    callbacks.number_float = sax_number_float;
    callbacks.boolean = sax_boolean;
    callbacks.null = sax_null;
    
    json_parser_t parser;
    json_parser_init(&parser, input, &callbacks, &builder);
    
    json_error_t err = json_parse_sax(&parser);
    
    if (builder.stack) free(builder.stack);
    if (builder.pending_key) free(builder.pending_key);
    
    if (err != JSON_OK || builder.has_error) {
        if (error) *error = (err != JSON_OK) ? err : JSON_ERROR_MEMORY;
        json_free(builder.root);
        return NULL;
    }
    
    if (error) *error = JSON_OK;
    return builder.root;
}

void json_free(json_value_t *value)
{
    if (!value) return;
    
    switch (value->type) {
        case JSON_STRING:
            free(value->u.str_val);
            break;
        case JSON_ARRAY:
            if (value->u.array_val) {
                for (size_t i = 0; i < value->u.array_val->count; i++) {
                    json_free(value->u.array_val->elements[i]);
                }
                free(value->u.array_val->elements);
                free(value->u.array_val);
            }
            break;
        case JSON_OBJECT:
            if (value->u.object_val) {
                for (size_t i = 0; i < value->u.object_val->count; i++) {
                    free(value->u.object_val->members[i].key);
                    json_free(value->u.object_val->members[i].value);
                }
                free(value->u.object_val->members);
                free(value->u.object_val);
            }
            break;
        default:
            break;
    }
    free(value);
}

json_type_t json_type(const json_value_t *value)
{
    if (!value) return JSON_NULL;
    return value->type;
}

bool json_get_bool(const json_value_t *value, bool *out)
{
    if (!value || value->type != JSON_BOOL || !out) return false;
    *out = value->u.bool_val;
    return true;
}

bool json_get_int(const json_value_t *value, long long *out)
{
    if (!value || value->type != JSON_INT || !out) return false;
    *out = value->u.int_val;
    return true;
}

bool json_get_float(const json_value_t *value, double *out)
{
    if (!value || !out) return false;
    if (value->type == JSON_FLOAT) {
        *out = value->u.float_val;
        return true;
    }
    if (value->type == JSON_INT) {
        *out = (double)value->u.int_val;
        return true;
    }
    return false;
}

bool json_get_string(const json_value_t *value, const char **out)
{
    if (!value || value->type != JSON_STRING || !out) return false;
    *out = value->u.str_val;
    return true;
}

size_t json_array_length(const json_value_t *array)
{
    if (!array || array->type != JSON_ARRAY || !array->u.array_val)
        return 0;
    return array->u.array_val->count;
}

json_value_t *json_get_at(const json_value_t *array, size_t index)
{
    if (!array || array->type != JSON_ARRAY || !array->u.array_val)
        return NULL;
    if (index >= array->u.array_val->count)
        return NULL;
    return array->u.array_val->elements[index];
}

size_t json_object_size(const json_value_t *obj)
{
    if (!obj || obj->type != JSON_OBJECT || !obj->u.object_val)
        return 0;
    return obj->u.object_val->count;
}

json_value_t *json_get(const json_value_t *obj, const char *key)
{
    if (!obj || obj->type != JSON_OBJECT || !obj->u.object_val || !key)
        return NULL;
    
    for (size_t i = 0; i < obj->u.object_val->count; i++) {
        if (strcmp(obj->u.object_val->members[i].key, key) == 0)
            return obj->u.object_val->members[i].value;
    }
    return NULL;
}
