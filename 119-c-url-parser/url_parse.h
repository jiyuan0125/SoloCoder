#ifndef URL_PARSE_H
#define URL_PARSE_H

#include "query_params.h"

#define URL_PARSE_OK 0
#define URL_PARSE_ERROR -1
#define URL_PARSE_INVALID_PROTOCOL -2
#define URL_PARSE_INVALID_HOST -3
#define URL_PARSE_INVALID_PORT -4
#define URL_PARSE_PATH_TRAVERSAL -5

typedef enum {
    URL_PROTOCOL_HTTP,
    URL_PROTOCOL_HTTPS,
    URL_PROTOCOL_UNKNOWN
} URLProtocol;

typedef struct {
    URLProtocol protocol;
    char *protocol_str;
    char *username;
    char *password;
    char *hostname;
    char *port_str;
    int port;
    char *path;
    char *normalized_path;
    char *fragment;
    QueryParams *query_params;
    int path_traversal;
    int valid;
} URL;

URL *url_parse(const char *url_str);
void url_destroy(URL *url);
int url_get_default_port(URLProtocol protocol);
const char *url_protocol_to_string(URLProtocol protocol);
URLProtocol url_string_to_protocol(const char *str);
char *normalize_path(const char *path, int *path_traversal);

#endif
