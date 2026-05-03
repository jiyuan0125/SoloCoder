#ifndef DOCUMENT_H
#define DOCUMENT_H

#include "operation.h"
#include <pthread.h>
#include <stddef.h>

typedef struct {
    char *text;
    size_t length;
    size_t capacity;
    pthread_rwlock_t lock;
} Document;

Document *document_create(void);
void document_destroy(Document *doc);
size_t document_length(Document *doc);
char *document_copy(Document *doc);
void document_snapshot(Document *doc, char *buffer, size_t buf_size);

int document_apply_operation(Document *doc, const Operation *op);
int document_apply_group(Document *doc, const OperationGroup *group);

#endif
