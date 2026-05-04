#ifndef XML_BUILDER_H
#define XML_BUILDER_H

#include <stdio.h>
#include <stddef.h>

typedef enum {
    XML_OUTPUT_FILE,
    XML_OUTPUT_BUFFER
} XML_OutputType;

typedef struct {
    XML_OutputType type;
    union {
        FILE *file;
        struct {
            char *buffer;
            size_t size;
            size_t used;
        } buf;
    } data;
    int indent_level;
    int pretty_print;
} XML_Builder;

XML_Builder* XML_Builder_ToFile(const char *filepath, int pretty_print);
XML_Builder* XML_Builder_ToBuffer(char *buffer, size_t size, int pretty_print);
void XML_Builder_Free(XML_Builder *builder);

int XML_WriteDeclaration(XML_Builder *builder);

int XML_OpenElement(XML_Builder *builder, const char *name);
int XML_CloseElement(XML_Builder *builder, const char *name);

int XML_ElementWithText(XML_Builder *builder, const char *name, const char *text);
int XML_ElementWithCData(XML_Builder *builder, const char *name, const char *content);

int XML_Attribute(XML_Builder *builder, const char *name, const char *value);

int XML_StartElementWithAttrs(XML_Builder *builder, const char *name, 
                               const char **attr_names, const char **attr_values, 
                               int attr_count);

void XML_IncreaseIndent(XML_Builder *builder);
void XML_DecreaseIndent(XML_Builder *builder);

int XML_Flush(XML_Builder *builder);

#endif
