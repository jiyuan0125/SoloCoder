package com.example.client.output;

import com.example.common.dto.*;
import java.util.List;
import java.util.Map;

public class OutputFormatter {

    public void printEmployee(EmployeeDTO employee) {
        System.out.println("========================================");
        System.out.println("员工信息");
        System.out.println("========================================");
        System.out.printf("ID: %s%n", employee.getId());
        System.out.printf("姓名: %s%n", employee.getName());
        System.out.printf("邮箱: %s%n", employee.getEmail() != null ? employee.getEmail() : "-");
        System.out.printf("岗位: %s%n", employee.getPositionType() != null 
                ? employee.getPositionType().getDescription() : "-");
        System.out.printf("部门: %s%n", employee.getDepartment() != null ? employee.getDepartment() : "-");
        System.out.printf("入职日期: %s%n", employee.getOnboardingDate() != null 
                ? employee.getOnboardingDate() : "-");
        System.out.printf("入职状态: %s%n", employee.getStatus() != null 
                ? employee.getStatus().getDescription() : "-");
        System.out.println();

        if (employee.getChecklistItems() != null && !employee.getChecklistItems().isEmpty()) {
            System.out.println("清单事项:");
            System.out.println("----------------------------------------");
            for (ChecklistItemDTO item : employee.getChecklistItems()) {
                printChecklistItemShort(item);
            }
        }
        System.out.println("========================================");
    }

    public void printEmployeeList(List<EmployeeDTO> employees) {
        System.out.println("========================================");
        System.out.println("员工列表");
        System.out.println("========================================");
        System.out.printf("共 %d 条记录%n%n", employees.size());

        if (employees.isEmpty()) {
            System.out.println("暂无员工记录");
        } else {
            System.out.printf("%-36s %-12s %-12s %-16s %s%n", 
                    "ID", "姓名", "岗位", "入职日期", "状态");
            System.out.println("-----------------------------------------------------------------------------");
            for (EmployeeDTO employee : employees) {
                System.out.printf("%-36s %-12s %-12s %-16s %s%n",
                        employee.getId(),
                        employee.getName(),
                        employee.getPositionType() != null ? employee.getPositionType().getDescription() : "-",
                        employee.getOnboardingDate() != null ? employee.getOnboardingDate().toString() : "-",
                        employee.getStatus() != null ? employee.getStatus().getDescription() : "-"
                );
            }
        }
        System.out.println("========================================");
    }

    public void printChecklistItem(ChecklistItemDTO item) {
        System.out.println("========================================");
        System.out.println("清单事项详情");
        System.out.println("========================================");
        System.out.printf("ID: %s%n", item.getId());
        System.out.printf("名称: %s%n", item.getName());
        System.out.printf("描述: %s%n", item.getDescription() != null ? item.getDescription() : "-");
        System.out.printf("负责人: %s%n", item.getResponsiblePerson() != null ? item.getResponsiblePerson() : "-");
        System.out.printf("部门: %s%n", item.getDepartment() != null ? item.getDepartment() : "-");
        System.out.printf("截止日期: %s%n", item.getDueDate() != null ? item.getDueDate() : "-");
        System.out.printf("完成日期: %s%n", item.getCompletedDate() != null ? item.getCompletedDate() : "-");
        System.out.printf("状态: %s%n", item.getStatus() != null ? item.getStatus().getDescription() : "-");
        System.out.printf("是否必填: %s%n", item.isRequired() ? "是" : "否");
        System.out.printf("入职当天必须完成: %s%n", item.isOnboardingDayRequired() ? "是" : "否");
        System.out.printf("入职前事项: %s%n", item.isPreOnboarding() ? "是" : "否");
        System.out.printf("是否已升级: %s%n", item.isEscalated() ? "是" : "否");
        if (item.isEscalated()) {
            System.out.printf("升级至: %s%n", item.getEscalatedTo() != null ? item.getEscalatedTo() : "-");
        }
        System.out.println("========================================");
    }

    private void printChecklistItemShort(ChecklistItemDTO item) {
        String statusIcon = switch (item.getStatus()) {
            case COMPLETED -> "[✓]";
            case OVERDUE -> "[!]";
            case ESCALATED -> "[↑]";
            case IN_PROGRESS -> "[→]";
            default -> "[ ]";
        };
        System.out.printf("%s %s (负责人: %s, 截止: %s)%n",
                statusIcon,
                item.getName(),
                item.getResponsiblePerson() != null ? item.getResponsiblePerson() : "-",
                item.getDueDate() != null ? item.getDueDate() : "-"
        );
    }

    public void printTemplate(ChecklistTemplateDTO template) {
        System.out.println("========================================");
        System.out.println("模板信息");
        System.out.println("========================================");
        System.out.printf("ID: %s%n", template.getId());
        System.out.printf("名称: %s%n", template.getName());
        System.out.printf("岗位类型: %s%n", template.getPositionType() != null 
                ? template.getPositionType().getDescription() : "-");
        System.out.printf("描述: %s%n", template.getDescription() != null ? template.getDescription() : "-");
        System.out.printf("是否默认: %s%n", template.isDefault() ? "是" : "否");
        System.out.println();

        if (template.getItems() != null && !template.getItems().isEmpty()) {
            System.out.println("模板事项:");
            System.out.println("----------------------------------------");
            for (TemplateItemDTO item : template.getItems()) {
                System.out.printf("- %s (相对入职日期: %s天, 负责人: %s)%n",
                        item.getName(),
                        item.getDaysRelativeToOnboarding() >= 0 ? "+" + item.getDaysRelativeToOnboarding() : item.getDaysRelativeToOnboarding(),
                        item.getResponsiblePerson() != null ? item.getResponsiblePerson() : "-"
                );
            }
        }
        System.out.println("========================================");
    }

    public void printTemplateList(List<ChecklistTemplateDTO> templates) {
        System.out.println("========================================");
        System.out.println("模板列表");
        System.out.println("========================================");
        System.out.printf("共 %d 条记录%n%n", templates.size());

        if (templates.isEmpty()) {
            System.out.println("暂无模板");
        } else {
            System.out.printf("%-36s %-20s %-12s %s%n", 
                    "ID", "名称", "岗位", "默认");
            System.out.println("---------------------------------------------------------------------------------");
            for (ChecklistTemplateDTO template : templates) {
                System.out.printf("%-36s %-20s %-12s %s%n",
                        template.getId(),
                        template.getName(),
                        template.getPositionType() != null ? template.getPositionType().getDescription() : "-",
                        template.isDefault() ? "是" : "否"
                );
            }
        }
        System.out.println("========================================");
    }

    public void printAlert(AlertDTO alert) {
        System.out.println("========================================");
        System.out.println("告警信息");
        System.out.println("========================================");
        System.out.printf("ID: %s%n", alert.getId());
        System.out.printf("员工ID: %s%n", alert.getEmployeeId() != null ? alert.getEmployeeId() : "-");
        System.out.printf("员工姓名: %s%n", alert.getEmployeeName() != null ? alert.getEmployeeName() : "-");
        System.out.printf("负责人: %s%n", alert.getResponsiblePerson() != null ? alert.getResponsiblePerson() : "-");
        System.out.printf("级别: %s%n", alert.getLevel() != null ? alert.getLevel().getDescription() : "-");
        System.out.printf("消息: %s%n", alert.getMessage() != null ? alert.getMessage() : "-");
        System.out.printf("逾期数量: %d%n", alert.getOverdueCount());
        System.out.printf("是否已读: %s%n", alert.isRead() ? "是" : "否");
        System.out.printf("创建时间: %s%n", alert.getCreatedAt() != null ? alert.getCreatedAt() : "-");
        System.out.println("========================================");
    }

    public void printAlertList(List<AlertDTO> alerts) {
        System.out.println("========================================");
        System.out.println("告警列表");
        System.out.println("========================================");
        System.out.printf("共 %d 条记录%n%n", alerts.size());

        if (alerts.isEmpty()) {
            System.out.println("暂无告警");
        } else {
            System.out.printf("%-36s %-12s %-10s %s%n", 
                    "ID", "负责人", "级别", "消息");
            System.out.println("---------------------------------------------------------------------------------");
            for (AlertDTO alert : alerts) {
                System.out.printf("%-36s %-12s %-10s %s%n",
                        alert.getId(),
                        alert.getResponsiblePerson() != null ? alert.getResponsiblePerson() : "-",
                        alert.getLevel() != null ? alert.getLevel().getDescription() : "-",
                        alert.getMessage() != null ? truncate(alert.getMessage(), 40) : "-"
                );
            }
        }
        System.out.println("========================================");
    }

    public void printStatistics(OnboardingStatisticsDTO stats) {
        System.out.println("========================================");
        System.out.println("入职进度统计");
        System.out.println("========================================");
        System.out.printf("员工总数: %d%n", stats.getTotalEmployees());
        System.out.printf("完成入职: %d%n", stats.getCompletedOnboarding());
        System.out.printf("入职不完整: %d%n", stats.getIncompleteOnboarding());
        System.out.printf("整体完成率: %.2f%%%n", stats.getCompletionRate());
        System.out.printf("平均完成天数: %.2f 天%n", stats.getAverageCompletionTime());
        System.out.println();

        if (stats.getItemsByStatus() != null && !stats.getItemsByStatus().isEmpty()) {
            System.out.println("事项状态统计:");
            for (Map.Entry<String, Long> entry : stats.getItemsByStatus().entrySet()) {
                System.out.printf("  %s: %d%n", entry.getKey(), entry.getValue());
            }
        }

        if (stats.getOverdueItemsByResponsible() != null && !stats.getOverdueItemsByResponsible().isEmpty()) {
            System.out.println();
            System.out.println("按负责人统计逾期事项:");
            for (Map.Entry<String, Long> entry : stats.getOverdueItemsByResponsible().entrySet()) {
                System.out.printf("  %s: %d 个%n", entry.getKey(), entry.getValue());
            }
        }

        if (stats.getEmployeesByPosition() != null && !stats.getEmployeesByPosition().isEmpty()) {
            System.out.println();
            System.out.println("按岗位统计员工:");
            for (Map.Entry<String, Long> entry : stats.getEmployeesByPosition().entrySet()) {
                System.out.printf("  %s: %d 人%n", entry.getKey(), entry.getValue());
            }
        }
        System.out.println("========================================");
    }

    public void printSuccess(String message) {
        System.out.println("[成功] " + message);
    }

    public void printError(String message) {
        System.err.println("[错误] " + message);
    }

    public void printHelp() {
        System.out.println("========================================");
        System.out.println("入职清单管理系统 - 命令行客户端");
        System.out.println("========================================");
        System.out.println();
        System.out.println("使用方法:");
        System.out.println("  java -jar client.jar <命令> [选项]");
        System.out.println();
        System.out.println("员工管理命令:");
        System.out.println("  employee-list                                    列出所有员工");
        System.out.println("  employee-get --id <员工ID>                      获取员工详情");
        System.out.println("  employee-create --name <姓名> --position <岗位> --onboarding-date <入职日期>");
        System.out.println("                                                   创建新员工");
        System.out.println("  employee-delete --id <员工ID>                   删除员工");
        System.out.println();
        System.out.println("清单事项命令:");
        System.out.println("  item-add --employee-id <员工ID> --name <事项名> --responsible <负责人> --due-date <截止日期>");
        System.out.println("                                                   添加新事项");
        System.out.println("  item-update --employee-id <员工ID> --item-id <事项ID> [选项]");
        System.out.println("                                                   更新事项");
        System.out.println();
        System.out.println("模板管理命令:");
        System.out.println("  template-list                                    列出所有模板");
        System.out.println("  template-get --id <模板ID>                      获取模板详情");
        System.out.println("  template-copy --source-id <源模板ID> --new-name <新名称>");
        System.out.println("                                                   复制模板");
        System.out.println();
        System.out.println("告警管理命令:");
        System.out.println("  alert-list                                       列出所有告警");
        System.out.println("  alert-unread                                     列出未读告警");
        System.out.println("  alert-get --id <告警ID>                         获取告警详情");
        System.out.println("  alert-read --id <告警ID>                        标记为已读");
        System.out.println();
        System.out.println("统计命令:");
        System.out.println("  stats                                            查看统计信息");
        System.out.println();
        System.out.println("岗位类型选项:");
        System.out.println("  DEVELOPER, SALES, HR, FINANCE, ADMIN, MARKETING");
        System.out.println();
        System.out.println("日期格式:");
        System.out.println("  YYYY-MM-DD (例如: 2024-01-15)");
        System.out.println();
        System.out.println("示例:");
        System.out.println("  java -jar client.jar employee-create --name 张三 --position DEVELOPER --onboarding-date 2024-01-15");
        System.out.println("  java -jar client.jar employee-list");
        System.out.println("  java -jar client.jar stats");
        System.out.println("========================================");
    }

    private String truncate(String str, int maxLength) {
        if (str == null) return "";
        return str.length() <= maxLength ? str : str.substring(0, maxLength) + "...";
    }
}
