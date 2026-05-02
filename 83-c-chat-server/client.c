#include "server.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include <sys/socket.h>

extern Client *clients_head;
extern Room *rooms_head;
extern pthread_mutex_t clients_mutex;
extern pthread_mutex_t rooms_mutex;

Client *create_client(int socket) {
    Client *client = (Client *)malloc(sizeof(Client));
    if (!client) return NULL;

    client->socket = socket;
    client->username[0] = '\0';
    client->current_room = NULL;
    client->last_activity = time(NULL);
    client->is_authenticated = 0;
    client->next = NULL;

    pthread_mutex_lock(&clients_mutex);
    client->next = clients_head;
    clients_head = client;
    pthread_mutex_unlock(&clients_mutex);

    return client;
}

void remove_client(Client *client) {
    if (client == NULL) return;

    if (client->current_room != NULL) {
        Room *room = find_room(client->current_room);
        if (room != NULL) {
            remove_client_from_room(client, room);
        }
    }

    pthread_mutex_lock(&clients_mutex);

    if (clients_head == client) {
        clients_head = client->next;
    } else {
        Client *current = clients_head;
        while (current != NULL && current->next != client) {
            current = current->next;
        }
        if (current != NULL) {
            current->next = client->next;
        }
    }

    pthread_mutex_unlock(&clients_mutex);

    close(client->socket);
    if (client->current_room != NULL) {
        free(client->current_room);
    }
    free(client);
}

Client *find_client_by_username(const char *username) {
    if (username == NULL) return NULL;

    pthread_mutex_lock(&clients_mutex);

    Client *current = clients_head;
    while (current != NULL) {
        if (strcmp(current->username, username) == 0) {
            pthread_mutex_unlock(&clients_mutex);
            return current;
        }
        current = current->next;
    }

    pthread_mutex_unlock(&clients_mutex);
    return NULL;
}

void send_to_client(Client *client, const char *message) {
    if (client == NULL || message == NULL) return;
    send(client->socket, message, strlen(message), 0);
}

void send_error_to_client(Client *client, const char *description) {
    if (client == NULL || description == NULL) return;
    
    char error_msg[MAX_MSG_LEN];
    snprintf(error_msg, sizeof(error_msg), "ERROR: %s\n", description);
    send_to_client(client, error_msg);
}

void handle_enter_room(Client *client, char *command) {
    char *token = strtok(command, " ");
    if (token == NULL) {
        send_error_to_client(client, "Invalid ENTER_ROOM command");
        return;
    }

    token = strtok(NULL, " ");
    if (token == NULL) {
        send_error_to_client(client, "Room name required");
        return;
    }

    char room_name[MAX_ROOMNAME_LEN];
    strncpy(room_name, token, MAX_ROOMNAME_LEN - 1);
    room_name[MAX_ROOMNAME_LEN - 1] = '\0';

    token = strtok(NULL, " ");
    if (token == NULL) {
        send_error_to_client(client, "Username required");
        return;
    }

    char username[MAX_USERNAME_LEN];
    strncpy(username, token, MAX_USERNAME_LEN - 1);
    username[MAX_USERNAME_LEN - 1] = '\0';

    if (strlen(username) == 0) {
        send_error_to_client(client, "Invalid username");
        return;
    }

    if (find_client_by_username(username) != NULL && strcmp(client->username, username) != 0) {
        send_error_to_client(client, "Username already taken");
        return;
    }

    pthread_mutex_lock(&rooms_mutex);
    Room *room = find_room(room_name);

    if (room == NULL) {
        room = create_room(room_name, username);
        if (room == NULL) {
            pthread_mutex_unlock(&rooms_mutex);
            send_error_to_client(client, "Failed to create room");
            return;
        }
        room->next = rooms_head;
        rooms_head = room;
    }
    pthread_mutex_unlock(&rooms_mutex);

    strncpy(client->username, username, MAX_USERNAME_LEN - 1);
    client->username[MAX_USERNAME_LEN - 1] = '\0';

    add_client_to_room(client, room);

    char welcome_msg[MAX_MSG_LEN];
    snprintf(welcome_msg, sizeof(welcome_msg), "Welcome to room %s, %s!\n", room_name, username);
    send_to_client(client, welcome_msg);

    send_history_to_client(client, room);
}

void handle_leave(Client *client) {
    if (client->current_room == NULL) {
        send_error_to_client(client, "Not in any room");
        return;
    }

    Room *room = find_room(client->current_room);
    if (room == NULL) {
        send_error_to_client(client, "Room not found");
        return;
    }

    char leave_msg[MAX_MSG_LEN];
    snprintf(leave_msg, sizeof(leave_msg), "You have left room %s\n", room->name);
    send_to_client(client, leave_msg);

    remove_client_from_room(client, room);
}

void handle_msg(Client *client, char *command) {
    if (!client->is_authenticated) {
        send_error_to_client(client, "Not authenticated. Use ENTER_ROOM first");
        return;
    }

    char *token = strtok(command, " ");
    if (token == NULL) {
        send_error_to_client(client, "Invalid MSG command");
        return;
    }

    token = strtok(NULL, " ");
    if (token == NULL) {
        send_error_to_client(client, "Room name required");
        return;
    }

    char room_name[MAX_ROOMNAME_LEN];
    strncpy(room_name, token, MAX_ROOMNAME_LEN - 1);
    room_name[MAX_ROOMNAME_LEN - 1] = '\0';

    token = strtok(NULL, "");
    if (token == NULL) {
        send_error_to_client(client, "Message required");
        return;
    }

    Room *room = find_room(room_name);
    if (room == NULL) {
        send_error_to_client(client, "Room not found");
        return;
    }

    char formatted_msg[MAX_MSG_LEN + MAX_USERNAME_LEN + 10];
    snprintf(formatted_msg, sizeof(formatted_msg), "[%s] %s\n", client->username, token);
    
    broadcast_to_room(room, formatted_msg, NULL);
    add_message_to_history(room, client->username, token);
}

void handle_privmsg(Client *client, char *command) {
    if (!client->is_authenticated) {
        send_error_to_client(client, "Not authenticated. Use ENTER_ROOM first");
        return;
    }

    char *token = strtok(command, " ");
    if (token == NULL) {
        send_error_to_client(client, "Invalid PRIVMSG command");
        return;
    }

    token = strtok(NULL, " ");
    if (token == NULL) {
        send_error_to_client(client, "Username required");
        return;
    }

    char target_username[MAX_USERNAME_LEN];
    strncpy(target_username, token, MAX_USERNAME_LEN - 1);
    target_username[MAX_USERNAME_LEN - 1] = '\0';

    token = strtok(NULL, "");
    if (token == NULL) {
        send_error_to_client(client, "Message required");
        return;
    }

    Client *target = find_client_by_username(target_username);
    if (target == NULL) {
        send_error_to_client(client, "User not found");
        return;
    }

    char formatted_msg[MAX_MSG_LEN + MAX_USERNAME_LEN + 10];
    snprintf(formatted_msg, sizeof(formatted_msg), "[%s] (private) %s\n", client->username, token);
    
    send_to_client(target, formatted_msg);
}

void handle_list(Client *client) {
    list_rooms(client);
}

void handle_who(Client *client, char *command) {
    char *token = strtok(command, " ");
    if (token == NULL) {
        send_error_to_client(client, "Invalid WHO command");
        return;
    }

    token = strtok(NULL, " ");
    if (token == NULL) {
        send_error_to_client(client, "Room name required");
        return;
    }

    char room_name[MAX_ROOMNAME_LEN];
    strncpy(room_name, token, MAX_ROOMNAME_LEN - 1);
    room_name[MAX_ROOMNAME_LEN - 1] = '\0';

    list_room_members(client, room_name);
}

void handle_kick(Client *client, char *command) {
    if (!client->is_authenticated) {
        send_error_to_client(client, "Not authenticated. Use ENTER_ROOM first");
        return;
    }

    char *token = strtok(command, " ");
    if (token == NULL) {
        send_error_to_client(client, "Invalid KICK command");
        return;
    }

    token = strtok(NULL, " ");
    if (token == NULL) {
        send_error_to_client(client, "Room name required");
        return;
    }

    char room_name[MAX_ROOMNAME_LEN];
    strncpy(room_name, token, MAX_ROOMNAME_LEN - 1);
    room_name[MAX_ROOMNAME_LEN - 1] = '\0';

    token = strtok(NULL, " ");
    if (token == NULL) {
        send_error_to_client(client, "Username required");
        return;
    }

    char username[MAX_USERNAME_LEN];
    strncpy(username, token, MAX_USERNAME_LEN - 1);
    username[MAX_USERNAME_LEN - 1] = '\0';

    kick_user_from_room(client, room_name, username);
}

void handle_ping(Client *client) {
    send_to_client(client, "PONG\n");
    client->last_activity = time(NULL);
}

void handle_client_command(Client *client, char *command) {
    if (command == NULL || strlen(command) == 0) return;

    char command_copy[MAX_MSG_LEN];
    strncpy(command_copy, command, MAX_MSG_LEN - 1);
    command_copy[MAX_MSG_LEN - 1] = '\0';

    char *cmd = strtok(command_copy, " ");
    if (cmd == NULL) return;

    if (strcmp(cmd, "ENTER_ROOM") == 0) {
        handle_enter_room(client, command);
    } else if (strcmp(cmd, "LEAVE") == 0) {
        handle_leave(client);
    } else if (strcmp(cmd, "MSG") == 0) {
        handle_msg(client, command);
    } else if (strcmp(cmd, "PRIVMSG") == 0) {
        handle_privmsg(client, command);
    } else if (strcmp(cmd, "PING") == 0) {
        handle_ping(client);
    } else if (strcmp(cmd, "LIST") == 0) {
        handle_list(client);
    } else if (strcmp(cmd, "WHO") == 0) {
        handle_who(client, command);
    } else if (strcmp(cmd, "KICK") == 0) {
        handle_kick(client, command);
    } else {
        send_error_to_client(client, "Unknown command");
    }

    client->last_activity = time(NULL);
}

void *client_handler(void *arg) {
    if (arg == NULL) return NULL;

    Client *client = (Client *)arg;
    char buffer[MAX_MSG_LEN];
    char line_buffer[MAX_MSG_LEN * 2];
    size_t line_buffer_len = 0;
    ssize_t bytes_read;

    printf("Client connected on socket %d\n", client->socket);

    while (1) {
        bytes_read = recv(client->socket, buffer, sizeof(buffer) - 1, 0);
        
        if (bytes_read <= 0) {
            printf("Client disconnected on socket %d\n", client->socket);
            break;
        }

        buffer[bytes_read] = '\0';

        for (int i = 0; i < bytes_read; i++) {
            if (buffer[i] == '\n') {
                line_buffer[line_buffer_len] = '\0';
                
                if (line_buffer_len > 0 && line_buffer[line_buffer_len - 1] == '\r') {
                    line_buffer[line_buffer_len - 1] = '\0';
                }
                
                if (strlen(line_buffer) > 0) {
                    printf("Received from socket %d: %s\n", client->socket, line_buffer);
                    handle_client_command(client, line_buffer);
                }
                
                line_buffer_len = 0;
            } else {
                if (line_buffer_len < sizeof(line_buffer) - 1) {
                    line_buffer[line_buffer_len++] = buffer[i];
                }
            }
        }
    }

    remove_client(client);
    printf("Client handler exiting for socket %d\n", client->socket);
    return NULL;
}

void check_timeouts() {
    time_t now = time(NULL);
    
    pthread_mutex_lock(&clients_mutex);
    
    Client *current = clients_head;
    Client *next = NULL;
    
    while (current != NULL) {
        next = current->next;
        
        if (now - current->last_activity > TIMEOUT_SECONDS) {
            printf("Client %s timed out\n", current->username);
            
            pthread_mutex_unlock(&clients_mutex);
            remove_client(current);
            pthread_mutex_lock(&clients_mutex);
            
            current = clients_head;
            next = current ? current->next : NULL;
        } else {
            current = next;
        }
    }
    
    pthread_mutex_unlock(&clients_mutex);
}
