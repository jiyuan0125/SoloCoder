CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c11 -D_POSIX_C_SOURCE=200809L -D_XOPEN_SOURCE=700
LDFLAGS =

TARGET = dupfinder
SRCS = common.c dir_scan.c fingerprint.c duplicate.c main.c
OBJS = $(SRCS:.c=.o)

.PHONY: all clean test

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

common.o: common.c common.h
dir_scan.o: dir_scan.c dir_scan.h common.h
fingerprint.o: fingerprint.c fingerprint.h common.h
duplicate.o: duplicate.c duplicate.h fingerprint.h common.h
main.o: main.c common.h dir_scan.h fingerprint.h duplicate.h

test: $(TARGET)
	mkdir -p test_dir/subdir
	echo "test content" > test_dir/file1.txt
	echo "test content" > test_dir/file2.txt
	echo "test content" > test_dir/subdir/file3.txt
	echo "different content" > test_dir/unique.txt
	./$(TARGET) test_dir

clean:
	rm -f $(TARGET) $(OBJS)
	rm -rf test_dir
