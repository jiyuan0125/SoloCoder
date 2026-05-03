#ifndef CSV_PARSER_H
#define CSV_PARSER_H

#include <stddef.h>
#include "csv_quote.h"

typedef enum {
    CSV_DELIMITER_COMMA,
    CSV_DELIMITER_TAB
} CSV_Delimiter;

typedef enum {
    CSV_EMPTY_LINE_SKIP,
    CSV_EMPTY_LINE_KEEP
} CSV_EmptyLinePolicy;

typedef enum {
    CSV_PARSE_OK,
    CSV_PARSE_NEED_MORE_DATA,
    CSV_PARSE_ERROR
} CSV_ParseStatus;

typedef struct CSV_Row {
    char **fields;
    size_t field_count;
    size_t field_capacity;
    int is_empty;
} CSV_Row;

typedef void (*CSV_RowCallback)(const CSV_Row *row, void *user_data);

typedef struct CSV_Parser {
    char *buffer;
    size_t buffer_size;
    size_t buffer_used;
    size_t buffer_pos;

    CSV_Delimiter delimiter;
    CSV_EmptyLinePolicy empty_line_policy;

    char delimiter_char;

    CSV_RowCallback row_callback;
    void *user_data;

    int error_code;
    const char *error_message;

    CSV_QuoteHandler quote_handler;
    CSV_Row current_row;
} CSV_Parser;

CSV_Parser* CSV_ParserCreate(void);
void CSV_ParserDestroy(CSV_Parser *parser);

void CSV_ParserSetDelimiter(CSV_Parser *parser, CSV_Delimiter delimiter);
void CSV_ParserSetEmptyLinePolicy(CSV_Parser *parser, CSV_EmptyLinePolicy policy);
void CSV_ParserSetRowCallback(CSV_Parser *parser, CSV_RowCallback callback, void *user_data);

CSV_ParseStatus CSV_ParserFeed(CSV_Parser *parser, const char *data, size_t size);
CSV_ParseStatus CSV_ParserFinish(CSV_Parser *parser);

int CSV_ParserGetErrorCode(const CSV_Parser *parser);
const char* CSV_ParserGetErrorMessage(const CSV_Parser *parser);

void CSV_RowInit(CSV_Row *row);
void CSV_RowFreeFields(CSV_Row *row);
char* CSV_RowGetField(const CSV_Row *row, size_t index);
size_t CSV_RowGetFieldCount(const CSV_Row *row);

#endif
