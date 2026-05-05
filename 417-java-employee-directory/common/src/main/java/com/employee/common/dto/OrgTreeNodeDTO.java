package com.employee.common.dto;

import java.util.ArrayList;
import java.util.List;

public class OrgTreeNodeDTO {
    private String name;
    private String type;
    private int employeeCount;
    private List<OrgTreeNodeDTO> children;

    public OrgTreeNodeDTO() {
        this.children = new ArrayList<>();
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public int getEmployeeCount() {
        return employeeCount;
    }

    public void setEmployeeCount(int employeeCount) {
        this.employeeCount = employeeCount;
    }

    public List<OrgTreeNodeDTO> getChildren() {
        return children;
    }

    public void setChildren(List<OrgTreeNodeDTO> children) {
        this.children = children;
    }

    public void addChild(OrgTreeNodeDTO child) {
        this.children.add(child);
    }
}
