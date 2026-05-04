CC = gcc
CFLAGS = -Wall -Wextra -O2 -g
TARGET = timefmt
SRCS = main.c timezone.c time_formatter.c time_parser.c
OBJS = $(SRCS:.c=.o)
HDRS = timezone.h time_formatter.h time_parser.h

.PHONY: all clean test

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^

%.o: %.c $(HDRS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(TARGET) $(OBJS)

test: $(TARGET)
	./$(TARGET) demo
