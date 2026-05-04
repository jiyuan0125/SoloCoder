#define _DEFAULT_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <time.h>
#include <string.h>
#include "progress.h"
#include "terminal.h"

#define ONE_KB (1024ULL)
#define ONE_MB (1024ULL * 1024ULL)
#define ONE_GB (1024ULL * 1024ULL * 1024ULL)

static void sleep_ms(int milliseconds) {
    struct timespec ts;
    ts.tv_sec = milliseconds / 1000;
    ts.tv_nsec = (milliseconds % 1000) * 1000000L;
    nanosleep(&ts, NULL);
}

static void demo_single_progress(void) {
    printf("\n=== 演示：单进度条 (1GB 文件复制) ===\n\n");
    
    ProgressManager manager;
    progress_manager_init(&manager, 4);
    
    size_t file_idx = progress_manager_add(&manager, ONE_GB, DEFAULT_BAR_WIDTH);
    
    terminal_init();
    
    uint64_t chunk_size = ONE_MB;
    uint64_t total_written = 0;
    
    while (total_written < ONE_GB) {
        sleep_ms(10);
        
        uint64_t write_size = chunk_size;
        if (total_written + write_size > ONE_GB) {
            write_size = ONE_GB - total_written;
        }
        
        progress_manager_update(&manager, file_idx, write_size);
        total_written += write_size;
        
        progress_manager_detect_terminal(&manager);
        terminal_draw_all(&manager);
    }
    
    if (manager.terminal_supported) {
        fprintf(stdout, "\n");
        fflush(stdout);
    }
    
    terminal_cleanup();
    progress_manager_destroy(&manager);
    
    printf("\n✓ 文件复制完成！\n");
}

static void demo_multi_progress(void) {
    printf("\n=== 演示：多进度条 (3个文件同时下载) ===\n\n");
    
    ProgressManager manager;
    progress_manager_init(&manager, 4);
    
    uint64_t file_sizes[] = {
        500 * ONE_MB,
        ONE_GB,
        750 * ONE_MB
    };
    int num_files = 3;
    
    for (int i = 0; i < num_files; i++) {
        progress_manager_add(&manager, file_sizes[i], DEFAULT_BAR_WIDTH);
    }
    
    terminal_init();
    
    uint64_t chunk_size = 2 * ONE_MB;
    int done = 0;
    
    while (!done) {
        sleep_ms(15);
        
        for (int i = 0; i < num_files; i++) {
            ProgressState *state = &manager.states[i];
            if (state->is_complete) continue;
            
            uint64_t write_size = chunk_size;
            if (i == 1) {
                write_size = chunk_size / 2;
            } else if (i == 2) {
                write_size = chunk_size * 3 / 2;
            }
            
            if (state->current_bytes + write_size > state->total_bytes) {
                write_size = state->total_bytes - state->current_bytes;
            }
            
            progress_manager_update(&manager, i, write_size);
        }
        
        progress_manager_detect_terminal(&manager);
        terminal_draw_all(&manager);
        
        done = progress_manager_all_complete(&manager);
    }
    
    if (manager.terminal_supported) {
        fprintf(stdout, "\n");
        fflush(stdout);
    }
    
    terminal_cleanup();
    progress_manager_destroy(&manager);
    
    printf("\n✓ 所有文件下载完成！\n");
}

int main(int argc, char *argv[]) {
    srand((unsigned int)time(NULL));
    
    int demo_mode = 0;
    
    if (argc > 1) {
        if (strcmp(argv[1], "single") == 0) {
            demo_mode = 1;
        } else if (strcmp(argv[1], "multi") == 0) {
            demo_mode = 2;
        }
    }
    
    if (demo_mode == 0 || demo_mode == 1) {
        demo_single_progress();
    }
    
    if (demo_mode == 0 || demo_mode == 2) {
        demo_multi_progress();
    }
    
    printf("\n使用方法：\n");
    printf("  ./progress_bar        # 运行所有演示\n");
    printf("  ./progress_bar single # 仅单进度条演示\n");
    printf("  ./progress_bar multi  # 仅多进度条演示\n");
    
    return 0;
}
