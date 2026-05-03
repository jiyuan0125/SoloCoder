#ifndef OPERATION_H
#define OPERATION_H

#include <stddef.h>

typedef enum {
    OP_INSERT,
    OP_DELETE,
    OP_REPLACE
} OpType;

typedef struct {
    OpType type;
    size_t position;
    char *old_text;
    size_t old_len;
    char *new_text;
    size_t new_len;
} Operation;

typedef struct OpNode {
    Operation *op;
    struct OpNode *next;
} OpNode;

typedef struct {
    int is_block;
    OpNode *head;
    OpNode *tail;
    size_t count;
} OperationGroup;

Operation *operation_create_insert(size_t pos, const char *text, size_t len);
Operation *operation_create_delete(size_t pos, const char *text, size_t len);
Operation *operation_create_replace(size_t pos, const char *old_text, size_t old_len,
                                     const char *new_text, size_t new_len);
void operation_destroy(Operation *op);
Operation *operation_inverse(const Operation *op);

OperationGroup *operation_group_create(int is_block);
void operation_group_add(OperationGroup *group, Operation *op);
void operation_group_destroy(OperationGroup *group);
OperationGroup *operation_group_inverse(const OperationGroup *group);

#endif
