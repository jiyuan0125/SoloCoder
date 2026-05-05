package com.workshift.common.dto;

import com.workshift.common.enums.ShiftType;
import com.workshift.common.enums.SwapStatus;
import java.time.LocalDate;
import java.time.LocalDateTime;

public class ShiftSwapDTO {
    private String id;
    private String requesterEmployeeId;
    private String targetEmployeeId;
    private LocalDate requesterDate;
    private ShiftType requesterShiftType;
    private LocalDate targetDate;
    private ShiftType targetShiftType;
    private SwapStatus status;
    private String requesterSignature;
    private LocalDateTime requesterSignatureTime;
    private String targetSignature;
    private LocalDateTime targetSignatureTime;
    private LocalDateTime createTime;
    private String rejectReason;

    public ShiftSwapDTO() {
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getRequesterEmployeeId() {
        return requesterEmployeeId;
    }

    public void setRequesterEmployeeId(String requesterEmployeeId) {
        this.requesterEmployeeId = requesterEmployeeId;
    }

    public String getTargetEmployeeId() {
        return targetEmployeeId;
    }

    public void setTargetEmployeeId(String targetEmployeeId) {
        this.targetEmployeeId = targetEmployeeId;
    }

    public LocalDate getRequesterDate() {
        return requesterDate;
    }

    public void setRequesterDate(LocalDate requesterDate) {
        this.requesterDate = requesterDate;
    }

    public ShiftType getRequesterShiftType() {
        return requesterShiftType;
    }

    public void setRequesterShiftType(ShiftType requesterShiftType) {
        this.requesterShiftType = requesterShiftType;
    }

    public LocalDate getTargetDate() {
        return targetDate;
    }

    public void setTargetDate(LocalDate targetDate) {
        this.targetDate = targetDate;
    }

    public ShiftType getTargetShiftType() {
        return targetShiftType;
    }

    public void setTargetShiftType(ShiftType targetShiftType) {
        this.targetShiftType = targetShiftType;
    }

    public SwapStatus getStatus() {
        return status;
    }

    public void setStatus(SwapStatus status) {
        this.status = status;
    }

    public String getRequesterSignature() {
        return requesterSignature;
    }

    public void setRequesterSignature(String requesterSignature) {
        this.requesterSignature = requesterSignature;
    }

    public LocalDateTime getRequesterSignatureTime() {
        return requesterSignatureTime;
    }

    public void setRequesterSignatureTime(LocalDateTime requesterSignatureTime) {
        this.requesterSignatureTime = requesterSignatureTime;
    }

    public String getTargetSignature() {
        return targetSignature;
    }

    public void setTargetSignature(String targetSignature) {
        this.targetSignature = targetSignature;
    }

    public LocalDateTime getTargetSignatureTime() {
        return targetSignatureTime;
    }

    public void setTargetSignatureTime(LocalDateTime targetSignatureTime) {
        this.targetSignatureTime = targetSignatureTime;
    }

    public LocalDateTime getCreateTime() {
        return createTime;
    }

    public void setCreateTime(LocalDateTime createTime) {
        this.createTime = createTime;
    }

    public String getRejectReason() {
        return rejectReason;
    }

    public void setRejectReason(String rejectReason) {
        this.rejectReason = rejectReason;
    }
}