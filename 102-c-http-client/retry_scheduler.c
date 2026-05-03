#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include "retry_scheduler.h"

static int is_success_response(HttpResponse *resp) {
    if (resp == NULL) return 0;
    return resp->status_code >= 200 && resp->status_code < 300;
}

static int should_retry(int error_code, HttpResponse *resp) {
    if (error_code == HTTP_ALERT_OK && resp != NULL) {
        if (resp->status_code >= 500 && resp->status_code < 600) {
            return 1;
        }
        if (resp->status_code == 429) {
            return 1;
        }
        return 0;
    }
    
    switch (error_code) {
        case HTTP_ALERT_ERR_CONNECT:
        case HTTP_ALERT_ERR_CONNECT_TIMEOUT:
        case HTTP_ALERT_ERR_RECV_TIMEOUT:
        case HTTP_ALERT_ERR_SEND:
        case HTTP_ALERT_ERR_RECV:
        case HTTP_ALERT_ERR_RESOLVE:
        case HTTP_ALERT_ERR_SSL_CONNECT:
            return 1;
        default:
            return 0;
    }
}

static void sleep_seconds(int seconds) {
    struct timespec req, rem;
    req.tv_sec = seconds;
    req.tv_nsec = 0;
    
    while (nanosleep(&req, &rem) == -1) {
        req = rem;
    }
}

RetryResult retry_scheduler_execute(const HttpRequest *req, 
                                     const TimeoutConfig *timeout,
                                     const RetryConfig *retry_config,
                                     AlertContext *ctx) {
    RetryResult result = {0};
    result.success = 0;
    result.last_error = HTTP_ALERT_ERR_INVALID_ARG;
    result.attempt_count = 0;
    result.response = NULL;
    
    if (req == NULL || timeout == NULL || retry_config == NULL) {
        return result;
    }
    
    time_t overall_start = time(NULL);
    int max_retries = retry_config->max_retries > 0 ? retry_config->max_retries : 3;
    int total_attempts = max_retries + 1;
    int delay = retry_config->initial_delay_sec > 0 ? retry_config->initial_delay_sec : 1;
    
    for (int attempt = 0; attempt <= max_retries; attempt++) {
        result.attempt_count = attempt + 1;
        
        if (retry_config->overall_timeout_sec > 0) {
            time_t elapsed = time(NULL) - overall_start;
            if (elapsed >= retry_config->overall_timeout_sec) {
                http_alert_log("WARN", "Overall timeout exceeded after %d seconds, giving up",
                              (int)elapsed);
                result.last_error = HTTP_ALERT_ERR_OVERALL_TIMEOUT;
                
                if (ctx != NULL && ctx->callback != NULL) {
                    ctx->callback(ctx->webhook_url,
                                  ctx->alert,
                                  result.attempt_count,
                                  total_attempts,
                                  0,
                                  result.last_error,
                                  NULL);
                }
                break;
            }
        }
        
        if (result.response != NULL) {
            http_alert_free_response(result.response);
            result.response = NULL;
        }
        
        result.response = http_alert_create_response();
        if (result.response == NULL) {
            result.last_error = HTTP_ALERT_ERR_MEMORY;
            http_alert_log("ERROR", "Failed to create response object");
            
            if (ctx != NULL && ctx->callback != NULL) {
                ctx->callback(ctx->webhook_url,
                              ctx->alert,
                              result.attempt_count,
                              total_attempts,
                              0,
                              result.last_error,
                              NULL);
            }
            break;
        }
        
        if (attempt > 0) {
            http_alert_log("INFO", "Retry attempt %d/%d, delaying %d seconds...",
                          attempt, max_retries, delay);
            sleep_seconds(delay);
            delay *= 2;
        }
        
        http_alert_log("INFO", "Attempt %d: %s %s",
                      result.attempt_count,
                      http_method_to_string(req->method),
                      req->url ? req->url : req->path);
        
        int ret = http_request_execute(req, result.response, timeout);
        result.last_error = ret;
        
        int attempt_success = 0;
        if (ret == HTTP_ALERT_OK) {
            http_alert_log("INFO", "Attempt %d: HTTP %d",
                          result.attempt_count, result.response->status_code);
            
            if (is_success_response(result.response)) {
                result.success = 1;
                attempt_success = 1;
            } else {
                result.last_error = HTTP_ALERT_ERR_HTTP_STATUS;
            }
        } else {
            http_alert_log("ERROR", "Attempt %d failed: %s",
                          result.attempt_count, http_alert_strerror(ret));
        }
        
        if (ctx != NULL && ctx->callback != NULL) {
            ctx->callback(ctx->webhook_url,
                          ctx->alert,
                          result.attempt_count,
                          total_attempts,
                          attempt_success,
                          result.last_error,
                          result.response);
        }
        
        if (result.success) {
            break;
        }
        
        if (attempt < max_retries && should_retry(ret, result.response)) {
            continue;
        }
        
        break;
    }
    
    if (!result.success) {
        http_alert_log("ERROR", "All %d attempts failed. Last error: %s",
                      result.attempt_count, http_alert_strerror(result.last_error));
    }
    
    return result;
}

void retry_result_free(RetryResult *result) {
    if (result == NULL) return;
    if (result->response != NULL) {
        http_alert_free_response(result->response);
        result->response = NULL;
    }
}

int push_alert_with_retry(const char *webhook_url,
                           HttpMethod method,
                           const HttpHeader *custom_headers,
                           const char *body,
                           size_t body_len,
                           const TimeoutConfig *timeout,
                           const RetryConfig *retry_config,
                           AlertCallback callback,
                           void *user_data,
                           HttpResponse **out_response) {
    if (webhook_url == NULL || timeout == NULL || retry_config == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    HttpRequest *req = http_alert_create_request();
    if (req == NULL) {
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    req->method = method;
    
    int ret = http_alert_parse_url(webhook_url, req);
    if (ret != HTTP_ALERT_OK) {
        http_alert_free_request(req);
        return ret;
    }
    
    if (custom_headers != NULL) {
        req->headers = http_alert_copy_headers(custom_headers);
    }
    
    if (body != NULL && body_len > 0) {
        ret = http_alert_set_body(req, body, body_len);
        if (ret != HTTP_ALERT_OK) {
            http_alert_free_request(req);
            return ret;
        }
    }
    
    AlertContext ctx = {0};
    ctx.webhook_url = webhook_url;
    ctx.alert = NULL;
    ctx.callback = callback;
    ctx.user_data = user_data;
    
    RetryResult result = retry_scheduler_execute(req, timeout, retry_config, &ctx);
    
    if (out_response != NULL) {
        *out_response = result.response;
        result.response = NULL;
    }
    
    http_alert_free_request(req);
    retry_result_free(&result);
    
    return result.success ? HTTP_ALERT_OK : result.last_error;
}

int push_alert_json(const char *webhook_url,
                     const AlertData *alert,
                     const TimeoutConfig *timeout,
                     const RetryConfig *retry_config,
                     AlertCallback callback,
                     void *user_data,
                     HttpResponse **out_response) {
    if (webhook_url == NULL || alert == NULL || timeout == NULL || retry_config == NULL) {
        return HTTP_ALERT_ERR_INVALID_ARG;
    }
    
    char *json_body = alert_data_to_json(alert);
    if (json_body == NULL) {
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    HttpHeader *headers = NULL;
    HttpHeader *h1 = (HttpHeader *)malloc(sizeof(HttpHeader));
    HttpHeader *h2 = (HttpHeader *)malloc(sizeof(HttpHeader));
    
    if (h1 == NULL || h2 == NULL) {
        free(h1);
        free(h2);
        free(json_body);
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    h1->key = strdup("Content-Type");
    h1->value = strdup("application/json");
    h1->next = NULL;
    
    h2->key = strdup("Accept");
    h2->value = strdup("application/json");
    h2->next = NULL;
    
    if (h1->key == NULL || h1->value == NULL || h2->key == NULL || h2->value == NULL) {
        http_alert_free_headers(h1);
        http_alert_free_headers(h2);
        free(json_body);
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    h1->next = h2;
    headers = h1;
    
    AlertContext ctx = {0};
    ctx.webhook_url = webhook_url;
    ctx.alert = alert;
    ctx.callback = callback;
    ctx.user_data = user_data;
    
    HttpRequest *req = http_alert_create_request();
    if (req == NULL) {
        http_alert_free_headers(headers);
        free(json_body);
        return HTTP_ALERT_ERR_MEMORY;
    }
    
    req->method = HTTP_METHOD_POST;
    
    int ret = http_alert_parse_url(webhook_url, req);
    if (ret != HTTP_ALERT_OK) {
        http_alert_free_request(req);
        http_alert_free_headers(headers);
        free(json_body);
        return ret;
    }
    
    req->headers = http_alert_copy_headers(headers);
    
    ret = http_alert_set_body(req, json_body, strlen(json_body));
    if (ret != HTTP_ALERT_OK) {
        http_alert_free_request(req);
        http_alert_free_headers(headers);
        free(json_body);
        return ret;
    }
    
    RetryResult result = retry_scheduler_execute(req, timeout, retry_config, &ctx);
    
    if (out_response != NULL) {
        *out_response = result.response;
        result.response = NULL;
    }
    
    http_alert_free_request(req);
    http_alert_free_headers(headers);
    free(json_body);
    retry_result_free(&result);
    
    return result.success ? HTTP_ALERT_OK : result.last_error;
}
