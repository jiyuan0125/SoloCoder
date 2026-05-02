CC = gcc
CFLAGS = -Wall -Wextra -g -O2 -pthread
LDFLAGS = -pthread

TARGET = kvstore

SRCS = kv_store.c wal.c memtable.c sstable.c compaction.c db.c main.c
OBJS = $(SRCS:.c=.o)
DEPS = kv_store.h wal.h memtable.h sstable.h compaction.h db.h

.PHONY: all clean test

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c $(DEPS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(TARGET) $(OBJS)
	rm -rf ./test_db_*

test: $(TARGET)
	./$(TARGET)
