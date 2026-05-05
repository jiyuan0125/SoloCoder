package com.orgchart.common.dto;

import java.util.ArrayList;
import java.util.List;

public class ReportingLineDTO {

    private String employeeId;
    private String employeeName;
    private List<ManagerNode> managers = new ArrayList<>();
    private List<DepartmentNode> departmentPath = new ArrayList<>();

    public ReportingLineDTO() {
    }

    public String getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(String employeeId) {
        this.employeeId = employeeId;
    }

    public String getEmployeeName() {
        return employeeName;
    }

    public void setEmployeeName(String employeeName) {
        this.employeeName = employeeName;
    }

    public List<ManagerNode> getManagers() {
        return managers;
    }

    public void setManagers(List<ManagerNode> managers) {
        this.managers = managers;
    }

    public List<DepartmentNode> getDepartmentPath() {
        return departmentPath;
    }

    public void setDepartmentPath(List<DepartmentNode> departmentPath) {
        this.departmentPath = departmentPath;
    }

    public static class ManagerNode {
        private String id;
        private String name;
        private String departmentName;
        private int level;

        public ManagerNode() {
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

        public String getDepartmentName() {
            return departmentName;
        }

        public void setDepartmentName(String departmentName) {
            this.departmentName = departmentName;
        }

        public int getLevel() {
            return level;
        }

        public void setLevel(int level) {
            this.level = level;
        }
    }

    public static class DepartmentNode {
        private String id;
        private String name;
        private int level;

        public DepartmentNode() {
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

        public int getLevel() {
            return level;
        }

        public void setLevel(int level) {
            this.level = level;
        }
    }
}
