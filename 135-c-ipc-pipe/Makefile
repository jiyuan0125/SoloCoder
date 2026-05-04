CC = gcc
CFLAGS = -Wall -Wextra -g -O0
LDFLAGS = -lpthread

TARGET = ipc_pipe_demo

SRCS = main.c message_frame.c pipe_manager.c process_manager.c
OBJS = $(SRCS:.c=.o)
HEADERS = message_frame.h pipe_manager.h process_manager.h

.PHONY: all clean run

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c $(HEADERS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(TARGET) $(OBJS)

run: $(TARGET)
	./$(TARGET)
