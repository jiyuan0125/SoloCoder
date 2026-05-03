#ifndef UNDO_REDO_H
#define UNDO_REDO_H

#include "document.h"
#include "operation.h"
#include <pthread.h>
#include <stddef.h>

#define MAX_HISTORY 1000

typedef struct HistoryNode {
    OperationGroup *group;
    struct HistoryNode *next;
    struct HistoryNode *prev;
    int refcount;
} HistoryNode;

typedef struct {
    Document *doc;
    
    HistoryNode *undo_head;
    HistoryNode *undo_tail;
    size_t undo_count;
    
    HistoryNode *redo_stack;
    size_t redo_count;
    
    pthread_rwlock_t stack_rwlock;
} UndoRedoManager;

UndoRedoManager *undo_redo_create(Document *doc);
void undo_redo_destroy(UndoRedoManager *manager);

int undo_redo_record(UndoRedoManager *manager, OperationGroup *group);
int undo_redo_record_single(UndoRedoManager *manager, Operation *op);

int undo_redo_can_undo(UndoRedoManager *manager);
int undo_redo_can_redo(UndoRedoManager *manager);

int undo_redo_undo(UndoRedoManager *manager);
int undo_redo_redo(UndoRedoManager *manager);

size_t undo_redo_undo_count(UndoRedoManager *manager);
size_t undo_redo_redo_count(UndoRedoManager *manager);

#endif
