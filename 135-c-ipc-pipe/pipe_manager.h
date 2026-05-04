#ifndef PIPE_MANAGER_H
#define PIPE_MANAGER_H

#include <stdint.h>
#include <stddef.h>
#include <sys/types.h>
#include "message_frame.h"

#define PIPE_READ_END 0
#define PIPE_WRITE_END 1

typedef enum {
    PIPE_OK = 0,
    PIPE_ERROR = -1,
    PIPE_EOF = -2,
    PIPE_AGAIN = -3,
    PIPE_FULL = -4
} pipe_result_t;

typedef struct {
    int read_fd;
    int write_fd;
    int is_nonblocking;
} pipe_pair_t;

typedef struct {
    pipe_pair_t *pipes;
    size_t pipe_count;
    size_t pipe_capacity;
} pipe_manager_t;

int pipe_manager_init(pipe_manager_t *pm, size_t initial_capacity);
void pipe_manager_destroy(pipe_manager_t *pm);

int pipe_pair_create(pipe_pair_t *pp);
void pipe_pair_close(pipe_pair_t *pp);
void pipe_pair_close_read(pipe_pair_t *pp);
void pipe_pair_close_write(pipe_pair_t *pp);

int pipe_set_nonblocking(int fd);
int pipe_set_blocking(int fd);

ssize_t pipe_write_message(int fd, const uint8_t *msg_data, size_t msg_len);
ssize_t pipe_read_partial(int fd, message_buffer_t *mb);

int pipe_manager_add_pipe(pipe_manager_t *pm, const pipe_pair_t *pp);
pipe_pair_t *pipe_manager_get_pipe(pipe_manager_t *pm, size_t index);
int pipe_manager_remove_pipe(pipe_manager_t *pm, size_t index);
void pipe_manager_close_all(pipe_manager_t *pm);

#endif
