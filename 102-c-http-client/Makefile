CC = gcc
CFLAGS = -Wall -Wextra -g -O2

PKG_CONFIG := pkg-config
HAVE_OPENSSL := $(shell $(PKG_CONFIG) --exists openssl && echo yes || echo no)

ifeq ($(HAVE_OPENSSL),yes)
    CFLAGS += -DWITH_SSL
    LDFLAGS += $(shell $(PKG_CONFIG) --libs openssl)
    $(info OpenSSL found: HTTPS support enabled)
else
    $(info WARNING: OpenSSL not found. HTTPS support will be disabled.)
    $(info Install libssl-dev to enable HTTPS: sudo apt-get install libssl-dev)
endif

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

ssl:
	@if [ "$(HAVE_OPENSSL)" != "yes" ]; then \
	    echo "Error: OpenSSL development files not found."; \
	    echo "Please install: sudo apt-get install libssl-dev"; \
	    exit 1; \
	fi
	$(MAKE) all

clean:
	rm -f $(OBJS) $(TARGET)

http_alert.o: http_alert.c http_alert.h
http_connection.o: http_connection.c http_connection.h http_alert.h
http_request.o: http_request.c http_request.h http_connection.h http_alert.h
retry_scheduler.o: retry_scheduler.c retry_scheduler.h http_request.h http_alert.h
main.o: main.c http_alert.h http_request.h retry_scheduler.h
