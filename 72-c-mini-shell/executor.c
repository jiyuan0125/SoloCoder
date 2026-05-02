#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <fcntl.h>
#include <errno.h>
#include <signal.h>
#include "shell.h"

int last_exit_status = 0;

static int apply_redirect(const Command *cmd) {
    if (cmd->redirect_file != NULL) {
        int flags = O_WRONLY | O_CREAT;
        if (cmd->redirect_append) {
            flags |= O_APPEND;
        } else {
            flags |= O_TRUNC;
        }
        
        int fd = open(cmd->redirect_file, flags, 0644);
        if (fd < 0) {
            perror("open");
            return -1;
        }
        
        if (dup2(fd, STDOUT_FILENO) < 0) {
            perror("dup2");
            close(fd);
            return -1;
        }
        close(fd);
    }
    return 0;
}

static void execute_external(const Command *cmd) {
    if (cmd->argv[0] == NULL) {
        exit(EXIT_SUCCESS);
    }
    
    if (apply_redirect(cmd) < 0) {
        exit(EXIT_FAILURE);
    }
    
    execvp(cmd->argv[0], cmd->argv);
    
    if (errno == ENOENT) {
        fprintf(stderr, "%s: command not found\n", cmd->argv[0]);
        exit(127);
    } else {
        fprintf(stderr, "%s: %s\n", cmd->argv[0], strerror(errno));
        exit(126);
    }
}

int execute_builtin(const Command *cmd) {
    if (cmd->argv[0] == NULL) return -1;
    
    if (strcmp(cmd->argv[0], "cd") == 0) {
        char *dir = cmd->argv[1];
        if (dir == NULL) {
            dir = getenv("HOME");
            if (dir == NULL) {
                fprintf(stderr, "cd: HOME not set\n");
                return 1;
            }
        }
        
        if (chdir(dir) < 0) {
            fprintf(stderr, "cd: %s: %s\n", dir, strerror(errno));
            return 1;
        }
        return 0;
    }
    
    if (strcmp(cmd->argv[0], "exit") == 0) {
        int code = 0;
        if (cmd->argv[1] != NULL) {
            code = atoi(cmd->argv[1]);
        }
        exit(code);
    }
    
    if (strcmp(cmd->argv[0], "export") == 0) {
        if (cmd->argv[1] == NULL) {
            return 0;
        }
        
        char *eq = strchr(cmd->argv[1], '=');
        if (eq != NULL) {
            *eq = '\0';
            char *var = cmd->argv[1];
            char *value = eq + 1;
            if (setenv(var, value, 1) < 0) {
                perror("setenv");
                return 1;
            }
        }
        return 0;
    }
    
    return -1;
}

int execute_pipeline(Pipeline *pipeline) {
    if (pipeline->num_commands == 0) return 0;
    
    int is_background = pipeline->commands[pipeline->num_commands - 1].background;
    
    if (pipeline->num_commands == 1) {
        Command *cmd = &pipeline->commands[0];
        
        BuiltinType builtin = check_builtin(cmd);
        if (builtin != BUILTIN_NONE) {
            return execute_builtin(cmd);
        }
    }
    
    int pipes[MAX_PIPES - 1][2];
    for (int i = 0; i < pipeline->num_commands - 1; i++) {
        if (pipe(pipes[i]) < 0) {
            perror("pipe");
            return 1;
        }
    }
    
    pid_t pgid = 0;
    pid_t child_pids[MAX_PIPES];
    
    for (int i = 0; i < pipeline->num_commands; i++) {
        Command *cmd = &pipeline->commands[i];
        
        pid_t pid = fork();
        if (pid < 0) {
            perror("fork");
            for (int j = 0; j < pipeline->num_commands - 1; j++) {
                close(pipes[j][0]);
                close(pipes[j][1]);
            }
            return 1;
        }
        
        if (pid == 0) {
            signal(SIGINT, SIG_DFL);
            signal(SIGQUIT, SIG_DFL);
            signal(SIGTSTP, SIG_DFL);
            signal(SIGTTIN, SIG_DFL);
            signal(SIGTTOU, SIG_DFL);
            
            if (!is_background) {
                if (i == 0) {
                    setpgrp();
                } else {
                    setpgid(0, pgid);
                }
            } else {
                setpgrp();
            }
            
            if (i > 0) {
                if (dup2(pipes[i - 1][0], STDIN_FILENO) < 0) {
                    perror("dup2");
                    exit(EXIT_FAILURE);
                }
            }
            
            if (i < pipeline->num_commands - 1) {
                if (dup2(pipes[i][1], STDOUT_FILENO) < 0) {
                    perror("dup2");
                    exit(EXIT_FAILURE);
                }
            }
            
            for (int j = 0; j < pipeline->num_commands - 1; j++) {
                close(pipes[j][0]);
                close(pipes[j][1]);
            }
            
            if (i == pipeline->num_commands - 1) {
                BuiltinType builtin = check_builtin(cmd);
                if (builtin != BUILTIN_NONE) {
                    exit(execute_builtin(cmd));
                }
            }
            
            execute_external(cmd);
        }
        
        child_pids[i] = pid;
        
        if (i == 0) {
            pgid = pid;
        }
        
        if (!is_background) {
            setpgid(pid, pgid);
        }
    }
    
    for (int i = 0; i < pipeline->num_commands - 1; i++) {
        close(pipes[i][0]);
        close(pipes[i][1]);
    }
    
    if (is_background) {
        printf("[1] %d\n", child_pids[pipeline->num_commands - 1]);
        fflush(stdout);
        return 0;
    } else {
        int status;
        int exit_code = 0;
        
        signal(SIGTTOU, SIG_IGN);
        tcsetpgrp(STDIN_FILENO, pgid);
        signal(SIGTTOU, SIG_DFL);
        
        for (int i = 0; i < pipeline->num_commands; i++) {
            waitpid(child_pids[i], &status, 0);
            
            if (i == pipeline->num_commands - 1) {
                if (WIFEXITED(status)) {
                    exit_code = WEXITSTATUS(status);
                } else if (WIFSIGNALED(status)) {
                    exit_code = 128 + WTERMSIG(status);
                }
            }
        }
        
        signal(SIGTTOU, SIG_IGN);
        tcsetpgrp(STDIN_FILENO, getpgrp());
        signal(SIGTTOU, SIG_DFL);
        
        last_exit_status = exit_code;
        return exit_code;
    }
}
