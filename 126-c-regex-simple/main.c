#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "regex.h"

static void test_match(const char *pattern, const char *input, int flags, const char *description) {
    RegexError err = REGEX_OK;
    int err_pos = 0;
    Regex *re = regex_compile(pattern, flags, &err, &err_pos);
    
    printf("\n================================================\n");
    printf("测试: %s\n", description);
    printf("模式: \"%s\"\n", pattern);
    printf("输入: \"%s\"\n", input);
    printf("------------------------------------------------\n");
    
    if (re == NULL) {
        printf("编译失败: %s (位置: %d)\n", regex_strerror(err), err_pos);
        return;
    }
    
    RegexMatch matches[5];
    int result = regex_match(re, input, matches, 5);
    
    printf("全文匹配结果: %s\n", result ? "匹配成功" : "匹配失败");
    if (result) {
        printf("  匹配位置: start=%d, length=%d\n", matches[0].start, matches[0].length);
        char matched[256];
        strncpy(matched, input + matches[0].start, matches[0].length);
        matched[matches[0].length] = '\0';
        printf("  匹配内容: \"%s\"\n", matched);
        
        for (int i = 1; i <= re->group_count && i < 5; i++) {
            printf("  分组%d: start=%d, length=%d", i, matches[i].start, matches[i].length);
            if (matches[i].length > 0) {
                char group[256];
                strncpy(group, input + matches[i].start, matches[i].length);
                group[matches[i].length] = '\0';
                printf(" => \"%s\"", group);
            }
            printf("\n");
        }
    }
    
    int search_result = regex_search(re, input, matches, 5);
    printf("搜索匹配结果: %s\n", search_result ? "找到匹配" : "未找到匹配");
    if (search_result && !result) {
        printf("  匹配位置: start=%d, length=%d\n", matches[0].start, matches[0].length);
        char matched[256];
        strncpy(matched, input + matches[0].start, matches[0].length);
        matched[matches[0].length] = '\0';
        printf("  匹配内容: \"%s\"\n", matched);
    }
    
    regex_free(re);
}

static void test_compile_error(const char *pattern, const char *description) {
    RegexError err = REGEX_OK;
    int err_pos = 0;
    
    printf("\n================================================\n");
    printf("错误测试: %s\n", description);
    printf("模式: \"%s\"\n", pattern);
    printf("------------------------------------------------\n");
    
    Regex *re = regex_compile(pattern, 0, &err, &err_pos);
    
    if (re == NULL) {
        printf("编译失败 (预期): %s\n", regex_strerror(err));
        printf("错误位置: 第 %d 个字符\n", err_pos);
        printf("模式: %s\n", pattern);
        printf("      ");
        for (int i = 0; i < err_pos; i++) printf(" ");
        printf("^\n");
    } else {
        printf("意外: 编译成功了\n");
        regex_free(re);
    }
}

int main() {
    printf("================================================");
    printf("\n简易模式匹配引擎演示 - 配置文件格式校验工具\n");
    printf("================================================");
    
    printf("\n\n========== 1. 基础匹配测试 ==========");
    
    test_match("hello", "hello", 0, "普通字符精确匹配");
    test_match("hello", "Hello", 0, "区分大小写匹配 - 不匹配");
    test_match("hello", "Hello", REGEX_CASE_INSENSITIVE, "不区分大小写匹配");
    test_match("h.llo", "hello", 0, ". 匹配任意单个字符");
    test_match("h.llo", "h1llo", 0, ". 匹配数字");
    test_match("h.llo", "h llo", 0, ". 匹配空格");
    
    printf("\n\n========== 2. 量词匹配测试 (* + ?) ==========");
    
    test_match("a*", "aaabbb", 0, "* 匹配零次或多次");
    test_match("a+", "aaabbb", 0, "+ 匹配一次或多次");
    test_match("a?", "aaabbb", 0, "? 匹配零次或一次");
    test_match("colou?r", "color", 0, "? 用于可选字符 (美国英语)");
    test_match("colou?r", "colour", 0, "? 用于可选字符 (英国英语)");
    test_match(".*test", "this is a test", 0, ".* 贪婪匹配示例");
    
    printf("\n\n========== 3. 字符集测试 ([abc] [a-z] [^abc]) ==========");
    
    test_match("[abc]", "a", 0, "[abc] 匹配集合中的任意字符");
    test_match("[abc]", "x", 0, "[abc] 不匹配集合外字符");
    test_match("[^abc]", "x", 0, "[^abc] 匹配不在集合中的字符");
    test_match("[^abc]", "a", 0, "[^abc] 不匹配集合内字符");
    test_match("[0-9]", "5", 0, "[0-9] 匹配数字范围");
    test_match("[a-z]", "m", 0, "[a-z] 匹配小写字母范围");
    test_match("[A-Z]", "M", 0, "[A-Z] 匹配大写字母范围");
    test_match("[a-zA-Z0-9]", "A", 0, "[a-zA-Z0-9] 匹配字母数字");
    
    printf("\n\n========== 4. 锚点测试 (^ $) ==========");
    
    test_match("^hello", "hello world", 0, "^ 匹配行首 - 成功");
    test_match("^hello", "say hello", 0, "^ 匹配行首 - 失败");
    test_match("world$", "hello world", 0, "$ 匹配行尾 - 成功");
    test_match("world$", "world hello", 0, "$ 匹配行尾 - 失败");
    test_match("^hello$", "hello", 0, "^...$ 精确匹配整个字符串");
    test_match("^hello$", "hello1", 0, "^...$ 不匹配多余字符");
    
    printf("\n\n========== 5. 模拟配置文件场景 ==========");
    
    test_match("^[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*$", 
               "192.168.1.1", 0, "模拟IP地址格式校验 (简化版)");
    test_match("^[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*$", 
               "192.168.1", 0, "无效IP地址 - 缺少段");
    test_match("^[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*$", 
               "256.168.1.1", 0, "IP格式匹配(仅检查格式，不检查范围)");
    
    test_match("^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z][a-zA-Z]*$",
               "test@example.com", 0, "模拟邮箱格式校验");
    test_match("^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z][a-zA-Z]*$",
               "invalid-email", 0, "无效邮箱格式");
    
    printf("\n\n========== 6. 分组提取测试 ==========");
    
    test_match("value=([0-9]+)", "value=12345", 0, "提取配置值中的数字");
    test_match("^name: ([a-zA-Z]+)$", "name: John", 0, "提取配置项的值");
    test_match("(.*)=(.*)", "key=value", 0, "简单键值对分组");
    
    printf("\n\n========== 7. 全文匹配 vs 搜索匹配 ==========");
    
    test_match("test", "this is a test string", 0, "全文匹配 - 整个字符串必须匹配");
    test_match(".*test.*", "this is a test string", 0, "用.*实现搜索效果");
    
    printf("\n\n========== 8. 语法错误测试 ==========");
    
    test_compile_error("[abc", "不匹配的左方括号");
    test_compile_error("*abc", "量词前没有字符");
    test_compile_error("+abc", "量词前没有字符");
    test_compile_error("?abc", "量词前没有字符");
    test_compile_error("a**", "连续量词");
    test_compile_error("(abc", "不匹配的左括号");
    test_compile_error("abc)", "不匹配的右括号");
    test_compile_error("[z-a]", "无效的字符范围");
    test_compile_error("abc\\", "不完整的转义序列");
    
    printf("\n\n================================================\n");
    printf("演示完成！\n");
    printf("================================================\n");
    
    return 0;
}
