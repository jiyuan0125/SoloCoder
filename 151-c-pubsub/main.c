#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <time.h>

#include "common.h"
#include "message.h"
#include "subscription.h"
#include "channel.h"

#define DEMO_CHANNEL1 "技术讨论"
#define DEMO_CHANNEL2 "闲聊"
#define DEMO_CHANNEL3 "公告"

typedef struct {
    UserManager *um;
    ChannelManager *cm;
    char username[USER_NAME_MAX];
    int exit_flag;
} ReceiverThreadArgs;

void *receiver_thread(void *arg) {
    ReceiverThreadArgs *args = (ReceiverThreadArgs *)arg;
    User *user = user_find(args->um, args->username);
    
    if (!user) {
        printf("[%s] 错误：找不到用户\n", args->username);
        return NULL;
    }
    
    printf("[%s] 开始接收消息...\n", args->username);
    
    while (!args->exit_flag) {
        Message *msg = user_receive_message_timed(user, 500);
        if (msg) {
            char time_buf[32];
            struct tm *tm_info = localtime(&msg->timestamp);
            strftime(time_buf, sizeof(time_buf), "%H:%M:%S", tm_info);
            
            printf("[%s] [%s] [%s] %s: %s\n", 
                   args->username, time_buf, msg->channel, msg->sender, msg->content);
            message_destroy(msg);
        }
    }
    
    printf("[%s] 接收线程退出\n", args->username);
    return NULL;
}

void print_error(const char *op, int err) {
    switch (err) {
        case CHAT_OK:
            printf("%s: 成功\n", op);
            break;
        case CHAT_ERR_CHANNEL_EXISTS:
            printf("%s: 错误 - 频道已存在\n", op);
            break;
        case CHAT_ERR_CHANNEL_NOT_FOUND:
            printf("%s: 错误 - 频道不存在\n", op);
            break;
        case CHAT_ERR_USER_NOT_FOUND:
            printf("%s: 错误 - 用户不存在\n", op);
            break;
        case CHAT_ERR_USER_ALREADY_SUBSCRIBED:
            printf("%s: 错误 - 用户已订阅\n", op);
            break;
        case CHAT_ERR_USER_NOT_SUBSCRIBED:
            printf("%s: 错误 - 用户未订阅\n", op);
            break;
        case CHAT_ERR_TOO_MANY_CHANNELS:
            printf("%s: 错误 - 频道数量超限\n", op);
            break;
        case CHAT_ERR_TOO_MANY_SUBSCRIPTIONS:
            printf("%s: 错误 - 订阅数量超限\n", op);
            break;
        case CHAT_ERR_INVALID_ARG:
            printf("%s: 错误 - 无效参数\n", op);
            break;
        case CHAT_ERR_MEMORY:
            printf("%s: 错误 - 内存分配失败\n", op);
            break;
        case CHAT_ERR_CHANNEL_DESTROYED:
            printf("%s: 错误 - 频道已销毁\n", op);
            break;
        default:
            printf("%s: 错误 - 未知错误 (%d)\n", op, err);
    }
}

void demo_basic_operations(UserManager *um, ChannelManager *cm) {
    printf("\n========== 演示1: 基础操作 ==========\n");
    
    printf("\n--- 创建频道 ---\n");
    Channel *ch1 = channel_create(cm, DEMO_CHANNEL1);
    printf("创建频道「%s」: %s\n", DEMO_CHANNEL1, ch1 ? "成功" : "失败");
    
    Channel *ch2 = channel_create(cm, DEMO_CHANNEL2);
    printf("创建频道「%s」: %s\n", DEMO_CHANNEL2, ch2 ? "成功" : "失败");
    
    Channel *ch1_dup = channel_create(cm, DEMO_CHANNEL1);
    printf("重复创建「%s」: %s (预期失败)\n", DEMO_CHANNEL1, ch1_dup ? "成功" : "失败");
    
    printf("\n--- 创建用户 ---\n");
    User *alice = user_create(um, "Alice");
    printf("创建用户 Alice: %s\n", alice ? "成功" : "失败");
    
    User *bob = user_create(um, "Bob");
    printf("创建用户 Bob: %s\n", bob ? "成功" : "失败");
    
    printf("\n--- 验证频道隔离 ---\n");
    printf("Alice 加入「%s」\n", DEMO_CHANNEL1);
    channel_add_subscriber(ch1, alice);
    
    printf("Bob 加入「%s」\n", DEMO_CHANNEL2);
    channel_add_subscriber(ch2, bob);
    
    printf("发送消息到「%s」: \"技术问题请教\"\n", DEMO_CHANNEL1);
    channel_publish_message(ch1, "Charlie", "技术问题请教");
    
    printf("发送消息到「%s」: \"今天天气不错\"\n", DEMO_CHANNEL2);
    channel_publish_message(ch2, "Dave", "今天天气不错");
    
    printf("Alice 等待消息...\n");
    Message *msg = user_receive_message_timed(alice, 1000);
    if (msg) {
        printf("Alice 收到: [%s] %s: %s\n", msg->channel, msg->sender, msg->content);
        message_destroy(msg);
    } else {
        printf("Alice 没有收到消息\n");
    }
    
    printf("Bob 等待消息...\n");
    msg = user_receive_message_timed(bob, 1000);
    if (msg) {
        printf("Bob 收到: [%s] %s: %s\n", msg->channel, msg->sender, msg->content);
        message_destroy(msg);
    } else {
        printf("Bob 没有收到消息\n");
    }
    
    printf("\n--- 验证新人看不到历史消息 ---\n");
    printf("先往「%s」发一条消息\n", DEMO_CHANNEL1);
    channel_publish_message(ch1, "OldUser", "这是历史消息");
    
    User *carol = user_create(um, "Carol");
    printf("Carol 刚创建，还没加入任何频道\n");
    
    msg = user_receive_message_timed(carol, 500);
    printf("Carol 检查消息: %s (预期没有)\n", msg ? "有消息" : "无消息");
    if (msg) message_destroy(msg);
    
    printf("Carol 加入「%s」\n", DEMO_CHANNEL1);
    channel_add_subscriber(ch1, carol);
    
    printf("再发一条新消息\n");
    channel_publish_message(ch1, "NewUser", "这是新消息");
    
    msg = user_receive_message_timed(carol, 1000);
    if (msg) {
        printf("Carol 收到: %s (应该是「这是新消息」)\n", msg->content);
        message_destroy(msg);
    } else {
        printf("Carol 没收到消息\n");
    }
    
    printf("\n--- 清理演示1 ---\n");
    channel_remove_subscriber(ch1, alice);
    channel_remove_subscriber(ch1, carol);
    channel_remove_subscriber(ch2, bob);
    
    while ((msg = user_receive_message_timed(alice, 100))) { message_destroy(msg); }
    while ((msg = user_receive_message_timed(bob, 100))) { message_destroy(msg); }
    while ((msg = user_receive_message_timed(carol, 100))) { message_destroy(msg); }
    
    user_destroy(um, "Alice");
    user_destroy(um, "Bob");
    user_destroy(um, "Carol");
    
    channel_destroy(cm, DEMO_CHANNEL1);
    channel_destroy(cm, DEMO_CHANNEL2);
}

void demo_empty_channel(UserManager *um, ChannelManager *cm) {
    printf("\n========== 演示2: 空频道消息丢弃 ==========\n");
    
    Channel *ch = channel_create(cm, "空测试频道");
    if (!ch) {
        printf("创建频道失败\n");
        return;
    }
    
    printf("频道「空测试频道」当前订阅者数: %d\n", channel_subscriber_count(ch));
    
    printf("发送消息到空频道: \"这条消息应该被丢弃\"\n");
    int ret = channel_publish_message(ch, "Sender", "这条消息应该被丢弃");
    print_error("发送到空频道", ret);
    
    printf("现在加入一个用户\n");
    User *tester = user_create(um, "Tester");
    channel_add_subscriber(ch, tester);
    
    printf("发送消息: \"这条消息应该能收到\"\n");
    channel_publish_message(ch, "Sender", "这条消息应该能收到");
    
    Message *msg = user_receive_message_timed(tester, 1000);
    if (msg) {
        printf("Tester 收到: %s\n", msg->content);
        message_destroy(msg);
    } else {
        printf("Tester 没收到消息 (错误)\n");
    }
    
    printf("\n--- 清理演示2 ---\n");
    channel_remove_subscriber(ch, tester);
    while ((msg = user_receive_message_timed(tester, 100))) { message_destroy(msg); }
    user_destroy(um, "Tester");
    channel_destroy(cm, "空测试频道");
}

void demo_unsubscribe(UserManager *um, ChannelManager *cm) {
    printf("\n========== 演示3: 退出频道后不再收到消息 ==========\n");
    
    Channel *ch = channel_create(cm, "测试频道");
    User *user = user_create(um, "Subscriber");
    
    channel_add_subscriber(ch, user);
    
    printf("发送消息1: \"加入时能收到\"\n");
    channel_publish_message(ch, "Admin", "加入时能收到");
    
    Message *msg = user_receive_message_timed(user, 1000);
    if (msg) {
        printf("收到: %s\n", msg->content);
        message_destroy(msg);
    }
    
    printf("\n用户退出频道\n");
    channel_remove_subscriber(ch, user);
    
    printf("发送消息2: \"退出后不应该收到\"\n");
    channel_publish_message(ch, "Admin", "退出后不应该收到");
    
    msg = user_receive_message_timed(user, 1000);
    if (msg) {
        printf("错误: 收到了退出后的消息: %s\n", msg->content);
        message_destroy(msg);
    } else {
        printf("正确: 没有收到退出后的消息\n");
    }
    
    printf("\n--- 清理演示3 ---\n");
    while ((msg = user_receive_message_timed(user, 100))) { message_destroy(msg); }
    user_destroy(um, "Subscriber");
    channel_destroy(cm, "测试频道");
}

void demo_concurrent_behavior(UserManager *um, ChannelManager *cm) {
    printf("\n========== 演示4: 并发场景行为定义 ==========\n");
    printf("行为定义: 只有在消息发送操作开始前已成功加入的用户才能收到消息\n");
    printf("          发送过程中加入的用户收不到本条消息\n\n");
    
    Channel *ch = channel_create(cm, "并发测试");
    
    User *early = user_create(um, "EarlyBird");
    channel_add_subscriber(ch, early);
    
    User *late = user_create(um, "LateComer");
    
    printf("EarlyBird 已加入频道\n");
    printf("发送消息: \"Hello\" (此时 LateComer 还没加入)\n");
    
    channel_publish_message(ch, "Sender", "Hello");
    
    printf("消息发送后，LateComer 才加入频道\n");
    channel_add_subscriber(ch, late);
    
    Message *msg;
    printf("\nEarlyBird 检查消息:\n");
    msg = user_receive_message_timed(early, 1000);
    if (msg) {
        printf("  收到: %s (正确)\n", msg->content);
        message_destroy(msg);
    } else {
        printf("  没收到 (错误)\n");
    }
    
    printf("LateComer 检查消息 (应该收不到，因为加入前发的):\n");
    msg = user_receive_message_timed(late, 500);
    if (msg) {
        printf("  收到: %s (错误 - 不该收到历史消息)\n", msg->content);
        message_destroy(msg);
    } else {
        printf("  没收到 (正确 - 加入前的消息看不到)\n");
    }
    
    printf("\n再发一条新消息:\n");
    channel_publish_message(ch, "Sender", "New Message");
    
    printf("EarlyBird 检查:\n");
    msg = user_receive_message_timed(early, 1000);
    if (msg) { printf("  收到: %s\n", msg->content); message_destroy(msg); }
    
    printf("LateComer 检查 (应该能收到，因为已加入):\n");
    msg = user_receive_message_timed(late, 1000);
    if (msg) { 
        printf("  收到: %s (正确)\n", msg->content); 
        message_destroy(msg); 
    } else {
        printf("  没收到 (错误)\n");
    }
    
    printf("\n--- 清理演示4 ---\n");
    channel_remove_subscriber(ch, early);
    channel_remove_subscriber(ch, late);
    while ((msg = user_receive_message_timed(early, 100))) { message_destroy(msg); }
    while ((msg = user_receive_message_timed(late, 100))) { message_destroy(msg); }
    user_destroy(um, "EarlyBird");
    user_destroy(um, "LateComer");
    channel_destroy(cm, "并发测试");
}

void demo_channel_destroy_notify(UserManager *um, ChannelManager *cm) {
    printf("\n========== 演示5: 频道销毁时通知订阅者 ==========\n");
    
    Channel *ch = channel_create(cm, "即将销毁的频道");
    
    User *u1 = user_create(um, "User1");
    User *u2 = user_create(um, "User2");
    
    channel_add_subscriber(ch, u1);
    channel_add_subscriber(ch, u2);
    
    printf("User1 和 User2 都加入了频道\n");
    printf("现在销毁频道...\n");
    
    channel_destroy(cm, "即将销毁的频道");
    
    printf("检查用户是否收到销毁通知:\n");
    
    Message *msg;
    msg = user_receive_message_timed(u1, 1000);
    if (msg) {
        printf("User1 收到: [%s] %s: %s\n", msg->channel, msg->sender, msg->content);
        message_destroy(msg);
    }
    
    msg = user_receive_message_timed(u2, 1000);
    if (msg) {
        printf("User2 收到: [%s] %s: %s\n", msg->channel, msg->sender, msg->content);
        message_destroy(msg);
    }
    
    printf("\n--- 清理演示5 ---\n");
    while ((msg = user_receive_message_timed(u1, 100))) { message_destroy(msg); }
    while ((msg = user_receive_message_timed(u2, 100))) { message_destroy(msg); }
    user_destroy(um, "User1");
    user_destroy(um, "User2");
}

void demo_multithreaded(UserManager *um, ChannelManager *cm) {
    printf("\n========== 演示6: 多线程并发接收（5个频道）==========\n");
    
    const char *channels[] = {"技术讨论", "闲聊", "公告", "游戏", "音乐"};
    const int num_channels = 5;
    
    for (int i = 0; i < num_channels; i++) {
        channel_create(cm, channels[i]);
    }
    
    User *super_user = user_create(um, "SuperUser");
    
    printf("SuperUser 同时加入 %d 个频道\n", num_channels);
    for (int i = 0; i < num_channels; i++) {
        Channel *ch = channel_find(cm, channels[i]);
        channel_add_subscriber(ch, super_user);
    }
    
    pthread_t recv_thread;
    ReceiverThreadArgs args;
    args.um = um;
    args.cm = cm;
    args.exit_flag = 0;
    strncpy(args.username, "SuperUser", USER_NAME_MAX - 1);
    
    pthread_create(&recv_thread, NULL, receiver_thread, &args);
    
    usleep(100000);
    
    printf("\n同时向 %d 个频道发送消息...\n", num_channels);
    for (int i = 0; i < num_channels; i++) {
        Channel *ch = channel_find(cm, channels[i]);
        char content[128];
        snprintf(content, sizeof(content), "来自「%s」的消息 #%d", channels[i], i + 1);
        channel_publish_message(ch, "Broadcaster", content);
        usleep(10000);
    }
    
    usleep(200000);
    
    printf("\n发送更多消息测试不丢失...\n");
    for (int round = 0; round < 3; round++) {
        for (int i = 0; i < num_channels; i++) {
            Channel *ch = channel_find(cm, channels[i]);
            char content[128];
            snprintf(content, sizeof(content), "第%d轮 - %s消息", round + 1, channels[i]);
            channel_publish_message(ch, "Broadcaster", content);
        }
    }
    
    usleep(500000);
    
    args.exit_flag = 1;
    message_queue_wakeup(&super_user->msg_queue);
    pthread_join(recv_thread, NULL);
    
    int remaining = 0;
    Message *msg;
    while ((msg = user_receive_message_timed(super_user, 100))) {
        remaining++;
        message_destroy(msg);
    }
    printf("\nSuperUser 消息队列剩余: %d 条\n", remaining);
    
    printf("\n--- 清理演示6 ---\n");
    for (int i = 0; i < num_channels; i++) {
        Channel *ch = channel_find(cm, channels[i]);
        if (ch) {
            channel_remove_subscriber(ch, super_user);
        }
    }
    while ((msg = user_receive_message_timed(super_user, 100))) { message_destroy(msg); }
    user_destroy(um, "SuperUser");
    for (int i = 0; i < num_channels; i++) {
        channel_destroy(cm, channels[i]);
    }
}

int main() {
    printf("========================================\n");
    printf("聊天室消息分发模块演示\n");
    printf("========================================\n");
    
    UserManager um;
    ChannelManager cm;
    
    printf("\n初始化管理器...\n");
    if (user_manager_init(&um) != CHAT_OK) {
        printf("用户管理器初始化失败\n");
        return 1;
    }
    
    if (channel_manager_init(&cm, &um) != CHAT_OK) {
        printf("频道管理器初始化失败\n");
        user_manager_destroy(&um);
        return 1;
    }
    printf("初始化完成\n");
    
    demo_basic_operations(&um, &cm);
    demo_empty_channel(&um, &cm);
    demo_unsubscribe(&um, &cm);
    demo_concurrent_behavior(&um, &cm);
    demo_channel_destroy_notify(&um, &cm);
    demo_multithreaded(&um, &cm);
    
    printf("\n========================================\n");
    printf("所有演示完成\n");
    printf("========================================\n");
    
    printf("\n清理资源...\n");
    channel_manager_destroy(&cm);
    user_manager_destroy(&um);
    printf("清理完成\n");
    
    return 0;
}
