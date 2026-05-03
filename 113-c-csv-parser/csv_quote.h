#ifndef CSV_QUOTE_H
#define CSV_QUOTE_H

#include <stddef.h>

typedef enum {
    CSV_QUOTE_STATE_NORMAL,
    CSV_QUOTE_STATE_IN_QUOTED,
    CSV_QUOTE_STATE_EXPECT_DELIM
} CSV_QuoteState;

typedef struct CSV_QuoteHandler {
    CSV_QuoteState state;
    char *field_buffer;
    size_t field_size;
    size_t field_used;
    int last_was_quote;
} CSV_QuoteHandler;

void CSV_QuoteHandlerInit(CSV_QuoteHandler *handler);
void CSV_QuoteHandlerReset(CSV_QuoteHandler *handler);
void CSV_QuoteHandlerDestroy(CSV_QuoteHandler *handler);

void CSV_QuoteHandlerStartField(CSV_QuoteHandler *handler);
void CSV_QuoteHandlerAppendChar(CSV_QuoteHandler *handler, char c);
void CSV_QuoteHandlerEndField(CSV_QuoteHandler *handler);

char* CSV_QuoteHandlerGetFieldValue(CSV_QuoteHandler *handler);
size_t CSV_QuoteHandlerGetFieldLength(CSV_QuoteHandler *handler);

int CSV_QuoteHandlerProcessChar(CSV_QuoteHandler *handler, char c, char delimiter,
                                  int *field_complete, int *row_complete);

#endif
