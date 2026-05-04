#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include "ref_count.h"

int sp_atomic_init(sp_atomic_int_t *ai, size_t init_val)
{
    if (ai == NULL) {
        errno = EINVAL;
        return -1;
    }
    ai->value = init_val;
    if (pthread_mutex_init(&ai->mutex, NULL) != 0) {
        return -1;
    }
    return 0;
}

void sp_atomic_destroy(sp_atomic_int_t *ai)
{
    if (ai != NULL) {
        pthread_mutex_destroy(&ai->mutex);
    }
}

size_t sp_atomic_load(const sp_atomic_int_t *ai)
{
    size_t val;
    pthread_mutex_lock((pthread_mutex_t *)&ai->mutex);
    val = ai->value;
    pthread_mutex_unlock((pthread_mutex_t *)&ai->mutex);
    return val;
}

size_t sp_atomic_fetch_add(sp_atomic_int_t *ai, size_t val)
{
    size_t old_val;
    pthread_mutex_lock(&ai->mutex);
    old_val = ai->value;
    ai->value += val;
    pthread_mutex_unlock(&ai->mutex);
    return old_val;
}

size_t sp_atomic_fetch_sub(sp_atomic_int_t *ai, size_t val)
{
    size_t old_val;
    pthread_mutex_lock(&ai->mutex);
    old_val = ai->value;
    if (ai->value >= val) {
        ai->value -= val;
    } else {
        ai->value = 0;
    }
    pthread_mutex_unlock(&ai->mutex);
    return old_val;
}

int sp_control_block_init(sp_control_block_t *cb, void *ptr,
                           sp_destructor_fn destructor,
                           sp_object_type_t type, size_t obj_size)
{
    if (cb == NULL) {
        errno = EINVAL;
        return -1;
    }

    memset(cb, 0, sizeof(sp_control_block_t));
    cb->cb_ptr = ptr;
    cb->cb_destructor = destructor;
    cb->cb_ref_count = 1;
    cb->cb_weak_count = 0;
    cb->cb_is_destroyed = 0;
    cb->cb_obj_type = type;
    cb->cb_obj_size = obj_size;

    if (pthread_mutex_init(&cb->cb_mutex, NULL) != 0) {
        return -1;
    }

    return 0;
}

void sp_control_block_cleanup(sp_control_block_t *cb)
{
    if (cb == NULL) {
        return;
    }

    pthread_mutex_destroy(&cb->cb_mutex);
}

int sp_control_block_try_destroy(sp_control_block_t *cb)
{
    if (cb == NULL) {
        return -1;
    }

    pthread_mutex_lock(&cb->cb_mutex);

    if (cb->cb_ref_count == 0 && !cb->cb_is_destroyed) {
        cb->cb_is_destroyed = 1;
        if (cb->cb_destructor != NULL && cb->cb_ptr != NULL) {
            pthread_mutex_unlock(&cb->cb_mutex);
            cb->cb_destructor(cb->cb_ptr);
            pthread_mutex_lock(&cb->cb_mutex);
        }
    }

    int should_free = (cb->cb_ref_count == 0 && cb->cb_weak_count == 0);
    pthread_mutex_unlock(&cb->cb_mutex);

    if (should_free) {
        sp_control_block_cleanup(cb);
        free(cb);
    }

    return 0;
}

int sp_control_block_create(sp_control_block_t **cb, void *ptr,
                             sp_destructor_fn destructor,
                             sp_object_type_t type, size_t obj_size)
{
    if (cb == NULL) {
        errno = EINVAL;
        return -1;
    }

    sp_control_block_t *new_cb = (sp_control_block_t *)malloc(sizeof(sp_control_block_t));
    if (new_cb == NULL) {
        return -1;
    }

    if (sp_control_block_init(new_cb, ptr, destructor, type, obj_size) != 0) {
        free(new_cb);
        return -1;
    }

    *cb = new_cb;
    return 0;
}

void sp_control_block_add_ref(sp_control_block_t *cb)
{
    if (cb == NULL) {
        return;
    }

    pthread_mutex_lock(&cb->cb_mutex);
    if (!cb->cb_is_destroyed) {
        cb->cb_ref_count++;
    }
    pthread_mutex_unlock(&cb->cb_mutex);
}

void sp_control_block_release(sp_control_block_t *cb)
{
    if (cb == NULL) {
        return;
    }

    pthread_mutex_lock(&cb->cb_mutex);
    if (cb->cb_ref_count > 0) {
        cb->cb_ref_count--;
    }
    pthread_mutex_unlock(&cb->cb_mutex);

    sp_control_block_try_destroy(cb);
}

void sp_control_block_add_weak_ref(sp_control_block_t *cb)
{
    if (cb == NULL) {
        return;
    }

    pthread_mutex_lock(&cb->cb_mutex);
    cb->cb_weak_count++;
    pthread_mutex_unlock(&cb->cb_mutex);
}

void sp_control_block_release_weak(sp_control_block_t *cb)
{
    if (cb == NULL) {
        return;
    }

    pthread_mutex_lock(&cb->cb_mutex);
    if (cb->cb_weak_count > 0) {
        cb->cb_weak_count--;
    }
    pthread_mutex_unlock(&cb->cb_mutex);

    sp_control_block_try_destroy(cb);
}

int sp_control_block_is_valid(const sp_control_block_t *cb)
{
    if (cb == NULL) {
        return 0;
    }

    pthread_mutex_lock((pthread_mutex_t *)&cb->cb_mutex);
    int valid = (cb->cb_ptr != NULL && !cb->cb_is_destroyed && cb->cb_ref_count > 0);
    pthread_mutex_unlock((pthread_mutex_t *)&cb->cb_mutex);

    return valid;
}

int sp_shared_ptr_init(sp_shared_ptr_t *ptr)
{
    if (ptr == NULL) {
        errno = EINVAL;
        return -1;
    }
    ptr->sp_cb = NULL;
    return 0;
}

void sp_shared_ptr_reset(sp_shared_ptr_t *ptr)
{
    if (ptr == NULL) {
        return;
    }

    if (ptr->sp_cb != NULL) {
        sp_control_block_release(ptr->sp_cb);
        ptr->sp_cb = NULL;
    }
}

int sp_shared_ptr_copy(sp_shared_ptr_t *dest, const sp_shared_ptr_t *src)
{
    if (dest == NULL || src == NULL) {
        errno = EINVAL;
        return -1;
    }

    if (dest == src) {
        return 0;
    }

    if (dest->sp_cb == src->sp_cb) {
        return 0;
    }

    sp_shared_ptr_reset(dest);

    if (src->sp_cb != NULL) {
        sp_control_block_add_ref(src->sp_cb);
        dest->sp_cb = src->sp_cb;
    }

    return 0;
}

int sp_shared_ptr_move(sp_shared_ptr_t *dest, sp_shared_ptr_t *src)
{
    if (dest == NULL || src == NULL) {
        errno = EINVAL;
        return -1;
    }

    if (dest == src) {
        return 0;
    }

    sp_shared_ptr_reset(dest);
    dest->sp_cb = src->sp_cb;
    src->sp_cb = NULL;

    return 0;
}

size_t sp_shared_ptr_use_count(const sp_shared_ptr_t *ptr)
{
    if (ptr == NULL || ptr->sp_cb == NULL) {
        return 0;
    }

    pthread_mutex_lock(&ptr->sp_cb->cb_mutex);
    size_t count = ptr->sp_cb->cb_ref_count;
    pthread_mutex_unlock(&ptr->sp_cb->cb_mutex);

    return count;
}

int sp_shared_ptr_unique(const sp_shared_ptr_t *ptr)
{
    return (sp_shared_ptr_use_count(ptr) == 1);
}

int sp_shared_ptr_expired(const sp_shared_ptr_t *ptr)
{
    if (ptr == NULL || ptr->sp_cb == NULL) {
        return 1;
    }

    pthread_mutex_lock(&ptr->sp_cb->cb_mutex);
    int expired = (ptr->sp_cb->cb_ref_count == 0 || ptr->sp_cb->cb_is_destroyed);
    pthread_mutex_unlock(&ptr->sp_cb->cb_mutex);

    return expired;
}

void *sp_shared_ptr_get(const sp_shared_ptr_t *ptr)
{
    if (ptr == NULL || ptr->sp_cb == NULL) {
        return NULL;
    }

    pthread_mutex_lock(&ptr->sp_cb->cb_mutex);
    if (ptr->sp_cb->cb_is_destroyed) {
        pthread_mutex_unlock(&ptr->sp_cb->cb_mutex);
        return NULL;
    }
    void *result = ptr->sp_cb->cb_ptr;
    pthread_mutex_unlock(&ptr->sp_cb->cb_mutex);

    return result;
}

int sp_weak_ptr_init(sp_weak_ptr_t *ptr)
{
    if (ptr == NULL) {
        errno = EINVAL;
        return -1;
    }
    ptr->wp_cb = NULL;
    return 0;
}

void sp_weak_ptr_reset(sp_weak_ptr_t *ptr)
{
    if (ptr == NULL) {
        return;
    }

    if (ptr->wp_cb != NULL) {
        sp_control_block_release_weak(ptr->wp_cb);
        ptr->wp_cb = NULL;
    }
}

int sp_weak_ptr_copy(sp_weak_ptr_t *dest, const sp_weak_ptr_t *src)
{
    if (dest == NULL || src == NULL) {
        errno = EINVAL;
        return -1;
    }

    if (dest == src) {
        return 0;
    }

    if (dest->wp_cb == src->wp_cb) {
        return 0;
    }

    sp_weak_ptr_reset(dest);

    if (src->wp_cb != NULL) {
        sp_control_block_add_weak_ref(src->wp_cb);
        dest->wp_cb = src->wp_cb;
    }

    return 0;
}

int sp_weak_ptr_from_shared(sp_weak_ptr_t *dest, const sp_shared_ptr_t *src)
{
    if (dest == NULL || src == NULL) {
        errno = EINVAL;
        return -1;
    }

    sp_weak_ptr_reset(dest);

    if (src->sp_cb != NULL) {
        sp_control_block_add_weak_ref(src->sp_cb);
        dest->wp_cb = src->sp_cb;
    }

    return 0;
}

int sp_weak_ptr_lock(sp_shared_ptr_t *dest, const sp_weak_ptr_t *src)
{
    if (dest == NULL || src == NULL) {
        errno = EINVAL;
        return -1;
    }

    sp_shared_ptr_reset(dest);

    if (src->wp_cb == NULL) {
        return 0;
    }

    pthread_mutex_lock(&src->wp_cb->cb_mutex);
    if (src->wp_cb->cb_is_destroyed || src->wp_cb->cb_ref_count == 0) {
        pthread_mutex_unlock(&src->wp_cb->cb_mutex);
        return 0;
    }
    src->wp_cb->cb_ref_count++;
    dest->sp_cb = src->wp_cb;
    pthread_mutex_unlock(&src->wp_cb->cb_mutex);

    return 1;
}

size_t sp_weak_ptr_use_count(const sp_weak_ptr_t *ptr)
{
    if (ptr == NULL || ptr->wp_cb == NULL) {
        return 0;
    }

    pthread_mutex_lock(&ptr->wp_cb->cb_mutex);
    size_t count = ptr->wp_cb->cb_ref_count;
    pthread_mutex_unlock(&ptr->wp_cb->cb_mutex);

    return count;
}

int sp_weak_ptr_expired(const sp_weak_ptr_t *ptr)
{
    if (ptr == NULL || ptr->wp_cb == NULL) {
        return 1;
    }

    pthread_mutex_lock(&ptr->wp_cb->cb_mutex);
    int expired = (ptr->wp_cb->cb_ref_count == 0 || ptr->wp_cb->cb_is_destroyed);
    pthread_mutex_unlock(&ptr->wp_cb->cb_mutex);

    return expired;
}
