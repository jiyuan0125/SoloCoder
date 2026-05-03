#ifndef SHM_QUEUE_H
#define SHM_QUEUE_H

#include <stddef.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

#define SHMQ_OK           0
#define SHMQ_ERROR       -1
#define SHMQ_TIMEOUT     -2
#define SHMQ_FULL        -3
#define SHMQ_EMPTY       -4
#define SHMQ_EXISTS      -5
#define SHMQ_NOT_EXISTS  -6

#define SHMQ_WAIT_INFINITE  (-1)
#define SHMQ_WAIT_NOWAIT    (0)

#define SHMQ_DEFAULT_NAME     "/shmq_default"
#define SHMQ_MIN_BUFFER_SIZE  (1024)
#define SHMQ_MAX_MSG_LEN      (1024 * 1024)

typedef enum {
    SHMQ_MODE_BLOCKING = 0,
    SHMQ_MODE_NONBLOCKING = 1,
} shmq_mode_t;

typedef enum {
    SHMQ_ROLE_UNKNOWN = 0,
    SHMQ_ROLE_CREATOR = 1,
    SHMQ_ROLE_ATTACHER = 2,
} shmq_role_t;

typedef struct shm_queue_config {
    const char *name;
    size_t buffer_size;
    size_t max_msg_len;
    int default_timeout_ms;
    bool reuse_existing;
    bool cleanup_on_destroy;
} shm_queue_config_t;

typedef struct shm_queue {
    shm_queue_config_t config;
    shmq_role_t role;
    void *private_data;
} shm_queue_t;

void shmq_config_init(shm_queue_config_t *cfg);

int shmq_create(shm_queue_t *q, const shm_queue_config_t *cfg);
int shmq_open(shm_queue_t *q, const char *name);
int shmq_close(shm_queue_t *q);
int shmq_destroy(const char *name);
int shmq_destroy_queue(shm_queue_t *q);

bool shmq_exists(const char *name);
size_t shmq_calculate_total_size(size_t buffer_size);

int shmq_send(shm_queue_t *q, const void *data, size_t len, int timeout_ms);
int shmq_recv(shm_queue_t *q, void *data, size_t *len, int timeout_ms);
int shmq_try_send(shm_queue_t *q, const void *data, size_t len);
int shmq_try_recv(shm_queue_t *q, void *data, size_t *len);

size_t shmq_available_space(shm_queue_t *q);
size_t shmq_available_data(shm_queue_t *q);
bool shmq_is_full(shm_queue_t *q);
bool shmq_is_empty(shm_queue_t *q);

int shmq_get_max_msg_len(shm_queue_t *q, size_t *out_len);
int shmq_get_buffer_size(shm_queue_t *q, size_t *out_size);

#ifdef __cplusplus
}
#endif

#endif
