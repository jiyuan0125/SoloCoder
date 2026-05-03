#include "shm_manager.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <errno.h>

void shm_manager_init(shm_manager_t *mgr) {
    memset(mgr, 0, sizeof(shm_manager_t));
    mgr->fd = -1;
}

static int shm_do_open(shm_manager_t *mgr, const char *name, size_t size, int oflag, mode_t mode) {
    shm_manager_init(mgr);
    
    mgr->name = strdup(name);
    if (mgr->name == NULL) {
        return -1;
    }
    
    mgr->fd = shm_open(name, oflag, mode);
    if (mgr->fd < 0) {
        free((void *)mgr->name);
        mgr->name = NULL;
        return -1;
    }
    
    if (oflag & O_CREAT) {
        if (ftruncate(mgr->fd, (off_t)size) < 0) {
            close(mgr->fd);
            shm_unlink(name);
            free((void *)mgr->name);
            mgr->name = NULL;
            mgr->fd = -1;
            return -1;
        }
        mgr->owner = true;
    }
    
    mgr->addr = mmap(NULL, size, PROT_READ | PROT_WRITE, MAP_SHARED, mgr->fd, 0);
    if (mgr->addr == MAP_FAILED) {
        close(mgr->fd);
        if (mgr->owner) {
            shm_unlink(name);
        }
        free((void *)mgr->name);
        mgr->name = NULL;
        mgr->fd = -1;
        return -1;
    }
    
    mgr->size = size;
    return 0;
}

int shm_manager_create(shm_manager_t *mgr, const char *name, size_t size) {
    return shm_do_open(mgr, name, size, O_RDWR | O_CREAT | O_EXCL, S_IRUSR | S_IWUSR);
}

int shm_manager_open(shm_manager_t *mgr, const char *name, size_t size) {
    return shm_do_open(mgr, name, size, O_RDWR, S_IRUSR | S_IWUSR);
}

int shm_manager_close(shm_manager_t *mgr) {
    int ret = 0;
    
    if (mgr->addr != NULL && mgr->addr != MAP_FAILED) {
        if (munmap(mgr->addr, mgr->size) < 0) {
            ret = -1;
        }
        mgr->addr = NULL;
    }
    
    if (mgr->fd >= 0) {
        if (close(mgr->fd) < 0) {
            ret = -1;
        }
        mgr->fd = -1;
    }
    
    if (mgr->name != NULL) {
        free((void *)mgr->name);
        mgr->name = NULL;
    }
    
    return ret;
}

int shm_manager_unlink(const char *name) {
    return shm_unlink(name);
}

bool shm_manager_exists(const char *name) {
    int fd = shm_open(name, O_RDONLY, 0);
    if (fd >= 0) {
        close(fd);
        return true;
    }
    return (errno != ENOENT);
}
