#include "server.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <errno.h>

#define PORT 8080
#define BUFFER_SIZE 1024
#define BACKLOG 10

Client *clients_head = NULL;
Room *rooms_head = NULL;
pthread_mutex_t clients_mutex = PTHREAD_MUTEX_INITIALIZER;
pthread_mutex_t rooms_mutex = PTHREAD_MUTEX_INITIALIZER;

int server_socket = -1;
int running = 1;

void cleanup() {
    running = 0;

    if (server_socket != -1) {
        close(server_socket);
        server_socket = -1;
    }

    pthread_mutex_lock(&clients_mutex);
    while (clients_head != NULL) {
        Client *temp = clients_head;
        clients_head = clients_head->next;
        close(temp->socket);
        if (temp->current_room != NULL) {
            free(temp->current_room);
        }
        free(temp);
    }
    pthread_mutex_unlock(&clients_mutex);

    pthread_mutex_lock(&rooms_mutex);
    while (rooms_head != NULL) {
        Room *temp = rooms_head;
        rooms_head = rooms_head->next;
        while (temp->clients != NULL) {
            Client *client_temp = temp->clients;
            temp->clients = temp->clients->next;
            free(client_temp);
        }
        pthread_mutex_destroy(&temp->mutex);
        free(temp);
    }
    pthread_mutex_unlock(&rooms_mutex);

    pthread_mutex_destroy(&clients_mutex);
    pthread_mutex_destroy(&rooms_mutex);

    printf("Server cleaned up and exiting\n");
}

void signal_handler(int sig) {
    printf("\nReceived signal %d, shutting down...\n", sig);
    cleanup();
    exit(EXIT_SUCCESS);
}

void *timeout_checker(void *arg) {
    (void)arg;
    printf("Timeout checker thread started\n");

    while (running) {
        sleep(10);
        if (!running) break;
        check_timeouts();
    }

    printf("Timeout checker thread exiting\n");
    return NULL;
}

int main() {
    struct sockaddr_in server_addr, client_addr;
    socklen_t client_len;
    pthread_t timeout_thread;

    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);
    atexit(cleanup);

    server_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (server_socket == -1) {
        perror("socket");
        exit(EXIT_FAILURE);
    }

    int opt = 1;
    if (setsockopt(server_socket, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) == -1) {
        perror("setsockopt");
        close(server_socket);
        exit(EXIT_FAILURE);
    }

    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = INADDR_ANY;
    server_addr.sin_port = htons(PORT);

    if (bind(server_socket, (struct sockaddr *)&server_addr, sizeof(server_addr)) == -1) {
        perror("bind");
        close(server_socket);
        exit(EXIT_FAILURE);
    }

    if (listen(server_socket, BACKLOG) == -1) {
        perror("listen");
        close(server_socket);
        exit(EXIT_FAILURE);
    }

    printf("TCP Chat Server starting on port %d...\n", PORT);
    printf("Server is running. Press Ctrl+C to stop.\n");

    if (pthread_create(&timeout_thread, NULL, timeout_checker, NULL) != 0) {
        perror("pthread_create timeout_checker");
        close(server_socket);
        exit(EXIT_FAILURE);
    }

    while (running) {
        client_len = sizeof(client_addr);
        int client_socket = accept(server_socket, (struct sockaddr *)&client_addr, &client_len);

        if (client_socket == -1) {
            if (running) {
                perror("accept");
            }
            continue;
        }

        printf("New connection from %s:%d\n", 
               inet_ntoa(client_addr.sin_addr), ntohs(client_addr.sin_port));

        Client *client = create_client(client_socket);
        if (client == NULL) {
            printf("Failed to create client structure\n");
            close(client_socket);
            continue;
        }

        if (pthread_create(&client->thread, NULL, client_handler, (void *)client) != 0) {
            perror("pthread_create client_handler");
            remove_client(client);
            continue;
        }

        pthread_detach(client->thread);
    }

    printf("Server shutting down...\n");
    return 0;
}
