#include "xml_builder.h"
#include "xml_escape.h"
#include <stdlib.h>
#include <string.h>

#define INDENT_SPACES 2

static void XML_WriteIndent(XML_Builder *builder) {
    if (!builder->pretty_print) return;
    
    int spaces = builder->indent_level * INDENT_SPACES;
    for (int i = 0; i < spaces; i++) {
        if (builder->type == XML_OUTPUT_FILE) {
            fputc(' ', builder->data.file);
        } else {
            if (builder->data.buf.used < builder->data.buf.size - 1) {
                builder->data.buf.buffer[builder->data.buf.used++] = ' ';
            }
        }
    }
}

static void XML_WriteNewline(XML_Builder *builder) {
    if (!builder->pretty_print) return;
    
    if (builder->type == XML_OUTPUT_FILE) {
        fputc('\n', builder->data.file);
    } else {
        if (builder->data.buf.used < builder->data.buf.size - 1) {
            builder->data.buf.buffer[builder->data.buf.used++] = '\n';
        }
    }
}

static int XML_WriteRaw(XML_Builder *builder, const char *str) {
    if (builder == NULL || str == NULL) return -1;
    
    if (builder->type == XML_OUTPUT_FILE) {
        return fputs(str, builder->data.file) >= 0 ? 0 : -1;
    } else {
        size_t len = strlen(str);
        size_t available = builder->data.buf.size - builder->data.buf.used - 1;
        size_t to_write = (len < available) ? len : available;
        
        memcpy(builder->data.buf.buffer + builder->data.buf.used, str, to_write);
        builder->data.buf.used += to_write;
        builder->data.buf.buffer[builder->data.buf.used] = '\0';
        
        return 0;
    }
}

static int XML_WriteEscaped(XML_Builder *builder, const char *text) {
    if (builder == NULL || text == NULL) return -1;
    
    size_t escaped_len = XML_Escape_Length(text);
    char *escaped = (char *)malloc(escaped_len + 1);
    if (escaped == NULL) return -1;
    
    XML_Escape(escaped, escaped_len + 1, text);
    int result = XML_WriteRaw(builder, escaped);
    free(escaped);
    
    return result;
}

static int XML_WriteCDataContent(XML_Builder *builder, const char *content) {
    if (builder == NULL || content == NULL) return -1;
    
    size_t safe_len = XML_CDATA_SafeLength(content);
    char *safe_content = (char *)malloc(safe_len + 1);
    if (safe_content == NULL) return -1;
    
    XML_CDATA_Safe(safe_content, safe_len + 1, content);
    
    const char *start = safe_content;
    const char *p = safe_content;
    
    while (*p != '\0') {
        if (*p == ']' && *(p + 1) == ']' && *(p + 2) == '>') {
            char temp[3] = {0};
            strncpy(temp, start, p - start + 2);
            XML_WriteRaw(builder, temp);
            XML_WriteRaw(builder, "]]><![CDATA[");
            XML_WriteRaw(builder, ">");
            p += 3;
            start = p;
        } else {
            p++;
        }
    }
    
    if (start < p) {
        XML_WriteRaw(builder, start);
    }
    
    free(safe_content);
    return 0;
}

XML_Builder* XML_Builder_ToFile(const char *filepath, int pretty_print) {
    if (filepath == NULL) return NULL;
    
    FILE *file = fopen(filepath, "w");
    if (file == NULL) return NULL;
    
    XML_Builder *builder = (XML_Builder *)malloc(sizeof(XML_Builder));
    if (builder == NULL) {
        fclose(file);
        return NULL;
    }
    
    builder->type = XML_OUTPUT_FILE;
    builder->data.file = file;
    builder->indent_level = 0;
    builder->pretty_print = pretty_print;
    
    return builder;
}

XML_Builder* XML_Builder_ToBuffer(char *buffer, size_t size, int pretty_print) {
    if (buffer == NULL || size == 0) return NULL;
    
    XML_Builder *builder = (XML_Builder *)malloc(sizeof(XML_Builder));
    if (builder == NULL) return NULL;
    
    builder->type = XML_OUTPUT_BUFFER;
    builder->data.buf.buffer = buffer;
    builder->data.buf.size = size;
    builder->data.buf.used = 0;
    builder->indent_level = 0;
    builder->pretty_print = pretty_print;
    
    buffer[0] = '\0';
    
    return builder;
}

void XML_Builder_Free(XML_Builder *builder) {
    if (builder == NULL) return;
    
    if (builder->type == XML_OUTPUT_FILE && builder->data.file != NULL) {
        fclose(builder->data.file);
    }
    
    free(builder);
}

int XML_WriteDeclaration(XML_Builder *builder) {
    if (builder == NULL) return -1;
    
    int result = XML_WriteRaw(builder, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>");
    if (result == 0) {
        XML_WriteNewline(builder);
    }
    return result;
}

int XML_OpenElement(XML_Builder *builder, const char *name) {
    if (builder == NULL || name == NULL) return -1;
    
    XML_WriteIndent(builder);
    XML_WriteRaw(builder, "<");
    XML_WriteRaw(builder, name);
    XML_WriteRaw(builder, ">");
    XML_WriteNewline(builder);
    
    XML_IncreaseIndent(builder);
    return 0;
}

int XML_CloseElement(XML_Builder *builder, const char *name) {
    if (builder == NULL || name == NULL) return -1;
    
    XML_DecreaseIndent(builder);
    XML_WriteIndent(builder);
    XML_WriteRaw(builder, "</");
    XML_WriteRaw(builder, name);
    XML_WriteRaw(builder, ">");
    XML_WriteNewline(builder);
    
    return 0;
}

int XML_ElementWithText(XML_Builder *builder, const char *name, const char *text) {
    if (builder == NULL || name == NULL) return -1;
    if (text == NULL) text = "";
    
    XML_WriteIndent(builder);
    XML_WriteRaw(builder, "<");
    XML_WriteRaw(builder, name);
    XML_WriteRaw(builder, ">");
    XML_WriteEscaped(builder, text);
    XML_WriteRaw(builder, "</");
    XML_WriteRaw(builder, name);
    XML_WriteRaw(builder, ">");
    XML_WriteNewline(builder);
    
    return 0;
}

int XML_ElementWithCData(XML_Builder *builder, const char *name, const char *content) {
    if (builder == NULL || name == NULL) return -1;
    if (content == NULL) content = "";
    
    XML_WriteIndent(builder);
    XML_WriteRaw(builder, "<");
    XML_WriteRaw(builder, name);
    XML_WriteRaw(builder, "><![CDATA[");
    XML_WriteCDataContent(builder, content);
    XML_WriteRaw(builder, "]]></");
    XML_WriteRaw(builder, name);
    XML_WriteRaw(builder, ">");
    XML_WriteNewline(builder);
    
    return 0;
}

int XML_Attribute(XML_Builder *builder, const char *name, const char *value) {
    if (builder == NULL || name == NULL || value == NULL) return -1;
    
    XML_WriteRaw(builder, " ");
    XML_WriteRaw(builder, name);
    XML_WriteRaw(builder, "=\"");
    XML_WriteEscaped(builder, value);
    XML_WriteRaw(builder, "\"");
    
    return 0;
}

int XML_StartElementWithAttrs(XML_Builder *builder, const char *name, 
                               const char **attr_names, const char **attr_values, 
                               int attr_count) {
    if (builder == NULL || name == NULL) return -1;
    
    XML_WriteIndent(builder);
    XML_WriteRaw(builder, "<");
    XML_WriteRaw(builder, name);
    
    if (attr_names != NULL && attr_values != NULL && attr_count > 0) {
        for (int i = 0; i < attr_count; i++) {
            if (attr_names[i] != NULL && attr_values[i] != NULL) {
                XML_Attribute(builder, attr_names[i], attr_values[i]);
            }
        }
    }
    
    XML_WriteRaw(builder, ">");
    XML_WriteNewline(builder);
    XML_IncreaseIndent(builder);
    
    return 0;
}

void XML_IncreaseIndent(XML_Builder *builder) {
    if (builder == NULL) return;
    builder->indent_level++;
}

void XML_DecreaseIndent(XML_Builder *builder) {
    if (builder == NULL) return;
    if (builder->indent_level > 0) {
        builder->indent_level--;
    }
}

int XML_Flush(XML_Builder *builder) {
    if (builder == NULL) return -1;
    
    if (builder->type == XML_OUTPUT_FILE) {
        return fflush(builder->data.file);
    }
    
    return 0;
}
