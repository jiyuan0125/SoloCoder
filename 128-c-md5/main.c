#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "md5_core.h"
#include "md5_stream.h"
#include "md5_file.h"

static void create_test_files(void) {
    FILE *f1 = fopen("test_empty.txt", "wb");
    if (f1) fclose(f1);

    FILE *f2 = fopen("test_hello.txt", "wb");
    if (f2) {
        fprintf(f2, "Hello, world!");
        fclose(f2);
    }

    FILE *f3 = fopen("test_binary.bin", "wb");
    if (f3) {
        int i;
        for (i = 0; i < 256; i++) {
            uint8_t b = (uint8_t)i;
            fwrite(&b, 1, 1, f3);
        }
        fclose(f3);
    }

    FILE *f4 = fopen("test_large.txt", "wb");
    if (f4) {
        int i;
        const char *pattern = "The quick brown fox jumps over the lazy dog. ";
        size_t len = strlen(pattern);
        for (i = 0; i < 10000; i++) {
            fwrite(pattern, 1, len, f4);
        }
        fclose(f4);
    }
}

static void test_memory_md5(void) {
    const char *test1 = "";
    const char *test2 = "a";
    const char *test3 = "abc";
    const char *test4 = "message digest";
    const char *test5 = "abcdefghijklmnopqrstuvwxyz";

    char result[MD5_STRING_LENGTH];

    printf("=== 内存数据MD5测试 ===\n");

    md5_memory_string((const uint8_t *)test1, strlen(test1), result);
    printf("空字符串:        %s (预期: d41d8cd98f00b204e9800998ecf8427e)\n", result);

    md5_memory_string((const uint8_t *)test2, strlen(test2), result);
    printf("\"a\":             %s (预期: 0cc175b9c0f1b6a831c399e269772661)\n", result);

    md5_memory_string((const uint8_t *)test3, strlen(test3), result);
    printf("\"abc\":           %s (预期: 900150983cd24fb0d6963f7d28e17f72)\n", result);

    md5_memory_string((const uint8_t *)test4, strlen(test4), result);
    printf("\"message digest\":%s (预期: f96b697d7cb7938d525a2f31aaf161d0)\n", result);

    md5_memory_string((const uint8_t *)test5, strlen(test5), result);
    printf("小写字母表:      %s (预期: c3fcd3d76192e4007dfb496cca67e13b)\n", result);

    printf("\n");
}

static void test_stream_md5(void) {
    const char *data = "The quick brown fox jumps over the lazy dog";
    size_t len = strlen(data);
    MD5_CTX ctx;
    char result[MD5_STRING_LENGTH];

    printf("=== 流式计算测试 ===\n");

    md5_stream_init(&ctx);
    md5_stream_update(&ctx, (const uint8_t *)data, len / 2);
    md5_stream_update(&ctx, (const uint8_t *)data + len / 2, len - len / 2);
    md5_stream_final_string(&ctx, result);

    printf("分块处理完整句子: %s (预期: 9e107d9d372bb6826bd81d3542a419d6)\n", result);
    printf("\n");
}

static void test_file_md5(void) {
    char result[MD5_STRING_LENGTH];
    int verify_result;

    printf("=== 文件MD5测试 ===\n");

    if (md5_file_string("test_empty.txt", result) == 0) {
        printf("空文件:              %s\n", result);
        printf("验证匹配(正确hash):  %d (预期: 1)\n", 
               md5_file_verify("test_empty.txt", "d41d8cd98f00b204e9800998ecf8427e"));
        printf("验证匹配(错误hash):  %d (预期: 0)\n",
               md5_file_verify("test_empty.txt", "00000000000000000000000000000000"));
    }

    if (md5_file_string("test_hello.txt", result) == 0) {
        printf("\n\"Hello, world!\"文件: %s\n", result);
    }

    if (md5_file_string("test_binary.bin", result) == 0) {
        printf("二进制文件(0-255):   %s\n", result);
    }

    if (md5_file_string("test_large.txt", result) == 0) {
        printf("较大测试文件:        %s\n", result);
    }

    printf("\n");
}

static void run_system_md5sum(void) {
    printf("=== 与系统md5sum对比 ===\n");
    printf("请手动运行以下命令验证：\n");
    printf("  md5sum test_empty.txt\n");
    printf("  md5sum test_hello.txt\n");
    printf("  md5sum test_binary.bin\n");
    printf("  md5sum test_large.txt\n");
    printf("\n");
}

int main(int argc, char *argv[]) {
    printf("========================================\n");
    printf("    MD5 文件完整性校验模块演示\n");
    printf("========================================\n\n");

    printf("创建测试文件...\n\n");
    create_test_files();

    test_memory_md5();
    test_stream_md5();
    test_file_md5();
    run_system_md5sum();

    if (argc > 1) {
        int i;
        printf("=== 命令行参数指定的文件 ===\n");
        for (i = 1; i < argc; i++) {
            char hash[MD5_STRING_LENGTH];
            if (md5_file_string(argv[i], hash) == 0) {
                printf("%s  %s\n", hash, argv[i]);
            } else {
                printf("无法读取文件: %s\n", argv[i]);
            }
        }
    }

    return 0;
}
