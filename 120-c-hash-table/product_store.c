#include "product_store.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

#define LOAD_FACTOR_THRESHOLD  0.7
#define INITIAL_CAPACITY        16

unsigned long hash_function(const char* str, size_t bucket_count) {
    unsigned long hash = 5381;
    int c;
    while ((c = (unsigned char)*str++)) {
        hash = ((hash << 5) + hash) + tolower((unsigned char)c);
    }
    return hash % bucket_count;
}

void str_tolower(char* dest, const char* src, size_t max_len) {
    size_t i = 0;
    while (src[i] != '\0' && i < max_len - 1) {
        dest[i] = (char)tolower((unsigned char)src[i]);
        i++;
    }
    dest[i] = '\0';
}

int strcasecmp_custom(const char* s1, const char* s2) {
    while (*s1 && *s2) {
        int c1 = tolower((unsigned char)*s1);
        int c2 = tolower((unsigned char)*s2);
        if (c1 != c2) {
            return c1 - c2;
        }
        s1++;
        s2++;
    }
    return tolower((unsigned char)*s1) - tolower((unsigned char)*s2);
}

static HashNode* hash_node_create(const Product* product) {
    HashNode* node = (HashNode*)malloc(sizeof(HashNode));
    if (!node) return NULL;
    
    memcpy(&node->product, product, sizeof(Product));
    node->next = NULL;
    return node;
}

static void hash_node_destroy(HashNode* node) {
    free(node);
}

InventoryStore* inventory_store_create(size_t initial_capacity) {
    if (initial_capacity == 0) {
        initial_capacity = INITIAL_CAPACITY;
    }
    
    InventoryStore* store = (InventoryStore*)malloc(sizeof(InventoryStore));
    if (!store) return NULL;
    
    store->buckets = (HashNode**)calloc(initial_capacity, sizeof(HashNode*));
    if (!store->buckets) {
        free(store);
        return NULL;
    }
    
    store->bucket_count = initial_capacity;
    store->size = 0;
    store->deleted_count = 0;
    
    return store;
}

void inventory_store_destroy(InventoryStore* store) {
    if (!store) return;
    
    for (size_t i = 0; i < store->bucket_count; i++) {
        HashNode* curr = store->buckets[i];
        while (curr) {
            HashNode* temp = curr;
            curr = curr->next;
            hash_node_destroy(temp);
        }
    }
    
    free(store->buckets);
    free(store);
}

static int inventory_store_needs_resize(const InventoryStore* store) {
    double load_factor = (double)(store->size + store->deleted_count) / store->bucket_count;
    return load_factor > LOAD_FACTOR_THRESHOLD;
}

static int inventory_store_rehash(InventoryStore* store, size_t new_capacity) {
    HashNode** new_buckets = (HashNode**)calloc(new_capacity, sizeof(HashNode*));
    if (!new_buckets) {
        return INVENTORY_STORE_ERR_MEMORY;
    }
    
    for (size_t i = 0; i < store->bucket_count; i++) {
        HashNode* curr = store->buckets[i];
        while (curr) {
            HashNode* next = curr->next;
            
            unsigned long new_idx = hash_function(curr->product.id, new_capacity);
            curr->next = new_buckets[new_idx];
            new_buckets[new_idx] = curr;
            
            curr = next;
        }
    }
    
    free(store->buckets);
    store->buckets = new_buckets;
    store->bucket_count = new_capacity;
    store->deleted_count = 0;
    
    return INVENTORY_STORE_OK;
}

int inventory_store_reserve(InventoryStore* store, size_t new_capacity) {
    if (!store) return INVENTORY_STORE_ERR_NULL;
    
    if (new_capacity <= store->bucket_count) {
        return INVENTORY_STORE_OK;
    }
    
    size_t actual_capacity = 1;
    while (actual_capacity < new_capacity) {
        actual_capacity *= 2;
    }
    
    return inventory_store_rehash(store, actual_capacity);
}

int inventory_store_add(InventoryStore* store, const Product* product) {
    if (!store || !product) {
        return INVENTORY_STORE_ERR_NULL;
    }
    
    if (product->id[0] == '\0') {
        return INVENTORY_STORE_ERR_NULL;
    }
    
    if (inventory_store_needs_resize(store)) {
        int ret = inventory_store_rehash(store, store->bucket_count * 2);
        if (ret != INVENTORY_STORE_OK) {
            return ret;
        }
    }
    
    Product* existing = inventory_store_find(store, product->id);
    if (existing && !existing->is_deleted) {
        return INVENTORY_STORE_ERR_DUPLICATE;
    }
    
    unsigned long idx = hash_function(product->id, store->bucket_count);
    
    HashNode* new_node = hash_node_create(product);
    if (!new_node) {
        return INVENTORY_STORE_ERR_MEMORY;
    }
    
    new_node->next = store->buckets[idx];
    store->buckets[idx] = new_node;
    store->size++;
    
    return INVENTORY_STORE_OK;
}

Product* inventory_store_find(const InventoryStore* store, const char* product_id) {
    if (!store || !product_id) {
        return NULL;
    }
    
    unsigned long idx = hash_function(product_id, store->bucket_count);
    
    HashNode* curr = store->buckets[idx];
    while (curr) {
        if (strcasecmp_custom(curr->product.id, product_id) == 0) {
            if (curr->product.is_deleted) {
                return NULL;
            }
            return &curr->product;
        }
        curr = curr->next;
    }
    
    return NULL;
}

int inventory_store_update_quantity(InventoryStore* store, const char* product_id, int delta) {
    if (!store || !product_id) {
        return INVENTORY_STORE_ERR_NULL;
    }
    
    Product* product = inventory_store_find(store, product_id);
    if (!product) {
        return INVENTORY_STORE_ERR_NOT_FOUND;
    }
    
    int new_quantity = product->quantity + delta;
    if (new_quantity < 0) {
        return INVENTORY_STORE_ERR_NEGATIVE;
    }
    
    product->quantity = new_quantity;
    return INVENTORY_STORE_OK;
}

int inventory_store_mark_deleted(InventoryStore* store, const char* product_id) {
    if (!store || !product_id) {
        return INVENTORY_STORE_ERR_NULL;
    }
    
    unsigned long idx = hash_function(product_id, store->bucket_count);
    
    HashNode* curr = store->buckets[idx];
    while (curr) {
        if (strcasecmp_custom(curr->product.id, product_id) == 0) {
            if (curr->product.is_deleted) {
                return INVENTORY_STORE_ERR_NOT_FOUND;
            }
            curr->product.is_deleted = 1;
            store->deleted_count++;
            return INVENTORY_STORE_OK;
        }
        curr = curr->next;
    }
    
    return INVENTORY_STORE_ERR_NOT_FOUND;
}

size_t inventory_store_size(const InventoryStore* store) {
    if (!store) return 0;
    return store->size;
}

size_t inventory_store_active_count(const InventoryStore* store) {
    if (!store) return 0;
    return store->size - store->deleted_count;
}

Product** inventory_store_get_all_active(const InventoryStore* store, size_t* out_count) {
    if (!store || !out_count) return NULL;
    
    size_t active_count = inventory_store_active_count(store);
    *out_count = 0;
    
    if (active_count == 0) {
        return NULL;
    }
    
    Product** result = (Product**)malloc(active_count * sizeof(Product*));
    if (!result) return NULL;
    
    size_t idx = 0;
    for (size_t i = 0; i < store->bucket_count && idx < active_count; i++) {
        HashNode* curr = store->buckets[i];
        while (curr) {
            if (!curr->product.is_deleted) {
                result[idx++] = &curr->product;
            }
            curr = curr->next;
        }
    }
    
    *out_count = idx;
    return result;
}

void inventory_store_free_product_array(Product** array, size_t count) {
    (void)count;
    free(array);
}

int inventory_store_restore_deleted(InventoryStore* store, const char* product_id, int old_is_deleted) {
    if (!store || !product_id) return INVENTORY_STORE_ERR_NULL;
    
    unsigned long idx = hash_function(product_id, store->bucket_count);
    
    HashNode* curr = store->buckets[idx];
    while (curr) {
        if (strcasecmp_custom(curr->product.id, product_id) == 0) {
            int was_deleted = curr->product.is_deleted;
            curr->product.is_deleted = old_is_deleted;
            
            if (was_deleted && !old_is_deleted) {
                store->deleted_count--;
            } else if (!was_deleted && old_is_deleted) {
                store->deleted_count++;
            }
            
            return INVENTORY_STORE_OK;
        }
        curr = curr->next;
    }
    
    return INVENTORY_STORE_ERR_NOT_FOUND;
}
