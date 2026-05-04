#include "rss.h"
#include "xml_builder.h"
#include "xml_escape.h"
#include "time_utils.h"
#include <stdlib.h>
#include <string.h>

#define DEFAULT_MAX_ITEMS 20

RSS_Writer* RSS_Writer_ToFile(const char *filepath, int max_items) {
    if (filepath == NULL) return NULL;
    
    RSS_Writer *writer = (RSS_Writer *)malloc(sizeof(RSS_Writer));
    if (writer == NULL) return NULL;
    
    writer->type = RSS_OUTPUT_FILE;
    writer->data.file = fopen(filepath, "w");
    if (writer->data.file == NULL) {
        free(writer);
        return NULL;
    }
    
    writer->indent_level = 0;
    writer->max_items = (max_items > 0) ? max_items : DEFAULT_MAX_ITEMS;
    writer->item_count = 0;
    writer->started = 0;
    writer->closed = 0;
    
    return writer;
}

RSS_Writer* RSS_Writer_ToBuffer(char *buffer, size_t size, int max_items) {
    if (buffer == NULL || size == 0) return NULL;
    
    RSS_Writer *writer = (RSS_Writer *)malloc(sizeof(RSS_Writer));
    if (writer == NULL) return NULL;
    
    writer->type = RSS_OUTPUT_BUFFER;
    writer->data.buf.buffer = buffer;
    writer->data.buf.size = size;
    writer->data.buf.used = 0;
    
    buffer[0] = '\0';
    
    writer->indent_level = 0;
    writer->max_items = (max_items > 0) ? max_items : DEFAULT_MAX_ITEMS;
    writer->item_count = 0;
    writer->started = 0;
    writer->closed = 0;
    
    return writer;
}

void RSS_Writer_Free(RSS_Writer *writer) {
    if (writer == NULL) return;
    
    if (!writer->closed) {
        RSS_EndChannel(writer);
    }
    
    if (writer->type == RSS_OUTPUT_FILE && writer->data.file != NULL) {
        fclose(writer->data.file);
    }
    
    free(writer);
}

static int RSS_WriteDeclarationToWriter(RSS_Writer *writer) {
    const char *decl = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n";
    if (writer->type == RSS_OUTPUT_FILE) {
        return fputs(decl, writer->data.file) >= 0 ? 0 : -1;
    } else {
        size_t len = strlen(decl);
        if (writer->data.buf.used + len >= writer->data.buf.size) {
            return -1;
        }
        memcpy(writer->data.buf.buffer + writer->data.buf.used, decl, len);
        writer->data.buf.used += len;
        return 0;
    }
}

static int RSS_WriteIndent(RSS_Writer *writer) {
    int spaces = writer->indent_level * 2;
    for (int i = 0; i < spaces; i++) {
        if (writer->type == RSS_OUTPUT_FILE) {
            fputc(' ', writer->data.file);
        } else {
            if (writer->data.buf.used < writer->data.buf.size - 1) {
                writer->data.buf.buffer[writer->data.buf.used++] = ' ';
            }
        }
    }
    return 0;
}

static int RSS_WriteRaw(RSS_Writer *writer, const char *str) {
    if (writer->type == RSS_OUTPUT_FILE) {
        return fputs(str, writer->data.file) >= 0 ? 0 : -1;
    } else {
        size_t len = strlen(str);
        size_t available = writer->data.buf.size - writer->data.buf.used - 1;
        size_t to_write = (len < available) ? len : available;
        
        memcpy(writer->data.buf.buffer + writer->data.buf.used, str, to_write);
        writer->data.buf.used += to_write;
        writer->data.buf.buffer[writer->data.buf.used] = '\0';
        
        return 0;
    }
}

static int RSS_WriteEscaped(RSS_Writer *writer, const char *text) {
    size_t escaped_len = XML_Escape_Length(text);
    char *escaped = (char *)malloc(escaped_len + 1);
    if (escaped == NULL) return -1;
    
    XML_Escape(escaped, escaped_len + 1, text);
    int result = RSS_WriteRaw(writer, escaped);
    free(escaped);
    
    return result;
}

static int RSS_WriteCData(RSS_Writer *writer, const char *content) {
    if (content == NULL) {
        return RSS_WriteRaw(writer, "");
    }
    
    char fixed_input[8192];
    XML_ValidateAndFixUTF8(fixed_input, sizeof(fixed_input), content);
    
    const char *start = fixed_input;
    const char *p = fixed_input;
    
    while (*p != '\0') {
        if (*p == ']' && *(p + 1) == ']' && *(p + 2) == '>') {
            size_t copy_len = p - start + 2;
            char *temp = (char *)malloc(copy_len + 1);
            if (temp == NULL) return -1;
            strncpy(temp, start, copy_len);
            temp[copy_len] = '\0';
            RSS_WriteRaw(writer, temp);
            free(temp);
            RSS_WriteRaw(writer, "]]><![CDATA[");
            RSS_WriteRaw(writer, ">");
            p += 3;
            start = p;
        } else {
            p++;
        }
    }
    
    if (start < p) {
        RSS_WriteRaw(writer, start);
    }
    
    return 0;
}

int RSS_StartChannel(RSS_Writer *writer, const RSS_Channel *channel) {
    if (writer == NULL || channel == NULL || writer->started) {
        return -1;
    }
    
    RSS_WriteDeclarationToWriter(writer);
    
    RSS_WriteRaw(writer, "<rss version=\"");
    RSS_WriteRaw(writer, RSS_VERSION);
    RSS_WriteRaw(writer, "\">\n");
    writer->indent_level++;
    
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "<channel>\n");
    writer->indent_level++;
    
    char time_buf[RFC822_DATE_LEN];
    
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "<title>");
    RSS_WriteEscaped(writer, channel->title);
    RSS_WriteRaw(writer, "</title>\n");
    
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "<link>");
    RSS_WriteEscaped(writer, channel->link);
    RSS_WriteRaw(writer, "</link>\n");
    
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "<description>");
    RSS_WriteEscaped(writer, channel->description);
    RSS_WriteRaw(writer, "</description>\n");
    
    if (channel->language[0] != '\0') {
        RSS_WriteIndent(writer);
        RSS_WriteRaw(writer, "<language>");
        RSS_WriteEscaped(writer, channel->language);
        RSS_WriteRaw(writer, "</language>\n");
    }
    
    if (channel->last_build_date > 0) {
        Time_ToRFC822(time_buf, sizeof(time_buf), channel->last_build_date);
        RSS_WriteIndent(writer);
        RSS_WriteRaw(writer, "<lastBuildDate>");
        RSS_WriteRaw(writer, time_buf);
        RSS_WriteRaw(writer, "</lastBuildDate>\n");
    }
    
    if (channel->pub_date > 0) {
        Time_ToRFC822(time_buf, sizeof(time_buf), channel->pub_date);
        RSS_WriteIndent(writer);
        RSS_WriteRaw(writer, "<pubDate>");
        RSS_WriteRaw(writer, time_buf);
        RSS_WriteRaw(writer, "</pubDate>\n");
    }
    
    writer->started = 1;
    
    if (writer->type == RSS_OUTPUT_FILE) {
        fflush(writer->data.file);
    }
    
    return 0;
}

int RSS_AddItem(RSS_Writer *writer, const RSS_Item *item) {
    if (writer == NULL || item == NULL || !writer->started || writer->closed) {
        return -1;
    }
    
    if (writer->item_count >= writer->max_items) {
        return 0;
    }
    
    char time_buf[RFC822_DATE_LEN];
    
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "<item>\n");
    writer->indent_level++;
    
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "<title>");
    RSS_WriteEscaped(writer, item->title);
    RSS_WriteRaw(writer, "</title>\n");
    
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "<link>");
    RSS_WriteEscaped(writer, item->link);
    RSS_WriteRaw(writer, "</link>\n");
    
    if (item->is_html_description) {
        RSS_WriteIndent(writer);
        RSS_WriteRaw(writer, "<description><![CDATA[");
        RSS_WriteCData(writer, item->description);
        RSS_WriteRaw(writer, "]]></description>\n");
    } else {
        RSS_WriteIndent(writer);
        RSS_WriteRaw(writer, "<description>");
        RSS_WriteEscaped(writer, item->description);
        RSS_WriteRaw(writer, "</description>\n");
    }
    
    if (item->pub_date > 0) {
        Time_ToRFC822(time_buf, sizeof(time_buf), item->pub_date);
        RSS_WriteIndent(writer);
        RSS_WriteRaw(writer, "<pubDate>");
        RSS_WriteRaw(writer, time_buf);
        RSS_WriteRaw(writer, "</pubDate>\n");
    }
    
    if (item->author[0] != '\0') {
        RSS_WriteIndent(writer);
        RSS_WriteRaw(writer, "<author>");
        RSS_WriteEscaped(writer, item->author);
        RSS_WriteRaw(writer, "</author>\n");
    }
    
    for (int i = 0; i < item->category_count; i++) {
        if (item->categories[i][0] != '\0') {
            RSS_WriteIndent(writer);
            RSS_WriteRaw(writer, "<category>");
            RSS_WriteEscaped(writer, item->categories[i]);
            RSS_WriteRaw(writer, "</category>\n");
        }
    }
    
    writer->indent_level--;
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "</item>\n");
    
    writer->item_count++;
    
    if (writer->type == RSS_OUTPUT_FILE) {
        fflush(writer->data.file);
    }
    
    return 0;
}

int RSS_EndChannel(RSS_Writer *writer) {
    if (writer == NULL || writer->closed) {
        return -1;
    }
    
    if (!writer->started) {
        writer->closed = 1;
        return 0;
    }
    
    writer->indent_level--;
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "</channel>\n");
    
    writer->indent_level--;
    RSS_WriteIndent(writer);
    RSS_WriteRaw(writer, "</rss>\n");
    
    writer->closed = 1;
    
    if (writer->type == RSS_OUTPUT_FILE) {
        fflush(writer->data.file);
    }
    
    return 0;
}

static int RSS_CompareItemsByDate(const void *a, const void *b) {
    const RSS_Item *item_a = (const RSS_Item *)a;
    const RSS_Item *item_b = (const RSS_Item *)b;
    
    if (item_a->pub_date < item_b->pub_date) return 1;
    if (item_a->pub_date > item_b->pub_date) return -1;
    return 0;
}

int RSS_GenerateToFile(const char *filepath, const RSS_Channel *channel, 
                        const RSS_Item *items, int item_count, int max_items) {
    if (filepath == NULL || channel == NULL) {
        return -1;
    }
    
    RSS_Writer *writer = RSS_Writer_ToFile(filepath, max_items);
    if (writer == NULL) {
        return -1;
    }
    
    if (RSS_StartChannel(writer, channel) != 0) {
        RSS_Writer_Free(writer);
        return -1;
    }
    
    RSS_Item *sorted_items = NULL;
    const RSS_Item *items_to_use = items;
    
    if (items != NULL && item_count > 0) {
        sorted_items = (RSS_Item *)malloc(item_count * sizeof(RSS_Item));
        if (sorted_items == NULL) {
            RSS_Writer_Free(writer);
            return -1;
        }
        memcpy(sorted_items, items, item_count * sizeof(RSS_Item));
        qsort(sorted_items, item_count, sizeof(RSS_Item), RSS_CompareItemsByDate);
        items_to_use = sorted_items;
    }
    
    for (int i = 0; i < item_count && writer->item_count < writer->max_items; i++) {
        if (RSS_AddItem(writer, &items_to_use[i]) != 0) {
            free(sorted_items);
            RSS_Writer_Free(writer);
            return -1;
        }
    }
    
    free(sorted_items);
    
    if (RSS_EndChannel(writer) != 0) {
        RSS_Writer_Free(writer);
        return -1;
    }
    
    RSS_Writer_Free(writer);
    return 0;
}

int RSS_GenerateToBuffer(char *buffer, size_t size, const RSS_Channel *channel,
                          const RSS_Item *items, int item_count, int max_items) {
    if (buffer == NULL || size == 0 || channel == NULL) {
        return -1;
    }
    
    RSS_Writer *writer = RSS_Writer_ToBuffer(buffer, size, max_items);
    if (writer == NULL) {
        return -1;
    }
    
    if (RSS_StartChannel(writer, channel) != 0) {
        RSS_Writer_Free(writer);
        return -1;
    }
    
    RSS_Item *sorted_items = NULL;
    const RSS_Item *items_to_use = items;
    
    if (items != NULL && item_count > 0) {
        sorted_items = (RSS_Item *)malloc(item_count * sizeof(RSS_Item));
        if (sorted_items == NULL) {
            RSS_Writer_Free(writer);
            return -1;
        }
        memcpy(sorted_items, items, item_count * sizeof(RSS_Item));
        qsort(sorted_items, item_count, sizeof(RSS_Item), RSS_CompareItemsByDate);
        items_to_use = sorted_items;
    }
    
    for (int i = 0; i < item_count && writer->item_count < writer->max_items; i++) {
        if (RSS_AddItem(writer, &items_to_use[i]) != 0) {
            free(sorted_items);
            RSS_Writer_Free(writer);
            return -1;
        }
    }
    
    free(sorted_items);
    
    if (RSS_EndChannel(writer) != 0) {
        RSS_Writer_Free(writer);
        return -1;
    }
    
    size_t used = writer->data.buf.used;
    RSS_Writer_Free(writer);
    
    return (int)used;
}
