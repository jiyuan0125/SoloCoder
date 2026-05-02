#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "ini_parser.h"

int main(int argc, char* argv[]) {
    printf("=== INI Parser Test ===\n\n");
    
    setenv("TEST_ENV_VAR", "hello_world", 1);
    
    ini_parser_t* parser = ini_parse("test_config.ini");
    if (!parser) {
        fprintf(stderr, "Parse failed: line %d: %s\n", 
                ini_get_error_line(parser), ini_get_error_message(parser));
        return 1;
    }
    
    printf("1. Testing basic section and key-value parsing...\n");
    
    const char* server_name = ini_get(parser, "server", "name", "default");
    printf("  server.name = %s (expected: myserver)\n", server_name);
    
    int port = ini_get_int(parser, "server", "port", 0);
    printf("  server.port = %d (expected: 8080)\n", port);
    
    printf("\n2. Testing environment variable expansion...\n");
    const char* env_val = ini_get(parser, "test", "env_var", "");
    printf("  test.env_var = %s (expected: prefix_hello_world_suffix)\n", env_val);
    
    const char* nonexistent_env = ini_get(parser, "test", "nonexistent_env", "");
    printf("  test.nonexistent_env = %s (expected: ${NONEXISTENT_VAR})\n", nonexistent_env);
    
    printf("\n3. Testing type inference (int)...\n");
    int positive = ini_get_int(parser, "types", "positive_int", 0);
    printf("  types.positive_int = %d (expected: 123)\n", positive);
    
    int negative = ini_get_int(parser, "types", "negative_int", 0);
    printf("  types.negative_int = %d (expected: -456)\n", negative);
    
    int spaced = ini_get_int(parser, "types", "spaced_int", 0);
    printf("  types.spaced_int = %d (expected: 789)\n", spaced);
    
    int invalid_int = ini_get_int(parser, "types", "invalid_int", 999);
    printf("  types.invalid_int = %d (expected: 999 [default])\n", invalid_int);
    
    printf("\n4. Testing type inference (bool)...\n");
    int bool_true = ini_get_bool(parser, "types", "bool_true", -1);
    printf("  types.bool_true = %d (expected: 1)\n", bool_true);
    
    int bool_false = ini_get_bool(parser, "types", "bool_false", -1);
    printf("  types.bool_false = %d (expected: 0)\n", bool_false);
    
    int bool_yes = ini_get_bool(parser, "types", "bool_yes", -1);
    printf("  types.bool_yes = %d (expected: 1)\n", bool_yes);
    
    int bool_no = ini_get_bool(parser, "types", "bool_no", -1);
    printf("  types.bool_no = %d (expected: 0)\n", bool_no);
    
    int bool_one = ini_get_bool(parser, "types", "bool_one", -1);
    printf("  types.bool_one = %d (expected: 1)\n", bool_one);
    
    int bool_zero = ini_get_bool(parser, "types", "bool_zero", -1);
    printf("  types.bool_zero = %d (expected: 0)\n", bool_zero);
    
    int invalid_bool = ini_get_bool(parser, "types", "invalid_bool", -1);
    printf("  types.invalid_bool = %d (expected: -1 [default])\n", invalid_bool);
    
    printf("\n5. Testing case insensitivity...\n");
    const char* case_test = ini_get(parser, "SERVER", "NAME", "not_found");
    printf("  SERVER.NAME = %s (expected: myserver)\n", case_test);
    
    const char* case_test2 = ini_get(parser, "Server", "Port", 0);
    int port2 = ini_get_int(parser, "Server", "Port", 0);
    printf("  Server.Port = %d (expected: 8080)\n", port2);
    
    printf("\n6. Testing include directive...\n");
    const char* included_val = ini_get(parser, "included_section", "included_key", "not_found");
    printf("  included_section.included_key = %s (expected: included_value)\n", included_val);
    
    printf("\n7. Testing ini_set and ini_save...\n");
    
    printf("  Setting new values...\n");
    ini_set(parser, "new_section", "new_key", "new_value");
    ini_set(parser, "server", "port", "9090");
    
    if (ini_save(parser, "test_config_saved.ini") == 0) {
        printf("  Saved to test_config_saved.ini successfully\n");
    } else {
        printf("  Failed to save\n");
    }
    
    printf("\n8. Verifying saved file...\n");
    ini_parser_t* parser2 = ini_parse("test_config_saved.ini");
    if (parser2) {
        const char* new_val = ini_get(parser2, "new_section", "new_key", "not_found");
        printf("  new_section.new_key = %s (expected: new_value)\n", new_val);
        
        int new_port = ini_get_int(parser2, "server", "port", 0);
        printf("  server.port = %d (expected: 9090)\n", new_port);
        
        ini_free(parser2);
    } else {
        printf("  Failed to re-parse saved file\n");
    }
    
    printf("\n9. Testing value trim behavior...\n");
    const char* trimmed = ini_get(parser, "test", "trim_me", "");
    printf("  test.trim_me = '%s' (expected: 'trimmed_value')\n", trimmed);
    
    printf("\n10. Testing semicolon and hash in values...\n");
    const char* with_semi = ini_get(parser, "test", "with_semicolon", "");
    printf("  test.with_semicolon = '%s' (expected: 'value;with;semicolon')\n", with_semi);
    
    const char* with_hash = ini_get(parser, "test", "with_hash", "");
    printf("  test.with_hash = '%s' (expected: 'value#with#hash')\n", with_hash);
    
    ini_free(parser);
    
    printf("\n=== All tests completed ===\n");
    
    return 0;
}
