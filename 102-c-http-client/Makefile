CC = gcc
CFLAGS = -Wall -Wextra -g -O2
LDFLAGS =

TARGET = http_alert_demo

SRCS = main.c \
       http_alert.c \
       http_connection.c \
       http_request.c \
       retry_scheduler.c

OBJS = $(SRCS:.c=.o)

.PHONY: all clean ssl

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

ssl: CFLAGS += -DWITH_SSL
ssl: LDFLAGS += -lssl -lcrypto
ssl: clean $(TARGET)

clean:
	rm -f $(OBJS) $(TARGET)

http_alert.o: http_alert.c http_alert.h
http_connection.o: http_connection.c http_connection.h http_alert.h
http_request.o: http_request.c http_request.h http_connection.h http_alert.h
retry_scheduler.o: retry_scheduler.c retry_scheduler.h http_request.h http_alert.h
main.o: main.c http_alert.h http_request.h retry_scheduler.h
