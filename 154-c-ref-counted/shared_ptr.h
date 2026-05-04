#ifndef SHARED_PTR_H
#define SHARED_PTR_H

#include <stddef.h>
#include <pthread.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef void (*sp_destructor_fn)(void *ptr);

typedef enum {
    SP_OBJ_TYPE_IMAGE,
    SP_OBJ_TYPE_TEMPLATE,
    SP_OBJ_TYPE_GENERIC,
    SP_OBJ_TYPE_MAX
} sp_object_type_t;

typedef struct sp_control_block {
    void                    *cb_ptr;
    sp_destructor_fn        cb_destructor;
    size_t                  cb_ref_count;
    size_t                  cb_weak_count;
    pthread_mutex_t         cb_mutex;
    int                     cb_is_destroyed;
    sp_object_type_t        cb_obj_type;
    size_t                  cb_obj_size;
} sp_control_block_t;

typedef struct sp_shared_ptr {
    sp_control_block_t      *sp_cb;
} sp_shared_ptr_t;

typedef struct sp_weak_ptr {
    sp_control_block_t      *wp_cb;
} sp_weak_ptr_t;

#define SP_SHARED_PTR_INIT   { NULL }
#define SP_WEAK_PTR_INIT     { NULL }

int sp_shared_ptr_init(sp_shared_ptr_t *ptr);
void sp_shared_ptr_reset(sp_shared_ptr_t *ptr);
int sp_shared_ptr_copy(sp_shared_ptr_t *dest, const sp_shared_ptr_t *src);
int sp_shared_ptr_move(sp_shared_ptr_t *dest, sp_shared_ptr_t *src);
size_t sp_shared_ptr_use_count(const sp_shared_ptr_t *ptr);
int sp_shared_ptr_unique(const sp_shared_ptr_t *ptr);
int sp_shared_ptr_expired(const sp_shared_ptr_t *ptr);
void *sp_shared_ptr_get(const sp_shared_ptr_t *ptr);

int sp_weak_ptr_init(sp_weak_ptr_t *ptr);
void sp_weak_ptr_reset(sp_weak_ptr_t *ptr);
int sp_weak_ptr_copy(sp_weak_ptr_t *dest, const sp_weak_ptr_t *src);
int sp_weak_ptr_from_shared(sp_weak_ptr_t *dest, const sp_shared_ptr_t *src);
int sp_weak_ptr_lock(sp_shared_ptr_t *dest, const sp_weak_ptr_t *src);
size_t sp_weak_ptr_use_count(const sp_weak_ptr_t *ptr);
int sp_weak_ptr_expired(const sp_weak_ptr_t *ptr);

int sp_control_block_create(sp_control_block_t **cb, void *ptr,
                             sp_destructor_fn destructor,
                             sp_object_type_t type, size_t obj_size);
void sp_control_block_add_ref(sp_control_block_t *cb);
void sp_control_block_release(sp_control_block_t *cb);
void sp_control_block_add_weak_ref(sp_control_block_t *cb);
void sp_control_block_release_weak(sp_control_block_t *cb);
int sp_control_block_is_valid(const sp_control_block_t *cb);

#ifdef __cplusplus
}
#endif

#endif
