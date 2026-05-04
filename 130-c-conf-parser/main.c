#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "conf.h"

static void print_separator(void) {
    printf("========================================\n");
}

static void test_basic_types(ConfContext* ctx) {
    print_separator();
    printf("Test 1: Basic Types\n");
    print_separator();
    
    long port = conf_get_int(ctx, "server.port", 8080);
    printf("  server.port (int): %ld\n", port);
    
    double rate = conf_get_float(ctx, "server.rate", 1.0);
    printf("  server.rate (float): %g\n", rate);
    
    bool debug = conf_get_bool(ctx, "debug", false);
    printf("  debug (bool): %s\n", debug ? "true" : "false");
    
    const char* app_name = conf_get_string(ctx, "app.name", "default");
    printf("  app.name (string): %s\n", app_name);
}

static void test_default_values(ConfContext* ctx) {
    print_separator();
    printf("Test 2: Default Values\n");
    print_separator();
    
    long missing_int = conf_get_int(ctx, "non.existent", 999);
    printf("  non.existent (int, default 999): %ld\n", missing_int);
    
    const char* missing_str = conf_get_string(ctx, "missing.key", "my_default");
    printf("  missing.key (string, default 'my_default'): %s\n", missing_str);
    
    bool missing_bool = conf_get_bool(ctx, "no.such.key", true);
    printf("  no.such.key (bool, default true): %s\n", missing_bool ? "true" : "false");
}

static void test_variable_references(ConfContext* ctx) {
    print_separator();
    printf("Test 3: Variable References\n");
    print_separator();
    
    const char* base_dir = conf_get_string(ctx, "paths.base_dir", "");
    printf("  paths.base_dir: %s\n", base_dir);
    
    const char* log_path = conf_get_string(ctx, "paths.log_path", "");
    printf("  paths.log_path (ref base_dir): %s\n", log_path);
    
    const char* config_path = conf_get_string(ctx, "paths.config_path", "");
    printf("  paths.config_path (ref log_path): %s\n", config_path);
    
    const char* welcome = conf_get_string(ctx, "messages.welcome", "");
    printf("  messages.welcome (ref app.name): %s\n", welcome);
}

static void test_dynamic_reference(ConfContext* ctx) {
    print_separator();
    printf("Test 4: Dynamic Reference (modify and re-query)\n");
    print_separator();
    
    const char* before = conf_get_string(ctx, "paths.log_path", "");
    printf("  Before modification:\n");
    printf("    paths.log_path: %s\n", before);
    
    conf_set_string(ctx, "paths", "base_dir", "/new/base");
    printf("  Modified paths.base_dir to '/new/base'\n");
    
    const char* after = conf_get_string(ctx, "paths.log_path", "");
    printf("  After modification:\n");
    printf("    paths.log_path: %s\n", after);
    
    conf_set_string(ctx, "paths", "base_dir", "/opt/myapp");
}

static void test_circular_reference(ConfContext* ctx) {
    print_separator();
    printf("Test 5: Circular Reference Detection\n");
    print_separator();
    
    conf_set_string(ctx, "circular", "a", "${circular.b}");
    conf_set_string(ctx, "circular", "b", "${circular.a}");
    
    char result[256];
    ConfError err = conf_get_string_ex(ctx, "circular", "a", "default", result, sizeof(result));
    
    if (err == CONF_ERR_CIRCULAR_REF) {
        printf("  Circular reference detected as expected!\n");
        printf("  Returned default value: %s\n", result);
    } else {
        printf("  Unexpected result: %s (error: %s)\n", result, conf_error_string(err));
    }
}

static void test_multiline_value(ConfContext* ctx) {
    print_separator();
    printf("Test 6: Multi-line Value (Triple Quotes)\n");
    print_separator();
    
    const char* desc = conf_get_string(ctx, "app.description", "");
    printf("  app.description:\n");
    printf("  ----------------\n");
    printf("%s\n", desc);
    printf("  ----------------\n");
}

static void test_section_priority(ConfContext* ctx) {
    print_separator();
    printf("Test 7: Section Priority\n");
    print_separator();
    
    const char* global = conf_get_string(ctx, "timeout", "");
    printf("  Global 'timeout': %s\n", global);
    
    const char* server_timeout = conf_get_string(ctx, "server.timeout", "");
    printf("  server.timeout (section has priority): %s\n", server_timeout);
    
    const char* same_key = conf_get_string(ctx, "database.timeout", "");
    printf("  database.timeout: %s\n", same_key);
}

static void test_file_watch(ConfContext* ctx) {
    print_separator();
    printf("Test 8: File Watch\n");
    print_separator();
    
    if (conf_needs_reload(ctx)) {
        printf("  File needs reload (modified externally)\n");
        ConfError err = conf_reload(ctx);
        if (err == CONF_OK) {
            printf("  Reloaded successfully\n");
        } else {
            printf("  Reload failed: %s\n", conf_error_string(err));
        }
    } else {
        printf("  File unchanged, no reload needed\n");
    }
}

static void test_type_conversion(ConfContext* ctx) {
    print_separator();
    printf("Test 9: Type Conversion\n");
    print_separator();
    
    long int_from_str = conf_get_int(ctx, "conversion.str_num", 0);
    printf("  conversion.str_num (string '42' as int): %ld\n", int_from_str);
    
    double float_from_int = conf_get_float(ctx, "server.port", 0.0);
    printf("  server.port (int 8080 as float): %g\n", float_from_int);
    
    bool bool_from_int = conf_get_bool(ctx, "server.port", false);
    printf("  server.port (int 8080 as bool): %s\n", bool_from_int ? "true" : "false");
    
    long int_from_bool = conf_get_int(ctx, "debug", 0);
    printf("  debug (bool true as int): %ld\n", int_from_bool);
}

static void test_case_insensitive_keys(ConfContext* ctx) {
    print_separator();
    printf("Test 10: Case Insensitive Keys\n");
    print_separator();
    
    const char* name1 = conf_get_string(ctx, "APP.NAME", "");
    printf("  APP.NAME (uppercase): %s\n", name1);
    
    const char* name2 = conf_get_string(ctx, "App.Name", "");
    printf("  App.Name (mixed case): %s\n", name2);
    
    const char* name3 = conf_get_string(ctx, "app.name", "");
    printf("  app.name (lowercase): %s\n", name3);
}

int main(int argc, char* argv[]) {
    const char* config_file = "config.ini";
    if (argc > 1) {
        config_file = argv[1];
    }
    
    printf("C Configuration Parser Demo\n");
    printf("============================\n\n");
    
    ConfContext* ctx = conf_create();
    if (!ctx) {
        fprintf(stderr, "Failed to create config context\n");
        return 1;
    }
    
    printf("Loading config file: %s\n", config_file);
    ConfError err = conf_load_file(ctx, config_file);
    if (err != CONF_OK) {
        fprintf(stderr, "Failed to load config: %s\n", conf_last_error(ctx));
        conf_destroy(ctx);
        return 1;
    }
    printf("Config loaded successfully.\n\n");
    
    test_basic_types(ctx);
    test_default_values(ctx);
    test_variable_references(ctx);
    test_dynamic_reference(ctx);
    test_circular_reference(ctx);
    test_multiline_value(ctx);
    test_section_priority(ctx);
    test_type_conversion(ctx);
    test_case_insensitive_keys(ctx);
    test_file_watch(ctx);
    
    print_separator();
    printf("Demo Complete\n");
    print_separator();
    
    conf_destroy(ctx);
    return 0;
}
