#ifndef HTTP_REQUEST_H
#define HTTP_REQUEST_H

#include "http_alert.h"
#include "http_connection.h"

int http_request_send(HttpConnection *conn, const HttpRequest *req);
int http_request_receive(HttpConnection *conn, HttpResponse *resp);
int http_request_execute(const HttpRequest *req, HttpResponse *resp, 
                          const TimeoutConfig *timeout);
const char *http_method_to_string(HttpMethod method);
char *alert_data_to_json(const AlertData *alert);

#endif
