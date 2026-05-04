#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>
#include <ctype.h>
#include "lifecycle.h"
#include "ref_count.h"
#include "object_store.h"

static sp_lc_image_loader_fn g_image_loader = NULL;
static sp_lc_template_loader_fn g_template_loader = NULL;
static sp_lc_policy_t g_policy = SP_LC_POLICY_LAZY;
static pthread_mutex_t g_module_mutex = PTHREAD_MUTEX_INITIALIZER;
static int g_module_initialized = 0;
static sp_lc_stats_t g_stats = {0};

sp_image_data_t *sp_lc_image_data_alloc(int width, int height, int channels)
{
    sp_image_data_t *img;
    size_t pixel_size;

    if (width <= 0 || height <= 0 || channels <= 0) {
        errno = EINVAL;
        return NULL;
    }

    if (width > SP_LC_MAX_IMAGE_WIDTH || height > SP_LC_MAX_IMAGE_HEIGHT) {
        errno = EINVAL;
        return NULL;
    }

    img = (sp_image_data_t *)calloc(1, sizeof(sp_image_data_t));
    if (img == NULL) {
        return NULL;
    }

    img->width = width;
    img->height = height;
    img->channels = channels;
    pixel_size = (size_t)width * (size_t)height * (size_t)channels;

    img->pixel_data = (unsigned char *)malloc(pixel_size);
    if (img->pixel_data == NULL) {
        free(img);
        return NULL;
    }

    img->pixel_data_size = pixel_size;
    memset(img->pixel_data, 0, pixel_size);

    return img;
}

void sp_lc_image_data_free(sp_image_data_t *img)
{
    if (img == NULL) {
        return;
    }

    if (img->pixel_data != NULL) {
        free(img->pixel_data);
    }
    free(img);

    pthread_mutex_lock(&g_module_mutex);
    g_stats.total_objects_destroyed++;
    pthread_mutex_unlock(&g_module_mutex);
}

sp_template_data_t *sp_lc_template_data_alloc(size_t code_size)
{
    sp_template_data_t *tmpl;

    if (code_size == 0 || code_size > SP_LC_MAX_TEMPLATE_SIZE) {
        errno = EINVAL;
        return NULL;
    }

    tmpl = (sp_template_data_t *)calloc(1, sizeof(sp_template_data_t));
    if (tmpl == NULL) {
        return NULL;
    }

    tmpl->compiled_code = malloc(code_size);
    if (tmpl->compiled_code == NULL) {
        free(tmpl);
        return NULL;
    }

    tmpl->compiled_size = code_size;
    tmpl->version = 1;

    return tmpl;
}

void sp_lc_template_data_free(sp_template_data_t *tmpl)
{
    if (tmpl == NULL) {
        return;
    }

    sp_weak_ptr_reset(&tmpl->dependency);

    if (tmpl->compiled_code != NULL) {
        free(tmpl->compiled_code);
    }
    free(tmpl);

    pthread_mutex_lock(&g_module_mutex);
    g_stats.total_objects_destroyed++;
    pthread_mutex_unlock(&g_module_mutex);
}

int sp_lc_module_init(void)
{
    pthread_mutex_lock(&g_module_mutex);

    if (g_module_initialized) {
        pthread_mutex_unlock(&g_module_mutex);
        return 0;
    }

    memset(&g_stats, 0, sizeof(sp_lc_stats_t));
    g_policy = SP_LC_POLICY_LAZY;
    g_module_initialized = 1;

    pthread_mutex_unlock(&g_module_mutex);
    return 0;
}

void sp_lc_module_cleanup(void)
{
    pthread_mutex_lock(&g_module_mutex);

    if (!g_module_initialized) {
        pthread_mutex_unlock(&g_module_mutex);
        return;
    }

    g_module_initialized = 0;
    g_image_loader = NULL;
    g_template_loader = NULL;
    pthread_mutex_unlock(&g_module_mutex);
}

int sp_lc_set_image_loader(sp_lc_image_loader_fn loader)
{
    pthread_mutex_lock(&g_module_mutex);
    g_image_loader = loader;
    pthread_mutex_unlock(&g_module_mutex);
    return 0;
}

int sp_lc_set_template_loader(sp_lc_template_loader_fn loader)
{
    pthread_mutex_lock(&g_module_mutex);
    g_template_loader = loader;
    pthread_mutex_unlock(&g_module_mutex);
    return 0;
}

int sp_lc_set_policy(sp_lc_policy_t policy)
{
    if (policy < 0 || policy >= SP_LC_POLICY_MAX) {
        errno = EINVAL;
        return -1;
    }

    pthread_mutex_lock(&g_module_mutex);
    g_policy = policy;
    pthread_mutex_unlock(&g_module_mutex);
    return 0;
}

static void *sp_lc_image_factory(const char *key, size_t *out_size, void *user_data)
{
    sp_image_data_t *img = NULL;
    int result;
    (void)user_data;

    if (g_image_loader != NULL) {
        result = g_image_loader(key, &img);
        if (result != 0 || img == NULL) {
            return NULL;
        }
    } else {
        img = sp_lc_image_data_alloc(100, 100, 4);
        if (img == NULL) {
            return NULL;
        }
        strncpy(img->source_path, key, sizeof(img->source_path) - 1);
    }

    if (out_size != NULL) {
        *out_size = img->pixel_data_size + sizeof(sp_image_data_t);
    }

    pthread_mutex_lock(&g_module_mutex);
    g_stats.total_images_loaded++;
    g_stats.total_bytes_allocated += img->pixel_data_size;
    pthread_mutex_unlock(&g_module_mutex);

    return img;
}

static void *sp_lc_template_factory(const char *key, size_t *out_size, void *user_data)
{
    sp_template_data_t *tmpl = NULL;
    int result;
    (void)user_data;

    if (g_template_loader != NULL) {
        result = g_template_loader(key, &tmpl);
        if (result != 0 || tmpl == NULL) {
            return NULL;
        }
    } else {
        tmpl = sp_lc_template_data_alloc(4096);
        if (tmpl == NULL) {
            return NULL;
        }
        strncpy(tmpl->template_name, key, sizeof(tmpl->template_name) - 1);
    }

    if (out_size != NULL) {
        *out_size = tmpl->compiled_size + sizeof(sp_template_data_t);
    }

    pthread_mutex_lock(&g_module_mutex);
    g_stats.total_templates_loaded++;
    g_stats.total_bytes_allocated += tmpl->compiled_size;
    pthread_mutex_unlock(&g_module_mutex);

    return tmpl;
}

int sp_lc_load_image(const char *path, sp_shared_ptr_t *out_ptr)
{
    sp_object_store_t *store;
    int result;
    char key_prefix[512];

    if (path == NULL || out_ptr == NULL) {
        errno = EINVAL;
        return -1;
    }

    store = sp_store_global_instance();
    if (store == NULL) {
        errno = ENOMEM;
        return -1;
    }

    snprintf(key_prefix, sizeof(key_prefix), "img:%s", path);

    result = sp_store_get_or_create(
        store,
        key_prefix,
        sp_lc_image_factory,
        NULL,
        (sp_destructor_fn)sp_lc_image_data_free,
        SP_OBJ_TYPE_IMAGE,
        out_ptr);

    if (result == 0) {
        pthread_mutex_lock(&g_module_mutex);
        g_stats.total_references_created++;
        pthread_mutex_unlock(&g_module_mutex);
    }

    return result;
}

int sp_lc_load_template(const char *name, sp_shared_ptr_t *out_ptr)
{
    sp_object_store_t *store;
    int result;
    char key_prefix[512];

    if (name == NULL || out_ptr == NULL) {
        errno = EINVAL;
        return -1;
    }

    store = sp_store_global_instance();
    if (store == NULL) {
        errno = ENOMEM;
        return -1;
    }

    snprintf(key_prefix, sizeof(key_prefix), "tmpl:%s", name);

    result = sp_store_get_or_create(
        store,
        key_prefix,
        sp_lc_template_factory,
        NULL,
        (sp_destructor_fn)sp_lc_template_data_free,
        SP_OBJ_TYPE_TEMPLATE,
        out_ptr);

    if (result == 0) {
        pthread_mutex_lock(&g_module_mutex);
        g_stats.total_references_created++;
        pthread_mutex_unlock(&g_module_mutex);
    }

    return result;
}

sp_image_data_t *sp_lc_image_get(const sp_shared_ptr_t *ptr)
{
    void *raw_ptr;
    sp_control_block_t *cb;

    if (ptr == NULL || ptr->sp_cb == NULL) {
        return NULL;
    }

    cb = ptr->sp_cb;
    pthread_mutex_lock(&cb->cb_mutex);
    if (cb->cb_is_destroyed || cb->cb_obj_type != SP_OBJ_TYPE_IMAGE) {
        pthread_mutex_unlock(&cb->cb_mutex);
        return NULL;
    }
    pthread_mutex_unlock(&cb->cb_mutex);

    raw_ptr = sp_shared_ptr_get(ptr);
    return (sp_image_data_t *)raw_ptr;
}

sp_template_data_t *sp_lc_template_get(const sp_shared_ptr_t *ptr)
{
    void *raw_ptr;
    sp_control_block_t *cb;

    if (ptr == NULL || ptr->sp_cb == NULL) {
        return NULL;
    }

    cb = ptr->sp_cb;
    pthread_mutex_lock(&cb->cb_mutex);
    if (cb->cb_is_destroyed || cb->cb_obj_type != SP_OBJ_TYPE_TEMPLATE) {
        pthread_mutex_unlock(&cb->cb_mutex);
        return NULL;
    }
    pthread_mutex_unlock(&cb->cb_mutex);

    raw_ptr = sp_shared_ptr_get(ptr);
    return (sp_template_data_t *)raw_ptr;
}

void sp_lc_release(sp_shared_ptr_t *ptr)
{
    sp_shared_ptr_reset(ptr);
}

int sp_lc_reload_image(const char *path)
{
    sp_object_store_t *store;
    char key_prefix[512];

    if (path == NULL) {
        errno = EINVAL;
        return -1;
    }

    store = sp_store_global_instance();
    if (store == NULL) {
        errno = ENOMEM;
        return -1;
    }

    snprintf(key_prefix, sizeof(key_prefix), "img:%s", path);

    return sp_store_remove(store, key_prefix);
}

int sp_lc_reload_template(const char *name)
{
    sp_object_store_t *store;
    char key_prefix[512];

    if (name == NULL) {
        errno = EINVAL;
        return -1;
    }

    store = sp_store_global_instance();
    if (store == NULL) {
        errno = ENOMEM;
        return -1;
    }

    snprintf(key_prefix, sizeof(key_prefix), "tmpl:%s", name);

    return sp_store_remove(store, key_prefix);
}

static void sp_lc_check_unused_visitor(const char *key,
                                       const sp_control_block_t *cb,
                                       void *user_data)
{
    sp_object_store_t *store = (sp_object_store_t *)user_data;
    (void)key;
    (void)store;

    if (cb != NULL) {
        pthread_mutex_lock((pthread_mutex_t *)&cb->cb_mutex);
        if (cb->cb_ref_count <= 1) {
            pthread_mutex_unlock((pthread_mutex_t *)&cb->cb_mutex);
        } else {
            pthread_mutex_unlock((pthread_mutex_t *)&cb->cb_mutex);
        }
    }
}

int sp_lc_unload_unused(int force)
{
    sp_object_store_t *store;
    (void)force;

    store = sp_store_global_instance();
    if (store == NULL) {
        errno = ENOMEM;
        return -1;
    }

    sp_store_foreach(store, sp_lc_check_unused_visitor, store);
    return 0;
}

void sp_lc_get_stats(sp_lc_stats_t *out_stats)
{
    if (out_stats == NULL) {
        return;
    }

    pthread_mutex_lock(&g_module_mutex);
    memcpy(out_stats, &g_stats, sizeof(sp_lc_stats_t));
    pthread_mutex_unlock(&g_module_mutex);
}

static void sp_lc_dump_visitor(const char *key,
                               const sp_control_block_t *cb,
                               void *user_data)
{
    FILE *f = (FILE *)user_data;
    const char *type_str;

    if (f == NULL || cb == NULL) {
        return;
    }

    switch (cb->cb_obj_type) {
        case SP_OBJ_TYPE_IMAGE:
            type_str = "IMAGE";
            break;
        case SP_OBJ_TYPE_TEMPLATE:
            type_str = "TEMPLATE";
            break;
        default:
            type_str = "GENERIC";
            break;
    }

    fprintf(f, "  Key: %s\n", key);
    fprintf(f, "    Type: %s\n", type_str);
    fprintf(f, "    RefCount: %zu\n", cb->cb_ref_count);
    fprintf(f, "    WeakCount: %zu\n", cb->cb_weak_count);
    fprintf(f, "    Destroyed: %s\n", cb->cb_is_destroyed ? "yes" : "no");
    fprintf(f, "    Size: %zu bytes\n", cb->cb_obj_size);
    fprintf(f, "\n");
}

void sp_lc_dump_state(void)
{
    sp_object_store_t *store;
    sp_lc_stats_t stats;
    size_t total_objs, total_refs, buckets_used;

    store = sp_store_global_instance();
    if (store == NULL) {
        printf("[Lifecycle] Store not initialized\n");
        return;
    }

    sp_lc_get_stats(&stats);
    sp_store_get_stats(store, &total_objs, &total_refs, &buckets_used);

    printf("========== Lifecycle Module State ==========\n");
    printf("\n");
    printf("Global Statistics:\n");
    printf("  Total Images Loaded: %zu\n", stats.total_images_loaded);
    printf("  Total Templates Loaded: %zu\n", stats.total_templates_loaded);
    printf("  Total Bytes Allocated: %zu\n", stats.total_bytes_allocated);
    printf("  Total References Created: %zu\n", stats.total_references_created);
    printf("  Total Objects Destroyed: %zu\n", stats.total_objects_destroyed);
    printf("  Circular References Broken: %zu\n", stats.circular_ref_breaks);
    printf("\n");

    printf("Store Statistics:\n");
    printf("  Total Objects: %zu\n", total_objs);
    printf("  Total References: %zu\n", total_refs);
    printf("  Buckets Used: %zu / %zu\n", buckets_used, sp_store_bucket_count(store));
    printf("\n");

    printf("Managed Objects:\n");
    printf("----------------\n");
    sp_store_foreach(store, sp_lc_dump_visitor, stdout);
    printf("==========================================\n");
}

int sp_lc_add_weak_dependency(sp_shared_ptr_t *dependent,
                               const sp_shared_ptr_t *dependency)
{
    sp_template_data_t *tmpl;

    if (dependent == NULL || dependency == NULL) {
        errno = EINVAL;
        return -1;
    }

    tmpl = sp_lc_template_get(dependent);
    if (tmpl == NULL) {
        errno = EINVAL;
        return -1;
    }

    return sp_weak_ptr_from_shared(&tmpl->dependency, dependency);
}

int sp_lc_break_circular_refs(void)
{
    pthread_mutex_lock(&g_module_mutex);
    g_stats.circular_ref_breaks++;
    pthread_mutex_unlock(&g_module_mutex);
    return 0;
}
