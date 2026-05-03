#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "http_alert.h"
#include "http_request.h"
#include "retry_scheduler.h"

static void print_help(const char *program_name) {
    printf("HTTP Alert Push Module - Demo Program\n");
    printf("Usage: %s [options]\n", program_name);
    printf("\nOptions:\n");
    printf("  -u <url>       Target webhook URL (default: http://httpbin.org/post)\n");
    printf("  -t <title>     Alert title (default: Test Alert)\n");
    printf("  -m <message>   Alert message (default: This is a test alert message)\n");
    printf("  -l <level>     Alert level: info, warning, error, critical (default: warning)\n");
    printf("  -s <source>    Alert source (default: monitoring-system)\n");
    printf("  -c <timeout>   Connection timeout in seconds (default: 5)\n");
    printf("  -r <timeout>   Read timeout in seconds (default: 10)\n");
    printf("  -R <retries>   Max retry count (default: 3)\n");
    printf("  -o <timeout>   Overall timeout in seconds (default: 30)\n");
    printf("  -h             Show this help message\n");
    printf("\nExamples:\n");
    printf("  %s\n", program_name);
    printf("  %s -u http://example.com/webhook -t \"Server Down\" -m \"Web server not responding\" -l error\n", program_name);
    printf("  %s -u https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx ...\n", program_name);
}

static void print_response(HttpResponse *resp) {
    if (resp == NULL) {
        printf("No response\n");
        return;
    }
    
    printf("\n=== HTTP Response ===\n");
    printf("Status Code: %d\n", resp->status_code);
    if (resp->status_message) {
        printf("Status Message: %s\n", resp->status_message);
    }
    
    printf("\nHeaders:\n");
    HttpHeader *hdr = resp->headers;
    while (hdr != NULL) {
        printf("  %s: %s\n", hdr->key, hdr->value);
        hdr = hdr->next;
    }
    
    if (resp->body && resp->body_len > 0) {
        printf("\nBody (%zu bytes):\n", resp->body_len);
        printf("%.*s\n", (int)resp->body_len, resp->body);
    }
}

int main(int argc, char *argv[]) {
    const char *url = "http://httpbin.org/post";
    const char *title = "Test Alert";
    const char *message = "This is a test alert message from the monitoring system.";
    const char *level = "warning";
    const char *source = "monitoring-system";
    
    int connect_timeout = 5;
    int read_timeout = 10;
    int max_retries = 3;
    int overall_timeout = 30;
    
    for (int i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-h") == 0 || strcmp(argv[i], "--help") == 0) {
            print_help(argv[0]);
            return 0;
        } else if (strcmp(argv[i], "-u") == 0 && i + 1 < argc) {
            url = argv[++i];
        } else if (strcmp(argv[i], "-t") == 0 && i + 1 < argc) {
            title = argv[++i];
        } else if (strcmp(argv[i], "-m") == 0 && i + 1 < argc) {
            message = argv[++i];
        } else if (strcmp(argv[i], "-l") == 0 && i + 1 < argc) {
            level = argv[++i];
        } else if (strcmp(argv[i], "-s") == 0 && i + 1 < argc) {
            source = argv[++i];
        } else if (strcmp(argv[i], "-c") == 0 && i + 1 < argc) {
            connect_timeout = atoi(argv[++i]);
        } else if (strcmp(argv[i], "-r") == 0 && i + 1 < argc) {
            read_timeout = atoi(argv[++i]);
        } else if (strcmp(argv[i], "-R") == 0 && i + 1 < argc) {
            max_retries = atoi(argv[++i]);
        } else if (strcmp(argv[i], "-o") == 0 && i + 1 < argc) {
            overall_timeout = atoi(argv[++i]);
        }
    }
    
    printf("=== HTTP Alert Push Module - Demo ===\n\n");
    
    AlertData alert;
    memset(&alert, 0, sizeof(alert));
    alert.title = (char *)title;
    alert.message = (char *)message;
    alert.level = (char *)level;
    alert.source = (char *)source;
    alert.timestamp = time(NULL);
    
    TimeoutConfig timeout_cfg;
    memset(&timeout_cfg, 0, sizeof(timeout_cfg));
    timeout_cfg.connect_timeout_sec = connect_timeout;
    timeout_cfg.connect_timeout_usec = 0;
    timeout_cfg.read_timeout_sec = read_timeout;
    timeout_cfg.read_timeout_usec = 0;
    
    RetryConfig retry_cfg;
    memset(&retry_cfg, 0, sizeof(retry_cfg));
    retry_cfg.max_retries = max_retries;
    retry_cfg.initial_delay_sec = 1;
    retry_cfg.overall_timeout_sec = overall_timeout;
    
    printf("Configuration:\n");
    printf("  URL: %s\n", url);
    printf("  Alert Title: %s\n", title);
    printf("  Alert Level: %s\n", level);
    printf("  Connection Timeout: %d sec\n", connect_timeout);
    printf("  Read Timeout: %d sec\n", read_timeout);
    printf("  Max Retries: %d\n", max_retries);
    printf("  Overall Timeout: %d sec\n\n", overall_timeout);
    
    printf("Sending alert...\n");
    
    HttpResponse *response = NULL;
    int ret = push_alert_json(url, &alert, &timeout_cfg, &retry_cfg, &response);
    
    if (ret == HTTP_ALERT_OK && response != NULL) {
        printf("\n*** Alert pushed successfully! ***\n");
        print_response(response);
        http_alert_free_response(response);
        return 0;
    } else {
        printf("\n*** Failed to push alert ***\n");
        printf("Error: %s (code: %d)\n", http_alert_strerror(ret), ret);
        if (response != NULL) {
            print_response(response);
            http_alert_free_response(response);
        }
        return 1;
    }
}
