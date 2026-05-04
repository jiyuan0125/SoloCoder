#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <fcntl.h>
#include <signal.h>
#include <errno.h>
#include <time.h>
#include <poll.h>

#include "cmd_exec.h"
#include "common.h"

static size_t initial_buf_size = 4096;

int cmd_config_init(cmd_config_t *config) {
    if (!config) return -1;
    memset(config, 0, sizeof(cmd_config_t));
    config->timeout_sec = DEFAULT_TIMEOUT_SEC;
    config->max_output_size = DEFAULT_MAX_OUTPUT_SIZE;
    return 0;
}

void cmd_config_free(cmd_config_t *config) {
    if (!config) return;
    if (config->env_vars) {
        for (size_t i = 0; i < config->env_count; i++) {
            free(config->env_vars[i].key);
            free(config->env_vars[i].value);
        }
        free(config->env_vars);
    }
    memset(config, 0, sizeof(cmd_config_t));
}

int cmd_config_set_env(cmd_config_t *config, const char *key, const char *value, bool overwrite) {
    if (!config || !key || !value) return -1;
    
    env_var_t *new_vars = realloc(config->env_vars, (config->env_count + 1) * sizeof(env_var_t));
    if (!new_vars) return -1;
    
    config->env_vars = new_vars;
    env_var_t *var = &config->env_vars[config->env_count];
    
    var->key = strdup(key);
    var->value = strdup(value);
    var->overwrite = overwrite;
    
    if (!var->key || !var->value) {
        free(var->key);
        free(var->value);
        return -1;
    }
    
    config->env_count++;
    return 0;
}

int cmd_result_init(cmd_result_t *result, cmd_config_t *config) {
    if (!result || !config) return -1;
    
    memset(result, 0, sizeof(cmd_result_t));
    result->config = config;
    result->status = CMD_STATUS_PENDING;
    result->pid = -1;
    result->stdout_pipe[0] = -1;
    result->stdout_pipe[1] = -1;
    result->stderr_pipe[0] = -1;
    result->stderr_pipe[1] = -1;
    
    if (output_buf_init(&result->output, config->max_output_size) != 0) {
        return -1;
    }
    
    return 0;
}

void cmd_result_free(cmd_result_t *result) {
    if (!result) return;
    output_buf_free(&result->output);
    if (result->stdout_pipe[0] >= 0) close(result->stdout_pipe[0]);
    if (result->stdout_pipe[1] >= 0) close(result->stdout_pipe[1]);
    if (result->stderr_pipe[0] >= 0) close(result->stderr_pipe[0]);
    if (result->stderr_pipe[1] >= 0) close(result->stderr_pipe[1]);
    memset(result, 0, sizeof(cmd_result_t));
}

int output_buf_init(output_buf_t *buf, size_t max_size) {
    if (!buf) return -1;
    memset(buf, 0, sizeof(output_buf_t));
    
    buf->stdout_alloc = initial_buf_size;
    buf->stdout_buf = malloc(buf->stdout_alloc);
    if (!buf->stdout_buf) return -1;
    buf->stdout_buf[0] = '\0';
    
    buf->stderr_alloc = initial_buf_size;
    buf->stderr_buf = malloc(buf->stderr_alloc);
    if (!buf->stderr_buf) {
        free(buf->stdout_buf);
        return -1;
    }
    buf->stderr_buf[0] = '\0';
    
    return 0;
}

void output_buf_free(output_buf_t *buf) {
    if (!buf) return;
    free(buf->stdout_buf);
    free(buf->stderr_buf);
    memset(buf, 0, sizeof(output_buf_t));
}

ssize_t output_buf_append(output_buf_t *buf, const char *data, size_t len, bool is_stdout) {
    if (!buf || !data || len == 0) return 0;
    
    char **buf_ptr;
    size_t *len_ptr;
    size_t *alloc_ptr;
    bool *trunc_ptr;
    size_t max_size = buf->stdout_alloc > buf->stderr_alloc ? 
                       (buf->stdout_alloc > initial_buf_size ? buf->stdout_alloc : DEFAULT_MAX_OUTPUT_SIZE) :
                       (buf->stderr_alloc > initial_buf_size ? buf->stderr_alloc : DEFAULT_MAX_OUTPUT_SIZE);
    
    if (is_stdout) {
        buf_ptr = &buf->stdout_buf;
        len_ptr = &buf->stdout_len;
        alloc_ptr = &buf->stdout_alloc;
        trunc_ptr = &buf->stdout_truncated;
    } else {
        buf_ptr = &buf->stderr_buf;
        len_ptr = &buf->stderr_len;
        alloc_ptr = &buf->stderr_alloc;
        trunc_ptr = &buf->stderr_truncated;
    }
    
    if (*trunc_ptr) {
        return len;
    }
    
    if (*len_ptr + len >= max_size) {
        size_t copy_len = max_size - *len_ptr - 1;
        if (copy_len > 0) {
            memcpy(*buf_ptr + *len_ptr, data, copy_len);
            *len_ptr += copy_len;
            (*buf_ptr)[*len_ptr] = '\0';
        }
        *trunc_ptr = true;
        return len;
    }
    
    if (*len_ptr + len + 1 > *alloc_ptr) {
        size_t new_alloc = *alloc_ptr * 2;
        while (new_alloc < *len_ptr + len + 1) {
            new_alloc *= 2;
        }
        if (new_alloc > max_size) {
            new_alloc = max_size;
        }
        char *new_buf = realloc(*buf_ptr, new_alloc);
        if (!new_buf) {
            *trunc_ptr = true;
            return -1;
        }
        *buf_ptr = new_buf;
        *alloc_ptr = new_alloc;
    }
    
    memcpy(*buf_ptr + *len_ptr, data, len);
    *len_ptr += len;
    (*buf_ptr)[*len_ptr] = '\0';
    
    return len;
}

char **build_argv(const char *command) {
    if (!command || command[0] == '\0') return NULL;
    
    size_t argc = 0;
    size_t buf_size = 16;
    char **argv = malloc(buf_size * sizeof(char *));
    if (!argv) return NULL;
    
    const char *ptr = command;
    while (*ptr) {
        while (*ptr == ' ' || *ptr == '\t') ptr++;
        if (*ptr == '\0') break;
        
        const char *start = ptr;
        bool in_quote = false;
        char quote_char = 0;
        
        while (*ptr) {
            if (!in_quote && (*ptr == ' ' || *ptr == '\t')) {
                break;
            }
            if (*ptr == '"' || *ptr == '\'') {
                if (!in_quote) {
                    in_quote = true;
                    quote_char = *ptr;
                } else if (*ptr == quote_char) {
                    in_quote = false;
                    quote_char = 0;
                }
            }
            ptr++;
        }
        
        size_t len = ptr - start;
        char *arg = malloc(len + 1);
        if (!arg) {
            free_argv(argv);
            return NULL;
        }
        
        char *dst = arg;
        const char *src = start;
        in_quote = false;
        quote_char = 0;
        
        while (src < ptr) {
            if (*src == '"' || *src == '\'') {
                if (!in_quote) {
                    in_quote = true;
                    quote_char = *src;
                    src++;
                    continue;
                } else if (*src == quote_char) {
                    in_quote = false;
                    quote_char = 0;
                    src++;
                    continue;
                }
            }
            *dst++ = *src++;
        }
        *dst = '\0';
        
        if (argc + 1 >= buf_size) {
            size_t new_size = buf_size * 2;
            char **new_argv = realloc(argv, new_size * sizeof(char *));
            if (!new_argv) {
                free(arg);
                free_argv(argv);
                return NULL;
            }
            argv = new_argv;
            buf_size = new_size;
        }
        
        argv[argc++] = arg;
    }
    
    argv[argc] = NULL;
    return argv;
}

void free_argv(char **argv) {
    if (!argv) return;
    for (size_t i = 0; argv[i]; i++) {
        free(argv[i]);
    }
    free(argv);
}

char **build_envp(env_var_t *env_vars, size_t env_count) {
    extern char **environ;
    
    size_t base_count = 0;
    if (environ) {
        while (environ[base_count]) base_count++;
    }
    
    char **new_envp = malloc((base_count + env_count + 1) * sizeof(char *));
    if (!new_envp) return NULL;
    
    size_t idx = 0;
    for (size_t i = 0; i < base_count; i++) {
        bool should_override = false;
        for (size_t j = 0; j < env_count; j++) {
            if (env_vars[j].overwrite) {
                size_t keylen = strlen(env_vars[j].key);
                if (strncmp(environ[i], env_vars[j].key, keylen) == 0 && environ[i][keylen] == '=') {
                    should_override = true;
                    break;
                }
            }
        }
        if (!should_override) {
            new_envp[idx++] = strdup(environ[i]);
        }
    }
    
    for (size_t j = 0; j < env_count; j++) {
        size_t var_len = strlen(env_vars[j].key) + strlen(env_vars[j].value) + 2;
        char *var = malloc(var_len);
        if (!var) {
            free_envp(new_envp);
            return NULL;
        }
        snprintf(var, var_len, "%s=%s", env_vars[j].key, env_vars[j].value);
        new_envp[idx++] = var;
    }
    
    new_envp[idx] = NULL;
    return new_envp;
}

void free_envp(char **envp) {
    if (!envp) return;
    for (size_t i = 0; envp[i]; i++) {
        free(envp[i]);
    }
    free(envp);
}

int cmd_kill_process_group(pid_t pgid) {
    if (pgid <= 0) return -1;
    return killpg(pgid, SIGKILL);
}

int cmd_read_pipes(cmd_result_t *result) {
    if (!result) return -1;
    
    struct pollfd pfds[2];
    int nfds = 0;
    
    if (result->stdout_pipe[0] >= 0) {
        pfds[nfds].fd = result->stdout_pipe[0];
        pfds[nfds].events = POLLIN;
        nfds++;
    }
    if (result->stderr_pipe[0] >= 0) {
        pfds[nfds].fd = result->stderr_pipe[0];
        pfds[nfds].events = POLLIN;
        nfds++;
    }
    
    if (nfds == 0) return 0;
    
    int ret = poll(pfds, nfds, 100);
    if (ret <= 0) return ret;
    
    char buf[READ_BUFFER_SIZE];
    ssize_t n;
    ssize_t total_read = 0;
    
    for (int i = 0; i < nfds; i++) {
        if (pfds[i].revents & (POLLIN | POLLHUP)) {
            if (pfds[i].fd == result->stdout_pipe[0]) {
                while (true) {
                    n = read(result->stdout_pipe[0], buf, sizeof(buf) - 1);
                    if (n > 0) {
                        output_buf_append(&result->output, buf, n, true);
                        total_read += n;
                    } else if (n == 0) {
                        close(result->stdout_pipe[0]);
                        result->stdout_pipe[0] = -1;
                        break;
                    } else {
                        if (errno == EAGAIN || errno == EWOULDBLOCK) {
                            break;
                        }
                        break;
                    }
                }
            } else if (pfds[i].fd == result->stderr_pipe[0]) {
                while (true) {
                    n = read(result->stderr_pipe[0], buf, sizeof(buf) - 1);
                    if (n > 0) {
                        output_buf_append(&result->output, buf, n, false);
                        total_read += n;
                    } else if (n == 0) {
                        close(result->stderr_pipe[0]);
                        result->stderr_pipe[0] = -1;
                        break;
                    } else {
                        if (errno == EAGAIN || errno == EWOULDBLOCK) {
                            break;
                        }
                        break;
                    }
                }
            }
        }
    }
    
    return total_read > 0 ? (int)total_read : 0;
}

int cmd_start(cmd_result_t *result) {
    if (!result || !result->config) return -1;
    
    if (pipe(result->stdout_pipe) != 0) return -1;
    if (pipe(result->stderr_pipe) != 0) {
        close(result->stdout_pipe[0]);
        close(result->stdout_pipe[1]);
        return -1;
    }
    
    int flags;
    flags = fcntl(result->stdout_pipe[0], F_GETFL);
    fcntl(result->stdout_pipe[0], F_SETFL, flags | O_NONBLOCK);
    flags = fcntl(result->stderr_pipe[0], F_GETFL);
    fcntl(result->stderr_pipe[0], F_SETFL, flags | O_NONBLOCK);
    
    pid_t pid = fork();
    
    if (pid < 0) {
        close(result->stdout_pipe[0]);
        close(result->stdout_pipe[1]);
        close(result->stderr_pipe[0]);
        close(result->stderr_pipe[1]);
        return -1;
    }
    
    if (pid == 0) {
        setpgid(0, 0);
        
        close(result->stdout_pipe[0]);
        close(result->stderr_pipe[0]);
        
        int stdout_fd = result->stdout_pipe[1];
        int stderr_fd = result->stderr_pipe[1];
        
        if (result->config->redirect_stdout) {
            int fd = open(result->config->redirect_stdout, 
                         O_WRONLY | O_CREAT | O_TRUNC, 0644);
            if (fd >= 0) {
                stdout_fd = fd;
            }
        }
        
        if (result->config->redirect_stderr) {
            int fd = open(result->config->redirect_stderr,
                         O_WRONLY | O_CREAT | O_TRUNC, 0644);
            if (fd >= 0) {
                stderr_fd = fd;
            }
        }
        
        if (dup2(stdout_fd, STDOUT_FILENO) < 0) {
            perror("dup2 stdout");
            _exit(CMD_EXIT_EXEC_FAILED);
        }
        
        if (dup2(stderr_fd, STDERR_FILENO) < 0) {
            perror("dup2 stderr");
            _exit(CMD_EXIT_EXEC_FAILED);
        }
        
        close(result->stdout_pipe[1]);
        close(result->stderr_pipe[1]);
        
        if (result->config->work_dir) {
            if (chdir(result->config->work_dir) != 0) {
                perror("chdir");
                _exit(CMD_EXIT_EXEC_FAILED);
            }
        }
        
        char **envp = build_envp(result->config->env_vars, result->config->env_count);
        
        char *argv[4];
        argv[0] = "/bin/sh";
        argv[1] = "-c";
        argv[2] = (char *)result->config->command;
        argv[3] = NULL;
        
        execve("/bin/sh", argv, envp);
        
        perror("execve");
        if (envp) free_envp(envp);
        _exit(CMD_EXIT_EXEC_FAILED);
    }
    
    setpgid(pid, pid);
    
    close(result->stdout_pipe[1]);
    close(result->stderr_pipe[1]);
    result->stdout_pipe[1] = -1;
    result->stderr_pipe[1] = -1;
    
    result->pid = pid;
    result->status = CMD_STATUS_RUNNING;
    result->start_time = time(NULL);
    
    return 0;
}

int cmd_wait(cmd_result_t *result, int timeout_sec) {
    if (!result || result->pid <= 0) return -1;
    
    time_t start = time(NULL);
    int timeout = timeout_sec > 0 ? timeout_sec : DEFAULT_TIMEOUT_SEC;
    
    while (true) {
        while (cmd_read_pipes(result) > 0);
        
        int status;
        pid_t w = waitpid(result->pid, &status, WNOHANG);
        
        if (w < 0) {
            if (errno == EINTR) continue;
            return -1;
        }
        
        if (w > 0) {
            while (cmd_read_pipes(result) > 0);
            
            result->end_time = time(NULL);
            result->elapsed_sec = difftime(result->end_time, result->start_time);
            
            if (result->timed_out) {
                result->status = CMD_STATUS_TIMEOUT;
                result->exit_code = CMD_EXIT_TIMEOUT;
            } else if (WIFEXITED(status)) {
                result->exit_code = WEXITSTATUS(status);
                result->status = (result->exit_code == 0) ? CMD_STATUS_SUCCESS : CMD_STATUS_FAILED;
            } else if (WIFSIGNALED(status)) {
                result->exit_code = 128 + WTERMSIG(status);
                result->status = CMD_STATUS_KILLED;
            }
            
            return 0;
        }
        
        time_t now = time(NULL);
        if (difftime(now, start) >= timeout) {
            result->timed_out = true;
            cmd_kill_process_group(result->pid);
            usleep(100000);
            continue;
        }
        
        usleep(100000);
    }
}

int cmd_execute_sync(cmd_config_t *config, cmd_result_t *result) {
    if (!config || !result) return -1;
    
    if (cmd_result_init(result, config) != 0) {
        return -1;
    }
    
    if (cmd_start(result) != 0) {
        cmd_result_free(result);
        return -1;
    }
    
    if (cmd_wait(result, config->timeout_sec) != 0) {
        cmd_result_free(result);
        return -1;
    }
    
    return 0;
}
