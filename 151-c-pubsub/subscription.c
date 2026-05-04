#include "subscription.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

int user_manager_init(UserManager *um) {
    if (!um) return CHAT_ERR_INVALID_ARG;
    
    memset(um, 0, sizeof(UserManager));
    
    if (pthread_mutex_init(&um->mutex, NULL) != 0) {
        return CHAT_ERR_MEMORY;
    }
    
    um->initialized = 1;
    return CHAT_OK;
}

void user_manager_destroy(UserManager *um) {
    if (!um || !um->initialized) return;
    
    pthread_mutex_lock(&um->mutex);
    
    for (int i = 0; i < MAX_USERS; i++) {
        if (um->users[i] != NULL) {
            pthread_mutex_lock(&um->users[i]->mutex);
            message_queue_destroy(&um->users[i]->msg_queue);
            pthread_mutex_unlock(&um->users[i]->mutex);
            pthread_mutex_destroy(&um->users[i]->mutex);
            free(um->users[i]);
            um->users[i] = NULL;
        }
    }
    
    um->user_count = 0;
    um->initialized = 0;
    
    pthread_mutex_unlock(&um->mutex);
    pthread_mutex_destroy(&um->mutex);
}

User *user_create(UserManager *um, const char *username) {
    if (!um || !um->initialized || !username) return NULL;
    if (strlen(username) >= USER_NAME_MAX) return NULL;
    
    pthread_mutex_lock(&um->mutex);
    
    if (um->user_count >= MAX_USERS) {
        pthread_mutex_unlock(&um->mutex);
        return NULL;
    }
    
    for (int i = 0; i < MAX_USERS; i++) {
        if (um->users[i] != NULL && strcmp(um->users[i]->name, username) == 0) {
            pthread_mutex_unlock(&um->mutex);
            return NULL;
        }
    }
    
    int idx = -1;
    for (int i = 0; i < MAX_USERS; i++) {
        if (um->users[i] == NULL) {
            idx = i;
            break;
        }
    }
    
    if (idx == -1) {
        pthread_mutex_unlock(&um->mutex);
        return NULL;
    }
    
    User *user = (User *)malloc(sizeof(User));
    if (!user) {
        pthread_mutex_unlock(&um->mutex);
        return NULL;
    }
    
    memset(user, 0, sizeof(User));
    strncpy(user->name, username, USER_NAME_MAX - 1);
    user->subscription_count = 0;
    
    if (pthread_mutex_init(&user->mutex, NULL) != 0) {
        free(user);
        pthread_mutex_unlock(&um->mutex);
        return NULL;
    }
    
    if (message_queue_init(&user->msg_queue) != CHAT_OK) {
        pthread_mutex_destroy(&user->mutex);
        free(user);
        pthread_mutex_unlock(&um->mutex);
        return NULL;
    }
    
    user->active = 1;
    um->users[idx] = user;
    um->user_count++;
    
    pthread_mutex_unlock(&um->mutex);
    return user;
}

int user_destroy(UserManager *um, const char *username) {
    if (!um || !um->initialized || !username) return CHAT_ERR_INVALID_ARG;
    
    pthread_mutex_lock(&um->mutex);
    
    for (int i = 0; i < MAX_USERS; i++) {
        if (um->users[i] != NULL && strcmp(um->users[i]->name, username) == 0) {
            User *user = um->users[i];
            pthread_mutex_lock(&user->mutex);
            message_queue_destroy(&user->msg_queue);
            pthread_mutex_unlock(&user->mutex);
            pthread_mutex_destroy(&user->mutex);
            free(user);
            um->users[i] = NULL;
            um->user_count--;
            pthread_mutex_unlock(&um->mutex);
            return CHAT_OK;
        }
    }
    
    pthread_mutex_unlock(&um->mutex);
    return CHAT_ERR_USER_NOT_FOUND;
}

User *user_find(UserManager *um, const char *username) {
    if (!um || !um->initialized || !username) return NULL;
    
    pthread_mutex_lock(&um->mutex);
    
    for (int i = 0; i < MAX_USERS; i++) {
        if (um->users[i] != NULL && strcmp(um->users[i]->name, username) == 0) {
            pthread_mutex_unlock(&um->mutex);
            return um->users[i];
        }
    }
    
    pthread_mutex_unlock(&um->mutex);
    return NULL;
}

int user_subscribe(User *user, const char *channel_name) {
    if (!user || !channel_name) return CHAT_ERR_INVALID_ARG;
    if (strlen(channel_name) >= CHANNEL_NAME_MAX) return CHAT_ERR_INVALID_ARG;
    
    pthread_mutex_lock(&user->mutex);
    
    if (user->subscription_count >= MAX_USER_CHANNELS) {
        pthread_mutex_unlock(&user->mutex);
        return CHAT_ERR_TOO_MANY_SUBSCRIPTIONS;
    }
    
    for (int i = 0; i < user->subscription_count; i++) {
        if (strcmp(user->subscribed_channels[i], channel_name) == 0) {
            pthread_mutex_unlock(&user->mutex);
            return CHAT_ERR_USER_ALREADY_SUBSCRIBED;
        }
    }
    
    strncpy(user->subscribed_channels[user->subscription_count], channel_name, CHANNEL_NAME_MAX - 1);
    user->subscription_count++;
    
    pthread_mutex_unlock(&user->mutex);
    return CHAT_OK;
}

int user_unsubscribe(User *user, const char *channel_name) {
    if (!user || !channel_name) return CHAT_ERR_INVALID_ARG;
    
    pthread_mutex_lock(&user->mutex);
    
    for (int i = 0; i < user->subscription_count; i++) {
        if (strcmp(user->subscribed_channels[i], channel_name) == 0) {
            if (i < user->subscription_count - 1) {
                strncpy(user->subscribed_channels[i], 
                        user->subscribed_channels[user->subscription_count - 1], 
                        CHANNEL_NAME_MAX - 1);
            }
            memset(user->subscribed_channels[user->subscription_count - 1], 0, CHANNEL_NAME_MAX);
            user->subscription_count--;
            pthread_mutex_unlock(&user->mutex);
            return CHAT_OK;
        }
    }
    
    pthread_mutex_unlock(&user->mutex);
    return CHAT_ERR_USER_NOT_SUBSCRIBED;
}

int user_is_subscribed(User *user, const char *channel_name) {
    if (!user || !channel_name) return 0;
    
    pthread_mutex_lock(&user->mutex);
    
    for (int i = 0; i < user->subscription_count; i++) {
        if (strcmp(user->subscribed_channels[i], channel_name) == 0) {
            pthread_mutex_unlock(&user->mutex);
            return 1;
        }
    }
    
    pthread_mutex_unlock(&user->mutex);
    return 0;
}

int user_get_subscriptions(User *user, char channels[][CHANNEL_NAME_MAX], int max_channels) {
    if (!user || !channels || max_channels <= 0) return 0;
    
    pthread_mutex_lock(&user->mutex);
    
    int count = (user->subscription_count < max_channels) ? user->subscription_count : max_channels;
    for (int i = 0; i < count; i++) {
        strncpy(channels[i], user->subscribed_channels[i], CHANNEL_NAME_MAX - 1);
        channels[i][CHANNEL_NAME_MAX - 1] = '\0';
    }
    
    pthread_mutex_unlock(&user->mutex);
    return count;
}

Message *user_receive_message(User *user) {
    if (!user) return NULL;
    return message_queue_pop(&user->msg_queue);
}

Message *user_receive_message_timed(User *user, int timeout_ms) {
    if (!user) return NULL;
    return message_queue_timedpop(&user->msg_queue, timeout_ms);
}

int user_deliver_message(User *user, Message *msg) {
    if (!user || !msg) {
        message_destroy(msg);
        return CHAT_ERR_INVALID_ARG;
    }
    
    int ret = message_queue_push(&user->msg_queue, msg);
    if (ret != CHAT_OK) {
        message_destroy(msg);
    }
    
    return ret;
}
