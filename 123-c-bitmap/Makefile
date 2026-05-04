CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c11
LDFLAGS =

TARGET = checkin_demo

SRCS = checkin_common.c checkin_bitmap.c checkin_streak.c checkin_user.c main.c
OBJS = $(SRCS:.c=.o)
HDRS = checkin_common.h checkin_bitmap.h checkin_streak.h checkin_user.h

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
