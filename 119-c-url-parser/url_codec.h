#ifndef URL_CODEC_H
#define URL_CODEC_H

#include <stddef.h>

size_t url_encode(const char *src, char *dst, size_t dst_size);
size_t url_decode(const char *src, char *dst, size_t dst_size);
char *url_encode_alloc(const char *src);
char *url_decode_alloc(const char *src);

#endif
