package com.workshift.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.workshift.common.dto.ObjectionDTO;
import com.workshift.common.request.CreateObjectionRequest;
import com.workshift.common.response.ApiResponse;
import com.workshift.client.http.HttpClient;
import com.workshift.client.util.OutputFormatter;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class ObjectionCommand implements Command {

    private final HttpClient httpClient;

    public ObjectionCommand(HttpClient httpClient) {
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
                    listObjectionsByEmployee(args[2]);
                } else {
                    listAllObjections();
                }
                break;
            case "get":
                if (args.length < 3) {
                    System.out.println("用法: objection get <objectionId>");
                    return;
                }
                getObjection(args[2]);
                break;
            case "create":
                if (args.length < 5) {
                    System.out.println("用法: objection create <shiftId> <employeeId> <reason>");
                    return;
                }
                createObjection(args[2], args[3], args, 4);
                break;
            case "resolve":
                if (args.length < 4) {
                    System.out.println("用法: objection resolve <objectionId> <note>");
                    return;
                }
                resolveObjection(args[2], args, 3);
                break;
            case "dismiss":
                if (args.length < 4) {
                    System.out.println("用法: objection dismiss <objectionId> <note>");
                    return;
                }
                dismissObjection(args[2], args, 3);
                break;
            default:
                System.out.println("未知操作: " + action);
                printHelp();
        }
    }

    private void listAllObjections() throws Exception {
        TypeReference<ApiResponse<List<ObjectionDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<ObjectionDTO>>>() {};
        ApiResponse<List<ObjectionDTO>> response = httpClient.get("/api/objections", typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printObjections(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void listObjectionsByEmployee(String employeeId) throws Exception {
        TypeReference<ApiResponse<List<ObjectionDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<ObjectionDTO>>>() {};
        ApiResponse<List<ObjectionDTO>> response = httpClient.get("/api/objections/employee/" + employeeId, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printObjections(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void getObjection(String objectionId) throws Exception {
        TypeReference<ApiResponse<ObjectionDTO>> typeRef = 
                new TypeReference<ApiResponse<ObjectionDTO>>() {};
        ApiResponse<ObjectionDTO> response = httpClient.get("/api/objections/" + objectionId, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printObjection(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void createObjection(String shiftId, String employeeId, String[] args, int startIndex) throws Exception {
        StringBuilder reason = new StringBuilder();
        for (int i = startIndex; i < args.length; i++) {
            if (i > startIndex) reason.append(" ");
            reason.append(args[i]);
        }

        CreateObjectionRequest request = new CreateObjectionRequest();
        request.setShiftId(shiftId);
        request.setEmployeeId(employeeId);
        request.setReason(reason.toString());

        TypeReference<ApiResponse<ObjectionDTO>> typeRef = 
                new TypeReference<ApiResponse<ObjectionDTO>>() {};
        ApiResponse<ObjectionDTO> response = httpClient.post("/api/objections", request, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("异议创建成功");
            OutputFormatter.printObjection(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void resolveObjection(String objectionId, String[] args, int startIndex) throws Exception {
        StringBuilder note = new StringBuilder();
        for (int i = startIndex; i < args.length; i++) {
            if (i > startIndex) note.append(" ");
            note.append(args[i]);
        }

        Map<String, String> params = new HashMap<>();
        params.put("handlerNote", note.toString());

        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        String path = httpClient.buildPathWithParams("/api/objections/" + objectionId + "/resolve", params);
        ApiResponse<Void> response = httpClient.put(path, null, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("异议已解决");
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void dismissObjection(String objectionId, String[] args, int startIndex) throws Exception {
        StringBuilder note = new StringBuilder();
        for (int i = startIndex; i < args.length; i++) {
            if (i > startIndex) note.append(" ");
            note.append(args[i]);
        }

        Map<String, String> params = new HashMap<>();
        params.put("handlerNote", note.toString());

        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        String path = httpClient.buildPathWithParams("/api/objections/" + objectionId + "/dismiss", params);
        ApiResponse<Void> response = httpClient.put(path, null, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("异议已驳回");
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void printHelp() {
        System.out.println("异议管理命令:");
        System.out.println("  list                              - 列出所有异议");
        System.out.println("  list <employeeId>                 - 列出指定员工的异议");
        System.out.println("  get <objectionId>                 - 获取异议详情");
        System.out.println("  create <shiftId> <empId> <reason> - 创建异议");
        System.out.println("  resolve <objectionId> <note>      - 解决异议");
        System.out.println("  dismiss <objectionId> <note>      - 驳回异议");
    }
}