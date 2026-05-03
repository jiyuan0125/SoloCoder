#ifndef PRODUCT_STORE_H
#define PRODUCT_STORE_H

#include <stddef.h>

#define PRODUCT_ID_MAX_LEN    64
#define PRODUCT_NAME_MAX_LEN  128
#define PRODUCT_LOC_MAX_LEN   64

#define INVENTORY_STORE_OK             0
#define INVENTORY_STORE_ERR_NULL       -1
#define INVENTORY_STORE_ERR_DUPLICATE  -2
#define INVENTORY_STORE_ERR_NOT_FOUND  -3
#define INVENTORY_STORE_ERR_MEMORY     -4
#define INVENTORY_STORE_ERR_NEGATIVE   -5

typedef struct {
    char id[PRODUCT_ID_MAX_LEN];
    char name[PRODUCT_NAME_MAX_LEN];
    int quantity;
    double price;
    char location[PRODUCT_LOC_MAX_LEN];
    int is_deleted;
} Product;

typedef struct HashNode {
    Product product;
    struct HashNode* next;
} HashNode;

typedef struct {
    HashNode** buckets;
    size_t bucket_count;
    size_t size;
    size_t deleted_count;
} InventoryStore;

InventoryStore* inventory_store_create(size_t initial_capacity);
void inventory_store_destroy(InventoryStore* store);

int inventory_store_add(InventoryStore* store, const Product* product);
Product* inventory_store_find(const InventoryStore* store, const char* product_id);
int inventory_store_update_quantity(InventoryStore* store, const char* product_id, int delta);
int inventory_store_mark_deleted(InventoryStore* store, const char* product_id);

size_t inventory_store_size(const InventoryStore* store);
size_t inventory_store_active_count(const InventoryStore* store);

void str_tolower(char* dest, const char* src, size_t max_len);
int strcasecmp_custom(const char* s1, const char* s2);

int inventory_store_reserve(InventoryStore* store, size_t new_capacity);

Product** inventory_store_get_all_active(const InventoryStore* store, size_t* out_count);
void inventory_store_free_product_array(Product** array, size_t count);

unsigned long hash_function(const char* str, size_t bucket_count);
int inventory_store_restore_deleted(InventoryStore* store, const char* product_id, int old_is_deleted);

#endif
