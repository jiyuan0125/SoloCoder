CC = gcc
CFLAGS = -Wall -Wextra -g -O0
LDFLAGS =

TARGET = conf_demo
SRCS = main.c conf_parser.c conf_resolver.c conf_query.c
OBJS = $(SRCS:.c=.o)
HDRS = conf.h

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

valgrind: $(TARGET)
	valgrind --leak-check=full --show-leak-kinds=all ./$(TARGET)
