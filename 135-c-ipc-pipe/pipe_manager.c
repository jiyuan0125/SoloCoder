#include "pipe_manager.h"
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>

int pipe_manager_init(pipe_manager_t *pm, size_t initial_capacity) {
    if (initial_capacity == 0) {
        initial_capacity = 4;
    }
    
    pm->pipes = (pipe_pair_t *)malloc(initial_capacity * sizeof(pipe_pair_t));
    if (pm->pipes == NULL) {
        return -1;
    }
    
    pm->pipe_count = 0;
    pm->pipe_capacity = initial_capacity;
    return 0;
}

void pipe_manager_destroy(pipe_manager_t *pm) {
    pipe_manager_close_all(pm);
    if (pm->pipes != NULL) {
        free(pm->pipes);
        pm->pipes = NULL;
    }
    pm->pipe_count = 0;
    pm->pipe_capacity = 0;
}

static int pipe_manager_resize(pipe_manager_t *pm, size_t new_capacity) {
    if (new_capacity <= pm->pipe_capacity) {
        return 0;
    }
    
    pipe_pair_t *new_pipes = (pipe_pair_t *)realloc(pm->pipes, 
                                                      new_capacity * sizeof(pipe_pair_t));
    if (new_pipes == NULL) {
        return -1;
    }
    
    pm->pipes = new_pipes;
    pm->pipe_capacity = new_capacity;
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

ssize_t pipe_write_message(int fd, const uint8_t *msg_data, size_t msg_len) {
    size_t total_size = message_frame_calculate_total_size(msg_len);
    uint8_t *buffer = (uint8_t *)malloc(total_size);
    
    if (buffer == NULL) {
        return -1;
    }
    
    int pack_result = message_frame_pack(buffer, total_size, msg_data, msg_len);
    if (pack_result < 0) {
        free(buffer);
        return -1;
    }
    
    ssize_t written = write(fd, buffer, total_size);
    free(buffer);
    
    if (written < 0) {
        if (errno == EAGAIN || errno == EWOULDBLOCK) {
            return PIPE_AGAIN;
        }
        return PIPE_ERROR;
    }
    
    if ((size_t)written < total_size) {
        return PIPE_AGAIN;
    }
    
    return written;
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

int pipe_manager_add_pipe(pipe_manager_t *pm, const pipe_pair_t *pp) {
    if (pm->pipe_count >= pm->pipe_capacity) {
        size_t new_capacity = pm->pipe_capacity * 2;
        if (pipe_manager_resize(pm, new_capacity) != 0) {
            return -1;
        }
    }
    
    pm->pipes[pm->pipe_count] = *pp;
    pm->pipe_count++;
    return (int)(pm->pipe_count - 1);
}

pipe_pair_t *pipe_manager_get_pipe(pipe_manager_t *pm, size_t index) {
    if (index >= pm->pipe_count) {
        return NULL;
    }
    return &pm->pipes[index];
}

int pipe_manager_remove_pipe(pipe_manager_t *pm, size_t index) {
    if (index >= pm->pipe_count) {
        return -1;
    }
    
    pipe_pair_close(&pm->pipes[index]);
    
    if (index < pm->pipe_count - 1) {
        memmove(&pm->pipes[index], &pm->pipes[index + 1], 
                (pm->pipe_count - index - 1) * sizeof(pipe_pair_t));
    }
    
    pm->pipe_count--;
    return 0;
}

void pipe_manager_close_all(pipe_manager_t *pm) {
    for (size_t i = 0; i < pm->pipe_count; i++) {
        pipe_pair_close(&pm->pipes[i]);
    }
    pm->pipe_count = 0;
}
