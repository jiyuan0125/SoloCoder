CC = gcc
CFLAGS = -Wall -Wextra -O2 -std=c99 -D_POSIX_C_SOURCE=200809L
LDFLAGS =

TARGET = invite_code_tool

SRCS = common.c code_generator.c unique_check.c storage.c invite_service.c main.c
OBJS = $(SRCS:.c=.o)
HDRS = common.h code_generator.h unique_check.h storage.h invite_service.h

.PHONY: all clean test

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS)

%.o: %.c $(HDRS)
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(TARGET) $(OBJS) invite_codes.dat

test: $(TARGET)
	@echo "=== 测试邀请码系统 ==="
	@echo ""
	@echo "1. 查看系统状态:"
	./$(TARGET) status
	@echo ""
	@echo "2. 为用户 user_001 生成单个邀请码:"
	./$(TARGET) generate user_001
	@echo ""
	@echo "3. 为用户 user_002 批量生成 5 个邀请码:"
	./$(TARGET) batch user_002 5
	@echo ""
	@echo "4. 再次查看系统状态:"
	./$(TARGET) status
