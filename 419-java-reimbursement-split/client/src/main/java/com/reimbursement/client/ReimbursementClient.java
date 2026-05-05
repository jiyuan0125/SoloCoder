package com.reimbursement.client;

import com.reimbursement.client.util.HttpClientUtil;
import com.reimbursement.client.util.JsonUtil;
import com.reimbursement.client.util.OutputFormatter;
import com.reimbursement.common.dto.*;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class ReimbursementClient {
    private static final BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));

    public static void main(String[] args) {
        try {
            if (args.length == 0) {
                OutputFormatter.printHelp();
                return;
            }

            String command = args[0].toLowerCase();

            switch (command) {
                case "list":
                    listReimbursements();
                    break;
                case "get":
                    if (args.length < 2) {
                        System.out.println("用法: get <报销单号>");
                        return;
                    }
                    getReimbursement(args[1]);
                    break;
                case "cost-centers":
                    listCostCenters();
                    break;
                case "report":
                    if (args.length < 4) {
                        System.out.println("用法: report <年> <月> <成本中心ID>");
                        return;
                    }
                    getMonthlyReport(Integer.parseInt(args[1]), Integer.parseInt(args[2]), args[3]);
                    break;
                case "create":
                    createReimbursement();
                    break;
                case "submit":
                    if (args.length < 2) {
                        System.out.println("用法: submit <报销单号>");
                        return;
                    }
                    submitReimbursement(args[1]);
                    break;
                case "approve":
                    approveReimbursement();
                    break;
                case "final-review":
                    finalReview();
                    break;
                case "help":
                default:
                    OutputFormatter.printHelp();
                    break;
            }
        } catch (Exception e) {
            System.out.println("执行出错: " + e.getMessage());
            e.printStackTrace();
        }
    }

    private static void listReimbursements() throws Exception {
        String response = HttpClientUtil.get("/list");
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        
        if (apiResponse.getCode() == 200) {
            List<?> data = (List<?>) apiResponse.getData();
            List<ReimbursementDetailDTO> list = new ArrayList<>();
            for (Object item : data) {
                if (item instanceof Map) {
                    @SuppressWarnings("unchecked")
                    ReimbursementDetailDTO dto = (ReimbursementDetailDTO) JsonUtil.parseToObject(
                        (Map<String, Object>) item, ReimbursementDetailDTO.class);
                    list.add(dto);
                }
            }
            OutputFormatter.printReimbursementList(list);
        } else {
            OutputFormatter.printApiResponse(apiResponse);
        }
    }

    private static void getReimbursement(String id) throws Exception {
        String response = HttpClientUtil.get("/" + id);
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        
        if (apiResponse.getCode() == 200 && apiResponse.getData() instanceof Map) {
            @SuppressWarnings("unchecked")
            ReimbursementDetailDTO dto = JsonUtil.parseToObject(
                (Map<String, Object>) apiResponse.getData(), ReimbursementDetailDTO.class);
            OutputFormatter.printReimbursementDetail(dto);
        } else {
            OutputFormatter.printApiResponse(apiResponse);
        }
    }

    private static void listCostCenters() throws Exception {
        String response = HttpClientUtil.get("/cost-centers");
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        
        if (apiResponse.getCode() == 200) {
            List<?> data = (List<?>) apiResponse.getData();
            List<CostCenterDTO> list = new ArrayList<>();
            for (Object item : data) {
                if (item instanceof Map) {
                    @SuppressWarnings("unchecked")
                    CostCenterDTO cc = JsonUtil.parseToObject(
                        (Map<String, Object>) item, CostCenterDTO.class);
                    list.add(cc);
                }
            }
            OutputFormatter.printCostCenterList(list);
        } else {
            OutputFormatter.printApiResponse(apiResponse);
        }
    }

    private static void getMonthlyReport(int year, int month, String costCenterId) throws Exception {
        String response = HttpClientUtil.get("/report/" + year + "/" + month + "/" + costCenterId);
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        
        if (apiResponse.getCode() == 200 && apiResponse.getData() instanceof Map) {
            @SuppressWarnings("unchecked")
            MonthlyReportDTO dto = JsonUtil.parseToObject(
                (Map<String, Object>) apiResponse.getData(), MonthlyReportDTO.class);
            OutputFormatter.printMonthlyReport(dto);
        } else {
            OutputFormatter.printApiResponse(apiResponse);
        }
    }

    private static void createReimbursement() throws Exception {
        System.out.println();
        System.out.println("=== 创建报销单 ===");
        
        CreateReimbursementRequest request = new CreateReimbursementRequest();
        
        System.out.print("员工ID: ");
        request.setEmployeeId(reader.readLine().trim());
        
        System.out.print("员工姓名: ");
        request.setEmployeeName(reader.readLine().trim());
        
        System.out.print("报销类型 (如: 差旅、办公、招待): ");
        request.setReimbursementType(reader.readLine().trim());
        
        System.out.print("描述: ");
        request.setDescription(reader.readLine().trim());
        
        System.out.print("总金额: ");
        request.setTotalAmount(new BigDecimal(reader.readLine().trim()));
        
        System.out.println("\n可用成本中心:");
        listCostCenters();
        
        System.out.print("\n分配到几个成本中心? (最多5个，超过需特殊审批): ");
        int count = Integer.parseInt(reader.readLine().trim());
        
        if (count > 5) {
            System.out.print("超过5个成本中心，是否申请特殊审批? (y/n): ");
            String special = reader.readLine().trim().toLowerCase();
            request.setSpecialApprovalRequested("y".equals(special) || "yes".equals(special));
        }
        
        List<CreateReimbursementRequest.AllocationItem> allocations = new ArrayList<>();
        for (int i = 1; i <= count; i++) {
            CreateReimbursementRequest.AllocationItem item = new CreateReimbursementRequest.AllocationItem();
            
            System.out.printf("\n第 %d 个成本中心:%n", i);
            System.out.print("  成本中心ID: ");
            item.setCostCenterId(reader.readLine().trim());
            
            System.out.print("  分配百分比 (1-100): ");
            item.setPercentage(Integer.parseInt(reader.readLine().trim()));
            
            allocations.add(item);
        }
        request.setAllocations(allocations);
        
        String json = JsonUtil.toJson(request);
        String response = HttpClientUtil.post("/create", json);
        
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        OutputFormatter.printApiResponse(apiResponse);
        
        if (apiResponse.getCode() == 200 && apiResponse.getData() instanceof Map) {
            @SuppressWarnings("unchecked")
            ReimbursementDetailDTO dto = JsonUtil.parseToObject(
                (Map<String, Object>) apiResponse.getData(), ReimbursementDetailDTO.class);
            OutputFormatter.printReimbursementDetail(dto);
        }
    }

    private static void submitReimbursement(String id) throws Exception {
        String response = HttpClientUtil.post("/submit/" + id, "");
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        OutputFormatter.printApiResponse(apiResponse);
        
        if (apiResponse.getCode() == 200 && apiResponse.getData() instanceof Map) {
            @SuppressWarnings("unchecked")
            ReimbursementDetailDTO dto = JsonUtil.parseToObject(
                (Map<String, Object>) apiResponse.getData(), ReimbursementDetailDTO.class);
            OutputFormatter.printReimbursementDetail(dto);
        }
    }

    private static void approveReimbursement() throws Exception {
        System.out.println();
        System.out.println("=== 审批报销单 ===");
        
        ApprovalRequest request = new ApprovalRequest();
        
        System.out.print("报销单号: ");
        request.setReimbursementId(reader.readLine().trim());
        
        System.out.print("成本中心ID: ");
        request.setCostCenterId(reader.readLine().trim());
        
        System.out.print("审批人ID: ");
        request.setApproverId(reader.readLine().trim());
        
        System.out.print("审批人姓名: ");
        request.setApproverName(reader.readLine().trim());
        
        System.out.print("审批结果 (通过: y / 驳回: n): ");
        String result = reader.readLine().trim().toLowerCase();
        request.setApproved("y".equals(result) || "yes".equals(result));
        
        System.out.print("审批意见 (可选): ");
        request.setComment(reader.readLine().trim());
        
        String json = JsonUtil.toJson(request);
        String response = HttpClientUtil.post("/approve", json);
        
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        OutputFormatter.printApiResponse(apiResponse);
        
        if (apiResponse.getCode() == 200 && apiResponse.getData() instanceof Map) {
            @SuppressWarnings("unchecked")
            ReimbursementDetailDTO dto = JsonUtil.parseToObject(
                (Map<String, Object>) apiResponse.getData(), ReimbursementDetailDTO.class);
            OutputFormatter.printReimbursementDetail(dto);
        }
    }

    private static void finalReview() throws Exception {
        System.out.println();
        System.out.println("=== 财务复核 ===");
        
        FinalReviewRequest request = new FinalReviewRequest();
        
        System.out.print("报销单号: ");
        request.setReimbursementId(reader.readLine().trim());
        
        System.out.print("复核人ID: ");
        request.setReviewerId(reader.readLine().trim());
        
        System.out.print("复核人姓名: ");
        request.setReviewerName(reader.readLine().trim());
        
        System.out.print("复核结果 (通过: y / 驳回: n): ");
        String result = reader.readLine().trim().toLowerCase();
        request.setApproved("y".equals(result) || "yes".equals(result));
        
        System.out.print("复核意见 (可选): ");
        request.setComment(reader.readLine().trim());
        
        String json = JsonUtil.toJson(request);
        String response = HttpClientUtil.post("/final-review", json);
        
        ApiResponse<?> apiResponse = JsonUtil.fromJson(response, ApiResponse.class);
        OutputFormatter.printApiResponse(apiResponse);
        
        if (apiResponse.getCode() == 200 && apiResponse.getData() instanceof Map) {
            @SuppressWarnings("unchecked")
            ReimbursementDetailDTO dto = JsonUtil.parseToObject(
                (Map<String, Object>) apiResponse.getData(), ReimbursementDetailDTO.class);
            OutputFormatter.printReimbursementDetail(dto);
        }
    }
}
