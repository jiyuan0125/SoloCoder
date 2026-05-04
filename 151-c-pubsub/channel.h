#ifndef CHANNEL_H
#define CHANNEL_H

#include "common.h"
#include "subscription.h"
#include "message.h"

#define MAX_SUBSCRIBERS 1024

typedef struct Channel {
    char name[CHANNEL_NAME_MAX];
    User *subscribers[MAX_SUBSCRIBERS];
    int subscriber_count;
    pthread_rwlock_t rwlock;
    int destroyed;
    int active;
} Channel;

typedef struct ChannelManager {
    Channel *channels[MAX_CHANNELS];
    int channel_count;
    pthread_mutex_t mutex;
    UserManager *user_manager;
    int initialized;
} ChannelManager;

int channel_manager_init(ChannelManager *cm, UserManager *um);
void channel_manager_destroy(ChannelManager *cm);

Channel *channel_create(ChannelManager *cm, const char *channel_name);
int channel_destroy(ChannelManager *cm, const char *channel_name);
Channel *channel_find(ChannelManager *cm, const char *channel_name);
int channel_exists(ChannelManager *cm, const char *channel_name);

int channel_add_subscriber(Channel *ch, User *user);
int channel_remove_subscriber(Channel *ch, User *user);
int channel_has_subscriber(Channel *ch, User *user);
int channel_get_subscribers(Channel *ch, User *subscribers[], int max_subscribers);
int channel_subscriber_count(Channel *ch);

int channel_publish_message(Channel *ch, const char *sender, const char *content);
int channel_publish_system_message(Channel *ch, const char *content);

#endif
