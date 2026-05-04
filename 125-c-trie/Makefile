CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c11
LDFLAGS = 

TARGET = search_suggest
SRCS = trie.c query.c dict.c main.c
OBJS = $(SRCS:.c=.o)
HDRS = trie.h query.h dict.h

.PHONY: all clean run

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c $(HDRS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(TARGET) $(OBJS)

run: $(TARGET)
	./$(TARGET)
