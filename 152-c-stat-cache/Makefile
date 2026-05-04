CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c99 -D_POSIX_C_SOURCE=200809L
LDFLAGS =

TARGET = filewatch
SRCS = main.c snapshot.c scanner.c diff.c
OBJS = $(SRCS:.c=.o)
HDRS = snapshot.h scanner.h diff.h

.PHONY: all clean debug

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c $(HDRS)
	$(CC) $(CFLAGS) -c -o $@ $<

debug: CFLAGS = -Wall -Wextra -g -O0 -std=c99 -D_POSIX_C_SOURCE=200809L -DDEBUG
debug: clean $(TARGET)

clean:
	rm -f $(TARGET) $(OBJS)
	rm -f .filewatch.snap .filewatch.snap.bak .filewatch.snap.pending
