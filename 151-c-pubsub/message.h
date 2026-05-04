#ifndef MESSAGE_H
#define MESSAGE_H

#include "common.h"
#include <time.h>

typedef struct Message {
    char sender[USER_NAME_MAX];
    char channel[CHANNEL_NAME_MAX];
    char content[MESSAGE_CONTENT_MAX];
    time_t timestamp;
    struct Message *next;
} Message;

typedef struct MessageQueue {
    Message *head;
    Message *tail;
    pthread_mutex_t mutex;
    pthread_cond_t cond;
    int count;
    int destroyed;
} MessageQueue;

Message *message_create(const char *sender, const char *channel, const char *content);
void message_destroy(Message *msg);

int message_queue_init(MessageQueue *mq);
void message_queue_destroy(MessageQueue *mq);
int message_queue_push(MessageQueue *mq, Message *msg);
Message *message_queue_pop(MessageQueue *mq);
Message *message_queue_timedpop(MessageQueue *mq, int timeout_ms);
int message_queue_size(MessageQueue *mq);
void message_queue_wakeup(MessageQueue *mq);

#endif
