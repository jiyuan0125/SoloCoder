#define _GNU_SOURCE
#define _POSIX_C_SOURCE 200809L

#include "pidfile.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <signal.h>
#include <sys/stat.h>
#include <sys/file.h>

static char g_pidfile_path[512] = {0};
static int g_pidfile_fd = -1;

pid_t pidfile_get_current_pid(void) {
    return getpid();
}

const char *pidfile_get_current_path(void) {
    return g_pidfile_path[0] != '\0' ? g_pidfile_path : NULL;
}

static bool process_is_running(pid_t pid) {
    if (pid <= 0) {
        return false;
    }
    
    if (kill(pid, 0) == 0) {
        return true;
    }
    
    if (errno == ESRCH) {
        return false;
    }
    
    return true;
}

static pid_t read_pid_from_file(const char *path) {
    FILE *f = fopen(path, "r");
    if (f == NULL) {
        return -1;
    }
    
    pid_t pid = -1;
    if (fscanf(f, "%d", &pid) != 1) {
        fclose(f);
        return -1;
    }
    
    fclose(f);
    return pid;
}

PidFileResult pidfile_check(const char *path, bool *is_running) {
    if (is_running == NULL) {
        return PIDFILE_ERROR_INVALID_PATH;
    }
    
    *is_running = false;
    
    if (path == NULL || path[0] == '\0') {
        return PIDFILE_ERROR_INVALID_PATH;
    }
    
    if (access(path, F_OK) != 0) {
        return PIDFILE_OK;
    }
    
    pid_t pid = read_pid_from_file(path);
    if (pid <= 0) {
        return PIDFILE_OK;
    }
    
    if (process_is_running(pid)) {
        *is_running = true;
        return PIDFILE_OK;
    }
    
    return PIDFILE_OK;
}

PidFileResult pidfile_create(const char *path) {
    if (path == NULL || path[0] == '\0') {
        return PIDFILE_ERROR_INVALID_PATH;
    }
    
    if (g_pidfile_fd >= 0) {
        return PIDFILE_OK;
    }
    
    bool is_running = false;
    PidFileResult check_result = pidfile_check(path, &is_running);
    if (check_result != PIDFILE_OK) {
        return check_result;
    }
    
    if (is_running) {
        return PIDFILE_ERROR_ALREADY_RUNNING;
    }
    
    int fd = open(path, O_RDWR | O_CREAT, S_IRUSR | S_IWUSR | S_IRGRP | S_IROTH);
    if (fd < 0) {
        return PIDFILE_ERROR_OPEN;
    }
    
    if (flock(fd, LOCK_EX | LOCK_NB) != 0) {
        int saved_errno = errno;
        close(fd);
        errno = saved_errno;
        return PIDFILE_ERROR_LOCK;
    }
    
    if (ftruncate(fd, 0) != 0) {
        int saved_errno = errno;
        close(fd);
        errno = saved_errno;
        return PIDFILE_ERROR_WRITE;
    }
    
    char pid_str[32];
    snprintf(pid_str, sizeof(pid_str), "%d\n", (int)getpid());
    
    ssize_t written = write(fd, pid_str, strlen(pid_str));
    if (written < 0 || (size_t)written != strlen(pid_str)) {
        int saved_errno = errno;
        close(fd);
        errno = saved_errno;
        return PIDFILE_ERROR_WRITE;
    }
    
    strncpy(g_pidfile_path, path, sizeof(g_pidfile_path) - 1);
    g_pidfile_path[sizeof(g_pidfile_path) - 1] = '\0';
    g_pidfile_fd = fd;
    
    return PIDFILE_OK;
}

PidFileResult pidfile_remove(void) {
    if (g_pidfile_fd < 0) {
        return PIDFILE_OK;
    }
    
    if (g_pidfile_path[0] != '\0') {
        unlink(g_pidfile_path);
    }
    
    if (g_pidfile_fd >= 0) {
        flock(g_pidfile_fd, LOCK_UN);
        close(g_pidfile_fd);
        g_pidfile_fd = -1;
    }
    
    g_pidfile_path[0] = '\0';
    
    return PIDFILE_OK;
}
