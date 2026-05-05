package com.training.client;

import com.training.client.service.TrainingApiService;
import com.training.client.util.OutputFormatter;
import com.training.common.dto.request.ApproveCancellationRequest;
import com.training.common.dto.request.CancelRegistrationRequest;
import com.training.common.dto.request.CreateCourseRequest;
import com.training.common.dto.request.CreateDepartmentRequest;
import com.training.common.dto.request.CreateEmployeeRequest;
import com.training.common.dto.request.CreateInstructorRequest;
import com.training.common.dto.request.EvaluationRequest;
import com.training.common.dto.request.RegisterRequest;
import com.training.common.dto.request.ScoreRequest;
import com.training.common.dto.response.CourseDTO;
import com.training.common.dto.response.CreditSummaryDTO;
import com.training.common.dto.response.DepartmentDTO;
import com.training.common.dto.response.EmployeeDTO;
import com.training.common.dto.response.InstructorDTO;
import com.training.common.dto.response.InstructorEvaluationSummaryDTO;
import com.training.common.dto.response.RegistrationDTO;
import com.training.common.dto.response.StatisticsDTO;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Arrays;
import java.util.List;

public class TrainingClient {

    private static final String DEFAULT_BASE_URL = "http://localhost:8080";
    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm");
    
    private final TrainingApiService apiService;
    private final BufferedReader reader;

    public TrainingClient(String baseUrl) {
        this.apiService = new TrainingApiService(baseUrl);
        this.reader = new BufferedReader(new InputStreamReader(System.in));
    }

    public static void main(String[] args) {
        String baseUrl = System.getenv("TRAINING_API_URL");
        if (baseUrl == null || baseUrl.isEmpty()) {
            baseUrl = DEFAULT_BASE_URL;
        }

        TrainingClient client = new TrainingClient(baseUrl);

        if (args.length == 0) {
            client.printHelp();
            return;
        }

        try {
            client.executeCommand(args);
        } catch (Exception e) {
            System.err.println("错误: " + e.getMessage());
            System.exit(1);
        }
    }

    private void executeCommand(String[] args) throws Exception {
        String command = args[0].toLowerCase();

        switch (command) {
            case "help":
            case "-h":
            case "--help":
                printHelp();
                break;

            case "dept-list":
                listDepartments();
                break;
            case "dept-get":
                requireArgs(args, 2, "用法: dept-get <部门ID>");
                getDepartment(args[1]);
                break;
            case "dept-create":
                createDepartment(args);
                break;

            case "emp-list":
                listEmployees();
                break;
            case "emp-get":
                requireArgs(args, 2, "用法: emp-get <员工ID>");
                getEmployee(args[1]);
                break;
            case "emp-create":
                createEmployee(args);
                break;

            case "inst-list":
                listInstructors();
                break;
            case "inst-get":
                requireArgs(args, 2, "用法: inst-get <讲师ID>");
                getInstructor(args[1]);
                break;
            case "inst-create":
                createInstructor(args);
                break;
            case "inst-evaluations":
                requireArgs(args, 2, "用法: inst-evaluations <讲师ID>");
                getInstructorEvaluations(args[1]);
                break;
            case "inst-pending-cancellations":
                requireArgs(args, 2, "用法: inst-pending-cancellations <讲师ID>");
                getPendingCancellations(args[1]);
                break;

            case "course-list":
                boolean publishedOnly = args.length > 1 && args[1].equals("--published");
                listCourses(publishedOnly);
                break;
            case "course-get":
                requireArgs(args, 2, "用法: course-get <课程ID>");
                getCourse(args[1]);
                break;
            case "course-create":
                createCourse(args);
                break;

            case "register":
                requireArgs(args, 3, "用法: register <员工ID> <课程ID>");
                register(args[1], args[2]);
                break;
            case "cancel-registration":
                requireArgs(args, 2, "用法: cancel-registration <报名ID>");
                cancelRegistration(args[1]);
                break;
            case "approve-cancellation":
                approveCancellation(args);
                break;
            case "reg-list-by-emp":
                requireArgs(args, 2, "用法: reg-list-by-emp <员工ID>");
                listRegistrationsByEmployee(args[1]);
                break;
            case "reg-list-by-course":
                requireArgs(args, 2, "用法: reg-list-by-course <课程ID>");
                listRegistrationsByCourse(args[1]);
                break;
            case "reg-get":
                requireArgs(args, 2, "用法: reg-get <报名ID>");
                getRegistration(args[1]);
                break;

            case "score":
                submitScore(args);
                break;
            case "evaluate":
                submitEvaluation(args);
                break;

            case "stats":
                getStatistics();
                break;
            case "credits":
                getCredits(args);
                break;

            default:
                System.out.println("未知命令: " + command);
                printHelp();
        }
    }

    private void listDepartments() throws Exception {
        List<DepartmentDTO> departments = apiService.getAllDepartments();
        System.out.println(OutputFormatter.formatDepartments(departments));
    }

    private void getDepartment(String id) throws Exception {
        DepartmentDTO dept = apiService.getDepartment(id);
        System.out.println(OutputFormatter.formatDepartment(dept));
    }

    private void createDepartment(String[] args) throws Exception {
        CreateDepartmentRequest request = new CreateDepartmentRequest();
        
        if (args.length >= 2) {
            request.setName(args[1]);
        } else {
            System.out.print("部门名称: ");
            request.setName(reader.readLine());
        }
        
        System.out.print("经理ID (可选, 直接回车跳过): ");
        String managerId = reader.readLine();
        if (!managerId.isEmpty()) {
            request.setManagerId(managerId);
        }

        DepartmentDTO dept = apiService.createDepartment(request);
        System.out.println("部门创建成功:");
        System.out.println(OutputFormatter.formatDepartment(dept));
    }

    private void listEmployees() throws Exception {
        List<EmployeeDTO> employees = apiService.getAllEmployees();
        System.out.println(OutputFormatter.formatEmployees(employees));
    }

    private void getEmployee(String id) throws Exception {
        EmployeeDTO emp = apiService.getEmployee(id);
        System.out.println(OutputFormatter.formatEmployee(emp));
    }

    private void createEmployee(String[] args) throws Exception {
        CreateEmployeeRequest request = new CreateEmployeeRequest();

        System.out.print("员工姓名: ");
        request.setName(reader.readLine());

        System.out.print("部门ID: ");
        request.setDepartmentId(reader.readLine());

        System.out.print("邮箱: ");
        request.setEmail(reader.readLine());

        System.out.print("薪资 (使用BigDecimal格式): ");
        String salaryStr = reader.readLine();
        if (!salaryStr.isEmpty()) {
            request.setSalary(new BigDecimal(salaryStr));
        }

        EmployeeDTO emp = apiService.createEmployee(request);
        System.out.println("员工创建成功:");
        System.out.println(OutputFormatter.formatEmployee(emp));
    }

    private void listInstructors() throws Exception {
        List<InstructorDTO> instructors = apiService.getAllInstructors();
        System.out.println(OutputFormatter.formatInstructors(instructors));
    }

    private void getInstructor(String id) throws Exception {
        InstructorDTO inst = apiService.getInstructor(id);
        System.out.println(OutputFormatter.formatInstructor(inst));
    }

    private void createInstructor(String[] args) throws Exception {
        CreateInstructorRequest request = new CreateInstructorRequest();

        System.out.print("讲师姓名: ");
        request.setName(reader.readLine());

        System.out.print("邮箱: ");
        request.setEmail(reader.readLine());

        System.out.print("专长 (逗号分隔): ");
        String specialtiesStr = reader.readLine();
        if (!specialtiesStr.isEmpty()) {
            request.setSpecialties(Arrays.asList(specialtiesStr.split(",")));
        }

        System.out.print("时薪 (使用BigDecimal格式): ");
        String hourlyRateStr = reader.readLine();
        if (!hourlyRateStr.isEmpty()) {
            request.setHourlyRate(new BigDecimal(hourlyRateStr));
        }

        InstructorDTO inst = apiService.createInstructor(request);
        System.out.println("讲师创建成功:");
        System.out.println(OutputFormatter.formatInstructor(inst));
    }

    private void getInstructorEvaluations(String instructorId) throws Exception {
        InstructorEvaluationSummaryDTO summary = apiService.getInstructorEvaluations(instructorId);
        System.out.println(OutputFormatter.formatInstructorEvaluations(summary));
    }

    private void getPendingCancellations(String instructorId) throws Exception {
        List<RegistrationDTO> registrations = apiService.getPendingCancellations(instructorId);
        System.out.println(OutputFormatter.formatRegistrations(registrations));
    }

    private void listCourses(boolean publishedOnly) throws Exception {
        List<CourseDTO> courses = apiService.getAllCourses(publishedOnly);
        System.out.println(OutputFormatter.formatCourses(courses));
    }

    private void getCourse(String id) throws Exception {
        CourseDTO course = apiService.getCourse(id);
        System.out.println(OutputFormatter.formatCourse(course));
    }

    private void createCourse(String[] args) throws Exception {
        CreateCourseRequest request = new CreateCourseRequest();

        System.out.print("课程名称: ");
        request.setName(reader.readLine());

        System.out.print("课程描述: ");
        request.setDescription(reader.readLine());

        System.out.print("讲师ID: ");
        request.setInstructorId(reader.readLine());

        System.out.print("时长(分钟): ");
        String durationStr = reader.readLine();
        if (!durationStr.isEmpty()) {
            request.setDurationMinutes(Integer.parseInt(durationStr));
        }

        System.out.print("名额上限: ");
        String capacityStr = reader.readLine();
        if (!capacityStr.isEmpty()) {
            request.setMaxCapacity(Integer.parseInt(capacityStr));
        }

        System.out.print("开始时间 (格式: yyyy-MM-dd HH:mm): ");
        String startTimeStr = reader.readLine();
        if (!startTimeStr.isEmpty()) {
            request.setStartTime(LocalDateTime.parse(startTimeStr, DATE_FORMATTER));
        }

        System.out.print("结束时间 (格式: yyyy-MM-dd HH:mm): ");
        String endTimeStr = reader.readLine();
        if (!endTimeStr.isEmpty()) {
            request.setEndTime(LocalDateTime.parse(endTimeStr, DATE_FORMATTER));
        }

        System.out.print("学分: ");
        String creditsStr = reader.readLine();
        if (!creditsStr.isEmpty()) {
            request.setCredits(Integer.parseInt(creditsStr));
        }

        System.out.print("费用 (可选): ");
        String feeStr = reader.readLine();
        if (!feeStr.isEmpty()) {
            request.setFee(new BigDecimal(feeStr));
        }

        CourseDTO course = apiService.createCourse(request);
        System.out.println("课程创建成功:");
        System.out.println(OutputFormatter.formatCourse(course));
    }

    private void register(String employeeId, String courseId) throws Exception {
        RegisterRequest request = new RegisterRequest();
        request.setEmployeeId(employeeId);
        request.setCourseId(courseId);

        RegistrationDTO reg = apiService.register(request);
        System.out.println("报名成功:");
        System.out.println(OutputFormatter.formatRegistration(reg));
    }

    private void cancelRegistration(String registrationId) throws Exception {
        CancelRegistrationRequest request = new CancelRegistrationRequest();
        request.setRegistrationId(registrationId);

        System.out.print("取消原因: ");
        request.setReason(reader.readLine());

        RegistrationDTO reg = apiService.cancelRegistration(request);
        System.out.println("取消申请已提交:");
        System.out.println(OutputFormatter.formatRegistration(reg));
    }

    private void approveCancellation(String[] args) throws Exception {
        ApproveCancellationRequest request = new ApproveCancellationRequest();
        
        if (args.length >= 2) {
            request.setRegistrationId(args[1]);
        } else {
            System.out.print("报名ID: ");
            request.setRegistrationId(reader.readLine());
        }

        System.out.print("是否批准? (yes/no): ");
        String approvedStr = reader.readLine().toLowerCase();
        request.setApproved(approvedStr.equals("yes") || approvedStr.equals("y"));

        System.out.print("备注 (可选): ");
        request.setComment(reader.readLine());

        RegistrationDTO reg = apiService.approveCancellation(request);
        System.out.println("处理完成:");
        System.out.println(OutputFormatter.formatRegistration(reg));
    }

    private void listRegistrationsByEmployee(String employeeId) throws Exception {
        List<RegistrationDTO> registrations = apiService.getRegistrationsByEmployee(employeeId);
        System.out.println(OutputFormatter.formatRegistrations(registrations));
    }

    private void listRegistrationsByCourse(String courseId) throws Exception {
        List<RegistrationDTO> registrations = apiService.getRegistrationsByCourse(courseId);
        System.out.println(OutputFormatter.formatRegistrations(registrations));
    }

    private void getRegistration(String id) throws Exception {
        RegistrationDTO reg = apiService.getRegistration(id);
        System.out.println(OutputFormatter.formatRegistration(reg));
    }

    private void submitScore(String[] args) throws Exception {
        ScoreRequest request = new ScoreRequest();

        if (args.length >= 2) {
            request.setRegistrationId(args[1]);
        } else {
            System.out.print("报名ID: ");
            request.setRegistrationId(reader.readLine());
        }

        System.out.print("分数 (0-100): ");
        request.setScore(Integer.parseInt(reader.readLine()));

        System.out.print("评语 (可选): ");
        request.setComment(reader.readLine());

        RegistrationDTO reg = apiService.submitScore(request);
        System.out.println("打分成功:");
        System.out.println(OutputFormatter.formatRegistration(reg));
    }

    private void submitEvaluation(String[] args) throws Exception {
        EvaluationRequest request = new EvaluationRequest();

        if (args.length >= 2) {
            request.setRegistrationId(args[1]);
        } else {
            System.out.print("报名ID: ");
            request.setRegistrationId(reader.readLine());
        }

        System.out.print("评分 (1-5星): ");
        request.setRating(Integer.parseInt(reader.readLine()));

        System.out.print("评价内容: ");
        request.setComment(reader.readLine());

        RegistrationDTO reg = apiService.submitEvaluation(request);
        System.out.println("评价提交成功:");
        System.out.println(OutputFormatter.formatRegistration(reg));
    }

    private void getStatistics() throws Exception {
        StatisticsDTO stats = apiService.getStatistics();
        System.out.println(OutputFormatter.formatStatistics(stats));
    }

    private void getCredits(String[] args) throws Exception {
        String employeeId;
        Integer year = null;

        if (args.length >= 2) {
            employeeId = args[1];
        } else {
            System.out.print("员工ID: ");
            employeeId = reader.readLine();
        }

        if (args.length >= 3) {
            year = Integer.parseInt(args[2]);
        }

        CreditSummaryDTO summary = apiService.getCreditSummary(employeeId, year);
        System.out.println(OutputFormatter.formatCreditSummary(summary));
    }

    private void requireArgs(String[] args, int minCount, String message) {
        if (args.length < minCount) {
            System.out.println(message);
            throw new RuntimeException("参数不足");
        }
    }

    private void printHelp() {
        String separator = "=".repeat(80);
        System.out.println(separator);
        System.out.println("                        培训管理平台客户端");
        System.out.println(separator);
        System.out.println("\n【基础命令】:");
        System.out.println("  help, -h, --help          显示此帮助信息");
        System.out.println("\n【部门管理】:");
        System.out.println("  dept-list                列出所有部门");
        System.out.println("  dept-get <部门ID>       获取部门详情");
        System.out.println("  dept-create              创建新部门");
        System.out.println("\n【员工管理】:");
        System.out.println("  emp-list                 列出所有员工");
        System.out.println("  emp-get <员工ID>         获取员工详情");
        System.out.println("  emp-create               创建新员工");
        System.out.println("\n【讲师管理】:");
        System.out.println("  inst-list                列出所有讲师");
        System.out.println("  inst-get <讲师ID>        获取讲师详情");
        System.out.println("  inst-create              创建新讲师");
        System.out.println("  inst-evaluations <讲师ID> 查看讲师评价");
        System.out.println("  inst-pending-cancellations <讲师ID> 查看待审批的取消申请");
        System.out.println("\n【课程管理】:");
        System.out.println("  course-list [--published] 列出课程 (--published 只显示已发布)");
        System.out.println("  course-get <课程ID>       获取课程详情");
        System.out.println("  course-create             创建新课程");
        System.out.println("\n【报名管理】:");
        System.out.println("  register <员工ID> <课程ID>  员工报名课程");
        System.out.println("  cancel-registration <报名ID> 申请取消报名");
        System.out.println("  approve-cancellation       审批取消申请");
        System.out.println("  reg-list-by-emp <员工ID>   查看员工的报名记录");
        System.out.println("  reg-list-by-course <课程ID> 查看课程的报名记录");
        System.out.println("  reg-get <报名ID>            获取报名详情");
        System.out.println("\n【评分与评价】:");
        System.out.println("  score [报名ID]              讲师给学员打分");
        System.out.println("  evaluate [报名ID]           学员评价课程");
        System.out.println("\n【统计与学分】:");
        System.out.println("  stats                      查看统计概览");
        System.out.println("  credits <员工ID> [年份]     查看员工学分情况");
        System.out.println("\n");
        System.out.println("环境变量:");
        System.out.println("  TRAINING_API_URL          API服务地址 (默认: http://localhost:8080)");
        System.out.println(separator);
    }
}
