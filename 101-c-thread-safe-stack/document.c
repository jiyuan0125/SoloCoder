#include "document.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define INITIAL_CAPACITY 1024

Document *document_create(void) {
    Document *doc = (Document *)malloc(sizeof(Document));
    if (!doc) return NULL;
    doc->capacity = INITIAL_CAPACITY;
    doc->text = (char *)malloc(doc->capacity);
    if (!doc->text) {
        free(doc);
        return NULL;
    }
    doc->text[0] = '\0';
    doc->length = 0;
    pthread_rwlock_init(&doc->lock, NULL);
    return doc;
}

void document_destroy(Document *doc) {
    if (!doc) return;
    pthread_rwlock_destroy(&doc->lock);
    free(doc->text);
    free(doc);
}

size_t document_length(Document *doc) {
    if (!doc) return 0;
    size_t len;
    pthread_rwlock_rdlock(&doc->lock);
    len = doc->length;
    pthread_rwlock_unlock(&doc->lock);
    return len;
}

char *document_copy(Document *doc) {
    if (!doc) return NULL;
    pthread_rwlock_rdlock(&doc->lock);
    char *copy = (char *)malloc(doc->length + 1);
    if (copy) {
        memcpy(copy, doc->text, doc->length);
        copy[doc->length] = '\0';
    }
    pthread_rwlock_unlock(&doc->lock);
    return copy;
}

void document_snapshot(Document *doc, char *buffer, size_t buf_size) {
    if (!doc || !buffer || buf_size == 0) return;
    pthread_rwlock_rdlock(&doc->lock);
    size_t copy_len = (doc->length < buf_size - 1) ? doc->length : buf_size - 1;
    memcpy(buffer, doc->text, copy_len);
    buffer[copy_len] = '\0';
    pthread_rwlock_unlock(&doc->lock);
}

static int ensure_capacity(Document *doc, size_t needed) {
    if (doc->capacity >= needed) return 1;
    size_t new_cap = doc->capacity;
    while (new_cap < needed) {
        new_cap *= 2;
    }
    char *new_text = (char *)realloc(doc->text, new_cap);
    if (!new_text) return 0;
    doc->text = new_text;
    doc->capacity = new_cap;
    return 1;
}

static int apply_insert(Document *doc, const Operation *op) {
    if (op->position > doc->length) return 0;
    size_t new_len = doc->length + op->new_len;
    if (!ensure_capacity(doc, new_len + 1)) return 0;
    
    memmove(doc->text + op->position + op->new_len,
            doc->text + op->position,
            doc->length - op->position);
    memcpy(doc->text + op->position, op->new_text, op->new_len);
    doc->length = new_len;
    doc->text[doc->length] = '\0';
    return 1;
}

static int apply_delete(Document *doc, const Operation *op) {
    if (op->position + op->old_len > doc->length) return 0;
    
    memmove(doc->text + op->position,
            doc->text + op->position + op->old_len,
            doc->length - op->position - op->old_len);
    doc->length -= op->old_len;
    doc->text[doc->length] = '\0';
    return 1;
}

static int apply_replace(Document *doc, const Operation *op) {
    if (op->position + op->old_len > doc->length) return 0;
    
    size_t len_diff = op->new_len - op->old_len;
    if (len_diff > 0) {
        size_t new_len = doc->length + len_diff;
        if (!ensure_capacity(doc, new_len + 1)) return 0;
    }
    
    if (len_diff != 0) {
        memmove(doc->text + op->position + op->new_len,
                doc->text + op->position + op->old_len,
                doc->length - op->position - op->old_len);
    }
    
    memcpy(doc->text + op->position, op->new_text, op->new_len);
    doc->length += len_diff;
    doc->text[doc->length] = '\0';
    return 1;
}

int document_apply_operation(Document *doc, const Operation *op) {
    if (!doc || !op) return 0;
    pthread_rwlock_wrlock(&doc->lock);
    int result = 0;
    switch (op->type) {
        case OP_INSERT:
            result = apply_insert(doc, op);
            break;
        case OP_DELETE:
            result = apply_delete(doc, op);
            break;
        case OP_REPLACE:
            result = apply_replace(doc, op);
            break;
        default:
            result = 0;
    }
    pthread_rwlock_unlock(&doc->lock);
    return result;
}

int document_apply_group(Document *doc, const OperationGroup *group) {
    if (!doc || !group) return 0;
    pthread_rwlock_wrlock(&doc->lock);
    OpNode *current = group->head;
    while (current) {
        int result = 0;
        switch (current->op->type) {
            case OP_INSERT:
                result = apply_insert(doc, current->op);
                break;
            case OP_DELETE:
                result = apply_delete(doc, current->op);
                break;
            case OP_REPLACE:
                result = apply_replace(doc, current->op);
                break;
            default:
                result = 0;
        }
        if (!result) {
            pthread_rwlock_unlock(&doc->lock);
            return 0;
        }
        current = current->next;
    }
    pthread_rwlock_unlock(&doc->lock);
    return 1;
}
