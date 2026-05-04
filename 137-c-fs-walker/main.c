#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <getopt.h>
#include <inttypes.h>

#include "common.h"
#include "dir_scan.h"
#include "fingerprint.h"
#include "duplicate.h"

static void print_usage(const char *prog_name) {
    printf("Disk Duplicate File Finder\n");
    printf("Usage: %s [OPTIONS] <directory>\n\n", prog_name);
    printf("Options:\n");
    printf("  -h, --help                Show this help message\n");
    printf("  -e, --exclude <pattern>   Exclude files matching pattern (e.g., *.tmp, *.bak)\n");
    printf("  -d, --exclude-dir <name>  Exclude directory name (e.g., .git, node_modules)\n");
    printf("  -s, --min-size <size>     Minimum file size to check (e.g., 1K, 2M, 1G)\n");
    printf("  -S, --max-size <size>     Maximum file size to check\n");
    printf("  -f, --follow-symlinks     Follow symbolic links (default: skip)\n");
    printf("\n");
    printf("Examples:\n");
    printf("  %s /home/user/Documents\n", prog_name);
    printf("  %s -d .git -d node_modules -e *.tmp /home/user\n", prog_name);
    printf("  %s -s 1K -S 100M /data\n", prog_name);
}

static off_t parse_size(const char *str) {
    if (!str || *str == '\0') return 0;
    
    off_t value = 0;
    char suffix = '\0';
    int scanned = sscanf(str, "%" SCNu64 "%c", (uint64_t *)&value, &suffix);
    
    if (scanned < 1) return 0;
    
    if (suffix == 'K' || suffix == 'k') {
        value *= 1024;
    } else if (suffix == 'M' || suffix == 'm') {
        value *= 1024 * 1024;
    } else if (suffix == 'G' || suffix == 'g') {
        value *= 1024 * 1024 * 1024;
    }
    
    return value;
}

int main(int argc, char *argv[]) {
    const char *prog_name = argv[0];
    ScanConfig config;
    scan_config_init(&config);
    
    static struct option long_options[] = {
        {"help",           no_argument,       0, 'h'},
        {"exclude",        required_argument, 0, 'e'},
        {"exclude-dir",    required_argument, 0, 'd'},
        {"min-size",       required_argument, 0, 's'},
        {"max-size",       required_argument, 0, 'S'},
        {"follow-symlinks", no_argument,      0, 'f'},
        {0, 0, 0, 0}
    };
    
    int opt;
    int option_index = 0;
    
    while ((opt = getopt_long(argc, argv, "he:d:s:S:f", long_options, &option_index)) != -1) {
        switch (opt) {
            case 'h':
                print_usage(prog_name);
                scan_config_free(&config);
                return 0;
            case 'e':
                pattern_list_append(&config.exclude_patterns, optarg);
                break;
            case 'd':
                pattern_list_append(&config.exclude_dirs, optarg);
                break;
            case 's':
                config.min_size = parse_size(optarg);
                break;
            case 'S':
                config.max_size = parse_size(optarg);
                break;
            case 'f':
                config.follow_symlinks = true;
                break;
            case '?':
                scan_config_free(&config);
                return 1;
            default:
                break;
        }
    }
    
    if (optind >= argc) {
        fprintf(stderr, "Error: No directory specified\n\n");
        print_usage(prog_name);
        scan_config_free(&config);
        return 1;
    }
    
    const char *dir_path = argv[optind];
    
    printf("Disk Duplicate File Finder\n");
    printf("Scanning: %s\n", dir_path);
    printf("============================================================\n");
    
    printf("Scanning directory...\n");
    FileList *files = scan_directory(dir_path, &config);
    if (!files) {
        fprintf(stderr, "Error: Failed to scan directory '%s'\n", dir_path);
        scan_config_free(&config);
        return 1;
    }
    
    printf("Found %zu files to analyze\n", files->count);
    
    if (files->count < 2) {
        printf("Not enough files to compare. Need at least 2 files.\n");
        file_list_free(files);
        scan_config_free(&config);
        return 0;
    }
    
    printf("Finding duplicate files...\n");
    DuplicateReport *report = find_duplicates(files);
    if (!report) {
        fprintf(stderr, "Error: Failed to find duplicates\n");
        file_list_free(files);
        scan_config_free(&config);
        return 1;
    }
    
    print_duplicate_report(report);
    
    duplicate_report_free(report);
    file_list_free(files);
    scan_config_free(&config);
    
    return 0;
}
