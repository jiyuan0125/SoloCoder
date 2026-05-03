#include "undo_redo.h"
#include <stdlib.h>
#include <stdio.h>

static HistoryNode *node_create(OperationGroup *group) {
    HistoryNode *node = (HistoryNode *)malloc(sizeof(HistoryNode));
    if (!node) return NULL;
    node->group = group;
    node->next = NULL;
    node->prev = NULL;
    node->refcount = 1;
    return node;
}

static void node_destroy(HistoryNode *node) {
    if (!node) return;
    operation_group_destroy(node->group);
    free(node);
}

static void node_release(HistoryNode *node) {
    if (!node) return;
    node->refcount--;
    if (node->refcount == 0) {
        node_destroy(node);
    }
}

static void clear_redo_stack(UndoRedoManager *manager) {
    HistoryNode *current = manager->redo_stack;
    manager->redo_stack = NULL;
    manager->redo_count = 0;
    
    while (current) {
        HistoryNode *next = current->next;
        current->next = NULL;
        current->prev = NULL;
        node_release(current);
        current = next;
    }
}

static void trim_undo_history(UndoRedoManager *manager) {
    while (manager->undo_count > MAX_HISTORY) {
        HistoryNode *old_head = manager->undo_head;
        if (!old_head) break;
        
        manager->undo_head = old_head->next;
        if (manager->undo_head) {
            manager->undo_head->prev = NULL;
        } else {
            manager->undo_tail = NULL;
        }
        
        old_head->next = NULL;
        old_head->prev = NULL;
        node_release(old_head);
        manager->undo_count--;
    }
}

UndoRedoManager *undo_redo_create(Document *doc) {
    if (!doc) return NULL;
    UndoRedoManager *manager = (UndoRedoManager *)malloc(sizeof(UndoRedoManager));
    if (!manager) return NULL;
    
    manager->doc = doc;
    manager->undo_head = NULL;
    manager->undo_tail = NULL;
    manager->undo_count = 0;
    manager->redo_stack = NULL;
    manager->redo_count = 0;
    
    pthread_rwlock_init(&manager->stack_rwlock, NULL);
    return manager;
}

void undo_redo_destroy(UndoRedoManager *manager) {
    if (!manager) return;
    
    pthread_rwlock_wrlock(&manager->stack_rwlock);
    while (manager->undo_head) {
        HistoryNode *next = manager->undo_head->next;
        node_destroy(manager->undo_head);
        manager->undo_head = next;
    }
    while (manager->redo_stack) {
        HistoryNode *next = manager->redo_stack->next;
        node_destroy(manager->redo_stack);
        manager->redo_stack = next;
    }
    pthread_rwlock_unlock(&manager->stack_rwlock);
    
    pthread_rwlock_destroy(&manager->stack_rwlock);
    free(manager);
}

int undo_redo_record(UndoRedoManager *manager, OperationGroup *group) {
    if (!manager || !group) return 0;
    
    pthread_rwlock_wrlock(&manager->stack_rwlock);
    
    clear_redo_stack(manager);
    
    HistoryNode *node = node_create(group);
    if (!node) {
        pthread_rwlock_unlock(&manager->stack_rwlock);
        return 0;
    }
    
    if (!manager->undo_tail) {
        manager->undo_head = node;
        manager->undo_tail = node;
    } else {
        manager->undo_tail->next = node;
        node->prev = manager->undo_tail;
        manager->undo_tail = node;
    }
    manager->undo_count++;
    
    trim_undo_history(manager);
    
    pthread_rwlock_unlock(&manager->stack_rwlock);
    
    document_apply_group(manager->doc, group);
    
    return 1;
}

int undo_redo_record_single(UndoRedoManager *manager, Operation *op) {
    if (!manager || !op) return 0;
    
    OperationGroup *group = operation_group_create(0);
    if (!group) {
        operation_destroy(op);
        return 0;
    }
    operation_group_add(group, op);
    
    return undo_redo_record(manager, group);
}

int undo_redo_can_undo(UndoRedoManager *manager) {
    if (!manager) return 0;
    pthread_rwlock_rdlock(&manager->stack_rwlock);
    int can_undo = (manager->undo_tail != NULL);
    pthread_rwlock_unlock(&manager->stack_rwlock);
    return can_undo;
}

int undo_redo_can_redo(UndoRedoManager *manager) {
    if (!manager) return 0;
    pthread_rwlock_rdlock(&manager->stack_rwlock);
    int can_redo = (manager->redo_stack != NULL);
    pthread_rwlock_unlock(&manager->stack_rwlock);
    return can_redo;
}

int undo_redo_undo(UndoRedoManager *manager) {
    if (!manager) return 0;
    
    pthread_rwlock_wrlock(&manager->stack_rwlock);
    
    if (!manager->undo_tail) {
        pthread_rwlock_unlock(&manager->stack_rwlock);
        return 0;
    }
    
    HistoryNode *node = manager->undo_tail;
    manager->undo_tail = node->prev;
    if (manager->undo_tail) {
        manager->undo_tail->next = NULL;
    } else {
        manager->undo_head = NULL;
    }
    manager->undo_count--;
    
    node->next = manager->redo_stack;
    manager->redo_stack = node;
    manager->redo_count++;
    
    node->refcount++;
    
    OperationGroup *inv_group = operation_group_inverse(node->group);
    
    pthread_rwlock_unlock(&manager->stack_rwlock);
    
    if (!inv_group) {
        pthread_rwlock_wrlock(&manager->stack_rwlock);
        node_release(node);
        pthread_rwlock_unlock(&manager->stack_rwlock);
        return 0;
    }
    
    int result = document_apply_group(manager->doc, inv_group);
    operation_group_destroy(inv_group);
    
    pthread_rwlock_wrlock(&manager->stack_rwlock);
    node_release(node);
    pthread_rwlock_unlock(&manager->stack_rwlock);
    
    return result;
}

int undo_redo_redo(UndoRedoManager *manager) {
    if (!manager) return 0;
    
    pthread_rwlock_wrlock(&manager->stack_rwlock);
    
    if (!manager->redo_stack) {
        pthread_rwlock_unlock(&manager->stack_rwlock);
        return 0;
    }
    
    HistoryNode *node = manager->redo_stack;
    manager->redo_stack = node->next;
    manager->redo_count--;
    
    node->refcount++;
    
    if (!manager->undo_tail) {
        manager->undo_head = node;
        manager->undo_tail = node;
        node->prev = NULL;
        node->next = NULL;
    } else {
        manager->undo_tail->next = node;
        node->prev = manager->undo_tail;
        node->next = NULL;
        manager->undo_tail = node;
    }
    manager->undo_count++;
    
    OperationGroup *group = node->group;
    
    pthread_rwlock_unlock(&manager->stack_rwlock);
    
    int result = document_apply_group(manager->doc, group);
    
    pthread_rwlock_wrlock(&manager->stack_rwlock);
    node_release(node);
    pthread_rwlock_unlock(&manager->stack_rwlock);
    
    return result;
}

size_t undo_redo_undo_count(UndoRedoManager *manager) {
    if (!manager) return 0;
    pthread_rwlock_rdlock(&manager->stack_rwlock);
    size_t count = manager->undo_count;
    pthread_rwlock_unlock(&manager->stack_rwlock);
    return count;
}

size_t undo_redo_redo_count(UndoRedoManager *manager) {
    if (!manager) return 0;
    pthread_rwlock_rdlock(&manager->stack_rwlock);
    size_t count = manager->redo_count;
    pthread_rwlock_unlock(&manager->stack_rwlock);
    return count;
}
