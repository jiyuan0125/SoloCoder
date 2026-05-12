package com.hospital.prescription.service;

import com.hospital.prescription.dto.PrescriptionItemRequest;
import com.hospital.prescription.dto.PrescriptionRequest;
import com.hospital.prescription.dto.PrescriptionValidationResult;
import com.hospital.prescription.dto.ValidationResult;
import com.hospital.prescription.entity.*;
import com.hospital.prescription.enums.*;
import com.hospital.prescription.repository.*;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class PrescriptionValidationService {

    private final DrugRepository drugRepository;
    private final AllergyRecordRepository allergyRecordRepository;
    private final DrugInteractionRepository drugInteractionRepository;
    private final ChildDosageRuleRepository childDosageRuleRepository;
    private final PatientRepository patientRepository;

    public PrescriptionValidationResult validate(PrescriptionRequest request) {
        List<ValidationResult> validations = new ArrayList<>();

        Patient patient = null;
        if (request.getPatientId() != null) {
            patient = patientRepository.findById(request.getPatientId()).orElse(null);
        }

        List<Drug> drugs = new ArrayList<>();
        Map<Long, PrescriptionItemRequest> itemMap = new HashMap<>();

        for (PrescriptionItemRequest itemReq : request.getItems()) {
            Drug drug = null;
            if (itemReq.getDrugId() != null) {
                drug = drugRepository.findById(itemReq.getDrugId()).orElse(null);
            }
            if (drug != null) {
                drugs.add(drug);
                itemMap.put(drug.getId(), itemReq);
            }
        }

        if (patient != null) {
            validateAllergyHistory(patient, drugs, validations);
            validateSpecialPopulation(patient, drugs, itemMap, validations);
        }

        validateDosage(drugs, itemMap, validations);
        validateDrugInteractions(drugs, validations);
        validateAdministrationRoute(drugs, itemMap, validations);
        validateInventory(drugs, validations);
        validatePsychotropicNarcoticRequirements(drugs, request, validations);

        boolean hasBlockingErrors = validations.stream().anyMatch(ValidationResult::isBlocking);
        boolean hasWarnings = validations.stream().anyMatch(v -> !v.isBlocking());
        
        boolean hasOverDose = validations.stream()
                .anyMatch(v -> "DOSAGE".equals(v.getValidationType()) && v.getSeverity() == SeverityLevel.MODERATE);
        
        boolean canSubmit = !hasBlockingErrors && (!hasOverDose || request.isOverDoseConfirmed());

        String confirmationMessage = null;
        if (hasOverDose && !request.isOverDoseConfirmed()) {
            confirmationMessage = "处方中存在超剂量用药，请确认后继续。确认后处方将标记为超剂量处方。";
        }

        return PrescriptionValidationResult.builder()
                .canSubmit(canSubmit)
                .hasBlockingErrors(hasBlockingErrors)
                .hasWarnings(hasWarnings)
                .validations(validations)
                .requiresOverDoseConfirmation(hasOverDose && !request.isOverDoseConfirmed())
                .confirmationMessage(confirmationMessage)
                .build();
    }

    private void validateDosage(List<Drug> drugs, Map<Long, PrescriptionItemRequest> itemMap, 
                                  List<ValidationResult> validations) {
        for (Drug drug : drugs) {
            PrescriptionItemRequest item = itemMap.get(drug.getId());
            if (item == null) continue;

            BigDecimal normalizedPrescriptionDose = convertToStandardUnit(item.getSingleDose(), item.getDoseUnit());
            BigDecimal normalizedDrugMaxDose = convertToStandardUnit(drug.getMaxSingleDose(), drug.getDoseUnit());

            if (normalizedPrescriptionDose.setScale(4, RoundingMode.HALF_UP).compareTo(normalizedDrugMaxDose) > 0) {
                validations.add(ValidationResult.builder()
                        .validationType("DOSAGE")
                        .message(String.format("药品[%s]单次剂量超过最大限制。处方剂量: %s%s，最大限制: %s%s",
                                drug.getName(),
                                item.getSingleDose(), item.getDoseUnit(),
                                drug.getMaxSingleDose(), drug.getDoseUnit()))
                        .severity(SeverityLevel.MODERATE)
                        .relatedDrugNames(drug.getName())
                        .suggestion(String.format("建议将单次剂量调整至 %s%s 以下，或确认该超剂量用药的必要性",
                                drug.getMaxSingleDose(), drug.getDoseUnit()))
                        .build());
            }
        }
    }
    
    private BigDecimal convertToStandardUnit(BigDecimal dose, String unit) {
        if (dose == null || unit == null) {
            return dose;
        }
        
        String lowerUnit = unit.toLowerCase().trim();
        
        if (lowerUnit.equals("mg") || lowerUnit.equals("毫克")) {
            return dose.divide(new BigDecimal("1000"), 6, RoundingMode.HALF_UP);
        } else if (lowerUnit.equals("g") || lowerUnit.equals("克")) {
            return dose;
        } else if (lowerUnit.equals("μg") || lowerUnit.equals("mcg") || lowerUnit.equals("微克")) {
            return dose.divide(new BigDecimal("1000000"), 9, RoundingMode.HALF_UP);
        } else if (lowerUnit.equals("kg") || lowerUnit.equals("千克")) {
            return dose.multiply(new BigDecimal("1000"));
        }
        
        return dose;
    }

    private void validateDrugInteractions(List<Drug> drugs, List<ValidationResult> validations) {
        if (drugs.size() < 2) return;

        Set<String> checkedPairs = new HashSet<>();

        for (int i = 0; i < drugs.size(); i++) {
            for (int j = i + 1; j < drugs.size(); j++) {
                Drug drugA = drugs.get(i);
                Drug drugB = drugs.get(j);
                
                String pairKey = Math.min(drugA.getId(), drugB.getId()) + "-" + Math.max(drugA.getId(), drugB.getId());
                if (checkedPairs.contains(pairKey)) continue;
                checkedPairs.add(pairKey);

                List<DrugInteraction> interactions = drugInteractionRepository.findInteractionBetween(drugA, drugB);
                
                for (DrugInteraction interaction : interactions) {
                    validations.add(ValidationResult.builder()
                            .validationType("DRUG_INTERACTION")
                            .message(String.format("药品配伍禁忌：[%s] 与 [%s] 联用存在风险。%s",
                                    drugA.getName(), drugB.getName(), interaction.getDescription()))
                            .severity(interaction.getSeverity())
                            .relatedDrugNames(drugA.getName() + ", " + drugB.getName())
                            .suggestion(interaction.getAlternative())
                            .build());
                }
            }
        }
    }

    private void validateAdministrationRoute(List<Drug> drugs, Map<Long, PrescriptionItemRequest> itemMap,
                                                List<ValidationResult> validations) {
        Map<String, List<Drug>> drugByGenericName = drugs.stream()
                .collect(Collectors.groupingBy(drug -> 
                    (drug.getGenericName() != null && !drug.getGenericName().isEmpty()) 
                        ? drug.getGenericName() 
                        : extractGenericName(drug.getName())));

        for (Map.Entry<String, List<Drug>> entry : drugByGenericName.entrySet()) {
            if (entry.getValue().size() > 1) {
                List<DosageForm> forms = entry.getValue().stream()
                        .map(Drug::getDosageForm)
                        .distinct()
                        .toList();
                
                if (forms.size() > 1) {
                    String drugNames = entry.getValue().stream()
                            .map(Drug::getName)
                            .collect(Collectors.joining("、"));
                    
                    String formNames = forms.stream()
                            .map(DosageForm::getDescription)
                            .collect(Collectors.joining("、"));
                    
                    validations.add(ValidationResult.builder()
                            .validationType("ADMINISTRATION_ROUTE")
                            .message(String.format("同一药品[%s]在处方中同时出现多种剂型：%s",
                                    drugNames, formNames))
                            .severity(SeverityLevel.SEVERE)
                            .relatedDrugNames(drugNames)
                            .suggestion("请选择单一剂型使用")
                            .build());
                }
            }
        }

        for (Drug drug : drugs) {
            if (drug.getDosageForm() == DosageForm.INJECTION && 
                drug.getRequiredSolvent() != null && 
                drug.getRequiredSolvent() != SolventType.NONE) {
                
                PrescriptionItemRequest item = itemMap.get(drug.getId());
                if (item == null) continue;

                if (item.getSolventType() == null) {
                    validations.add(ValidationResult.builder()
                            .validationType("SOLVENT")
                            .message(String.format("注射液[%s]需要指定溶媒，推荐溶媒：%s",
                                    drug.getName(), drug.getRequiredSolvent().getDescription()))
                            .severity(SeverityLevel.SEVERE)
                            .relatedDrugNames(drug.getName())
                            .suggestion(String.format("请选择 %s 作为溶媒", drug.getRequiredSolvent().getDescription()))
                            .build());
                } else if (item.getSolventType() != drug.getRequiredSolvent()) {
                    validations.add(ValidationResult.builder()
                            .validationType("SOLVENT")
                            .message(String.format("注射液[%s]溶媒选择错误。处方选择: %s，推荐溶媒: %s",
                                    drug.getName(),
                                    item.getSolventType().getDescription(),
                                    drug.getRequiredSolvent().getDescription()))
                            .severity(SeverityLevel.SEVERE)
                            .relatedDrugNames(drug.getName())
                            .suggestion(String.format("请将溶媒更换为 %s", drug.getRequiredSolvent().getDescription()))
                            .build());
                }
            }
        }
    }

    private void validateAllergyHistory(Patient patient, List<Drug> drugs, List<ValidationResult> validations) {
        List<AllergyRecord> allergies = allergyRecordRepository.findByPatientId(patient.getId());
        
        for (Drug drug : drugs) {
            String drugGenericName = extractGenericName(drug.getName());
            
            for (AllergyRecord allergy : allergies) {
                String allergen = allergy.getAllergen();
                
                if (matchesAllergen(drug.getName(), drugGenericName, allergen)) {
                    SeverityLevel severity = allergy.getAllergyType() == AllergyType.CONFIRMED 
                            ? SeverityLevel.SEVERE 
                            : SeverityLevel.MODERATE;
                    
                    validations.add(ValidationResult.builder()
                            .validationType("ALLERGY")
                            .message(String.format("患者[%s]存在%s：%s。药品[%s]可能引发过敏反应。",
                                    patient.getName(),
                                    allergy.getAllergyType().getDescription(),
                                    allergy.getAllergen(),
                                    drug.getName()))
                            .severity(severity)
                            .relatedDrugNames(drug.getName())
                            .suggestion("建议更换其他不含该过敏原的药品")
                            .build());
                }
            }
        }
    }
    
    private boolean matchesAllergen(String drugName, String drugGenericName, String allergen) {
        String lowerDrugName = drugName.toLowerCase();
        String lowerGenericName = drugGenericName.toLowerCase();
        String lowerAllergen = allergen.toLowerCase();
        
        if (lowerDrugName.contains(lowerAllergen)) {
            return true;
        }
        
        if (lowerGenericName.contains(lowerAllergen)) {
            return true;
        }
        
        String allergenBase = lowerAllergen.replace("类", "").replace("素", "");
        if (lowerGenericName.startsWith(allergenBase) || lowerDrugName.startsWith(allergenBase)) {
            return true;
        }
        
        return false;
    }

    private void validateSpecialPopulation(Patient patient, List<Drug> drugs, 
                                            Map<Long, PrescriptionItemRequest> itemMap,
                                            List<ValidationResult> validations) {
        if (patient.isChild()) {
            validateChildDosage(patient, drugs, itemMap, validations);
        }

        if (patient.isElderly()) {
            validateElderlyDosage(patient, drugs, itemMap, validations);
        }
    }

    private void validateChildDosage(Patient patient, List<Drug> drugs,
                                       Map<Long, PrescriptionItemRequest> itemMap,
                                       List<ValidationResult> validations) {
        ChildAgeGroup ageGroup = ChildAgeGroup.fromAgeInDays(patient.getAgeInDays());
        if (ageGroup == null) return;

        for (Drug drug : drugs) {
            PrescriptionItemRequest item = itemMap.get(drug.getId());
            if (item == null) continue;

            Optional<ChildDosageRule> ruleOpt = childDosageRuleRepository.findByDrugIdAndAgeGroup(
                    drug.getId(), ageGroup);

            if (ruleOpt.isPresent()) {
                ChildDosageRule rule = ruleOpt.get();
                BigDecimal maxAllowedDose;
                
                if (rule.getMaxSingleDose() != null) {
                    maxAllowedDose = rule.getMaxSingleDose();
                } else {
                    maxAllowedDose = drug.getMaxSingleDose().multiply(rule.getDoseRatio());
                }

                BigDecimal normalizedPrescriptionDose = convertToStandardUnit(item.getSingleDose(), item.getDoseUnit());
                BigDecimal normalizedMaxDose = convertToStandardUnit(maxAllowedDose, drug.getDoseUnit());

                if (normalizedPrescriptionDose.setScale(4, RoundingMode.HALF_UP).compareTo(normalizedMaxDose) > 0) {
                    validations.add(ValidationResult.builder()
                            .validationType("CHILD_DOSAGE")
                            .message(String.format("儿童患者[%s]（%s）超剂量用药。药品[%s]处方剂量: %s%s，年龄段建议最大剂量: %s%s",
                                    patient.getName(), ageGroup.getDescription(),
                                    drug.getName(),
                                    item.getSingleDose(), item.getDoseUnit(),
                                    maxAllowedDose, drug.getDoseUnit()))
                            .severity(SeverityLevel.SEVERE)
                            .relatedDrugNames(drug.getName())
                            .suggestion(String.format("建议将剂量调整至 %s%s 以下", maxAllowedDose, drug.getDoseUnit()))
                            .build());
                }
            }
        }
    }

    private void validateElderlyDosage(Patient patient, List<Drug> drugs,
                                         Map<Long, PrescriptionItemRequest> itemMap,
                                         List<ValidationResult> validations) {
        for (Drug drug : drugs) {
            PrescriptionItemRequest item = itemMap.get(drug.getId());
            if (item == null) continue;

            BigDecimal elderlyMaxDose = drug.getMaxSingleDose().multiply(new BigDecimal("0.75"));
            
            BigDecimal normalizedPrescriptionDose = convertToStandardUnit(item.getSingleDose(), item.getDoseUnit());
            BigDecimal normalizedElderlyMaxDose = convertToStandardUnit(elderlyMaxDose, drug.getDoseUnit());
            
            if (normalizedPrescriptionDose.setScale(4, RoundingMode.HALF_UP).compareTo(normalizedElderlyMaxDose) > 0) {
                validations.add(ValidationResult.builder()
                        .validationType("ELDERLY_DOSAGE")
                        .message(String.format("老年患者[%s]（%d岁）用药需注意肝肾功能减退。药品[%s]处方剂量: %s%s，建议老年人剂量: %s%s以下",
                                patient.getName(), patient.getAgeInYears(),
                                drug.getName(),
                                item.getSingleDose(), item.getDoseUnit(),
                                elderlyMaxDose, drug.getDoseUnit()))
                        .severity(SeverityLevel.MODERATE)
                        .relatedDrugNames(drug.getName())
                        .suggestion(String.format("建议将剂量调整至 %s%s 以下，或密切监测肝肾功能",
                                elderlyMaxDose, drug.getDoseUnit()))
                        .build());
            }
        }
    }

    private void validateInventory(List<Drug> drugs, List<ValidationResult> validations) {
        for (Drug drug : drugs) {
            if (drug.isLowInventory()) {
                validations.add(ValidationResult.builder()
                        .validationType("INVENTORY")
                        .message(String.format("药品[%s]库存不足", drug.getName()))
                        .severity(SeverityLevel.MILD)
                        .relatedDrugNames(drug.getName())
                        .suggestion("药房将安排调货，请确认是否继续开具处方")
                        .build());
            }
        }
    }

    private void validatePsychotropicNarcoticRequirements(List<Drug> drugs, PrescriptionRequest request,
                                                            List<ValidationResult> validations) {
        boolean hasSpecialDrugs = drugs.stream()
                .anyMatch(d -> d.getCategory() == DrugCategory.PSYCHOTROPIC || d.getCategory() == DrugCategory.NARCOTIC);
        
        if (hasSpecialDrugs) {
            List<String> specialDrugNames = drugs.stream()
                    .filter(d -> d.getCategory() == DrugCategory.PSYCHOTROPIC || d.getCategory() == DrugCategory.NARCOTIC)
                    .map(Drug::getName)
                    .toList();

            validations.add(ValidationResult.builder()
                    .validationType("SPECIAL_DRUG")
                    .message(String.format("处方包含精麻类药品：%s。需要双人签名确认，处方保存期限延长。",
                            String.join("、", specialDrugNames)))
                    .severity(SeverityLevel.MILD)
                    .relatedDrugNames(String.join(", ", specialDrugNames))
                    .suggestion("请确保有第二位医师/药师签字确认")
                    .build());
        }
    }

    private String extractGenericName(String drugName) {
        String[] dosageForms = {
            "肠溶片", "缓释片", "控释片", "分散片", "咀嚼片", "泡腾片", "舌下片",
            "缓释胶囊", "控释胶囊", "肠溶胶囊", "注射液", "注射用",
            "片", "胶囊", "糖浆", "栓剂", "软膏", "粉剂", "滴剂", "颗粒剂", "口服溶液"
        };
        
        String[] saltsAndAcids = {
            "盐酸", "硫酸", "磷酸", "醋酸", "马来酸", "富马酸", "苯磺酸", "甲磺酸",
            "钠", "钾", "钙", "镁", "氯", "枸橼酸", "酒石酸"
        };
        
        String genericName = drugName;
        
        for (String form : dosageForms) {
            genericName = genericName.replace(form, "");
        }
        
        for (String salt : saltsAndAcids) {
            genericName = genericName.replace(salt, "");
        }
        
        return genericName.trim();
    }
}
