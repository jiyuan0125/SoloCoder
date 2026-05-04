#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <time.h>
#include "crc32.h"
#include "crc32_stream.h"
#include "crc32_table.h"

static const char *test_string = "The quick brown fox jumps over the lazy dog";
static const char *test_string_expected_crc32 = "414FA339";

static void test_basic_compute(void)
{
    uint32_t crc;
    char hex_str[CRC32_HEX_STRING_LENGTH];
    
    printf("=== 基础 CRC32 计算测试 ===\n");
    printf("测试字符串: \"%s\"\n", test_string);
    printf("预期 CRC32:  %s\n", test_string_expected_crc32);
    
    crc = crc32_compute((const uint8_t *)test_string, strlen(test_string));
    crc32_to_hex(crc, hex_str);
    
    printf("计算结果:    0x%08X (%s)\n", crc, hex_str);
    
    if (strcmp(hex_str, test_string_expected_crc32) == 0) {
        printf("测试结果:    ✓ 通过\n\n");
    } else {
        printf("测试结果:    ✗ 失败 (预期: %s)\n\n", test_string_expected_crc32);
    }
}

static void test_verify(void)
{
    int result;
    uint32_t expected_crc;
    
    printf("=== 验证功能测试 ===\n");
    
    expected_crc = crc32_parse_hex(test_string_expected_crc32);
    
    result = crc32_verify((const uint8_t *)test_string, strlen(test_string), expected_crc);
    printf("使用 uint32_t 验证: %s\n", result ? "✓ 通过" : "✗ 失败");
    
    result = crc32_verify_hex((const uint8_t *)test_string, strlen(test_string), test_string_expected_crc32);
    printf("使用十六进制字符串验证: %s\n", result ? "✓ 通过" : "✗ 失败");
    
    printf("修改一个字节后的验证测试:\n");
    {
        char modified[256];
        strcpy(modified, test_string);
        modified[0] = 'X';
        result = crc32_verify_hex((const uint8_t *)modified, strlen(modified), test_string_expected_crc32);
        printf("  修改后验证: %s (预期失败)\n\n", result ? "✗ 意外通过" : "✓ 正确失败");
    }
}

static void test_streaming(void)
{
    crc32_stream_t stream;
    uint32_t crc_stream, crc_direct;
    char hex_stream[CRC32_HEX_STRING_LENGTH];
    char hex_direct[CRC32_HEX_STRING_LENGTH];
    
    const char *test_data = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    size_t len = strlen(test_data);
    
    printf("=== 流式计算测试 ===\n");
    printf("测试数据: \"%s\"\n", test_data);
    
    crc_direct = crc32_compute((const uint8_t *)test_data, len);
    crc32_to_hex(crc_direct, hex_direct);
    printf("一次性计算结果: %s\n", hex_direct);
    
    crc32_stream_init(&stream);
    
    crc32_stream_update(&stream, (const uint8_t *)test_data, 5);
    crc32_stream_update(&stream, (const uint8_t *)(test_data + 5), 10);
    crc32_stream_update(&stream, (const uint8_t *)(test_data + 15), len - 15);
    
    crc_stream = crc32_stream_final(&stream);
    crc32_to_hex(crc_stream, hex_stream);
    printf("分块计算结果:   %s\n", hex_stream);
    
    if (crc_stream == crc_direct) {
        printf("测试结果:       ✓ 流式计算与一次性计算结果一致\n\n");
    } else {
        printf("测试结果:       ✗ 结果不一致 (直接: %s, 流式: %s)\n\n", hex_direct, hex_stream);
    }
}

static void test_combine(void)
{
    const char *part1 = "Hello, ";
    const char *part2 = "World!";
    const char *combined = "Hello, World!";
    
    uint32_t crc1, crc2, crc_combined_direct, crc_combined_computed;
    char hex1[CRC32_HEX_STRING_LENGTH];
    char hex2[CRC32_HEX_STRING_LENGTH];
    char hex_direct[CRC32_HEX_STRING_LENGTH];
    char hex_computed[CRC32_HEX_STRING_LENGTH];
    
    printf("=== 合并计算测试 ===\n");
    printf("数据段1: \"%s\"\n", part1);
    printf("数据段2: \"%s\"\n", part2);
    printf("合并后:  \"%s\"\n", combined);
    
    crc1 = crc32_compute((const uint8_t *)part1, strlen(part1));
    crc2 = crc32_compute((const uint8_t *)part2, strlen(part2));
    crc_combined_direct = crc32_compute((const uint8_t *)combined, strlen(combined));
    
    crc32_to_hex(crc1, hex1);
    crc32_to_hex(crc2, hex2);
    crc32_to_hex(crc_combined_direct, hex_direct);
    
    printf("CRC1:  %s\n", hex1);
    printf("CRC2:  %s\n", hex2);
    printf("直接计算合并 CRC: %s\n", hex_direct);
    
    crc_combined_computed = crc32_combine(crc1, crc2, strlen(part2));
    crc32_to_hex(crc_combined_computed, hex_computed);
    printf("使用 crc32_combine 计算: %s\n", hex_computed);
    
    if (crc_combined_computed == crc_combined_direct) {
        printf("测试结果: ✓ 合并计算结果正确\n\n");
    } else {
        printf("测试结果: ✗ 合并计算结果错误 (预期: %s, 实际: %s)\n\n", hex_direct, hex_computed);
    }
}

static void test_performance_estimate(void)
{
    const size_t test_size = 10 * 1024 * 1024;
    uint8_t *test_data;
    uint32_t crc;
    clock_t start, end;
    double elapsed, throughput;
    
    printf("=== 性能估算测试 ===\n");
    printf("测试数据大小: %zu MB\n", test_size / (1024 * 1024));
    
    test_data = (uint8_t *)malloc(test_size);
    if (test_data == NULL) {
        printf("无法分配内存，跳过性能测试\n\n");
        return;
    }
    
    memset(test_data, 0xAA, test_size);
    
    crc32_table_init();
    
    start = clock();
    crc = crc32_compute(test_data, test_size);
    end = clock();
    
    elapsed = (double)(end - start) / CLOCKS_PER_SEC;
    throughput = (double)test_size / elapsed / (1024 * 1024);
    
    printf("计算耗时:   %.4f 秒\n", elapsed);
    printf("处理速度:   %.2f MB/s\n", throughput);
    printf("CRC32 结果: 0x%08X\n", crc);
    
    if (throughput > 100) {
        printf("性能评估:   ✓ 良好 (超过 100 MB/s)\n");
    } else if (throughput > 50) {
        printf("性能评估:   ☐ 一般 (50-100 MB/s)\n");
    } else {
        printf("性能评估:   ✗ 较慢 (低于 50 MB/s)\n");
    }
    printf("(注: 优化编译后查表法 CRC32 应能达到数 GB/s)\n\n");
    
    free(test_data);
}

int main(void)
{
    printf("========================================\n");
    printf("    CRC32 数据完整性校验模块演示\n");
    printf("========================================\n\n");
    
    printf("模块信息:\n");
    printf("  - 多项式:     0x%08X (反向表示法)\n", CRC32_POLYNOMIAL);
    printf("  - 初始值:     0x%08X\n", CRC32_INIT_VALUE);
    printf("  - 最终异或:   0x%08X\n", CRC32_FINAL_XOR);
    printf("  - 兼容性:     与以太网、ZIP、crc32 命令等标准兼容\n\n");
    
    test_basic_compute();
    test_verify();
    test_streaming();
    test_combine();
    test_performance_estimate();
    
    printf("========================================\n");
    printf("    演示完成\n");
    printf("========================================\n");
    
    return 0;
}
