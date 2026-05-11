package com.hospital.prescription.config;

import com.hospital.prescription.entity.*;
import com.hospital.prescription.enums.*;
import com.hospital.prescription.repository.*;
import lombok.RequiredArgsConstructor;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.LocalDate;

@Component
@RequiredArgsConstructor
public class DataInitializer implements CommandLineRunner {

    private final DrugRepository drugRepository;
    private final PatientRepository patientRepository;
    private final AllergyRecordRepository allergyRecordRepository;
    private final DrugInteractionRepository drugInteractionRepository;
    private final ChildDosageRuleRepository childDosageRuleRepository;

    @Override
    public void run(String... args) {
        if (drugRepository.count() > 0) {
            return;
        }

        initDrugs();
        initPatients();
        initAllergyRecords();
        initDrugInteractions();
        initChildDosageRules();
    }

    private void initDrugs() {
        Drug warfarin = createDrug("华法林钠片", DosageForm.TABLET, AdministrationRoute.ORAL,
                DrugCategory.NORMAL, new BigDecimal("5.0"), "mg", SolventType.NONE, new BigDecimal("100"));
        drugRepository.save(warfarin);

        Drug aspirin = createDrug("阿司匹林肠溶片", DosageForm.TABLET, AdministrationRoute.ORAL,
                DrugCategory.NORMAL, new BigDecimal("100.0"), "mg", SolventType.NONE, new BigDecimal("50"));
        drugRepository.save(aspirin);

        Drug cefazolin = createDrug("注射用头孢唑林钠", DosageForm.INJECTION, AdministrationRoute.INTRAVENOUS,
                DrugCategory.NORMAL, new BigDecimal("2.0"), "g", SolventType.NORMAL_SALINE, new BigDecimal("30"));
        drugRepository.save(cefazolin);

        Drug cefazolinOral = createDrug("头孢唑林胶囊", DosageForm.CAPSULE, AdministrationRoute.ORAL,
                DrugCategory.NORMAL, new BigDecimal("0.5"), "g", SolventType.NONE, new BigDecimal("80"));
        drugRepository.save(cefazolinOral);

        Drug amoxicillin = createDrug("阿莫西林胶囊", DosageForm.CAPSULE, AdministrationRoute.ORAL,
                DrugCategory.NORMAL, new BigDecimal("1.0"), "g", SolventType.NONE, new BigDecimal("0"));
        drugRepository.save(amoxicillin);

        Drug penicillin = createDrug("注射用青霉素钠", DosageForm.INJECTION, AdministrationRoute.INTRAVENOUS,
                DrugCategory.NORMAL, new BigDecimal("800"), "万单位", SolventType.NORMAL_SALINE, new BigDecimal("40"));
        drugRepository.save(penicillin);

        Drug morphine = createDrug("盐酸吗啡注射液", DosageForm.INJECTION, AdministrationRoute.INTRAMUSCULAR,
                DrugCategory.NARCOTIC, new BigDecimal("10"), "mg", SolventType.NORMAL_SALINE, new BigDecimal("20"));
        drugRepository.save(morphine);

        Drug diazepam = createDrug("地西泮注射液", DosageForm.INJECTION, AdministrationRoute.INTRAVENOUS,
                DrugCategory.PSYCHOTROPIC, new BigDecimal("10"), "mg", SolventType.GLUCOSE, new BigDecimal("25"));
        drugRepository.save(diazepam);

        Drug metformin = createDrug("盐酸二甲双胍片", DosageForm.TABLET, AdministrationRoute.ORAL,
                DrugCategory.NORMAL, new BigDecimal("1000"), "mg", SolventType.NONE, new BigDecimal("120"));
        drugRepository.save(metformin);

        Drug levofloxacin = createDrug("左氧氟沙星注射液", DosageForm.INJECTION, AdministrationRoute.INTRAVENOUS,
                DrugCategory.NORMAL, new BigDecimal("500"), "mg", SolventType.GLUCOSE, new BigDecimal("60"));
        drugRepository.save(levofloxacin);
    }

    private Drug createDrug(String name, DosageForm form, AdministrationRoute route,
                            DrugCategory category, BigDecimal maxDose, String unit,
                            SolventType solvent, BigDecimal inventory) {
        Drug drug = new Drug();
        drug.setName(name);
        drug.setDosageForm(form);
        drug.setAdministrationRoute(route);
        drug.setCategory(category);
        drug.setMaxSingleDose(maxDose);
        drug.setDoseUnit(unit);
        drug.setRequiredSolvent(solvent);
        drug.setInventoryQuantity(inventory);
        return drug;
    }

    private void initPatients() {
        Patient patient1 = new Patient();
        patient1.setName("张三");
        patient1.setIdCard("110101199001011234");
        patient1.setDateOfBirth(LocalDate.of(1990, 1, 1));
        patient1.setGender("男");
        patient1.setPhone("13800138001");
        patientRepository.save(patient1);

        Patient patient2 = new Patient();
        patient2.setName("李四");
        patient2.setIdCard("110101195501015678");
        patient2.setDateOfBirth(LocalDate.of(1955, 1, 1));
        patient2.setGender("女");
        patient2.setPhone("13800138002");
        patientRepository.save(patient2);

        Patient patient3 = new Patient();
        patient3.setName("王小宝");
        patient3.setIdCard("110101202306017890");
        patient3.setDateOfBirth(LocalDate.of(2023, 6, 1));
        patient3.setGender("男");
        patient3.setPhone("13800138003");
        patientRepository.save(patient3);

        Patient patient4 = new Patient();
        patient4.setName("王爷爷");
        patient4.setIdCard("110101195003151122");
        patient4.setDateOfBirth(LocalDate.of(1950, 3, 15));
        patient4.setGender("男");
        patient4.setPhone("13800138004");
        patientRepository.save(patient4);
    }

    private void initAllergyRecords() {
        Patient patient4 = patientRepository.findById(4L).orElse(null);
        if (patient4 != null) {
            AllergyRecord allergy1 = new AllergyRecord();
            allergy1.setPatient(patient4);
            allergy1.setAllergen("青霉素");
            allergy1.setAllergyType(AllergyType.CONFIRMED);
            allergy1.setReaction("过敏性休克");
            allergyRecordRepository.save(allergy1);

            AllergyRecord allergy2 = new AllergyRecord();
            allergy2.setPatient(patient4);
            allergy2.setAllergen("头孢类");
            allergy2.setAllergyType(AllergyType.SUSPECTED);
            allergy2.setReaction("皮疹");
            allergyRecordRepository.save(allergy2);
        }
    }

    private void initDrugInteractions() {
        Drug warfarin = drugRepository.findByName("华法林钠片").orElse(null);
        Drug aspirin = drugRepository.findByName("阿司匹林肠溶片").orElse(null);
        Drug cefazolin = drugRepository.findByName("注射用头孢唑林钠").orElse(null);
        Drug cefazolinOral = drugRepository.findByName("头孢唑林胶囊").orElse(null);
        Drug metformin = drugRepository.findByName("盐酸二甲双胍片").orElse(null);
        Drug levofloxacin = drugRepository.findByName("左氧氟沙星注射液").orElse(null);

        if (warfarin != null && aspirin != null) {
            DrugInteraction interaction1 = new DrugInteraction();
            interaction1.setDrugA(warfarin);
            interaction1.setDrugB(aspirin);
            interaction1.setSeverity(SeverityLevel.SEVERE);
            interaction1.setDescription("华法林与阿司匹林联用会显著增加出血风险，可能导致严重出血事件。");
            interaction1.setAlternative("建议单独使用其中一种药物，或在医生严密监测下使用，并密切监测凝血功能。");
            drugInteractionRepository.save(interaction1);
        }

        if (cefazolin != null) {
            DrugInteraction interaction2 = new DrugInteraction();
            interaction2.setDrugA(cefazolin);
            interaction2.setDrugB(cefazolin);
            interaction2.setSeverity(SeverityLevel.SEVERE);
            interaction2.setDescription("使用头孢类药物期间及停药后一周内禁止饮酒，否则可能产生双硫仑样反应，表现为面部潮红、头痛、恶心、呕吐、心悸等症状，严重时可危及生命。");
            interaction2.setAlternative("建议患者在用药期间及停药后一周内避免饮酒和含酒精的饮料。");
            drugInteractionRepository.save(interaction2);
        }

        if (metformin != null && levofloxacin != null) {
            DrugInteraction interaction3 = new DrugInteraction();
            interaction3.setDrugA(metformin);
            interaction3.setDrugB(levofloxacin);
            interaction3.setSeverity(SeverityLevel.MODERATE);
            interaction3.setDescription("左氧氟沙星可能增加二甲双胍的血药浓度，增加乳酸酸中毒风险。");
            interaction3.setAlternative("建议密切监测血糖水平，必要时调整二甲双胍剂量。");
            drugInteractionRepository.save(interaction3);
        }
    }

    private void initChildDosageRules() {
        Drug amoxicillin = drugRepository.findByName("阿莫西林胶囊").orElse(null);
        Drug cefazolinOral = drugRepository.findByName("头孢唑林胶囊").orElse(null);

        if (amoxicillin != null) {
            createChildDosageRule(amoxicillin, ChildAgeGroup.NEWBORN, new BigDecimal("0.3"));
            createChildDosageRule(amoxicillin, ChildAgeGroup.INFANT, new BigDecimal("0.4"));
            createChildDosageRule(amoxicillin, ChildAgeGroup.PRESCHOOL, new BigDecimal("0.5"));
            createChildDosageRule(amoxicillin, ChildAgeGroup.SCHOOL_AGE, new BigDecimal("0.7"));
            createChildDosageRule(amoxicillin, ChildAgeGroup.ADOLESCENT, new BigDecimal("0.85"));
        }

        if (cefazolinOral != null) {
            createChildDosageRule(cefazolinOral, ChildAgeGroup.NEWBORN, new BigDecimal("0.25"));
            createChildDosageRule(cefazolinOral, ChildAgeGroup.INFANT, new BigDecimal("0.35"));
            createChildDosageRule(cefazolinOral, ChildAgeGroup.PRESCHOOL, new BigDecimal("0.45"));
            createChildDosageRule(cefazolinOral, ChildAgeGroup.SCHOOL_AGE, new BigDecimal("0.6"));
            createChildDosageRule(cefazolinOral, ChildAgeGroup.ADOLESCENT, new BigDecimal("0.8"));
        }
    }

    private void createChildDosageRule(Drug drug, ChildAgeGroup ageGroup, BigDecimal ratio) {
        ChildDosageRule rule = new ChildDosageRule();
        rule.setDrug(drug);
        rule.setAgeGroup(ageGroup);
        rule.setDoseRatio(ratio);
        childDosageRuleRepository.save(rule);
    }
}
