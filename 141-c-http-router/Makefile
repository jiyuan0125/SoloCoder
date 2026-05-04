CC = gcc
CFLAGS = -Wall -Wextra -O2 -g
LDFLAGS = 

TARGET = router_demo

SRCS = main.c types.c router.c matcher.c handler.c
OBJS = $(SRCS:.c=.o)
HDRS = types.h router.h matcher.h handler.h

.PHONY: all clean run

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c $(HDRS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJS) $(TARGET)

run: $(TARGET)
	./$(TARGET)
