package com.hospital.prescription.service;

import com.hospital.prescription.dto.PrescriptionItemRequest;
import com.hospital.prescription.dto.PrescriptionRequest;
import com.hospital.prescription.dto.PrescriptionValidationResult;
import com.hospital.prescription.dto.ValidationResult;
import com.hospital.prescription.entity.*;
import com.hospital.prescription.enums.PrescriptionStatus;
import com.hospital.prescription.enums.SeverityLevel;
import com.hospital.prescription.repository.*;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;

@Service
@RequiredArgsConstructor
public class PrescriptionService {

    private final PrescriptionRepository prescriptionRepository;
    private final PrescriptionValidationService validationService;
    private final DrugRepository drugRepository;
    private final PatientRepository patientRepository;

    @Transactional
    public Prescription createPrescription(PrescriptionRequest request) {
        PrescriptionValidationResult validationResult = validationService.validate(request);

        if (validationResult.isHasBlockingErrors()) {
            throw new RuntimeException("处方存在严重问题，无法提交：" + 
                validationResult.getBlockingValidations().get(0).getMessage());
        }

        Patient patient = patientRepository.findById(request.getPatientId())
                .orElseThrow(() -> new RuntimeException("患者不存在"));

        Prescription prescription = new Prescription();
        prescription.setPatient(patient);
        prescription.setDiagnosis(request.getDiagnosis());
        prescription.setDoctorName(request.getDoctorName());
        prescription.setDepartment(request.getDepartment());
        prescription.setStatus(PrescriptionStatus.DRAFT);

        for (PrescriptionItemRequest itemReq : request.getItems()) {
            Drug drug = drugRepository.findById(itemReq.getDrugId())
                    .orElseThrow(() -> new RuntimeException("药品不存在"));
            
            PrescriptionItem item = new PrescriptionItem();
            item.setDrug(drug);
            item.setDrugName(drug.getName());
            item.setDosageForm(drug.getDosageForm());
            item.setSingleDose(itemReq.getSingleDose());
            item.setDoseUnit(drug.getDoseUnit());
            item.setFrequencyPerDay(itemReq.getFrequencyPerDay());
            item.setDurationDays(itemReq.getDurationDays());
            item.setSolventType(itemReq.getSolventType());
            prescription.addItem(item);
        }

        saveValidations(prescription, validationResult.getValidations());

        boolean hasOverDose = validationResult.getValidations().stream()
                .anyMatch(v -> "DOSAGE".equals(v.getValidationType()) && v.getSeverity() == SeverityLevel.MODERATE);
        
        if (hasOverDose && request.isOverDoseConfirmed()) {
            prescription.setOverDose(true);
            prescription.setOverDoseReason("医生确认超剂量用药");
        }

        return prescriptionRepository.save(prescription);
    }

    @Transactional
    public Prescription submitPrescription(Long prescriptionId, String overDoseReason) {
        Prescription prescription = prescriptionRepository.findById(prescriptionId)
                .orElseThrow(() -> new RuntimeException("处方不存在"));

        if (prescription.getStatus() != PrescriptionStatus.DRAFT) {
            throw new RuntimeException("只有草稿状态的处方可以提交");
        }

        if (prescription.isOverDose() && overDoseReason != null && !overDoseReason.isEmpty()) {
            prescription.setOverDoseReason(overDoseReason);
        }

        prescription.setStatus(PrescriptionStatus.SUBMITTED);
        prescription.setSubmittedAt(LocalDateTime.now());

        PrescriptionSignature signature = new PrescriptionSignature();
        signature.setSignerName(prescription.getDoctorName());
        signature.setSignerRole("开方医生");
        prescription.addSignature(signature);

        return prescriptionRepository.save(prescription);
    }

    @Transactional
    public Prescription pharmacistReview(Long prescriptionId, String pharmacistName,
                                         boolean approve, String comments, 
                                         List<PrescriptionItemRequest> modifiedItems) {
        Prescription prescription = prescriptionRepository.findById(prescriptionId)
                .orElseThrow(() -> new RuntimeException("处方不存在"));

        if (prescription.getStatus() != PrescriptionStatus.SUBMITTED) {
            throw new RuntimeException("只有已提交的处方可以审核");
        }

        prescription.setPharmacistName(pharmacistName);

        if (approve) {
            if (modifiedItems != null && !modifiedItems.isEmpty()) {
                updatePrescriptionItems(prescription, modifiedItems);
                prescription.setStatus(PrescriptionStatus.MODIFIED_BY_PHARMACIST);
            } else {
                prescription.setStatus(PrescriptionStatus.APPROVED);
                prescription.setApprovedAt(LocalDateTime.now());
                
                if (prescription.hasPsychotropicOrNarcoticDrugs()) {
                    PrescriptionSignature signature = new PrescriptionSignature();
                    signature.setSignerName(pharmacistName);
                    signature.setSignerRole("审核药师");
                    prescription.addSignature(signature);
                }
            }
        } else {
            prescription.setStatus(PrescriptionStatus.REJECTED);
        }

        return prescriptionRepository.save(prescription);
    }

    @Transactional
    public Prescription doctorConfirmModified(Long prescriptionId, boolean confirm, String comments) {
        Prescription prescription = prescriptionRepository.findById(prescriptionId)
                .orElseThrow(() -> new RuntimeException("处方不存在"));

        if (prescription.getStatus() != PrescriptionStatus.MODIFIED_BY_PHARMACIST) {
            throw new RuntimeException("只有药师修改过的处方需要医生确认");
        }

        if (confirm) {
            prescription.setStatus(PrescriptionStatus.APPROVED);
            prescription.setApprovedAt(LocalDateTime.now());
        } else {
            prescription.setStatus(PrescriptionStatus.CONSULTATION);
            prescription.setConsultationNotes(comments);
        }

        return prescriptionRepository.save(prescription);
    }

    @Transactional
    public Prescription dispensePrescription(Long prescriptionId) {
        Prescription prescription = prescriptionRepository.findById(prescriptionId)
                .orElseThrow(() -> new RuntimeException("处方不存在"));

        if (prescription.getStatus() != PrescriptionStatus.APPROVED) {
            throw new RuntimeException("只有已审核通过的处方可以发药");
        }

        prescription.setStatus(PrescriptionStatus.DISPENSED);

        for (PrescriptionItem item : prescription.getItems()) {
            Drug drug = item.getDrug();
            if (drug.getInventoryQuantity() != null) {
                drug.setInventoryQuantity(
                    drug.getInventoryQuantity().subtract(
                        item.getSingleDose().multiply(
                            java.math.BigDecimal.valueOf(item.getFrequencyPerDay() * item.getDurationDays())
                        )
                    )
                );
                drugRepository.save(drug);
            }
        }

        return prescriptionRepository.save(prescription);
    }

    @Transactional
    public Prescription completePrescription(Long prescriptionId) {
        Prescription prescription = prescriptionRepository.findById(prescriptionId)
                .orElseThrow(() -> new RuntimeException("处方不存在"));

        if (prescription.getStatus() != PrescriptionStatus.DISPENSED) {
            throw new RuntimeException("只有已发药的处方可以完成");
        }

        prescription.setStatus(PrescriptionStatus.COMPLETED);
        return prescriptionRepository.save(prescription);
    }

    public Prescription getPrescription(Long prescriptionId) {
        return prescriptionRepository.findById(prescriptionId)
                .orElseThrow(() -> new RuntimeException("处方不存在"));
    }

    public List<Prescription> getPatientPrescriptions(Long patientId) {
        return prescriptionRepository.findByPatientIdOrderByCreatedAtDesc(patientId);
    }

    public List<Prescription> getPrescriptionsByStatus(PrescriptionStatus status) {
        return prescriptionRepository.findByStatus(status);
    }

    private void saveValidations(Prescription prescription, List<ValidationResult> validations) {
        for (ValidationResult result : validations) {
            PrescriptionValidation pv = new PrescriptionValidation();
            pv.setValidationType(result.getValidationType());
            pv.setMessage(result.getMessage());
            pv.setSeverity(result.getSeverity());
            pv.setRelatedDrugNames(result.getRelatedDrugNames());
            pv.setSuggestion(result.getSuggestion());
            prescription.addValidation(pv);
        }
    }

    private void updatePrescriptionItems(Prescription prescription, List<PrescriptionItemRequest> modifiedItems) {
        prescription.getItems().clear();
        
        for (PrescriptionItemRequest itemReq : modifiedItems) {
            Drug drug = drugRepository.findById(itemReq.getDrugId())
                    .orElseThrow(() -> new RuntimeException("药品不存在"));
            
            PrescriptionItem item = new PrescriptionItem();
            item.setDrug(drug);
            item.setDrugName(drug.getName());
            item.setDosageForm(drug.getDosageForm());
            item.setSingleDose(itemReq.getSingleDose());
            item.setDoseUnit(drug.getDoseUnit());
            item.setFrequencyPerDay(itemReq.getFrequencyPerDay());
            item.setDurationDays(itemReq.getDurationDays());
            item.setSolventType(itemReq.getSolventType());
            prescription.addItem(item);
        }
    }
}
