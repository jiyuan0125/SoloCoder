CC = gcc
CFLAGS = -Wall -Wextra -O2 -D_GNU_SOURCE
LDFLAGS =

TARGET = logsearch
SRCS = main.c common.c index.c file_search.c query.c
OBJS = $(SRCS:.c=.o)
DEPS = common.h index.h file_search.h query.h

.PHONY: all clean test

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c $(DEPS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(TARGET) $(OBJS) *.idx

test: $(TARGET)
	@echo "Creating test log file..."
	@echo '127.0.0.1 - - [01/Jan/2025:10:00:00 +0000] "GET /index.html HTTP/1.1" 200 1234' > test.log
	@echo '192.168.1.1 - - [01/Jan/2025:10:00:01 +0000] "POST /api/login HTTP/1.1" 500 567' >> test.log
	@echo '10.0.0.1 - - [01/Jan/2025:10:00:02 +0000] "GET /css/style.css HTTP/1.1" 200 9999' >> test.log
	@echo '127.0.0.1 - - [01/Jan/2025:10:00:03 +0000] "GET /api/users HTTP/1.1" 404 100' >> test.log
	@echo '192.168.1.2 - - [01/Jan/2025:10:00:04 +0000] "POST /api/data HTTP/1.1" 200 200' >> test.log
	@echo '10.0.0.5 - - [01/Jan/2025:10:00:05 +0000] "GET /favicon.ico HTTP/1.1" 500 123' >> test.log
	@echo '127.0.0.1 - - [01/Jan/2025:10:00:06 +0000] "GET /index.html HTTP/1.1" 200 1234' >> test.log
	@echo ''
	@echo "=== Test 1: Basic search for 'error' (should not find) ==="
	./$(TARGET) test.log error || true
	@echo ''
	@echo "=== Test 2: Search for '500' ==="
	./$(TARGET) test.log 500
	@echo ''
	@echo "=== Test 3: Search for 'GET' with context ==="
	./$(TARGET) -c 2 test.log GET
	@echo ''
	@echo "=== Test 4: Multiple keywords search ==="
	./$(TARGET) -c 1 test.log 500 404
	@echo ''
	@echo "=== Test 5: Search again (should use index, faster) ==="
	./$(TARGET) test.log 500
	@echo ''
	@echo "=== Test 6: Show help ==="
	./$(TARGET) -h
	@echo ''
	@echo "Cleaning up..."
	rm -f test.log test.log.idx
