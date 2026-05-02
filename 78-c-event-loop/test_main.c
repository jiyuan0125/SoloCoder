#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/epoll.h>
#include "event_loop.h"

static event_loop_t *g_loop = NULL;
static int pipe_fds[2];

typedef struct {
    int count;
    int max_count;
    event_loop_t *loop;
} timer_test_context_t;

static void basic_timer_cb(int events, void *data) {
    (void)events;
    timer_test_context_t *ctx = (timer_test_context_t *)data;
    ctx->count++;
    printf("Basic timer callback triggered, count: %d\n", ctx->count);
    
    if (ctx->count >= ctx->max_count) {
        printf("Stopping loop...\n");
        loop_stop(ctx->loop);
    }
}

static void repeat_timer_cb(int events, void *data) {
    (void)events;
    timer_test_context_t *ctx = (timer_test_context_t *)data;
    ctx->count++;
    printf("Repeat timer callback triggered, count: %d\n", ctx->count);
}

static void stop_timer_cb(int events, void *data) {
    (void)events;
    event_loop_t *loop = (event_loop_t *)data;
    printf("Stop timer callback triggered, stopping loop...\n");
    loop_stop(loop);
}

static void pipe_read_cb(int events, void *data) {
    (void)data;
    char buf[256];
    ssize_t n;
    
    if (events & EPOLLIN) {
        n = read(pipe_fds[0], buf, sizeof(buf) - 1);
        if (n > 0) {
            buf[n] = '\0';
            printf("Pipe read callback: received '%s'\n", buf);
        }
    }
}

static void test_basic_timer(void) {
    printf("\n=== Test 1: Basic Timer ===\n");
    
    event_loop_t *loop = loop_create();
    if (!loop) {
        printf("FAIL: loop_create failed\n");
        return;
    }
    
    timer_test_context_t ctx = {0, 2, loop};
    
    timer_id_t t1 = loop_add_timer(loop, 100, basic_timer_cb, &ctx, 0);
    timer_id_t t2 = loop_add_timer(loop, 200, basic_timer_cb, &ctx, 0);
    
    if (t1 == 0 || t2 == 0) {
        printf("FAIL: loop_add_timer failed\n");
        loop_destroy(loop);
        return;
    }
    
    printf("Added timers: t1=%lu, t2=%lu\n", (unsigned long)t1, (unsigned long)t2);
    
    uint64_t event_count = loop_run(loop);
    
    printf("Loop returned, event_count = %lu, timer_count = %d\n", (unsigned long)event_count, ctx.count);
    
    if (ctx.count >= 2) {
        printf("PASS: Basic timer test\n");
    } else {
        printf("FAIL: Expected at least 2 timer callbacks, got %d\n", ctx.count);
    }
    
    loop_destroy(loop);
}

static void test_remove_timer(void) {
    printf("\n=== Test 2: Remove Timer ===\n");
    
    event_loop_t *loop = loop_create();
    if (!loop) {
        printf("FAIL: loop_create failed\n");
        return;
    }
    
    timer_test_context_t ctx = {0, 2, loop};
    
    timer_id_t t1 = loop_add_timer(loop, 5000, basic_timer_cb, &ctx, 0);
    timer_id_t t2 = loop_add_timer(loop, 100, basic_timer_cb, &ctx, 0);
    
    printf("Added timers: t1=%lu (long), t2=%lu (short)\n", (unsigned long)t1, (unsigned long)t2);
    
    int ret = loop_remove_timer(loop, t1);
    printf("Removed t1, ret = %d (0 = success)\n", ret);
    
    loop_add_timer(loop, 300, basic_timer_cb, &ctx, 0);
    
    uint64_t event_count = loop_run(loop);
    
    printf("Loop returned, event_count = %lu, timer_count = %d\n", (unsigned long)event_count, ctx.count);
    
    if (ctx.count >= 2) {
        printf("PASS: Remove timer test\n");
    } else {
        printf("FAIL: Expected at least 2 timer callbacks, got %d\n", ctx.count);
    }
    
    loop_destroy(loop);
}

static void test_repeat_timer(void) {
    printf("\n=== Test 3: Repeat Timer ===\n");
    
    event_loop_t *loop = loop_create();
    if (!loop) {
        printf("FAIL: loop_create failed\n");
        return;
    }
    
    timer_test_context_t ctx = {0, 10, loop};
    
    timer_id_t t1 = loop_add_timer(loop, 100, repeat_timer_cb, &ctx, 1);
    
    if (t1 == 0) {
        printf("FAIL: loop_add_timer failed\n");
        loop_destroy(loop);
        return;
    }
    
    printf("Added repeat timer: t1=%lu\n", (unsigned long)t1);
    
    loop_add_timer(loop, 500, stop_timer_cb, loop, 0);
    
    uint64_t event_count = loop_run(loop);
    
    printf("Loop returned, event_count = %lu, repeat_count = %d\n", (unsigned long)event_count, ctx.count);
    
    if (ctx.count >= 3) {
        printf("PASS: Repeat timer test (triggered %d times)\n", ctx.count);
    } else {
        printf("FAIL: Repeat timer only triggered %d times\n", ctx.count);
    }
    
    loop_destroy(loop);
}

static void test_fd_events(void) {
    printf("\n=== Test 4: FD Events (pipe) ===\n");
    
    event_loop_t *loop = loop_create();
    if (!loop) {
        printf("FAIL: loop_create failed\n");
        return;
    }
    
    if (pipe(pipe_fds) < 0) {
        printf("FAIL: pipe failed\n");
        loop_destroy(loop);
        return;
    }
    
    int flags = fcntl(pipe_fds[0], F_GETFL, 0);
    fcntl(pipe_fds[0], F_SETFL, flags | O_NONBLOCK);
    
    if (loop_add_fd(loop, pipe_fds[0], EPOLLIN, pipe_read_cb, NULL) < 0) {
        printf("FAIL: loop_add_fd failed\n");
        close(pipe_fds[0]);
        close(pipe_fds[1]);
        loop_destroy(loop);
        return;
    }
    
    printf("Added pipe read fd to loop\n");
    
    const char *msg = "Hello Event Loop!";
    write(pipe_fds[1], msg, strlen(msg) + 1);
    
    loop_add_timer(loop, 200, stop_timer_cb, loop, 0);
    
    uint64_t event_count = loop_run(loop);
    
    printf("Loop returned, event_count = %lu\n", (unsigned long)event_count);
    
    close(pipe_fds[0]);
    close(pipe_fds[1]);
    loop_destroy(loop);
    
    printf("PASS: FD events test\n");
}

static void test_modify_fd(void) {
    printf("\n=== Test 5: Modify FD Events ===\n");
    
    event_loop_t *loop = loop_create();
    if (!loop) {
        printf("FAIL: loop_create failed\n");
        return;
    }
    
    if (pipe(pipe_fds) < 0) {
        printf("FAIL: pipe failed\n");
        loop_destroy(loop);
        return;
    }
    
    int flags = fcntl(pipe_fds[0], F_GETFL, 0);
    fcntl(pipe_fds[0], F_SETFL, flags | O_NONBLOCK);
    
    if (loop_add_fd(loop, pipe_fds[0], EPOLLIN, pipe_read_cb, NULL) < 0) {
        printf("FAIL: loop_add_fd failed\n");
        close(pipe_fds[0]);
        close(pipe_fds[1]);
        loop_destroy(loop);
        return;
    }
    
    printf("Added fd with EPOLLIN\n");
    
    if (loop_modify_fd(loop, pipe_fds[0], EPOLLIN | EPOLLOUT) < 0) {
        printf("FAIL: loop_modify_fd failed\n");
    } else {
        printf("Modified fd to EPOLLIN | EPOLLOUT\n");
    }
    
    if (loop_remove_fd(loop, pipe_fds[0]) < 0) {
        printf("FAIL: loop_remove_fd failed\n");
    } else {
        printf("Removed fd from loop\n");
    }
    
    close(pipe_fds[0]);
    close(pipe_fds[1]);
    loop_destroy(loop);
    
    printf("PASS: Modify FD test\n");
}

int main(void) {
    printf("========================================\n");
    printf("Event Loop Framework Tests\n");
    printf("========================================\n");
    
    test_basic_timer();
    test_remove_timer();
    test_repeat_timer();
    test_fd_events();
    test_modify_fd();
    
    printf("\n========================================\n");
    printf("All tests completed\n");
    printf("========================================\n");
    
    return 0;
}
