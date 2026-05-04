#ifndef SUBSCRIPTION_H
#define SUBSCRIPTION_H

#include "common.h"
#include "message.h"

typedef struct User {
    char name[USER_NAME_MAX];
    MessageQueue msg_queue;
    char subscribed_channels[MAX_USER_CHANNELS][CHANNEL_NAME_MAX];
    int subscription_count;
    pthread_mutex_t mutex;
    int active;
} User;

typedef struct UserManager {
    User *users[MAX_USERS];
    int user_count;
    pthread_mutex_t mutex;
    int initialized;
} UserManager;

int user_manager_init(UserManager *um);
void user_manager_destroy(UserManager *um);

User *user_create(UserManager *um, const char *username);
int user_destroy(UserManager *um, const char *username);
User *user_find(UserManager *um, const char *username);

int user_subscribe(User *user, const char *channel_name);
int user_unsubscribe(User *user, const char *channel_name);
int user_is_subscribed(User *user, const char *channel_name);
int user_get_subscriptions(User *user, char channels[][CHANNEL_NAME_MAX], int max_channels);

Message *user_receive_message(User *user);
Message *user_receive_message_timed(User *user, int timeout_ms);
int user_deliver_message(User *user, Message *msg);

#endif
