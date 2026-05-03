#include <stdio.h>
#include <stdlib.h>
#include <float.h>
#include <math.h>
#include "json_value.h"
#include "json_formatter.h"
#include "json_allocator.h"

static int add_to_object(JsonValue *obj, const char *key, JsonValue *value) {
    if (value == NULL) {
        return -1;
    }
    if (json_object_add(obj, key, value) != 0) {
        json_free(value);
        return -1;
    }
    return 0;
}

static int add_to_array(JsonValue *arr, JsonValue *value) {
    if (value == NULL) {
        return -1;
    }
    if (json_array_add(arr, value) != 0) {
        json_free(value);
        return -1;
    }
    return 0;
}

int main(void) {
    int ret = 0;
    JsonValue *root = json_create_object();
    if (root == NULL) {
        fprintf(stderr, "Failed to create root JSON object\n");
        return 1;
    }
    
    printf("========== 构建复杂 JSON 结构 ==========\n");
    
    if (add_to_object(root, "name", json_create_string("张三")) != 0) {
        fprintf(stderr, "Failed to add 'name'\n");
    }
    
    if (add_to_object(root, "age", json_create_int(28)) != 0) {
        fprintf(stderr, "Failed to add 'age'\n");
    }
    
    if (add_to_object(root, "is_active", json_create_bool(1)) != 0) {
        fprintf(stderr, "Failed to add 'is_active'\n");
    }
    
    if (add_to_object(root, "score", json_create_float(95.5)) != 0) {
        fprintf(stderr, "Failed to add 'score'\n");
    }
    
    if (add_to_object(root, "null_value", json_create_null()) != 0) {
        fprintf(stderr, "Failed to add 'null_value'\n");
    }
    
    if (add_to_object(root, "special_chars", 
        json_create_string("包含\"引号\"和\\反斜杠，还有\n换行\t制表符")) != 0) {
        fprintf(stderr, "Failed to add 'special_chars'\n");
    }
    
    JsonValue *address = json_create_object();
    if (address != NULL) {
        add_to_object(address, "city", json_create_string("北京"));
        add_to_object(address, "district", json_create_string("海淀区"));
        add_to_object(address, "street", json_create_string("中关村大街1号"));
        add_to_object(address, "zip_code", json_create_int(100080));
        if (add_to_object(root, "address", address) != 0) {
            json_free(address);
            fprintf(stderr, "Failed to add 'address'\n");
        }
    }
    
    JsonValue *tags = json_create_array();
    if (tags != NULL) {
        add_to_array(tags, json_create_string("程序员"));
        add_to_array(tags, json_create_string("C语言"));
        add_to_array(tags, json_create_string("JSON"));
        add_to_array(tags, json_create_int(2024));
        if (add_to_object(root, "tags", tags) != 0) {
            json_free(tags);
            fprintf(stderr, "Failed to add 'tags'\n");
        }
    }
    
    JsonValue *experience = json_create_array();
    if (experience != NULL) {
        JsonValue *job1 = json_create_object();
        if (job1 != NULL) {
            add_to_object(job1, "company", json_create_string("科技公司A"));
            add_to_object(job1, "position", json_create_string("初级工程师"));
            add_to_object(job1, "years", json_create_int(2));
            add_to_object(job1, "salary", json_create_float(15000.5));
            add_to_array(experience, job1);
        }
        
        JsonValue *job2 = json_create_object();
        if (job2 != NULL) {
            add_to_object(job2, "company", json_create_string("互联网公司B"));
            add_to_object(job2, "position", json_create_string("高级工程师"));
            add_to_object(job2, "years", json_create_int(3));
            add_to_object(job2, "salary", json_create_float(25000.0));
            add_to_array(experience, job2);
        }
        
        if (add_to_object(root, "experience", experience) != 0) {
            json_free(experience);
            fprintf(stderr, "Failed to add 'experience'\n");
        }
    }
    
    printf("\n========== 特殊浮点数测试 (NaN/Infinity) ==========\n");
    
    JsonValue *special_floats = json_create_object();
    if (special_floats != NULL) {
        double nan_val = NAN;
        double inf_val = INFINITY;
        double neg_inf_val = -INFINITY;
        
        printf("添加 NaN 值到 JSON...\n");
        printf("添加 Infinity 值到 JSON...\n");
        printf("添加 -Infinity 值到 JSON...\n");
        
        add_to_object(special_floats, "nan_value", json_create_float(nan_val));
        add_to_object(special_floats, "infinity_value", json_create_float(inf_val));
        add_to_object(special_floats, "negative_infinity", json_create_float(neg_inf_val));
        
        add_to_object(special_floats, "normal_float", json_create_float(3.14159));
        add_to_object(special_floats, "zero", json_create_float(0.0));
        add_to_object(special_floats, "one_point_zero", json_create_float(1.0));
        
        if (add_to_object(root, "special_floating_points", special_floats) != 0) {
            json_free(special_floats);
            fprintf(stderr, "Failed to add 'special_floating_points'\n");
        }
    }
    
    JsonValue *numbers = json_create_array();
    if (numbers != NULL) {
        add_to_array(numbers, json_create_int(0));
        add_to_array(numbers, json_create_int(-1));
        add_to_array(numbers, json_create_int(999999999));
        add_to_array(numbers, json_create_float(3.1415926535));
        add_to_array(numbers, json_create_float(3.0));
        add_to_array(numbers, json_create_float(0.0000001));
        add_to_array(numbers, json_create_float(1.0e10));
        if (add_to_object(root, "test_numbers", numbers) != 0) {
            json_free(numbers);
            fprintf(stderr, "Failed to add 'test_numbers'\n");
        }
    }
    
    JsonValue *skills = json_create_object();
    if (skills != NULL) {
        JsonValue *languages = json_create_array();
        if (languages != NULL) {
            add_to_array(languages, json_create_string("C"));
            add_to_array(languages, json_create_string("C++"));
            add_to_array(languages, json_create_string("Python"));
            add_to_object(skills, "languages", languages);
        }
        
        JsonValue *tools = json_create_array();
        if (tools != NULL) {
            add_to_array(tools, json_create_string("Git"));
            add_to_array(tools, json_create_string("Docker"));
            add_to_array(tools, json_create_string("Linux"));
            add_to_object(skills, "tools", tools);
        }
        
        add_to_object(skills, "level", json_create_int(5));
        
        if (add_to_object(root, "skills", skills) != 0) {
            json_free(skills);
            fprintf(stderr, "Failed to add 'skills'\n");
        }
    }
    
    printf("\n========== 压缩格式输出 ==========\n");
    char *compact_str = json_to_string(root);
    if (compact_str != NULL) {
        printf("%s\n\n", compact_str);
        free(compact_str);
    } else {
        printf("格式化失败\n\n");
        ret = 1;
    }
    
    printf("========== 美化格式输出 ==========\n");
    char *pretty_str = json_to_string_pretty(root);
    if (pretty_str != NULL) {
        printf("%s\n", pretty_str);
        free(pretty_str);
    } else {
        printf("格式化失败\n");
        ret = 1;
    }
    
    printf("\n========== 释放所有内存 ==========\n");
    json_free(root);
    printf("内存已成功释放\n");
    
    printf("\n========== 测试完成 ==========\n");
    if (ret == 0) {
        printf("所有测试通过！\n");
    } else {
        printf("部分测试失败！\n");
    }
    
    return ret;
}
