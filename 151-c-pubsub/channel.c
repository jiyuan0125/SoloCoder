#include "channel.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

int channel_manager_init(ChannelManager *cm, UserManager *um) {
    if (!cm || !um) return CHAT_ERR_INVALID_ARG;
    
    memset(cm, 0, sizeof(ChannelManager));
    
    if (pthread_mutex_init(&cm->mutex, NULL) != 0) {
        return CHAT_ERR_MEMORY;
    }
    
    cm->user_manager = um;
    cm->initialized = 1;
    
    return CHAT_OK;
}

void channel_manager_destroy(ChannelManager *cm) {
    if (!cm || !cm->initialized) return;
    
    pthread_mutex_lock(&cm->mutex);
    
    for (int i = 0; i < MAX_CHANNELS; i++) {
        if (cm->channels[i] != NULL) {
            Channel *ch = cm->channels[i];
            
            pthread_rwlock_wrlock(&ch->rwlock);
            ch->destroyed = 1;
            
            char sys_msg[256];
            snprintf(sys_msg, sizeof(sys_msg), "频道「%s」已被销毁", ch->name);
            
            User *sub_snapshot[MAX_SUBSCRIBERS];
            int sub_count = ch->subscriber_count;
            for (int j = 0; j < sub_count; j++) {
                sub_snapshot[j] = ch->subscribers[j];
            }
            
            pthread_rwlock_unlock(&ch->rwlock);
            
            for (int j = 0; j < sub_count; j++) {
                if (sub_snapshot[j]) {
                    Message *notif = message_create("SYSTEM", ch->name, sys_msg);
                    if (notif) {
                        user_deliver_message(sub_snapshot[j], notif);
                    }
                    user_unsubscribe(sub_snapshot[j], ch->name);
                }
            }
            
            pthread_rwlock_destroy(&ch->rwlock);
            free(ch);
            cm->channels[i] = NULL;
        }
    }
    
    cm->channel_count = 0;
    cm->initialized = 0;
    
    pthread_mutex_unlock(&cm->mutex);
    pthread_mutex_destroy(&cm->mutex);
}

Channel *channel_create(ChannelManager *cm, const char *channel_name) {
    if (!cm || !cm->initialized || !channel_name) return NULL;
    if (strlen(channel_name) >= CHANNEL_NAME_MAX) return NULL;
    
    pthread_mutex_lock(&cm->mutex);
    
    if (cm->channel_count >= MAX_CHANNELS) {
        pthread_mutex_unlock(&cm->mutex);
        return NULL;
    }
    
    for (int i = 0; i < MAX_CHANNELS; i++) {
        if (cm->channels[i] != NULL && strcmp(cm->channels[i]->name, channel_name) == 0) {
            pthread_mutex_unlock(&cm->mutex);
            return NULL;
        }
    }
    
    int idx = -1;
    for (int i = 0; i < MAX_CHANNELS; i++) {
        if (cm->channels[i] == NULL) {
            idx = i;
            break;
        }
    }
    
    if (idx == -1) {
        pthread_mutex_unlock(&cm->mutex);
        return NULL;
    }
    
    Channel *ch = (Channel *)malloc(sizeof(Channel));
    if (!ch) {
        pthread_mutex_unlock(&cm->mutex);
        return NULL;
    }
    
    memset(ch, 0, sizeof(Channel));
    strncpy(ch->name, channel_name, CHANNEL_NAME_MAX - 1);
    ch->subscriber_count = 0;
    ch->destroyed = 0;
    
    if (pthread_rwlock_init(&ch->rwlock, NULL) != 0) {
        free(ch);
        pthread_mutex_unlock(&cm->mutex);
        return NULL;
    }
    
    ch->active = 1;
    cm->channels[idx] = ch;
    cm->channel_count++;
    
    pthread_mutex_unlock(&cm->mutex);
    return ch;
}

int channel_destroy(ChannelManager *cm, const char *channel_name) {
    if (!cm || !cm->initialized || !channel_name) return CHAT_ERR_INVALID_ARG;
    
    pthread_mutex_lock(&cm->mutex);
    
    Channel *target_ch = NULL;
    int target_idx = -1;
    
    for (int i = 0; i < MAX_CHANNELS; i++) {
        if (cm->channels[i] != NULL && strcmp(cm->channels[i]->name, channel_name) == 0) {
            target_ch = cm->channels[i];
            target_idx = i;
            break;
        }
    }
    
    if (!target_ch) {
        pthread_mutex_unlock(&cm->mutex);
        return CHAT_ERR_CHANNEL_NOT_FOUND;
    }
    
    pthread_rwlock_wrlock(&target_ch->rwlock);
    target_ch->destroyed = 1;
    
    char sys_msg[256];
    snprintf(sys_msg, sizeof(sys_msg), "频道「%s」已被销毁", target_ch->name);
    
    User *sub_snapshot[MAX_SUBSCRIBERS];
    int sub_count = target_ch->subscriber_count;
    for (int j = 0; j < sub_count; j++) {
        sub_snapshot[j] = target_ch->subscribers[j];
    }
    
    pthread_rwlock_unlock(&target_ch->rwlock);
    
    for (int j = 0; j < sub_count; j++) {
        if (sub_snapshot[j]) {
            Message *notif = message_create("SYSTEM", target_ch->name, sys_msg);
            if (notif) {
                user_deliver_message(sub_snapshot[j], notif);
            }
            user_unsubscribe(sub_snapshot[j], target_ch->name);
        }
    }
    
    pthread_rwlock_destroy(&target_ch->rwlock);
    free(target_ch);
    cm->channels[target_idx] = NULL;
    cm->channel_count--;
    
    pthread_mutex_unlock(&cm->mutex);
    return CHAT_OK;
}

Channel *channel_find(ChannelManager *cm, const char *channel_name) {
    if (!cm || !cm->initialized || !channel_name) return NULL;
    
    pthread_mutex_lock(&cm->mutex);
    
    for (int i = 0; i < MAX_CHANNELS; i++) {
        if (cm->channels[i] != NULL && strcmp(cm->channels[i]->name, channel_name) == 0) {
            pthread_mutex_unlock(&cm->mutex);
            return cm->channels[i];
        }
    }
    
    pthread_mutex_unlock(&cm->mutex);
    return NULL;
}

int channel_exists(ChannelManager *cm, const char *channel_name) {
    return (channel_find(cm, channel_name) != NULL) ? 1 : 0;
}

int channel_add_subscriber(Channel *ch, User *user) {
    if (!ch || !user) return CHAT_ERR_INVALID_ARG;
    
    pthread_rwlock_wrlock(&ch->rwlock);
    
    if (ch->destroyed) {
        pthread_rwlock_unlock(&ch->rwlock);
        return CHAT_ERR_CHANNEL_DESTROYED;
    }
    
    if (ch->subscriber_count >= MAX_SUBSCRIBERS) {
        pthread_rwlock_unlock(&ch->rwlock);
        return CHAT_ERR_TOO_MANY_SUBSCRIPTIONS;
    }
    
    for (int i = 0; i < ch->subscriber_count; i++) {
        if (ch->subscribers[i] == user) {
            pthread_rwlock_unlock(&ch->rwlock);
            return CHAT_ERR_USER_ALREADY_SUBSCRIBED;
        }
    }
    
    ch->subscribers[ch->subscriber_count] = user;
    ch->subscriber_count++;
    
    pthread_rwlock_unlock(&ch->rwlock);
    
    int ret = user_subscribe(user, ch->name);
    if (ret != CHAT_OK && ret != CHAT_ERR_USER_ALREADY_SUBSCRIBED) {
        pthread_rwlock_wrlock(&ch->rwlock);
        for (int i = 0; i < ch->subscriber_count; i++) {
            if (ch->subscribers[i] == user) {
                if (i < ch->subscriber_count - 1) {
                    ch->subscribers[i] = ch->subscribers[ch->subscriber_count - 1];
                }
                ch->subscribers[ch->subscriber_count - 1] = NULL;
                ch->subscriber_count--;
                break;
            }
        }
        pthread_rwlock_unlock(&ch->rwlock);
        return ret;
    }
    
    return CHAT_OK;
}

int channel_remove_subscriber(Channel *ch, User *user) {
    if (!ch || !user) return CHAT_ERR_INVALID_ARG;
    
    pthread_rwlock_wrlock(&ch->rwlock);
    
    int found = 0;
    for (int i = 0; i < ch->subscriber_count; i++) {
        if (ch->subscribers[i] == user) {
            if (i < ch->subscriber_count - 1) {
                ch->subscribers[i] = ch->subscribers[ch->subscriber_count - 1];
            }
            ch->subscribers[ch->subscriber_count - 1] = NULL;
            ch->subscriber_count--;
            found = 1;
            break;
        }
    }
    
    pthread_rwlock_unlock(&ch->rwlock);
    
    if (!found) {
        return CHAT_ERR_USER_NOT_SUBSCRIBED;
    }
    
    user_unsubscribe(user, ch->name);
    
    return CHAT_OK;
}

int channel_has_subscriber(Channel *ch, User *user) {
    if (!ch || !user) return 0;
    
    pthread_rwlock_rdlock(&ch->rwlock);
    
    for (int i = 0; i < ch->subscriber_count; i++) {
        if (ch->subscribers[i] == user) {
            pthread_rwlock_unlock(&ch->rwlock);
            return 1;
        }
    }
    
    pthread_rwlock_unlock(&ch->rwlock);
    return 0;
}

int channel_get_subscribers(Channel *ch, User *subscribers[], int max_subscribers) {
    if (!ch || !subscribers || max_subscribers <= 0) return 0;
    
    pthread_rwlock_rdlock(&ch->rwlock);
    
    int count = (ch->subscriber_count < max_subscribers) ? ch->subscriber_count : max_subscribers;
    for (int i = 0; i < count; i++) {
        subscribers[i] = ch->subscribers[i];
    }
    
    pthread_rwlock_unlock(&ch->rwlock);
    return count;
}

int channel_subscriber_count(Channel *ch) {
    if (!ch) return 0;
    
    pthread_rwlock_rdlock(&ch->rwlock);
    int count = ch->subscriber_count;
    pthread_rwlock_unlock(&ch->rwlock);
    
    return count;
}

int channel_publish_message(Channel *ch, const char *sender, const char *content) {
    if (!ch || !sender || !content) return CHAT_ERR_INVALID_ARG;
    
    Message *msg = message_create(sender, ch->name, content);
    if (!msg) return CHAT_ERR_MEMORY;
    
    pthread_rwlock_rdlock(&ch->rwlock);
    
    if (ch->destroyed) {
        pthread_rwlock_unlock(&ch->rwlock);
        message_destroy(msg);
        return CHAT_ERR_CHANNEL_DESTROYED;
    }
    
    if (ch->subscriber_count == 0) {
        pthread_rwlock_unlock(&ch->rwlock);
        message_destroy(msg);
        return CHAT_OK;
    }
    
    User *sub_snapshot[MAX_SUBSCRIBERS];
    int sub_count = ch->subscriber_count;
    for (int i = 0; i < sub_count; i++) {
        sub_snapshot[i] = ch->subscribers[i];
    }
    
    pthread_rwlock_unlock(&ch->rwlock);
    
    int delivered = 0;
    for (int i = 0; i < sub_count; i++) {
        if (sub_snapshot[i]) {
            Message *copy = message_create(msg->sender, msg->channel, msg->content);
            if (copy) {
                if (user_deliver_message(sub_snapshot[i], copy) == CHAT_OK) {
                    delivered++;
                } else {
                    message_destroy(copy);
                }
            }
        }
    }
    
    message_destroy(msg);
    return CHAT_OK;
}

int channel_publish_system_message(Channel *ch, const char *content) {
    return channel_publish_message(ch, "SYSTEM", content);
}
