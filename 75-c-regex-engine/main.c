#include <stdio.h>
#include <stdlib.h>
#include "regex.h"

static void test_match(const char *pattern, const char *text) {
    printf("\n=== Testing pattern: \"%s\" against text: \"%s\" ===\n", pattern, text);
    
    Regex *regex = NULL;
    int compile_result = regex_compile(pattern, &regex);
    
    if (compile_result != 0) {
        if (regex != NULL && regex->error_msg != NULL) {
            printf("  Compile error at position %d: %s\n", regex->error_pos, regex->error_msg);
        } else {
            printf("  Compile error\n");
        }
        regex_free(regex);
        return;
    }
    
    printf("  Compiled successfully. Number of states: %d\n", regex->state_count);
    printf("  Number of capture groups: %d\n", regex->capture_count);
    
    MatchResult result;
    int match_result = regex_match(regex, text, &result);
    
    if (match_result == -1) {
        printf("  Match timeout (exceeded max steps)\n");
    } else if (match_result == 1) {
        printf("  Match found!\n");
        printf("    Start position: %d\n", result.start);
        printf("    End position: %d\n", result.end);
        printf("    Matched text: \"%.*s\"\n", result.end - result.start, text + result.start);
        printf("    Number of capture groups: %d\n", result.count);
        
        for (int i = 0; i < result.count; i++) {
            if (result.captures[i].start != -1) {
                printf("    Group %d: [%d, %d] \"%.*s\"\n", 
                       i, 
                       result.captures[i].start, 
                       result.captures[i].end,
                       result.captures[i].end - result.captures[i].start,
                       text + result.captures[i].start);
            } else {
                printf("    Group %d: (not captured)\n", i);
            }
        }
    } else {
        printf("  No match found\n");
    }
    
    regex_free(regex);
}

static void test_compile_error(const char *pattern) {
    printf("\n=== Testing compile error for pattern: \"%s\" ===\n", pattern);
    
    Regex *regex = NULL;
    int compile_result = regex_compile(pattern, &regex);
    
    if (compile_result != 0) {
        if (regex != NULL && regex->error_msg != NULL) {
            printf("  Compile error at position %d: %s\n", regex->error_pos, regex->error_msg);
        } else {
            printf("  Compile error\n");
        }
    } else {
        printf("  Unexpected: pattern compiled successfully\n");
    }
    
    regex_free(regex);
}

int main(void) {
    printf("========================================\n");
    printf("Regex Engine Tests\n");
    printf("========================================\n");

    printf("\n\n----- Basic character matching -----\n");
    test_match("a", "a");
    test_match("a", "b");
    test_match("ab", "ab");
    test_match("ab", "abc");

    printf("\n\n----- Any character (.) -----\n");
    test_match("a.b", "acb");
    test_match("a.b", "a1b");
    test_match("a.b", "a\nb");
    test_match("...", "abc");

    printf("\n\n----- Zero or more (*) -----\n");
    test_match("a*", "");
    test_match("a*", "a");
    test_match("a*", "aaa");
    test_match("a*b", "b");
    test_match("a*b", "ab");
    test_match("a*b", "aaab");

    printf("\n\n----- One or more (+) -----\n");
    test_match("a+", "a");
    test_match("a+", "aaa");
    test_match("a+", "");
    test_match("a+b", "ab");
    test_match("a+b", "aaab");
    test_match("a+b", "b");

    printf("\n\n----- Zero or one (?) -----\n");
    test_match("a?", "");
    test_match("a?", "a");
    test_match("a?", "aa");
    test_match("colou?r", "color");
    test_match("colou?r", "colour");

    printf("\n\n----- Alternation (|) -----\n");
    test_match("a|b", "a");
    test_match("a|b", "b");
    test_match("a|b", "c");
    test_match("cat|dog", "cat");
    test_match("cat|dog", "dog");
    test_match("cat|dog", "bird");

    printf("\n\n----- Capture groups () -----\n");
    test_match("(a)", "a");
    test_match("(ab)c", "abc");
    test_match("(a)(b)", "ab");
    test_match("(a(b)c)d", "abcd");

    printf("\n\n----- Complex patterns -----\n");
    test_match("a*b+c?", "aaabb");
    test_match("a*b+c?", "aaabbc");
    test_match("a*b+c?", "b");
    test_match("(a|b)*", "abab");
    test_match("(a|b)*c", "ababc");

    printf("\n\n----- Start of line (^) -----\n");
    test_match("^a", "a");
    test_match("^a", "ba");
    test_match("^a", "\na");
    test_match("^ab", "ab");

    printf("\n\n----- End of line ($) -----\n");
    test_match("a$", "a");
    test_match("a$", "ab");
    test_match("a$", "a\n");
    test_match("ab$", "ab");

    printf("\n\n----- Compile error tests -----\n");
    test_compile_error("*");
    test_compile_error("+");
    test_compile_error("?");
    test_compile_error(")");
    test_compile_error("(");
    test_compile_error("a**");
    test_compile_error("|");

    printf("\n\n========================================\n");
    printf("All tests completed\n");
    printf("========================================\n");

    return 0;
}
