#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <signal.h>
#include <sys/wait.h>
#include <termios.h>
#include "shell.h"

static volatile sig_atomic_t sigint_received = 0;

static void sigint_handler(int sig) {
    sigint_received = 1;
    killpg(getpgrp(), sig);
}

static void sigchld_handler(int sig) {
    int status;
    pid_t pid;
    
    while ((pid = waitpid(-1, &status, WNOHANG)) > 0) {
    }
}

static void setup_signals(void) {
    struct sigaction sa;
    
    sa.sa_handler = sigint_handler;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = 0;
    sigaction(SIGINT, &sa, NULL);
    
    sa.sa_handler = sigchld_handler;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_RESTART | SA_NOCLDSTOP;
    sigaction(SIGCHLD, &sa, NULL);
    
    signal(SIGQUIT, SIG_IGN);
    signal(SIGTSTP, SIG_IGN);
    signal(SIGTTIN, SIG_IGN);
    signal(SIGTTOU, SIG_IGN);
}

static void print_prompt(void) {
    char cwd[4096];
    char *home = getenv("HOME");
    const char *display_path;
    
    if (getcwd(cwd, sizeof(cwd)) != NULL) {
        if (home != NULL && strncmp(cwd, home, strlen(home)) == 0) {
            display_path = cwd + strlen(home);
            if (*display_path == '/' || *display_path == '\0') {
                printf("~%s $ ", display_path);
            } else {
                printf("%s $ ", cwd);
            }
        } else {
            printf("%s $ ", cwd);
        }
    } else {
        printf("$ ");
    }
    fflush(stdout);
}

int main(void) {
    char input[MAX_LINE];
    Pipeline pipeline;
    
    setpgrp();
    tcsetpgrp(STDIN_FILENO, getpgrp());
    
    setup_signals();
    
    while (1) {
        sigint_received = 0;
        
        print_prompt();
        
        if (fgets(input, sizeof(input), stdin) == NULL) {
            if (sigint_received) {
                clearerr(stdin);
                printf("\n");
                continue;
            }
            printf("\n");
            break;
        }
        
        size_t len = strlen(input);
        if (len > 0 && input[len - 1] == '\n') {
            input[len - 1] = '\0';
        }
        
        if (input[0] == '\0') {
            continue;
        }
        
        if (parse_input(input, &pipeline)) {
            execute_pipeline(&pipeline);
            free_pipeline(&pipeline);
        }
    }
    
    return last_exit_status;
}
