#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "template_engine.h"

static void print_separator(void)
{
    printf("========================================\n");
}

static TE_Value *create_sample_data(void)
{
    TE_Value *data = te_value_object();
    
    te_object_set(data, "title", te_value_string("Welcome to Our Newsletter!"));
    te_object_set(data, "company", te_value_string("TechCorp Inc."));
    te_object_set(data, "year", te_value_int(2025));
    te_object_set(data, "show_promo", te_value_bool(1));
    te_object_set(data, "is_vip", te_value_bool(0));
    te_object_set(data, "empty_str", te_value_string(""));
    te_object_set(data, "zero_num", te_value_int(0));
    te_object_set(data, "null_val", te_value_null());
    
    TE_Value *user = te_value_object();
    te_object_set(user, "name", te_value_string("John Doe"));
    te_object_set(user, "email", te_value_string("john@example.com"));
    te_object_set(user, "active", te_value_bool(1));
    
    TE_Value *address = te_value_object();
    te_object_set(address, "city", te_value_string("New York"));
    te_object_set(address, "country", te_value_string("USA"));
    te_object_set(user, "address", address);
    
    te_object_set(data, "user", user);
    
    TE_Value *products = te_value_array();
    TE_Value *p1 = te_value_object();
    te_object_set(p1, "name", te_value_string("Laptop Pro"));
    te_object_set(p1, "price", te_value_double(1299.99));
    te_object_set(p1, "in_stock", te_value_bool(1));
    te_array_push(products, p1);
    
    TE_Value *p2 = te_value_object();
    te_object_set(p2, "name", te_value_string("Wireless Mouse"));
    te_object_set(p2, "price", te_value_double(49.99));
    te_object_set(p2, "in_stock", te_value_bool(0));
    te_array_push(products, p2);
    
    TE_Value *p3 = te_value_object();
    te_object_set(p3, "name", te_value_string("USB-C Hub"));
    te_object_set(p3, "price", te_value_double(79.99));
    te_object_set(p3, "in_stock", te_value_bool(1));
    te_array_push(products, p3);
    
    te_object_set(data, "products", products);
    
    TE_Value *unsafe_html = te_value_string("<script>alert('xss')</script>");
    te_object_set(data, "unsafe_html", unsafe_html);
    
    TE_Value *safe_html = te_value_string("<strong>Important!</strong>");
    te_object_set(data, "safe_html", safe_html);
    
    return data;
}

static void test_basic_variables(TE_Engine *engine, TE_Value *data)
{
    print_separator();
    printf("Test 1: Basic Variable Replacement\n");
    print_separator();
    
    const char *template = 
        "<h1>{{title}}</h1>\n"
        "<p>Hello {{user.name}}!</p>\n"
        "<p>Your email: {{user.email}}</p>\n"
        "<p>Location: {{user.address.city}}, {{user.address.country}}</p>\n"
        "<p>Year: {{year}}</p>\n"
        "<p>Undefined: [{{undefined_var}}]</p>\n";
    
    char *result = te_render(engine, template, data);
    if (result) {
        printf("Result:\n%s\n", result);
        free(result);
    } else if (engine->error.has_error) {
        printf("Error: %s\n", engine->error.message);
    }
}

static void test_conditional_rendering(TE_Engine *engine, TE_Value *data)
{
    print_separator();
    printf("Test 2: Conditional Rendering (if/unless)\n");
    print_separator();
    
    const char *template = 
        "{{#if show_promo}}\n"
        "  <div class=\"promo\">Special Promotion!</div>\n"
        "{{/if}}\n"
        "{{#unless is_vip}}\n"
        "  <p>Become a VIP for exclusive offers!</p>\n"
        "{{/unless}}\n"
        "{{#if user.active}}\n"
        "  <p>Your account is active.</p>\n"
        "{{/if}}\n"
        "Truthy tests:\n"
        "- empty_str is falsy? {{#if empty_str}}NO{{/if}}{{#unless empty_str}}YES{{/unless}}\n"
        "- zero_num is falsy? {{#if zero_num}}NO{{/if}}{{#unless zero_num}}YES{{/unless}}\n"
        "- null_val is falsy? {{#if null_val}}NO{{/if}}{{#unless null_val}}YES{{/unless}}\n";
    
    char *result = te_render(engine, template, data);
    if (result) {
        printf("Result:\n%s\n", result);
        free(result);
    } else if (engine->error.has_error) {
        printf("Error: %s\n", engine->error.message);
    }
}

static void test_loop_rendering(TE_Engine *engine, TE_Value *data)
{
    print_separator();
    printf("Test 3: Loop Rendering (each)\n");
    print_separator();
    
    const char *template = 
        "<h2>Featured Products</h2>\n"
        "<ul>\n"
        "{{#each products}}\n"
        "  <li>\n"
        "    [{{@index}}] {{this.name}} - ${{this.price}}\n"
        "{{#if this.in_stock}}\n"
        "    <span class=\"available\">In Stock</span>\n"
        "{{/if}}\n"
        "{{#unless this.in_stock}}\n"
        "    <span class=\"out\">Out of Stock</span>\n"
        "{{/unless}}\n"
        "  </li>\n"
        "{{/each}}\n"
        "</ul>\n";
    
    char *result = te_render(engine, template, data);
    if (result) {
        printf("Result:\n%s\n", result);
        free(result);
    } else if (engine->error.has_error) {
        printf("Error: %s\n", engine->error.message);
    }
}

static void test_html_escaping(TE_Engine *engine, TE_Value *data)
{
    print_separator();
    printf("Test 4: HTML Escaping\n");
    print_separator();
    
    const char *template = 
        "Escaped (safe): {{unsafe_html}}\n"
        "Raw (unsafe):   {{{unsafe_html}}}\n"
        "Safe HTML:      {{{safe_html}}}\n";
    
    char *result = te_render(engine, template, data);
    if (result) {
        printf("Result:\n%s\n", result);
        free(result);
    } else if (engine->error.has_error) {
        printf("Error: %s\n", engine->error.message);
    }
}

static void test_partials(TE_Engine *engine, TE_Value *data)
{
    print_separator();
    printf("Test 5: Template Partials\n");
    print_separator();
    
    te_engine_add_partial(engine, "header",
        "<header>\n"
        "  <h1>{{title}}</h1>\n"
        "  <p>Welcome, {{user.name}}!</p>\n"
        "</header>\n");
    
    te_engine_add_partial(engine, "footer",
        "<footer>\n"
        "  <p>&copy; {{year}} {{company}}</p>\n"
        "</footer>\n");
    
    const char *template =
        "{{> header}}\n"
        "<main>\n"
        "  <p>This is the main content area.</p>\n"
        "</main>\n"
        "{{> footer}}\n";
    
    char *result = te_render(engine, template, data);
    if (result) {
        printf("Result:\n%s\n", result);
        free(result);
    } else if (engine->error.has_error) {
        printf("Error: %s\n", engine->error.message);
    }
}

static void test_custom_delimiters(TE_Engine *engine, TE_Value *data)
{
    print_separator();
    printf("Test 6: Custom Delimiters\n");
    print_separator();
    
    te_engine_set_delimiters(engine, "[[", "]]");
    
    const char *template =
        "Mixing Vue-like {{syntax}} with our [[title]]\n"
        "User: [[user.name]]\n"
        "Raw: [[[safe_html]]]\n";
    
    char *result = te_render(engine, template, data);
    if (result) {
        printf("Result:\n%s\n", result);
        free(result);
    } else if (engine->error.has_error) {
        printf("Error: %s\n", engine->error.message);
    }
    
    te_engine_set_delimiters(engine, "{{", "}}");
}

static void test_error_handling(TE_Engine *engine)
{
    print_separator();
    printf("Test 7: Error Handling\n");
    print_separator();
    
    printf("Testing unclosed if block...\n");
    const char *bad_template1 = "{{#if show_promo}}Missing close";
    char *result1 = te_render(engine, bad_template1, NULL);
    if (engine->error.has_error) {
        printf("Error caught: %s\n", engine->error.message);
    }
    free(result1);
    
    printf("\nTesting mismatched /if...\n");
    const char *bad_template2 = "Content{{/if}}More";
    char *result2 = te_render(engine, bad_template2, NULL);
    if (engine->error.has_error) {
        printf("Error caught: %s\n", engine->error.message);
    }
    free(result2);
}

static void test_marketing_email(TE_Engine *engine, TE_Value *data)
{
    print_separator();
    printf("Test 8: Complete Marketing Email Example\n");
    print_separator();
    
    const char *email_template =
        "<!DOCTYPE html>\n"
        "<html>\n"
        "<head>\n"
        "  <title>{{title}}</title>\n"
        "  <style>\n"
        "    body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; }\n"
        "    .header { background: #2563eb; color: white; padding: 20px; text-align: center; }\n"
        "    .content { padding: 20px; }\n"
        "    .product { border: 1px solid #e5e7eb; padding: 15px; margin: 10px 0; border-radius: 8px; }\n"
        "    .price { color: #2563eb; font-weight: bold; font-size: 1.2em; }\n"
        "    .in-stock { color: #16a34a; }\n"
        "    .out-stock { color: #dc2626; }\n"
        "    .footer { background: #f3f4f6; padding: 20px; text-align: center; color: #6b7280; }\n"
        "  </style>\n"
        "</head>\n"
        "<body>\n"
        "  <div class=\"header\">\n"
        "    <h1>{{title}}</h1>\n"
        "  </div>\n"
        "  <div class=\"content\">\n"
        "    <p>Hello {{user.name}},</p>\n"
        "    <p>Check out our latest products!</p>\n"
        "{{#if show_promo}}\n"
        "    <div style=\"background: #fef3c7; padding: 15px; border-radius: 8px; margin: 15px 0;\">\n"
        "      <strong>🎉 Limited Time Offer!</strong><br>\n"
        "      Use code SAVE20 for 20% off your order!\n"
        "    </div>\n"
        "{{/if}}\n"
        "{{#each products}}\n"
        "    <div class=\"product\">\n"
        "      <h3>{{this.name}}</h3>\n"
        "      <p class=\"price\">${{this.price}}</p>\n"
        "{{#if this.in_stock}}\n"
        "      <p class=\"in-stock\">✓ In Stock</p>\n"
        "{{/if}}\n"
        "{{#unless this.in_stock}}\n"
        "      <p class=\"out-stock\">✗ Out of Stock</p>\n"
        "{{/unless}}\n"
        "    </div>\n"
        "{{/each}}\n"
        "  </div>\n"
        "  <div class=\"footer\">\n"
        "    <p>&copy; {{year}} {{company}}</p>\n"
        "    <p>You're receiving this email because you subscribed to our newsletter.</p>\n"
        "  </div>\n"
        "</body>\n"
        "</html>\n";
    
    char *result = te_render(engine, email_template, data);
    if (result) {
        printf("Rendered Email HTML:\n");
        printf("%s\n", result);
        free(result);
    } else if (engine->error.has_error) {
        printf("Error: %s\n", engine->error.message);
    }
}

int main(void)
{
    printf("C Template Engine Demo\n");
    printf("=======================\n\n");
    
    TE_Engine *engine = te_engine_create();
    if (!engine) {
        fprintf(stderr, "Failed to create template engine\n");
        return 1;
    }
    
    TE_Value *data = create_sample_data();
    
    test_basic_variables(engine, data);
    test_conditional_rendering(engine, data);
    test_loop_rendering(engine, data);
    test_html_escaping(engine, data);
    test_partials(engine, data);
    test_custom_delimiters(engine, data);
    test_error_handling(engine);
    test_marketing_email(engine, data);
    
    te_value_free(data);
    te_engine_destroy(engine);
    
    printf("\n=======================\n");
    printf("All tests completed!\n");
    
    return 0;
}
