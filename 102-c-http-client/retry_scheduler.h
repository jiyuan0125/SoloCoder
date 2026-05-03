#ifndef RETRY_SCHEDULER_H
#define RETRY_SCHEDULER_H

#include "http_alert.h"
#include "http_request.h"

typedef struct RetryResult {
    int success;
    int last_error;
    int attempt_count;
    HttpResponse *response;
} RetryResult;

typedef void (*AlertCallback)(const char *webhook_url, const AlertData *alert, int success, int error_code);

RetryResult retry_scheduler_execute(const HttpRequest *req, 
                                     const TimeoutConfig *timeout,
                                     const RetryConfig *retry_config);

int push_alert_with_retry(const char *webhook_url,
                           HttpMethod method,
                           const HttpHeader *custom_headers,
                           const char *body,
                           size_t body_len,
                           const TimeoutConfig *timeout,
                           const RetryConfig *retry_config,
                           HttpResponse **out_response);

int push_alert_json(const char *webhook_url,
                     const AlertData *alert,
                     const TimeoutConfig *timeout,
                     const RetryConfig *retry_config,
                     HttpResponse **out_response);

void retry_result_free(RetryResult *result);

#endif
