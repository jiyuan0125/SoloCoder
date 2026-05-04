#include <stdio.h>
#include <stdlib.h>
#include "template_engine.h"

int main(void)
{
    TE_Engine *engine = te_engine_create();
    
    TE_Value *data = te_value_object();
    te_object_set(data, "show", te_value_bool(0));
    te_object_set(data, "name", te_value_string("Test"));
    
    const char *template = "{{#if show}}显示A{{#else}}显示B{{/if}}";
    printf("模板: %s\n", template);
    printf("show = false\n");
    
    char *result = te_render(engine, template, data);
    if (result) {
        printf("结果: [%s]\n", result);
        printf("预期: [显示B]\n");
        if (strcmp(result, "显示B") == 0) {
            printf("✅ 测试通过！\n");
        } else {
            printf("❌ 测试失败！\n");
        }
        free(result);
    } else if (engine->error.has_error) {
        printf("错误: %s\n", engine->error.message);
    }
    
    printf("\n--- 测试 show = true ---\n");
    te_object_set(data, "show", te_value_bool(1));
    result = te_render(engine, template, data);
    if (result) {
        printf("结果: [%s]\n", result);
        printf("预期: [显示A]\n");
        if (strcmp(result, "显示A") == 0) {
            printf("✅ 测试通过！\n");
        } else {
            printf("❌ 测试失败！\n");
        }
        free(result);
    }
    
    printf("\n--- 测试 unless + else ---\n");
    const char *template2 = "{{#unless show}}unless显示A{{#else}}unless显示B{{/unless}}";
    printf("模板: %s\n", template2);
    printf("show = true\n");
    
    result = te_render(engine, template2, data);
    if (result) {
        printf("结果: [%s]\n", result);
        printf("预期: [unless显示B]\n");
        if (strcmp(result, "unless显示B") == 0) {
            printf("✅ 测试通过！\n");
        } else {
            printf("❌ 测试失败！\n");
        }
        free(result);
    }
    
    te_value_free(data);
    te_engine_destroy(engine);
    
    return 0;
}
