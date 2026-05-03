#include "operation.h"
#include <stdlib.h>
#include <string.h>

static char *str_dup(const char *s, size_t len) {
    char *dup = (char *)malloc(len + 1);
    if (dup) {
        memcpy(dup, s, len);
        dup[len] = '\0';
    }
    return dup;
}

Operation *operation_create_insert(size_t pos, const char *text, size_t len) {
    Operation *op = (Operation *)malloc(sizeof(Operation));
    if (!op) return NULL;
    op->type = OP_INSERT;
    op->position = pos;
    op->old_text = NULL;
    op->old_len = 0;
    op->new_text = str_dup(text, len);
    op->new_len = len;
    if (!op->new_text) {
        free(op);
        return NULL;
    }
    return op;
}

Operation *operation_create_delete(size_t pos, const char *text, size_t len) {
    Operation *op = (Operation *)malloc(sizeof(Operation));
    if (!op) return NULL;
    op->type = OP_DELETE;
    op->position = pos;
    op->old_text = str_dup(text, len);
    op->old_len = len;
    op->new_text = NULL;
    op->new_len = 0;
    if (!op->old_text) {
        free(op);
        return NULL;
    }
    return op;
}

Operation *operation_create_replace(size_t pos, const char *old_text, size_t old_len,
                                     const char *new_text, size_t new_len) {
    Operation *op = (Operation *)malloc(sizeof(Operation));
    if (!op) return NULL;
    op->type = OP_REPLACE;
    op->position = pos;
    op->old_text = str_dup(old_text, old_len);
    op->old_len = old_len;
    op->new_text = str_dup(new_text, new_len);
    op->new_len = new_len;
    if (!op->old_text || !op->new_text) {
        free(op->old_text);
        free(op->new_text);
        free(op);
        return NULL;
    }
    return op;
}

void operation_destroy(Operation *op) {
    if (!op) return;
    free(op->old_text);
    free(op->new_text);
    free(op);
}

Operation *operation_inverse(const Operation *op) {
    if (!op) return NULL;
    switch (op->type) {
        case OP_INSERT:
            return operation_create_delete(op->position, op->new_text, op->new_len);
        case OP_DELETE:
            return operation_create_insert(op->position, op->old_text, op->old_len);
        case OP_REPLACE:
            return operation_create_replace(op->position, op->new_text, op->new_len,
                                             op->old_text, op->old_len);
        default:
            return NULL;
    }
}

OperationGroup *operation_group_create(int is_block) {
    OperationGroup *group = (OperationGroup *)malloc(sizeof(OperationGroup));
    if (!group) return NULL;
    group->is_block = is_block;
    group->head = NULL;
    group->tail = NULL;
    group->count = 0;
    return group;
}

void operation_group_add(OperationGroup *group, Operation *op) {
    if (!group || !op) return;
    OpNode *node = (OpNode *)malloc(sizeof(OpNode));
    if (!node) {
        operation_destroy(op);
        return;
    }
    node->op = op;
    node->next = NULL;
    if (!group->head) {
        group->head = node;
        group->tail = node;
    } else {
        group->tail->next = node;
        group->tail = node;
    }
    group->count++;
}

void operation_group_destroy(OperationGroup *group) {
    if (!group) return;
    OpNode *current = group->head;
    while (current) {
        OpNode *next = current->next;
        operation_destroy(current->op);
        free(current);
        current = next;
    }
    free(group);
}

OperationGroup *operation_group_inverse(const OperationGroup *group) {
    if (!group) return NULL;
    OperationGroup *inv_group = operation_group_create(group->is_block);
    if (!inv_group) return NULL;
    
    OpNode *nodes[1024];
    size_t count = 0;
    OpNode *current = group->head;
    while (current && count < 1024) {
        nodes[count++] = current;
        current = current->next;
    }
    
    for (ssize_t i = count - 1; i >= 0; i--) {
        Operation *inv_op = operation_inverse(nodes[i]->op);
        if (!inv_op) {
            operation_group_destroy(inv_group);
            return NULL;
        }
        operation_group_add(inv_group, inv_op);
    }
    return inv_group;
}
