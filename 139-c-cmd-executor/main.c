#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "common.h"
#include "cmd_exec.h"
#include "scheduler.h"
#include "result.h"

static void demo_single_command(void) {
    printf("\n=== 演示1: 执行单个命令 ===\n\n");
    
    cmd_config_t config;
    cmd_config_init(&config);
    config.command = "ls -la /tmp";
    config.label = "列出/tmp目录";
    config.timeout_sec = 10;
    
    cmd_result_t result;
    if (cmd_execute_sync(&config, &result) == 0) {
        print_result_summary(&result, 10);
        cmd_result_free(&result);
    } else {
        printf("命令执行失败\n");
    }
    
    cmd_config_free(&config);
}

static void demo_concurrent_commands(void) {
    printf("\n=== 演示2: 并发执行多个命令 (最大并发数: 3) ===\n\n");
    
    scheduler_t *sched = scheduler_create(3);
    if (!sched) {
        printf("创建调度器失败\n");
        return;
    }
    
    const char *commands[] = {
        "echo 'Command 1: Hello' && sleep 1 && echo 'Command 1 done'",
        "echo 'Command 2: Starting' && sleep 2 && echo 'Command 2 done'",
        "echo 'Command 3: Quick' && sleep 0.5 && echo 'Command 3 done'",
        "echo 'Command 4: Longer' && sleep 1.5 && echo 'Command 4 done'",
        "echo 'Command 5: Last' && sleep 1 && echo 'Command 5 done'",
    };
    
    for (int i = 0; i < 5; i++) {
        cmd_config_t config;
        cmd_config_init(&config);
        config.command = commands[i];
        config.id = i + 1;
        char label[32];
        snprintf(label, sizeof(label), "并发测试命令 #%d", i + 1);
        config.label = strdup(label);
        config.timeout_sec = 10;
        
        scheduler_add_task(sched, &config);
    }
    
    printf("开始执行 %d 个命令，最大并发数: 3\n", scheduler_get_task_count(sched));
    printf("请观察执行顺序（前一批有完成的才会启动新的）...\n\n");
    
    scheduler_run(sched);
    
    size_t result_count;
    cmd_result_t *results = scheduler_get_results(sched, &result_count);
    
    if (results) {
        print_complete_report(results, result_count, DEFAULT_OUTPUT_SUMMARY);
        free(results);
    }
    
    scheduler_destroy(sched);
}

static void demo_timeout(void) {
    printf("\n=== 演示3: 超时处理 (超时时间: 2秒，命令需要睡5秒) ===\n\n");
    
    cmd_config_t config;
    cmd_config_init(&config);
    config.command = "sleep 5 && echo '应该不会看到这条消息'";
    config.label = "超时测试";
    config.timeout_sec = 2;
    
    cmd_result_t result;
    if (cmd_execute_sync(&config, &result) == 0) {
        print_result_summary(&result, 5);
        cmd_result_free(&result);
    }
    
    cmd_config_free(&config);
}

static void demo_environment_and_cwd(void) {
    printf("\n=== 演示4: 环境变量和工作目录 ===\n\n");
    
    scheduler_t *sched = scheduler_create(2);
    if (!sched) return;
    
    cmd_config_t config1;
    cmd_config_init(&config1);
    config1.command = "echo \"HOME=$HOME\\nMY_VAR=$MY_VAR\\nOVERRIDE_VAR=$OVERRIDE_VAR\"";
    config1.label = "环境变量测试-追加";
    config1.id = 1;
    cmd_config_set_env(&config1, "MY_VAR", "Hello from custom env", false);
    cmd_config_set_env(&config1, "OVERRIDE_VAR", "OriginalValue", false);
    scheduler_add_task(sched, &config1);
    
    cmd_config_t config2;
    cmd_config_init(&config2);
    config2.command = "pwd && ls -la";
    config2.label = "工作目录测试(/tmp)";
    config2.id = 2;
    config2.work_dir = "/tmp";
    scheduler_add_task(sched, &config2);
    
    cmd_config_t config3;
    cmd_config_init(&config3);
    config3.command = "echo \"OVERRIDE_VAR=$OVERRIDE_VAR\"";
    config3.label = "环境变量测试-覆盖";
    config3.id = 3;
    cmd_config_set_env(&config3, "OVERRIDE_VAR", "OverriddenValue", true);
    scheduler_add_task(sched, &config3);
    
    scheduler_run(sched);
    
    size_t result_count;
    cmd_result_t *results = scheduler_get_results(sched, &result_count);
    if (results) {
        print_complete_report(results, result_count, 10);
        free(results);
    }
    
    scheduler_destroy(sched);
}

static void demo_output_redirect(void) {
    printf("\n=== 演示5: 输出重定向到文件 ===\n\n");
    
    const char *stdout_file = "/tmp/cmd_exec_stdout.txt";
    const char *stderr_file = "/tmp/cmd_exec_stderr.txt";
    
    scheduler_t *sched = scheduler_create(2);
    if (!sched) return;
    
    cmd_config_t config1;
    cmd_config_init(&config1);
    config1.command = "echo 'Hello to stdout' && ls -la /nonexistent 2>&1";
    config1.label = "测试输出重定向";
    config1.id = 1;
    config1.redirect_stdout = stdout_file;
    config1.redirect_stderr = stderr_file;
    scheduler_add_task(sched, &config1);
    
    scheduler_run(sched);
    
    size_t result_count;
    cmd_result_t *results = scheduler_get_results(sched, &result_count);
    if (results) {
        print_complete_report(results, result_count, 5);
        free(results);
    }
    
    printf("\n--- 重定向文件内容 ---\n");
    printf("%s内容:\n", stdout_file);
    char cmd[256];
    snprintf(cmd, sizeof(cmd), "cat '%s' 2>/dev/null", stdout_file);
    system(cmd);
    
    printf("\n%s内容:\n", stderr_file);
    snprintf(cmd, sizeof(cmd), "cat '%s' 2>/dev/null", stderr_file);
    system(cmd);
    
    unlink(stdout_file);
    unlink(stderr_file);
    
    scheduler_destroy(sched);
}

static void demo_mixed_scenarios(void) {
    printf("\n=== 演示6: 混合场景（成功、失败、超时组合） ===\n\n");
    
    scheduler_t *sched = scheduler_create(4);
    if (!sched) return;
    
    cmd_config_t config;
    
    cmd_config_init(&config);
    config.command = "echo 'Success command' && exit 0";
    config.label = "成功命令";
    config.id = 1;
    config.timeout_sec = 5;
    scheduler_add_task(sched, &config);
    
    cmd_config_init(&config);
    config.command = "echo 'Failure command' && exit 1";
    config.label = "失败命令(退出码1)";
    config.id = 2;
    config.timeout_sec = 5;
    scheduler_add_task(sched, &config);
    
    cmd_config_init(&config);
    config.command = "sleep 10";
    config.label = "超时命令";
    config.id = 3;
    config.timeout_sec = 2;
    scheduler_add_task(sched, &config);
    
    cmd_config_init(&config);
    config.command = "nonexistent_command_12345";
    config.label = "命令不存在";
    config.id = 4;
    config.timeout_sec = 5;
    scheduler_add_task(sched, &config);
    
    cmd_config_init(&config);
    config.command = "echo 'Another success' && echo 'Multi-line' && echo 'Output'";
    config.label = "多行输出";
    config.id = 5;
    config.timeout_sec = 5;
    scheduler_add_task(sched, &config);
    
    printf("执行混合场景测试...\n");
    scheduler_run(sched);
    
    size_t result_count;
    cmd_result_t *results = scheduler_get_results(sched, &result_count);
    if (results) {
        print_complete_report(results, result_count, DEFAULT_OUTPUT_SUMMARY);
        free(results);
    }
    
    scheduler_destroy(sched);
}

int main(void) {
    printf("========================================\n");
    printf("    C 语言命令批量执行模块演示\n");
    printf("========================================\n");
    
    demo_single_command();
    demo_concurrent_commands();
    demo_timeout();
    demo_environment_and_cwd();
    demo_output_redirect();
    demo_mixed_scenarios();
    
    printf("\n=== 所有演示完成 ===\n\n");
    
    return 0;
}
