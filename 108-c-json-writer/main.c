#include <stdio.h>
#include <stdlib.h>
#include "json_value.h"
#include "json_formatter.h"
#include "json_allocator.h"

int main(void) {
    JsonValue *root = json_create_object();
    if (root == NULL) {
        fprintf(stderr, "Failed to create JSON object\n");
        return 1;
    }
    
    json_object_add(root, "name", json_create_string("张三"));
    json_object_add(root, "age", json_create_int(28));
    json_object_add(root, "is_active", json_create_bool(1));
    json_object_add(root, "score", json_create_float(95.5));
    json_object_add(root, "null_value", json_create_null());
    
    json_object_add(root, "special_chars", 
        json_create_string("包含\"引号\"和\\反斜杠，还有\n换行\t制表符"));
    
    JsonValue *address = json_create_object();
    json_object_add(address, "city", json_create_string("北京"));
    json_object_add(address, "district", json_create_string("海淀区"));
    json_object_add(address, "street", json_create_string("中关村大街1号"));
    json_object_add(address, "zip_code", json_create_int(100080));
    json_object_add(root, "address", address);
    
    JsonValue *tags = json_create_array();
    json_array_add(tags, json_create_string("程序员"));
    json_array_add(tags, json_create_string("C语言"));
    json_array_add(tags, json_create_string("JSON"));
    json_array_add(tags, json_create_int(2024));
    json_object_add(root, "tags", tags);
    
    JsonValue *experience = json_create_array();
    
    JsonValue *job1 = json_create_object();
    json_object_add(job1, "company", json_create_string("科技公司A"));
    json_object_add(job1, "position", json_create_string("初级工程师"));
    json_object_add(job1, "years", json_create_int(2));
    json_object_add(job1, "salary", json_create_float(15000.5));
    json_array_add(experience, job1);
    
    JsonValue *job2 = json_create_object();
    json_object_add(job2, "company", json_create_string("互联网公司B"));
    json_object_add(job2, "position", json_create_string("高级工程师"));
    json_object_add(job2, "years", json_create_int(3));
    json_object_add(job2, "salary", json_create_float(25000.0));
    json_array_add(experience, job2);
    
    json_object_add(root, "experience", experience);
    
    JsonValue *numbers = json_create_array();
    json_array_add(numbers, json_create_int(0));
    json_array_add(numbers, json_create_int(-1));
    json_array_add(numbers, json_create_int(999999999));
    json_array_add(numbers, json_create_float(3.1415926535));
    json_array_add(numbers, json_create_float(3.0));
    json_array_add(numbers, json_create_float(0.0000001));
    json_array_add(numbers, json_create_float(1.0e10));
    json_object_add(root, "test_numbers", numbers);
    
    JsonValue *skills = json_create_object();
    
    JsonValue *languages = json_create_array();
    json_array_add(languages, json_create_string("C"));
    json_array_add(languages, json_create_string("C++"));
    json_array_add(languages, json_create_string("Python"));
    
    JsonValue *tools = json_create_array();
    json_array_add(tools, json_create_string("Git"));
    json_array_add(tools, json_create_string("Docker"));
    json_array_add(tools, json_create_string("Linux"));
    
    json_object_add(skills, "languages", languages);
    json_object_add(skills, "tools", tools);
    json_object_add(skills, "level", json_create_int(5));
    json_object_add(root, "skills", skills);
    
    printf("========== 压缩格式输出 ==========\n");
    char *compact_str = json_to_string(root);
    if (compact_str != NULL) {
        printf("%s\n\n", compact_str);
        free(compact_str);
    } else {
        printf("格式化失败\n\n");
    }
    
    printf("========== 美化格式输出 ==========\n");
    char *pretty_str = json_to_string_pretty(root);
    if (pretty_str != NULL) {
        printf("%s\n", pretty_str);
        free(pretty_str);
    } else {
        printf("格式化失败\n");
    }
    
    json_free(root);
    
    printf("\n========== 内存已成功释放 ==========\n");
    
    return 0;
}
