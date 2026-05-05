package com.orgchart.server.service;

import com.orgchart.common.dto.EmployeeDTO;
import com.orgchart.common.dto.ReportingLineDTO;
import com.orgchart.common.dto.TransferHistoryDTO;
import com.orgchart.common.dto.request.CreateEmployeeRequest;
import com.orgchart.common.dto.request.TransferEmployeeRequest;
import com.orgchart.common.dto.request.UpdateEmployeeRequest;
import com.orgchart.common.enums.ErrorCode;
import com.orgchart.server.entity.Department;
import com.orgchart.server.entity.Employee;
import com.orgchart.server.entity.TransferHistory;
import com.orgchart.server.entity.VirtualTeam;
import com.orgchart.server.exception.BusinessException;
import com.orgchart.server.repository.DepartmentRepository;
import com.orgchart.server.repository.EmployeeRepository;
import com.orgchart.server.repository.TransferHistoryRepository;
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
public class EmployeeService {

    private final EmployeeRepository employeeRepository;
    private final DepartmentRepository departmentRepository;
    private final VirtualTeamRepository virtualTeamRepository;
    private final TransferHistoryRepository transferHistoryRepository;
    private final OperationLogService operationLogService;

    public EmployeeService(EmployeeRepository employeeRepository,
                           DepartmentRepository departmentRepository,
                           VirtualTeamRepository virtualTeamRepository,
                           TransferHistoryRepository transferHistoryRepository,
                           OperationLogService operationLogService) {
        this.employeeRepository = employeeRepository;
        this.departmentRepository = departmentRepository;
        this.virtualTeamRepository = virtualTeamRepository;
        this.transferHistoryRepository = transferHistoryRepository;
        this.operationLogService = operationLogService;
    }

    public EmployeeDTO createEmployee(CreateEmployeeRequest request) {
        if (request.getName() == null || request.getName().trim().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "员工姓名不能为空");
        }

        if (request.getDepartmentId() == null || request.getDepartmentId().isEmpty()) {
            throw new BusinessException(ErrorCode.EMPLOYEE_DEPARTMENT_NOT_FOUND);
        }

        Department department = departmentRepository.findById(request.getDepartmentId())
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_DEPARTMENT_NOT_FOUND));

        if (request.getEmail() != null && employeeRepository.existsByEmail(request.getEmail())) {
            throw new BusinessException(ErrorCode.EMPLOYEE_ALREADY_EXISTS.getCode(), "邮箱已存在");
        }

        if (request.getManagerId() != null && !request.getManagerId().isEmpty()) {
            if (!employeeRepository.existsById(request.getManagerId())) {
                throw new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND.getCode(), "上级员工不存在");
            }
            if (hasCircularReference(request.getManagerId(), null)) {
                throw new BusinessException(ErrorCode.CIRCULAR_REFERENCE);
            }
        }

        if (request.getVirtualTeamIds() != null) {
            for (String teamId : request.getVirtualTeamIds()) {
                if (!virtualTeamRepository.existsById(teamId)) {
                    throw new BusinessException(ErrorCode.VIRTUAL_TEAM_NOT_FOUND);
                }
            }
        }

        Employee employee = new Employee();
        employee.setId(IdGenerator.generateEmployeeId());
        employee.setName(request.getName());
        employee.setEmail(request.getEmail());
        employee.setPhone(request.getPhone());
        employee.setDepartmentId(request.getDepartmentId());
        employee.setManagerId(request.getManagerId());
        if (request.getVirtualTeamIds() != null) {
            employee.setVirtualTeamIds(new ArrayList<>(request.getVirtualTeamIds()));
        }
        employee.setCreatedAt(LocalDateTime.now());
        employee.setUpdatedAt(LocalDateTime.now());

        Employee saved = employeeRepository.save(employee);

        if (request.getVirtualTeamIds() != null) {
            for (String teamId : request.getVirtualTeamIds()) {
                VirtualTeam team = virtualTeamRepository.findById(teamId).orElse(null);
                if (team != null && !team.getMemberIds().contains(saved.getId())) {
                    team.getMemberIds().add(saved.getId());
                    virtualTeamRepository.save(team);
                }
            }
        }

        operationLogService.logCreate("EMPLOYEE", saved.getId(), saved.getName(),
                "姓名: " + saved.getName() + ", 部门: " + department.getName());

        return toDTO(saved);
    }

    public EmployeeDTO updateEmployee(String id, UpdateEmployeeRequest request) {
        Employee employee = employeeRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND));

        String oldName = employee.getName();
        String oldDepartmentId = employee.getDepartmentId();
        String oldManagerId = employee.getManagerId();
        List<String> oldVirtualTeamIds = new ArrayList<>(employee.getVirtualTeamIds());

        StringBuilder changes = new StringBuilder();

        if (request.getName() != null && !request.getName().trim().isEmpty()) {
            if (!request.getName().equals(employee.getName())) {
                changes.append("姓名: ").append(oldName).append(" -> ").append(request.getName()).append("; ");
                employee.setName(request.getName());
            }
        }

        if (request.getEmail() != null) {
            if (!request.getEmail().equals(employee.getEmail())) {
                if (employeeRepository.existsByEmailExcludingId(request.getEmail(), id)) {
                    throw new BusinessException(ErrorCode.EMPLOYEE_ALREADY_EXISTS.getCode(), "邮箱已存在");
                }
                changes.append("邮箱: ").append(employee.getEmail()).append(" -> ").append(request.getEmail()).append("; ");
                employee.setEmail(request.getEmail());
            }
        }

        if (request.getPhone() != null) {
            if (!request.getPhone().equals(employee.getPhone())) {
                changes.append("电话: ").append(employee.getPhone()).append(" -> ").append(request.getPhone()).append("; ");
                employee.setPhone(request.getPhone());
            }
        }

        if (request.getDepartmentId() != null && !request.getDepartmentId().isEmpty()) {
            if (!request.getDepartmentId().equals(employee.getDepartmentId())) {
                Department newDept = departmentRepository.findById(request.getDepartmentId())
                        .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_DEPARTMENT_NOT_FOUND));
                String oldDeptName = oldDepartmentId != null ? 
                        departmentRepository.findById(oldDepartmentId).map(Department::getName).orElse(null) : null;
                changes.append("部门: ").append(oldDeptName).append(" -> ").append(newDept.getName()).append("; ");
                employee.setDepartmentId(request.getDepartmentId());
            }
        }

        if (request.getManagerId() != null) {
            if (!request.getManagerId().equals(employee.getManagerId())) {
                if (request.getManagerId().equals(id)) {
                    throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "不能将自己设为上级");
                }
                if (!request.getManagerId().isEmpty() && !employeeRepository.existsById(request.getManagerId())) {
                    throw new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND.getCode(), "上级员工不存在");
                }
                if (hasCircularReference(request.getManagerId(), id)) {
                    throw new BusinessException(ErrorCode.CIRCULAR_REFERENCE);
                }
                Employee oldManager = oldManagerId != null ? employeeRepository.findById(oldManagerId).orElse(null) : null;
                Employee newManager = request.getManagerId() != null && !request.getManagerId().isEmpty() ? 
                        employeeRepository.findById(request.getManagerId()).orElse(null) : null;
                changes.append("上级: ").append(oldManager != null ? oldManager.getName() : "无")
                        .append(" -> ").append(newManager != null ? newManager.getName() : "无").append("; ");
                employee.setManagerId(request.getManagerId().isEmpty() ? null : request.getManagerId());
            }
        }

        if (request.getVirtualTeamIds() != null) {
            for (String teamId : request.getVirtualTeamIds()) {
                if (!virtualTeamRepository.existsById(teamId)) {
                    throw new BusinessException(ErrorCode.VIRTUAL_TEAM_NOT_FOUND);
                }
            }

            Set<String> oldTeamSet = new HashSet<>(oldVirtualTeamIds);
            Set<String> newTeamSet = new HashSet<>(request.getVirtualTeamIds());

            Set<String> toAdd = new HashSet<>(newTeamSet);
            toAdd.removeAll(oldTeamSet);

            Set<String> toRemove = new HashSet<>(oldTeamSet);
            toRemove.removeAll(newTeamSet);

            for (String teamId : toAdd) {
                VirtualTeam team = virtualTeamRepository.findById(teamId).orElse(null);
                if (team != null && !team.getMemberIds().contains(id)) {
                    team.getMemberIds().add(id);
                    virtualTeamRepository.save(team);
                }
            }

            for (String teamId : toRemove) {
                VirtualTeam team = virtualTeamRepository.findById(teamId).orElse(null);
                if (team != null) {
                    team.getMemberIds().remove(id);
                    virtualTeamRepository.save(team);
                }
            }

            employee.setVirtualTeamIds(new ArrayList<>(request.getVirtualTeamIds()));
            changes.append("虚拟团队变更; ");
        }

        employee.setUpdatedAt(LocalDateTime.now());
        Employee saved = employeeRepository.save(employee);

        if (!changes.isEmpty()) {
            operationLogService.logUpdate("EMPLOYEE", saved.getId(), saved.getName(),
                    "原信息: " + oldName, "变更: " + changes);
        }

        return toDTO(saved);
    }

    public void deleteEmployee(String id) {
        Employee employee = employeeRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND));

        List<Employee> subordinates = employeeRepository.findByManagerId(id);
        for (Employee subordinate : subordinates) {
            subordinate.setManagerId(null);
            subordinate.setUpdatedAt(LocalDateTime.now());
            employeeRepository.save(subordinate);
        }

        List<String> teamIds = new ArrayList<>(employee.getVirtualTeamIds());
        for (String teamId : teamIds) {
            VirtualTeam team = virtualTeamRepository.findById(teamId).orElse(null);
            if (team != null) {
                team.getMemberIds().remove(id);
                virtualTeamRepository.save(team);
            }
        }

        employeeRepository.deleteById(id);

        operationLogService.logDelete("EMPLOYEE", id, employee.getName(),
                "删除员工: " + employee.getName());
    }

    public EmployeeDTO getEmployeeById(String id) {
        Employee employee = employeeRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND));
        return toDTO(employee);
    }

    public List<EmployeeDTO> getAllEmployees() {
        List<Employee> employees = employeeRepository.findAll();
        return employees.stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public List<EmployeeDTO> getEmployeesByDepartment(String departmentId, boolean includeSubDepartments) {
        List<Employee> employees = new ArrayList<>();
        
        employees.addAll(employeeRepository.findByDepartmentId(departmentId));

        if (includeSubDepartments) {
            List<Department> descendants = departmentRepository.findAllDescendants(departmentId);
            for (Department dept : descendants) {
                employees.addAll(employeeRepository.findByDepartmentId(dept.getId()));
            }
        }

        return employees.stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public EmployeeDTO transferEmployee(String id, TransferEmployeeRequest request) {
        Employee employee = employeeRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND));

        if (request.getNewDepartmentId() == null || request.getNewDepartmentId().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "目标部门不能为空");
        }

        Department targetDept = departmentRepository.findById(request.getNewDepartmentId())
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_DEPARTMENT_NOT_FOUND));

        String oldDepartmentId = employee.getDepartmentId();
        if (request.getNewDepartmentId().equals(oldDepartmentId)) {
            throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "员工已在目标部门");
        }

        Department oldDept = oldDepartmentId != null ? 
                departmentRepository.findById(oldDepartmentId).orElse(null) : null;

        employee.setDepartmentId(request.getNewDepartmentId());
        employee.setUpdatedAt(LocalDateTime.now());
        Employee saved = employeeRepository.save(employee);

        recordTransferHistory(saved, 
                oldDepartmentId, oldDept != null ? oldDept.getName() : null,
                targetDept.getId(), targetDept.getName(),
                request.getReason());

        operationLogService.logTransfer("EMPLOYEE", saved.getId(), saved.getName(),
                "原部门: " + (oldDept != null ? oldDept.getName() : "无"),
                "新部门: " + targetDept.getName());

        return toDTO(saved);
    }

    public void recordTransferHistory(Employee employee, 
                                       String fromDeptId, String fromDeptName,
                                       String toDeptId, String toDeptName,
                                       String reason) {
        TransferHistory history = new TransferHistory();
        history.setId(IdGenerator.generateTransferHistoryId());
        history.setEmployeeId(employee.getId());
        history.setEmployeeName(employee.getName());
        history.setFromDepartmentId(fromDeptId);
        history.setFromDepartmentName(fromDeptName);
        history.setToDepartmentId(toDeptId);
        history.setToDepartmentName(toDeptName);
        history.setReason(reason);
        history.setTransferTime(LocalDateTime.now());
        transferHistoryRepository.save(history);
    }

    public ReportingLineDTO getReportingLine(String employeeId) {
        Employee employee = employeeRepository.findById(employeeId)
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND));

        ReportingLineDTO dto = new ReportingLineDTO();
        dto.setEmployeeId(employee.getId());
        dto.setEmployeeName(employee.getName());

        List<ReportingLineDTO.ManagerNode> managers = new ArrayList<>();
        Set<String> visited = new HashSet<>();

        String currentManagerId = employee.getManagerId();
        int level = 1;

        while (currentManagerId != null && !currentManagerId.isEmpty()) {
            if (visited.contains(currentManagerId)) {
                break;
            }
            visited.add(currentManagerId);

            Employee manager = employeeRepository.findById(currentManagerId).orElse(null);
            if (manager == null) {
                break;
            }

            ReportingLineDTO.ManagerNode node = new ReportingLineDTO.ManagerNode();
            node.setId(manager.getId());
            node.setName(manager.getName());
            node.setLevel(level);

            if (manager.getDepartmentId() != null) {
                Department dept = departmentRepository.findById(manager.getDepartmentId()).orElse(null);
                if (dept != null) {
                    node.setDepartmentName(dept.getName());
                }
            }

            managers.add(node);
            currentManagerId = manager.getManagerId();
            level++;
        }

        dto.setManagers(managers);

        List<ReportingLineDTO.DepartmentNode> deptPath = new ArrayList<>();
        String currentDeptId = employee.getDepartmentId();
        Set<String> deptVisited = new HashSet<>();
        int deptLevel = 1;

        while (currentDeptId != null) {
            if (deptVisited.contains(currentDeptId)) {
                break;
            }
            deptVisited.add(currentDeptId);

            Department dept = departmentRepository.findById(currentDeptId).orElse(null);
            if (dept == null) {
                break;
            }

            ReportingLineDTO.DepartmentNode node = new ReportingLineDTO.DepartmentNode();
            node.setId(dept.getId());
            node.setName(dept.getName());
            node.setLevel(deptLevel);
            deptPath.add(0, node);

            currentDeptId = dept.getParentId();
            deptLevel++;
        }

        dto.setDepartmentPath(deptPath);

        return dto;
    }

    public List<TransferHistoryDTO> getTransferHistory(String employeeId) {
        return transferHistoryRepository.findByEmployeeId(employeeId).stream()
                .map(this::toHistoryDTO)
                .collect(Collectors.toList());
    }

    private boolean hasCircularReference(String newManagerId, String employeeId) {
        if (newManagerId == null || newManagerId.isEmpty()) {
            return false;
        }

        Set<String> visited = new HashSet<>();
        String current = newManagerId;

        while (current != null && !current.isEmpty()) {
            if (employeeId != null && employeeId.equals(current)) {
                return true;
            }
            if (visited.contains(current)) {
                return true;
            }
            visited.add(current);

            Employee emp = employeeRepository.findById(current).orElse(null);
            if (emp == null) {
                break;
            }
            current = emp.getManagerId();
        }
        return false;
    }

    private EmployeeDTO toDTO(Employee entity) {
        EmployeeDTO dto = new EmployeeDTO();
        BeanUtils.copyProperties(entity, dto);

        if (entity.getDepartmentId() != null) {
            Department dept = departmentRepository.findById(entity.getDepartmentId()).orElse(null);
            if (dept != null) {
                dto.setDepartmentName(dept.getName());
            }
        }

        if (entity.getManagerId() != null) {
            Employee manager = employeeRepository.findById(entity.getManagerId()).orElse(null);
            if (manager != null) {
                dto.setManagerName(manager.getName());
            }
        }

        List<String> teamNames = new ArrayList<>();
        for (String teamId : entity.getVirtualTeamIds()) {
            VirtualTeam team = virtualTeamRepository.findById(teamId).orElse(null);
            if (team != null) {
                teamNames.add(team.getName());
            }
        }
        dto.setVirtualTeamNames(teamNames);

        return dto;
    }

    private TransferHistoryDTO toHistoryDTO(TransferHistory entity) {
        TransferHistoryDTO dto = new TransferHistoryDTO();
        BeanUtils.copyProperties(entity, dto);
        return dto;
    }
}
