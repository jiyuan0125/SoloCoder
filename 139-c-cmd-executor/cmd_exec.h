#ifndef CMD_EXEC_H
#define CMD_EXEC_H

#include "common.h"

int cmd_config_init(cmd_config_t *config);
void cmd_config_free(cmd_config_t *config);
int cmd_config_set_env(cmd_config_t *config, const char *key, const char *value, bool overwrite);

int cmd_result_init(cmd_result_t *result, cmd_config_t *config);
void cmd_result_free(cmd_result_t *result);

int output_buf_init(output_buf_t *buf, size_t max_size);
void output_buf_free(output_buf_t *buf);
ssize_t output_buf_append(output_buf_t *buf, const char *data, size_t len, bool is_stdout);

int cmd_start(cmd_result_t *result);
int cmd_wait(cmd_result_t *result, int timeout_sec);
int cmd_kill_process_group(pid_t pgid);
int cmd_read_pipes(cmd_result_t *result);
int cmd_execute_sync(cmd_config_t *config, cmd_result_t *result);

#endif
