package com.orgchart.common.dto.request;

import java.util.ArrayList;
import java.util.List;

public class CreateVirtualTeamRequest {

    private String name;
    private String description;
    private List<String> memberIds = new ArrayList<>();

    public CreateVirtualTeamRequest() {
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public List<String> getMemberIds() {
        return memberIds;
    }

    public void setMemberIds(List<String> memberIds) {
        this.memberIds = memberIds;
    }
}
