CC = gcc
CFLAGS = -Wall -Wextra -O2 -g
LDFLAGS = -lpthread

SERVER_SRCS = server_main.c heartbeat_io.c client_manager.c timeout_checker.c
SERVER_HDRS = heartbeat_protocol.h heartbeat_io.h client_manager.h timeout_checker.h
SERVER_OBJS = $(SERVER_SRCS:.c=.o)

CLIENT_SRCS = client_sim.c
CLIENT_OBJS = $(CLIENT_SRCS:.c=.o)

TARGETS = heartbeat_server heartbeat_client

.PHONY: all clean

all: $(TARGETS)

heartbeat_server: $(SERVER_OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

heartbeat_client: $(CLIENT_OBJS)
	$(CC) $(CFLAGS) -o $@ $^

%.o: %.c $(SERVER_HDRS)
	$(CC) $(CFLAGS) -c $< -o $@

client_sim.o: client_sim.c
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -f $(TARGETS) $(SERVER_OBJS) $(CLIENT_OBJS)
