CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c11
LDFLAGS = 

SRCS = main.c pattern_parser.c matcher.c batch_matcher.c
OBJS = $(SRCS:.c=.o)
TARGET = wildcard_matcher

.PHONY: all clean test

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c wildcard.h
	$(CC) $(CFLAGS) -c -o $@ $<

test: $(TARGET)
	./$(TARGET)

clean:
	rm -f $(OBJS) $(TARGET)
