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

typedef void (*AlertCallback)(const char *webhook_url, 
                                const AlertData *alert, 
                                int attempt,
                                int total_attempts,
                                int success, 
                                int error_code,
                                HttpResponse *response);

typedef struct AlertContext {
    const char *webhook_url;
    const AlertData *alert;
    AlertCallback callback;
    void *user_data;
} AlertContext;

RetryResult retry_scheduler_execute(const HttpRequest *req, 
                                     const TimeoutConfig *timeout,
                                     const RetryConfig *retry_config,
                                     AlertContext *ctx);

int push_alert_with_retry(const char *webhook_url,
                           HttpMethod method,
                           const HttpHeader *custom_headers,
                           const char *body,
                           size_t body_len,
                           const TimeoutConfig *timeout,
                           const RetryConfig *retry_config,
                           AlertCallback callback,
                           void *user_data,
                           HttpResponse **out_response);

int push_alert_json(const char *webhook_url,
                     const AlertData *alert,
                     const TimeoutConfig *timeout,
                     const RetryConfig *retry_config,
                     AlertCallback callback,
                     void *user_data,
                     HttpResponse **out_response);

void retry_result_free(RetryResult *result);

#endif
