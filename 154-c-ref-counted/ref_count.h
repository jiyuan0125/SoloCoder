#ifndef REF_COUNT_H
#define REF_COUNT_H

#include "shared_ptr.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct sp_atomic_int {
    volatile size_t         value;
    pthread_mutex_t         mutex;
} sp_atomic_int_t;

int sp_atomic_init(sp_atomic_int_t *ai, size_t init_val);
void sp_atomic_destroy(sp_atomic_int_t *ai);
size_t sp_atomic_load(const sp_atomic_int_t *ai);
size_t sp_atomic_fetch_add(sp_atomic_int_t *ai, size_t val);
size_t sp_atomic_fetch_sub(sp_atomic_int_t *ai, size_t val);

int sp_ref_count_init(sp_control_block_t *cb);
void sp_ref_count_destroy(sp_control_block_t *cb);
size_t sp_ref_count_inc(sp_control_block_t *cb);
size_t sp_ref_count_dec(sp_control_block_t *cb);
size_t sp_ref_count_get(const sp_control_block_t *cb);

int sp_weak_count_init(sp_control_block_t *cb);
void sp_weak_count_destroy(sp_control_block_t *cb);
size_t sp_weak_count_inc(sp_control_block_t *cb);
size_t sp_weak_count_dec(sp_control_block_t *cb);
size_t sp_weak_count_get(const sp_control_block_t *cb);

int sp_control_block_init(sp_control_block_t *cb, void *ptr,
                           sp_destructor_fn destructor,
                           sp_object_type_t type, size_t obj_size);
void sp_control_block_cleanup(sp_control_block_t *cb);
int sp_control_block_try_destroy(sp_control_block_t *cb);

#ifdef __cplusplus
}
#endif

#endif
