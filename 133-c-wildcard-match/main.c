#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "wildcard.h"

typedef struct {
    const char *pattern;
    const char *description;
} test_rule_t;

typedef struct {
    const char *pattern;
    const char *input;
    bool should_match;
} single_test_t;

static void print_separator(void) {
    printf("========================================\n");
}

static void test_single_pattern(const char *pattern, const char *input, 
                                wc_case_mode_t case_mode, bool expected) {
    wc_pattern_t *p = wc_pattern_create(pattern, case_mode);
    if (!p) {
        printf("  [ERROR] Failed to parse pattern '%s': %s\n", 
               pattern, wc_pattern_get_error());
        return;
    }
    
    wc_match_result_t result;
    bool matched = wc_match(p, input, case_mode, &result);
    
    const char *status = (matched == expected) ? "PASS" : "FAIL";
    printf("  [%s] Pattern: '%s', Input: '%s'\n", 
           status, pattern, input);
    printf("         Expected: %s, Got: %s\n", 
           expected ? "match" : "no match",
           matched ? "match" : "no match");
    
    if (matched && result.count > 0) {
        printf("         Captures (%zu):\n", result.count);
        for (size_t i = 0; i < result.count; i++) {
            const char *text = wc_capture_get_text(&result, i);
            if (text) {
                printf("           [%zu]: '%s'\n", i, text);
                if (strlen(text) > 0) {
                    free((void*)text);
                }
            }
        }
    }
    
    wc_pattern_destroy(p);
}

static void run_single_tests(void) {
    printf("\n");
    print_separator();
    printf("单模式匹配测试\n");
    print_separator();
    
    single_test_t tests[] = {
        {"test.txt", "test.txt", true},
        {"test.txt", "test.exe", false},
        
        {"*.txt", "test.txt", true},
        {"*.txt", "a.txt", true},
        {"*.txt", ".txt", true},
        {"*.txt", "test.exe", false},
        
        {"test.*", "test.txt", true},
        {"test.*", "test.exe", true},
        {"test.*", "test.", true},
        
        {"src/*.c", "src/main.c", true},
        {"src/*.c", "src/utils.c", true},
        {"src/*.c", "src/include/foo/bar.c", false},
        
        {"src/**/*.c", "src/main.c", true},
        {"src/**/*.c", "src/lib/foo.c", true},
        {"src/**/*.c", "src/a/b/c/test.c", true},
        
        {"file.???", "file.abc", true},
        {"file.???", "file.abcd", false},
        {"file.???", "file.ab", false},
        
        {"test[abc]", "testa", true},
        {"test[abc]", "testb", true},
        {"test[abc]", "testx", false},
        
        {"test[a-z]", "testm", true},
        {"test[a-z]", "testA", false},
        {"test[a-z]", "test5", false},
        
        {"test[0-9]", "test5", true},
        {"test[0-9]", "testa", false},
        
        {"test[!abc]", "testx", true},
        {"test[!abc]", "testa", false},
        
        {"test[^abc]", "testx", true},
        {"test[^abc]", "testa", false},
        
        {"a*?", "ab", true},
        {"a*?", "abc", true},
        {"a*?", "abcd", true},
        {"a*?", "a", false},
        
        {"**/*.txt", "test.txt", true},
        {"**/*.txt", "a/test.txt", true},
        {"**/*.txt", "a/b/c/test.txt", true},
        
        {"a**z", "az", true},
        {"a**z", "abz", true},
        {"a**z", "a/b/z", true},
        
        {"test\\*file", "test*file", true},
        {"test\\*file", "testabcfile", false},
        
        {"test\\?file", "test?file", true},
        {"test\\?file", "testafile", false},
        
        {NULL, NULL, false}
    };
    
    for (int i = 0; tests[i].input != NULL; i++) {
        test_single_pattern(tests[i].pattern, tests[i].input, 
                            WC_CASE_SENSITIVE, tests[i].should_match);
    }
    
    printf("\n--- 不区分大小写测试 ---\n");
    test_single_pattern("*.EXE", "test.exe", WC_CASE_INSENSITIVE, true);
    test_single_pattern("TEST.txt", "test.TXT", WC_CASE_INSENSITIVE, true);
    test_single_pattern("[A-Z]*", "test.txt", WC_CASE_INSENSITIVE, true);
}

static void run_batch_matcher_tests(void) {
    printf("\n");
    print_separator();
    printf("批量匹配器测试\n");
    print_separator();
    
    test_rule_t rules[] = {
        {"*.example.com", "所有 example.com 子域名"},
        {"www.*.com", "www 开头的二级域名"},
        {"**/static/*.css", "static 目录下的 CSS 文件"},
        {"src/**/*.c", "src 目录下的所有 C 文件"},
        {"*.txt", "所有 txt 文件"},
        {"test[0-9]*", "test 开头后跟数字"},
        {"**/tmp/**", "tmp 目录下的任何内容"},
        {NULL, NULL}
    };
    
    wc_batch_matcher_t *matcher = wc_batch_matcher_create(WC_CASE_SENSITIVE);
    if (!matcher) {
        printf("  [ERROR] Failed to create batch matcher\n");
        return;
    }
    
    printf("\n添加规则:\n");
    for (int i = 0; rules[i].pattern != NULL; i++) {
        int idx = wc_batch_matcher_add_rule(matcher, rules[i].pattern, 
                                              (void*)rules[i].description);
        if (idx < 0) {
            printf("  [ERROR] Failed to add rule '%s': %s\n", 
                   rules[i].pattern, wc_pattern_get_error());
        } else {
            printf("  Rule %d: '%s' - %s\n", 
                   idx, rules[i].pattern, rules[i].description);
        }
    }
    
    wc_batch_matcher_build(matcher);
    
    const char *test_inputs[] = {
        "sub.example.com",
        "www.google.com",
        "static/style.css",
        "assets/static/main.css",
        "src/main.c",
        "src/lib/utils.c",
        "a/b/c/src/test.c",
        "document.txt",
        "test123_file",
        "testabc_file",
        "/var/tmp/logs/error.log",
        "home/user/tmp/file.txt",
        "notmatching.xyz",
        NULL
    };
    
    printf("\n批量匹配测试:\n");
    for (int i = 0; test_inputs[i] != NULL; i++) {
        const char *input = test_inputs[i];
        size_t match_count = 0;
        int *matches = wc_batch_match(matcher, input, &match_count);
        
        printf("\n  Input: '%s'\n", input);
        if (matches && match_count > 0) {
            printf("    匹配 %zu 条规则:\n", match_count);
            for (size_t j = 0; j < match_count; j++) {
                int rule_idx = matches[j];
                wc_rule_t *rule = &matcher->rules[rule_idx];
                printf("      [%d] '%s' - %s\n", 
                       rule_idx, 
                       rule->pattern.original_pattern,
                       (char*)rule->user_data);
                
                wc_match_result_t result;
                if (wc_match_rule(matcher, input, rule_idx, &result)) {
                    if (result.count > 0) {
                        printf("          捕获: ");
                        for (size_t k = 0; k < result.count; k++) {
                            const char *text = wc_capture_get_text(&result, k);
                            if (text) {
                                printf("'%s' ", text);
                                if (strlen(text) > 0) free((void*)text);
                            }
                        }
                        printf("\n");
                    }
                }
            }
        } else {
            printf("    无匹配\n");
        }
        
        if (matches) {
            wc_batch_match_free(matches);
        }
    }
    
    wc_batch_matcher_destroy(matcher);
}

static void run_special_cases_tests(void) {
    printf("\n");
    print_separator();
    printf("特殊情况测试\n");
    print_separator();
    
    printf("\n--- 连续星号测试 ---\n");
    test_single_pattern("a***b", "axyzb", WC_CASE_SENSITIVE, true);
    test_single_pattern("a*****b", "ab", WC_CASE_SENSITIVE, true);
    test_single_pattern("****", "anything", WC_CASE_SENSITIVE, true);
    
    printf("\n--- 空字符串测试 ---\n");
    test_single_pattern("", "", WC_CASE_SENSITIVE, true);
    test_single_pattern("*", "", WC_CASE_SENSITIVE, true);
    test_single_pattern("**", "", WC_CASE_SENSITIVE, true);
    test_single_pattern("a*", "a", WC_CASE_SENSITIVE, true);
    test_single_pattern("*a", "a", WC_CASE_SENSITIVE, true);
    
    printf("\n--- *? 和 ?* 测试 ---\n");
    test_single_pattern("*?", "a", WC_CASE_SENSITIVE, true);
    test_single_pattern("*?", "ab", WC_CASE_SENSITIVE, true);
    test_single_pattern("*?", "", WC_CASE_SENSITIVE, false);
    test_single_pattern("?*", "a", WC_CASE_SENSITIVE, true);
    test_single_pattern("?*", "ab", WC_CASE_SENSITIVE, true);
    test_single_pattern("?*", "", WC_CASE_SENSITIVE, false);
    
    printf("\n--- 复杂模式组合测试 ---\n");
    test_single_pattern("a*b?c", "axyzbxc", WC_CASE_SENSITIVE, true);
    test_single_pattern("a*b?c", "abbc", WC_CASE_SENSITIVE, true);
    test_single_pattern("a*b?c", "abxc", WC_CASE_SENSITIVE, true);
    
    printf("\n--- 目录分隔符测试 ---\n");
    test_single_pattern("a*b", "a/b", WC_CASE_SENSITIVE, false);
    test_single_pattern("a**b", "a/b", WC_CASE_SENSITIVE, true);
    test_single_pattern("a**b", "a/x/y/b", WC_CASE_SENSITIVE, true);
}

int main(void) {
    printf("========================================\n");
    printf("通配符模式匹配引擎 - 防火墙规则匹配\n");
    printf("========================================\n");
    
    run_single_tests();
    run_batch_matcher_tests();
    run_special_cases_tests();
    
    printf("\n");
    print_separator();
    printf("所有测试完成\n");
    print_separator();
    
    return 0;
}
