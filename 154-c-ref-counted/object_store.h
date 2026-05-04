#ifndef OBJECT_STORE_H
#define OBJECT_STORE_H

#include "shared_ptr.h"

#ifdef __cplusplus
extern "C" {
#endif

#define SP_STORE_DEFAULT_BUCKETS  64
#define SP_STORE_MAX_KEY_LEN      256

typedef struct sp_store_entry {
    char                     key[SP_STORE_MAX_KEY_LEN];
    sp_control_block_t       *cb;
    struct sp_store_entry    *next;
} sp_store_entry_t;

typedef struct sp_object_store {
    sp_store_entry_t         **buckets;
    size_t                    bucket_count;
    size_t                    entry_count;
    pthread_mutex_t           mutex;
    int                       is_initialized;
} sp_object_store_t;

typedef void *(*sp_factory_fn)(const char *key, size_t *out_size, void *user_data);

int sp_store_init(sp_object_store_t *store, size_t bucket_count);
void sp_store_destroy(sp_object_store_t *store);
int sp_store_clear(sp_object_store_t *store);

int sp_store_insert(sp_object_store_t *store, const char *key,
                    void *ptr, sp_destructor_fn destructor,
                    sp_object_type_t type, size_t obj_size);

int sp_store_get(sp_object_store_t *store, const char *key,
                 sp_shared_ptr_t *out_ptr);

int sp_store_get_or_create(sp_object_store_t *store, const char *key,
                            sp_factory_fn factory, void *factory_data,
                            sp_destructor_fn destructor,
                            sp_object_type_t type,
                            sp_shared_ptr_t *out_ptr);

int sp_store_remove(sp_object_store_t *store, const char *key);
int sp_store_contains(sp_object_store_t *store, const char *key);

size_t sp_store_size(const sp_object_store_t *store);
size_t sp_store_bucket_count(const sp_object_store_t *store);

typedef void (*sp_store_visitor_fn)(const char *key,
                                      const sp_control_block_t *cb,
                                      void *user_data);

void sp_store_foreach(sp_object_store_t *store,
                       sp_store_visitor_fn visitor,
                       void *user_data);

int sp_store_get_stats(sp_object_store_t *store,
                        size_t *out_total_objects,
                        size_t *out_total_refs,
                        size_t *out_buckets_used);

sp_object_store_t *sp_store_global_instance(void);

#ifdef __cplusplus
}
#endif

#endif
