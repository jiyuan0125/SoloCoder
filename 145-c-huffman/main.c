#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <time.h>
#include "huffman.h"
#include "compress.h"
#include "decompress.h"

#define BUFFER_SIZE (1024 * 1024)

static int compare_files(const char *file1, const char *file2) {
    FILE *f1 = fopen(file1, "rb");
    FILE *f2 = fopen(file2, "rb");
    
    if (f1 == NULL || f2 == NULL) {
        if (f1 != NULL) fclose(f1);
        if (f2 != NULL) fclose(f2);
        return -1;
    }
    
    uint8_t *buf1 = (uint8_t *)malloc(BUFFER_SIZE);
    uint8_t *buf2 = (uint8_t *)malloc(BUFFER_SIZE);
    
    if (buf1 == NULL || buf2 == NULL) {
        free(buf1);
        free(buf2);
        fclose(f1);
        fclose(f2);
        return -1;
    }
    
    int result = 0;
    size_t read1, read2;
    
    do {
        read1 = fread(buf1, 1, BUFFER_SIZE, f1);
        read2 = fread(buf2, 1, BUFFER_SIZE, f2);
        
        if (read1 != read2) {
            result = -1;
            break;
        }
        
        if (read1 > 0 && memcmp(buf1, buf2, read1) != 0) {
            result = -1;
            break;
        }
    } while (read1 > 0);
    
    free(buf1);
    free(buf2);
    fclose(f1);
    fclose(f2);
    
    return result;
}

static off_t get_file_size(const char *filename) {
    struct stat st;
    if (stat(filename, &st) != 0) {
        return -1;
    }
    return st.st_size;
}

static void generate_test_file(const char *filename, size_t size) {
    FILE *f = fopen(filename, "wb");
    if (f == NULL) {
        printf("Error: Cannot create test file\n");
        return;
    }
    
    const char *words[] = {
        "INFO", "WARN", "ERROR", "DEBUG", "TRACE",
        "User", "System", "Database", "Network", "Server",
        "connection", "failed", "success", "timeout", "retry",
        "0123456789", "abcdefghij", "klmnopqrst", "uvwxyz"
    };
    int num_words = sizeof(words) / sizeof(words[0]);
    
    srand((unsigned int)time(NULL));
    
    size_t written = 0;
    char line[256];
    
    while (written < size) {
        int len = snprintf(line, sizeof(line), 
            "%s [%s] %s %s - %s %s %s\n",
            "2024-01-15 10:30:45",
            words[rand() % 5],
            words[5 + rand() % 5],
            words[10 + rand() % 5],
            words[rand() % num_words],
            words[rand() % num_words],
            words[rand() % num_words]
        );
        
        size_t to_write = (written + len > size) ? (size - written) : (size_t)len;
        fwrite(line, 1, to_write, f);
        written += to_write;
    }
    
    fclose(f);
    printf("Generated test file: %s (%zu bytes)\n", filename, size);
}

static void generate_single_byte_file(const char *filename, size_t size, uint8_t byte) {
    FILE *f = fopen(filename, "wb");
    if (f == NULL) {
        printf("Error: Cannot create single byte file\n");
        return;
    }
    
    uint8_t *buffer = (uint8_t *)malloc(BUFFER_SIZE);
    if (buffer == NULL) {
        fclose(f);
        return;
    }
    
    memset(buffer, byte, BUFFER_SIZE);
    
    size_t remaining = size;
    while (remaining > 0) {
        size_t to_write = (remaining > BUFFER_SIZE) ? BUFFER_SIZE : remaining;
        fwrite(buffer, 1, to_write, f);
        remaining -= to_write;
    }
    
    free(buffer);
    fclose(f);
    printf("Generated single byte file: %s (%zu bytes, byte=0x%02X)\n", 
           filename, size, byte);
}

int main(int argc, char *argv[]) {
    printf("========================================\n");
    printf("Huffman Compression Tool Demo\n");
    printf("========================================\n\n");
    
    const char *test_file = "test_log.txt";
    const char *compressed_file = "test_log.huff";
    const char *decompressed_file = "test_log_decompressed.txt";
    
    const char *single_byte_file = "test_single.txt";
    const char *single_compressed = "test_single.huff";
    const char *single_decompressed = "test_single_decompressed.txt";
    
    printf("--- Test 1: Normal Text File ---\n");
    generate_test_file(test_file, 1024 * 1024);
    
    clock_t start = clock();
    if (compress_file(test_file, compressed_file) != 0) {
        printf("Error: Compression failed\n");
        return 1;
    }
    clock_t compress_time = clock() - start;
    
    start = clock();
    if (decompress_file(compressed_file, decompressed_file) != 0) {
        printf("Error: Decompression failed\n");
        return 1;
    }
    clock_t decompress_time = clock() - start;
    
    off_t original_size = get_file_size(test_file);
    off_t compressed_size = get_file_size(compressed_file);
    
    printf("Original size:   %lld bytes\n", (long long)original_size);
    printf("Compressed size: %lld bytes\n", (long long)compressed_size);
    printf("Compression ratio: %.2f%%\n", 
           (1.0 - (double)compressed_size / original_size) * 100);
    printf("Compression time: %.3f seconds\n", 
           (double)compress_time / CLOCKS_PER_SEC);
    printf("Decompression time: %.3f seconds\n", 
           (double)decompress_time / CLOCKS_PER_SEC);
    
    if (compare_files(test_file, decompressed_file) == 0) {
        printf("✓ Success: Decompressed file matches original!\n\n");
    } else {
        printf("✗ Error: Decompressed file does NOT match original!\n\n");
        return 1;
    }
    
    printf("--- Test 2: Single Byte File (Edge Case) ---\n");
    generate_single_byte_file(single_byte_file, 500 * 1024, 0x41);
    
    if (compress_file(single_byte_file, single_compressed) != 0) {
        printf("Error: Single byte compression failed\n");
        return 1;
    }
    
    if (decompress_file(single_compressed, single_decompressed) != 0) {
        printf("Error: Single byte decompression failed\n");
        return 1;
    }
    
    off_t single_original = get_file_size(single_byte_file);
    off_t single_compressed_size = get_file_size(single_compressed);
    
    printf("Original size:   %lld bytes\n", (long long)single_original);
    printf("Compressed size: %lld bytes\n", (long long)single_compressed_size);
    printf("Compression ratio: %.2f%%\n", 
           (1.0 - (double)single_compressed_size / single_original) * 100);
    
    if (compare_files(single_byte_file, single_decompressed) == 0) {
        printf("✓ Success: Single byte file decompressed correctly!\n\n");
    } else {
        printf("✗ Error: Single byte decompression failed!\n\n");
        return 1;
    }
    
    printf("--- Test 3: Buffer-based Compression ---\n");
    const char *test_data = "Hello, Huffman encoding! This is a test string for compression.";
    size_t data_len = strlen(test_data);
    
    uint8_t *compressed_buf = NULL;
    size_t compressed_buf_size = 0;
    
    if (compress_buffer((const uint8_t *)test_data, data_len, 
                        &compressed_buf, &compressed_buf_size) != 0) {
        printf("Error: Buffer compression failed\n");
        return 1;
    }
    
    printf("Original data size: %zu bytes\n", data_len);
    printf("Compressed buffer size: %zu bytes\n", compressed_buf_size);
    
    uint8_t *decompressed_buf = NULL;
    size_t decompressed_buf_size = 0;
    
    if (decompress_buffer(compressed_buf, compressed_buf_size,
                          &decompressed_buf, &decompressed_buf_size) != 0) {
        printf("Error: Buffer decompression failed\n");
        free(compressed_buf);
        return 1;
    }
    
    if (decompressed_buf_size == data_len && 
        memcmp(test_data, decompressed_buf, data_len) == 0) {
        printf("✓ Success: Buffer compression/decompression works!\n");
        printf("  Original: \"%s\"\n", test_data);
        printf("  Decompressed: \"%.*s\"\n", (int)decompressed_buf_size, decompressed_buf);
    } else {
        printf("✗ Error: Buffer data mismatch!\n");
    }
    
    free(compressed_buf);
    free(decompressed_buf);
    
    printf("\n========================================\n");
    printf("All tests passed!\n");
    printf("========================================\n");
    
    return 0;
}
