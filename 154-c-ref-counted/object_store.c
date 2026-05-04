#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <stdio.h>
#include "object_store.h"
#include "ref_count.h"

static size_t sp_hash_func(const char *key, size_t bucket_count)
{
    size_t hash = 5381;
    int c;
    while ((c = *key++) != 0) {
        hash = ((hash << 5) + hash) + (unsigned char)c;
    }
    return hash % bucket_count;
}

int sp_store_init(sp_object_store_t *store, size_t bucket_count)
{
    size_t i;

    if (store == NULL) {
        errno = EINVAL;
        return -1;
    }

    if (bucket_count == 0) {
        bucket_count = SP_STORE_DEFAULT_BUCKETS;
    }

    memset(store, 0, sizeof(sp_object_store_t));

    store->buckets = (sp_store_entry_t **)malloc(
        bucket_count * sizeof(sp_store_entry_t *));
    if (store->buckets == NULL) {
        return -1;
    }

    for (i = 0; i < bucket_count; i++) {
        store->buckets[i] = NULL;
    }

    store->bucket_count = bucket_count;
    store->entry_count = 0;

    if (pthread_mutex_init(&store->mutex, NULL) != 0) {
        free(store->buckets);
        store->buckets = NULL;
        return -1;
    }

    store->is_initialized = 1;
    return 0;
}

static void sp_store_free_entry(sp_store_entry_t *entry)
{
    if (entry == NULL) {
        return;
    }
    free(entry);
}

void sp_store_destroy(sp_object_store_t *store)
{
    size_t i;
    sp_store_entry_t *entry, *next;

    if (store == NULL || !store->is_initialized) {
        return;
    }

    pthread_mutex_lock(&store->mutex);

    for (i = 0; i < store->bucket_count; i++) {
        entry = store->buckets[i];
        while (entry != NULL) {
            next = entry->next;
            sp_store_free_entry(entry);
            entry = next;
        }
    }

    free(store->buckets);
    store->buckets = NULL;

    pthread_mutex_unlock(&store->mutex);
    pthread_mutex_destroy(&store->mutex);

    memset(store, 0, sizeof(sp_object_store_t));
}

int sp_store_clear(sp_object_store_t *store)
{
    size_t i;
    sp_store_entry_t *entry, *next;

    if (store == NULL || !store->is_initialized) {
        errno = EINVAL;
        return -1;
    }

    pthread_mutex_lock(&store->mutex);

    for (i = 0; i < store->bucket_count; i++) {
        entry = store->buckets[i];
        while (entry != NULL) {
            next = entry->next;
            sp_store_free_entry(entry);
            entry = next;
        }
        store->buckets[i] = NULL;
    }

    store->entry_count = 0;
    pthread_mutex_unlock(&store->mutex);

    return 0;
}

static sp_store_entry_t *sp_store_find_entry_locked(
    sp_object_store_t *store, const char *key, size_t *out_bucket)
{
    size_t bucket;
    sp_store_entry_t *entry;

    if (key == NULL || key[0] == '\0') {
        return NULL;
    }

    bucket = sp_hash_func(key, store->bucket_count);
    if (out_bucket != NULL) {
        *out_bucket = bucket;
    }

    entry = store->buckets[bucket];
    while (entry != NULL) {
        if (strcmp(entry->key, key) == 0) {
            return entry;
        }
        entry = entry->next;
    }

    return NULL;
}

int sp_store_insert(sp_object_store_t *store, const char *key,
                    void *ptr, sp_destructor_fn destructor,
                    sp_object_type_t type, size_t obj_size)
{
    size_t bucket;
    sp_store_entry_t *entry;
    sp_control_block_t *cb = NULL;

    if (store == NULL || !store->is_initialized ||
        key == NULL || key[0] == '\0' || ptr == NULL) {
        errno = EINVAL;
        return -1;
    }

    if (strlen(key) >= SP_STORE_MAX_KEY_LEN) {
        errno = ENAMETOOLONG;
        return -1;
    }

    if (sp_control_block_create(&cb, ptr, destructor, type, obj_size) != 0) {
        return -1;
    }

    pthread_mutex_lock(&store->mutex);

    if (sp_store_find_entry_locked(store, key, &bucket) != NULL) {
        sp_control_block_release(cb);
        pthread_mutex_unlock(&store->mutex);
        errno = EEXIST;
        return -1;
    }

    entry = (sp_store_entry_t *)malloc(sizeof(sp_store_entry_t));
    if (entry == NULL) {
        sp_control_block_release(cb);
        pthread_mutex_unlock(&store->mutex);
        return -1;
    }

    strncpy(entry->key, key, SP_STORE_MAX_KEY_LEN - 1);
    entry->key[SP_STORE_MAX_KEY_LEN - 1] = '\0';
    entry->cb = cb;
    entry->next = store->buckets[bucket];

    store->buckets[bucket] = entry;
    store->entry_count++;

    pthread_mutex_unlock(&store->mutex);
    return 0;
}

int sp_store_get(sp_object_store_t *store, const char *key,
                 sp_shared_ptr_t *out_ptr)
{
    sp_store_entry_t *entry;

    if (store == NULL || !store->is_initialized ||
        key == NULL || key[0] == '\0' || out_ptr == NULL) {
        errno = EINVAL;
        return -1;
    }

    sp_shared_ptr_init(out_ptr);

    pthread_mutex_lock(&store->mutex);

    entry = sp_store_find_entry_locked(store, key, NULL);
    if (entry == NULL) {
        pthread_mutex_unlock(&store->mutex);
        errno = ENOENT;
        return -1;
    }

    if (!sp_control_block_is_valid(entry->cb)) {
        pthread_mutex_unlock(&store->mutex);
        errno = ENOENT;
        return -1;
    }

    sp_control_block_add_ref(entry->cb);
    out_ptr->sp_cb = entry->cb;

    pthread_mutex_unlock(&store->mutex);
    return 0;
}

int sp_store_get_or_create(sp_object_store_t *store, const char *key,
                            sp_factory_fn factory, void *factory_data,
                            sp_destructor_fn destructor,
                            sp_object_type_t type,
                            sp_shared_ptr_t *out_ptr)
{
    sp_store_entry_t *entry;
    void *new_ptr = NULL;
    size_t obj_size = 0;
    sp_control_block_t *cb = NULL;
    size_t bucket;

    if (store == NULL || !store->is_initialized ||
        key == NULL || key[0] == '\0' || factory == NULL || out_ptr == NULL) {
        errno = EINVAL;
        return -1;
    }

    if (strlen(key) >= SP_STORE_MAX_KEY_LEN) {
        errno = ENAMETOOLONG;
        return -1;
    }

    sp_shared_ptr_init(out_ptr);

    pthread_mutex_lock(&store->mutex);

    entry = sp_store_find_entry_locked(store, key, &bucket);
    if (entry != NULL && sp_control_block_is_valid(entry->cb)) {
        sp_control_block_add_ref(entry->cb);
        out_ptr->sp_cb = entry->cb;
        pthread_mutex_unlock(&store->mutex);
        return 0;
    }

    if (entry != NULL) {
        sp_store_entry_t **prev = &store->buckets[bucket];
        sp_store_entry_t *e = store->buckets[bucket];
        while (e != NULL) {
            if (e == entry) {
                *prev = e->next;
                sp_store_free_entry(e);
                store->entry_count--;
                break;
            }
            prev = &e->next;
            e = e->next;
        }
    }

    pthread_mutex_unlock(&store->mutex);

    new_ptr = factory(key, &obj_size, factory_data);
    if (new_ptr == NULL) {
        errno = ENOMEM;
        return -1;
    }

    if (sp_control_block_create(&cb, new_ptr, destructor, type, obj_size) != 0) {
        if (destructor != NULL) {
            destructor(new_ptr);
        } else {
            free(new_ptr);
        }
        return -1;
    }

    pthread_mutex_lock(&store->mutex);

    entry = sp_store_find_entry_locked(store, key, &bucket);
    if (entry != NULL && sp_control_block_is_valid(entry->cb)) {
        sp_control_block_release(cb);
        sp_control_block_add_ref(entry->cb);
        out_ptr->sp_cb = entry->cb;
        pthread_mutex_unlock(&store->mutex);
        return 0;
    }

    entry = (sp_store_entry_t *)malloc(sizeof(sp_store_entry_t));
    if (entry == NULL) {
        sp_control_block_release(cb);
        pthread_mutex_unlock(&store->mutex);
        return -1;
    }

    strncpy(entry->key, key, SP_STORE_MAX_KEY_LEN - 1);
    entry->key[SP_STORE_MAX_KEY_LEN - 1] = '\0';
    entry->cb = cb;
    entry->next = store->buckets[bucket];

    store->buckets[bucket] = entry;
    store->entry_count++;

    out_ptr->sp_cb = cb;

    pthread_mutex_unlock(&store->mutex);
    return 0;
}

int sp_store_remove(sp_object_store_t *store, const char *key)
{
    size_t bucket;
    sp_store_entry_t *entry, **prev;

    if (store == NULL || !store->is_initialized ||
        key == NULL || key[0] == '\0') {
        errno = EINVAL;
        return -1;
    }

    pthread_mutex_lock(&store->mutex);

    bucket = sp_hash_func(key, store->bucket_count);
    prev = &store->buckets[bucket];
    entry = store->buckets[bucket];

    while (entry != NULL) {
        if (strcmp(entry->key, key) == 0) {
            *prev = entry->next;
            sp_store_free_entry(entry);
            store->entry_count--;
            pthread_mutex_unlock(&store->mutex);
            return 0;
        }
        prev = &entry->next;
        entry = entry->next;
    }

    pthread_mutex_unlock(&store->mutex);
    errno = ENOENT;
    return -1;
}

int sp_store_contains(sp_object_store_t *store, const char *key)
{
    sp_store_entry_t *entry;
    int result;

    if (store == NULL || !store->is_initialized ||
        key == NULL || key[0] == '\0') {
        return 0;
    }

    pthread_mutex_lock(&store->mutex);

    entry = sp_store_find_entry_locked(store, key, NULL);
    result = (entry != NULL && sp_control_block_is_valid(entry->cb));

    pthread_mutex_unlock(&store->mutex);

    return result;
}

size_t sp_store_size(const sp_object_store_t *store)
{
    size_t count;

    if (store == NULL || !store->is_initialized) {
        return 0;
    }

    pthread_mutex_lock((pthread_mutex_t *)&store->mutex);
    count = store->entry_count;
    pthread_mutex_unlock((pthread_mutex_t *)&store->mutex);

    return count;
}

size_t sp_store_bucket_count(const sp_object_store_t *store)
{
    if (store == NULL || !store->is_initialized) {
        return 0;
    }
    return store->bucket_count;
}

void sp_store_foreach(sp_object_store_t *store,
                       sp_store_visitor_fn visitor,
                       void *user_data)
{
    size_t i;
    sp_store_entry_t *entry;

    if (store == NULL || !store->is_initialized || visitor == NULL) {
        return;
    }

    pthread_mutex_lock(&store->mutex);

    for (i = 0; i < store->bucket_count; i++) {
        entry = store->buckets[i];
        while (entry != NULL) {
            visitor(entry->key, entry->cb, user_data);
            entry = entry->next;
        }
    }

    pthread_mutex_unlock(&store->mutex);
}

int sp_store_get_stats(sp_object_store_t *store,
                        size_t *out_total_objects,
                        size_t *out_total_refs,
                        size_t *out_buckets_used)
{
    size_t i, buckets_used = 0, total_refs = 0;
    sp_store_entry_t *entry;

    if (store == NULL || !store->is_initialized) {
        errno = EINVAL;
        return -1;
    }

    pthread_mutex_lock(&store->mutex);

    for (i = 0; i < store->bucket_count; i++) {
        entry = store->buckets[i];
        if (entry != NULL) {
            buckets_used++;
        }
        while (entry != NULL) {
            total_refs += entry->cb->cb_ref_count;
            entry = entry->next;
        }
    }

    if (out_total_objects != NULL) {
        *out_total_objects = store->entry_count;
    }
    if (out_total_refs != NULL) {
        *out_total_refs = total_refs;
    }
    if (out_buckets_used != NULL) {
        *out_buckets_used = buckets_used;
    }

    pthread_mutex_unlock(&store->mutex);
    return 0;
}

static sp_object_store_t g_global_store;
static void sp_init_global_store(void)
{
    sp_store_init(&g_global_store, SP_STORE_DEFAULT_BUCKETS);
}

sp_object_store_t *sp_store_global_instance(void)
{
    static pthread_once_t once_control = PTHREAD_ONCE_INIT;

    pthread_once(&once_control, sp_init_global_store);

    return &g_global_store;
}
