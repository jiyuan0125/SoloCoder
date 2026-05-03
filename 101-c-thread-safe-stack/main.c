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
    printf("\n\n========== 测试 4: 多线程并发访问 ==========\n\n");
    
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

int main(void) {
    printf("========================================\n");
    printf("   C 语言撤销/重做模块演示程序\n");
    printf("========================================\n");
    
    test_basic_operations();
    test_block_operation();
    test_history_limit();
    test_concurrent_access();
    
    printf("\n\n========== 所有测试完成 ==========\n");
    
    return 0;
}
