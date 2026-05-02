#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <math.h>
#include "json.h"

static int sax_events = 0;

static void test_sax_object_start(json_parser_t *p, void *ud, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] object_start at %d:%d\n", pos.line, pos.column); }
static void test_sax_object_end(json_parser_t *p, void *ud, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] object_end at %d:%d\n", pos.line, pos.column); }
static void test_sax_array_start(json_parser_t *p, void *ud, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] array_start at %d:%d\n", pos.line, pos.column); }
static void test_sax_array_end(json_parser_t *p, void *ud, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] array_end at %d:%d\n", pos.line, pos.column); }
static void test_sax_key(json_parser_t *p, void *ud, const char *key, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] key: \"%s\" at %d:%d\n", key, pos.line, pos.column); }
static void test_sax_string(json_parser_t *p, void *ud, const char *val, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] string: \"%s\" at %d:%d\n", val, pos.line, pos.column); }
static void test_sax_int(json_parser_t *p, void *ud, long long val, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] int: %lld at %d:%d\n", val, pos.line, pos.column); }
static void test_sax_float(json_parser_t *p, void *ud, double val, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] float: %g at %d:%d\n", val, pos.line, pos.column); }
static void test_sax_bool(json_parser_t *p, void *ud, bool val, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] bool: %s at %d:%d\n", val?"true":"false", pos.line, pos.column); }
static void test_sax_null(json_parser_t *p, void *ud, json_pos_t pos) { (void)p; (void)ud; sax_events++; printf("[SAX] null at %d:%d\n", pos.line, pos.column); }

static void test_sax(void)
{
    printf("\n========== TEST SAX PARSER ==========\n");
    const char *json_str = 
        "{ \"name\": \"test\", \"id\": 42, \"pi\": 3.14159, \"valid\": true, \"nothing\": null, \"arr\": [1, 2, 3] }";
    
    FILE *f = fmemopen((void*)json_str, strlen(json_str), "r");
    assert(f);
    
    json_sax_callbacks_t cb = {0};
    cb.object_start = test_sax_object_start;
    cb.object_end = test_sax_object_end;
    cb.array_start = test_sax_array_start;
    cb.array_end = test_sax_array_end;
    cb.key = test_sax_key;
    cb.string = test_sax_string;
    cb.number_int = test_sax_int;
    cb.number_float = test_sax_float;
    cb.boolean = test_sax_bool;
    cb.null = test_sax_null;
    
    sax_events = 0;
    json_parser_t parser;
    json_parser_init(&parser, f, &cb, NULL);
    
    json_error_t err = json_parse_sax(&parser);
    fclose(f);
    
    assert(err == JSON_OK);
    assert(sax_events > 0);
    printf("SAX test passed, events: %d\n", sax_events);
}

static void test_dom_basic(void)
{
    printf("\n========== TEST DOM BASIC ==========\n");
    const char *json_str = 
        "{\n"
        "  \"name\": \"Alice\",\n"
        "  \"age\": 30,\n"
        "  \"is_student\": false,\n"
        "  \"scores\": [95.5, 88, 92],\n"
        "  \"address\": {\n"
        "    \"city\": \"Beijing\",\n"
        "    \"zip\": 100000\n"
        "  }\n"
        "}";
    
    FILE *f = fmemopen((void*)json_str, strlen(json_str), "r");
    assert(f);
    
    json_error_t err;
    json_value_t *root = json_parse_dom(f, &err);
    fclose(f);
    
    assert(err == JSON_OK);
    assert(root);
    assert(json_type(root) == JSON_OBJECT);
    
    json_value_t *name = json_get(root, "name");
    assert(name && json_type(name) == JSON_STRING);
    const char *name_str;
    assert(json_get_string(name, &name_str));
    printf("name = %s\n", name_str);
    assert(strcmp(name_str, "Alice") == 0);
    
    json_value_t *age = json_get(root, "age");
    assert(age && json_type(age) == JSON_INT);
    long long age_val;
    assert(json_get_int(age, &age_val));
    printf("age = %lld\n", age_val);
    assert(age_val == 30LL);
    
    json_value_t *scores = json_get(root, "scores");
    assert(scores && json_type(scores) == JSON_ARRAY);
    size_t score_count = json_array_length(scores);
    printf("scores count = %zu\n", score_count);
    assert(score_count == 3);
    
    json_value_t *score0 = json_get_at(scores, 0);
    double d;
    assert(json_get_float(score0, &d));
    printf("scores[0] = %g\n", d);
    assert(fabs(d - 95.5) < 0.001);
    
    json_value_t *addr = json_get(root, "address");
    assert(addr && json_type(addr) == JSON_OBJECT);
    json_value_t *city = json_get(addr, "city");
    assert(json_get_string(city, &name_str));
    printf("address.city = %s\n", name_str);
    assert(strcmp(name_str, "Beijing") == 0);
    
    json_free(root);
    printf("DOM basic test passed\n");
}

static void test_number_types(void)
{
    printf("\n========== TEST NUMBER TYPES (INT/FLOAT) ==========\n");
    
    struct { const char *json; json_type_t expected; } tests[] = {
        {"42", JSON_INT},
        {"-123", JSON_INT},
        {"9223372036854775807", JSON_INT},
        {"3.14", JSON_FLOAT},
        {"1e10", JSON_FLOAT},
        {"-0.5", JSON_FLOAT},
        {"2.5e-3", JSON_FLOAT},
        {"0", JSON_INT},
        {NULL, JSON_NULL}
    };
    
    for (int i = 0; tests[i].json; i++) {
        FILE *f = fmemopen((void*)tests[i].json, strlen(tests[i].json), "r");
        json_error_t err;
        json_value_t *v = json_parse_dom(f, &err);
        fclose(f);
        
        assert(err == JSON_OK);
        printf("JSON: %s => type %s\n", tests[i].json, 
               json_type(v) == JSON_INT ? "INT" : 
               json_type(v) == JSON_FLOAT ? "FLOAT" : "OTHER");
        assert(json_type(v) == tests[i].expected);
        
        if (strcmp(tests[i].json, "9223372036854775807") == 0) {
            long long ll;
            assert(json_get_int(v, &ll));
            printf("  value = %lld (max long long)\n", ll);
            assert(ll == 9223372036854775807LL);
        }
        
        json_free(v);
    }
    printf("Number type distinction test passed\n");
}

static void test_string_escape(void)
{
    printf("\n========== TEST STRING ESCAPE & UNICODE ==========\n");
    
    struct { const char *json_in; const char *expected; } tests[] = {
        {"\"hello\"", "hello"},
        {"\"line1\\nline2\"", "line1\nline2"},
        {"\"tab\\there\"", "tab\there"},
        {"\"quote:\\\"test\\\"\"", "quote:\"test\""},
        {"\"backslash:\\\\test\"", "backslash:\\test"},
        {"\"unicode: \\u4E2D\\u6587\"", "unicode: 中文"},
        {NULL, NULL}
    };
    
    for (int i = 0; tests[i].json_in; i++) {
        FILE *f = fmemopen((void*)tests[i].json_in, strlen(tests[i].json_in), "r");
        json_error_t err;
        json_value_t *v = json_parse_dom(f, &err);
        fclose(f);
        
        if (err != JSON_OK) {
            printf("FAIL: json=%s error=%d\n", tests[i].json_in, err);
            assert(0);
        }
        
        const char *s;
        assert(json_get_string(v, &s));
        printf("input: %s\n", tests[i].json_in);
        printf("output: \"%s\"\n", s);
        printf("expected: \"%s\"\n\n", tests[i].expected);
        assert(strcmp(s, tests[i].expected) == 0);
        
        json_free(v);
    }
    printf("String escape test passed\n");
}

static void test_error_handling(void)
{
    printf("\n========== TEST ERROR HANDLING ==========\n");
    
    struct { const char *json; const char *desc; } tests[] = {
        {"{ \"a\": 1 } extra", "trailing content"},
        {"[1, 2, 3", "unclosed array"},
        {"\"unclosed", "unclosed string"},
        {"0123", "leading zero"},
        {"-+1", "invalid number"},
        {"nul", "truncated null"},
        {"[true, fals, true]", "truncated false"},
        {"invalid", "invalid token"},
        {NULL, NULL}
    };
    
    for (int i = 0; tests[i].json; i++) {
        FILE *f = fmemopen((void*)tests[i].json, strlen(tests[i].json), "r");
        json_parser_t parser;
        json_sax_callbacks_t empty_cb = {0};
        json_parser_init(&parser, f, &empty_cb, NULL);
        
        json_error_t err = json_parse_sax(&parser);
        fclose(f);
        
        json_pos_t pos = json_error_position(&parser);
        printf("Test: %s\n", tests[i].desc);
        printf("  Input: \"%s\"\n", tests[i].json);
        printf("  Error code: %d (non-zero = error)\n", err);
        printf("  Error pos: %d:%d\n", pos.line, pos.column);
        printf("  Error msg: %s\n\n", json_error_message(&parser));
        
        assert(err != JSON_OK);
        assert(pos.line >= 1);
        assert(pos.column >= 1);
    }
    printf("Error handling test passed\n");
}

static void test_empty_containers(void)
{
    printf("\n========== TEST EMPTY CONTAINERS ==========\n");
    
    const char *tests[] = {
        "{}",
        "[]",
        "{ \"empty_obj\": {}, \"empty_arr\": [] }",
        NULL
    };
    
    for (int i = 0; tests[i]; i++) {
        FILE *f = fmemopen((void*)tests[i], strlen(tests[i]), "r");
        json_error_t err;
        json_value_t *v = json_parse_dom(f, &err);
        fclose(f);
        
        printf("JSON: %s => err=%d\n", tests[i], err);
        assert(err == JSON_OK);
        assert(v);
        json_free(v);
    }
    printf("Empty containers test passed\n");
}

int main(void)
{
    printf("========================================\n");
    printf("    JSON PARSER TEST SUITE\n");
    printf("========================================\n");
    
    test_sax();
    test_dom_basic();
    test_number_types();
    test_string_escape();
    test_error_handling();
    test_empty_containers();
    
    printf("\n========================================\n");
    printf("    ALL TESTS PASSED!\n");
    printf("========================================\n");
    
    return 0;
}
