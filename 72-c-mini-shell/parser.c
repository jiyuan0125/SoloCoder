#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include "shell.h"

static char *trim(char *str) {
    while (isspace((unsigned char)*str)) str++;
    if (*str == '\0') return str;
    char *end = str + strlen(str) - 1;
    while (end > str && isspace((unsigned char)*end)) end--;
    end[1] = '\0';
    return str;
}

static char *skip_whitespace(const char *str) {
    while (*str && isspace((unsigned char)*str)) str++;
    return (char *)str;
}

static char *get_token(char **str) {
    char *start = skip_whitespace(*str);
    if (*start == '\0') return NULL;
    
    char *end = start;
    int in_single_quote = 0;
    int in_double_quote = 0;
    
    while (*end) {
        if (*end == '\'' && !in_double_quote) {
            in_single_quote = !in_single_quote;
            end++;
        } else if (*end == '"' && !in_single_quote) {
            in_double_quote = !in_double_quote;
            end++;
        } else if (!in_single_quote && !in_double_quote && isspace((unsigned char)*end)) {
            break;
        } else {
            end++;
        }
    }
    
    if (*end != '\0') {
        *end = '\0';
        *str = end + 1;
    } else {
        *str = end;
    }
    
    return start;
}

static char *remove_quotes(const char *str) {
    if (str == NULL) return NULL;
    size_t len = strlen(str);
    char *result = malloc(len + 1);
    if (result == NULL) return NULL;
    
    char *dst = result;
    int in_single_quote = 0;
    int in_double_quote = 0;
    
    for (const char *src = str; *src; src++) {
        if (*src == '\'' && !in_double_quote) {
            in_single_quote = !in_single_quote;
        } else if (*src == '"' && !in_single_quote) {
            in_double_quote = !in_double_quote;
        } else {
            *dst++ = *src;
        }
    }
    *dst = '\0';
    
    return result;
}

char *expand_env_vars(const char *str) {
    if (str == NULL) return NULL;
    
    size_t result_len = strlen(str) * 2 + 1;
    char *result = malloc(result_len);
    if (result == NULL) return NULL;
    
    char *dst = result;
    const char *src = str;
    int in_single_quote = 0;
    
    while (*src) {
        if (*src == '\'') {
            in_single_quote = !in_single_quote;
            src++;
        } else if (*src == '$' && !in_single_quote) {
            src++;
            char var_name[256];
            int var_len = 0;
            
            if (*src == '{') {
                src++;
                while (*src && *src != '}' && var_len < 255) {
                    var_name[var_len++] = *src++;
                }
                if (*src == '}') src++;
            } else {
                while (*src && (isalnum((unsigned char)*src) || *src == '_') && var_len < 255) {
                    var_name[var_len++] = *src++;
                }
            }
            
            var_name[var_len] = '\0';
            
            if (var_len > 0) {
                char *value = getenv(var_name);
                if (value != NULL) {
                    size_t value_len = strlen(value);
                    if ((dst - result) + value_len >= result_len) {
                        size_t pos = dst - result;
                        result_len = result_len * 2 + value_len;
                        result = realloc(result, result_len);
                        if (result == NULL) return NULL;
                        dst = result + pos;
                    }
                    strcpy(dst, value);
                    dst += value_len;
                }
            } else {
                if ((dst - result) + 1 >= result_len) {
                    size_t pos = dst - result;
                    result_len = result_len * 2 + 2;
                    result = realloc(result, result_len);
                    if (result == NULL) return NULL;
                    dst = result + pos;
                }
                *dst++ = '$';
            }
        } else {
            if ((dst - result) + 1 >= result_len) {
                size_t pos = dst - result;
                result_len = result_len * 2 + 2;
                result = realloc(result, result_len);
                if (result == NULL) return NULL;
                dst = result + pos;
            }
            *dst++ = *src++;
        }
    }
    *dst = '\0';
    
    return result;
}

static void init_command(Command *cmd) {
    memset(cmd->argv, 0, sizeof(cmd->argv));
    cmd->redirect_file = NULL;
    cmd->redirect_append = 0;
    cmd->background = 0;
}

BuiltinType check_builtin(const Command *cmd) {
    if (cmd->argv[0] == NULL) return BUILTIN_NONE;
    
    if (strcmp(cmd->argv[0], "cd") == 0) return BUILTIN_CD;
    if (strcmp(cmd->argv[0], "exit") == 0) return BUILTIN_EXIT;
    if (strcmp(cmd->argv[0], "export") == 0) return BUILTIN_EXPORT;
    
    return BUILTIN_NONE;
}

static int parse_command_segment(char *segment, Command *cmd) {
    init_command(cmd);
    segment = trim(segment);
    
    if (*segment == '\0') return 0;
    
    char *ptr = segment;
    char *token;
    int argc = 0;
    
    while ((token = get_token(&ptr)) != NULL) {
        if (strcmp(token, ">") == 0 || strcmp(token, ">>") == 0) {
            int append = (strcmp(token, ">>") == 0);
            char *file = get_token(&ptr);
            if (file != NULL && *file != '\0') {
                char *unquoted = remove_quotes(file);
                char *expanded = expand_env_vars(unquoted);
                free(unquoted);
                cmd->redirect_file = expanded;
                cmd->redirect_append = append;
            }
        } else if (strcmp(token, "&") == 0) {
            if (*ptr == '\0') {
                cmd->background = 1;
            } else {
                if (argc < MAX_ARGS - 1) {
                    char *expanded = expand_env_vars(token);
                    cmd->argv[argc++] = expanded;
                }
            }
        } else {
            if (argc < MAX_ARGS - 1) {
                char *unquoted = remove_quotes(token);
                char *expanded = expand_env_vars(unquoted);
                free(unquoted);
                cmd->argv[argc++] = expanded;
            }
        }
    }
    
    cmd->argv[argc] = NULL;
    return argc > 0 ? 1 : 0;
}

int parse_input(const char *input, Pipeline *pipeline) {
    if (input == NULL || pipeline == NULL) return 0;
    
    pipeline->num_commands = 0;
    for (int i = 0; i < MAX_PIPES; i++) {
        init_command(&pipeline->commands[i]);
    }
    
    char *input_copy = strdup(input);
    if (input_copy == NULL) return 0;
    
    char *trimmed = trim(input_copy);
    if (*trimmed == '\0') {
        free(input_copy);
        return 0;
    }
    
    char *segments[MAX_PIPES];
    int num_segments = 0;
    char *saveptr = NULL;
    
    char *token = strtok_r(trimmed, "|", &saveptr);
    while (token != NULL && num_segments < MAX_PIPES) {
        segments[num_segments++] = token;
        token = strtok_r(NULL, "|", &saveptr);
    }
    
    for (int i = 0; i < num_segments; i++) {
        if (parse_command_segment(segments[i], &pipeline->commands[i])) {
            pipeline->num_commands++;
        }
    }
    
    if (pipeline->num_commands > 0) {
        Command *last_cmd = &pipeline->commands[pipeline->num_commands - 1];
        char *segment = segments[num_segments - 1];
        size_t len = strlen(segment);
        while (len > 0 && isspace((unsigned char)segment[len - 1])) len--;
        if (len > 0 && segment[len - 1] == '&') {
            last_cmd->background = 1;
        }
    }
    
    free(input_copy);
    return pipeline->num_commands > 0 ? 1 : 0;
}

void free_pipeline(Pipeline *pipeline) {
    if (pipeline == NULL) return;
    
    for (int i = 0; i < pipeline->num_commands; i++) {
        Command *cmd = &pipeline->commands[i];
        for (int j = 0; cmd->argv[j] != NULL; j++) {
            free(cmd->argv[j]);
            cmd->argv[j] = NULL;
        }
        if (cmd->redirect_file != NULL) {
            free(cmd->redirect_file);
            cmd->redirect_file = NULL;
        }
    }
    pipeline->num_commands = 0;
}
