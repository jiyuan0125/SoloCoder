#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <time.h>
#include <unistd.h>
#include <sys/epoll.h>
#include "event_loop.h"
#include "event_loop_internal.h"

#define MAX_EPOLL_EVENTS 1024
#define INITIAL_FD_MAP_SIZE 256

struct fd_entry {
    int fd;
    int events;
    event_callback_t callback;
    void *data;
    int pending_remove;
};

struct event_loop {
    int epoll_fd;
    struct epoll_event *epoll_events;
    int max_events;
    
    struct fd_entry *fd_map;
    int fd_map_size;
    
    struct timer_heap *timer_heap;
    timer_id_t next_timer_id;
    
    int is_running;
    int stop_requested;
    uint64_t event_count;
    
    int in_callback;
    timer_node_t **pending_timers;
    int pending_timers_count;
    int pending_timers_capacity;
};

static uint64_t get_current_time_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000 + (uint64_t)ts.tv_nsec / 1000000;
}

event_loop_t* loop_create(void) {
    event_loop_t *loop = (event_loop_t *)malloc(sizeof(event_loop_t));
    if (!loop) return NULL;
    
    loop->epoll_fd = epoll_create1(0);
    if (loop->epoll_fd < 0) {
        free(loop);
        return NULL;
    }
    
    loop->max_events = MAX_EPOLL_EVENTS;
    loop->epoll_events = (struct epoll_event *)malloc(sizeof(struct epoll_event) * loop->max_events);
    if (!loop->epoll_events) {
        close(loop->epoll_fd);
        free(loop);
        return NULL;
    }
    
    loop->fd_map_size = INITIAL_FD_MAP_SIZE;
    loop->fd_map = (struct fd_entry *)calloc(loop->fd_map_size, sizeof(struct fd_entry));
    if (!loop->fd_map) {
        free(loop->epoll_events);
        close(loop->epoll_fd);
        free(loop);
        return NULL;
    }
    
    for (int i = 0; i < loop->fd_map_size; i++) {
        loop->fd_map[i].fd = -1;
    }
    
    loop->timer_heap = timer_heap_create();
    if (!loop->timer_heap) {
        free(loop->fd_map);
        free(loop->epoll_events);
        close(loop->epoll_fd);
        free(loop);
        return NULL;
    }
    
    loop->next_timer_id = 1;
    loop->is_running = 0;
    loop->stop_requested = 0;
    loop->event_count = 0;
    loop->in_callback = 0;
    
    loop->pending_timers_capacity = 64;
    loop->pending_timers_count = 0;
    loop->pending_timers = (timer_node_t **)malloc(sizeof(timer_node_t *) * loop->pending_timers_capacity);
    if (!loop->pending_timers) {
        timer_heap_destroy(loop->timer_heap);
        free(loop->fd_map);
        free(loop->epoll_events);
        close(loop->epoll_fd);
        free(loop);
        return NULL;
    }
    
    return loop;
}

void loop_destroy(event_loop_t *loop) {
    if (!loop) return;
    
    timer_heap_destroy(loop->timer_heap);
    free(loop->pending_timers);
    free(loop->fd_map);
    free(loop->epoll_events);
    close(loop->epoll_fd);
    free(loop);
}

static int ensure_fd_map_size(event_loop_t *loop, int fd) {
    if (fd < 0) return -1;
    
    if (fd >= loop->fd_map_size) {
        int new_size = fd + 1;
        if (new_size < loop->fd_map_size * 2) {
            new_size = loop->fd_map_size * 2;
        }
        
        struct fd_entry *new_map = (struct fd_entry *)realloc(loop->fd_map, sizeof(struct fd_entry) * new_size);
        if (!new_map) return -1;
        
        for (int i = loop->fd_map_size; i < new_size; i++) {
            new_map[i].fd = -1;
        }
        
        loop->fd_map = new_map;
        loop->fd_map_size = new_size;
    }
    
    return 0;
}

int loop_add_fd(event_loop_t *loop, int fd, int events, event_callback_t callback, void *data) {
    if (!loop || fd < 0 || !callback) return -1;
    
    if (ensure_fd_map_size(loop, fd) < 0) return -1;
    
    if (loop->fd_map[fd].fd != -1) {
        return loop_modify_fd(loop, fd, events);
    }
    
    struct epoll_event ev;
    memset(&ev, 0, sizeof(ev));
    ev.events = events;
    ev.data.fd = fd;
    
    if (epoll_ctl(loop->epoll_fd, EPOLL_CTL_ADD, fd, &ev) < 0) {
        return -1;
    }
    
    loop->fd_map[fd].fd = fd;
    loop->fd_map[fd].events = events;
    loop->fd_map[fd].callback = callback;
    loop->fd_map[fd].data = data;
    loop->fd_map[fd].pending_remove = 0;
    
    return 0;
}

int loop_remove_fd(event_loop_t *loop, int fd) {
    if (!loop || fd < 0) return -1;
    
    if (fd >= loop->fd_map_size || loop->fd_map[fd].fd == -1) {
        return -1;
    }
    
    if (loop->in_callback) {
        loop->fd_map[fd].pending_remove = 1;
        return 0;
    }
    
    if (epoll_ctl(loop->epoll_fd, EPOLL_CTL_DEL, fd, NULL) < 0) {
        return -1;
    }
    
    loop->fd_map[fd].fd = -1;
    loop->fd_map[fd].callback = NULL;
    loop->fd_map[fd].data = NULL;
    loop->fd_map[fd].pending_remove = 0;
    
    return 0;
}

int loop_modify_fd(event_loop_t *loop, int fd, int events) {
    if (!loop || fd < 0) return -1;
    
    if (fd >= loop->fd_map_size || loop->fd_map[fd].fd == -1) {
        return -1;
    }
    
    struct epoll_event ev;
    memset(&ev, 0, sizeof(ev));
    ev.events = events;
    ev.data.fd = fd;
    
    if (epoll_ctl(loop->epoll_fd, EPOLL_CTL_MOD, fd, &ev) < 0) {
        return -1;
    }
    
    loop->fd_map[fd].events = events;
    
    return 0;
}

timer_id_t loop_add_timer(event_loop_t *loop, uint64_t delay_ms, event_callback_t callback, void *data, int is_repeat) {
    if (!loop || !callback) return 0;
    
    timer_node_t *node = (timer_node_t *)malloc(sizeof(timer_node_t));
    if (!node) return 0;
    
    uint64_t now = get_current_time_ms();
    
    node->id = loop->next_timer_id++;
    node->expire_ms = now + delay_ms;
    node->interval_ms = delay_ms;
    node->is_repeat = is_repeat;
    node->callback = callback;
    node->data = data;
    node->heap_index = -1;
    node->is_pending = 0;
    
    if (timer_heap_insert(loop->timer_heap, node) < 0) {
        free(node);
        return 0;
    }
    
    return node->id;
}

static timer_node_t* find_timer_by_id(event_loop_t *loop, timer_id_t timer_id) {
    if (!loop) return NULL;
    
    for (int i = 0; i < loop->timer_heap->size; i++) {
        if (loop->timer_heap->nodes[i]->id == timer_id) {
            return loop->timer_heap->nodes[i];
        }
    }
    
    for (int i = 0; i < loop->pending_timers_count; i++) {
        if (loop->pending_timers[i] && loop->pending_timers[i]->id == timer_id) {
            return loop->pending_timers[i];
        }
    }
    
    return NULL;
}

int loop_remove_timer(event_loop_t *loop, timer_id_t timer_id) {
    if (!loop || timer_id == 0) return -1;
    
    timer_node_t *node = find_timer_by_id(loop, timer_id);
    if (!node) return -1;
    
    if (node->is_pending) {
        for (int i = 0; i < loop->pending_timers_count; i++) {
            if (loop->pending_timers[i] == node) {
                loop->pending_timers[i] = NULL;
                free(node);
                return 0;
            }
        }
        return -1;
    }
    
    if (timer_heap_remove(loop->timer_heap, node) < 0) {
        return -1;
    }
    
    free(node);
    return 0;
}

void loop_stop(event_loop_t *loop) {
    if (!loop) return;
    loop->stop_requested = 1;
}

static int add_pending_timer(event_loop_t *loop, timer_node_t *node) {
    if (!loop || !node) return -1;
    
    if (loop->pending_timers_count >= loop->pending_timers_capacity) {
        int new_capacity = loop->pending_timers_capacity * 2;
        timer_node_t **new_pending = (timer_node_t **)realloc(loop->pending_timers, sizeof(timer_node_t *) * new_capacity);
        if (!new_pending) return -1;
        
        loop->pending_timers = new_pending;
        loop->pending_timers_capacity = new_capacity;
    }
    
    node->is_pending = 1;
    loop->pending_timers[loop->pending_timers_count++] = node;
    return 0;
}

static uint64_t process_expired_timers(event_loop_t *loop) {
    if (!loop) return 0;
    
    uint64_t count = 0;
    uint64_t now = get_current_time_ms();
    
    loop->pending_timers_count = 0;
    
    while (loop->timer_heap->size > 0) {
        timer_node_t *node = timer_heap_peek(loop->timer_heap);
        if (!node || node->expire_ms > now) {
            break;
        }
        
        timer_heap_pop(loop->timer_heap);
        
        if (add_pending_timer(loop, node) < 0) {
            free(node);
            continue;
        }
    }
    
    loop->in_callback = 1;
    
    for (int i = 0; i < loop->pending_timers_count; i++) {
        timer_node_t *node = loop->pending_timers[i];
        if (!node) continue;
        
        node->callback(0, node->data);
        count++;
        loop->event_count++;
        node->is_pending = 0;
        
        if (node->is_repeat && !loop->stop_requested) {
            node->expire_ms = get_current_time_ms() + node->interval_ms;
            if (timer_heap_insert(loop->timer_heap, node) < 0) {
                free(node);
            }
        } else {
            free(node);
        }
        
        loop->pending_timers[i] = NULL;
    }
    
    loop->in_callback = 0;
    loop->pending_timers_count = 0;
    
    return count;
}

static void cleanup_pending_removals(event_loop_t *loop) {
    if (!loop) return;
    
    for (int i = 0; i < loop->fd_map_size; i++) {
        if (loop->fd_map[i].fd != -1 && loop->fd_map[i].pending_remove) {
            epoll_ctl(loop->epoll_fd, EPOLL_CTL_DEL, loop->fd_map[i].fd, NULL);
            loop->fd_map[i].fd = -1;
            loop->fd_map[i].callback = NULL;
            loop->fd_map[i].data = NULL;
            loop->fd_map[i].pending_remove = 0;
        }
    }
}

static int calculate_epoll_timeout(event_loop_t *loop) {
    if (!loop || loop->timer_heap->size == 0) {
        return -1;
    }
    
    uint64_t now = get_current_time_ms();
    timer_node_t *next_timer = timer_heap_peek(loop->timer_heap);
    
    if (!next_timer) {
        return -1;
    }
    
    if (next_timer->expire_ms <= now) {
        return 0;
    }
    
    uint64_t timeout_ms = next_timer->expire_ms - now;
    
    if (timeout_ms > INT32_MAX) {
        return INT32_MAX;
    }
    
    return (int)timeout_ms;
}

uint64_t loop_run(event_loop_t *loop) {
    if (!loop) return 0;
    
    loop->is_running = 1;
    loop->stop_requested = 0;
    loop->event_count = 0;
    
    while (loop->is_running && !loop->stop_requested) {
        int timeout = calculate_epoll_timeout(loop);
        
        int nfds = epoll_wait(loop->epoll_fd, loop->epoll_events, loop->max_events, timeout);
        
        process_expired_timers(loop);
        
        if (nfds < 0) {
            if (errno == EINTR) {
                continue;
            }
            break;
        }
        
        loop->in_callback = 1;
        
        for (int i = 0; i < nfds; i++) {
            int fd = loop->epoll_events[i].data.fd;
            int events = loop->epoll_events[i].events;
            
            if (fd >= loop->fd_map_size || loop->fd_map[fd].fd == -1) {
                continue;
            }
            
            if (loop->fd_map[fd].pending_remove) {
                continue;
            }
            
            if (loop->fd_map[fd].callback) {
                loop->fd_map[fd].callback(events, loop->fd_map[fd].data);
                loop->event_count++;
            }
        }
        
        loop->in_callback = 0;
        
        cleanup_pending_removals(loop);
    }
    
    loop->is_running = 0;
    return loop->event_count;
}
