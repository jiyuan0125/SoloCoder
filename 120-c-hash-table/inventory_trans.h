#ifndef INVENTORY_TRANS_H
#define INVENTORY_TRANS_H

#include "product_store.h"

#define TRANS_OK                  0
#define TRANS_ERR_NULL           -1
#define TRANS_ERR_NOT_FOUND      -2
#define TRANS_ERR_NEGATIVE       -3
#define TRANS_ERR_MEMORY         -4
#define TRANS_ERR_NO_ACTIVE      -5
#define TRANS_ERR_DUPLICATE      -6

typedef enum {
    OP_ADD,
    OP_UPDATE_QUANTITY,
    OP_DELETE
} TransactionOpType;

typedef struct {
    TransactionOpType type;
    char product_id[PRODUCT_ID_MAX_LEN];
    int old_quantity;
    double old_price;
    char old_name[PRODUCT_NAME_MAX_LEN];
    char old_location[PRODUCT_LOC_MAX_LEN];
    int old_is_deleted;
    int was_existing;
    Product saved_product;
} TransactionLogEntry;

typedef struct {
    InventoryStore* store;
    TransactionLogEntry* logs;
    size_t log_count;
    size_t log_capacity;
    int in_transaction;
} InventoryTransaction;

typedef struct {
    const char* product_id;
    int quantity;
} StockChangeItem;

InventoryTransaction* inventory_transaction_create(InventoryStore* store);
void inventory_transaction_destroy(InventoryTransaction* trans);

int inventory_transaction_begin(InventoryTransaction* trans);
int inventory_transaction_commit(InventoryTransaction* trans);
int inventory_transaction_rollback(InventoryTransaction* trans);

int inventory_transaction_add(InventoryTransaction* trans, const Product* product);
int inventory_transaction_update_quantity(InventoryTransaction* trans, const char* product_id, int delta, int* available);
int inventory_transaction_delete(InventoryTransaction* trans, const char* product_id);

int inventory_batch_inbound(InventoryTransaction* trans, const StockChangeItem* items, size_t count);
int inventory_batch_outbound(InventoryTransaction* trans, const StockChangeItem* items, size_t count, 
                              const char** failed_id, int* available);

#endif
