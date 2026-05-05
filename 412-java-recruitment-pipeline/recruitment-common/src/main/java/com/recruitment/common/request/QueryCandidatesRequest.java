package com.recruitment.common.request;

import com.recruitment.common.enums.SourceChannel;
import com.recruitment.common.enums.Stage;

public class QueryCandidatesRequest {
    private String position;
    private Stage stage;
    private SourceChannel sourceChannel;
    private Integer page;
    private Integer size;

    public QueryCandidatesRequest() {
        this.page = 0;
        this.size = 20;
    }

    public String getPosition() {
        return position;
    }

    public void setPosition(String position) {
        this.position = position;
    }

    public Stage getStage() {
        return stage;
    }

    public void setStage(Stage stage) {
        this.stage = stage;
    }

    public SourceChannel getSourceChannel() {
        return sourceChannel;
    }

    public void setSourceChannel(SourceChannel sourceChannel) {
        this.sourceChannel = sourceChannel;
    }

    public Integer getPage() {
        return page;
    }

    public void setPage(Integer page) {
        this.page = page;
    }

    public Integer getSize() {
        return size;
    }

    public void setSize(Integer size) {
        this.size = size;
    }
}
