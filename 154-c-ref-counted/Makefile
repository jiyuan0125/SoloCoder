CC = gcc
CFLAGS = -Wall -Wextra -Werror -O2 -g -pthread -std=c11
LDFLAGS = -pthread
CPPFLAGS = -I.

SRCS = ref_count.c object_store.c lifecycle.c main.c
OBJS = $(SRCS:.c=.o)
TARGET = shared_ptr_test

.PHONY: all clean run debug

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c
	$(CC) $(CPPFLAGS) $(CFLAGS) -c -o $@ $<

ref_count.o: ref_count.c ref_count.h shared_ptr.h
object_store.o: object_store.c object_store.h shared_ptr.h ref_count.h
lifecycle.o: lifecycle.c lifecycle.h shared_ptr.h object_store.h ref_count.h
main.o: main.c shared_ptr.h ref_count.h object_store.h lifecycle.h

run: $(TARGET)
	./$(TARGET)

debug: CFLAGS = -Wall -Wextra -O0 -g -pthread -std=c11 -DDEBUG
debug: clean $(TARGET)

clean:
	rm -f $(OBJS) $(TARGET)
	rm -f *.core core.*
