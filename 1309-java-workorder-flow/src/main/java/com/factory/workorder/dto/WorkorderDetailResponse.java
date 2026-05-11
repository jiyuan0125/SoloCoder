package com.factory.workorder.dto;

import com.factory.workorder.entity.Workorder;
import com.factory.workorder.entity.WorkorderLog;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class WorkorderDetailResponse {

    private Workorder workorder;
    private List<WorkorderLog> logs;
}
