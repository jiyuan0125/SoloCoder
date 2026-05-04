#ifndef RSS_H
#define RSS_H

#include <stdio.h>
#include <time.h>

#define RSS_VERSION "2.0"
#define RSS_MAX_TITLE_LEN 256
#define RSS_MAX_LINK_LEN 512
#define RSS_MAX_DESCRIPTION_LEN 4096
#define RSS_MAX_AUTHOR_LEN 128
#define RSS_MAX_CATEGORY_LEN 128
#define RSS_MAX_LANGUAGE_LEN 32

typedef struct {
    char title[RSS_MAX_TITLE_LEN];
    char link[RSS_MAX_LINK_LEN];
    char description[RSS_MAX_DESCRIPTION_LEN];
    char language[RSS_MAX_LANGUAGE_LEN];
    time_t last_build_date;
    time_t pub_date;
} RSS_Channel;

typedef struct {
    char title[RSS_MAX_TITLE_LEN];
    char link[RSS_MAX_LINK_LEN];
    char description[RSS_MAX_DESCRIPTION_LEN];
    char author[RSS_MAX_AUTHOR_LEN];
    char categories[RSS_MAX_CATEGORY_LEN][RSS_MAX_CATEGORY_LEN];
    int category_count;
    time_t pub_date;
    int is_html_description;
} RSS_Item;

typedef enum {
    RSS_OUTPUT_FILE,
    RSS_OUTPUT_BUFFER
} RSS_OutputType;

typedef struct {
    RSS_OutputType type;
    union {
        FILE *file;
        struct {
            char *buffer;
            size_t size;
            size_t used;
        } buf;
    } data;
    int indent_level;
    int max_items;
    int item_count;
    int started;
    int closed;
} RSS_Writer;

RSS_Writer* RSS_Writer_ToFile(const char *filepath, int max_items);
RSS_Writer* RSS_Writer_ToBuffer(char *buffer, size_t size, int max_items);
void RSS_Writer_Free(RSS_Writer *writer);

int RSS_StartChannel(RSS_Writer *writer, const RSS_Channel *channel);
int RSS_AddItem(RSS_Writer *writer, const RSS_Item *item);
int RSS_EndChannel(RSS_Writer *writer);

int RSS_GenerateToFile(const char *filepath, const RSS_Channel *channel, 
                        const RSS_Item *items, int item_count, int max_items);
int RSS_GenerateToBuffer(char *buffer, size_t size, const RSS_Channel *channel,
                          const RSS_Item *items, int item_count, int max_items);

#endif
