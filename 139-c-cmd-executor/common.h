#ifndef COMMON_H
#define COMMON_H

#include <sys/types.h>
#include <stdbool.h>
#include <time.h>

#define DEFAULT_TIMEOUT_SEC       30
#define DEFAULT_MAX_CONCURRENT    10
#define DEFAULT_MAX_OUTPUT_SIZE   (10 * 1024 * 1024)
#define DEFAULT_OUTPUT_SUMMARY    5
#define READ_BUFFER_SIZE          4096

#define CMD_EXIT_SUCCESS          0
#define CMD_EXIT_GENERAL_ERROR    1
#define CMD_EXIT_TIMEOUT          124
#define CMD_EXIT_KILLED           137
#define CMD_EXIT_EXEC_FAILED      126

typedef enum {
    CMD_STATUS_PENDING = 0,
    CMD_STATUS_RUNNING,
    CMD_STATUS_SUCCESS,
    CMD_STATUS_FAILED,
    CMD_STATUS_TIMEOUT,
    CMD_STATUS_KILLED
} cmd_status_t;

typedef struct {
    char *key;
    char *value;
    bool overwrite;
} env_var_t;

typedef struct {
    const char *command;
    const char *work_dir;
    const char *redirect_stdout;
    const char *redirect_stderr;
    env_var_t *env_vars;
    size_t env_count;
    int timeout_sec;
    size_t max_output_size;
    int id;
    const char *label;
} cmd_config_t;

typedef struct {
    char *stdout_buf;
    size_t stdout_len;
    size_t stdout_alloc;
    bool stdout_truncated;
    char *stderr_buf;
    size_t stderr_len;
    size_t stderr_alloc;
    bool stderr_truncated;
    size_t max_size;
} output_buf_t;

typedef struct {
    cmd_config_t *config;
    cmd_status_t status;
    int exit_code;
    output_buf_t output;
    time_t start_time;
    time_t end_time;
    double elapsed_sec;
    pid_t pid;
    int stdout_pipe[2];
    int stderr_pipe[2];
    bool timed_out;
} cmd_result_t;

typedef struct {
    int total;
    int success;
    int failed;
    int timeout;
    int killed;
} summary_stats_t;

#endif
