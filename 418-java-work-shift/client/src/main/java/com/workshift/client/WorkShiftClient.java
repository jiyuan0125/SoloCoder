package com.workshift.client;

import com.workshift.client.command.CommandDispatcher;

public class WorkShiftClient {

    public static void main(String[] args) {
        String baseUrl = System.getProperty("workshift.server.url", "http://localhost:8080");
        
        if (args.length == 0) {
            printHelp();
            System.exit(0);
        }

        CommandDispatcher dispatcher = new CommandDispatcher(baseUrl);
        try {
            dispatcher.dispatch(args);
        } catch (Exception e) {
            System.err.println("错误: " + e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }

    private static void printHelp() {
        System.out.println("========================================");
        System.out.println("    工厂排班管理系统 CLI 客户端");
        System.out.println("========================================");
        System.out.println();
        System.out.println("用法: java -jar work-shift-client.jar <命令> [选项]");
        System.out.println();
        System.out.println("员工管理命令:");
        System.out.println("  employee list                              - 列出所有员工");
        System.out.println("  employee get <id>                          - 获取指定员工信息");
        System.out.println("  employee create <name> <department> <hourlyWage> - 创建员工");
        System.out.println("  employee update <id> [name=<name>] [dept=<dept>] [wage=<wage>] - 更新员工");
        System.out.println("  employee delete <id>                       - 删除员工");
        System.out.println();
        System.out.println("排班管理命令:");
        System.out.println("  shift list                                  - 列出所有排班");
        System.out.println("  shift list <employeeId>                     - 列出指定员工的排班");
        System.out.println("  shift range <startDate> <endDate>          - 列出指定日期范围的排班");
        System.out.println("  shift create <employeeId> <date> <shiftType> - 创建排班");
        System.out.println("  shift publish <shiftId>                     - 发布排班");
        System.out.println("  shift publish-batch <shiftId1,shiftId2...> - 批量发布排班");
        System.out.println("  shift delete <shiftId>                      - 删除排班");
        System.out.println();
        System.out.println("换班管理命令:");
        System.out.println("  swap list                                    - 列出所有换班申请");
        System.out.println("  swap list <employeeId>                       - 列出指定员工的换班申请");
        System.out.println("  swap get <swapId>                            - 获取换班申请详情");
        System.out.println("  swap create <requesterId> <targetId> <requesterDate> <targetDate> <signature> - 创建换班申请");
        System.out.println("  swap confirm <swapId> <targetEmployeeId> <confirmed> [signature] [rejectReason] - 确认换班");
        System.out.println();
        System.out.println("异议管理命令:");
        System.out.println("  objection list                               - 列出所有异议");
        System.out.println("  objection list <employeeId>                  - 列出指定员工的异议");
        System.out.println("  objection get <objectionId>                  - 获取异议详情");
        System.out.println("  objection create <shiftId> <employeeId> <reason> - 创建异议");
        System.out.println("  objection resolve <objectionId> <note>      - 解决异议");
        System.out.println("  objection dismiss <objectionId> <note>      - 驳回异议");
        System.out.println();
        System.out.println("统计报表命令:");
        System.out.println("  stats employee <employeeId> <year> <month>  - 生成员工月度统计");
        System.out.println("  stats all <year> <month>                     - 生成所有员工月度统计");
        System.out.println("  stats holiday add <date> <name>              - 添加法定节假日");
        System.out.println("  stats holiday check <date>                    - 检查日期是否为节假日");
        System.out.println();
        System.out.println("班次类型: MORNING(早班), AFTERNOON(中班), NIGHT(夜班)");
        System.out.println("日期格式: yyyy-MM-dd (例如: 2025-01-01)");
        System.out.println("========================================");
    }
}