#include "inventory_trans.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#define INITIAL_LOG_CAPACITY  16

static int transaction_log_reserve(InventoryTransaction* trans, size_t new_capacity) {
    if (new_capacity <= trans->log_capacity) {
        return TRANS_OK;
    }
    
    TransactionLogEntry* new_logs = (TransactionLogEntry*)realloc(
        trans->logs, new_capacity * sizeof(TransactionLogEntry));
    if (!new_logs) {
        return TRANS_ERR_MEMORY;
    }
    
    trans->logs = new_logs;
    trans->log_capacity = new_capacity;
    return TRANS_OK;
}

InventoryTransaction* inventory_transaction_create(InventoryStore* store) {
    if (!store) return NULL;
    
    InventoryTransaction* trans = (InventoryTransaction*)malloc(sizeof(InventoryTransaction));
    if (!trans) return NULL;
    
    trans->logs = (TransactionLogEntry*)malloc(INITIAL_LOG_CAPACITY * sizeof(TransactionLogEntry));
    if (!trans->logs) {
        free(trans);
        return NULL;
    }
    
    trans->store = store;
    trans->log_count = 0;
    trans->log_capacity = INITIAL_LOG_CAPACITY;
    trans->in_transaction = 0;
    
    return trans;
}

void inventory_transaction_destroy(InventoryTransaction* trans) {
    if (!trans) return;
    
    if (trans->in_transaction) {
        inventory_transaction_rollback(trans);
    }
    
    free(trans->logs);
    free(trans);
}

int inventory_transaction_begin(InventoryTransaction* trans) {
    if (!trans) return TRANS_ERR_NULL;
    
    if (trans->in_transaction) {
        inventory_transaction_rollback(trans);
    }
    
    trans->log_count = 0;
    trans->in_transaction = 1;
    
    return TRANS_OK;
}

int inventory_transaction_commit(InventoryTransaction* trans) {
    if (!trans) return TRANS_ERR_NULL;
    
    if (!trans->in_transaction) {
        return TRANS_ERR_NO_ACTIVE;
    }
    
    trans->log_count = 0;
    trans->in_transaction = 0;
    
    return TRANS_OK;
}

static void remove_last_node(InventoryStore* store, const char* product_id) {
    unsigned long idx = hash_function(product_id, store->bucket_count);
    
    HashNode* prev = NULL;
    HashNode* curr = store->buckets[idx];
    
    while (curr) {
        if (strcasecmp_custom(curr->product.id, product_id) == 0) {
            if (prev) {
                prev->next = curr->next;
            } else {
                store->buckets[idx] = curr->next;
            }
            free(curr);
            store->size--;
            return;
        }
        prev = curr;
        curr = curr->next;
    }
}

int inventory_transaction_rollback(InventoryTransaction* trans) {
    if (!trans) return TRANS_ERR_NULL;
    
    if (!trans->in_transaction) {
        return TRANS_ERR_NO_ACTIVE;
    }
    
    for (size_t i = trans->log_count; i > 0; i--) {
        TransactionLogEntry* entry = &trans->logs[i - 1];
        
        switch (entry->type) {
            case OP_ADD:
                if (!entry->was_existing) {
                    remove_last_node(trans->store, entry->product_id);
                } else {
                    Product* p = inventory_store_find(trans->store, entry->product_id);
                    if (p) {
                        *p = entry->saved_product;
                    }
                }
                break;
                
            case OP_UPDATE_QUANTITY: {
                Product* p = inventory_store_find(trans->store, entry->product_id);
                if (p) {
                    p->quantity = entry->old_quantity;
                }
                break;
            }
            
            case OP_DELETE: {
                inventory_store_restore_deleted(trans->store, entry->product_id, entry->old_is_deleted);
                break;
            }
        }
    }
    
    trans->log_count = 0;
    trans->in_transaction = 0;
    
    return TRANS_OK;
}

static int log_add_operation(InventoryTransaction* trans, TransactionOpType type, 
                              const char* product_id, const Product* existing) {
    if (trans->log_count >= trans->log_capacity) {
        int ret = transaction_log_reserve(trans, trans->log_capacity * 2);
        if (ret != TRANS_OK) {
            return ret;
        }
    }
    
    TransactionLogEntry* entry = &trans->logs[trans->log_count++];
    entry->type = type;
    strncpy(entry->product_id, product_id, PRODUCT_ID_MAX_LEN - 1);
    entry->product_id[PRODUCT_ID_MAX_LEN - 1] = '\0';
    
    if (existing) {
        entry->old_quantity = existing->quantity;
        entry->old_price = existing->price;
        strncpy(entry->old_name, existing->name, PRODUCT_NAME_MAX_LEN - 1);
        entry->old_name[PRODUCT_NAME_MAX_LEN - 1] = '\0';
        strncpy(entry->old_location, existing->location, PRODUCT_LOC_MAX_LEN - 1);
        entry->old_location[PRODUCT_LOC_MAX_LEN - 1] = '\0';
        entry->old_is_deleted = existing->is_deleted;
        entry->was_existing = 1;
        entry->saved_product = *existing;
    } else {
        entry->was_existing = 0;
        entry->old_is_deleted = 0;
    }
    
    return TRANS_OK;
}

int inventory_transaction_add(InventoryTransaction* trans, const Product* product) {
    if (!trans || !product) return TRANS_ERR_NULL;
    
    if (!trans->in_transaction) {
        return TRANS_ERR_NO_ACTIVE;
    }
    
    Product* existing = inventory_store_find(trans->store, product->id);
    
    if (existing) {
        if (!existing->is_deleted) {
            return TRANS_ERR_DUPLICATE;
        }
        log_add_operation(trans, OP_ADD, product->id, existing);
        *existing = *product;
        existing->is_deleted = 0;
        trans->store->deleted_count--;
    } else {
        log_add_operation(trans, OP_ADD, product->id, NULL);
        int ret = inventory_store_add(trans->store, product);
        if (ret != INVENTORY_STORE_OK) {
            trans->log_count--;
            if (ret == INVENTORY_STORE_ERR_MEMORY) return TRANS_ERR_MEMORY;
            if (ret == INVENTORY_STORE_ERR_DUPLICATE) return TRANS_ERR_DUPLICATE;
            return TRANS_ERR_NULL;
        }
    }
    
    return TRANS_OK;
}

int inventory_transaction_update_quantity(InventoryTransaction* trans, const char* product_id, 
                                            int delta, int* available) {
    if (!trans || !product_id) return TRANS_ERR_NULL;
    
    if (!trans->in_transaction) {
        return TRANS_ERR_NO_ACTIVE;
    }
    
    Product* product = inventory_store_find(trans->store, product_id);
    if (!product) {
        return TRANS_ERR_NOT_FOUND;
    }
    
    int new_quantity = product->quantity + delta;
    if (new_quantity < 0) {
        if (available) {
            *available = product->quantity;
        }
        return TRANS_ERR_NEGATIVE;
    }
    
    if (delta != 0) {
        log_add_operation(trans, OP_UPDATE_QUANTITY, product_id, product);
        product->quantity = new_quantity;
    }
    
    return TRANS_OK;
}

int inventory_transaction_delete(InventoryTransaction* trans, const char* product_id) {
    if (!trans || !product_id) return TRANS_ERR_NULL;
    
    if (!trans->in_transaction) {
        return TRANS_ERR_NO_ACTIVE;
    }
    
    Product* product = inventory_store_find(trans->store, product_id);
    if (!product) {
        return TRANS_ERR_NOT_FOUND;
    }
    
    log_add_operation(trans, OP_DELETE, product_id, product);
    product->is_deleted = 1;
    trans->store->deleted_count++;
    
    return TRANS_OK;
}

int inventory_batch_inbound(InventoryTransaction* trans, const StockChangeItem* items, size_t count) {
    if (!trans || !items || count == 0) return TRANS_ERR_NULL;
    
    int ret = inventory_transaction_begin(trans);
    if (ret != TRANS_OK) return ret;
    
    for (size_t i = 0; i < count; i++) {
        if (items[i].quantity < 0) {
            inventory_transaction_rollback(trans);
            return TRANS_ERR_NEGATIVE;
        }
        
        ret = inventory_transaction_update_quantity(trans, items[i].product_id, items[i].quantity, NULL);
        if (ret != TRANS_OK) {
            inventory_transaction_rollback(trans);
            return ret;
        }
    }
    
    ret = inventory_transaction_commit(trans);
    return ret;
}

int inventory_batch_outbound(InventoryTransaction* trans, const StockChangeItem* items, size_t count,
                              const char** failed_id, int* available) {
    if (!trans || !items || count == 0) return TRANS_ERR_NULL;
    
    if (failed_id) *failed_id = NULL;
    if (available) *available = 0;
    
    for (size_t i = 0; i < count; i++) {
        if (items[i].quantity < 0) {
            if (failed_id) *failed_id = items[i].product_id;
            return TRANS_ERR_NEGATIVE;
        }
    }
    
    int ret = inventory_transaction_begin(trans);
    if (ret != TRANS_OK) return ret;
    
    for (size_t i = 0; i < count; i++) {
        int avail = 0;
        ret = inventory_transaction_update_quantity(trans, items[i].product_id, -items[i].quantity, &avail);
        if (ret != TRANS_OK) {
            inventory_transaction_rollback(trans);
            if (failed_id) *failed_id = items[i].product_id;
            if (available) *available = avail;
            return ret;
        }
    }
    
    ret = inventory_transaction_commit(trans);
    return ret;
}
