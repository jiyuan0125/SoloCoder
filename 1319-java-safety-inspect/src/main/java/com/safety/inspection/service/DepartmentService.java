package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.common.BusinessException;
import com.safety.inspection.dto.DepartmentDTO;
import com.safety.inspection.entity.Department;
import com.safety.inspection.mapper.DepartmentMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.List;

@Service
@RequiredArgsConstructor
public class DepartmentService extends ServiceImpl<DepartmentMapper, Department> {

    private final DepartmentMapper departmentMapper;

    public List<Department> getAllDepartments(String deptName) {
        LambdaQueryWrapper<Department> wrapper = new LambdaQueryWrapper<>();
        if (StringUtils.hasText(deptName)) {
            wrapper.like(Department::getDeptName, deptName);
        }
        wrapper.orderByAsc(Department::getSort);
        return departmentMapper.selectList(wrapper);
    }

    public Department getDepartmentById(Long id) {
        return departmentMapper.selectById(id);
    }

    public void createDepartment(DepartmentDTO dto) {
        Department dept = new Department();
        dept.setDeptName(dto.getDeptName());
        dept.setParentId(dto.getParentId() != null ? dto.getParentId() : 0L);
        dept.setLeaderId(dto.getLeaderId());
        dept.setSort(dto.getSort() != null ? dto.getSort() : 0);
        dept.setStatus(dto.getStatus() != null ? dto.getStatus() : 1);
        departmentMapper.insert(dept);
    }

    public void updateDepartment(DepartmentDTO dto) {
        Department dept = departmentMapper.selectById(dto.getId());
        if (dept == null) {
            throw new BusinessException("部门不存在");
        }
        dept.setDeptName(dto.getDeptName());
        dept.setParentId(dto.getParentId());
        dept.setLeaderId(dto.getLeaderId());
        dept.setSort(dto.getSort());
        dept.setStatus(dto.getStatus());
        departmentMapper.updateById(dept);
    }

    public void deleteDepartment(Long id) {
        departmentMapper.deleteById(id);
    }
}
