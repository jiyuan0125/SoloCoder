#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include "event_channel.h"
#include "event_dispatcher.h"
#include "event_timer.h"

#define EVENT_MOUSE_CLICK      1
#define EVENT_KEYBOARD_INPUT   2
#define EVENT_TIMER            3
#define EVENT_SYSTEM_SHUTDOWN  4
#define EVENT_WINDOW_RESIZE    5
#define EVENT_NETWORK_DATA     6
#define EVENT_DIALOG_SHOW      7
#define EVENT_QUIT             8

typedef struct {
    event_channel_t *channel;
    int thread_id;
} producer_args_t;

static void on_mouse_click(event_t *event, void *user_data) {
    int *x = (int*)event->data;
    int *y = x + 1;
    printf("[Handler] Mouse click at (%d, %d), priority=%u\n", *x, *y, event->priority);
    
    printf("[Handler] Posting DIALOG_SHOW event from callback...\n");
    event_channel_post((event_channel_t*)user_data, EVENT_DIALOG_SHOW, 
                        EVENT_PRIORITY_NORMAL, NULL, 0, NULL);
}

static void on_keyboard_input(event_t *event, void *user_data) {
    char *key = (char*)event->data;
    printf("[Handler] Keyboard input: '%s', priority=%u\n", key, event->priority);
    (void)user_data;
}

static void on_timer(event_t *event, void *user_data) {
    uint64_t *timer_id = (uint64_t*)event->data;
    printf("[Handler] Timer event fired, timer_id=%lu, priority=%u\n", 
           (unsigned long)*timer_id, event->priority);
    (void)user_data;
}

static void on_system_shutdown(event_t *event, void *user_data) {
    printf("[Handler] SYSTEM SHUTDOWN event (HIGH PRIORITY), priority=%u\n", event->priority);
    (void)user_data;
}

static void on_window_resize(event_t *event, void *user_data) {
    int *w = (int*)event->data;
    int *h = w + 1;
    printf("[Handler] Window resize: %dx%d, priority=%u\n", *w, *h, event->priority);
    (void)user_data;
}

static void on_network_data(event_t *event, void *user_data) {
    char *data = (char*)event->data;
    printf("[Handler] Network data: '%s', priority=%u\n", data, event->priority);
    (void)user_data;
}

static void on_dialog_show(event_t *event, void *user_data) {
    printf("[Handler] Dialog shown (posted from callback), priority=%u\n", event->priority);
    (void)user_data;
}

static void on_quit(event_t *event, void *user_data) {
    printf("[Handler] Quit event received, stopping dispatcher...\n");
    event_dispatcher_stop((event_dispatcher_t*)user_data);
}

static void* producer_thread(void *arg) {
    producer_args_t *args = (producer_args_t*)arg;
    event_channel_t *channel = args->channel;
    int tid = args->thread_id;
    
    printf("[Producer %d] Starting...\n", tid);
    
    for (int i = 0; i < 3; i++) {
        char data[64];
        snprintf(data, sizeof(data), "Network packet #%d from thread %d", i, tid);
        event_channel_post(channel, EVENT_NETWORK_DATA, EVENT_PRIORITY_LOW,
                            data, strlen(data) + 1, NULL);
        printf("[Producer %d] Posted NETWORK_DATA #%d\n", tid, i);
        usleep(100000);
    }
    
    printf("[Producer %d] Exiting...\n", tid);
    return NULL;
}

int main(void) {
    printf("=== Event Dispatcher System Demo\n");
    printf("==================================\n\n");
    
    event_channel_t *channel = event_channel_create();
    if (!channel) {
        fprintf(stderr, "Failed to create event channel\n");
        return 1;
    }
    
    event_dispatcher_t *dispatcher = event_dispatcher_create(channel);
    if (!dispatcher) {
        fprintf(stderr, "Failed to create event dispatcher\n");
        event_channel_destroy(channel);
        return 1;
    }
    
    event_timer_t *timer = event_timer_create(channel);
    if (!timer) {
        fprintf(stderr, "Failed to create event timer\n");
        event_dispatcher_destroy(dispatcher);
        event_channel_destroy(channel);
        return 1;
    }
    
    printf("[Main] Registering event handlers...\n");
    
    event_dispatcher_register(dispatcher, EVENT_MOUSE_CLICK, on_mouse_click, channel);
    event_dispatcher_register(dispatcher, EVENT_KEYBOARD_INPUT, on_keyboard_input, NULL);
    event_dispatcher_register(dispatcher, EVENT_TIMER, on_timer, NULL);
    event_dispatcher_register(dispatcher, EVENT_SYSTEM_SHUTDOWN, on_system_shutdown, NULL);
    event_dispatcher_register(dispatcher, EVENT_WINDOW_RESIZE, on_window_resize, NULL);
    event_dispatcher_register(dispatcher, EVENT_NETWORK_DATA, on_network_data, NULL);
    event_dispatcher_register(dispatcher, EVENT_DIALOG_SHOW, on_dialog_show, NULL);
    event_dispatcher_register(dispatcher, EVENT_QUIT, on_quit, dispatcher);
    
    printf("\n[Main] Test 1: Priority ordering\n");
    printf("[Main] Posting events in arbitrary order...\n");
    
    int mouse_pos[2] = {100, 200};
    event_channel_post(channel, EVENT_MOUSE_CLICK, EVENT_PRIORITY_NORMAL,
                        mouse_pos, sizeof(mouse_pos), NULL);
    printf("[Main] Posted MOUSE_CLICK (NORMAL priority)\n");
    
    event_channel_post(channel, EVENT_SYSTEM_SHUTDOWN, EVENT_PRIORITY_HIGHEST,
                        NULL, 0, NULL);
    printf("[Main] Posted SYSTEM_SHUTDOWN (HIGHEST priority)\n");
    
    int window_size[2] = {800, 600};
    event_channel_post(channel, EVENT_WINDOW_RESIZE, EVENT_PRIORITY_HIGH,
                        window_size, sizeof(window_size), NULL);
    printf("[Main] Posted WINDOW_RESIZE (HIGH priority)\n");
    
    char key[] = "Enter";
    event_channel_post(channel, EVENT_KEYBOARD_INPUT, EVENT_PRIORITY_NORMAL,
                        key, sizeof(key), NULL);
    printf("[Main] Posted KEYBOARD_INPUT (NORMAL priority)\n");
    
    printf("\n[Main] Processing these events should be ordered by priority:\n");
    printf("  1. SYSTEM_SHUTDOWN (HIGHEST)\n");
    printf("  2. WINDOW_RESIZE (HIGH)\n");
    printf("  3. MOUSE_CLICK (NORMAL) - will post DIALOG_SHOW\n");
    printf("  4. KEYBOARD_INPUT (NORMAL)\n");
    printf("  5. DIALOG_SHOW (posted from MOUSE_CLICK handler)\n\n");
    
    printf("--- Processing events ---\n");
    for (int i = 0; i < 5; i++) {
        event_t *event = event_channel_try_get(channel);
        if (event) {
            event_dispatcher_dispatch(dispatcher, event);
            event_free(event);
        }
    }
    
    printf("\n[Main] Test 2: Multi-threaded producers\n");
    printf("[Main] Starting 2 producer threads...\n");
    
    pthread_t producers[2];
    producer_args_t args[2];
    for (int i = 0; i < 2; i++) {
        args[i].channel = channel;
        args[i].thread_id = i + 1;
        pthread_create(&producers[i], NULL, producer_thread, &args[i]);
    }
    
    for (int i = 0; i < 2; i++) {
        pthread_join(producers[i], NULL);
    }
    
    printf("\n[Main] All producer threads finished. Processing network events...\n");
    while (!event_channel_empty(channel)) {
        event_t *event = event_channel_try_get(channel);
        if (event) {
            event_dispatcher_dispatch(dispatcher, event);
            event_free(event);
        }
    }
    
    printf("\n[Main] Test 3: Timer events and cancellation\n");
    printf("[Main] Posting timer events (1s, 2s, 3s delays)...\n");
    
    uint64_t timer_data1 = 1;
    timer_id_t tid1 = event_timer_post_delayed(timer, EVENT_TIMER, EVENT_PRIORITY_NORMAL,
                                                &timer_data1, sizeof(timer_data1), 1000, NULL);
    printf("[Main] Posted timer 1 (1s delay, timer_id=%lu)\n", (unsigned long)tid1);
    
    uint64_t timer_data2 = 2;
    timer_id_t tid2 = event_timer_post_delayed(timer, EVENT_TIMER, EVENT_PRIORITY_HIGH,
                                                &timer_data2, sizeof(timer_data2), 2000, NULL);
    printf("[Main] Posted timer 2 (2s delay, timer_id=%lu, HIGH priority)\n", (unsigned long)tid2);
    
    uint64_t timer_data3 = 3;
    timer_id_t tid3 = event_timer_post_delayed(timer, EVENT_TIMER, EVENT_PRIORITY_NORMAL,
                                                &timer_data3, sizeof(timer_data3), 3000, NULL);
    printf("[Main] Posted timer 3 (3s delay, timer_id=%lu)\n", (unsigned long)tid3);
    
    printf("[Main] Cancelling timer 2 immediately...\n");
    if (event_timer_cancel(timer, tid2) == 0) {
        printf("[Main] Timer 2 cancelled successfully (should NOT fire)\n");
    } else {
        printf("[Main] Failed to cancel timer 2\n");
    }
    
    printf("\n[Main] Waiting for timers (timer 1 at 1s, timer 3 at 3s)...\n");
    
    event_id_t quit_event_id;
    event_channel_post(channel, EVENT_QUIT, EVENT_PRIORITY_HIGHEST,
                        NULL, 0, &quit_event_id);
    
    printf("\n[Main] Test 4: Event cancellation\n");
    printf("[Main] Posted QUIT event (id=%lu) and will cancel it...\n", 
           (unsigned long)quit_event_id);
    
    if (event_channel_cancel(channel, quit_event_id) == 0) {
        printf("[Main] QUIT event cancelled successfully\n");
    } else {
        printf("[Main] Failed to cancel QUIT event\n");
    }
    
    printf("\n[Main] Posting QUIT timer (4s delay) and starting dispatcher loop...\n");
    printf("[Main] Expected order: timer 1 (1s) -> timer 3 (3s) -> QUIT (4s)\n");
    event_timer_post_delayed(timer, EVENT_QUIT, EVENT_PRIORITY_HIGHEST,
                              NULL, 0, 4000, NULL);
    
    printf("\n--- Dispatcher Loop Started (waiting for timers) ---\n");
    event_dispatcher_run(dispatcher);
    printf("--- Dispatcher Loop Stopped ---\n");
    
    printf("\n[Main] Cleaning up...\n");
    event_timer_destroy(timer);
    event_dispatcher_destroy(dispatcher);
    event_channel_destroy(channel);
    
    printf("\n=== Demo Complete ===\n");
    return 0;
}
