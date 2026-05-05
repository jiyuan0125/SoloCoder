package com.orgchart.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.VirtualTeamDTO;
import com.orgchart.common.dto.request.CreateVirtualTeamRequest;
import com.orgchart.client.http.HttpClient;
import com.orgchart.client.util.OutputFormatter;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class VirtualTeamCommands implements Command {

    @Override
    public String getName() {
        return "virtual-team";
    }

    @Override
    public String getDescription() {
        return "虚拟团队管理命令 (list, get, create, update, delete)";
    }

    @Override
    public void execute(String[] args, HttpClient httpClient) throws Exception {
        OutputFormatter formatter = new OutputFormatter(httpClient.getObjectMapper());

        if (args.length == 0) {
            printHelp();
            return;
        }

        String subCommand = args[0];
        String[] remainingArgs = args.length > 1 ? Arrays.copyOfRange(args, 1, args.length) : new String[0];

        switch (subCommand) {
            case "list":
                listVirtualTeams(httpClient, formatter);
                break;
            case "get":
                getVirtualTeam(remainingArgs, httpClient, formatter);
                break;
            case "create":
                createVirtualTeam(remainingArgs, httpClient, formatter);
                break;
            case "update":
                updateVirtualTeam(remainingArgs, httpClient, formatter);
                break;
            case "delete":
                deleteVirtualTeam(remainingArgs, httpClient, formatter);
                break;
            default:
                System.out.println("未知的子命令: " + subCommand);
                printHelp();
        }
    }

    private void listVirtualTeams(HttpClient httpClient, OutputFormatter formatter) throws Exception {
        String path = "/api/virtual-teams";
        TypeReference<ApiResponse<List<VirtualTeamDTO>>> typeRef = new TypeReference<ApiResponse<List<VirtualTeamDTO>>>() {};
        ApiResponse<List<VirtualTeamDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<VirtualTeamDTO> teams = response.getData();
            if (teams == null || teams.isEmpty()) {
                System.out.println("暂无虚拟团队数据");
                return;
            }

            System.out.println("虚拟团队列表:");
            System.out.println();

            List<String> headers = Arrays.asList("ID", "名称", "描述", "成员数");
            List<List<String>> rows = new ArrayList<>();
            for (VirtualTeamDTO team : teams) {
                List<String> row = new ArrayList<>();
                row.add(team.getId());
                row.add(team.getName());
                row.add(team.getDescription() != null ? team.getDescription() : "-");
                row.add(String.valueOf(team.getMemberIds() != null ? team.getMemberIds().size() : 0));
                rows.add(row);
            }
            formatter.printTable(headers, rows);
        } else {
            formatter.printResult(response);
        }
    }

    private void getVirtualTeam(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: virtual-team get <id>");
            return;
        }

        String id = args[0];
        String path = "/api/virtual-teams/" + id;
        TypeReference<ApiResponse<VirtualTeamDTO>> typeRef = new TypeReference<ApiResponse<VirtualTeamDTO>>() {};
        ApiResponse<VirtualTeamDTO> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            VirtualTeamDTO team = response.getData();
            System.out.println("虚拟团队详情:");
            System.out.println();

            Map<String, String> map = new LinkedHashMap<>();
            map.put("ID", team.getId());
            map.put("名称", team.getName());
            map.put("描述", team.getDescription() != null ? team.getDescription() : "-");
            map.put("成员数", String.valueOf(team.getMemberIds() != null ? team.getMemberIds().size() : 0));
            
            String members = team.getMemberNames() != null && !team.getMemberNames().isEmpty()
                    ? String.join(", ", team.getMemberNames()) : "无";
            map.put("成员", members);
            
            map.put("创建时间", team.getCreatedAt() != null ? team.getCreatedAt().toString() : "-");
            map.put("更新时间", team.getUpdatedAt() != null ? team.getUpdatedAt().toString() : "-");

            formatter.printKeyValue(map);
        } else {
            formatter.printResult(response);
        }
    }

    private void createVirtualTeam(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        Map<String, String> params = parseParams(args);

        if (!params.containsKey("name")) {
            System.out.println("用法: virtual-team create --name <名称> [--description <描述>] [--memberIds <成员ID列表>]");
            return;
        }

        CreateVirtualTeamRequest request = new CreateVirtualTeamRequest();
        request.setName(params.get("name"));
        request.setDescription(params.get("description"));
        
        if (params.containsKey("memberIds")) {
            String memberIdsStr = params.get("memberIds");
            String[] memberIds = memberIdsStr.split(",");
            request.setMemberIds(Arrays.asList(memberIds));
        }

        String path = "/api/virtual-teams";
        TypeReference<ApiResponse<VirtualTeamDTO>> typeRef = new TypeReference<ApiResponse<VirtualTeamDTO>>() {};
        ApiResponse<VirtualTeamDTO> response = httpClient.post(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void updateVirtualTeam(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: virtual-team update <id> [--name <名称>] [--description <描述>] [--memberIds <成员ID列表>]");
            return;
        }

        String id = args[0];
        Map<String, String> params = parseParams(Arrays.copyOfRange(args, 1, args.length));

        if (params.isEmpty()) {
            System.out.println("至少需要指定一个更新字段");
            return;
        }

        CreateVirtualTeamRequest request = new CreateVirtualTeamRequest();
        request.setName(params.get("name"));
        request.setDescription(params.get("description"));
        
        if (params.containsKey("memberIds")) {
            String memberIdsStr = params.get("memberIds");
            String[] memberIds = memberIdsStr.split(",");
            request.setMemberIds(Arrays.asList(memberIds));
        }

        String path = "/api/virtual-teams/" + id;
        TypeReference<ApiResponse<VirtualTeamDTO>> typeRef = new TypeReference<ApiResponse<VirtualTeamDTO>>() {};
        ApiResponse<VirtualTeamDTO> response = httpClient.put(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void deleteVirtualTeam(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: virtual-team delete <id>");
            return;
        }

        String id = args[0];
        String path = "/api/virtual-teams/" + id;
        TypeReference<ApiResponse<Void>> typeRef = new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.delete(path, typeRef.getType());
        formatter.printResult(response);
    }

    private Map<String, String> parseParams(String[] args) {
        Map<String, String> params = new HashMap<>();
        for (int i = 0; i < args.length; i++) {
            String arg = args[i];
            if (arg.startsWith("--") && i + 1 < args.length) {
                String key = arg.substring(2);
                String value = args[i + 1];
                if (!value.startsWith("--")) {
                    params.put(key, value);
                    i++;
                }
            }
        }
        return params;
    }

    private void printHelp() {
        System.out.println();
        System.out.println("虚拟团队管理命令:");
        System.out.println();
        System.out.println("  virtual-team list                 列出所有虚拟团队");
        System.out.println("  virtual-team get <id>            获取虚拟团队详情");
        System.out.println("  virtual-team create --name <名称>");
        System.out.println("                                    创建虚拟团队");
        System.out.println("  virtual-team update <id> [--name <名称>] ...");
        System.out.println("                                    更新虚拟团队");
        System.out.println("  virtual-team delete <id>          删除虚拟团队");
        System.out.println();
    }
}
