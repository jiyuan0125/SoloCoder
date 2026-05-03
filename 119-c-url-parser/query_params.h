#ifndef QUERY_PARAMS_H
#define QUERY_PARAMS_H

typedef struct {
    char *key;
    char **values;
    int value_count;
} QueryParam;

typedef struct {
    QueryParam *params;
    int count;
    int capacity;
} QueryParams;

QueryParams *query_params_create(void);
void query_params_destroy(QueryParams *qp);
int query_params_add(QueryParams *qp, const char *key, const char *value);
const char *query_params_get(const QueryParams *qp, const char *key);
const char **query_params_get_all(const QueryParams *qp, const char *key, int *count);
int query_params_parse(QueryParams *qp, const char *query_string);
char *query_params_to_string(const QueryParams *qp, int encode_values);

#endif
