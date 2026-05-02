#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include "smtp_client.h"

int main(void) {
    printf("SMTP Client Library Test\n");
    printf("========================\n\n");

    /* 测试 base64 编解码 */
    const char *test_str = "Hello, World! 你好，世界！";
    size_t enc_len;
    char *encoded = base64_encode((const unsigned char *)test_str, strlen(test_str), &enc_len);
    
    printf("Original: %s\n", test_str);
    if (encoded) {
        printf("Encoded:  %s\n", encoded);
        
        size_t dec_len;
        unsigned char *decoded = base64_decode(encoded, strlen(encoded), &dec_len);
        if (decoded) {
            decoded[dec_len] = '\0';
            printf("Decoded:  %s\n", (char *)decoded);
            free(decoded);
        }
        free(encoded);
    }

    /* 初始化 SMTP 客户端 */
    if (smtp_client_init() != SMTP_OK) {
        printf("smtp_client_init failed: %s\n", smtp_last_error());
        return 1;
    }
    printf("\nSMTP client initialized successfully.\n");

    /* 配置示例（不会实际连接） */
    smtp_config_t config = {
        .server = "smtp.example.com",
        .port = 587,
        .username = "user@example.com",
        .password = "password",
        .use_tls = 1,
        .use_auth = 1
    };
    smtp_set_config(&config);

    printf("Configured server: %s:%d\n", config.server, config.port);
    printf("Username: %s\n", config.username);

    /* 清理 */
    smtp_client_cleanup();
    printf("\nTest completed.\n");

    return 0;
}
