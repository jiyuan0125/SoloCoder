#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <pthread.h>
#include <time.h>
#include <signal.h>

#include "common.h"
#include "queue_manager.h"
#include "table_manager.h"
#include "strategy.h"

static int g_running = 1;
static WaitQueue g_queue;
static SeatingStrategy g_strategy;

typedef struct {
    int customer_id;
    int table_number;
    int eating_time;
} CustomerEatingInfo;

void handle_sigint(int sig) {
    (void)sig;
    printf("\n\n正在关闭系统...\n");
    g_running = 0;
}

void print_menu(void) {
    printf("\n");
    printf("========================================\n");
    printf("        餐厅等位排队系统 - 管理界面\n");
    printf("========================================\n");
    printf("1. 客人取号\n");
    printf("2. VIP客人取号\n");
    printf("3. 客人取消排队\n");
    printf("4. 查看当前排队状态\n");
    printf("5. 查看桌台状态\n");
    printf("6. 客人用餐结束(释放桌台)\n");
    printf("7. 查询预计等待时间\n");
    printf("8. 手动触发超时检查\n");
    printf("9. 查看系统统计信息\n");
    printf("10. 启动模拟测试\n");
    printf("0. 退出系统\n");
    printf("========================================\n");
    printf("请选择操作: ");
}

void print_queue_status(void) {
    int count = queue_get_count(&g_queue);
    int vip_count = queue_get_vip_count(&g_queue);
    int max_cap = queue_get_max_capacity();
    
    printf("\n=== 当前排队状态 ===\n");
    printf("排队人数: %d / %d\n", count, max_cap);
    printf("VIP排队人数: %d\n", vip_count);
    
    if (count > 0) {
        Customer list[MAX_QUEUE];
        int actual = queue_list_all_customers(&g_queue, list, MAX_QUEUE);
        
        printf("\n排队名单:\n");
        printf("--------------------------------------------------------\n");
        printf("序号\t排队号\t姓名\t\t类型\t\t排队时长\n");
        printf("--------------------------------------------------------\n");
        
        for (int i = 0; i < actual; i++) {
            int pos = i + 1;
            const char *type_str = (list[i].type == CUSTOMER_VIP) ? "VIP" : "普通";
            time_t now = time(NULL);
            int wait_secs = (int)difftime(now, list[i].queue_time);
            
            double est_wait = queue_estimate_wait_time(&g_queue, i);
            
            printf("%d\t%d\t%s\t\t%s\t\t%02d:%02d (预计等%.0f秒)\n", 
                   pos, list[i].id, list[i].name, type_str,
                   wait_secs / 60, wait_secs % 60, est_wait);
        }
        printf("--------------------------------------------------------\n");
    }
}

void print_table_status(void) {
    Table tables[MAX_TABLES];
    int count = table_list_all(&g_table_manager, tables, MAX_TABLES);
    
    printf("\n=== 桌台状态 ===\n");
    printf("总桌台数: %d\n", MAX_TABLES);
    printf("空闲桌台: %d\n", table_get_free_count(&g_table_manager));
    printf("占用桌台: %d\n", table_get_occupied_count(&g_table_manager));
    
    printf("\n桌台详情:\n");
    printf("-----------------------------------------------------------------------\n");
    printf("桌台号\t状态\t\t客人ID\t\t占用时长\t用餐时长\n");
    printf("-----------------------------------------------------------------------\n");
    
    for (int i = 0; i < count; i++) {
        const char *status_str;
        switch (tables[i].status) {
            case TABLE_FREE: status_str = "空闲"; break;
            case TABLE_OCCUPIED: status_str = "已入座"; break;
            case TABLE_EATING: status_str = "用餐中"; break;
            case TABLE_DIRTY: status_str = "清洁中"; break;
            default: status_str = "未知"; break;
        }
        
        if (tables[i].status == TABLE_FREE) {
            printf("%d\t%s\t\t-\t\t-\t\t-\n", 
                   tables[i].number, status_str);
        } else {
            time_t now = time(NULL);
            int occupied_secs = (int)difftime(now, tables[i].occupied_time);
            int eating_secs = 0;
            if (tables[i].eating_start_time > 0) {
                eating_secs = (int)difftime(now, tables[i].eating_start_time);
            }
            
            printf("%d\t%s\t\t%d\t\t%02d:%02d\t\t%02d:%02d\n", 
                   tables[i].number, status_str, tables[i].customer_id,
                   occupied_secs / 60, occupied_secs % 60,
                   eating_secs / 60, eating_secs % 60);
        }
    }
    printf("-----------------------------------------------------------------------\n");
}

void print_system_stats(void) {
    int queue_count, free_tables, occupied_tables;
    double avg_eating_time;
    
    strategy_get_system_status(&g_strategy, &queue_count, &free_tables, 
                                &occupied_tables, &avg_eating_time);
    
    printf("\n=== 系统统计信息 ===\n");
    printf("排队人数: %d\n", queue_count);
    printf("空闲桌台: %d\n", free_tables);
    printf("占用桌台: %d\n", occupied_tables);
    printf("历史平均用餐时间: %.1f 秒 (%.1f 分钟)\n", 
           avg_eating_time, avg_eating_time / 60.0);
    printf("VIP功能: %s\n", strategy_is_vip_enabled(&g_strategy) ? "已启用" : "已禁用");
    printf("超时阈值: %d 秒\n", TIMEOUT_THRESHOLD);
}

void *timeout_check_thread(void *arg) {
    (void)arg;
    
    while (g_running) {
        sleep(10);
        if (!g_running) break;
        
        int released = strategy_force_timeout_check(&g_strategy, TIMEOUT_THRESHOLD);
        if (released > 0) {
            printf("\n[系统] 超时检查: 释放了 %d 个超时桌台\n", released);
        }
    }
    
    return NULL;
}

void *customer_eating_thread(void *arg) {
    CustomerEatingInfo *info = (CustomerEatingInfo *)arg;
    
    printf("[模拟] 客人 %d 在桌台 %d 开始用餐，预计用时 %d 秒\n", 
           info->customer_id, info->table_number, info->eating_time);
    
    table_start_eating(&g_table_manager, info->table_number, info->eating_time);
    
    sleep(info->eating_time);
    
    if (g_running) {
        printf("[模拟] 客人 %d 用餐结束，离开桌台 %d\n", 
               info->customer_id, info->table_number);
        table_release(&g_table_manager, info->table_number);
        
        Customer seated;
        int table_num;
        int ret = strategy_try_seat_next_customer(&g_strategy, &seated, &table_num);
        if (ret == STRATEGY_SUCCESS) {
            printf("[模拟] 桌台 %d 空出，安排客人 %d (%s) 入座\n",
                   table_num, seated.id, seated.name);
            
            CustomerEatingInfo *new_info = malloc(sizeof(CustomerEatingInfo));
            new_info->customer_id = seated.id;
            new_info->table_number = table_num;
            new_info->eating_time = 30 + rand() % 60;
            
            pthread_t tid;
            pthread_create(&tid, NULL, customer_eating_thread, new_info);
            pthread_detach(tid);
        }
    }
    
    free(info);
    return NULL;
}

void *simulation_thread(void *arg) {
    (void)arg;
    srand(time(NULL));
    
    int customer_count = 0;
    const char *names[] = {"张三", "李四", "王五", "赵六", "钱七", "孙八", "周九", "吴十",
                            "Alice", "Bob", "Charlie", "David", "Emma", "Frank", "Grace", "Henry"};
    int name_count = sizeof(names) / sizeof(names[0]);
    
    while (g_running && customer_count < 50) {
        int interval = 2 + rand() % 5;
        sleep(interval);
        
        if (!g_running) break;
        
        char name[50];
        int is_vip = (rand() % 100) < 20;
        int name_idx = rand() % name_count;
        
        snprintf(name, sizeof(name), "%s-%d", names[name_idx], customer_count + 1);
        
        int assigned_id;
        int ret;
        
        if (is_vip) {
            ret = queue_add_vip_customer(&g_queue, name, &assigned_id);
            if (ret == QUEUE_SUCCESS) {
                printf("\n[模拟] VIP客人 %s 取号成功，排队号: %d\n", name, assigned_id);
            }
        } else {
            ret = queue_add_customer(&g_queue, name, CUSTOMER_NORMAL, &assigned_id);
            if (ret == QUEUE_SUCCESS) {
                printf("\n[模拟] 客人 %s 取号成功，排队号: %d\n", name, assigned_id);
            }
        }
        
        if (ret == QUEUE_FULL) {
            printf("\n[模拟] 客人 %s 被告知：当前排队已满，请稍后再来\n", name);
        }
        
        if (ret == QUEUE_SUCCESS) {
            Customer seated;
            int table_num;
            int seat_ret = strategy_try_seat_next_customer(&g_strategy, &seated, &table_num);
            
            if (seat_ret == STRATEGY_SUCCESS) {
                printf("[模拟] 立即安排客人 %d (%s) 到桌台 %d 入座\n",
                       seated.id, seated.name, table_num);
                
                CustomerEatingInfo *info = malloc(sizeof(CustomerEatingInfo));
                info->customer_id = seated.id;
                info->table_number = table_num;
                info->eating_time = 30 + rand() % 60;
                
                pthread_t tid;
                pthread_create(&tid, NULL, customer_eating_thread, info);
                pthread_detach(tid);
            }
        }
        
        if (rand() % 100 < 10) {
            Customer list[MAX_QUEUE];
            int actual = queue_list_all_customers(&g_queue, list, MAX_QUEUE);
            if (actual > 0) {
                int idx = rand() % actual;
                printf("\n[模拟] 客人 %d (%s) 不耐烦了，取消排队离开\n",
                       list[idx].id, list[idx].name);
                queue_cancel_customer(&g_queue, list[idx].id);
            }
        }
        
        customer_count++;
    }
    
    printf("\n[模拟] 模拟线程结束，共产生 %d 个客人\n", customer_count);
    return NULL;
}

void start_simulation(void) {
    pthread_t sim_tid;
    pthread_create(&sim_tid, NULL, simulation_thread, NULL);
    pthread_detach(sim_tid);
    printf("\n模拟测试已启动，正在后台运行...\n");
}

int main(int argc, char *argv[]) {
    (void)argc;
    (void)argv;
    
    signal(SIGINT, handle_sigint);
    
    printf("正在初始化餐厅等位排队系统...\n");
    
    statistics_init();
    
    if (queue_init(&g_queue) != QUEUE_SUCCESS) {
        printf("错误: 初始化等待队列失败\n");
        return 1;
    }
    
    if (table_manager_init(&g_table_manager) != TABLE_SUCCESS) {
        printf("错误: 初始化桌台管理失败\n");
        queue_destroy(&g_queue);
        return 1;
    }
    
    if (strategy_init(&g_strategy, &g_queue, &g_table_manager) != STRATEGY_SUCCESS) {
        printf("错误: 初始化等位策略失败\n");
        table_manager_destroy(&g_table_manager);
        queue_destroy(&g_queue);
        return 1;
    }
    
    pthread_t timeout_tid;
    pthread_create(&timeout_tid, NULL, timeout_check_thread, NULL);
    pthread_detach(timeout_tid);
    
    printf("系统初始化完成！\n");
    printf("桌台数: %d, 最大排队人数: %d, 超时阈值: %d秒\n",
           MAX_TABLES, MAX_QUEUE, TIMEOUT_THRESHOLD);
    printf("VIP插队功能: 已支持\n");
    printf("自动跳过已离开客人: 已启用\n");
    
    while (g_running) {
        print_menu();
        
        int choice;
        if (scanf("%d", &choice) != 1) {
            while (getchar() != '\n');
            printf("\n无效输入，请重新选择\n");
            continue;
        }
        while (getchar() != '\n');
        
        switch (choice) {
            case 1: {
                char name[50];
                printf("请输入客人姓名: ");
                if (fgets(name, sizeof(name), stdin)) {
                    name[strcspn(name, "\n")] = '\0';
                }
                
                if (strlen(name) == 0) {
                    printf("姓名不能为空\n");
                    break;
                }
                
                int assigned_id;
                int ret = queue_add_customer(&g_queue, name, CUSTOMER_NORMAL, &assigned_id);
                
                if (ret == QUEUE_SUCCESS) {
                    printf("取号成功！排队号: %d\n", assigned_id);
                    
                    int pos = queue_get_customer_position(&g_queue, assigned_id);
                    if (pos >= 0) {
                        double est_wait = queue_estimate_wait_time(&g_queue, pos);
                        printf("当前排队位置: 第 %d 位\n", pos + 1);
                        printf("预计等待时间: %.0f 秒 (约 %.1f 分钟)\n", 
                               est_wait, est_wait / 60.0);
                    }
                } else if (ret == QUEUE_FULL) {
                    printf("当前排队已满，请稍后再来\n");
                } else {
                    printf("取号失败，错误码: %d\n", ret);
                }
                break;
            }
            
            case 2: {
                if (!strategy_is_vip_enabled(&g_strategy)) {
                    printf("VIP功能当前已禁用\n");
                    break;
                }
                
                char name[50];
                printf("请输入VIP客人姓名: ");
                if (fgets(name, sizeof(name), stdin)) {
                    name[strcspn(name, "\n")] = '\0';
                }
                
                if (strlen(name) == 0) {
                    printf("姓名不能为空\n");
                    break;
                }
                
                int assigned_id;
                int ret = queue_add_vip_customer(&g_queue, name, &assigned_id);
                
                if (ret == QUEUE_SUCCESS) {
                    printf("VIP取号成功！排队号: %d (VIP优先排队)\n", assigned_id);
                    
                    int pos = queue_get_customer_position(&g_queue, assigned_id);
                    if (pos >= 0) {
                        double est_wait = queue_estimate_wait_time(&g_queue, pos);
                        printf("当前排队位置: 第 %d 位\n", pos + 1);
                        printf("预计等待时间: %.0f 秒 (约 %.1f 分钟)\n", 
                               est_wait, est_wait / 60.0);
                    }
                } else if (ret == QUEUE_FULL) {
                    printf("当前排队已满，请稍后再来\n");
                } else {
                    printf("取号失败，错误码: %d\n", ret);
                }
                break;
            }
            
            case 3: {
                printf("请输入要取消的排队号: ");
                int customer_id;
                if (scanf("%d", &customer_id) != 1) {
                    while (getchar() != '\n');
                    printf("无效的排队号\n");
                    break;
                }
                while (getchar() != '\n');
                
                int ret = queue_cancel_customer(&g_queue, customer_id);
                
                if (ret == QUEUE_SUCCESS) {
                    printf("客人 %d 已成功取消排队\n", customer_id);
                } else if (ret == CUSTOMER_NOT_FOUND) {
                    printf("未找到该客人或客人已不在排队中\n");
                } else {
                    printf("取消失败，错误码: %d\n", ret);
                }
                break;
            }
            
            case 4:
                print_queue_status();
                break;
            
            case 5:
                print_table_status();
                break;
            
            case 6: {
                printf("请输入要释放的桌台号: ");
                int table_num;
                if (scanf("%d", &table_num) != 1) {
                    while (getchar() != '\n');
                    printf("无效的桌台号\n");
                    break;
                }
                while (getchar() != '\n');
                
                TableStatus status;
                if (table_get_status(&g_table_manager, table_num, &status) != TABLE_SUCCESS) {
                    printf("无效的桌台号\n");
                    break;
                }
                
                if (status == TABLE_FREE) {
                    printf("桌台 %d 已经是空闲状态\n", table_num);
                    break;
                }
                
                int ret = table_release(&g_table_manager, table_num);
                
                if (ret == TABLE_SUCCESS) {
                    printf("桌台 %d 已成功释放\n", table_num);
                    
                    Customer seated;
                    int new_table_num;
                    int seat_ret = strategy_try_seat_next_customer(&g_strategy, &seated, &new_table_num);
                    
                    if (seat_ret == STRATEGY_SUCCESS) {
                        printf("已安排下一位客人 %d (%s) 入座桌台 %d\n",
                               seated.id, seated.name, new_table_num);
                    } else if (seat_ret == STRATEGY_NO_CUSTOMER) {
                        printf("当前没有等待的客人\n");
                    }
                } else {
                    printf("释放桌台失败，错误码: %d\n", ret);
                }
                break;
            }
            
            case 7: {
                printf("请输入排队号查询预计等待时间: ");
                int customer_id;
                if (scanf("%d", &customer_id) != 1) {
                    while (getchar() != '\n');
                    printf("无效的排队号\n");
                    break;
                }
                while (getchar() != '\n');
                
                double est_wait = strategy_estimate_wait_for_customer(&g_strategy, customer_id);
                
                if (est_wait < 0) {
                    printf("未找到该客人或客人已不在排队中\n");
                } else {
                    int pos = queue_get_customer_position(&g_queue, customer_id);
                    printf("客人 %d 当前排队位置: 第 %d 位\n", customer_id, pos + 1);
                    printf("预计等待时间: %.0f 秒 (约 %.1f 分钟)\n", 
                           est_wait, est_wait / 60.0);
                }
                break;
            }
            
            case 8: {
                printf("请输入超时时间(秒): ");
                int timeout;
                if (scanf("%d", &timeout) != 1 || timeout <= 0) {
                    while (getchar() != '\n');
                    printf("使用默认超时时间: %d 秒\n", TIMEOUT_THRESHOLD);
                    timeout = TIMEOUT_THRESHOLD;
                }
                while (getchar() != '\n');
                
                int released = strategy_force_timeout_check(&g_strategy, timeout);
                printf("超时检查完成，释放了 %d 个桌台\n", released);
                break;
            }
            
            case 9:
                print_system_stats();
                break;
            
            case 10:
                start_simulation();
                break;
            
            case 0:
                printf("正在退出系统...\n");
                g_running = 0;
                break;
            
            default:
                printf("无效的选项，请重新选择\n");
                break;
        }
    }
    
    printf("正在清理资源...\n");
    strategy_destroy(&g_strategy);
    table_manager_destroy(&g_table_manager);
    queue_destroy(&g_queue);
    
    printf("系统已关闭，再见！\n");
    return 0;
}
