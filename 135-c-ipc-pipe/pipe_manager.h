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
    uint8_t *buffer;
    size_t buffer_size;
    size_t total_len;
    size_t written_len;
    int is_active;
} pipe_write_ctx_t;

int pipe_write_ctx_init(pipe_write_ctx_t *ctx);
void pipe_write_ctx_destroy(pipe_write_ctx_t *ctx);
void pipe_write_ctx_reset(pipe_write_ctx_t *ctx);

int pipe_pair_create(pipe_pair_t *pp);
void pipe_pair_close(pipe_pair_t *pp);
void pipe_pair_close_read(pipe_pair_t *pp);
void pipe_pair_close_write(pipe_pair_t *pp);

int pipe_set_nonblocking(int fd);
int pipe_set_blocking(int fd);

ssize_t pipe_write_message(int fd, pipe_write_ctx_t *ctx, 
                            const uint8_t *msg_data, size_t msg_len);
ssize_t pipe_write_message_continue(int fd, pipe_write_ctx_t *ctx);
int pipe_write_ctx_is_complete(const pipe_write_ctx_t *ctx);

ssize_t pipe_read_partial(int fd, message_buffer_t *mb);

#endif
