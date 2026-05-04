#ifndef XML_ESCAPE_H
#define XML_ESCAPE_H

#include <stddef.h>

size_t XML_Escape_Length(const char *input);
size_t XML_Escape(char *output, size_t output_size, const char *input);

int XML_IsValidUTF8(const char *input);
size_t XML_ValidateAndFixUTF8(char *output, size_t output_size, const char *input);

size_t XML_CDATA_SafeLength(const char *input);
size_t XML_CDATA_Safe(char *output, size_t output_size, const char *input);

#define XML_REPLACEMENT_CHAR 0xFFFD

#endif
