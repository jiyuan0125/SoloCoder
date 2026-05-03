#include "query_params.h"
#include "url_codec.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define INITIAL_CAPACITY 8
#define INITIAL_VALUE_CAPACITY 2

QueryParams *query_params_create(void) {
    QueryParams *qp = (QueryParams *)malloc(sizeof(QueryParams));
    if (!qp) return NULL;
    
    qp->params = (QueryParam *)malloc(sizeof(QueryParam) * INITIAL_CAPACITY);
    if (!qp->params) {
        free(qp);
        return NULL;
    }
    
    qp->count = 0;
    qp->capacity = INITIAL_CAPACITY;
    memset(qp->params, 0, sizeof(QueryParam) * INITIAL_CAPACITY);
    
    return qp;
}

void query_params_destroy(QueryParams *qp) {
    if (!qp) return;
    
    for (int i = 0; i < qp->count; i++) {
        free(qp->params[i].key);
        for (int j = 0; j < qp->params[i].value_count; j++) {
            free(qp->params[i].values[j]);
        }
        free(qp->params[i].values);
    }
    free(qp->params);
    free(qp);
}

static QueryParam *find_param(QueryParams *qp, const char *key) {
    for (int i = 0; i < qp->count; i++) {
        if (strcmp(qp->params[i].key, key) == 0) {
            return &qp->params[i];
        }
    }
    return NULL;
}

static int ensure_capacity(QueryParams *qp) {
    if (qp->count >= qp->capacity) {
        int new_capacity = qp->capacity * 2;
        QueryParam *new_params = (QueryParam *)realloc(qp->params, sizeof(QueryParam) * new_capacity);
        if (!new_params) return -1;
        memset(&new_params[qp->count], 0, sizeof(QueryParam) * (new_capacity - qp->count));
        qp->params = new_params;
        qp->capacity = new_capacity;
    }
    return 0;
}

int query_params_add(QueryParams *qp, const char *key, const char *value) {
    if (!qp || !key) return -1;
    
    QueryParam *param = find_param(qp, key);
    if (param) {
        if (param->value_count == 0) {
            param->values = (char **)malloc(sizeof(char *) * INITIAL_VALUE_CAPACITY);
            if (!param->values) return -1;
        } else if (param->value_count >= 2 && param->value_count % 2 == 0) {
            char **new_values = (char **)realloc(param->values, sizeof(char *) * (param->value_count + 2));
            if (!new_values) return -1;
            param->values = new_values;
        }
        
        param->values[param->value_count] = value ? strdup(value) : NULL;
        param->value_count++;
        return 0;
    }
    
    if (ensure_capacity(qp) < 0) return -1;
    
    int idx = qp->count;
    qp->params[idx].key = strdup(key);
    if (!qp->params[idx].key) return -1;
    
    qp->params[idx].values = (char **)malloc(sizeof(char *) * INITIAL_VALUE_CAPACITY);
    if (!qp->params[idx].values) {
        free(qp->params[idx].key);
        return -1;
    }
    
    qp->params[idx].values[0] = value ? strdup(value) : NULL;
    qp->params[idx].value_count = 1;
    qp->count++;
    
    return 0;
}

const char *query_params_get(const QueryParams *qp, const char *key) {
    if (!qp || !key) return NULL;
    
    for (int i = 0; i < qp->count; i++) {
        if (strcmp(qp->params[i].key, key) == 0 && qp->params[i].value_count > 0) {
            return qp->params[i].values[0];
        }
    }
    return NULL;
}

const char **query_params_get_all(const QueryParams *qp, const char *key, int *count) {
    if (!qp || !key || !count) return NULL;
    *count = 0;
    
    for (int i = 0; i < qp->count; i++) {
        if (strcmp(qp->params[i].key, key) == 0) {
            *count = qp->params[i].value_count;
            return (const char **)qp->params[i].values;
        }
    }
    return NULL;
}

int query_params_parse(QueryParams *qp, const char *query_string) {
    if (!qp || !query_string) return -1;
    
    const char *p = query_string;
    
    if (*p == '?') p++;
    
    while (*p) {
        const char *eq = NULL;
        const char *key_start = p;
        size_t key_len = 0;
        
        while (*p && *p != '=' && *p != '&') {
            p++;
        }
        
        if (*p == '=') {
            eq = p;
            key_len = eq - key_start;
            p++;
        } else {
            key_len = p - key_start;
        }
        
        char *key = NULL;
        if (key_len > 0) {
            key = (char *)malloc(key_len + 1);
            if (!key) return -1;
            strncpy(key, key_start, key_len);
            key[key_len] = '\0';
        }
        
        char *decoded_key = key ? url_decode_alloc(key) : NULL;
        free(key);
        
        const char *value_start = p;
        size_t value_len = 0;
        
        while (*p && *p != '&') {
            p++;
        }
        
        value_len = p - value_start;
        
        char *value = NULL;
        if (value_len > 0 && eq) {
            value = (char *)malloc(value_len + 1);
            if (!value) {
                free(decoded_key);
                return -1;
            }
            strncpy(value, value_start, value_len);
            value[value_len] = '\0';
        }
        
        char *decoded_value = value ? url_decode_alloc(value) : NULL;
        free(value);
        
        if (decoded_key && decoded_key[0] != '\0') {
            query_params_add(qp, decoded_key, decoded_value ? decoded_value : "");
        }
        
        free(decoded_key);
        free(decoded_value);
        
        if (*p == '&') p++;
    }
    
    return 0;
}

char *query_params_to_string(const QueryParams *qp, int encode_values) {
    if (!qp || qp->count == 0) return strdup("");
    
    size_t total_len = 1;
    
    for (int i = 0; i < qp->count; i++) {
        const QueryParam *param = &qp->params[i];
        for (int j = 0; j < param->value_count; j++) {
            if (i > 0 || j > 0) total_len++;
            
            size_t key_enc_len = encode_values ? 
                url_encode(param->key, NULL, 0) : strlen(param->key);
            total_len += key_enc_len;
            
            if (param->values[j]) {
                total_len++;
                size_t val_enc_len = encode_values ? 
                    url_encode(param->values[j], NULL, 0) : strlen(param->values[j]);
                total_len += val_enc_len;
            }
        }
    }
    
    char *result = (char *)malloc(total_len);
    if (!result) return NULL;
    
    char *dst = result;
    
    for (int i = 0; i < qp->count; i++) {
        const QueryParam *param = &qp->params[i];
        for (int j = 0; j < param->value_count; j++) {
            if (dst != result) {
                *dst++ = '&';
            }
            
            if (encode_values) {
                size_t enc_len = url_encode(param->key, dst, total_len - (dst - result));
                dst += enc_len;
            } else {
                size_t len = strlen(param->key);
                strcpy(dst, param->key);
                dst += len;
            }
            
            if (param->values[j]) {
                *dst++ = '=';
                if (encode_values) {
                    size_t enc_len = url_encode(param->values[j], dst, total_len - (dst - result));
                    dst += enc_len;
                } else {
                    size_t len = strlen(param->values[j]);
                    strcpy(dst, param->values[j]);
                    dst += len;
                }
            }
        }
    }
    
    *dst = '\0';
    return result;
}
