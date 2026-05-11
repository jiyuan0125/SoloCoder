package com.factory.workorder.service;

import com.factory.workorder.dto.*;
import com.factory.workorder.entity.Workorder;
import com.factory.workorder.entity.WorkorderLog;
import com.factory.workorder.enums.OperationType;
import com.factory.workorder.enums.WorkorderPriority;
import com.factory.workorder.enums.WorkorderStatus;

import java.util.List;

public interface WorkorderService {

    Workorder createWorkorder(CreateWorkorderRequest request);

    Workorder assignWorkorder(Long workorderId, AssignWorkorderRequest request);

    Workorder submitAcceptance(Long workorderId, SubmitAcceptanceRequest request);

    Workorder acceptWorkorder(Long workorderId, AcceptanceRequest request);

    Workorder rejectWorkorder(Long workorderId, AcceptanceRequest request);

    List<Workorder> getWorkordersByStatus(WorkorderStatus status);

    List<Workorder> getWorkordersByPriority(WorkorderPriority priority);

    List<Workorder> getWorkordersByHandler(String handler);

    WorkorderDetailResponse getWorkorderDetail(Long workorderId);

    List<Workorder> getAllWorkorders();
}
