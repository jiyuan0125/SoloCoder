package com.orgchart.server.service;

import com.orgchart.common.dto.DepartmentDTO;
import com.orgchart.common.dto.request.CreateDepartmentRequest;
import com.orgchart.common.dto.request.MergeDepartmentRequest;
import com.orgchart.common.dto.request.UpdateDepartmentRequest;
import com.orgchart.common.enums.ErrorCode;
import com.orgchart.server.entity.Department;
import com.orgchart.server.entity.Employee;
import com.orgchart.server.exception.BusinessException;
import com.orgchart.server.repository.DepartmentRepository;
import com.orgchart.server.repository.EmployeeRepository;
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
public class DepartmentService {

    private final DepartmentRepository departmentRepository;
    private final EmployeeRepository employeeRepository;
    private final EmployeeService employeeService;
    private final OperationLogService operationLogService;

    public DepartmentService(DepartmentRepository departmentRepository,
                              EmployeeRepository employeeRepository,
                              EmployeeService employeeService,
                              OperationLogService operationLogService) {
        this.departmentRepository = departmentRepository;
        this.employeeRepository = employeeRepository;
        this.employeeService = employeeService;
        this.operationLogService = operationLogService;
    }

    public DepartmentDTO createDepartment(CreateDepartmentRequest request) {
        if (request.getName() == null || request.getName().trim().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "部门名称不能为空");
        }

        String parentId = request.getParentId();
        Department parent = null;

        if (parentId != null && !parentId.isEmpty()) {
            parent = departmentRepository.findById(parentId)
                    .orElseThrow(() -> new BusinessException(ErrorCode.PARENT_DEPARTMENT_NOT_FOUND));
        }

        if (departmentRepository.existsByNameAndParentId(request.getName(), parentId)) {
            throw new BusinessException(ErrorCode.DEPARTMENT_NAME_DUPLICATE);
        }

        Department department = new Department();
        department.setId(IdGenerator.generateDepartmentId());
        department.setName(request.getName());
        department.setParentId(parentId);
        department.setLevel(parent != null ? parent.getLevel() + 1 : 1);
        department.setCreatedAt(LocalDateTime.now());
        department.setUpdatedAt(LocalDateTime.now());

        Department saved = departmentRepository.save(department);

        if (parent != null) {
            parent.getChildIds().add(saved.getId());
            departmentRepository.save(parent);
        }

        operationLogService.logCreate("DEPARTMENT", saved.getId(), saved.getName(), 
                "名称: " + saved.getName() + ", 父部门: " + (parent != null ? parent.getName() : "无"));

        return toDTO(saved);
    }

    public DepartmentDTO updateDepartment(String id, UpdateDepartmentRequest request) {
        Department department = departmentRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.DEPARTMENT_NOT_FOUND));

        String oldName = department.getName();
        String oldParentId = department.getParentId();
        Department oldParent = oldParentId != null ? departmentRepository.findById(oldParentId).orElse(null) : null;

        StringBuilder changes = new StringBuilder();

        if (request.getName() != null && !request.getName().trim().isEmpty()) {
            String newParentId = request.getParentId() != null ? request.getParentId() : department.getParentId();
            if (!request.getName().equals(department.getName())) {
                if (departmentRepository.existsByNameAndParentId(request.getName(), newParentId)) {
                    throw new BusinessException(ErrorCode.DEPARTMENT_NAME_DUPLICATE);
                }
                changes.append("名称: ").append(oldName).append(" -> ").append(request.getName()).append("; ");
                department.setName(request.getName());
            }
        }

        if (request.getParentId() != null && !request.getParentId().equals(id)) {
            throw new BusinessException(ErrorCode.DEPARTMENT_CANNOT_BE_OWN_PARENT);
        }

        if (request.getParentId() != null && !request.getParentId().isEmpty()) {
            if (oldParentId == null || !oldParentId.equals(request.getParentId())) {
                if (hasCircularReference(request.getParentId(), id)) {
                    throw new BusinessException(ErrorCode.CIRCULAR_REFERENCE);
                }

                Department newParent = departmentRepository.findById(request.getParentId())
                        .orElseThrow(() -> new BusinessException(ErrorCode.PARENT_DEPARTMENT_NOT_FOUND));

                if (department.getName() != null) {
                    if (departmentRepository.existsByNameAndParentId(department.getName(), request.getParentId())) {
                        throw new BusinessException(ErrorCode.DEPARTMENT_NAME_DUPLICATE);
                    }
                }

                if (oldParent != null) {
                    oldParent.getChildIds().remove(id);
                    departmentRepository.save(oldParent);
                }

                newParent.getChildIds().add(id);
                departmentRepository.save(newParent);

                department.setParentId(request.getParentId());
                updateDepartmentLevel(department, newParent.getLevel() + 1);

                String newParentName = newParent.getName();
                String oldParentName = oldParent != null ? oldParent.getName() : "无";
                changes.append("父部门: ").append(oldParentName).append(" -> ").append(newParentName);
            }
        } else if (request.getParentId() != null && request.getParentId().isEmpty()) {
            if (oldParentId != null && !oldParentId.isEmpty()) {
                if (oldParent != null) {
                    oldParent.getChildIds().remove(id);
                    departmentRepository.save(oldParent);
                }
                department.setParentId(null);
                updateDepartmentLevel(department, 1);
                changes.append("父部门: ").append(oldParent != null ? oldParent.getName() : "无").append(" -> 无");
            }
        }

        department.setUpdatedAt(LocalDateTime.now());
        Department saved = departmentRepository.save(department);

        if (!changes.isEmpty()) {
            operationLogService.logUpdate("DEPARTMENT", saved.getId(), saved.getName(),
                    "原信息: " + oldName, "变更: " + changes);
        }

        return toDTO(saved);
    }

    private boolean hasCircularReference(String newParentId, String departmentId) {
        Set<String> visited = new HashSet<>();
        String current = newParentId;

        while (current != null) {
            if (departmentId.equals(current)) {
                return true;
            }
            if (visited.contains(current)) {
                return true;
            }
            visited.add(current);

            Department dept = departmentRepository.findById(current).orElse(null);
            if (dept == null) {
                break;
            }
            current = dept.getParentId();
        }
        return false;
    }

    private void updateDepartmentLevel(Department department, int newLevel) {
        department.setLevel(newLevel);
        departmentRepository.save(department);

        List<String> childIds = new ArrayList<>(department.getChildIds());
        for (String childId : childIds) {
            Department child = departmentRepository.findById(childId).orElse(null);
            if (child != null) {
                updateDepartmentLevel(child, newLevel + 1);
            }
        }
    }

    public void deleteDepartment(String id) {
        Department department = departmentRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.DEPARTMENT_NOT_FOUND));

        List<Employee> employees = employeeRepository.findByDepartmentId(id);
        if (!employees.isEmpty()) {
            throw new BusinessException(ErrorCode.DEPARTMENT_HAS_EMPLOYEES);
        }

        List<Department> children = departmentRepository.findByParentId(id);
        if (!children.isEmpty()) {
            throw new BusinessException(ErrorCode.DEPARTMENT_HAS_EMPLOYEES.getCode(), 
                    "部门下存在子部门，请先处理子部门");
        }

        String parentId = department.getParentId();
        if (parentId != null) {
            Department parent = departmentRepository.findById(parentId).orElse(null);
            if (parent != null) {
                parent.getChildIds().remove(id);
                departmentRepository.save(parent);
            }
        }

        departmentRepository.deleteById(id);

        operationLogService.logDelete("DEPARTMENT", id, department.getName(), 
                "删除部门: " + department.getName());
    }

    public DepartmentDTO getDepartmentById(String id) {
        Department department = departmentRepository.findById(id)
                .orElseThrow(() -> new BusinessException(ErrorCode.DEPARTMENT_NOT_FOUND));
        return toDTO(department);
    }

    public List<DepartmentDTO> getAllDepartments() {
        List<Department> departments = departmentRepository.findAll();
        return departments.stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public List<DepartmentDTO> getDepartmentTree() {
        List<Department> roots = departmentRepository.findByParentId(null);
        return roots.stream()
                .map(this::buildDepartmentTree)
                .collect(Collectors.toList());
    }

    private DepartmentDTO buildDepartmentTree(Department department) {
        DepartmentDTO dto = toDTO(department);
        List<DepartmentDTO> children = department.getChildIds().stream()
                .map(childId -> departmentRepository.findById(childId).orElse(null))
                .filter(java.util.Objects::nonNull)
                .map(this::buildDepartmentTree)
                .collect(Collectors.toList());
        dto.setChildren(children);
        return dto;
    }

    public void mergeDepartment(String sourceId, MergeDepartmentRequest request) {
        Department source = departmentRepository.findById(sourceId)
                .orElseThrow(() -> new BusinessException(ErrorCode.DEPARTMENT_NOT_FOUND));

        Department target = departmentRepository.findById(request.getTargetDepartmentId())
                .orElseThrow(() -> new BusinessException(ErrorCode.DEPARTMENT_NOT_FOUND));

        if (sourceId.equals(request.getTargetDepartmentId())) {
            throw new BusinessException(ErrorCode.PARAM_ERROR.getCode(), "不能合并到自己");
        }

        if (isDescendant(target.getId(), source.getId())) {
            throw new BusinessException(ErrorCode.CIRCULAR_REFERENCE);
        }

        List<Employee> sourceEmployees = employeeRepository.findByDepartmentId(sourceId);
        for (Employee employee : sourceEmployees) {
            String oldDeptName = employee.getDepartmentId() != null ? 
                    departmentRepository.findById(employee.getDepartmentId()).map(Department::getName).orElse(null) : null;
            employee.setDepartmentId(target.getId());
            employee.setUpdatedAt(LocalDateTime.now());
            employeeRepository.save(employee);

            employeeService.recordTransferHistory(employee, sourceId, source.getName(), 
                    target.getId(), target.getName(), "部门合并");
        }

        List<String> childIds = new ArrayList<>(source.getChildIds());
        for (String childId : childIds) {
            Department child = departmentRepository.findById(childId).orElse(null);
            if (child != null) {
                child.setParentId(target.getId());
                child.setLevel(target.getLevel() + 1);
                child.setUpdatedAt(LocalDateTime.now());
                departmentRepository.save(child);

                target.getChildIds().add(childId);
                source.getChildIds().remove(childId);
            }
        }

        if (source.getParentId() != null) {
            Department parent = departmentRepository.findById(source.getParentId()).orElse(null);
            if (parent != null) {
                parent.getChildIds().remove(sourceId);
                departmentRepository.save(parent);
            }
        }

        departmentRepository.save(target);
        departmentRepository.deleteById(sourceId);

        operationLogService.logMerge("DEPARTMENT", sourceId, source.getName(),
                "源部门: " + source.getName(), "目标部门: " + target.getName());
    }

    private boolean isDescendant(String targetId, String ancestorId) {
        Department current = departmentRepository.findById(targetId).orElse(null);
        while (current != null) {
            if (ancestorId.equals(current.getParentId())) {
                return true;
            }
            current = current.getParentId() != null ? 
                    departmentRepository.findById(current.getParentId()).orElse(null) : null;
        }
        return false;
    }

    private DepartmentDTO toDTO(Department entity) {
        DepartmentDTO dto = new DepartmentDTO();
        BeanUtils.copyProperties(entity, dto);
        if (entity.getParentId() != null) {
            Department parent = departmentRepository.findById(entity.getParentId()).orElse(null);
            if (parent != null) {
                dto.setParentName(parent.getName());
            }
        }
        return dto;
    }
}
