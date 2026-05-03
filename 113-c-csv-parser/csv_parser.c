#include "csv_parser.h"
#include "csv_quote.h"
#include <stdlib.h>
#include <string.h>

#define INITIAL_BUFFER_SIZE 4096
#define INITIAL_FIELD_CAPACITY 16

static int CSV_ParserResizeBuffer(CSV_Parser *parser, size_t needed);
static CSV_ParseStatus CSV_ParserProcessBuffer(CSV_Parser *parser);
static int CSV_IsLineTerminator(char c);
static int CSV_IsRowEmpty(const CSV_Row *row);

CSV_Parser* CSV_ParserCreate(void) {
    CSV_Parser *parser = (CSV_Parser*)calloc(1, sizeof(CSV_Parser));
    if (!parser) return NULL;

    parser->buffer = (char*)malloc(INITIAL_BUFFER_SIZE);
    if (!parser->buffer) {
        free(parser);
        return NULL;
    }

    parser->buffer_size = INITIAL_BUFFER_SIZE;
    parser->delimiter = CSV_DELIMITER_COMMA;
    parser->delimiter_char = ',';
    parser->empty_line_policy = CSV_EMPTY_LINE_SKIP;

    CSV_QuoteHandlerInit(&parser->quote_handler);
    CSV_RowInit(&parser->current_row);

    return parser;
}

void CSV_ParserDestroy(CSV_Parser *parser) {
    if (!parser) return;
    if (parser->buffer) {
        free(parser->buffer);
    }
    CSV_QuoteHandlerDestroy(&parser->quote_handler);
    CSV_RowFreeFields(&parser->current_row);
    free(parser);
}

void CSV_ParserSetDelimiter(CSV_Parser *parser, CSV_Delimiter delimiter) {
    if (!parser) return;
    parser->delimiter = delimiter;
    if (delimiter == CSV_DELIMITER_COMMA) {
        parser->delimiter_char = ',';
    } else {
        parser->delimiter_char = '\t';
    }
}

void CSV_ParserSetEmptyLinePolicy(CSV_Parser *parser, CSV_EmptyLinePolicy policy) {
    if (!parser) return;
    parser->empty_line_policy = policy;
}

void CSV_ParserSetRowCallback(CSV_Parser *parser, CSV_RowCallback callback, void *user_data) {
    if (!parser) return;
    parser->row_callback = callback;
    parser->user_data = user_data;
}

CSV_ParseStatus CSV_ParserFeed(CSV_Parser *parser, const char *data, size_t size) {
    if (!parser || !data || size == 0) {
        return CSV_PARSE_OK;
    }

    size_t available = parser->buffer_size - parser->buffer_used;
    if (available < size) {
        size_t needed = parser->buffer_used + size;
        if (!CSV_ParserResizeBuffer(parser, needed)) {
            parser->error_code = 1;
            parser->error_message = "Memory allocation failed";
            return CSV_PARSE_ERROR;
        }
    }

    memcpy(parser->buffer + parser->buffer_used, data, size);
    parser->buffer_used += size;

    return CSV_ParserProcessBuffer(parser);
}

CSV_ParseStatus CSV_ParserFinish(CSV_Parser *parser) {
    if (!parser) return CSV_PARSE_OK;

    int has_pending_data = (parser->buffer_used > parser->buffer_pos) ||
                           (parser->quote_handler.field_used > 0) ||
                           (parser->current_row.field_count > 0);

    if (has_pending_data) {
        if (parser->buffer_used + 1 >= parser->buffer_size) {
            if (!CSV_ParserResizeBuffer(parser, parser->buffer_size + 2)) {
                parser->error_code = 1;
                parser->error_message = "Memory allocation failed";
                return CSV_PARSE_ERROR;
            }
        }
        parser->buffer[parser->buffer_used++] = '\n';
        return CSV_ParserProcessBuffer(parser);
    }

    return CSV_PARSE_OK;
}

int CSV_ParserGetErrorCode(const CSV_Parser *parser) {
    return parser ? parser->error_code : -1;
}

const char* CSV_ParserGetErrorMessage(const CSV_Parser *parser) {
    return parser ? parser->error_message : "Invalid parser";
}

void CSV_RowInit(CSV_Row *row) {
    if (!row) return;
    row->fields = NULL;
    row->field_count = 0;
    row->field_capacity = 0;
    row->is_empty = 0;
}

void CSV_RowFreeFields(CSV_Row *row) {
    if (!row || !row->fields) return;
    for (size_t i = 0; i < row->field_count; i++) {
        if (row->fields[i]) {
            free(row->fields[i]);
        }
    }
    free(row->fields);
    row->fields = NULL;
    row->field_count = 0;
    row->field_capacity = 0;
}

char* CSV_RowGetField(const CSV_Row *row, size_t index) {
    if (!row || index >= row->field_count) return NULL;
    return row->fields[index];
}

size_t CSV_RowGetFieldCount(const CSV_Row *row) {
    return row ? row->field_count : 0;
}

static int CSV_ParserResizeBuffer(CSV_Parser *parser, size_t needed) {
    size_t new_size = parser->buffer_size;
    while (new_size < needed) {
        new_size *= 2;
    }

    char *new_buffer = (char*)realloc(parser->buffer, new_size);
    if (!new_buffer) return 0;

    parser->buffer = new_buffer;
    parser->buffer_size = new_size;
    return 1;
}

static int CSV_IsLineTerminator(char c) {
    return (c == '\n' || c == '\r');
}

static int CSV_IsRowEmpty(const CSV_Row *row) {
    if (row->field_count == 0) return 1;
    if (row->field_count == 1) {
        const char *field = row->fields[0];
        if (field == NULL || field[0] == '\0') return 1;
    }
    return 0;
}

static int CSV_RowAddField(CSV_Row *row, CSV_QuoteHandler *quote_handler) {
    if (row->field_count >= row->field_capacity) {
        size_t new_capacity = (row->field_capacity == 0) ? INITIAL_FIELD_CAPACITY : row->field_capacity * 2;
        char **new_fields = (char**)realloc(row->fields, new_capacity * sizeof(char*));
        if (!new_fields) return 0;
        row->fields = new_fields;
        row->field_capacity = new_capacity;
    }

    char *value = CSV_QuoteHandlerGetFieldValue(quote_handler);
    size_t len = CSV_QuoteHandlerGetFieldLength(quote_handler);

    char *field_copy = (char*)malloc(len + 1);
    if (!field_copy) return 0;

    if (value) {
        memcpy(field_copy, value, len);
    }
    field_copy[len] = '\0';

    row->fields[row->field_count++] = field_copy;
    return 1;
}

static CSV_ParseStatus CSV_ParserProcessBuffer(CSV_Parser *parser) {
    size_t i = parser->buffer_pos;
    int skip_next = 0;

    while (i < parser->buffer_used) {
        char c = parser->buffer[i];

        if (skip_next) {
            skip_next = 0;
            i++;
            continue;
        }

        int field_complete = 0;
        int row_complete = 0;

        if (!CSV_QuoteHandlerProcessChar(&parser->quote_handler, c, parser->delimiter_char,
                                           &field_complete, &row_complete)) {
            parser->error_code = 2;
            parser->error_message = "Quote processing failed";
            return CSV_PARSE_ERROR;
        }

        if (CSV_IsLineTerminator(c) && parser->quote_handler.state == CSV_QUOTE_STATE_NORMAL) {
            row_complete = 1;
            if (c == '\r' && i + 1 < parser->buffer_used && parser->buffer[i + 1] == '\n') {
                skip_next = 1;
            }
        }

        if (field_complete) {
            if (!CSV_RowAddField(&parser->current_row, &parser->quote_handler)) {
                parser->error_code = 1;
                parser->error_message = "Memory allocation failed";
                return CSV_PARSE_ERROR;
            }
            CSV_QuoteHandlerReset(&parser->quote_handler);
        }

        if (row_complete) {
            if (parser->quote_handler.field_used > 0 || parser->current_row.field_count > 0) {
                if (!CSV_RowAddField(&parser->current_row, &parser->quote_handler)) {
                    parser->error_code = 1;
                    parser->error_message = "Memory allocation failed";
                    return CSV_PARSE_ERROR;
                }
                CSV_QuoteHandlerReset(&parser->quote_handler);
            }

            parser->current_row.is_empty = CSV_IsRowEmpty(&parser->current_row);

            int should_emit = 1;
            if (parser->current_row.is_empty && parser->empty_line_policy == CSV_EMPTY_LINE_SKIP) {
                should_emit = 0;
            }

            if (should_emit && parser->row_callback) {
                parser->row_callback(&parser->current_row, parser->user_data);
            }

            CSV_RowFreeFields(&parser->current_row);
            CSV_RowInit(&parser->current_row);
        }

        i++;
    }

    parser->buffer_pos = i;

    if (parser->buffer_pos > 0 && parser->buffer_used > parser->buffer_pos) {
        memmove(parser->buffer, parser->buffer + parser->buffer_pos,
                parser->buffer_used - parser->buffer_pos);
        parser->buffer_used -= parser->buffer_pos;
        parser->buffer_pos = 0;
    } else if (parser->buffer_used == parser->buffer_pos) {
        parser->buffer_used = 0;
        parser->buffer_pos = 0;
    }

    return CSV_PARSE_OK;
}
