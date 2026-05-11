package com.hospital.prescription.controller;

import com.hospital.prescription.dto.PrescriptionItemRequest;
import com.hospital.prescription.dto.PrescriptionRequest;
import com.hospital.prescription.dto.PrescriptionValidationResult;
import com.hospital.prescription.entity.Prescription;
import com.hospital.prescription.enums.PrescriptionStatus;
import com.hospital.prescription.service.PrescriptionService;
import com.hospital.prescription.service.PrescriptionValidationService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/prescriptions")
@RequiredArgsConstructor
public class PrescriptionController {

    private final PrescriptionService prescriptionService;
    private final PrescriptionValidationService validationService;

    @PostMapping("/validate")
    public ResponseEntity<PrescriptionValidationResult> validatePrescription(
            @RequestBody PrescriptionRequest request) {
        PrescriptionValidationResult result = validationService.validate(request);
        return ResponseEntity.ok(result);
    }

    @PostMapping
    public ResponseEntity<Prescription> createPrescription(@RequestBody PrescriptionRequest request) {
        Prescription prescription = prescriptionService.createPrescription(request);
        return ResponseEntity.ok(prescription);
    }

    @PostMapping("/{id}/submit")
    public ResponseEntity<Prescription> submitPrescription(
            @PathVariable Long id,
            @RequestBody(required = false) Map<String, String> body) {
        String overDoseReason = body != null ? body.get("overDoseReason") : null;
        Prescription prescription = prescriptionService.submitPrescription(id, overDoseReason);
        return ResponseEntity.ok(prescription);
    }

    @PostMapping("/{id}/review")
    public ResponseEntity<Prescription> pharmacistReview(
            @PathVariable Long id,
            @RequestBody Map<String, Object> body) {
        String pharmacistName = (String) body.get("pharmacistName");
        boolean approve = (Boolean) body.getOrDefault("approve", true);
        String comments = (String) body.get("comments");
        
        @SuppressWarnings("unchecked")
        List<PrescriptionItemRequest> modifiedItems = (List<PrescriptionItemRequest>) body.get("modifiedItems");
        
        Prescription prescription = prescriptionService.pharmacistReview(id, pharmacistName, approve, comments, modifiedItems);
        return ResponseEntity.ok(prescription);
    }

    @PostMapping("/{id}/confirm")
    public ResponseEntity<Prescription> doctorConfirm(
            @PathVariable Long id,
            @RequestBody Map<String, Object> body) {
        boolean confirm = (Boolean) body.getOrDefault("confirm", true);
        String comments = (String) body.get("comments");
        
        Prescription prescription = prescriptionService.doctorConfirmModified(id, confirm, comments);
        return ResponseEntity.ok(prescription);
    }

    @PostMapping("/{id}/dispense")
    public ResponseEntity<Prescription> dispensePrescription(@PathVariable Long id) {
        Prescription prescription = prescriptionService.dispensePrescription(id);
        return ResponseEntity.ok(prescription);
    }

    @PostMapping("/{id}/complete")
    public ResponseEntity<Prescription> completePrescription(@PathVariable Long id) {
        Prescription prescription = prescriptionService.completePrescription(id);
        return ResponseEntity.ok(prescription);
    }

    @GetMapping("/{id}")
    public ResponseEntity<Prescription> getPrescription(@PathVariable Long id) {
        Prescription prescription = prescriptionService.getPrescription(id);
        return ResponseEntity.ok(prescription);
    }

    @GetMapping("/patient/{patientId}")
    public ResponseEntity<List<Prescription>> getPatientPrescriptions(@PathVariable Long patientId) {
        List<Prescription> prescriptions = prescriptionService.getPatientPrescriptions(patientId);
        return ResponseEntity.ok(prescriptions);
    }

    @GetMapping("/status/{status}")
    public ResponseEntity<List<Prescription>> getPrescriptionsByStatus(@PathVariable PrescriptionStatus status) {
        List<Prescription> prescriptions = prescriptionService.getPrescriptionsByStatus(status);
        return ResponseEntity.ok(prescriptions);
    }
}
