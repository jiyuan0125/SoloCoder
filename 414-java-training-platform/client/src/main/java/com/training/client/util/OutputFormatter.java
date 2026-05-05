package com.training.client.util;

import com.training.common.dto.response.CourseDTO;
import com.training.common.dto.response.CreditDetailDTO;
import com.training.common.dto.response.CreditSummaryDTO;
import com.training.common.dto.response.DepartmentDTO;
import com.training.common.dto.response.DepartmentStatisticsDTO;
import com.training.common.dto.response.EmployeeDTO;
import com.training.common.dto.response.InstructorDTO;
import com.training.common.dto.response.InstructorEvaluationSummaryDTO;
import com.training.common.dto.response.RegistrationDTO;
import com.training.common.dto.response.StatisticsDTO;
import com.training.common.dto.response.StudentEvaluationDTO;
import com.training.common.enums.CourseType;

import java.time.format.DateTimeFormatter;
import java.util.List;

public class OutputFormatter {

    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    public static String formatCourse(CourseDTO course) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append(String.format("课程ID: %s\n", course.getId()));
        sb.append(String.format("课程名称: %s\n", course.getName()));
        sb.append(String.format("课程描述: %s\n", course.getDescription() != null ? course.getDescription() : "无"));
        sb.append(String.format("讲师: %s (%s)\n", 
                course.getInstructorName() != null ? course.getInstructorName() : "未知", 
                course.getInstructorId()));
        sb.append(String.format("时长: %d 分钟\n", course.getDurationMinutes()));
        sb.append(String.format("名额: %d / %d\n", course.getCurrentEnrollment(), course.getMaxCapacity()));
        sb.append(String.format("课程类型: %s\n", 
                course.getCourseType() == CourseType.REQUIRED ? "必修课" : "选修课"));
        sb.append(String.format("课程状态: %s\n", course.getStatus()));
        sb.append(String.format("学分: %d\n", course.getCredits()));
        if (course.getFee() != null) {
            sb.append(String.format("费用: %s\n", course.getFee().toString()));
        }
        if (course.getStartTime() != null) {
            sb.append(String.format("开始时间: %s\n", course.getStartTime().format(DATE_FORMATTER)));
        }
        if (course.getEndTime() != null) {
            sb.append(String.format("结束时间: %s\n", course.getEndTime().format(DATE_FORMATTER)));
        }
        if (course.getRegistrationDeadline() != null) {
            sb.append(String.format("报名截止: %s\n", course.getRegistrationDeadline().format(DATE_FORMATTER)));
        }
        if (course.getRequiredDepartments() != null && !course.getRequiredDepartments().isEmpty()) {
            sb.append(String.format("指定部门: %s\n", String.join(", ", course.getRequiredDepartments())));
        }
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatCourses(List<CourseDTO> courses) {
        StringBuilder sb = new StringBuilder();
        sb.append(String.format("共找到 %d 门课程\n", courses.size()));
        sb.append("-".repeat(80)).append("\n");
        sb.append(String.format("%-36s %-20s %-10s %-10s %s\n", 
                "课程ID", "名称", "类型", "状态", "报名情况"));
        sb.append("-".repeat(80)).append("\n");
        for (CourseDTO course : courses) {
            String type = course.getCourseType() == CourseType.REQUIRED ? "必修" : "选修";
            String status = course.getStatus().toString();
            String enrollment = String.format("%d/%d", course.getCurrentEnrollment(), course.getMaxCapacity());
            sb.append(String.format("%-36s %-20s %-10s %-10s %s\n", 
                    course.getId(), 
                    truncate(course.getName(), 18), 
                    type, 
                    status,
                    enrollment));
        }
        sb.append("-".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatRegistration(RegistrationDTO reg) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append(String.format("报名ID: %s\n", reg.getId()));
        sb.append(String.format("员工: %s (%s)\n", reg.getEmployeeName(), reg.getEmployeeId()));
        sb.append(String.format("部门: %s (%s)\n", 
                reg.getDepartmentName() != null ? reg.getDepartmentName() : "未知", 
                reg.getDepartmentId()));
        sb.append(String.format("课程: %s (%s)\n", reg.getCourseName(), reg.getCourseId()));
        sb.append(String.format("讲师: %s\n", reg.getInstructorName() != null ? reg.getInstructorName() : "未知"));
        sb.append(String.format("报名状态: %s\n", reg.getStatus()));
        if (reg.getScore() != null) {
            sb.append(String.format("分数: %d (%s)\n", reg.getScore(), reg.isPassed() ? "通过" : "未通过"));
        }
        if (reg.getScoreComment() != null) {
            sb.append(String.format("评语: %s\n", reg.getScoreComment()));
        }
        if (reg.getRating() != null) {
            sb.append(String.format("评价评分: %d 星\n", reg.getRating()));
        }
        if (reg.getEvaluationComment() != null) {
            sb.append(String.format("评价内容: %s\n", reg.getEvaluationComment()));
        }
        if (reg.getCancelReason() != null) {
            sb.append(String.format("取消原因: %s\n", reg.getCancelReason()));
        }
        if (reg.getRegisteredAt() != null) {
            sb.append(String.format("报名时间: %s\n", reg.getRegisteredAt().format(DATE_FORMATTER)));
        }
        if (reg.getCompletedAt() != null) {
            sb.append(String.format("完成时间: %s\n", reg.getCompletedAt().format(DATE_FORMATTER)));
        }
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatRegistrations(List<RegistrationDTO> registrations) {
        StringBuilder sb = new StringBuilder();
        sb.append(String.format("共找到 %d 条报名记录\n", registrations.size()));
        sb.append("-".repeat(80)).append("\n");
        sb.append(String.format("%-36s %-15s %-15s %-12s %s\n", 
                "报名ID", "员工", "课程", "状态", "分数"));
        sb.append("-".repeat(80)).append("\n");
        for (RegistrationDTO reg : registrations) {
            String score = reg.getScore() != null ? 
                    String.format("%d(%s)", reg.getScore(), reg.isPassed() ? "通过" : "未通过") : "未打分";
            sb.append(String.format("%-36s %-15s %-15s %-12s %s\n", 
                    reg.getId(),
                    truncate(reg.getEmployeeName() != null ? reg.getEmployeeName() : "未知", 13),
                    truncate(reg.getCourseName() != null ? reg.getCourseName() : "未知", 13),
                    reg.getStatus().toString(),
                    score));
        }
        sb.append("-".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatEmployee(EmployeeDTO employee) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append(String.format("员工ID: %s\n", employee.getId()));
        sb.append(String.format("姓名: %s\n", employee.getName()));
        sb.append(String.format("部门: %s (%s)\n", 
                employee.getDepartmentName() != null ? employee.getDepartmentName() : "未知", 
                employee.getDepartmentId() != null ? employee.getDepartmentId() : "未分配"));
        sb.append(String.format("邮箱: %s\n", employee.getEmail() != null ? employee.getEmail() : "无"));
        if (employee.getSalary() != null) {
            sb.append(String.format("薪资: %s\n", employee.getSalary().toString()));
        }
        sb.append(String.format("状态: %s\n", employee.isActive() ? "在职" : "离职"));
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatEmployees(List<EmployeeDTO> employees) {
        StringBuilder sb = new StringBuilder();
        sb.append(String.format("共找到 %d 名员工\n", employees.size()));
        sb.append("-".repeat(80)).append("\n");
        sb.append(String.format("%-36s %-15s %-20s %s\n", 
                "员工ID", "姓名", "部门", "状态"));
        sb.append("-".repeat(80)).append("\n");
        for (EmployeeDTO emp : employees) {
            String deptName = emp.getDepartmentName() != null ? emp.getDepartmentName() : "未分配";
            sb.append(String.format("%-36s %-15s %-20s %s\n", 
                    emp.getId(),
                    emp.getName(),
                    truncate(deptName, 18),
                    emp.isActive() ? "在职" : "离职"));
        }
        sb.append("-".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatDepartment(DepartmentDTO dept) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append(String.format("部门ID: %s\n", dept.getId()));
        sb.append(String.format("部门名称: %s\n", dept.getName()));
        if (dept.getManagerId() != null) {
            sb.append(String.format("经理ID: %s\n", dept.getManagerId()));
        }
        sb.append(String.format("员工数量: %d\n", dept.getEmployeeCount()));
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatDepartments(List<DepartmentDTO> departments) {
        StringBuilder sb = new StringBuilder();
        sb.append(String.format("共找到 %d 个部门\n", departments.size()));
        sb.append("-".repeat(80)).append("\n");
        sb.append(String.format("%-36s %-25s %s\n", 
                "部门ID", "名称", "员工数量"));
        sb.append("-".repeat(80)).append("\n");
        for (DepartmentDTO dept : departments) {
            sb.append(String.format("%-36s %-25s %d\n", 
                    dept.getId(),
                    truncate(dept.getName(), 23),
                    dept.getEmployeeCount()));
        }
        sb.append("-".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatInstructor(InstructorDTO instructor) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append(String.format("讲师ID: %s\n", instructor.getId()));
        sb.append(String.format("姓名: %s\n", instructor.getName()));
        sb.append(String.format("邮箱: %s\n", instructor.getEmail() != null ? instructor.getEmail() : "无"));
        if (instructor.getSpecialties() != null && !instructor.getSpecialties().isEmpty()) {
            sb.append(String.format("专长: %s\n", String.join(", ", instructor.getSpecialties())));
        }
        if (instructor.getHourlyRate() != null) {
            sb.append(String.format("时薪: %s\n", instructor.getHourlyRate().toString()));
        }
        sb.append(String.format("状态: %s\n", instructor.isActive() ? "活跃" : "不活跃"));
        sb.append(String.format("授课数量: %d\n", instructor.getCourseCount()));
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatInstructors(List<InstructorDTO> instructors) {
        StringBuilder sb = new StringBuilder();
        sb.append(String.format("共找到 %d 名讲师\n", instructors.size()));
        sb.append("-".repeat(80)).append("\n");
        sb.append(String.format("%-36s %-15s %-25s %s\n", 
                "讲师ID", "姓名", "专长", "状态"));
        sb.append("-".repeat(80)).append("\n");
        for (InstructorDTO inst : instructors) {
            String specialties = inst.getSpecialties() != null && !inst.getSpecialties().isEmpty()
                    ? String.join(", ", inst.getSpecialties())
                    : "无";
            sb.append(String.format("%-36s %-15s %-25s %s\n", 
                    inst.getId(),
                    inst.getName(),
                    truncate(specialties, 23),
                    inst.isActive() ? "活跃" : "不活跃"));
        }
        sb.append("-".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatStatistics(StatisticsDTO stats) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append("                            统 计 概 览\n");
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append("\n【课程统计】\n");
        sb.append(String.format("  总课程数: %d\n", stats.getTotalCourses()));
        sb.append(String.format("  已发布: %d | 进行中: %d | 已完成: %d\n", 
                stats.getPublishedCourses(), stats.getOngoingCourses(), stats.getCompletedCourses()));
        sb.append(String.format("  必修课: %d | 选修课: %d\n", 
                stats.getRequiredCourses(), stats.getElectiveCourses()));
        sb.append("\n【员工统计】\n");
        sb.append(String.format("  总员工数: %d\n", stats.getTotalEmployees()));
        sb.append("\n【报名统计】\n");
        sb.append(String.format("  总报名数: %d\n", stats.getTotalRegistrations()));
        sb.append(String.format("  已完成: %d | 通过: %d | 未通过: %d\n", 
                stats.getCompletedRegistrations(), stats.getPassedRegistrations(), stats.getFailedRegistrations()));
        if (stats.getCompletedRegistrations() > 0) {
            sb.append(String.format("  通过率: %.2f%%\n", stats.getPassRate()));
        }
        
        if (stats.getDepartmentStatistics() != null && !stats.getDepartmentStatistics().isEmpty()) {
            sb.append("\n【各部门统计】\n");
            sb.append(String.format("%-15s %-10s %-10s %-10s %s\n", 
                    "部门", "员工数", "报名数", "通过数", "学分"));
            sb.append("-".repeat(80)).append("\n");
            for (DepartmentStatisticsDTO deptStats : stats.getDepartmentStatistics()) {
                sb.append(String.format("%-15s %-10d %-10d %-10d %d\n", 
                        truncate(deptStats.getDepartmentName(), 13),
                        deptStats.getEmployeeCount(),
                        deptStats.getRegisteredCount(),
                        deptStats.getPassedCount(),
                        deptStats.getTotalCreditsEarned()));
            }
        }
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatCreditSummary(CreditSummaryDTO summary) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append("                            学 分 概 览\n");
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append(String.format("员工: %s (%s)\n", summary.getEmployeeName(), summary.getEmployeeId()));
        sb.append(String.format("部门: %s\n", summary.getDepartmentName() != null ? summary.getDepartmentName() : "未知"));
        sb.append(String.format("年度: %d\n", summary.getYear()));
        sb.append("-".repeat(80)).append("\n");
        sb.append("\n【必修课学分】\n");
        sb.append(String.format("  应修总学分: %d | 已获得学分: %d\n", 
                summary.getTotalRequiredCredits(), summary.getEarnedRequiredCredits()));
        sb.append("\n【选修课学分】\n");
        sb.append(String.format("  已修总学分: %d | 已获得学分: %d\n", 
                summary.getTotalElectiveCredits(), summary.getEarnedElectiveCredits()));
        sb.append(String.format("  优秀证书资格: %s\n", 
                summary.isEligibleForExcellentCertificate() ? "已满足 (选修课学分>=20)" : "未满足"));
        sb.append("\n【总学分】\n");
        sb.append(String.format("  累计获得学分: %d\n", summary.getTotalCredits()));

        if (summary.getCreditDetails() != null && !summary.getCreditDetails().isEmpty()) {
            sb.append("\n【学分明细】\n");
            sb.append(String.format("%-20s %-8s %-8s %-6s %s\n", 
                    "课程名称", "类型", "学分", "分数", "完成时间"));
            sb.append("-".repeat(80)).append("\n");
            for (CreditDetailDTO detail : summary.getCreditDetails()) {
                String type = detail.getCourseType() == CourseType.REQUIRED ? "必修" : "选修";
                String score = detail.getScore() != null ? detail.getScore().toString() : "-";
                String completedAt = detail.getCompletedAt() != null 
                        ? detail.getCompletedAt().format(DateTimeFormatter.ofPattern("yyyy-MM-dd"))
                        : "-";
                sb.append(String.format("%-20s %-8s %-8d %-6s %s\n", 
                        truncate(detail.getCourseName(), 18),
                        type,
                        detail.getCredits(),
                        score,
                        completedAt));
            }
        }
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    public static String formatInstructorEvaluations(InstructorEvaluationSummaryDTO summary) {
        StringBuilder sb = new StringBuilder();
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append("                        讲 师 评 价 概 览\n");
        sb.append("=").append("=".repeat(80)).append("\n");
        sb.append(String.format("讲师: %s (%s)\n", summary.getInstructorName(), summary.getInstructorId()));
        sb.append(String.format("授课总数: %d | 评价总数: %d | 平均评分: %.2f\n", 
                summary.getTotalCourses(), summary.getTotalEvaluations(), summary.getAverageRating()));

        if (summary.getCourseEvaluations() != null && !summary.getCourseEvaluations().isEmpty()) {
            for (var courseEval : summary.getCourseEvaluations()) {
                sb.append("\n").append("-".repeat(80)).append("\n");
                sb.append(String.format("课程: %s (%s)\n", courseEval.getCourseName(), courseEval.getCourseId()));
                sb.append(String.format("报名人数: %d | 评价数量: %d | 平均评分: %.2f\n", 
                        courseEval.getTotalEnrollments(), 
                        courseEval.getEvaluationCount(), 
                        courseEval.getAverageRating()));

                if (courseEval.getEvaluations() != null && !courseEval.getEvaluations().isEmpty()) {
                    sb.append("\n学员评价:\n");
                    for (StudentEvaluationDTO eval : courseEval.getEvaluations()) {
                        sb.append(String.format("\n  学员: %s\n", eval.getEmployeeName()));
                        sb.append(String.format("  评分: %d 星\n", eval.getRating()));
                        if (eval.getComment() != null) {
                            sb.append(String.format("  评价: %s\n", eval.getComment()));
                        }
                        if (eval.getEvaluatedAt() != null) {
                            sb.append(String.format("  时间: %s\n", eval.getEvaluatedAt().format(DATE_FORMATTER)));
                        }
                    }
                }
            }
        }
        sb.append("=").append("=".repeat(80)).append("\n");
        return sb.toString();
    }

    private static String truncate(String str, int maxLength) {
        if (str == null) return "";
        if (str.length() <= maxLength) return str;
        return str.substring(0, maxLength - 1) + "…";
    }
}
