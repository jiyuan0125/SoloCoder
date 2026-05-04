CC = gcc
CFLAGS = -Wall -Wextra -O2 -D_POSIX_C_SOURCE=200809L -D_DEFAULT_SOURCE
LDFLAGS =

TARGET = dns_resolver

SRCS = dns_protocol.c dns_query.c dns_cache.c main.c
OBJS = $(SRCS:.c=.o)

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

dns_protocol.o: dns_protocol.c dns_protocol.h
dns_query.o: dns_query.c dns_query.h dns_protocol.h
dns_cache.o: dns_cache.c dns_cache.h dns_query.h dns_protocol.h
main.o: main.c dns_cache.h dns_query.h dns_protocol.h

clean:
	rm -f $(TARGET) $(OBJS)

run: $(TARGET)
	./$(TARGET)
