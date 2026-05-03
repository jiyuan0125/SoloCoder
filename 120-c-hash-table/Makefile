CC = gcc
CFLAGS = -Wall -Wextra -O2 -g
LDFLAGS = 

TARGET = inventory_demo
SRCS = product_store.c inventory_trans.c filter_sort.c main.c
OBJS = $(SRCS:.c=.o)
HEADERS = product_store.h inventory_trans.h filter_sort.h

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
