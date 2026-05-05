package com.orgchart.common.dto.request;

import java.util.ArrayList;
import java.util.List;

public class UpdateEmployeeRequest {

    private String name;
    private String email;
    private String phone;
    private String departmentId;
    private String managerId;
    private List<String> virtualTeamIds = new ArrayList<>();

    public UpdateEmployeeRequest() {
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

    public String getManagerId() {
        return managerId;
    }

    public void setManagerId(String managerId) {
        this.managerId = managerId;
    }

    public List<String> getVirtualTeamIds() {
        return virtualTeamIds;
    }

    public void setVirtualTeamIds(List<String> virtualTeamIds) {
        this.virtualTeamIds = virtualTeamIds;
    }
}
