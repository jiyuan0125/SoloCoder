package com.orgchart.server.controller;

import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.VirtualTeamDTO;
import com.orgchart.common.dto.request.CreateVirtualTeamRequest;
import com.orgchart.server.service.VirtualTeamService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/virtual-teams")
public class VirtualTeamController {

    private final VirtualTeamService virtualTeamService;

    public VirtualTeamController(VirtualTeamService virtualTeamService) {
        this.virtualTeamService = virtualTeamService;
    }

    @PostMapping
    public ApiResponse<VirtualTeamDTO> createVirtualTeam(@RequestBody CreateVirtualTeamRequest request) {
        VirtualTeamDTO team = virtualTeamService.createVirtualTeam(request);
        return ApiResponse.success(team);
    }

    @PutMapping("/{id}")
    public ApiResponse<VirtualTeamDTO> updateVirtualTeam(@PathVariable String id, 
                                                           @RequestBody CreateVirtualTeamRequest request) {
        VirtualTeamDTO team = virtualTeamService.updateVirtualTeam(id, request);
        return ApiResponse.success(team);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteVirtualTeam(@PathVariable String id) {
        virtualTeamService.deleteVirtualTeam(id);
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<VirtualTeamDTO> getVirtualTeamById(@PathVariable String id) {
        VirtualTeamDTO team = virtualTeamService.getVirtualTeamById(id);
        return ApiResponse.success(team);
    }

    @GetMapping
    public ApiResponse<List<VirtualTeamDTO>> getAllVirtualTeams() {
        List<VirtualTeamDTO> teams = virtualTeamService.getAllVirtualTeams();
        return ApiResponse.success(teams);
    }

    @GetMapping("/employee/{employeeId}")
    public ApiResponse<List<VirtualTeamDTO>> getVirtualTeamsByEmployee(@PathVariable String employeeId) {
        List<VirtualTeamDTO> teams = virtualTeamService.getVirtualTeamsByEmployee(employeeId);
        return ApiResponse.success(teams);
    }
}
