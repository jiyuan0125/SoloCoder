#ifndef SMTP_CLIENT_H
#define SMTP_CLIENT_H

#include <stddef.h>
#include <time.h>

/* 错误码定义 */
enum {
    SMTP_OK = 0,
    SMTP_ERROR_INVALID_PARAM = -1,
    SMTP_ERROR_SOCKET_CREATE = -2,
    SMTP_ERROR_SOCKET_CONNECT = -3,
    SMTP_ERROR_SOCKET_TIMEOUT = -4,
    SMTP_ERROR_SOCKET_SEND = -5,
    SMTP_ERROR_SOCKET_RECV = -6,
    SMTP_ERROR_PROTOCOL = -7,
    SMTP_ERROR_AUTH_FAILED = -8,
    SMTP_ERROR_TLS_FAILED = -9,
    SMTP_ERROR_FILE_NOT_FOUND = -10,
    SMTP_ERROR_MEMORY = -11,
    SMTP_ERROR_SERVER_5XX = -12,
    SMTP_ERROR_SERVER_4XX = -13,
    SMTP_ERROR_CONNECTION_POOL = -14
};

/* SMTP 配置结构 */
typedef struct {
    const char *server;
    int port;
    const char *username;
    const char *password;
    int use_tls;
    int use_auth;
} smtp_config_t;

/* 邮件结构 */
typedef struct {
    const char *from;
    const char **to;
    size_t to_count;
    const char **cc;
    size_t cc_count;
    const char **bcc;
    size_t bcc_count;
    const char *subject;
    const char *body;
    const char **attachments;
    size_t attachment_count;
} smtp_mail_t;

/* 连接池结构 (前向声明) */
typedef struct smtp_connection_pool smtp_connection_pool_t;

/* API 函数 */

/**
 * @brief 初始化 SMTP 客户端库
 * @return 0 成功，非 0 失败
 */
int smtp_client_init(void);

/**
 * @brief 清理 SMTP 客户端库资源
 */
void smtp_client_cleanup(void);

/**
 * @brief 设置 SMTP 配置
 * @param config 配置结构体指针
 */
void smtp_set_config(const smtp_config_t *config);

/**
 * @brief 发送邮件
 * @param to_list 收件人列表（逗号分隔）
 * @param subject 邮件主题
 * @param body 邮件正文
 * @param attachment_paths 附件路径列表（NULL 结尾）
 * @return 0 成功，非 0 错误码
 */
int smtp_send(const char *to_list, const char *subject, const char *body, const char **attachment_paths);

/**
 * @brief 发送邮件（完整参数版本）
 * @param mail 邮件结构指针
 * @return 0 成功，非 0 错误码
 */
int smtp_send_mail(const smtp_mail_t *mail);

/**
 * @brief 发送 QUIT 命令并关闭所有连接
 */
void smtp_quit(void);

/**
 * @brief 获取最后一次错误信息
 * @return 错误信息字符串
 */
const char *smtp_last_error(void);

/* Base64 编解码函数（公开用于测试） */

/**
 * @brief Base64 编码
 * @param data 输入数据
 * @param input_length 输入长度
 * @param output_length 输出长度（可选）
 * @return 编码后的字符串（需要调用 free 释放），失败返回 NULL
 */
char *base64_encode(const unsigned char *data, size_t input_length, size_t *output_length);

/**
 * @brief Base64 解码
 * @param data 输入 Base64 字符串
 * @param input_length 输入长度
 * @param output_length 输出长度
 * @return 解码后的数据（需要调用 free 释放），失败返回 NULL
 */
unsigned char *base64_decode(const char *data, size_t input_length, size_t *output_length);

#endif /* SMTP_CLIENT_H */
