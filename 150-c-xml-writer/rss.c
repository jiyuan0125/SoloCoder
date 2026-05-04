#include "rss.h"
#include "xml_builder.h"
#include "time_utils.h"
#include <stdlib.h>
#include <string.h>

#define DEFAULT_MAX_ITEMS 20

RSS_Writer* RSS_Writer_ToFile(const char *filepath, int max_items) {
    if (filepath == NULL) return NULL;
    
    RSS_Writer *writer = (RSS_Writer *)malloc(sizeof(RSS_Writer));
    if (writer == NULL) return NULL;
    
    writer->builder = XML_Builder_ToFile(filepath, 1);
    if (writer->builder == NULL) {
        free(writer);
        return NULL;
    }
    
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
    
    writer->builder = XML_Builder_ToBuffer(buffer, size, 1);
    if (writer->builder == NULL) {
        free(writer);
        return NULL;
    }
    
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
    
    XML_Builder_Free(writer->builder);
    free(writer);
}

int RSS_StartChannel(RSS_Writer *writer, const RSS_Channel *channel) {
    if (writer == NULL || channel == NULL || writer->started) {
        return -1;
    }
    
    XML_WriteDeclaration(writer->builder);
    
    const char *attr_names[] = {"version"};
    const char *attr_values[] = {RSS_VERSION};
    XML_StartElementWithAttrs(writer->builder, "rss", attr_names, attr_values, 1);
    
    XML_OpenElement(writer->builder, "channel");
    
    char time_buf[RFC822_DATE_LEN];
    
    XML_ElementWithText(writer->builder, "title", channel->title);
    XML_ElementWithText(writer->builder, "link", channel->link);
    XML_ElementWithText(writer->builder, "description", channel->description);
    
    if (channel->language[0] != '\0') {
        XML_ElementWithText(writer->builder, "language", channel->language);
    }
    
    if (channel->last_build_date > 0) {
        Time_ToRFC822(time_buf, sizeof(time_buf), channel->last_build_date);
        XML_ElementWithText(writer->builder, "lastBuildDate", time_buf);
    }
    
    if (channel->pub_date > 0) {
        Time_ToRFC822(time_buf, sizeof(time_buf), channel->pub_date);
        XML_ElementWithText(writer->builder, "pubDate", time_buf);
    }
    
    writer->started = 1;
    XML_Flush(writer->builder);
    
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
    
    XML_OpenElement(writer->builder, "item");
    
    XML_ElementWithText(writer->builder, "title", item->title);
    XML_ElementWithText(writer->builder, "link", item->link);
    
    if (item->is_html_description) {
        XML_ElementWithCData(writer->builder, "description", item->description);
    } else {
        XML_ElementWithText(writer->builder, "description", item->description);
    }
    
    if (item->pub_date > 0) {
        Time_ToRFC822(time_buf, sizeof(time_buf), item->pub_date);
        XML_ElementWithText(writer->builder, "pubDate", time_buf);
    }
    
    if (item->author[0] != '\0') {
        XML_ElementWithText(writer->builder, "author", item->author);
    }
    
    for (int i = 0; i < item->category_count; i++) {
        if (item->categories[i][0] != '\0') {
            XML_ElementWithText(writer->builder, "category", item->categories[i]);
        }
    }
    
    XML_CloseElement(writer->builder, "item");
    
    writer->item_count++;
    XML_Flush(writer->builder);
    
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
    
    XML_CloseElement(writer->builder, "channel");
    XML_CloseElement(writer->builder, "rss");
    
    writer->closed = 1;
    XML_Flush(writer->builder);
    
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
    
    size_t used = 0;
    if (writer->builder != NULL) {
        used = writer->builder->data.buf.used;
    }
    
    RSS_Writer_Free(writer);
    
    return (int)used;
}
