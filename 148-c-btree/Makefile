CC = gcc
CFLAGS = -Wall -Wextra -g -pthread -O2
LDFLAGS = -pthread

TARGET = btree_demo

SRCS = main.c page.c cache.c btree.c
OBJS = $(SRCS:.c=.o)
HDRS = common.h page.h cache.h btree.h

.PHONY: all clean run

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c $(HDRS)
	$(CC) $(CFLAGS) -c -o $@ $<

run: $(TARGET)
	./$(TARGET)

clean:
	rm -f $(TARGET) $(OBJS) btree.db btree.wal
