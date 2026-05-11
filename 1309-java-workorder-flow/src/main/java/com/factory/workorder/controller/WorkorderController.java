package com.factory.workorder.controller;

import com.factory.workorder.dto.*;
import com.factory.workorder.entity.Workorder;
import com.factory.workorder.enums.WorkorderPriority;
import com.factory.workorder.enums.WorkorderStatus;
import com.factory.workorder.service.WorkorderService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/api/workorders")
@RequiredArgsConstructor
public class WorkorderController {

    private final WorkorderService workorderService;

    @PostMapping
    public ResponseEntity<Workorder> createWorkorder(@Valid @RequestBody CreateWorkorderRequest request) {
        Workorder workorder = workorderService.createWorkorder(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(workorder);
    }

    @PostMapping("/{workorderId}/assign")
    public ResponseEntity<Workorder> assignWorkorder(
            @PathVariable Long workorderId,
            @Valid @RequestBody AssignWorkorderRequest request) {
        Workorder workorder = workorderService.assignWorkorder(workorderId, request);
        return ResponseEntity.ok(workorder);
    }

    @PostMapping("/{workorderId}/submit-acceptance")
    public ResponseEntity<Workorder> submitAcceptance(
            @PathVariable Long workorderId,
            @Valid @RequestBody SubmitAcceptanceRequest request) {
        Workorder workorder = workorderService.submitAcceptance(workorderId, request);
        return ResponseEntity.ok(workorder);
    }

    @PostMapping("/{workorderId}/accept")
    public ResponseEntity<Workorder> acceptWorkorder(
            @PathVariable Long workorderId,
            @Valid @RequestBody AcceptanceRequest request) {
        Workorder workorder = workorderService.acceptWorkorder(workorderId, request);
        return ResponseEntity.ok(workorder);
    }

    @PostMapping("/{workorderId}/reject")
    public ResponseEntity<Workorder> rejectWorkorder(
            @PathVariable Long workorderId,
            @Valid @RequestBody AcceptanceRequest request) {
        Workorder workorder = workorderService.rejectWorkorder(workorderId, request);
        return ResponseEntity.ok(workorder);
    }

    @GetMapping
    public ResponseEntity<List<Workorder>> getAllWorkorders() {
        List<Workorder> workorders = workorderService.getAllWorkorders();
        return ResponseEntity.ok(workorders);
    }

    @GetMapping("/status/{status}")
    public ResponseEntity<List<Workorder>> getWorkordersByStatus(@PathVariable String status) {
        WorkorderStatus workorderStatus = parseStatus(status);
        List<Workorder> workorders = workorderService.getWorkordersByStatus(workorderStatus);
        return ResponseEntity.ok(workorders);
    }

    @GetMapping("/priority/{priority}")
    public ResponseEntity<List<Workorder>> getWorkordersByPriority(@PathVariable String priority) {
        WorkorderPriority workorderPriority = parsePriority(priority);
        List<Workorder> workorders = workorderService.getWorkordersByPriority(workorderPriority);
        return ResponseEntity.ok(workorders);
    }

    @GetMapping("/handler/{handler}")
    public ResponseEntity<List<Workorder>> getWorkordersByHandler(@PathVariable String handler) {
        List<Workorder> workorders = workorderService.getWorkordersByHandler(handler);
        return ResponseEntity.ok(workorders);
    }

    @GetMapping("/{workorderId}")
    public ResponseEntity<WorkorderDetailResponse> getWorkorderDetail(@PathVariable Long workorderId) {
        WorkorderDetailResponse detail = workorderService.getWorkorderDetail(workorderId);
        return ResponseEntity.ok(detail);
    }

    private WorkorderStatus parseStatus(String status) {
        if ("待分派".equals(status) || "PENDING_ASSIGN".equalsIgnoreCase(status)) {
            return WorkorderStatus.PENDING_ASSIGN;
        }
        if ("处理中".equals(status) || "IN_PROGRESS".equalsIgnoreCase(status)) {
            return WorkorderStatus.IN_PROGRESS;
        }
        if ("待验收".equals(status) || "PENDING_ACCEPTANCE".equalsIgnoreCase(status)) {
            return WorkorderStatus.PENDING_ACCEPTANCE;
        }
        if ("已完成".equals(status) || "COMPLETED".equalsIgnoreCase(status)) {
            return WorkorderStatus.COMPLETED;
        }
        return WorkorderStatus.valueOf(status.toUpperCase());
    }

    private WorkorderPriority parsePriority(String priority) {
        if ("普通".equals(priority) || "NORMAL".equalsIgnoreCase(priority)) {
            return WorkorderPriority.NORMAL;
        }
        if ("紧急".equals(priority) || "URGENT".equalsIgnoreCase(priority)) {
            return WorkorderPriority.URGENT;
        }
        if ("特急".equals(priority) || "CRITICAL".equalsIgnoreCase(priority)) {
            return WorkorderPriority.CRITICAL;
        }
        return WorkorderPriority.valueOf(priority.toUpperCase());
    }
}
