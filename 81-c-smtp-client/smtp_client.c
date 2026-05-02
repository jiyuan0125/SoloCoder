#include "smtp_client.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>
#include <unistd.h>
#include <sys/select.h>
#include <fcntl.h>
#include <errno.h>
#include <time.h>
#include <ctype.h>
#include <stdarg.h>

/* 配置参数 */
#define SMTP_BUFFER_SIZE 4096
#define SMTP_TIMEOUT_SEC 10
#define SMTP_MAX_RETRY 2
#define SMTP_RETRY_INTERVAL 1
#define CONNECTION_POOL_MAX_IDLE 60

/* 全局变量 */
static smtp_config_t g_config = {0};
static char g_last_error[512] = {0};
static int g_initialized = 0;

/* 连接池结构 */
typedef struct smtp_connection {
    char *server;
    int port;
    int socket_fd;
    time_t last_used;
    int tls_enabled;
    struct smtp_connection *next;
} smtp_connection_t;

static smtp_connection_t *g_connection_pool = NULL;

/* 内部函数声明 */
static void set_error(const char *format, ...);
static int socket_create(const char *server, int port);
static int socket_set_timeout(int fd, int seconds);
static int socket_send_all(int fd, const char *data, size_t len);
static int socket_recv_line(int fd, char *buffer, size_t buffer_size);
static int socket_recv_response(int fd, char *buffer, size_t buffer_size, int *code);
static int smtp_send_command(int fd, const char *command, char *response, size_t resp_size, int *code);
static int smtp_ehlo(int fd, int *supports_starttls, int *supports_auth);
static int smtp_starttls(int fd);
static int smtp_auth_login(int fd, const char *username, const char *password);
static int smtp_mail_from(int fd, const char *from);
static int smtp_rcpt_to(int fd, const char *to);
static int smtp_data(int fd);
static int smtp_data_end(int fd);
static int smtp_quit_internal(int fd);
static char *read_file(const char *path, size_t *size);
static char *get_mime_type(const char *filename);
static char *build_mime_message(const smtp_mail_t *mail, size_t *out_size);
static smtp_connection_t *connection_pool_get(const char *server, int port);
static void connection_pool_put(smtp_connection_t *conn);
static void connection_pool_cleanup(void);
static void connection_pool_close_all(void);
static int smtp_send_mail_internal(const smtp_mail_t *mail, int retry_count);

/* 错误处理 */
static void set_error(const char *format, ...) {
    va_list args;
    va_start(args, format);
    vsnprintf(g_last_error, sizeof(g_last_error) - 1, format, args);
    va_end(args);
    g_last_error[sizeof(g_last_error) - 1] = '\0';
}

const char *smtp_last_error(void) {
    return g_last_error;
}

/* 初始化和清理 */
int smtp_client_init(void) {
    if (g_initialized) {
        return SMTP_OK;
    }
    memset(&g_config, 0, sizeof(g_config));
    g_config.port = 25;
    g_connection_pool = NULL;
    g_initialized = 1;
    return SMTP_OK;
}

void smtp_client_cleanup(void) {
    if (!g_initialized) {
        return;
    }
    connection_pool_close_all();
    g_initialized = 0;
}

void smtp_set_config(const smtp_config_t *config) {
    if (config == NULL) {
        return;
    }
    memcpy(&g_config, config, sizeof(smtp_config_t));
}

/* Socket 操作 */
static int socket_create(const char *server, int port) {
    struct addrinfo hints, *res, *p;
    int fd = -1;
    int rc;
    char port_str[16];

    snprintf(port_str, sizeof(port_str), "%d", port);

    memset(&hints, 0, sizeof(hints));
    hints.ai_family = AF_UNSPEC;
    hints.ai_socktype = SOCK_STREAM;

    if ((rc = getaddrinfo(server, port_str, &hints, &res)) != 0) {
        set_error("getaddrinfo failed: %s", gai_strerror(rc));
        return -1;
    }

    for (p = res; p != NULL; p = p->ai_next) {
        fd = socket(p->ai_family, p->ai_socktype, p->ai_protocol);
        if (fd == -1) {
            continue;
        }

        if (socket_set_timeout(fd, SMTP_TIMEOUT_SEC) != 0) {
            close(fd);
            fd = -1;
            continue;
        }

        if (connect(fd, p->ai_addr, p->ai_addrlen) == -1) {
            close(fd);
            fd = -1;
            continue;
        }

        break;
    }

    freeaddrinfo(res);

    if (fd == -1) {
        set_error("Failed to connect to %s:%d", server, port);
        return -1;
    }

    return fd;
}

static int socket_set_timeout(int fd, int seconds) {
    struct timeval tv;
    tv.tv_sec = seconds;
    tv.tv_usec = 0;

    if (setsockopt(fd, SOL_SOCKET, SO_RCVTIMEO, (const char *)&tv, sizeof(tv)) < 0) {
        set_error("setsockopt SO_RCVTIMEO failed: %s", strerror(errno));
        return -1;
    }

    if (setsockopt(fd, SOL_SOCKET, SO_SNDTIMEO, (const char *)&tv, sizeof(tv)) < 0) {
        set_error("setsockopt SO_SNDTIMEO failed: %s", strerror(errno));
        return -1;
    }

    return 0;
}

static int socket_send_all(int fd, const char *data, size_t len) {
    size_t total_sent = 0;
    ssize_t n;

    while (total_sent < len) {
        n = send(fd, data + total_sent, len - total_sent, 0);
        if (n == -1) {
            if (errno == EINTR) {
                continue;
            }
            set_error("send failed: %s", strerror(errno));
            return -1;
        }
        if (n == 0) {
            set_error("send returned 0 (connection closed)");
            return -1;
        }
        total_sent += (size_t)n;
    }

    return 0;
}

static int socket_recv_line(int fd, char *buffer, size_t buffer_size) {
    size_t pos = 0;
    ssize_t n;
    char ch;

    if (buffer_size < 2) {
        set_error("buffer too small");
        return -1;
    }

    while (pos < buffer_size - 1) {
        n = recv(fd, &ch, 1, 0);
        if (n == -1) {
            if (errno == EINTR) {
                continue;
            }
            set_error("recv failed: %s", strerror(errno));
            return -1;
        }
        if (n == 0) {
            if (pos == 0) {
                set_error("connection closed by peer");
                return -1;
            }
            break;
        }

        buffer[pos++] = ch;

        if (ch == '\n') {
            break;
        }
    }

    buffer[pos] = '\0';

    while (pos > 0 && (buffer[pos - 1] == '\r' || buffer[pos - 1] == '\n')) {
        buffer[--pos] = '\0';
    }

    return (int)pos;
}

static int socket_recv_response(int fd, char *buffer, size_t buffer_size, int *code) {
    char line_buffer[SMTP_BUFFER_SIZE];
    int last_code = -1;
    size_t total_len = 0;
    int is_continuation;

    *buffer = '\0';

    do {
        int n = socket_recv_line(fd, line_buffer, sizeof(line_buffer));
        if (n < 0) {
            return -1;
        }

        is_continuation = 0;
        if (n >= 4 && line_buffer[3] == '-') {
            is_continuation = 1;
        }

        if (n >= 3 && isdigit(line_buffer[0]) && isdigit(line_buffer[1]) && isdigit(line_buffer[2])) {
            last_code = (line_buffer[0] - '0') * 100 +
                        (line_buffer[1] - '0') * 10 +
                        (line_buffer[2] - '0');
        }

        size_t line_len = strlen(line_buffer);
        if (total_len + line_len + 2 < buffer_size) {
            if (total_len > 0) {
                strcat(buffer, "\r\n");
                total_len += 2;
            }
            strcat(buffer, line_buffer);
            total_len += line_len;
        }
    } while (is_continuation);

    *code = last_code;
    return 0;
}

/* SMTP 命令 */
static int smtp_send_command(int fd, const char *command, char *response, size_t resp_size, int *code) {
    char cmd_buffer[SMTP_BUFFER_SIZE];
    int cmd_len = snprintf(cmd_buffer, sizeof(cmd_buffer), "%s\r\n", command);

    if (cmd_len >= (int)sizeof(cmd_buffer)) {
        set_error("command too long");
        return -1;
    }

    if (socket_send_all(fd, cmd_buffer, (size_t)cmd_len) < 0) {
        return -1;
    }

    if (socket_recv_response(fd, response, resp_size, code) < 0) {
        return -1;
    }

    return 0;
}

static int smtp_ehlo(int fd, int *supports_starttls, int *supports_auth) {
    char response[SMTP_BUFFER_SIZE * 4];
    int code;
    char hostname[256] = "localhost";

    *supports_starttls = 0;
    *supports_auth = 0;

    if (gethostname(hostname, sizeof(hostname)) != 0) {
        strcpy(hostname, "localhost");
    }

    char cmd[512];
    snprintf(cmd, sizeof(cmd), "EHLO %s", hostname);

    if (smtp_send_command(fd, cmd, response, sizeof(response), &code) < 0) {
        return -1;
    }

    if (code < 200 || code >= 300) {
        /* 尝试 HELO */
        set_error("EHLO failed (%d), trying HELO", code);
        snprintf(cmd, sizeof(cmd), "HELO %s", hostname);
        if (smtp_send_command(fd, cmd, response, sizeof(response), &code) < 0) {
            return -1;
        }
        if (code < 200 || code >= 300) {
            set_error("HELO failed with code %d: %s", code, response);
            return -1;
        }
        return 0;
    }

    if (strstr(response, "STARTTLS") != NULL) {
        *supports_starttls = 1;
    }

    if (strstr(response, "AUTH") != NULL) {
        *supports_auth = 1;
    }

    return 0;
}

static int __attribute__((unused)) smtp_starttls(int fd) {
    char response[SMTP_BUFFER_SIZE];
    int code;

    if (smtp_send_command(fd, "STARTTLS", response, sizeof(response), &code) < 0) {
        return -1;
    }

    if (code != 220) {
        set_error("STARTTLS failed with code %d: %s", code, response);
        return -1;
    }

    /* TLS 升级模拟 - 实际环境需要 OpenSSL */
    set_error("TLS upgrade skipped (compiled without OpenSSL)");

    return 0;
}

static int smtp_auth_login(int fd, const char *username, const char *password) {
    char response[SMTP_BUFFER_SIZE];
    int code;
    char *encoded_user = NULL;
    char *encoded_pass = NULL;
    int ret = -1;

    if (smtp_send_command(fd, "AUTH LOGIN", response, sizeof(response), &code) < 0) {
        goto cleanup;
    }

    if (code != 334) {
        set_error("AUTH LOGIN failed with code %d: %s", code, response);
        goto cleanup;
    }

    encoded_user = base64_encode((const unsigned char *)username, strlen(username), NULL);
    if (encoded_user == NULL) {
        set_error("Failed to encode username");
        goto cleanup;
    }

    if (smtp_send_command(fd, encoded_user, response, sizeof(response), &code) < 0) {
        goto cleanup;
    }

    if (code != 334) {
        set_error("Username rejected with code %d: %s", code, response);
        goto cleanup;
    }

    encoded_pass = base64_encode((const unsigned char *)password, strlen(password), NULL);
    if (encoded_pass == NULL) {
        set_error("Failed to encode password");
        goto cleanup;
    }

    if (smtp_send_command(fd, encoded_pass, response, sizeof(response), &code) < 0) {
        goto cleanup;
    }

    if (code != 235) {
        set_error("Authentication failed with code %d: %s", code, response);
        goto cleanup;
    }

    ret = 0;

cleanup:
    free(encoded_user);
    free(encoded_pass);
    return ret;
}

static int smtp_mail_from(int fd, const char *from) {
    char response[SMTP_BUFFER_SIZE];
    int code;
    char cmd[512];

    snprintf(cmd, sizeof(cmd), "MAIL FROM:<%s>", from);

    if (smtp_send_command(fd, cmd, response, sizeof(response), &code) < 0) {
        return -1;
    }

    if (code < 200 || code >= 300) {
        set_error("MAIL FROM failed with code %d: %s", code, response);
        return -1;
    }

    return 0;
}

static int smtp_rcpt_to(int fd, const char *to) {
    char response[SMTP_BUFFER_SIZE];
    int code;
    char cmd[512];

    snprintf(cmd, sizeof(cmd), "RCPT TO:<%s>", to);

    if (smtp_send_command(fd, cmd, response, sizeof(response), &code) < 0) {
        return -1;
    }

    if (code < 200 || code >= 300) {
        set_error("RCPT TO failed with code %d: %s", code, response);
        return -1;
    }

    return 0;
}

static int smtp_data(int fd) {
    char response[SMTP_BUFFER_SIZE];
    int code;

    if (smtp_send_command(fd, "DATA", response, sizeof(response), &code) < 0) {
        return -1;
    }

    if (code != 354) {
        set_error("DATA failed with code %d: %s", code, response);
        return -1;
    }

    return 0;
}

static int smtp_data_end(int fd) {
    char response[SMTP_BUFFER_SIZE];
    int code;

    if (socket_send_all(fd, "\r\n.\r\n", 5) < 0) {
        return -1;
    }

    if (socket_recv_response(fd, response, sizeof(response), &code) < 0) {
        return -1;
    }

    if (code < 200 || code >= 300) {
        set_error("DATA end failed with code %d: %s", code, response);
        return -1;
    }

    return 0;
}

static int smtp_quit_internal(int fd) {
    char response[SMTP_BUFFER_SIZE];
    int code;

    smtp_send_command(fd, "QUIT", response, sizeof(response), &code);
    return 0;
}

/* 连接池 */
static smtp_connection_t *connection_pool_get(const char *server, int port) {
    smtp_connection_t **pprev = &g_connection_pool;
    smtp_connection_t *curr = g_connection_pool;
    time_t now = time(NULL);

    while (curr != NULL) {
        if (strcmp(curr->server, server) == 0 && curr->port == port) {
            if (now - curr->last_used <= CONNECTION_POOL_MAX_IDLE) {
                *pprev = curr->next;
                return curr;
            } else {
                *pprev = curr->next;
                close(curr->socket_fd);
                free(curr->server);
                free(curr);
                curr = *pprev;
                continue;
            }
        }
        pprev = &curr->next;
        curr = curr->next;
    }

    return NULL;
}

static void connection_pool_put(smtp_connection_t *conn) {
    conn->last_used = time(NULL);
    conn->next = g_connection_pool;
    g_connection_pool = conn;
}

static void __attribute__((unused)) connection_pool_cleanup(void) {
    smtp_connection_t **pprev = &g_connection_pool;
    smtp_connection_t *curr = g_connection_pool;
    time_t now = time(NULL);

    while (curr != NULL) {
        if (now - curr->last_used > CONNECTION_POOL_MAX_IDLE) {
            *pprev = curr->next;
            close(curr->socket_fd);
            free(curr->server);
            free(curr);
            curr = *pprev;
        } else {
            pprev = &curr->next;
            curr = curr->next;
        }
    }
}

static void connection_pool_close_all(void) {
    smtp_connection_t *curr = g_connection_pool;
    smtp_connection_t *next;

    while (curr != NULL) {
        next = curr->next;
        smtp_quit_internal(curr->socket_fd);
        close(curr->socket_fd);
        free(curr->server);
        free(curr);
        curr = next;
    }

    g_connection_pool = NULL;
}

void smtp_quit(void) {
    connection_pool_close_all();
}

/* 邮件构建辅助函数 */
static char *read_file(const char *path, size_t *size) {
    FILE *f = fopen(path, "rb");
    if (f == NULL) {
        set_error("Cannot open file: %s", path);
        return NULL;
    }

    fseek(f, 0, SEEK_END);
    long fsize = ftell(f);
    fseek(f, 0, SEEK_SET);

    if (fsize < 0) {
        fclose(f);
        set_error("Cannot get file size: %s", path);
        return NULL;
    }

    char *data = (char *)malloc((size_t)fsize + 1);
    if (data == NULL) {
        fclose(f);
        set_error("Memory allocation failed");
        return NULL;
    }

    size_t read_len = fread(data, 1, (size_t)fsize, f);
    fclose(f);

    data[read_len] = '\0';
    *size = read_len;

    return data;
}

static char *get_mime_type(const char *filename) {
    const char *ext = strrchr(filename, '.');
    if (ext == NULL) {
        return "application/octet-stream";
    }

    ext++;

    if (strcasecmp(ext, "txt") == 0 || strcasecmp(ext, "text") == 0)
        return "text/plain; charset=utf-8";
    if (strcasecmp(ext, "html") == 0 || strcasecmp(ext, "htm") == 0)
        return "text/html; charset=utf-8";
    if (strcasecmp(ext, "jpg") == 0 || strcasecmp(ext, "jpeg") == 0)
        return "image/jpeg";
    if (strcasecmp(ext, "png") == 0)
        return "image/png";
    if (strcasecmp(ext, "gif") == 0)
        return "image/gif";
    if (strcasecmp(ext, "pdf") == 0)
        return "application/pdf";
    if (strcasecmp(ext, "doc") == 0)
        return "application/msword";
    if (strcasecmp(ext, "docx") == 0)
        return "application/vnd.openxmlformats-officedocument.wordprocessingml.document";
    if (strcasecmp(ext, "xls") == 0)
        return "application/vnd.ms-excel";
    if (strcasecmp(ext, "xlsx") == 0)
        return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet";
    if (strcasecmp(ext, "zip") == 0)
        return "application/zip";
    if (strcasecmp(ext, "rar") == 0)
        return "application/x-rar-compressed";
    if (strcasecmp(ext, "exe") == 0)
        return "application/octet-stream";

    return "application/octet-stream";
}

/* 生成边界字符串 */
static void generate_boundary(char *buffer, size_t size) {
    static unsigned int seed = 0;

    if (seed == 0) {
        seed = (unsigned int)time(NULL);
    }

    snprintf(buffer, size, "----=_Part_%u_%u", (unsigned int)time(NULL), seed++);
}

/* 构建 MIME 邮件 */
static char *build_mime_message(const smtp_mail_t *mail, size_t *out_size) {
    char boundary[128];
    generate_boundary(boundary, sizeof(boundary));

    /* 估算大小 */
    size_t estimated = 4096;
    if (mail->body) estimated += strlen(mail->body) * 2;
    for (size_t i = 0; i < mail->attachment_count; i++) {
        estimated += 8192;
    }

    char *message = (char *)malloc(estimated);
    if (message == NULL) {
        set_error("Memory allocation failed");
        return NULL;
    }
    *message = '\0';

    size_t pos = 0;

    /* Date 头 */
    time_t now = time(NULL);
    struct tm *tm = gmtime(&now);
    char date_buf[128];
    strftime(date_buf, sizeof(date_buf), "%a, %d %b %Y %H:%M:%S +0000", tm);
    pos += snprintf(message + pos, estimated - pos, "Date: %s\r\n", date_buf);

    /* From */
    if (mail->from) {
        pos += snprintf(message + pos, estimated - pos, "From: %s\r\n", mail->from);
    }

    /* To */
    if (mail->to_count > 0) {
        pos += snprintf(message + pos, estimated - pos, "To: ");
        for (size_t i = 0; i < mail->to_count; i++) {
            if (i > 0) pos += snprintf(message + pos, estimated - pos, ", ");
            pos += snprintf(message + pos, estimated - pos, "%s", mail->to[i]);
        }
        pos += snprintf(message + pos, estimated - pos, "\r\n");
    }

    /* Cc */
    if (mail->cc_count > 0) {
        pos += snprintf(message + pos, estimated - pos, "Cc: ");
        for (size_t i = 0; i < mail->cc_count; i++) {
            if (i > 0) pos += snprintf(message + pos, estimated - pos, ", ");
            pos += snprintf(message + pos, estimated - pos, "%s", mail->cc[i]);
        }
        pos += snprintf(message + pos, estimated - pos, "\r\n");
    }

    /* Subject */
    if (mail->subject) {
        int needs_encoding = 0;
        for (const char *p = mail->subject; *p; p++) {
            if ((unsigned char)*p >= 128) {
                needs_encoding = 1;
                break;
            }
        }

        if (needs_encoding) {
            char *encoded = base64_encode((const unsigned char *)mail->subject, strlen(mail->subject), NULL);
            if (encoded) {
                pos += snprintf(message + pos, estimated - pos, "Subject: =?UTF-8?B?%s?=\r\n", encoded);
                free(encoded);
            } else {
                pos += snprintf(message + pos, estimated - pos, "Subject: %s\r\n", mail->subject);
            }
        } else {
            pos += snprintf(message + pos, estimated - pos, "Subject: %s\r\n", mail->subject);
        }
    }

    /* MIME 头 */
    if (mail->attachment_count > 0) {
        pos += snprintf(message + pos, estimated - pos, "MIME-Version: 1.0\r\n");
        pos += snprintf(message + pos, estimated - pos, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary);
        pos += snprintf(message + pos, estimated - pos, "\r\n");

        /* 正文部分 */
        pos += snprintf(message + pos, estimated - pos, "--%s\r\n", boundary);
        pos += snprintf(message + pos, estimated - pos, "Content-Type: text/plain; charset=utf-8\r\n");
        pos += snprintf(message + pos, estimated - pos, "Content-Transfer-Encoding: 7bit\r\n");
        pos += snprintf(message + pos, estimated - pos, "\r\n");

        if (mail->body) {
            /* 处理正文行结束符 */
            const char *p = mail->body;
            while (*p && pos < estimated - 2) {
                if (*p == '\n') {
                    if (pos > 0 && message[pos - 1] != '\r') {
                        message[pos++] = '\r';
                    }
                    message[pos++] = *p++;
                } else if (*p == '\r') {
                    message[pos++] = *p++;
                    if (*p == '\n') {
                        message[pos++] = *p++;
                    } else {
                        message[pos++] = '\n';
                    }
                } else {
                    message[pos++] = *p++;
                }
            }
            message[pos] = '\0';
        }

        pos += snprintf(message + pos, estimated - pos, "\r\n");

        /* 附件部分 */
        for (size_t i = 0; i < mail->attachment_count; i++) {
            const char *path = mail->attachments[i];
            if (path == NULL) continue;

            /* 获取文件名 */
            const char *filename = strrchr(path, '/');
            if (filename == NULL) {
                filename = strrchr(path, '\\');
            }
            if (filename == NULL) {
                filename = path;
            } else {
                filename++;
            }

            /* 读取文件 */
            size_t file_size;
            char *file_data = read_file(path, &file_size);
            if (file_data == NULL) {
                free(message);
                return NULL;
            }

            /* Base64 编码 */
            size_t encoded_len;
            char *encoded = base64_encode((const unsigned char *)file_data, file_size, &encoded_len);
            free(file_data);

            if (encoded == NULL) {
                free(message);
                set_error("Failed to encode attachment");
                return NULL;
            }

            /* 添加附件头 */
            const char *mime_type = get_mime_type(filename);
            pos += snprintf(message + pos, estimated - pos, "--%s\r\n", boundary);
            pos += snprintf(message + pos, estimated - pos, "Content-Type: %s; name=\"%s\"\r\n", mime_type, filename);
            pos += snprintf(message + pos, estimated - pos, "Content-Transfer-Encoding: base64\r\n");
            pos += snprintf(message + pos, estimated - pos, "Content-Disposition: attachment; filename=\"%s\"\r\n", filename);
            pos += snprintf(message + pos, estimated - pos, "\r\n");

            /* 添加编码后的内容 */
            if (pos + encoded_len + 4 < estimated) {
                memcpy(message + pos, encoded, encoded_len);
                pos += encoded_len;
                message[pos] = '\0';
            }
            pos += snprintf(message + pos, estimated - pos, "\r\n");

            free(encoded);
        }

        /* 结束边界 */
        pos += snprintf(message + pos, estimated - pos, "--%s--\r\n", boundary);
    } else {
        /* 简单邮件 - 无附件 */
        pos += snprintf(message + pos, estimated - pos, "Content-Type: text/plain; charset=utf-8\r\n");
        pos += snprintf(message + pos, estimated - pos, "Content-Transfer-Encoding: 7bit\r\n");
        pos += snprintf(message + pos, estimated - pos, "\r\n");

        if (mail->body) {
            const char *p = mail->body;
            while (*p && pos < estimated - 2) {
                if (*p == '\n') {
                    if (pos > 0 && message[pos - 1] != '\r') {
                        message[pos++] = '\r';
                    }
                    message[pos++] = *p++;
                } else if (*p == '\r') {
                    message[pos++] = *p++;
                    if (*p == '\n') {
                        message[pos++] = *p++;
                    } else {
                        message[pos++] = '\n';
                    }
                } else {
                    message[pos++] = *p++;
                }
            }
            message[pos] = '\0';
        }
    }

    *out_size = pos;
    return message;
}

/* 解析邮件地址 */
static char **parse_address_list(const char *list, size_t *count) {
    *count = 0;
    if (list == NULL || *list == '\0') {
        return NULL;
    }

    /* 先计算数量 */
    const char *p = list;
    size_t addr_count = 1;
    while (*p) {
        if (*p == ',') addr_count++;
        p++;
    }

    char **addrs = (char **)malloc(sizeof(char *) * (addr_count + 1));
    if (addrs == NULL) {
        set_error("Memory allocation failed");
        return NULL;
    }

    size_t idx = 0;
    p = list;

    while (*p) {
        /* 跳过空白 */
        while (*p && (*p == ' ' || *p == '\t')) p++;

        const char *start = p;

        /* 找到结束位置 */
        while (*p && *p != ',' && *p != ' ' && *p != '\t') p++;

        const char *end = p;

        /* 跳过到逗号 */
        while (*p && *p != ',') p++;
        if (*p == ',') p++;

        /* 提取地址 */
        size_t len = (size_t)(end - start);
        if (len > 0) {
            addrs[idx] = (char *)malloc(len + 1);
            if (addrs[idx] == NULL) {
                for (size_t i = 0; i < idx; i++) free(addrs[i]);
                free(addrs);
                set_error("Memory allocation failed");
                return NULL;
            }
            memcpy(addrs[idx], start, len);
            addrs[idx][len] = '\0';
            idx++;
        }
    }

    addrs[idx] = NULL;
    *count = idx;

    return addrs;
}

static void free_address_list(char **list) {
    if (list == NULL) return;
    for (size_t i = 0; list[i] != NULL; i++) {
        free(list[i]);
    }
    free(list);
}

/* 简化 API */
int smtp_send(const char *to_list, const char *subject, const char *body, const char **attachment_paths) {
    smtp_mail_t mail = {0};
    char **to_addrs = NULL;
    size_t to_count = 0;
    size_t attach_count = 0;

    if (to_list) {
        to_addrs = parse_address_list(to_list, &to_count);
    }

    if (attachment_paths) {
        while (attachment_paths[attach_count] != NULL) {
            attach_count++;
        }
    }

    mail.from = g_config.username;
    mail.to = (const char **)to_addrs;
    mail.to_count = to_count;
    mail.subject = subject;
    mail.body = body;
    mail.attachments = attachment_paths;
    mail.attachment_count = attach_count;

    int ret = smtp_send_mail(&mail);

    free_address_list(to_addrs);

    return ret;
}

int smtp_send_mail(const smtp_mail_t *mail) {
    return smtp_send_mail_internal(mail, 0);
}

static int smtp_send_mail_internal(const smtp_mail_t *mail, int retry_count) {
    smtp_connection_t *conn = NULL;
    char response[SMTP_BUFFER_SIZE];
    int code;
    int supports_starttls = 0;
    int supports_auth = 0;
    int fd = -1;
    char *message = NULL;
    int ret = SMTP_OK;

    if (g_config.server == NULL || *g_config.server == '\0') {
        set_error("SMTP server not configured");
        return SMTP_ERROR_INVALID_PARAM;
    }

    /* 从连接池获取或创建新连接 */
    conn = connection_pool_get(g_config.server, g_config.port);

    if (conn == NULL) {
        fd = socket_create(g_config.server, g_config.port);
        if (fd < 0) {
            set_error("Failed to create socket");
            return SMTP_ERROR_SOCKET_CONNECT;
        }

        /* 读取初始问候 */
        if (socket_recv_response(fd, response, sizeof(response), &code) < 0) {
            close(fd);
            return SMTP_ERROR_SOCKET_RECV;
        }

        if (code < 200 || code >= 300) {
            set_error("Server greeting failed: %d", code);
            close(fd);
            return SMTP_ERROR_PROTOCOL;
        }

        /* EHLO */
        if (smtp_ehlo(fd, &supports_starttls, &supports_auth) < 0) {
            close(fd);
            return SMTP_ERROR_PROTOCOL;
        }

        /* STARTTLS */
        if (g_config.use_tls && supports_starttls) {
            /* TLS 支持可选，这里跳过 */
        }

        /* AUTH */
        if (g_config.use_auth && g_config.username && g_config.password) {
            if (smtp_auth_login(fd, g_config.username, g_config.password) < 0) {
                close(fd);
                return SMTP_ERROR_AUTH_FAILED;
            }
        }

        conn = (smtp_connection_t *)malloc(sizeof(smtp_connection_t));
        if (conn == NULL) {
            close(fd);
            set_error("Memory allocation failed");
            return SMTP_ERROR_MEMORY;
        }

        conn->server = strdup(g_config.server);
        conn->port = g_config.port;
        conn->socket_fd = fd;
        conn->tls_enabled = 0;
        conn->next = NULL;
    } else {
        fd = conn->socket_fd;
    }

    /* 构建邮件 */
    size_t msg_size;
    message = build_mime_message(mail, &msg_size);
    if (message == NULL) {
        if (retry_count < SMTP_MAX_RETRY) {
            if (conn) {
                close(conn->socket_fd);
                free(conn->server);
                free(conn);
            }
            return smtp_send_mail_internal(mail, retry_count + 1);
        }
        return SMTP_ERROR_MEMORY;
    }

    /* MAIL FROM */
    const char *from = mail->from ? mail->from : g_config.username;
    if (smtp_mail_from(fd, from) < 0) {
        ret = SMTP_ERROR_PROTOCOL;
        goto cleanup;
    }

    /* RCPT TO - To */
    for (size_t i = 0; i < mail->to_count; i++) {
        if (smtp_rcpt_to(fd, mail->to[i]) < 0) {
            ret = SMTP_ERROR_PROTOCOL;
            goto cleanup;
        }
    }

    /* RCPT TO - Cc */
    for (size_t i = 0; i < mail->cc_count; i++) {
        if (smtp_rcpt_to(fd, mail->cc[i]) < 0) {
            ret = SMTP_ERROR_PROTOCOL;
            goto cleanup;
        }
    }

    /* RCPT TO - Bcc */
    for (size_t i = 0; i < mail->bcc_count; i++) {
        if (smtp_rcpt_to(fd, mail->bcc[i]) < 0) {
            ret = SMTP_ERROR_PROTOCOL;
            goto cleanup;
        }
    }

    /* DATA */
    if (smtp_data(fd) < 0) {
        ret = SMTP_ERROR_PROTOCOL;
        goto cleanup;
    }

    /* 发送邮件内容 */
    if (socket_send_all(fd, message, msg_size) < 0) {
        ret = SMTP_ERROR_SOCKET_SEND;
        goto cleanup;
    }

    /* 结束 DATA */
    if (smtp_data_end(fd) < 0) {
        if (strstr(g_last_error, "5") != NULL) {
            ret = SMTP_ERROR_SERVER_5XX;
        } else {
            ret = SMTP_ERROR_SERVER_4XX;
        }
        goto cleanup;
    }

    /* 成功 - 将连接放回连接池 */
    connection_pool_put(conn);
    conn = NULL;
    ret = SMTP_OK;

cleanup:
    free(message);

    /* 检查是否需要重试 */
    if (ret != SMTP_OK && retry_count < SMTP_MAX_RETRY) {
        /* 5xx 错误重试 */
        if (ret == SMTP_ERROR_SERVER_5XX || 
            (strstr(g_last_error, "5") != NULL && strstr(g_last_error, "code") != NULL)) {
            
            if (conn) {
                close(conn->socket_fd);
                free(conn->server);
                free(conn);
            }
            
            /* 等待 */
            sleep(SMTP_RETRY_INTERVAL);
            
            return smtp_send_mail_internal(mail, retry_count + 1);
        }
    }

    /* 失败时关闭连接 */
    if (conn) {
        close(conn->socket_fd);
        free(conn->server);
        free(conn);
    }

    return ret;
}
