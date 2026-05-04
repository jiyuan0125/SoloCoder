CC=gcc
CFLAGS=-Wall -Wextra -O2 -g -D_GNU_SOURCE
LDFLAGS=

TARGET=cmd_executor
SRCS=main.c cmd_exec.c scheduler.c result.c
OBJS=$(SRCS:.c=.o)
HDRS=common.h cmd_exec.h scheduler.h result.h

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
