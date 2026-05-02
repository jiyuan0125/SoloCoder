CC = gcc
CFLAGS = -Wall -Wextra -g
LDFLAGS =

SRCS = regex_compile.c nfa.c regex_match.c main.c
OBJS = $(SRCS:.c=.o)
TARGET = regex_test

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJS) $(TARGET)

run: $(TARGET)
	./$(TARGET)
