package com.employee.common.dto;

public class SearchRequest {
    private String keyword;
    private boolean includeResigned;

    public String getKeyword() {
        return keyword;
    }

    public void setKeyword(String keyword) {
        this.keyword = keyword;
    }

    public boolean isIncludeResigned() {
        return includeResigned;
    }

    public void setIncludeResigned(boolean includeResigned) {
        this.includeResigned = includeResigned;
    }
}
