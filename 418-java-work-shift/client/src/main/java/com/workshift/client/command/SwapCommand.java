package com.workshift.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.workshift.common.dto.ShiftSwapDTO;
import com.workshift.common.request.ConfirmSwapRequest;
import com.workshift.common.request.CreateSwapRequest;
import com.workshift.common.response.ApiResponse;
import com.workshift.client.http.HttpClient;
import com.workshift.client.util.OutputFormatter;
import java.time.LocalDate;
import java.util.List;

public class SwapCommand implements Command {

    private final HttpClient httpClient;

    public SwapCommand(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    @Override
    public void execute(String[] args) throws Exception {
        if (args.length < 2) {
            printHelp();
            return;
        }

        String action = args[1];

        switch (action.toLowerCase()) {
            case "list":
                if (args.length >= 3) {
                    listSwapsByEmployee(args[2]);
                } else {
                    listAllSwaps();
                }
                break;
            case "get":
                if (args.length < 3) {
                    System.out.println("用法: swap get <swapId>");
                    return;
                }
                getSwap(args[2]);
                break;
            case "create":
                if (args.length < 7) {
                    System.out.println("用法: swap create <requesterId> <targetId> <requesterDate> <targetDate> <signature>");
                    return;
                }
                createSwap(args[2], args[3], args[4], args[5], args[6]);
                break;
            case "confirm":
                if (args.length < 5) {
                    System.out.println("用法: swap confirm <swapId> <targetEmployeeId> <confirmed> [signature] [rejectReason]");
                    System.out.println("  confirmed: true/false");
                    return;
                }
                confirmSwap(args);
                break;
            default:
                System.out.println("未知操作: " + action);
                printHelp();
        }
    }

    private void listAllSwaps() throws Exception {
        TypeReference<ApiResponse<List<ShiftSwapDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<ShiftSwapDTO>>>() {};
        ApiResponse<List<ShiftSwapDTO>> response = httpClient.get("/api/swaps", typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSwaps(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void listSwapsByEmployee(String employeeId) throws Exception {
        TypeReference<ApiResponse<List<ShiftSwapDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<ShiftSwapDTO>>>() {};
        ApiResponse<List<ShiftSwapDTO>> response = httpClient.get("/api/swaps/employee/" + employeeId, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSwaps(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void getSwap(String swapId) throws Exception {
        TypeReference<ApiResponse<ShiftSwapDTO>> typeRef = 
                new TypeReference<ApiResponse<ShiftSwapDTO>>() {};
        ApiResponse<ShiftSwapDTO> response = httpClient.get("/api/swaps/" + swapId, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSwap(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void createSwap(String requesterId, String targetId, String requesterDateStr, 
            String targetDateStr, String signature) throws Exception {
        
        LocalDate requesterDate;
        LocalDate targetDate;
        try {
            requesterDate = LocalDate.parse(requesterDateStr);
            targetDate = LocalDate.parse(targetDateStr);
        } catch (Exception e) {
            OutputFormatter.printError("无效的日期格式");
            System.out.println("日期格式应为: yyyy-MM-dd");
            return;
        }

        CreateSwapRequest request = new CreateSwapRequest();
        request.setRequesterEmployeeId(requesterId);
        request.setTargetEmployeeId(targetId);
        request.setRequesterDate(requesterDate);
        request.setTargetDate(targetDate);
        request.setRequesterSignature(signature);

        TypeReference<ApiResponse<ShiftSwapDTO>> typeRef = 
                new TypeReference<ApiResponse<ShiftSwapDTO>>() {};
        ApiResponse<ShiftSwapDTO> response = httpClient.post("/api/swaps", request, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("换班申请创建成功");
            OutputFormatter.printSwap(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void confirmSwap(String[] args) throws Exception {
        String swapId = args[2];
        String targetEmployeeId = args[3];
        boolean confirmed = Boolean.parseBoolean(args[4]);
        String signature = args.length > 5 ? args[5] : null;
        String rejectReason = args.length > 6 ? args[6] : null;

        ConfirmSwapRequest request = new ConfirmSwapRequest();
        request.setSwapId(swapId);
        request.setTargetEmployeeId(targetEmployeeId);
        request.setConfirmed(confirmed);
        request.setTargetSignature(signature);
        request.setRejectReason(rejectReason);

        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.put("/api/swaps/confirm", request, typeRef);
        
        if (response.getCode() == 0) {
            if (confirmed) {
                OutputFormatter.printSuccess("换班确认成功");
            } else {
                OutputFormatter.printSuccess("换班已拒绝");
            }
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void printHelp() {
        System.out.println("换班管理命令:");
        System.out.println("  list                                      - 列出所有换班申请");
        System.out.println("  list <employeeId>                         - 列出指定员工的换班申请");
        System.out.println("  get <swapId>                              - 获取换班申请详情");
        System.out.println("  create <reqId> <tarId> <reqDate> <tarDate> <sig> - 创建换班申请");
        System.out.println("  confirm <swapId> <tarEmpId> <confirmed> [sig] [reason] - 确认换班");
        System.out.println();
        System.out.println("日期格式: yyyy-MM-dd");
    }
}