#include "pipe_manager.h"
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>

int pipe_write_ctx_init(pipe_write_ctx_t *ctx) {
    ctx->buffer = NULL;
    ctx->buffer_size = 0;
    ctx->total_len = 0;
    ctx->written_len = 0;
    ctx->is_active = 0;
    return 0;
}

void pipe_write_ctx_destroy(pipe_write_ctx_t *ctx) {
    if (ctx->buffer != NULL) {
        free(ctx->buffer);
        ctx->buffer = NULL;
    }
    ctx->buffer_size = 0;
    ctx->total_len = 0;
    ctx->written_len = 0;
    ctx->is_active = 0;
}

void pipe_write_ctx_reset(pipe_write_ctx_t *ctx) {
    ctx->total_len = 0;
    ctx->written_len = 0;
    ctx->is_active = 0;
}

static int pipe_write_ctx_resize(pipe_write_ctx_t *ctx, size_t new_size) {
    if (new_size <= ctx->buffer_size) {
        return 0;
    }
    
    uint8_t *new_buf = (uint8_t *)realloc(ctx->buffer, new_size);
    if (new_buf == NULL) {
        return -1;
    }
    
    ctx->buffer = new_buf;
    ctx->buffer_size = new_size;
    return 0;
}

int pipe_pair_create(pipe_pair_t *pp) {
    int fds[2];
    
    if (pipe(fds) != 0) {
        return -1;
    }
    
    pp->read_fd = fds[PIPE_READ_END];
    pp->write_fd = fds[PIPE_WRITE_END];
    pp->is_nonblocking = 0;
    return 0;
}

void pipe_pair_close(pipe_pair_t *pp) {
    pipe_pair_close_read(pp);
    pipe_pair_close_write(pp);
}

void pipe_pair_close_read(pipe_pair_t *pp) {
    if (pp->read_fd >= 0) {
        close(pp->read_fd);
        pp->read_fd = -1;
    }
}

void pipe_pair_close_write(pipe_pair_t *pp) {
    if (pp->write_fd >= 0) {
        close(pp->write_fd);
        pp->write_fd = -1;
    }
}

int pipe_set_nonblocking(int fd) {
    if (fd < 0) {
        return -1;
    }
    
    int flags = fcntl(fd, F_GETFL, 0);
    if (flags == -1) {
        return -1;
    }
    
    if (fcntl(fd, F_SETFL, flags | O_NONBLOCK) == -1) {
        return -1;
    }
    
    return 0;
}

int pipe_set_blocking(int fd) {
    if (fd < 0) {
        return -1;
    }
    
    int flags = fcntl(fd, F_GETFL, 0);
    if (flags == -1) {
        return -1;
    }
    
    if (fcntl(fd, F_SETFL, flags & ~O_NONBLOCK) == -1) {
        return -1;
    }
    
    return 0;
}

int pipe_write_ctx_is_complete(const pipe_write_ctx_t *ctx) {
    if (!ctx->is_active) {
        return 1;
    }
    return (ctx->written_len >= ctx->total_len) ? 1 : 0;
}

ssize_t pipe_write_message(int fd, pipe_write_ctx_t *ctx, 
                            const uint8_t *msg_data, size_t msg_len) {
    if (ctx->is_active && ctx->written_len < ctx->total_len) {
        return pipe_write_message_continue(fd, ctx);
    }
    
    size_t total_size = message_frame_calculate_total_size(msg_len);
    
    if (pipe_write_ctx_resize(ctx, total_size) != 0) {
        return PIPE_ERROR;
    }
    
    int pack_result = message_frame_pack(ctx->buffer, total_size, msg_data, msg_len);
    if (pack_result < 0) {
        return PIPE_ERROR;
    }
    
    ctx->total_len = total_size;
    ctx->written_len = 0;
    ctx->is_active = 1;
    
    return pipe_write_message_continue(fd, ctx);
}

ssize_t pipe_write_message_continue(int fd, pipe_write_ctx_t *ctx) {
    if (!ctx->is_active) {
        return PIPE_ERROR;
    }
    
    if (ctx->written_len >= ctx->total_len) {
        return (ssize_t)ctx->total_len;
    }
    
    ssize_t written = write(fd, 
                            ctx->buffer + ctx->written_len, 
                            ctx->total_len - ctx->written_len);
    
    if (written < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return PIPE_AGAIN;
        }
        return PIPE_ERROR;
    }
    
    ctx->written_len += (size_t)written;
    
    if (ctx->written_len >= ctx->total_len) {
        ctx->is_active = 0;
        return (ssize_t)ctx->total_len;
    }
    
    return PIPE_AGAIN;
}

ssize_t pipe_read_partial(int fd, message_buffer_t *mb) {
    uint8_t temp_buf[4096];
    ssize_t read_bytes = read(fd, temp_buf, sizeof(temp_buf));
    
    if (read_bytes == 0) {
        return PIPE_EOF;
    }
    
    if (read_bytes < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return PIPE_AGAIN;
        }
        return PIPE_ERROR;
    }
    
    if (message_buffer_append(mb, temp_buf, (size_t)read_bytes) != 0) {
        return PIPE_ERROR;
    }
    
    return read_bytes;
}
