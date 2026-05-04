#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include "invite_service.h"

#define DEFAULT_STORAGE_FILE "invite_codes.dat"

static void print_usage(const char* prog_name) {
    printf("邀请码生成和校验工具\n\n");
    printf("用法: %s <命令> [选项]\n\n", prog_name);
    printf("命令:\n");
    printf("  generate <user_id>      为指定用户生成单个邀请码\n");
    printf("  batch <user_id> <count> 为指定用户批量生成邀请码\n");
    printf("  verify <code>           校验邀请码有效性\n");
    printf("  use <code>              使用邀请码(标记为已使用)\n");
    printf("  revoke <code>           作废邀请码\n");
    printf("  status                  显示系统状态和容量信息\n");
    printf("  help                    显示此帮助信息\n\n");
    printf("示例:\n");
    printf("  %s generate user_123\n", prog_name);
    printf("  %s batch user_123 10\n", prog_name);
    printf("  %s verify ABC123\n", prog_name);
    printf("  %s use ABC123\n", prog_name);
    printf("  %s revoke ABC123\n", prog_name);
}

static const char* status_to_string(CodeStatus status) {
    switch (status) {
        case CODE_STATUS_VALID: return "有效";
        case CODE_STATUS_USED: return "已使用";
        case CODE_STATUS_REVOKED: return "已作废";
        default: return "未知";
    }
}

static void format_time(time_t t, char* buf, size_t buf_size) {
    if (t == 0) {
        snprintf(buf, buf_size, "N/A");
        return;
    }
    struct tm* tm_info = localtime(&t);
    strftime(buf, buf_size, "%Y-%m-%d %H:%M:%S", tm_info);
}

static int cmd_generate(InviteService* service, int argc, char* argv[]) {
    if (argc < 3) {
        fprintf(stderr, "错误: 缺少 user_id 参数\n");
        return 1;
    }
    
    const char* user_id = argv[2];
    char code[INVITE_CODE_LENGTH + 1];
    
    ReturnCode rc = invite_service_generate(service, user_id, code, sizeof(code));
    if (rc != RC_SUCCESS) {
        fprintf(stderr, "生成邀请码失败: ");
        switch (rc) {
            case RC_ERROR_INVALID_PARAM: fprintf(stderr, "无效参数\n"); break;
            case RC_ERROR_CAPACITY_LIMIT: fprintf(stderr, "邀请码容量已耗尽\n"); break;
            case RC_ERROR_DUPLICATE_CODE: fprintf(stderr, "邀请码重复(请重试)\n"); break;
            default: fprintf(stderr, "错误码 %d\n", rc);
        }
        return 1;
    }
    
    rc = invite_service_save(service);
    if (rc != RC_SUCCESS) {
        fprintf(stderr, "警告: 保存到文件失败\n");
    }
    
    printf("邀请码: %s\n", code);
    printf("用户ID: %s\n", user_id);
    return 0;
}

static int cmd_batch(InviteService* service, int argc, char* argv[]) {
    if (argc < 4) {
        fprintf(stderr, "错误: 缺少参数\n");
        fprintf(stderr, "用法: %s batch <user_id> <count>\n", argv[0]);
        return 1;
    }
    
    const char* user_id = argv[2];
    int count = atoi(argv[3]);
    
    if (count <= 0 || count > 1000) {
        fprintf(stderr, "错误: 数量必须在 1-1000 之间\n");
        return 1;
    }
    
    char** codes = NULL;
    int actual_count = 0;
    
    ReturnCode rc = invite_service_generate_batch(service, user_id, count, &codes, &actual_count);
    if (rc != RC_SUCCESS && rc != RC_ERROR_CAPACITY_LIMIT) {
        fprintf(stderr, "批量生成邀请码失败: 错误码 %d\n", rc);
        return 1;
    }
    
    if (actual_count == 0) {
        fprintf(stderr, "未能生成任何邀请码，容量可能已耗尽\n");
        return 1;
    }
    
    rc = invite_service_save(service);
    if (rc != RC_SUCCESS) {
        fprintf(stderr, "警告: 保存到文件失败\n");
    }
    
    printf("成功生成 %d 个邀请码 (用户: %s):\n", actual_count, user_id);
    for (int i = 0; i < actual_count; i++) {
        printf("%d. %s\n", i + 1, codes[i]);
    }
    
    if (rc == RC_ERROR_CAPACITY_LIMIT) {
        printf("\n警告: 邀请码容量不足，仅生成了 %d 个\n", actual_count);
    }
    
    invite_service_free_codes(codes, actual_count);
    return 0;
}

static int cmd_verify(InviteService* service, int argc, char* argv[]) {
    if (argc < 3) {
        fprintf(stderr, "错误: 缺少邀请码参数\n");
        return 1;
    }
    
    const char* code = argv[2];
    InviteCodeInfo info;
    
    ReturnCode rc = invite_service_verify(service, code, &info);
    if (rc != RC_SUCCESS) {
        switch (rc) {
            case RC_ERROR_INVALID_CODE:
                printf("%s\n", INVALID_CODE_ERROR);
                break;
            case RC_ERROR_NOT_FOUND:
                printf("%s\n", INVALID_CODE_ERROR);
                break;
            default:
                fprintf(stderr, "校验失败: 错误码 %d\n", rc);
        }
        return 1;
    }
    
    char create_time[64], use_time[64];
    format_time(info.create_time, create_time, sizeof(create_time));
    format_time(info.use_time, use_time, sizeof(use_time));
    
    printf("邀请码有效\n");
    printf("邀请码: %s\n", info.code);
    printf("邀请人: %s\n", info.user_id);
    printf("状态: %s\n", status_to_string(info.status));
    printf("创建时间: %s\n", create_time);
    return 0;
}

static int cmd_use(InviteService* service, int argc, char* argv[]) {
    if (argc < 3) {
        fprintf(stderr, "错误: 缺少邀请码参数\n");
        return 1;
    }
    
    const char* code = argv[2];
    InviteCodeInfo info;
    
    ReturnCode rc = invite_service_verify(service, code, &info);
    if (rc != RC_SUCCESS) {
        printf("%s\n", INVALID_CODE_ERROR);
        return 1;
    }
    
    rc = invite_service_use(service, code);
    if (rc != RC_SUCCESS) {
        switch (rc) {
            case RC_ERROR_NOT_FOUND:
                printf("%s\n", INVALID_CODE_ERROR);
                break;
            case RC_ERROR_INVALID_CODE:
                printf("%s\n", INVALID_CODE_ERROR);
                break;
            default:
                fprintf(stderr, "使用邀请码失败: 错误码 %d\n", rc);
        }
        return 1;
    }
    
    rc = invite_service_save(service);
    if (rc != RC_SUCCESS) {
        fprintf(stderr, "警告: 保存到文件失败\n");
    }
    
    printf("邀请码已使用\n");
    printf("邀请码: %s\n", code);
    printf("邀请人: %s\n", info.user_id);
    return 0;
}

static int cmd_revoke(InviteService* service, int argc, char* argv[]) {
    if (argc < 3) {
        fprintf(stderr, "错误: 缺少邀请码参数\n");
        return 1;
    }
    
    const char* code = argv[2];
    
    ReturnCode rc = invite_service_revoke(service, code);
    if (rc != RC_SUCCESS) {
        switch (rc) {
            case RC_ERROR_NOT_FOUND:
                printf("%s\n", INVALID_CODE_ERROR);
                break;
            case RC_ERROR_INVALID_CODE:
                printf("%s\n", INVALID_CODE_ERROR);
                break;
            default:
                fprintf(stderr, "作废邀请码失败: 错误码 %d\n", rc);
        }
        return 1;
    }
    
    rc = invite_service_save(service);
    if (rc != RC_SUCCESS) {
        fprintf(stderr, "警告: 保存到文件失败\n");
    }
    
    printf("邀请码已作废: %s\n", code);
    return 0;
}

static int cmd_status(InviteService* service) {
    CapacityInfo cap = invite_service_get_capacity(service);
    uint64_t count = invite_service_get_total_count(service);
    
    printf("=== 邀请码系统状态 ===\n");
    printf("邀请码长度: %d 位\n", INVITE_CODE_LENGTH);
    printf("字符集大小: %d 个字符\n", BASE58_CHAR_COUNT);
    printf("\n");
    printf("总可能组合数: %llu\n", (unsigned long long)cap.total_possible);
    printf("已生成邀请码数: %llu\n", (unsigned long long)count);
    printf("使用率: %.6f%%\n", cap.usage_percent);
    printf("\n");
    
    if (cap.is_critical) {
        printf("!!! 警告: 邀请码容量已耗尽 !!!\n");
    } else if (cap.is_warning) {
        printf("!!! 警告: 邀请码使用率已超过 %.0f%% !!!\n", 
               CAPACITY_WARNING_THRESHOLD * 100.0);
    }
    
    return 0;
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        print_usage(argv[0]);
        return 1;
    }
    
    const char* cmd = argv[1];
    
    if (strcmp(cmd, "help") == 0 || strcmp(cmd, "-h") == 0 || strcmp(cmd, "--help") == 0) {
        print_usage(argv[0]);
        return 0;
    }
    
    InviteService* service = invite_service_create(DEFAULT_STORAGE_FILE);
    if (!service) {
        fprintf(stderr, "错误: 无法初始化邀请码服务\n");
        return 1;
    }
    
    ReturnCode rc = invite_service_load(service);
    if (rc != RC_SUCCESS && storage_file_exists(DEFAULT_STORAGE_FILE)) {
        fprintf(stderr, "警告: 加载数据文件失败，将创建新的数据库\n");
    }
    
    int ret = 0;
    
    if (strcmp(cmd, "generate") == 0) {
        ret = cmd_generate(service, argc, argv);
    } else if (strcmp(cmd, "batch") == 0) {
        ret = cmd_batch(service, argc, argv);
    } else if (strcmp(cmd, "verify") == 0) {
        ret = cmd_verify(service, argc, argv);
    } else if (strcmp(cmd, "use") == 0) {
        ret = cmd_use(service, argc, argv);
    } else if (strcmp(cmd, "revoke") == 0) {
        ret = cmd_revoke(service, argc, argv);
    } else if (strcmp(cmd, "status") == 0) {
        ret = cmd_status(service);
    } else {
        fprintf(stderr, "错误: 未知命令 '%s'\n", cmd);
        fprintf(stderr, "使用 '%s help' 查看帮助\n", argv[0]);
        ret = 1;
    }
    
    invite_service_destroy(service);
    return ret;
}
