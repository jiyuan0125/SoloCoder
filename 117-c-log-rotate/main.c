#include "log_rotate.h"
#include "backup_manager.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <dirent.h>
#include <sys/stat.h>

#define LOG_FILE "app.log"
#define TEST_SIZE ((size_t)512)
#define MAX_BACKUPS 5

static void list_log_files(const char *base_path)
{
    char dir_path[1024];
    const char *file_name;
    char *last_slash = strrchr(base_path, '/');
    
    if (last_slash != NULL) {
        size_t dir_len = last_slash - base_path;
        if (dir_len >= sizeof(dir_path) - 1) {
            printf("  Path too long\n");
            return;
        }
        strncpy(dir_path, base_path, dir_len);
        dir_path[dir_len] = '\0';
        file_name = last_slash + 1;
    } else {
        strncpy(dir_path, ".", sizeof(dir_path) - 1);
        dir_path[sizeof(dir_path) - 1] = '\0';
        file_name = base_path;
    }
    
    DIR *dir = opendir(dir_path);
    if (dir == NULL) {
        printf("  Cannot open directory: %s\n", dir_path);
        return;
    }
    
    printf("  Current log files:\n");
    
    size_t base_name_len = strlen(file_name);
    struct dirent *entry;
    
    while ((entry = readdir(dir)) != NULL) {
        const char *name = entry->d_name;
        
        if (strcmp(name, ".") == 0 || strcmp(name, "..") == 0) {
            continue;
        }
        
        if (strncmp(name, file_name, base_name_len) == 0) {
            char full_path[2048];
            snprintf(full_path, sizeof(full_path), "%s/%s", dir_path, name);
            
            struct stat st;
            if (stat(full_path, &st) == 0) {
                printf("    %-25s %10ld bytes\n", name, st.st_size);
            } else {
                printf("    %s\n", name);
            }
        }
    }
    
    closedir(dir);
}

static void clean_test_files(const char *base_path)
{
    DIR *dir = opendir(".");
    if (dir == NULL) {
        return;
    }
    
    const char *file_name = base_path;
    char *last_slash = strrchr(base_path, '/');
    if (last_slash != NULL) {
        file_name = last_slash + 1;
    }
    
    size_t base_name_len = strlen(file_name);
    struct dirent *entry;
    
    while ((entry = readdir(dir)) != NULL) {
        const char *name = entry->d_name;
        
        if (strncmp(name, file_name, base_name_len) == 0) {
            unlink(name);
        }
    }
    
    closedir(dir);
}

int main(void)
{
    printf("========================================\n");
    printf("  Log Rotate Module Demo\n");
    printf("========================================\n\n");
    
    printf("Configuration:\n");
    printf("  Log file: %s\n", LOG_FILE);
    printf("  Max size: %zu bytes (%zu KB)\n", TEST_SIZE, TEST_SIZE / 1024);
    printf("  Max backups: %d\n", MAX_BACKUPS);
    printf("  Compression: enabled\n");
    printf("  Rotate mode: by size\n\n");
    
    clean_test_files(LOG_FILE);
    
    log_rotate_config_t config;
    memset(&config, 0, sizeof(config));
    strncpy(config.base_path, LOG_FILE, sizeof(config.base_path) - 1);
    config.max_size = TEST_SIZE;
    config.max_backups = MAX_BACKUPS;
    config.mode = LOG_ROTATE_MODE_SIZE;
    config.compress_enabled = 1;
    
    log_rotate_t lr;
    int ret = log_rotate_init(&lr, &config);
    if (ret != 0) {
        printf("Failed to initialize log rotate!\n");
        return 1;
    }
    
    printf("Starting log writing...\n\n");
    
    char message[512];
    int total_writes = 0;
    int expected_rotations = 3;
    
    for (int i = 1; i <= 50; i++) {
        snprintf(message, sizeof(message), 
                 "[%04d] This is a log message to test the log rotate module. "
                 "Adding some padding to make the message longer. "
                 "The quick brown fox jumps over the lazy dog.\n",
                 i);
        
        int written = log_rotate_write(&lr, "%s", message);
        if (written < 0) {
            printf("Failed to write log!\n");
            break;
        }
        
        total_writes++;
        
        if (i % 5 == 0) {
            printf("  Wrote %d messages, checking files...\n", i);
            list_log_files(LOG_FILE);
            printf("\n");
            usleep(100000);
        }
    }
    
    printf("Waiting for async compression to complete...\n");
    sleep(2);
    
    printf("\nFinal state:\n");
    list_log_files(LOG_FILE);
    
    log_rotate_destroy(&lr);
    log_rotate_cleanup_compressor();
    
    printf("\n========================================\n");
    printf("  Demo completed successfully!\n");
    printf("========================================\n");
    
    printf("\nSummary:\n");
    printf("  Total writes: %d\n", total_writes);
    printf("  Check the log files listed above.\n");
    printf("  Backup files should be named: %s.1, %s.2, etc.\n", LOG_FILE, LOG_FILE);
    printf("  Oldest backups should be deleted when exceeding %d.\n", MAX_BACKUPS);
    printf("  Compressed files have .gz extension.\n");
    
    return 0;
}
