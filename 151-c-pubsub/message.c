#include "message.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>

Message *message_create(const char *sender, const char *channel, const char *content) {
    if (!sender || !channel || !content) return NULL;
    
    Message *msg = (Message *)malloc(sizeof(Message));
    if (!msg) return NULL;
    
    memset(msg, 0, sizeof(Message));
    strncpy(msg->sender, sender, USER_NAME_MAX - 1);
    strncpy(msg->channel, channel, CHANNEL_NAME_MAX - 1);
    strncpy(msg->content, content, MESSAGE_CONTENT_MAX - 1);
    msg->timestamp = time(NULL);
    msg->next = NULL;
    
    return msg;
}

void message_destroy(Message *msg) {
    if (msg) {
        free(msg);
    }
}

int message_queue_init(MessageQueue *mq) {
    if (!mq) return CHAT_ERR_INVALID_ARG;
    
    mq->head = NULL;
    mq->tail = NULL;
    mq->count = 0;
    mq->destroyed = 0;
    
    if (pthread_mutex_init(&mq->mutex, NULL) != 0) {
        return CHAT_ERR_MEMORY;
    }
    if (pthread_cond_init(&mq->cond, NULL) != 0) {
        pthread_mutex_destroy(&mq->mutex);
        return CHAT_ERR_MEMORY;
    }
    
    return CHAT_OK;
}

void message_queue_destroy(MessageQueue *mq) {
    if (!mq) return;
    
    pthread_mutex_lock(&mq->mutex);
    mq->destroyed = 1;
    
    while (mq->head != NULL) {
        Message *tmp = mq->head;
        mq->head = mq->head->next;
        message_destroy(tmp);
    }
    mq->tail = NULL;
    mq->count = 0;
    
    pthread_cond_broadcast(&mq->cond);
    pthread_mutex_unlock(&mq->mutex);
    
    pthread_cond_destroy(&mq->cond);
    pthread_mutex_destroy(&mq->mutex);
}

int message_queue_push(MessageQueue *mq, Message *msg) {
    if (!mq || !msg) return CHAT_ERR_INVALID_ARG;
    
    pthread_mutex_lock(&mq->mutex);
    
    if (mq->destroyed) {
        pthread_mutex_unlock(&mq->mutex);
        return CHAT_ERR_CHANNEL_DESTROYED;
    }
    
    msg->next = NULL;
    if (mq->tail == NULL) {
        mq->head = msg;
        mq->tail = msg;
    } else {
        mq->tail->next = msg;
        mq->tail = msg;
    }
    mq->count++;
    
    pthread_cond_signal(&mq->cond);
    pthread_mutex_unlock(&mq->mutex);
    
    return CHAT_OK;
}

Message *message_queue_pop(MessageQueue *mq) {
    if (!mq) return NULL;
    
    pthread_mutex_lock(&mq->mutex);
    
    while (mq->head == NULL && !mq->destroyed) {
        pthread_cond_wait(&mq->cond, &mq->mutex);
    }
    
    if (mq->destroyed) {
        pthread_mutex_unlock(&mq->mutex);
        return NULL;
    }
    
    Message *msg = mq->head;
    mq->head = mq->head->next;
    if (mq->head == NULL) {
        mq->tail = NULL;
    }
    mq->count--;
    
    pthread_mutex_unlock(&mq->mutex);
    
    return msg;
}

Message *message_queue_timedpop(MessageQueue *mq, int timeout_ms) {
    if (!mq) return NULL;
    
    pthread_mutex_lock(&mq->mutex);
    
    if (mq->head == NULL && !mq->destroyed) {
        struct timespec ts;
        if (clock_gettime(CLOCK_REALTIME, &ts) != 0) {
            pthread_mutex_unlock(&mq->mutex);
            return NULL;
        }
        ts.tv_sec += timeout_ms / 1000;
        ts.tv_nsec += (timeout_ms % 1000) * 1000000;
        if (ts.tv_nsec >= 1000000000) {
            ts.tv_sec++;
            ts.tv_nsec -= 1000000000;
        }
        
        int ret = pthread_cond_timedwait(&mq->cond, &mq->mutex, &ts);
        if (ret == ETIMEDOUT) {
            pthread_mutex_unlock(&mq->mutex);
            return NULL;
        }
    }
    
    if (mq->destroyed || mq->head == NULL) {
        pthread_mutex_unlock(&mq->mutex);
        return NULL;
    }
    
    Message *msg = mq->head;
    mq->head = mq->head->next;
    if (mq->head == NULL) {
        mq->tail = NULL;
    }
    mq->count--;
    
    pthread_mutex_unlock(&mq->mutex);
    
    return msg;
}

int message_queue_size(MessageQueue *mq) {
    if (!mq) return 0;
    
    pthread_mutex_lock(&mq->mutex);
    int size = mq->count;
    pthread_mutex_unlock(&mq->mutex);
    
    return size;
}

void message_queue_wakeup(MessageQueue *mq) {
    if (!mq) return;
    
    pthread_mutex_lock(&mq->mutex);
    pthread_cond_broadcast(&mq->cond);
    pthread_mutex_unlock(&mq->mutex);
}
