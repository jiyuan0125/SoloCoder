#ifndef LIFECYCLE_H
#define LIFECYCLE_H

#include "shared_ptr.h"
#include "object_store.h"

#ifdef __cplusplus
extern "C" {
#endif

#define SP_LC_MAX_IMAGE_WIDTH   8192
#define SP_LC_MAX_IMAGE_HEIGHT  8192
#define SP_LC_MAX_TEMPLATE_SIZE (10 * 1024 * 1024)

typedef enum {
    SP_LC_POLICY_LAZY,
    SP_LC_POLICY_EAGER,
    SP_LC_POLICY_AUTO_RELOAD,
    SP_LC_POLICY_MAX
} sp_lc_policy_t;

typedef struct sp_image_data {
    int                    width;
    int                    height;
    int                    channels;
    size_t                 pixel_data_size;
    unsigned char          *pixel_data;
    char                   source_path[512];
} sp_image_data_t;

typedef struct sp_template_data {
    size_t                 compiled_size;
    void                   *compiled_code;
    char                   template_name[256];
    int                    version;
    sp_weak_ptr_t          dependency;
} sp_template_data_t;

typedef struct sp_lc_stats {
    size_t                 total_images_loaded;
    size_t                 total_templates_loaded;
    size_t                 total_bytes_allocated;
    size_t                 total_references_created;
    size_t                 total_objects_destroyed;
    size_t                 circular_ref_breaks;
} sp_lc_stats_t;

typedef int (*sp_lc_image_loader_fn)(const char *path,
                                      sp_image_data_t **out_image);
typedef int (*sp_lc_template_loader_fn)(const char *name,
                                         sp_template_data_t **out_template);

int sp_lc_module_init(void);
void sp_lc_module_cleanup(void);

int sp_lc_set_image_loader(sp_lc_image_loader_fn loader);
int sp_lc_set_template_loader(sp_lc_template_loader_fn loader);
int sp_lc_set_policy(sp_lc_policy_t policy);

int sp_lc_load_image(const char *path, sp_shared_ptr_t *out_ptr);
int sp_lc_load_template(const char *name, sp_shared_ptr_t *out_ptr);

sp_image_data_t *sp_lc_image_get(const sp_shared_ptr_t *ptr);
sp_template_data_t *sp_lc_template_get(const sp_shared_ptr_t *ptr);

void sp_lc_release(sp_shared_ptr_t *ptr);

int sp_lc_reload_image(const char *path);
int sp_lc_reload_template(const char *name);

int sp_lc_unload_unused(int force);

void sp_lc_get_stats(sp_lc_stats_t *out_stats);
void sp_lc_dump_state(void);

int sp_lc_add_weak_dependency(sp_shared_ptr_t *dependent,
                               const sp_shared_ptr_t *dependency);
int sp_lc_break_circular_refs(void);

sp_image_data_t *sp_lc_image_data_alloc(int width, int height, int channels);
void sp_lc_image_data_free(sp_image_data_t *img);
sp_template_data_t *sp_lc_template_data_alloc(size_t code_size);
void sp_lc_template_data_free(sp_template_data_t *tmpl);

#ifdef __cplusplus
}
#endif

#endif
