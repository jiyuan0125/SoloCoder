#define _GNU_SOURCE
#include "scheduler.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>
#include <unistd.h>

static volatile sig_atomic_t g_running = 1;
static scheduler_t g_scheduler;

typedef struct {
    char name[TASK_NAME_MAX_LEN];
    int duration_seconds;
} task_context_t;

static void demo_task_func(void* user_data) {
    task_context_t* ctx = (task_context_t*)user_data;
    if (!ctx) return;

    time_t now = time(NULL);
    struct tm* tm_now = localtime(&now);
    char time_str[64];
    strftime(time_str, sizeof(time_str), "%Y-%m-%d %H:%M:%S", tm_now);

    printf("[%s] Task '%s' started (duration: %d sec)\n", time_str, ctx->name, ctx->duration_seconds);
    fflush(stdout);

    if (ctx->duration_seconds > 0) {
        sleep(ctx->duration_seconds);
    }

    now = time(NULL);
    tm_now = localtime(&now);
    strftime(time_str, sizeof(time_str), "%Y-%m-%d %H:%M:%S", tm_now);
    printf("[%s] Task '%s' completed\n", time_str, ctx->name);
    fflush(stdout);
}

static void signal_handler(int sig) {
    (void)sig;
    g_running = 0;
    printf("\nReceived signal, stopping scheduler...\n");
    fflush(stdout);
}

static void print_help(void) {
    printf("\nCron Scheduler Demo - Commands:\n");
    printf("  list      - List all scheduled tasks\n");
    printf("  add       - Add a new task (interactive)\n");
    printf("  remove    - Remove a task by ID\n");
    printf("  pause     - Pause a task by ID\n");
    printf("  resume    - Resume a paused task by ID\n");
    printf("  help      - Show this help message\n");
    printf("  quit      - Exit the scheduler\n");
    printf("\n");
}

static void cmd_list(void) {
    char buffer[8192];
    int count = scheduler_list_tasks(&g_scheduler, buffer, sizeof(buffer));
    if (count < 0) {
        printf("Error listing tasks\n");
    } else {
        printf("\n=== Scheduled Tasks (%d total) ===\n", count);
        printf("%s", buffer);
    }
    fflush(stdout);
}

static void cmd_add(void) {
    char name[TASK_NAME_MAX_LEN] = {0};
    char cron_expr[TASK_EXPR_MAX_LEN] = {0};
    char duration_str[32] = {0};
    int duration = 0;

    printf("Enter task name: ");
    fflush(stdout);
    if (fgets(name, sizeof(name), stdin) == NULL) return;
    name[strcspn(name, "\r\n")] = '\0';

    printf("Enter cron expression (5 fields): ");
    fflush(stdout);
    if (fgets(cron_expr, sizeof(cron_expr), stdin) == NULL) return;
    cron_expr[strcspn(cron_expr, "\r\n")] = '\0';

    printf("Enter task duration in seconds (0 for instant): ");
    fflush(stdout);
    if (fgets(duration_str, sizeof(duration_str), stdin) == NULL) return;
    duration = atoi(duration_str);

    task_context_t* ctx = (task_context_t*)malloc(sizeof(task_context_t));
    if (!ctx) {
        printf("Memory allocation failed\n");
        return;
    }
    strncpy(ctx->name, name, TASK_NAME_MAX_LEN - 1);
    ctx->name[TASK_NAME_MAX_LEN - 1] = '\0';
    ctx->duration_seconds = duration;

    task_id_t id = scheduler_add_task(&g_scheduler, name, cron_expr, demo_task_func, ctx);
    if (id == 0) {
        printf("Failed to add task. Invalid cron expression?\n");
        free(ctx);
    } else {
        printf("Task added with ID: %u\n", id);
    }
}

static void cmd_remove(void) {
    char id_str[32] = {0};
    printf("Enter task ID to remove: ");
    fflush(stdout);
    if (fgets(id_str, sizeof(id_str), stdin) == NULL) return;
    
    task_id_t id = (task_id_t)atoi(id_str);
    if (scheduler_remove_task(&g_scheduler, id) == 0) {
        printf("Task %u removed\n", id);
    } else {
        printf("Task %u not found\n", id);
    }
}

static void cmd_pause(void) {
    char id_str[32] = {0};
    printf("Enter task ID to pause: ");
    fflush(stdout);
    if (fgets(id_str, sizeof(id_str), stdin) == NULL) return;
    
    task_id_t id = (task_id_t)atoi(id_str);
    if (scheduler_pause_task(&g_scheduler, id) == 0) {
        printf("Task %u paused\n", id);
    } else {
        printf("Task %u not found or already paused\n", id);
    }
}

static void cmd_resume(void) {
    char id_str[32] = {0};
    printf("Enter task ID to resume: ");
    fflush(stdout);
    if (fgets(id_str, sizeof(id_str), stdin) == NULL) return;
    
    task_id_t id = (task_id_t)atoi(id_str);
    if (scheduler_resume_task(&g_scheduler, id) == 0) {
        printf("Task %u resumed\n", id);
    } else {
        printf("Task %u not found or not paused\n", id);
    }
}

int main(void) {
    printf("========================================\n");
    printf("  Cron Task Scheduler (C Implementation)\n");
    printf("========================================\n\n");

    if (scheduler_init(&g_scheduler) != 0) {
        fprintf(stderr, "Failed to initialize scheduler\n");
        return 1;
    }

    task_context_t* ctx1 = (task_context_t*)malloc(sizeof(task_context_t));
    strcpy(ctx1->name, "Every-Minute Demo");
    ctx1->duration_seconds = 2;
    scheduler_add_task(&g_scheduler, "Every-Minute Demo", "* * * * *", demo_task_func, ctx1);

    task_context_t* ctx2 = (task_context_t*)malloc(sizeof(task_context_t));
    strcpy(ctx2->name, "Every-5-Minutes Demo");
    ctx2->duration_seconds = 5;
    scheduler_add_task(&g_scheduler, "Every-5-Minutes Demo", "*/5 * * * *", demo_task_func, ctx2);

    task_context_t* ctx_db = (task_context_t*)malloc(sizeof(task_context_t));
    strcpy(ctx_db->name, "Daily DB Backup");
    ctx_db->duration_seconds = 10;
    scheduler_add_task(&g_scheduler, "Daily DB Backup", "0 3 * * *", demo_task_func, ctx_db);

    task_context_t* ctx_metrics = (task_context_t*)malloc(sizeof(task_context_t));
    strcpy(ctx_metrics->name, "Metrics Collection");
    ctx_metrics->duration_seconds = 1;
    scheduler_add_task(&g_scheduler, "Metrics Collection", "*/15 * * * *", demo_task_func, ctx_metrics);

    task_context_t* ctx_clean = (task_context_t*)malloc(sizeof(task_context_t));
    strcpy(ctx_clean->name, "Temp File Cleanup");
    ctx_clean->duration_seconds = 3;
    scheduler_add_task(&g_scheduler, "Temp File Cleanup", "0 * * * *", demo_task_func, ctx_clean);

    task_context_t* ctx_report = (task_context_t*)malloc(sizeof(task_context_t));
    strcpy(ctx_report->name, "Weekly Report");
    ctx_report->duration_seconds = 5;
    scheduler_add_task(&g_scheduler, "Weekly Report", "0 9 * * 1", demo_task_func, ctx_report);

    printf("Demo Tasks Added:\n");
    printf("  1. '* * * * *'     - Every minute (2 sec duration)\n");
    printf("  2. '*/5 * * * *'   - Every 5 minutes (5 sec duration)\n");
    printf("  3. '0 3 * * *'     - Daily at 3:00 AM (DB backup example)\n");
    printf("  4. '*/15 * * * *'  - Every 15 minutes (metrics example)\n");
    printf("  5. '0 * * * *'     - Every hour on the hour (cleanup example)\n");
    printf("  6. '0 9 * * 1'     - Every Monday at 9:00 AM (report example)\n\n");

    print_help();

    struct sigaction sa;
    sa.sa_handler = signal_handler;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = 0;
    sigaction(SIGINT, &sa, NULL);
    sigaction(SIGTERM, &sa, NULL);

    if (scheduler_start(&g_scheduler) != 0) {
        fprintf(stderr, "Failed to start scheduler\n");
        scheduler_destroy(&g_scheduler);
        return 1;
    }

    printf("Scheduler started. Type 'list' to see tasks, 'help' for commands.\n\n");
    fflush(stdout);

    char cmd[256];
    while (g_running) {
        printf("scheduler> ");
        fflush(stdout);

        fd_set readfds;
        struct timeval tv;
        FD_ZERO(&readfds);
        FD_SET(STDIN_FILENO, &readfds);
        tv.tv_sec = 1;
        tv.tv_usec = 0;

        int sel = select(STDIN_FILENO + 1, &readfds, NULL, NULL, &tv);
        if (sel < 0) {
            if (g_running) break;
            continue;
        }
        if (sel == 0) continue;

        if (fgets(cmd, sizeof(cmd), stdin) == NULL) continue;
        
        cmd[strcspn(cmd, "\r\n")] = '\0';

        if (strcmp(cmd, "list") == 0) {
            cmd_list();
        } else if (strcmp(cmd, "add") == 0) {
            cmd_add();
        } else if (strcmp(cmd, "remove") == 0) {
            cmd_remove();
        } else if (strcmp(cmd, "pause") == 0) {
            cmd_pause();
        } else if (strcmp(cmd, "resume") == 0) {
            cmd_resume();
        } else if (strcmp(cmd, "help") == 0) {
            print_help();
        } else if (strcmp(cmd, "quit") == 0 || strcmp(cmd, "exit") == 0) {
            g_running = 0;
        } else if (strlen(cmd) > 0) {
            printf("Unknown command: '%s'. Type 'help' for available commands.\n", cmd);
        }
    }

    printf("Stopping scheduler...\n");
    scheduler_stop(&g_scheduler);
    printf("Waiting for running tasks to complete...\n");
    scheduler_join_all(&g_scheduler);
    scheduler_destroy(&g_scheduler);
    printf("Scheduler stopped. Goodbye!\n");

    return 0;
}
