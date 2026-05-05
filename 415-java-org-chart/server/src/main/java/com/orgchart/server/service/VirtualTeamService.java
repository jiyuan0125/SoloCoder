package com.orgchart.server.service;

import com.orgchart.common.dto.VirtualTeamDTO;
import com.orgchart.common.dto.request.CreateVirtualTeamRequest;
import com.orgchart.common.enums.ErrorCode;
import com.orgchart.server.entity.Employee;
import com.orgchart.server.entity.VirtualTeam;
import com.orgchart.server.exception.BusinessException;
import com.orgchart.server.repository.EmployeeRepository;
import com.orgchart.server.repository.VirtualTeamRepository;
import com.orgchart.server.util.IdGenerator;
import org.springframework.beans.BeanUtils;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

@Service
public class VirtualTeamService {

    private final VirtualTeamRepository virtualTeamRepository;
    private final EmployeeRepository employeeRepository;
    private final OperationLogService operationLogService;

    public VirtualTeamService(VirtualTeamRepository virtualTeamRepository,
                               EmployeeRepository employeeRepository,
                               OperationLogService operationLogService) {
        this.virtualTeamRepository = virtualTeamRepository;
        this.employeeRepository = employeeRepository;
        this.operationLogService = operationLogService;
    }

    public VirtualTeamDTO createVirtualTeam(CreateVirtualTeamRequest request) {
        if (request.getName() == null || request.getName().trim().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "虚拟团队名称不能为空");
        }

        if (virtualTeamRepository.existsByName(request.getName())) {
            throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "虚拟团队名称已存在");
        }

        if (request.getMemberIds() != null) {
            for (String memberId : request.getMemberIds()) {
                if (!employeeRepository.existsById(memberId)) {
                    throw new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND);
                }
            }
        }

        VirtualTeam team = new VirtualTeam();
        team.setId(IdGenerator.generateVirtualTeamId());
        team.setName(request.getName());
        team.setDescription(request.getDescription());
        if (request.getMemberIds() != null) {
            team.setMemberIds(new ArrayList<>(request.getMemberIds()));
        }
        team.setCreatedAt(LocalDateTime.now());
        team.setUpdatedAt(LocalDateTime.now());

        VirtualTeam saved = virtualTeamRepository.save(team);

        if (request.getMemberIds() != null) {
            for (String memberId : request.getMemberIds()) {
                Employee employee = employeeRepository.findById(memberId).orElse(null);
                if (employee != null && !employee.getVirtualTeamIds().contains(saved.getId())) {
                    employee.getVirtualTeamIds().add(saved.getId());
                    employeeRepository.save(employee);
                }
            }
        }

        operationLogService.logCreate("VIRTUAL_TEAM", saved.getId(), saved.getName(),
                "名称: " + saved.getName() + ", 成员数: " + saved.getMemberIds().size());

        return toDTO(saved);
    }

    public VirtualTeamDTO updateVirtualTeam(String id, CreateVirtualTeamRequest request) {
        VirtualTeam team = virtualTeamRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.VIRTUAL_TEAM_NOT_FOUND));

        String oldName = team.getName();
        List<String> oldMemberIds = new ArrayList<>(team.getMemberIds());

        StringBuilder changes = new StringBuilder();

        if (request.getName() != null && !request.getName().trim().isEmpty()) {
            if (!request.getName().equals(team.getName())) {
                if (virtualTeamRepository.existsByNameExcludingId(request.getName(), id)) {
                    throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "虚拟团队名称已存在");
                }
                changes.append("名称: ").append(oldName).append(" -> ").append(request.getName()).append("; ");
                team.setName(request.getName());
            }
        }

        if (request.getDescription() != null) {
            if (!request.getDescription().equals(team.getDescription())) {
                changes.append("描述变更; ");
                team.setDescription(request.getDescription());
            }
        }

        if (request.getMemberIds() != null) {
            for (String memberId : request.getMemberIds()) {
                if (!employeeRepository.existsById(memberId)) {
                    throw new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND);
                }
            }

            Set<String> oldMemberSet = new HashSet<>(oldMemberIds);
            Set<String> newMemberSet = new HashSet<>(request.getMemberIds());

            Set<String> toAdd = new HashSet<>(newMemberSet);
            toAdd.removeAll(oldMemberSet);

            Set<String> toRemove = new HashSet<>(oldMemberSet);
            toRemove.removeAll(newMemberSet);

            for (String memberId : toAdd) {
                Employee employee = employeeRepository.findById(memberId).orElse(null);
                if (employee != null && !employee.getVirtualTeamIds().contains(id)) {
                    employee.getVirtualTeamIds().add(id);
                    employeeRepository.save(employee);
                }
            }

            for (String memberId : toRemove) {
                Employee employee = employeeRepository.findById(memberId).orElse(null);
                if (employee != null) {
                    employee.getVirtualTeamIds().remove(id);
                    employeeRepository.save(employee);
                }
            }

            team.setMemberIds(new ArrayList<>(request.getMemberIds()));
            changes.append("成员变更: ").append(oldMemberIds.size()).append("人 -> ").append(request.getMemberIds().size()).append("人; ");
        }

        team.setUpdatedAt(LocalDateTime.now());
        VirtualTeam saved = virtualTeamRepository.save(team);

        if (!changes.isEmpty()) {
            operationLogService.logUpdate("VIRTUAL_TEAM", saved.getId(), saved.getName(),
                    "原信息: " + oldName, "变更: " + changes);
        }

        return toDTO(saved);
    }

    public void deleteVirtualTeam(String id) {
        VirtualTeam team = virtualTeamRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.VIRTUAL_TEAM_NOT_FOUND));

        List<String> memberIds = new ArrayList<>(team.getMemberIds());
        for (String memberId : memberIds) {
            Employee employee = employeeRepository.findById(memberId).orElse(null);
            if (employee != null) {
                employee.getVirtualTeamIds().remove(id);
                employeeRepository.save(employee);
            }
        }

        virtualTeamRepository.deleteById(id);

        operationLogService.logDelete("VIRTUAL_TEAM", id, team.getName(),
                "删除虚拟团队: " + team.getName());
    }

    public VirtualTeamDTO getVirtualTeamById(String id) {
        VirtualTeam team = virtualTeamRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.VIRTUAL_TEAM_NOT_FOUND));
        return toDTO(team);
    }

    public List<VirtualTeamDTO> getAllVirtualTeams() {
        List<VirtualTeam> teams = virtualTeamRepository.findAll();
        return teams.stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public List<VirtualTeamDTO> getVirtualTeamsByEmployee(String employeeId) {
        if (!employeeRepository.existsById(employeeId)) {
            throw new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND);
        }
        return virtualTeamRepository.findByMemberId(employeeId).stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    private VirtualTeamDTO toDTO(VirtualTeam entity) {
        VirtualTeamDTO dto = new VirtualTeamDTO();
        BeanUtils.copyProperties(entity, dto);

        List<String> memberNames = new ArrayList<>();
        for (String memberId : entity.getMemberIds()) {
            Employee employee = employeeRepository.findById(memberId).orElse(null);
            if (employee != null) {
                memberNames.add(employee.getName());
            }
        }
        dto.setMemberNames(memberNames);

        return dto;
    }
}
