#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "text_replace.h"

#define OUTPUT_BUFFER_SIZE 4096

static void test_case(const char *name, const char *input) {
    char output[OUTPUT_BUFFER_SIZE];
    size_t required_size = 0;
    size_t result_len;
    
    printf("========== %s ==========\n", name);
    printf("原始: %s\n", input);
    
    result_len = desensitize_text(input, strlen(input), output, sizeof(output), &required_size);
    
    printf("脱敏: %s\n", output);
    printf("需要缓冲区大小: %zu, 实际输出长度: %zu\n\n", required_size, result_len);
}

static void test_buffer_overflow_case(void) {
    const char *input = "手机号: 13812345678, 身份证: 110101199001011234";
    size_t required_size = 0;
    size_t result_len;
    
    printf("========== 缓冲区不足测试 ==========\n");
    printf("原始: %s\n\n", input);
    
    printf("测试1: 缓冲区为0的情况 (获取所需大小)\n");
    result_len = desensitize_text(input, strlen(input), NULL, 0, &required_size);
    printf("  需要缓冲区大小: %zu, 实际输出长度: %zu\n\n", required_size, result_len);
    
    size_t small_size = 10;
    char *small_buffer = (char *)malloc(small_size);
    if (small_buffer != NULL) {
        printf("测试2: 缓冲区太小 (%zu bytes)\n", small_size);
        result_len = desensitize_text(input, strlen(input), small_buffer, small_size, &required_size);
        printf("  输出: \"%s\"\n", small_buffer);
        printf("  需要缓冲区大小: %zu, 实际输出长度: %zu\n\n", required_size, result_len);
        free(small_buffer);
    }
    
    printf("测试3: 精确分配缓冲区\n");
    char *exact_buffer = (char *)malloc(required_size + 1);
    if (exact_buffer != NULL) {
        result_len = desensitize_text(input, strlen(input), exact_buffer, required_size + 1, &required_size);
        printf("  脱敏: %s\n", exact_buffer);
        printf("  需要缓冲区大小: %zu, 实际输出长度: %zu\n\n", required_size, result_len);
        free(exact_buffer);
    }
}

int main(void) {
    printf("========================================\n");
    printf("      日志脱敏工具演示\n");
    printf("========================================\n\n");
    
    test_case("手机号测试", 
              "用户手机号: 13812345678, 联系电话: 13987654321");
    
    test_case("身份证测试", 
              "身份证号: 110101199001011234, 老身份证: 110101900101123");
    
    test_case("银行卡测试", 
              "银行卡号: 4111111111111111 (Visa测试卡), 另一张卡: 5555555555554444 (Mastercard测试卡)");
    
    test_case("邮箱测试", 
              "邮箱: testuser@gmail.com, 工作邮箱: admin@company.com.cn");
    
    test_case("超长邮箱测试", 
              "超长邮箱: this_is_a_very_long_email_address_with_many_characters_before_at_sign@this-is-a-very-long-domain-name-with-many-subdomains.example.com.cn");
    
    test_case("连续手机号测试", 
              "连续手机号: 1381234567813987654321");
    
    test_case("混合敏感信息测试", 
              "订单信息: 用户张三(手机号13812345678), "
              "身份证号110101199001011234, "
              "支付卡4111111111111111, "
              "邮箱zhangsan@example.com");
    
    test_case("避免误伤测试", 
              "订单号: 202401011234567890 (这是18位数字但不是身份证,因为出生日期无效), "
              "金额: 22345678901 (11位数字但不是手机号,因为不是1开头), "
              "短号: 12345 (太短不是手机号)");
    
    test_case("边界情况测试", 
              "文本开头手机号13812345678结尾, "
              "文本中间13987654321有手机号, "
              "多个连续1381111222213933334444手机号");
    
    test_buffer_overflow_case();
    
    printf("========================================\n");
    printf("      演示结束\n");
    printf("========================================\n");
    
    return 0;
}
