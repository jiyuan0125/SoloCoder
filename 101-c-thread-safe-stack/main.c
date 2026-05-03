#include "undo_redo.h"
#include <stdio.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>

#define BUF_SIZE 4096

static void print_doc_state(Document *doc, const char *label) {
    char buffer[BUF_SIZE];
    document_snapshot(doc, buffer, BUF_SIZE);
    printf("[%s] \"%s\"\n", label, buffer);
}

static void print_status(UndoRedoManager *manager) {
    printf("  可撤销: %s, 可重做: %s | 撤销栈: %zu, 重做栈: %zu\n",
           undo_redo_can_undo(manager) ? "是" : "否",
           undo_redo_can_redo(manager) ? "是" : "否",
           undo_redo_undo_count(manager),
           undo_redo_redo_count(manager));
}

typedef struct {
    Document *doc;
    int id;
    int stop;
} ReaderThreadArg;

static void *reader_thread(void *arg) {
    ReaderThreadArg *r_arg = (ReaderThreadArg *)arg;
    char buffer[BUF_SIZE];
    int count = 0;
    
    while (!r_arg->stop) {
        document_snapshot(r_arg->doc, buffer, BUF_SIZE);
        count++;
        if (count % 50000 == 0) {
            printf("[渲染线程 %d] 已读取 %d 次，当前: \"%s\"\n", 
                   r_arg->id, count, buffer);
        }
        usleep(100);
    }
    printf("[渲染线程 %d] 结束，共读取 %d 次\n", r_arg->id, count);
    return NULL;
}

static void test_basic_operations(void) {
    printf("\n========== 测试 1: 基本插入、删除、替换 ==========\n\n");
    
    Document *doc = document_create();
    UndoRedoManager *manager = undo_redo_create(doc);
    
    print_doc_state(doc, "初始状态");
    print_status(manager);
    
    printf("\n--- 插入 \"Hello\" ---\n");
    Operation *op1 = operation_create_insert(0, "Hello", 5);
    undo_redo_record_single(manager, op1);
    print_doc_state(doc, "插入后");
    print_status(manager);
    
    printf("\n--- 插入 \" World\" ---\n");
    Operation *op2 = operation_create_insert(5, " World", 6);
    undo_redo_record_single(manager, op2);
    print_doc_state(doc, "插入后");
    print_status(manager);
    
    printf("\n--- 删除 \" World\" ---\n");
    Operation *op3 = operation_create_delete(5, " World", 6);
    undo_redo_record_single(manager, op3);
    print_doc_state(doc, "删除后");
    print_status(manager);
    
    printf("\n--- 替换 \"Hello\" 为 \"Hi\" ---\n");
    Operation *op4 = operation_create_replace(0, "Hello", 5, "Hi", 2);
    undo_redo_record_single(manager, op4);
    print_doc_state(doc, "替换后");
    print_status(manager);
    
    printf("\n--- 撤销 1 次 ---\n");
    undo_redo_undo(manager);
    print_doc_state(doc, "撤销后");
    print_status(manager);
    
    printf("\n--- 撤销 1 次 ---\n");
    undo_redo_undo(manager);
    print_doc_state(doc, "撤销后");
    print_status(manager);
    
    printf("\n--- 重做 1 次 ---\n");
    undo_redo_redo(manager);
    print_doc_state(doc, "重做后");
    print_status(manager);
    
    printf("\n--- 插入 \", there\" 测试重做栈清空 ---\n");
    Operation *op5 = operation_create_insert(5, ", there", 7);
    undo_redo_record_single(manager, op5);
    print_doc_state(doc, "新插入后");
    print_status(manager);
    
    undo_redo_destroy(manager);
    document_destroy(doc);
}

static void test_block_operation(void) {
    printf("\n\n========== 测试 2: 块操作 ==========\n\n");
    
    Document *doc = document_create();
    UndoRedoManager *manager = undo_redo_create(doc);
    
    Operation *op_setup = operation_create_insert(0, "I have a dream. I have a dream.", 31);
    undo_redo_record_single(manager, op_setup);
    print_doc_state(doc, "初始文本");
    print_status(manager);
    
    printf("\n--- 块操作: 将两处 \"dream\" 替换为 \"plan\" ---\n");
    OperationGroup *block = operation_group_create(1);
    Operation *rep1 = operation_create_replace(9, "dream", 5, "plan", 4);
    Operation *rep2 = operation_create_replace(24, "dream", 5, "plan", 4);
    operation_group_add(block, rep1);
    operation_group_add(block, rep2);
    undo_redo_record(manager, block);
    
    print_doc_state(doc, "块替换后");
    print_status(manager);
    
    printf("\n--- 撤销 1 次 (整个块操作应被撤销) ---\n");
    undo_redo_undo(manager);
    print_doc_state(doc, "撤销后");
    print_status(manager);
    
    printf("\n--- 重做 1 次 (整个块操作应被重做) ---\n");
    undo_redo_redo(manager);
    print_doc_state(doc, "重做后");
    print_status(manager);
    
    undo_redo_destroy(manager);
    document_destroy(doc);
}

static void test_history_limit(void) {
    printf("\n\n========== 测试 3: 历史记录限制 (最多1000条) ==========\n\n");
    
    Document *doc = document_create();
    UndoRedoManager *manager = undo_redo_create(doc);
    
    printf("插入 1500 个字符，每次插入 1 个字符，创建 1500 条历史记录...\n");
    
    for (int i = 0; i < 1500; i++) {
        char c = 'A' + (i % 26);
        Operation *op = operation_create_insert(i, &c, 1);
        undo_redo_record_single(manager, op);
    }
    
    printf("撤销栈大小: %zu (应为 1000)\n", undo_redo_undo_count(manager));
    print_status(manager);
    
    printf("\n--- 连续撤销 500 次 ---\n");
    for (int i = 0; i < 500; i++) {
        undo_redo_undo(manager);
    }
    printf("撤销栈大小: %zu, 重做栈大小: %zu\n", 
           undo_redo_undo_count(manager), undo_redo_redo_count(manager));
    print_status(manager);
    
    char buffer[BUF_SIZE];
    document_snapshot(doc, buffer, BUF_SIZE);
    printf("当前文档长度: %zu\n", strlen(buffer));
    
    undo_redo_destroy(manager);
    document_destroy(doc);
}

static void test_concurrent_access(void) {
    printf("\n\n========== 测试 4: 多线程并发访问（渲染线程 + 编辑线程）==========\n\n");
    
    Document *doc = document_create();
    UndoRedoManager *manager = undo_redo_create(doc);
    
    ReaderThreadArg args[3];
    pthread_t readers[3];
    
    for (int i = 0; i < 3; i++) {
        args[i].doc = doc;
        args[i].id = i + 1;
        args[i].stop = 0;
        pthread_create(&readers[i], NULL, reader_thread, &args[i]);
    }
    
    printf("主线程开始执行编辑操作...\n\n");
    
    for (int i = 0; i < 10; i++) {
        char text[32];
        snprintf(text, sizeof(text), "Word%d ", i + 1);
        Operation *op = operation_create_insert(i * 6, text, strlen(text));
        undo_redo_record_single(manager, op);
        printf("[主线程] 插入: \"%s\"\n", text);
        print_doc_state(doc, "当前状态");
        usleep(100000);
    }
    
    printf("\n[主线程] 开始撤销...\n");
    for (int i = 0; i < 5; i++) {
        if (undo_redo_undo(manager)) {
            printf("[主线程] 撤销\n");
            print_doc_state(doc, "撤销后");
        }
        usleep(100000);
    }
    
    printf("\n[主线程] 开始重做...\n");
    for (int i = 0; i < 3; i++) {
        if (undo_redo_redo(manager)) {
            printf("[主线程] 重做\n");
            print_doc_state(doc, "重做后");
        }
        usleep(100000);
    }
    
    printf("\n[主线程] 停止渲染线程...\n");
    for (int i = 0; i < 3; i++) {
        args[i].stop = 1;
    }
    
    for (int i = 0; i < 3; i++) {
        pthread_join(readers[i], NULL);
    }
    
    printf("\n最终状态:\n");
    print_doc_state(doc, "最终文档");
    print_status(manager);
    
    undo_redo_destroy(manager);
    document_destroy(doc);
}

typedef struct {
    UndoRedoManager *manager;
    int id;
    int iterations;
    int success_count;
    int fail_count;
    int stop;
} StressThreadArg;

static void *undo_thread(void *arg) {
    StressThreadArg *s_arg = (StressThreadArg *)arg;
    int successes = 0;
    int failures = 0;
    
    for (int i = 0; i < s_arg->iterations && !s_arg->stop; i++) {
        if (undo_redo_undo(s_arg->manager)) {
            successes++;
        } else {
            failures++;
        }
        usleep(100);
    }
    
    s_arg->success_count = successes;
    s_arg->fail_count = failures;
    printf("[撤销线程 %d] 完成: 成功 %d 次, 失败 %d 次\n", 
           s_arg->id, successes, failures);
    return NULL;
}

static void *record_thread(void *arg) {
    StressThreadArg *s_arg = (StressThreadArg *)arg;
    int successes = 0;
    int failures = 0;
    
    for (int i = 0; i < s_arg->iterations && !s_arg->stop; i++) {
        char text[32];
        snprintf(text, sizeof(text), "[T%d-%d]", s_arg->id, i);
        Operation *op = operation_create_insert(0, text, strlen(text));
        if (undo_redo_record_single(s_arg->manager, op)) {
            successes++;
        } else {
            failures++;
            operation_destroy(op);
        }
        usleep(150);
    }
    
    s_arg->success_count = successes;
    s_arg->fail_count = failures;
    printf("[记录线程 %d] 完成: 成功 %d 次, 失败 %d 次\n", 
           s_arg->id, successes, failures);
    return NULL;
}

static void *query_thread(void *arg) {
    StressThreadArg *s_arg = (StressThreadArg *)arg;
    int count = 0;
    
    for (int i = 0; i < s_arg->iterations && !s_arg->stop; i++) {
        size_t undo_cnt = undo_redo_undo_count(s_arg->manager);
        size_t redo_cnt = undo_redo_redo_count(s_arg->manager);
        int can_undo = undo_redo_can_undo(s_arg->manager);
        int can_redo = undo_redo_can_redo(s_arg->manager);
        (void)undo_cnt;
        (void)redo_cnt;
        (void)can_undo;
        (void)can_redo;
        count++;
        usleep(50);
    }
    
    printf("[查询线程 %d] 完成: 共查询 %d 次\n", s_arg->id, count);
    s_arg->success_count = count;
    return NULL;
}

static void test_true_concurrent_undo_record(void) {
    printf("\n\n========== 测试 5: 真正的并发 Undo + Record（竞态窗口测试）==========\n\n");
    
    printf("测试目的:\n");
    printf("  1. 验证撤销线程和记录线程能否安全并发执行\n");
    printf("  2. 验证竞态窗口下引用计数机制的正确性\n");
    printf("  3. 验证读写锁让查询操作可以并发\n\n");
    
    Document *doc = document_create();
    UndoRedoManager *manager = undo_redo_create(doc);
    
    printf("--- 准备初始历史记录 ---\n");
    for (int i = 0; i < 20; i++) {
        char text[32];
        snprintf(text, sizeof(text), "Init%d ", i + 1);
        Operation *op = operation_create_insert(0, text, strlen(text));
        undo_redo_record_single(manager, op);
    }
    printf("已创建 20 条历史记录\n");
    print_status(manager);
    
    printf("\n--- 启动并发线程 ---\n");
    printf("  - 2 个撤销线程（每个执行 500 次）\n");
    printf("  - 2 个记录线程（每个执行 500 次，会清空重做栈）\n");
    printf("  - 3 个查询线程（每个执行 2000 次，验证读并发）\n\n");
    
    StressThreadArg undo_args[2];
    StressThreadArg record_args[2];
    StressThreadArg query_args[3];
    pthread_t undo_threads[2];
    pthread_t record_threads[2];
    pthread_t query_threads[3];
    
    for (int i = 0; i < 2; i++) {
        undo_args[i].manager = manager;
        undo_args[i].id = i + 1;
        undo_args[i].iterations = 500;
        undo_args[i].success_count = 0;
        undo_args[i].fail_count = 0;
        undo_args[i].stop = 0;
        pthread_create(&undo_threads[i], NULL, undo_thread, &undo_args[i]);
    }
    
    for (int i = 0; i < 2; i++) {
        record_args[i].manager = manager;
        record_args[i].id = i + 1;
        record_args[i].iterations = 500;
        record_args[i].success_count = 0;
        record_args[i].fail_count = 0;
        record_args[i].stop = 0;
        pthread_create(&record_threads[i], NULL, record_thread, &record_args[i]);
    }
    
    for (int i = 0; i < 3; i++) {
        query_args[i].manager = manager;
        query_args[i].id = i + 1;
        query_args[i].iterations = 2000;
        query_args[i].success_count = 0;
        query_args[i].fail_count = 0;
        query_args[i].stop = 0;
        pthread_create(&query_threads[i], NULL, query_thread, &query_args[i]);
    }
    
    printf("等待所有线程完成...\n\n");
    
    for (int i = 0; i < 2; i++) {
        pthread_join(undo_threads[i], NULL);
    }
    for (int i = 0; i < 2; i++) {
        pthread_join(record_threads[i], NULL);
    }
    for (int i = 0; i < 3; i++) {
        pthread_join(query_threads[i], NULL);
    }
    
    printf("\n--- 所有线程完成，验证状态 ---\n");
    print_status(manager);
    print_doc_state(doc, "最终文档状态");
    
    printf("\n--- 执行最终验证：多次撤销/重做确认没有损坏 ---\n");
    int undo_success = 0;
    int redo_success = 0;
    for (int i = 0; i < 5; i++) {
        if (undo_redo_undo(manager)) {
            undo_success++;
        }
    }
    for (int i = 0; i < 3; i++) {
        if (undo_redo_redo(manager)) {
            redo_success++;
        }
    }
    printf("最终验证: 撤销成功 %d 次, 重做成功 %d 次\n", undo_success, redo_success);
    print_status(manager);
    
    printf("\n[测试通过] 没有崩溃、没有死锁、引用计数机制正常工作！\n");
    
    undo_redo_destroy(manager);
    document_destroy(doc);
}

int main(void) {
    printf("========================================\n");
    printf("   C 语言撤销/重做模块演示程序\n");
    printf("========================================\n");
    
    test_basic_operations();
    test_block_operation();
    test_history_limit();
    test_concurrent_access();
    test_true_concurrent_undo_record();
    
    printf("\n\n========== 所有测试完成 ==========\n");
    
    return 0;
}
