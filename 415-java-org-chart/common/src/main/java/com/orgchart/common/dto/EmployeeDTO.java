package com.orgchart.common.dto;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

public class EmployeeDTO {

    private String id;
    private String name;
    private String email;
    private String phone;
    private String departmentId;
    private String departmentName;
    private String managerId;
    private String managerName;
    private List<String> virtualTeamIds = new ArrayList<>();
    private List<String> virtualTeamNames = new ArrayList<>();
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    public EmployeeDTO() {
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getPhone() {
        return phone;
    }

    public void setPhone(String phone) {
        this.phone = phone;
    }

    public String getDepartmentId() {
        return departmentId;
    }

    public void setDepartmentId(String departmentId) {
        this.departmentId = departmentId;
    }

    public String getDepartmentName() {
        return departmentName;
    }

    public void setDepartmentName(String departmentName) {
        this.departmentName = departmentName;
    }

    public String getManagerId() {
        return managerId;
    }

    public void setManagerId(String managerId) {
        this.managerId = managerId;
    }

    public String getManagerName() {
        return managerName;
    }

    public void setManagerName(String managerName) {
        this.managerName = managerName;
    }

    public List<String> getVirtualTeamIds() {
        return virtualTeamIds;
    }

    public void setVirtualTeamIds(List<String> virtualTeamIds) {
        this.virtualTeamIds = virtualTeamIds;
    }

    public List<String> getVirtualTeamNames() {
        return virtualTeamNames;
    }

    public void setVirtualTeamNames(List<String> virtualTeamNames) {
        this.virtualTeamNames = virtualTeamNames;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }

    public LocalDateTime getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(LocalDateTime updatedAt) {
        this.updatedAt = updatedAt;
    }
}
