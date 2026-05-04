#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/select.h>
#include <signal.h>
#include <ctype.h>
#include <errno.h>

#include "process_manager.h"
#include "message_frame.h"
#include "pipe_manager.h"

typedef enum {
    MSG_TYPE_ECHO = 1,
    MSG_TYPE_UPPER,
    MSG_TYPE_EXIT,
    MSG_TYPE_RESULT
} message_type_t;

typedef struct {
    uint8_t type;
    uint8_t data[1024];
} app_message_t;

static void worker_main(int read_fd, int write_fd) {
    message_buffer_t recv_buffer;
    
    if (message_buffer_init(&recv_buffer, 4096) != 0) {
        _exit(1);
    }
    
    printf("[Worker %d] Started, waiting for messages...\n", getpid());
    
    while (1) {
        fd_set readfds;
        struct timeval tv;
        int ret;
        
        FD_ZERO(&readfds);
        FD_SET(read_fd, &readfds);
        tv.tv_sec = 5;
        tv.tv_usec = 0;
        
        ret = select(read_fd + 1, &readfds, NULL, NULL, &tv);
        
        if (ret == 0) {
            continue;
        }
        
        if (ret < 0) {
            if (errno == EINTR) continue;
            break;
        }
        
        ssize_t read_result = pipe_read_partial(read_fd, &recv_buffer);
        
        if (read_result == PIPE_EOF) {
            printf("[Worker %d] Pipe closed, exiting...\n", getpid());
            break;
        }
        
        if (read_result == PIPE_AGAIN) {
            continue;
        }
        
        if (read_result < 0) {
            printf("[Worker %d] Read error, exiting...\n", getpid());
            break;
        }
        
        while (message_buffer_has_complete_message(&recv_buffer)) {
            uint8_t *msg_data = NULL;
            size_t msg_len = 0;
            
            if (message_buffer_extract_message(&recv_buffer, &msg_data, &msg_len) != 0) {
                continue;
            }
            
            if (msg_len < 1) {
                free(msg_data);
                continue;
            }
            
            app_message_t response;
            memset(&response, 0, sizeof(response));
            
            uint8_t msg_type = msg_data[0];
            
            switch (msg_type) {
                case MSG_TYPE_ECHO:
                    printf("[Worker %d] Received ECHO request\n", getpid());
                    response.type = MSG_TYPE_RESULT;
                    if (msg_len > 1) {
                        memcpy(response.data, msg_data + 1, 
                               (msg_len - 1 > 1023) ? 1023 : (msg_len - 1));
                    }
                    printf("[Worker %d] Sending echo response: %s\n", 
                           getpid(), response.data);
                    pipe_write_message(write_fd, (uint8_t *)&response, 
                                       1 + strlen((char *)response.data) + 1);
                    break;
                    
                case MSG_TYPE_UPPER:
                    printf("[Worker %d] Received UPPER request\n", getpid());
                    response.type = MSG_TYPE_RESULT;
                    if (msg_len > 1) {
                        size_t data_len = (msg_len - 1 > 1023) ? 1023 : (msg_len - 1);
                        for (size_t i = 0; i < data_len; i++) {
                            response.data[i] = (uint8_t)toupper(msg_data[1 + i]);
                        }
                    }
                    printf("[Worker %d] Sending upper response: %s\n", 
                           getpid(), response.data);
                    pipe_write_message(write_fd, (uint8_t *)&response, 
                                       1 + strlen((char *)response.data) + 1);
                    break;
                    
                case MSG_TYPE_EXIT:
                    printf("[Worker %d] Received EXIT command, exiting...\n", getpid());
                    free(msg_data);
                    message_buffer_destroy(&recv_buffer);
                    _exit(0);
                    
                default:
                    printf("[Worker %d] Unknown message type: %d\n", getpid(), msg_type);
                    break;
            }
            
            free(msg_data);
        }
    }
    
    message_buffer_destroy(&recv_buffer);
}

int main(int argc, char *argv[]) {
    (void)argc;
    (void)argv;
    
    process_manager_t pm;
    const int NUM_WORKERS = 3;
    int workers_created = 0;
    
    printf("[Master] Starting master process, PID: %d\n", getpid());
    
    if (process_manager_init(&pm) != 0) {
        fprintf(stderr, "[Master] Failed to initialize process manager\n");
        return 1;
    }
    
    if (process_manager_setup_sigchld(&pm) != 0) {
        fprintf(stderr, "[Master] Failed to setup SIGCHLD handler\n");
        process_manager_destroy(&pm);
        return 1;
    }
    
    printf("[Master] Creating %d worker processes...\n", NUM_WORKERS);
    for (int i = 0; i < NUM_WORKERS; i++) {
        int result = process_manager_create_worker(&pm, worker_main);
        if (result < 0) {
            fprintf(stderr, "[Master] Failed to create worker %d\n", i);
            continue;
        }
        workers_created++;
        printf("[Master] Created worker %d, PID: %d\n", i, 
               pm.workers[result].pid);
    }
    
    if (workers_created == 0) {
        fprintf(stderr, "[Master] No workers created, exiting\n");
        process_manager_destroy(&pm);
        return 1;
    }
    
    sleep(1);
    
    printf("\n[Master] Sending test messages to workers...\n");
    const char *test_messages[] = {
        "hello world",
        "test message",
        "pipe communication"
    };
    
    for (size_t i = 0; i < pm.worker_count; i++) {
        worker_process_t *worker = process_manager_get_worker(&pm, i);
        if (worker == NULL) continue;
        
        app_message_t msg;
        msg.type = (i % 2 == 0) ? MSG_TYPE_ECHO : MSG_TYPE_UPPER;
        strcpy((char *)msg.data, test_messages[i % 3]);
        
        printf("[Master] Sending to worker %zu (PID %d): type=%s, data='%s'\n",
               i, worker->pid,
               (msg.type == MSG_TYPE_ECHO) ? "ECHO" : "UPPER",
               msg.data);
        
        ssize_t send_result = worker_send_message(worker, (uint8_t *)&msg, 
                                                   1 + strlen((char *)msg.data) + 1);
        if (send_result < 0) {
            fprintf(stderr, "[Master] Failed to send message to worker %zu\n", i);
        }
    }
    
    printf("\n[Master] Waiting for responses...\n");
    
    int responses_received = 0;
    int max_wait_loops = 50;
    
    for (int loop = 0; loop < max_wait_loops && responses_received < workers_created; loop++) {
        if (pm.sigchld_received) {
            printf("[Master] SIGCHLD received, checking children...\n");
            process_manager_check_children(&pm);
            process_manager_cleanup_exited(&pm);
            pm.sigchld_received = 0;
        }
        
        fd_set readfds;
        struct timeval tv;
        int max_fd = -1;
        
        FD_ZERO(&readfds);
        
        for (size_t i = 0; i < pm.worker_count; i++) {
            worker_process_t *worker = process_manager_get_worker(&pm, i);
            if (worker == NULL || worker->state != WORKER_RUNNING) continue;
            
            int fd = worker->worker_to_master.read_fd;
            if (fd >= 0) {
                FD_SET(fd, &readfds);
                if (fd > max_fd) max_fd = fd;
            }
        }
        
        if (max_fd < 0) {
            usleep(100000);
            continue;
        }
        
        tv.tv_sec = 0;
        tv.tv_usec = 100000;
        
        int select_result = select(max_fd + 1, &readfds, NULL, NULL, &tv);
        
        if (select_result < 0) {
            if (errno == EINTR) continue;
            perror("[Master] select failed");
            break;
        }
        
        if (select_result == 0) {
            continue;
        }
        
        for (size_t i = 0; i < pm.worker_count; i++) {
            worker_process_t *worker = process_manager_get_worker(&pm, i);
            if (worker == NULL || worker->state != WORKER_RUNNING) continue;
            
            int fd = worker->worker_to_master.read_fd;
            if (fd >= 0 && FD_ISSET(fd, &readfds)) {
                uint8_t *response_data = NULL;
                size_t response_len = 0;
                
                ssize_t recv_result = worker_recv_message(worker, &response_data, &response_len);
                
                if (recv_result == PIPE_EOF) {
                    printf("[Master] Worker %zu (PID %d) closed pipe\n", 
                           i, worker->pid);
                    worker->state = WORKER_EXITED;
                } else if (recv_result == PIPE_AGAIN) {
                    continue;
                } else if (recv_result < 0) {
                    fprintf(stderr, "[Master] Error receiving from worker %zu\n", i);
                } else if (response_data != NULL && response_len >= 1) {
                    printf("[Master] Received response from worker %zu (PID %d): type=%d, data='%s'\n",
                           i, worker->pid,
                           response_data[0],
                           (response_len > 1) ? (char *)&response_data[1] : "");
                    responses_received++;
                    free(response_data);
                }
            }
        }
    }
    
    printf("\n[Master] Sending EXIT command to all workers...\n");
    for (size_t i = 0; i < pm.worker_count; i++) {
        worker_process_t *worker = process_manager_get_worker(&pm, i);
        if (worker == NULL || worker->state != WORKER_RUNNING) continue;
        
        app_message_t exit_msg;
        exit_msg.type = MSG_TYPE_EXIT;
        
        printf("[Master] Sending EXIT to worker %zu (PID %d)\n", i, worker->pid);
        worker_send_message(worker, (uint8_t *)&exit_msg, 1);
    }
    
    sleep(1);
    
    printf("[Master] Cleaning up...\n");
    process_manager_check_children(&pm);
    process_manager_destroy(&pm);
    
    printf("[Master] Done.\n");
    return 0;
}
