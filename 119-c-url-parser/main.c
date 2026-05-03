#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "url_codec.h"
#include "query_params.h"
#include "url_parse.h"

static void print_separator(void) {
    printf("============================================================\n");
}

static void test_url_encode_decode(void) {
    printf("\n=== 测试 URL 编码/解码 ===\n");
    print_separator();
    
    const char *test_cases[] = {
        "hello world",
        "中文测试",
        "key=value&another=test",
        "a/b/c",
        "user@example.com",
        "test~-_."
    };
    
    int num_cases = sizeof(test_cases) / sizeof(test_cases[0]);
    
    for (int i = 0; i < num_cases; i++) {
        const char *original = test_cases[i];
        char *encoded = url_encode_alloc(original);
        char *decoded = url_decode_alloc(encoded);
        
        printf("原文: \"%s\"\n", original);
        printf("编码: \"%s\"\n", encoded);
        printf("解码: \"%s\"\n", decoded);
        
        if (strcmp(original, decoded) == 0) {
            printf("状态: 编解码一致 ✓\n");
        } else {
            printf("状态: 编解码不一致 ✗\n");
        }
        printf("---\n");
        
        free(encoded);
        free(decoded);
    }
    
    const char *invalid_hex = "test%XXinvalid";
    char *decoded_invalid = url_decode_alloc(invalid_hex);
    printf("无效十六进制测试:\n");
    printf("原文: \"%s\"\n", invalid_hex);
    printf("解码: \"%s\"\n", decoded_invalid);
    free(decoded_invalid);
}

static void test_query_params(void) {
    printf("\n=== 测试查询参数解析 ===\n");
    print_separator();
    
    const char *test_queries[] = {
        "?name=john&age=25",
        "?key=%E4%B8%AD%E6%96%87&value=test",
        "?a=1&a=2&a=3",
        "?empty_key=&normal=value",
        "?key_only",
        "?special=%26%3D%25"
    };
    
    int num_queries = sizeof(test_queries) / sizeof(test_queries[0]);
    
    for (int i = 0; i < num_queries; i++) {
        const char *query = test_queries[i];
        printf("查询字符串: \"%s\"\n", query);
        
        QueryParams *qp = query_params_create();
        query_params_parse(qp, query);
        
        printf("解析结果:\n");
        for (int j = 0; j < qp->count; j++) {
            const QueryParam *param = &qp->params[j];
            printf("  \"%s\": ", param->key);
            if (param->value_count == 1) {
                printf("\"%s\"\n", param->values[0] ? param->values[0] : "(null)");
            } else {
                printf("[");
                for (int k = 0; k < param->value_count; k++) {
                    if (k > 0) printf(", ");
                    printf("\"%s\"", param->values[k] ? param->values[k] : "(null)");
                }
                printf("]\n");
            }
        }
        
        char *reconstructed = query_params_to_string(qp, 1);
        printf("重建字符串: \"%s\"\n", reconstructed);
        free(reconstructed);
        
        query_params_destroy(qp);
        printf("---\n");
    }
}

static void print_url_info(const URL *url) {
    printf("URL 有效性: %s\n", url->valid ? "有效" : "无效");
    
    printf("协议: ");
    if (url->protocol_str) {
        printf("%s", url->protocol_str);
        if (url->protocol != URL_PROTOCOL_UNKNOWN) {
            printf(" (已知协议)");
        } else {
            printf(" (未知协议)");
        }
    } else {
        printf("(未解析)");
    }
    printf("\n");
    
    printf("用户名: %s\n", url->username ? url->username : "(无)");
    printf("密码: %s\n", url->password ? url->password : "(无)");
    printf("主机名: %s\n", url->hostname ? url->hostname : "(无)");
    
    printf("端口: ");
    if (url->port_str) {
        printf("%s (显式指定)", url->port_str);
    } else if (url->port >= 0) {
        printf("%d (协议默认)", url->port);
    } else {
        printf("(无)");
    }
    printf("\n");
    
    printf("原始路径: %s\n", url->path ? url->path : "(无)");
    printf("归一化路径: %s\n", url->normalized_path ? url->normalized_path : "(无)");
    printf("路径穿越: %s\n", url->path_traversal ? "是 (危险!)" : "否");
    
    printf("片段: %s\n", url->fragment ? url->fragment : "(无)");
    
    printf("查询参数: ");
    if (url->query_params && url->query_params->count > 0) {
        printf("\n");
        for (int i = 0; i < url->query_params->count; i++) {
            const QueryParam *param = &url->query_params->params[i];
            printf("    \"%s\": ", param->key);
            if (param->value_count == 1) {
                printf("\"%s\"\n", param->values[0] ? param->values[0] : "");
            } else {
                printf("[");
                for (int k = 0; k < param->value_count; k++) {
                    if (k > 0) printf(", ");
                    printf("\"%s\"", param->values[k] ? param->values[k] : "");
                }
                printf("]\n");
            }
        }
    } else {
        printf("(无)\n");
    }
}

static void test_url_parse(void) {
    printf("\n=== 测试 URL 解析 ===\n");
    print_separator();
    
    const char *test_urls[] = {
        "http://www.example.com/",
        "https://user:pass@example.com:8080/path/to/resource?query=1#fragment",
        "http://[::1]:8080/api/v1/test",
        "http://example.com/a/b/../c/./d",
        "http://example.com/a/../../etc/passwd",
        "https://example.com?name=%E4%B8%AD%E6%96%87&lang=zh",
        "http://example.com?a=1&a=2&b=3",
        "http://user@example.com:443/path",
        "https://example.com/path?email=user%40example.com",
        "http://example.com:80/path#section1"
    };
    
    int num_urls = sizeof(test_urls) / sizeof(test_urls[0]);
    
    for (int i = 0; i < num_urls; i++) {
        const char *url_str = test_urls[i];
        printf("URL [%d/%d]: %s\n", i + 1, num_urls, url_str);
        print_separator();
        
        URL *url = url_parse(url_str);
        if (url) {
            print_url_info(url);
            url_destroy(url);
        } else {
            printf("解析失败!\n");
        }
        
        if (i < num_urls - 1) {
            printf("\n");
        }
    }
}

static void test_at_symbol_edge_case(void) {
    printf("\n=== 测试 @ 符号边界情况 ===\n");
    print_separator();
    
    const char *test_cases[] = {
        "http://user:pass@example.com/path",
        "http://example.com/path/user@domain",
        "http://user@example.com@another.com/path",
        "http://example.com?email=user@example.com"
    };
    
    int num_cases = sizeof(test_cases) / sizeof(test_cases[0]);
    
    for (int i = 0; i < num_cases; i++) {
        const char *url_str = test_cases[i];
        printf("URL: %s\n", url_str);
        
        URL *url = url_parse(url_str);
        if (url) {
            printf("  用户名: %s\n", url->username ? url->username : "(无)");
            printf("  密码: %s\n", url->password ? url->password : "(无)");
            printf("  主机名: %s\n", url->hostname ? url->hostname : "(无)");
            printf("  路径: %s\n", url->path ? url->path : "(无)");
            printf("  查询参数计数: %d\n", url->query_params ? url->query_params->count : 0);
            url_destroy(url);
        }
        printf("---\n");
    }
}

static void test_path_normalization(void) {
    printf("\n=== 测试路径归一化 ===\n");
    print_separator();
    
    const char *paths[] = {
        "/a/b/c",
        "/a/./b/./c",
        "/a/b/../c",
        "/a/../../etc/passwd",
        "/a/../b/../../c",
        "/./././",
        "/a//b///c",
        "a/b/c",
        "../a/b"
    };
    
    int num_paths = sizeof(paths) / sizeof(paths[0]);
    
    for (int i = 0; i < num_paths; i++) {
        const char *path = paths[i];
        int traversal = 0;
        
        char *normalized = normalize_path(path, &traversal);
        
        printf("原始路径: \"%s\"\n", path);
        printf("归一化:   \"%s\"\n", normalized ? normalized : "(null)");
        printf("路径穿越: %s\n", traversal ? "是" : "否");
        printf("---\n");
        
        free(normalized);
    }
}

int main(void) {
    printf("\n");
    printf("╔══════════════════════════════════════════════════════════╗\n");
    printf("║           URL 解析模块演示程序                            ║\n");
    printf("║           用于短链接服务后端的 URL 解析功能               ║\n");
    printf("╚══════════════════════════════════════════════════════════╝\n");
    
    test_url_encode_decode();
    test_query_params();
    test_url_parse();
    test_at_symbol_edge_case();
    test_path_normalization();
    
    printf("\n\n=== 所有测试完成 ===\n\n");
    
    return 0;
}
