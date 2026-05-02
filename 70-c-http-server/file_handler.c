#include "http_server.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>
#include <libgen.h>

const char *get_content_type(const char *path) {
    const char *ext = strrchr(path, '.');
    if (!ext) {
        return "application/octet-stream";
    }
    
    ext++;
    
    if (strcasecmp(ext, "html") == 0 || strcasecmp(ext, "htm") == 0) {
        return "text/html";
    } else if (strcasecmp(ext, "css") == 0) {
        return "text/css";
    } else if (strcasecmp(ext, "js") == 0) {
        return "application/javascript";
    } else if (strcasecmp(ext, "json") == 0) {
        return "application/json";
    } else if (strcasecmp(ext, "png") == 0) {
        return "image/png";
    } else if (strcasecmp(ext, "jpg") == 0 || strcasecmp(ext, "jpeg") == 0) {
        return "image/jpeg";
    } else if (strcasecmp(ext, "gif") == 0) {
        return "image/gif";
    } else if (strcasecmp(ext, "ico") == 0) {
        return "image/x-icon";
    } else if (strcasecmp(ext, "txt") == 0) {
        return "text/plain";
    } else if (strcasecmp(ext, "svg") == 0) {
        return "image/svg+xml";
    }
    
    return "application/octet-stream";
}

int is_directory_traversal(const char *path) {
    if (!path || path[0] == '\0') {
        return 0;
    }
    
    const char *ptr = path;
    while (*ptr) {
        if (*ptr == '.' && *(ptr + 1) == '.') {
            if (*(ptr + 2) == '/' || *(ptr + 2) == '\0' ||
                (ptr == path || *(ptr - 1) == '/')) {
                return 1;
            }
        }
        ptr++;
    }
    
    return 0;
}

int resolve_path(const char *root_dir, const char *url_path, char *resolved_path, size_t max_len) {
    if (is_directory_traversal(url_path)) {
        return -1;
    }
    
    size_t root_len = strlen(root_dir);
    size_t path_len = strlen(url_path);
    
    if (root_len + path_len + 2 > max_len) {
        return -1;
    }
    
    if (root_dir[root_len - 1] == '/') {
        if (url_path[0] == '/') {
            snprintf(resolved_path, max_len, "%s%s", root_dir, url_path + 1);
        } else {
            snprintf(resolved_path, max_len, "%s%s", root_dir, url_path);
        }
    } else {
        if (url_path[0] == '/') {
            snprintf(resolved_path, max_len, "%s%s", root_dir, url_path);
        } else {
            snprintf(resolved_path, max_len, "%s/%s", root_dir, url_path);
        }
    }
    
    return 0;
}

int file_handler_serve(HttpResponse *resp, const char *root_dir, const char *url_path) {
    char resolved_path[8192];
    
    if (resolve_path(root_dir, url_path, resolved_path, sizeof(resolved_path)) != 0) {
        http_response_set_status(resp, HTTP_FORBIDDEN);
        return -1;
    }
    
    struct stat st;
    if (stat(resolved_path, &st) != 0) {
        http_response_set_status(resp, HTTP_NOT_FOUND);
        return -1;
    }
    
    if (S_ISDIR(st.st_mode)) {
        size_t len = strlen(resolved_path);
        if (len + 12 >= sizeof(resolved_path)) {
            http_response_set_status(resp, HTTP_NOT_FOUND);
            return -1;
        }
        
        if (resolved_path[len - 1] != '/') {
            strcat(resolved_path, "/");
        }
        strcat(resolved_path, "index.html");
        
        if (stat(resolved_path, &st) != 0) {
            http_response_set_status(resp, HTTP_NOT_FOUND);
            return -1;
        }
    }
    
    if (!S_ISREG(st.st_mode)) {
        http_response_set_status(resp, HTTP_FORBIDDEN);
        return -1;
    }
    
    int fd = open(resolved_path, O_RDONLY);
    if (fd < 0) {
        http_response_set_status(resp, HTTP_FORBIDDEN);
        return -1;
    }
    
    char *file_content = (char *)malloc(st.st_size);
    if (!file_content) {
        close(fd);
        http_response_set_status(resp, HTTP_INTERNAL_SERVER_ERROR);
        return -1;
    }
    
    ssize_t bytes_read = read(fd, file_content, st.st_size);
    close(fd);
    
    if (bytes_read != st.st_size) {
        free(file_content);
        http_response_set_status(resp, HTTP_INTERNAL_SERVER_ERROR);
        return -1;
    }
    
    http_response_set_status(resp, HTTP_OK);
    http_response_set_body(resp, file_content, (size_t)bytes_read);
    free(file_content);
    
    const char *content_type = get_content_type(resolved_path);
    char cl_buf[32];
    snprintf(cl_buf, sizeof(cl_buf), "%zu", resp->body_length);
    
    http_response_add_header(resp, "Content-Type", content_type);
    http_response_add_header(resp, "Content-Length", cl_buf);
    
    return 0;
}
