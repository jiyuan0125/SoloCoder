#ifndef LINEEDIT_H
#define LINEEDIT_H

#include "history.h"
#include <stddef.h>

typedef struct {
    char *buf;
    size_t buf_size;
    size_t len;
    size_t cursor;
    size_t prompt_len;
    const char *prompt;
    History *history;
    char *saved_line;
    int interrupted;
} LineEdit;

typedef char **(*CompleteCallback)(const char *line, int *count);

void lineedit_init(LineEdit *le, History *h);
void lineedit_free(LineEdit *le);
void lineedit_set_prompt(LineEdit *le, const char *prompt);
char *lineedit_read(LineEdit *le);
void lineedit_set_complete_callback(CompleteCallback cb);

#endif
