#include "server.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>

extern Room *rooms_head;
extern pthread_mutex_t rooms_mutex;
extern pthread_mutex_t clients_mutex;

Room *create_room(const char *name, const char *creator) {
    Room *room = (Room *)malloc(sizeof(Room));
    if (!room) return NULL;

    strncpy(room->name, name, MAX_ROOMNAME_LEN - 1);
    room->name[MAX_ROOMNAME_LEN - 1] = '\0';

    strncpy(room->creator, creator, MAX_USERNAME_LEN - 1);
    room->creator[MAX_USERNAME_LEN - 1] = '\0';

    room->clients = NULL;
    room->client_count = 0;
    room->history_count = 0;
    room->history_start = 0;
    room->is_active = 1;
    room->last_used = time(NULL);
    room->next = NULL;

    pthread_mutex_init(&room->mutex, NULL);

    return room;
}

Room *find_room(const char *name) {
    Room *current = rooms_head;
    while (current != NULL) {
        if (strcmp(current->name, name) == 0) {
            return current;
        }
        current = current->next;
    }
    return NULL;
}

void add_client_to_room(Client *client, Room *room) {
    pthread_mutex_lock(&room->mutex);

    if (client->current_room != NULL) {
        Room *old_room = find_room(client->current_room);
        if (old_room != NULL) {
            remove_client_from_room(client, old_room);
        }
    }

    client->current_room = strdup(room->name);
    client->is_authenticated = 1;

    Client *new_client = (Client *)malloc(sizeof(Client));
    if (new_client == NULL) {
        pthread_mutex_unlock(&room->mutex);
        return;
    }
    memcpy(new_client, client, sizeof(Client));
    new_client->next = room->clients;
    room->clients = new_client;
    room->client_count++;
    room->is_active = 1;
    room->last_used = time(NULL);

    pthread_mutex_unlock(&room->mutex);
}

void remove_client_from_room(Client *client, Room *room) {
    pthread_mutex_lock(&room->mutex);

    Client *current = room->clients;
    Client *prev = NULL;

    while (current != NULL) {
        if (strcmp(current->username, client->username) == 0) {
            if (prev == NULL) {
                room->clients = current->next;
            } else {
                prev->next = current->next;
            }
            free(current);
            room->client_count--;
            break;
        }
        prev = current;
        current = current->next;
    }

    if (client->current_room != NULL) {
        free(client->current_room);
        client->current_room = NULL;
    }

    if (room->client_count == 0) {
        room->is_active = 0;
        room->last_used = time(NULL);
    }

    pthread_mutex_unlock(&room->mutex);
}

void add_message_to_history(Room *room, const char *sender, const char *content) {
    pthread_mutex_lock(&room->mutex);

    int index = (room->history_start + room->history_count) % MAX_HISTORY;

    strncpy(room->history[index].sender, sender, MAX_USERNAME_LEN - 1);
    room->history[index].sender[MAX_USERNAME_LEN - 1] = '\0';

    strncpy(room->history[index].content, content, MAX_MSG_LEN - 1);
    room->history[index].content[MAX_MSG_LEN - 1] = '\0';

    room->history[index].timestamp = time(NULL);

    if (room->history_count < MAX_HISTORY) {
        room->history_count++;
    } else {
        room->history_start = (room->history_start + 1) % MAX_HISTORY;
    }

    room->last_used = time(NULL);

    pthread_mutex_unlock(&room->mutex);
}

void send_history_to_client(Client *client, Room *room) {
    pthread_mutex_lock(&room->mutex);

    if (room->history_count == 0) {
        pthread_mutex_unlock(&room->mutex);
        return;
    }

    int start_index;
    int count;

    if (room->history_count <= MAX_SEND_HISTORY) {
        start_index = room->history_start;
        count = room->history_count;
    } else {
        start_index = (room->history_start + room->history_count - MAX_SEND_HISTORY) % MAX_HISTORY;
        count = MAX_SEND_HISTORY;
    }

    for (int i = 0; i < count; i++) {
        int index = (start_index + i) % MAX_HISTORY;
        char msg_buffer[MAX_MSG_LEN + MAX_USERNAME_LEN + 10];
        snprintf(msg_buffer, sizeof(msg_buffer), "[%s] %s\n", 
                 room->history[index].sender, room->history[index].content);
        send(client->socket, msg_buffer, strlen(msg_buffer), 0);
    }

    pthread_mutex_unlock(&room->mutex);
}

void list_rooms(Client *client) {
    pthread_mutex_lock(&rooms_mutex);

    if (rooms_head == NULL) {
        send_to_client(client, "No rooms available\n");
        pthread_mutex_unlock(&rooms_mutex);
        return;
    }

    Room *current = rooms_head;
    char buffer[MAX_MSG_LEN];
    int total_length = 0;

    buffer[0] = '\0';

    while (current != NULL) {
        char room_info[MAX_ROOMNAME_LEN + 32];
        snprintf(room_info, sizeof(room_info), "Room: %s (%d users)\n", 
                 current->name, current->client_count);
        
        int info_len = strlen(room_info);
        if (total_length + info_len < MAX_MSG_LEN) {
            strcat(buffer, room_info);
            total_length += info_len;
        } else {
            send_to_client(client, buffer);
            buffer[0] = '\0';
            strcat(buffer, room_info);
            total_length = info_len;
        }
        
        current = current->next;
    }

    if (strlen(buffer) > 0) {
        send_to_client(client, buffer);
    }

    pthread_mutex_unlock(&rooms_mutex);
}

void list_room_members(Client *client, const char *room_name) {
    Room *room = find_room(room_name);
    if (room == NULL) {
        send_error_to_client(client, "Room not found");
        return;
    }

    pthread_mutex_lock(&room->mutex);

    if (room->client_count == 0) {
        send_to_client(client, "Room is empty\n");
        pthread_mutex_unlock(&room->mutex);
        return;
    }

    char buffer[MAX_MSG_LEN];
    int total_length = 0;
    buffer[0] = '\0';

    snprintf(buffer, sizeof(buffer), "Users in %s:\n", room_name);
    total_length = strlen(buffer);

    Client *current = room->clients;
    while (current != NULL) {
        char user_info[MAX_USERNAME_LEN + 16];
        snprintf(user_info, sizeof(user_info), "  - %s\n", current->username);
        
        int info_len = strlen(user_info);
        if (total_length + info_len < MAX_MSG_LEN) {
            strcat(buffer, user_info);
            total_length += info_len;
        } else {
            send_to_client(client, buffer);
            buffer[0] = '\0';
            strcat(buffer, user_info);
            total_length = info_len;
        }
        
        current = current->next;
    }

    if (strlen(buffer) > 0) {
        send_to_client(client, buffer);
    }

    pthread_mutex_unlock(&room->mutex);
}

int kick_user_from_room(Client *kicker, const char *room_name, const char *username) {
    Room *room = find_room(room_name);
    if (room == NULL) {
        send_error_to_client(kicker, "Room not found");
        return -1;
    }

    pthread_mutex_lock(&room->mutex);

    if (strcmp(room->creator, kicker->username) != 0) {
        send_error_to_client(kicker, "Only room creator can kick users");
        pthread_mutex_unlock(&room->mutex);
        return -1;
    }

    if (strcmp(username, kicker->username) == 0) {
        send_error_to_client(kicker, "Cannot kick yourself");
        pthread_mutex_unlock(&room->mutex);
        return -1;
    }

    Client *current = room->clients;
    Client *to_kick = NULL;

    while (current != NULL) {
        if (strcmp(current->username, username) == 0) {
            to_kick = current;
            break;
        }
        current = current->next;
    }

    if (to_kick == NULL) {
        send_error_to_client(kicker, "User not in room");
        pthread_mutex_unlock(&room->mutex);
        return -1;
    }

    pthread_mutex_unlock(&room->mutex);

    Client *real_client = find_client_by_username(username);
    if (real_client != NULL) {
        char kick_msg[MAX_MSG_LEN];
        snprintf(kick_msg, sizeof(kick_msg), "You have been kicked from room %s\n", room_name);
        send_to_client(real_client, kick_msg);
        
        remove_client_from_room(real_client, room);
    } else {
        remove_client_from_room(to_kick, room);
    }

    char success_msg[MAX_MSG_LEN];
    snprintf(success_msg, sizeof(success_msg), "User %s has been kicked from room %s\n", 
             username, room_name);
    send_to_client(kicker, success_msg);

    return 0;
}

void broadcast_to_room(Room *room, const char *message, Client *exclude) {
    if (room == NULL) return;

    pthread_mutex_lock(&room->mutex);

    Client *current = room->clients;
    while (current != NULL) {
        if (exclude == NULL || strcmp(current->username, exclude->username) != 0) {
            send(current->socket, message, strlen(message), 0);
        }
        current = current->next;
    }

    pthread_mutex_unlock(&room->mutex);
}
