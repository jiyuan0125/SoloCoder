#include "csv_quote.h"
#include <stdlib.h>
#include <string.h>

#define INITIAL_FIELD_BUFFER_SIZE 64

void CSV_QuoteHandlerInit(CSV_QuoteHandler *handler) {
    if (!handler) return;
    handler->state = CSV_QUOTE_STATE_NORMAL;
    handler->field_buffer = NULL;
    handler->field_size = 0;
    handler->field_used = 0;
    handler->last_was_quote = 0;
}

void CSV_QuoteHandlerReset(CSV_QuoteHandler *handler) {
    if (!handler) return;
    handler->state = CSV_QUOTE_STATE_NORMAL;
    handler->field_used = 0;
    handler->last_was_quote = 0;
}

void CSV_QuoteHandlerDestroy(CSV_QuoteHandler *handler) {
    if (!handler) return;
    if (handler->field_buffer) {
        free(handler->field_buffer);
        handler->field_buffer = NULL;
    }
    handler->field_size = 0;
    handler->field_used = 0;
}

void CSV_QuoteHandlerStartField(CSV_QuoteHandler *handler) {
    if (!handler) return;
    handler->field_used = 0;
    handler->last_was_quote = 0;
}

static int CSV_QuoteHandlerEnsureCapacity(CSV_QuoteHandler *handler, size_t needed) {
    if (!handler) return 0;

    if (handler->field_size == 0) {
        size_t initial_size = INITIAL_FIELD_BUFFER_SIZE;
        while (initial_size < needed + 1) {
            initial_size *= 2;
        }
        handler->field_buffer = (char*)malloc(initial_size);
        if (!handler->field_buffer) return 0;
        handler->field_size = initial_size;
        return 1;
    }

    if (handler->field_used + needed < handler->field_size) {
        return 1;
    }

    size_t new_size = handler->field_size;
    while (new_size <= handler->field_used + needed) {
        new_size *= 2;
    }

    char *new_buffer = (char*)realloc(handler->field_buffer, new_size);
    if (!new_buffer) return 0;

    handler->field_buffer = new_buffer;
    handler->field_size = new_size;
    return 1;
}

void CSV_QuoteHandlerAppendChar(CSV_QuoteHandler *handler, char c) {
    if (!handler) return;
    if (!CSV_QuoteHandlerEnsureCapacity(handler, 1)) return;
    handler->field_buffer[handler->field_used++] = c;
}

void CSV_QuoteHandlerEndField(CSV_QuoteHandler *handler) {
    (void)handler;
}

char* CSV_QuoteHandlerGetFieldValue(CSV_QuoteHandler *handler) {
    if (!handler) return NULL;
    return handler->field_buffer;
}

size_t CSV_QuoteHandlerGetFieldLength(CSV_QuoteHandler *handler) {
    if (!handler) return 0;
    return handler->field_used;
}

int CSV_QuoteHandlerProcessChar(CSV_QuoteHandler *handler, char c, char delimiter,
                                  int *field_complete, int *row_complete) {
    if (!handler || !field_complete || !row_complete) return 0;

    *field_complete = 0;
    *row_complete = 0;

    switch (handler->state) {
        case CSV_QUOTE_STATE_NORMAL:
            if (handler->field_used == 0 && !handler->last_was_quote && c == '"') {
                handler->state = CSV_QUOTE_STATE_IN_QUOTED;
                handler->last_was_quote = 0;
                return 1;
            }

            if (c == delimiter) {
                *field_complete = 1;
                return 1;
            }

            if (c == '\n' || c == '\r') {
                *row_complete = 1;
                return 1;
            }

            CSV_QuoteHandlerAppendChar(handler, c);
            handler->last_was_quote = 0;
            return 1;

        case CSV_QUOTE_STATE_IN_QUOTED:
            if (c == '"') {
                if (handler->last_was_quote) {
                    CSV_QuoteHandlerAppendChar(handler, '"');
                    handler->last_was_quote = 0;
                } else {
                    handler->last_was_quote = 1;
                }
                return 1;
            }

            if (handler->last_was_quote) {
                if (c == delimiter) {
                    *field_complete = 1;
                    handler->state = CSV_QUOTE_STATE_NORMAL;
                    handler->last_was_quote = 0;
                    return 1;
                }
                if (c == '\n' || c == '\r') {
                    *row_complete = 1;
                    handler->state = CSV_QUOTE_STATE_NORMAL;
                    handler->last_was_quote = 0;
                    return 1;
                }
                handler->last_was_quote = 0;
            }

            CSV_QuoteHandlerAppendChar(handler, c);
            return 1;

        case CSV_QUOTE_STATE_EXPECT_DELIM:
            if (c == delimiter) {
                *field_complete = 1;
                handler->state = CSV_QUOTE_STATE_NORMAL;
                return 1;
            }
            if (c == '\n' || c == '\r') {
                *row_complete = 1;
                handler->state = CSV_QUOTE_STATE_NORMAL;
                return 1;
            }
            CSV_QuoteHandlerAppendChar(handler, c);
            return 1;

        default:
            return 0;
    }
}
