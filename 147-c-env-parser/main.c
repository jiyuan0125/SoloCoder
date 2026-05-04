#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "env_loader.h"

static void print_hex(const char *str) {
    if (!str) {
        printf("(null)");
        return;
    }
    const unsigned char *p = (const unsigned char *)str;
    while (*p) {
        if (*p >= 32 && *p <= 126) {
            putchar(*p);
        } else {
            printf("\\x%02x", *p);
        }
        p++;
    }
}

int main(int argc, char *argv[]) {
    const char *env_path = ".env";
    if (argc > 1) {
        env_path = argv[1];
    }
    
    printf("========================================\n");
    printf("  环境变量加载器演示\n");
    printf("========================================\n\n");
    
    env_expand_config_t config = env_expand_default_config();
    config.error_on_undefined = 0;
    
    env_loader_t *loader = env_loader_create(config);
    if (!loader) {
        fprintf(stderr, "错误：无法创建环境加载器\n");
        return 1;
    }
    
    printf("正在加载配置文件: %s\n\n", env_path);
    
    env_loader_error_t err = env_loader_load(loader, env_path);
    if (err != ENV_LOADER_OK) {
        fprintf(stderr, "错误：加载配置文件失败 (错误码: %d)\n", err);
        env_loader_destroy(loader);
        return 1;
    }
    
    printf("----------------------------------------\n");
    printf("  1. 字符串类型读取演示\n");
    printf("----------------------------------------\n");
    
    const char *db_host = env_loader_get_string(loader, "DATABASE_HOST", "localhost");
    printf("DATABASE_HOST = '%s'\n", db_host);
    
    const char *db_name = env_loader_get_string(loader, "DATABASE_NAME", "default_db");
    printf("DATABASE_NAME = '%s'\n", db_name);
    
    const char *password = env_loader_get_string(loader, "PASSWORD", "");
    printf("PASSWORD = '%s'\n", password);
    
    const char *api_key = env_loader_get_string(loader, "API_KEY", "");
    printf("API_KEY = '%s'\n", api_key);
    
    const char *message = env_loader_get_string(loader, "MESSAGE", "");
    printf("MESSAGE = '");
    print_hex(message);
    printf("'\n");
    
    const char *literal_msg = env_loader_get_string(loader, "LITERAL_MESSAGE", "");
    printf("LITERAL_MESSAGE = '");
    print_hex(literal_msg);
    printf("'\n");
    
    const char *multi_line = env_loader_get_string(loader, "MULTI_LINE_TEXT", "");
    printf("MULTI_LINE_TEXT = \n'''%s'''\n", multi_line);
    
    printf("\n----------------------------------------\n");
    printf("  2. 变量展开演示\n");
    printf("----------------------------------------\n");
    
    const char *db_url = env_loader_get_string(loader, "DATABASE_URL", "");
    printf("DATABASE_URL = '%s'\n", db_url);
    
    const char *full_name = env_loader_get_string(loader, "FULL_NAME", "");
    printf("FULL_NAME = '%s'\n", full_name);
    
    const char *service_url = env_loader_get_string(loader, "SERVICE_URL", "");
    printf("SERVICE_URL = '%s'\n", service_url);
    
    const char *undefined_ref = env_loader_get_string(loader, "UNDEFINED_REF", "");
    printf("UNDEFINED_REF = '%s'\n", undefined_ref);
    printf("  说明: $UNDEFINED_VAR_suffix 被解析为变量 UNDEFINED_VAR_suffix（下划线是变量名有效字符）\n");
    printf("        变量不存在，所以展开为空字符串\n");
    
    const char *explicit_boundary = env_loader_get_string(loader, "EXPLICIT_BOUNDARY", "");
    printf("EXPLICIT_BOUNDARY = '%s'\n", explicit_boundary);
    printf("  说明: 使用 ${DATABASE_NAME}_suffix 格式，明确变量名边界\n");
    printf("        ${DATABASE_NAME} 被解析为变量，_suffix 作为普通字符串保留\n");
    
    printf("\n  [验证] 多个 get_string 结果同时使用:\n");
    const char *a = env_loader_get_string(loader, "DATABASE_HOST", "");
    const char *b = env_loader_get_string(loader, "DATABASE_NAME", "");
    printf("  a (DATABASE_HOST) = '%s'\n", a);
    printf("  b (DATABASE_NAME) = '%s'\n", b);
    printf("  a && b 同时有效: a='%s', b='%s'\n", a, b);
    
    printf("\n----------------------------------------\n");
    printf("  3. 整数类型读取演示\n");
    printf("----------------------------------------\n");
    
    int port = env_loader_get_int(loader, "PORT", 8080);
    printf("PORT = %d (默认值: 8080)\n", port);
    
    int max_connections = env_loader_get_int(loader, "MAX_CONNECTIONS", 10);
    printf("MAX_CONNECTIONS = %d (默认值: 10)\n", max_connections);
    
    int invalid_int = env_loader_get_int(loader, "INVALID_INT", -1);
    printf("INVALID_INT = %d (默认值: -1)\n", invalid_int);
    
    int missing_int = env_loader_get_int(loader, "MISSING_INT", 999);
    printf("MISSING_INT = %d (默认值: 999)\n", missing_int);
    
    printf("\n----------------------------------------\n");
    printf("  4. 浮点数类型读取演示\n");
    printf("----------------------------------------\n");
    
    double timeout = env_loader_get_float(loader, "TIMEOUT", 30.0);
    printf("TIMEOUT = %.4f (默认值: 30.0)\n", timeout);
    
    double rate = env_loader_get_float(loader, "RATE_LIMIT", 1.0);
    printf("RATE_LIMIT = %.4f (默认值: 1.0)\n", rate);
    
    double invalid_float = env_loader_get_float(loader, "INVALID_FLOAT", -1.0);
    printf("INVALID_FLOAT = %.4f (默认值: -1.0)\n", invalid_float);
    
    printf("\n----------------------------------------\n");
    printf("  5. 布尔类型读取演示\n");
    printf("----------------------------------------\n");
    
    int debug = env_loader_get_bool(loader, "DEBUG_MODE", 0);
    printf("DEBUG_MODE = %s (默认值: false)\n", debug ? "true" : "false");
    
    int enable_log = env_loader_get_bool(loader, "ENABLE_LOGGING", 0);
    printf("ENABLE_LOGGING = %s (默认值: false)\n", enable_log ? "true" : "false");
    
    int feature_flag = env_loader_get_bool(loader, "FEATURE_FLAG", 0);
    printf("FEATURE_FLAG = %s (默认值: false)\n", feature_flag ? "true" : "false");
    
    int disabled = env_loader_get_bool(loader, "DISABLED", 1);
    printf("DISABLED = %s (默认值: true)\n", disabled ? "true" : "false");
    
    int invalid_bool = env_loader_get_bool(loader, "INVALID_BOOL", 0);
    printf("INVALID_BOOL = %s (默认值: false)\n", invalid_bool ? "true" : "false");
    
    printf("\n----------------------------------------\n");
    printf("  6. 包含等号的值演示\n");
    printf("----------------------------------------\n");
    
    const char *complex_password = env_loader_get_string(loader, "COMPLEX_PASSWORD", "");
    printf("COMPLEX_PASSWORD = '%s'\n", complex_password);
    
    const char *query_string = env_loader_get_string(loader, "QUERY_STRING", "");
    printf("QUERY_STRING = '%s'\n", query_string);
    
    printf("\n----------------------------------------\n");
    printf("  7. 系统环境变量优先级演示\n");
    printf("----------------------------------------\n");
    
    const char *system_env = getenv("SYSTEM_TEST_VAR");
    if (system_env) {
        printf("系统环境变量 SYSTEM_TEST_VAR = '%s'\n", system_env);
    } else {
        printf("提示: 请先设置系统环境变量测试:\n");
        printf("  export SYSTEM_TEST_VAR=\"system_value\"\n");
        printf("  然后重新运行程序查看系统变量优先级\n");
    }
    
    const char *test_var = env_loader_get_string(loader, "SYSTEM_TEST_VAR", "default");
    printf("读取 SYSTEM_TEST_VAR = '%s'\n", test_var);
    
    printf("\n========================================\n");
    printf("  演示完成\n");
    printf("========================================\n");
    
    env_loader_destroy(loader);
    return 0;
}
