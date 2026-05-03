CC = gcc
CFLAGS = -Wall -Wextra -O2 -g -D_POSIX_C_SOURCE=200809L
LDFLAGS = -pthread
TARGET = daemon_demo

SRCS = main.c signal_handler.c shutdown.c pidfile.c
OBJS = $(SRCS:.c=.o)
HDRS = signal_handler.h shutdown.h pidfile.h

.PHONY: all clean run

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c $(HDRS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(TARGET) $(OBJS)

run: $(TARGET)
	./$(TARGET) -f
