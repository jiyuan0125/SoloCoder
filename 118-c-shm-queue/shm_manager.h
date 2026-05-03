#ifndef SHM_MANAGER_H
#define SHM_MANAGER_H

#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct shm_manager {
    const char *name;
    int fd;
    void *addr;
    size_t size;
    bool owner;
} shm_manager_t;

int shm_manager_create(shm_manager_t *mgr, const char *name, size_t size);
int shm_manager_open(shm_manager_t *mgr, const char *name, size_t size);
int shm_manager_close(shm_manager_t *mgr);
int shm_manager_unlink(const char *name);
bool shm_manager_exists(const char *name);
void shm_manager_init(shm_manager_t *mgr);

#ifdef __cplusplus
}
#endif

#endif
