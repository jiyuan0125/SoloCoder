#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <sys/sysinfo.h>
#include <pthread.h>
#include <errno.h>
#include "thread_pool.h"
#include "mmap_search.h"

typedef struct {
    const char *search_pattern;
    int ignore_case;
    int show_line_numbers;
    int show_filename_only;
    int show_count_only;
    int thread_count;
    char **filenames;
    int file_count;
} options_t;

typedef struct {
    pthread_mutex_t mutex;
    int completed_blocks;
    int total_blocks;
    const char *current_filename;
} progress_t;

static progress_t g_progress = {
    .mutex = PTHREAD_MUTEX_INITIALIZER,
    .completed_blocks = 0,
    .total_blocks = 0,
    .current_filename = NULL
};

static void print_usage(const char *progname) {
    fprintf(stderr, "Usage: %s [OPTIONS] PATTERN FILE [FILE...]\n", progname);
    fprintf(stderr, "Options:\n");
    fprintf(stderr, "  -i        Ignore case\n");
    fprintf(stderr, "  -l        Only print filenames with matches\n");
    fprintf(stderr, "  -c        Only print count of matching lines\n");
    fprintf(stderr, "  -n        Print line numbers (default)\n");
    fprintf(stderr, "  -t NUM    Number of threads (default: CPU cores)\n");
}

static int parse_options(int argc, char *argv[], options_t *opts) {
    int opt;

    opts->ignore_case = 0;
    opts->show_line_numbers = 1;
    opts->show_filename_only = 0;
    opts->show_count_only = 0;
    opts->thread_count = get_nprocs();
    opts->filenames = NULL;
    opts->file_count = 0;

    while ((opt = getopt(argc, argv, "ilcnt:h")) != -1) {
        switch (opt) {
            case 'i':
                opts->ignore_case = 1;
                break;
            case 'l':
                opts->show_filename_only = 1;
                break;
            case 'c':
                opts->show_count_only = 1;
                break;
            case 'n':
                opts->show_line_numbers = 1;
                break;
            case 't':
                opts->thread_count = atoi(optarg);
                if (opts->thread_count <= 0) {
                    opts->thread_count = 1;
                }
                break;
            case 'h':
                print_usage(argv[0]);
                exit(0);
            default:
                print_usage(argv[0]);
                return -1;
        }
    }

    if (optind >= argc) {
        fprintf(stderr, "Error: Pattern required\n");
        print_usage(argv[0]);
        return -1;
    }

    opts->search_pattern = argv[optind++];

    if (optind >= argc) {
        fprintf(stderr, "Error: At least one file required\n");
        print_usage(argv[0]);
        return -1;
    }

    opts->filenames = &argv[optind];
    opts->file_count = argc - optind;

    return 0;
}

static void update_progress(const char *filename) {
    pthread_mutex_lock(&g_progress.mutex);
    
    g_progress.completed_blocks++;
    
    if (filename && g_progress.current_filename != filename) {
        g_progress.current_filename = filename;
        g_progress.completed_blocks = 0;
    }
    
    if (g_progress.total_blocks > 0 && g_progress.current_filename) {
        int percent = (g_progress.completed_blocks * 100) / g_progress.total_blocks;
        if (percent > 100) percent = 100;
        fprintf(stderr, "\rProcessing %s: %d%%", g_progress.current_filename, percent);
        fflush(stderr);
    }
    
    pthread_mutex_unlock(&g_progress.mutex);
}

static void worker_task(void *arg) {
    search_task_t *task = (search_task_t *)arg;
    
    search_in_block(task);
    update_progress(task->result->filename);
    
    free(task);
}

static int process_small_file(const char *filename, options_t *opts, file_result_t *result) {
    FILE *f = fopen(filename, "rb");
    if (!f) {
        fprintf(stderr, "\nError: Cannot open %s: %s\n", filename, strerror(errno));
        return -1;
    }

    fseek(f, 0, SEEK_END);
    long file_size = ftell(f);
    fseek(f, 0, SEEK_SET);

    if (file_size < 0) {
        fclose(f);
        return -1;
    }

    char *buffer = (char *)malloc(file_size + 1);
    if (!buffer) {
        fclose(f);
        return -1;
    }

    size_t read_size = fread(buffer, 1, file_size, f);
    fclose(f);

    if (read_size != (size_t)file_size) {
        free(buffer);
        return -1;
    }
    buffer[file_size] = '\0';

    if (is_binary_file(buffer, file_size)) {
        fprintf(stderr, "%s: binary file skipped\n", filename);
        free(buffer);
        return 0;
    }

    g_progress.current_filename = filename;
    g_progress.total_blocks = 1;
    g_progress.completed_blocks = 0;
    fprintf(stderr, "\rProcessing %s: 0%%", filename);
    fflush(stderr);

    search_task_t *task = (search_task_t *)calloc(1, sizeof(search_task_t));
    if (!task) {
        free(buffer);
        return -1;
    }

    task->search_pattern = opts->search_pattern;
    task->ignore_case = opts->ignore_case;
    task->show_line_numbers = opts->show_line_numbers;
    task->show_filename_only = opts->show_filename_only;
    task->show_count_only = opts->show_count_only;
    task->result = result;
    task->data = buffer;
    task->data_size = file_size;
    task->block_offset = 0;
    task->block_actual_size = file_size;
    task->block_index = 0;
    task->total_blocks = 1;
    task->is_first_block = 1;
    task->is_last_block = 1;
    task->file_size = file_size;

    search_in_block(task);
    update_progress(filename);
    fprintf(stderr, "\rProcessing %s: 100%%", filename);
    fflush(stderr);

    free(task);
    free(buffer);
    return 0;
}

static int process_large_file(const char *filename, options_t *opts, 
                               file_result_t *result, thread_pool_t *pool) {
    int fd = open(filename, O_RDONLY);
    if (fd < 0) {
        fprintf(stderr, "\nError: Cannot open %s: %s\n", filename, strerror(errno));
        return -1;
    }

    struct stat st;
    if (fstat(fd, &st) < 0) {
        close(fd);
        return -1;
    }

    size_t file_size = st.st_size;

    char *mapped = mmap(NULL, file_size, PROT_READ, MAP_PRIVATE, fd, 0);
    if (mapped == MAP_FAILED) {
        close(fd);
        fprintf(stderr, "\nError: Cannot mmap %s: %s\n", filename, strerror(errno));
        return -1;
    }

    close(fd);

    if (is_binary_file(mapped, file_size)) {
        fprintf(stderr, "%s: binary file skipped\n", filename);
        munmap(mapped, file_size);
        return 0;
    }

    int total_blocks = (file_size + BLOCK_SIZE - 1) / BLOCK_SIZE;

    g_progress.current_filename = filename;
    g_progress.total_blocks = total_blocks;
    g_progress.completed_blocks = 0;
    fprintf(stderr, "\rProcessing %s: 0%%", filename);
    fflush(stderr);

    size_t *block_boundaries = (size_t *)calloc(total_blocks + 1, sizeof(size_t));
    if (!block_boundaries) {
        munmap(mapped, file_size);
        return -1;
    }

    block_boundaries[0] = 0;
    for (int i = 1; i < total_blocks; i++) {
        size_t pos = i * BLOCK_SIZE;
        size_t nl_pos = find_next_newline(mapped, pos, file_size);
        if (nl_pos < file_size) {
            block_boundaries[i] = nl_pos + 1;
        } else {
            block_boundaries[i] = nl_pos;
        }
    }
    block_boundaries[total_blocks] = file_size;

    for (int i = 0; i < total_blocks; i++) {
        size_t start = block_boundaries[i];
        size_t end = block_boundaries[i + 1];
        
        if (start >= end && i < total_blocks - 1) {
            continue;
        }

        search_task_t *task = (search_task_t *)calloc(1, sizeof(search_task_t));
        if (!task) {
            continue;
        }

        task->search_pattern = opts->search_pattern;
        task->ignore_case = opts->ignore_case;
        task->show_line_numbers = opts->show_line_numbers;
        task->show_filename_only = opts->show_filename_only;
        task->show_count_only = opts->show_count_only;
        task->result = result;
        task->data = mapped;
        task->data_size = file_size;
        task->block_offset = start;
        task->block_actual_size = end - start;
        task->block_index = i;
        task->total_blocks = total_blocks;
        task->is_first_block = (i == 0);
        task->is_last_block = (i == total_blocks - 1);
        task->file_size = file_size;

        thread_pool_add(pool, worker_task, task);
    }

    free(block_boundaries);

    while (1) {
        pthread_mutex_lock(&g_progress.mutex);
        int done = (g_progress.completed_blocks >= total_blocks);
        pthread_mutex_unlock(&g_progress.mutex);
        
        if (done) break;
        usleep(10000);
    }

    munmap(mapped, file_size);
    return 0;
}

int main(int argc, char *argv[]) {
    options_t opts;
    if (parse_options(argc, argv, &opts) < 0) {
        return 1;
    }

    thread_pool_t *pool = thread_pool_create(opts.thread_count);
    if (!pool) {
        fprintf(stderr, "Error: Cannot create thread pool\n");
        return 1;
    }

    int total_files = opts.file_count;
    int matched_files = 0;
    int total_matches = 0;

    for (int i = 0; i < opts.file_count; i++) {
        const char *filename = opts.filenames[i];
        
        file_result_t result;
        result.filename = (char *)filename;
        result.matched = 0;
        result.match_count = 0;
        pthread_mutex_init(&result.mutex, NULL);

        struct stat st;
        if (stat(filename, &st) < 0) {
            fprintf(stderr, "Error: Cannot stat %s: %s\n", filename, strerror(errno));
            pthread_mutex_destroy(&result.mutex);
            continue;
        }

        if (S_ISDIR(st.st_mode)) {
            fprintf(stderr, "Error: %s is a directory\n", filename);
            pthread_mutex_destroy(&result.mutex);
            continue;
        }

        off_t file_size = st.st_size;

        if (file_size <= SMALL_FILE_THRESHOLD) {
            process_small_file(filename, &opts, &result);
        } else {
            process_large_file(filename, &opts, &result, pool);
        }

        fprintf(stderr, "\n");
        fflush(stderr);

        if (result.matched) {
            matched_files++;
            total_matches += result.match_count;
            
            if (opts.show_filename_only) {
                printf("%s\n", filename);
            }
            if (opts.show_count_only) {
                printf("%s:%d\n", filename, result.match_count);
            }
        }

        pthread_mutex_destroy(&result.mutex);
    }

    thread_pool_destroy(pool);

    fprintf(stderr, "\n%d files, %d matched, %d matches\n", 
            total_files, matched_files, total_matches);

    return 0;
}
