#ifndef CHAT_SERVER_H
#define CHAT_SERVER_H

#include <pthread.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <time.h>

#define MAX_USERNAME_LEN 64
#define MAX_ROOMNAME_LEN 64
#define MAX_MSG_LEN 1024
#define MAX_HISTORY 100
#define MAX_SEND_HISTORY 50

#define HEARTBEAT_INTERVAL 30
#define TIMEOUT_SECONDS 60

typedef struct message {
    char sender[MAX_USERNAME_LEN];
    char content[MAX_MSG_LEN];
    time_t timestamp;
} Message;

typedef struct client {
    int socket;
    char username[MAX_USERNAME_LEN];
    char *current_room;
    time_t last_activity;
    pthread_t thread;
    int is_authenticated;
    struct client *next;
} Client;

typedef struct room {
    char name[MAX_ROOMNAME_LEN];
    char creator[MAX_USERNAME_LEN];
    Client *clients;
    Message history[MAX_HISTORY];
    int history_count;
    int history_start;
    int client_count;
    int is_active;
    time_t last_used;
    pthread_mutex_t mutex;
    struct room *next;
} Room;

extern Client *clients_head;
extern Room *rooms_head;
extern pthread_mutex_t clients_mutex;
extern pthread_mutex_t rooms_mutex;

Client *create_client(int socket);
void remove_client(Client *client);
Client *find_client_by_username(const char *username);
void *client_handler(void *arg);
void handle_client_command(Client *client, char *command);

Room *create_room(const char *name, const char *creator);
Room *find_room(const char *name);
void add_client_to_room(Client *client, Room *room);
void remove_client_from_room(Client *client, Room *room);
void add_message_to_history(Room *room, const char *sender, const char *content);
void send_history_to_client(Client *client, Room *room);
void list_rooms(Client *client);
void list_room_members(Client *client, const char *room_name);
int kick_user_from_room(Client *kicker, const char *room_name, const char *username);

void broadcast_to_room(Room *room, const char *message, Client *exclude);
void send_to_client(Client *client, const char *message);
void send_error_to_client(Client *client, const char *description);

void check_timeouts();

#endif
