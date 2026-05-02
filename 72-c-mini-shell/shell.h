#ifndef SHELL_H
#define SHELL_H

#include <sys/types.h>

#define MAX_LINE 4096
#define MAX_ARGS 256
#define MAX_PIPES 16

typedef struct {
    char *argv[MAX_ARGS];
    char *redirect_file;
    int redirect_append;
    int background;
} Command;

typedef struct {
    Command commands[MAX_PIPES];
    int num_commands;
} Pipeline;

typedef enum {
    BUILTIN_NONE,
    BUILTIN_CD,
    BUILTIN_EXIT,
    BUILTIN_EXPORT
} BuiltinType;

int parse_input(const char *input, Pipeline *pipeline);
BuiltinType check_builtin(const Command *cmd);
int execute_builtin(const Command *cmd);
int execute_pipeline(Pipeline *pipeline);
void free_pipeline(Pipeline *pipeline);
char *expand_env_vars(const char *str);

extern int last_exit_status;

#endif
