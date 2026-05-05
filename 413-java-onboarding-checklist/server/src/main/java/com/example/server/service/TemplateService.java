package com.example.server.service;

import com.example.common.dto.ChecklistItemDTO;
import com.example.common.dto.ChecklistTemplateDTO;
import com.example.common.dto.TemplateItemDTO;
import com.example.common.enums.PositionType;
import com.example.common.request.CreateTemplateRequest;
import com.example.server.repository.TemplateRepository;
import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class TemplateService {

    @Autowired
    private TemplateRepository templateRepository;

    public ChecklistTemplateDTO createTemplate(CreateTemplateRequest request) {
        ChecklistTemplateDTO template = new ChecklistTemplateDTO();
        template.setName(request.getName());
        template.setPositionType(request.getPositionType());
        template.setDescription(request.getDescription());
        template.setItems(request.getItems() != null ? request.getItems() : new ArrayList<>());
        template.setDefault(request.isDefault());
        return templateRepository.save(template);
    }

    public Optional<ChecklistTemplateDTO> getTemplateById(String id) {
        return templateRepository.findById(id);
    }

    public List<ChecklistTemplateDTO> getAllTemplates() {
        return templateRepository.findAll();
    }

    public boolean deleteTemplate(String id) {
        if (templateRepository.existsById(id)) {
            templateRepository.deleteById(id);
            return true;
        }
        return false;
    }

    public Optional<ChecklistTemplateDTO> copyTemplate(String sourceId, String newName) {
        Optional<ChecklistTemplateDTO> sourceOpt = templateRepository.findById(sourceId);
        if (sourceOpt.isPresent() && !templateRepository.existsByName(newName)) {
            ChecklistTemplateDTO source = sourceOpt.get();
            ChecklistTemplateDTO newTemplate = new ChecklistTemplateDTO();
            newTemplate.setName(newName);
            newTemplate.setPositionType(source.getPositionType());
            newTemplate.setDescription(source.getDescription() + " (副本)");
            newTemplate.setDefault(false);
            
            List<TemplateItemDTO> newItems = new ArrayList<>();
            for (TemplateItemDTO sourceItem : source.getItems()) {
                TemplateItemDTO newItem = new TemplateItemDTO();
                newItem.setId(UUID.randomUUID().toString());
                newItem.setName(sourceItem.getName());
                newItem.setDescription(sourceItem.getDescription());
                newItem.setResponsiblePerson(sourceItem.getResponsiblePerson());
                newItem.setDepartment(sourceItem.getDepartment());
                newItem.setDaysRelativeToOnboarding(sourceItem.getDaysRelativeToOnboarding());
                newItem.setRequired(sourceItem.isRequired());
                newItem.setOnboardingDayRequired(sourceItem.isOnboardingDayRequired());
                newItem.setPreOnboarding(sourceItem.isPreOnboarding());
                newItems.add(newItem);
            }
            newTemplate.setItems(newItems);
            
            return Optional.of(templateRepository.save(newTemplate));
        }
        return Optional.empty();
    }

    public List<ChecklistItemDTO> generateChecklistFromTemplate(PositionType positionType, LocalDate onboardingDate) {
        List<ChecklistItemDTO> checklistItems = new ArrayList<>();
        
        Optional<ChecklistTemplateDTO> templateOpt = templateRepository.findDefaultByPositionType(positionType);
        if (templateOpt.isEmpty()) {
            templateOpt = templateRepository.findDefaultByPositionType(PositionType.ADMIN);
        }
        
        if (templateOpt.isPresent()) {
            ChecklistTemplateDTO template = templateOpt.get();
            for (TemplateItemDTO templateItem : template.getItems()) {
                ChecklistItemDTO item = convertToChecklistItem(templateItem, onboardingDate);
                checklistItems.add(item);
            }
        }
        
        return checklistItems;
    }

    private ChecklistItemDTO convertToChecklistItem(TemplateItemDTO templateItem, LocalDate onboardingDate) {
        ChecklistItemDTO item = new ChecklistItemDTO();
        item.setName(templateItem.getName());
        item.setDescription(templateItem.getDescription());
        item.setResponsiblePerson(templateItem.getResponsiblePerson());
        item.setDepartment(templateItem.getDepartment());
        item.setDueDate(onboardingDate.plusDays(templateItem.getDaysRelativeToOnboarding()));
        item.setRequired(templateItem.isRequired());
        item.setOnboardingDayRequired(templateItem.isOnboardingDayRequired());
        item.setPreOnboarding(templateItem.isPreOnboarding());
        return item;
    }

    @PostConstruct
    public void initDefaultTemplates() {
        if (templateRepository.findAll().isEmpty()) {
            createDefaultDeveloperTemplate();
            createDefaultSalesTemplate();
            createDefaultAdminTemplate();
        }
    }

    private void createDefaultDeveloperTemplate() {
        ChecklistTemplateDTO template = new ChecklistTemplateDTO();
        template.setName("开发岗入职清单");
        template.setPositionType(PositionType.DEVELOPER);
        template.setDescription("开发岗位默认入职清单模板");
        template.setDefault(true);
        
        List<TemplateItemDTO> items = new ArrayList<>();
        
        TemplateItemDTO item1 = new TemplateItemDTO();
        item1.setName("提前准备设备和工位");
        item1.setDescription("为新员工准备电脑、显示器、键盘等设备，并安排好工位");
        item1.setResponsiblePerson("IT管理员");
        item1.setDepartment("IT部门");
        item1.setDaysRelativeToOnboarding(-3);
        item1.setRequired(true);
        item1.setPreOnboarding(true);
        items.add(item1);
        
        TemplateItemDTO item2 = new TemplateItemDTO();
        item2.setName("开邮箱");
        item2.setDescription("为新员工开通企业邮箱");
        item2.setResponsiblePerson("IT管理员");
        item2.setDepartment("IT部门");
        item2.setDaysRelativeToOnboarding(0);
        item2.setRequired(true);
        item2.setOnboardingDayRequired(true);
        items.add(item2);
        
        TemplateItemDTO item3 = new TemplateItemDTO();
        item3.setName("签合同");
        item3.setDescription("与新员工签订劳动合同");
        item3.setResponsiblePerson("HR专员");
        item3.setDepartment("人力资源部");
        item3.setDaysRelativeToOnboarding(0);
        item3.setRequired(true);
        item3.setOnboardingDayRequired(true);
        items.add(item3);
        
        TemplateItemDTO item4 = new TemplateItemDTO();
        item4.setName("安全培训");
        item4.setDescription("为新员工进行信息安全和公司规章制度培训");
        item4.setResponsiblePerson("安全专员");
        item4.setDepartment("安全部");
        item4.setDaysRelativeToOnboarding(0);
        item4.setRequired(true);
        item4.setOnboardingDayRequired(true);
        items.add(item4);
        
        TemplateItemDTO item5 = new TemplateItemDTO();
        item5.setName("领门禁卡");
        item5.setDescription("为新员工办理并发放门禁卡");
        item5.setResponsiblePerson("行政专员");
        item5.setDepartment("行政部");
        item5.setDaysRelativeToOnboarding(0);
        item5.setRequired(true);
        items.add(item5);
        
        TemplateItemDTO item6 = new TemplateItemDTO();
        item6.setName("配置开发环境");
        item6.setDescription("为新开发人员配置开发环境，包括IDE、开发工具等");
        item6.setResponsiblePerson("技术主管");
        item6.setDepartment("技术部");
        item6.setDaysRelativeToOnboarding(1);
        item6.setRequired(true);
        items.add(item6);
        
        TemplateItemDTO item7 = new TemplateItemDTO();
        item7.setName("配置VPN");
        item7.setDescription("为新开发人员配置VPN访问权限");
        item7.setResponsiblePerson("IT管理员");
        item7.setDepartment("IT部门");
        item7.setDaysRelativeToOnboarding(1);
        item7.setRequired(true);
        items.add(item7);
        
        template.setItems(items);
        templateRepository.save(template);
    }

    private void createDefaultSalesTemplate() {
        ChecklistTemplateDTO template = new ChecklistTemplateDTO();
        template.setName("销售岗入职清单");
        template.setPositionType(PositionType.SALES);
        template.setDescription("销售岗位默认入职清单模板");
        template.setDefault(true);
        
        List<TemplateItemDTO> items = new ArrayList<>();
        
        TemplateItemDTO item1 = new TemplateItemDTO();
        item1.setName("提前准备设备和工位");
        item1.setDescription("为新员工准备电脑、显示器等设备，并安排好工位");
        item1.setResponsiblePerson("IT管理员");
        item1.setDepartment("IT部门");
        item1.setDaysRelativeToOnboarding(-3);
        item1.setRequired(true);
        item1.setPreOnboarding(true);
        items.add(item1);
        
        TemplateItemDTO item2 = new TemplateItemDTO();
        item2.setName("开邮箱");
        item2.setDescription("为新员工开通企业邮箱");
        item2.setResponsiblePerson("IT管理员");
        item2.setDepartment("IT部门");
        item2.setDaysRelativeToOnboarding(0);
        item2.setRequired(true);
        item2.setOnboardingDayRequired(true);
        items.add(item2);
        
        TemplateItemDTO item3 = new TemplateItemDTO();
        item3.setName("签合同");
        item3.setDescription("与新员工签订劳动合同");
        item3.setResponsiblePerson("HR专员");
        item3.setDepartment("人力资源部");
        item3.setDaysRelativeToOnboarding(0);
        item3.setRequired(true);
        item3.setOnboardingDayRequired(true);
        items.add(item3);
        
        TemplateItemDTO item4 = new TemplateItemDTO();
        item4.setName("安全培训");
        item4.setDescription("为新员工进行信息安全和公司规章制度培训");
        item4.setResponsiblePerson("安全专员");
        item4.setDepartment("安全部");
        item4.setDaysRelativeToOnboarding(0);
        item4.setRequired(true);
        item4.setOnboardingDayRequired(true);
        items.add(item4);
        
        TemplateItemDTO item5 = new TemplateItemDTO();
        item5.setName("领门禁卡");
        item5.setDescription("为新员工办理并发放门禁卡");
        item5.setResponsiblePerson("行政专员");
        item5.setDepartment("行政部");
        item5.setDaysRelativeToOnboarding(0);
        item5.setRequired(true);
        items.add(item5);
        
        TemplateItemDTO item6 = new TemplateItemDTO();
        item6.setName("配置手机");
        item6.setDescription("为新销售人员配置工作手机");
        item6.setResponsiblePerson("IT管理员");
        item6.setDepartment("IT部门");
        item6.setDaysRelativeToOnboarding(1);
        item6.setRequired(true);
        items.add(item6);
        
        TemplateItemDTO item7 = new TemplateItemDTO();
        item7.setName("配置CRM账号");
        item7.setDescription("为新销售人员配置CRM系统账号和权限");
        item7.setResponsiblePerson("销售主管");
        item7.setDepartment("销售部");
        item7.setDaysRelativeToOnboarding(1);
        item7.setRequired(true);
        items.add(item7);
        
        template.setItems(items);
        templateRepository.save(template);
    }

    private void createDefaultAdminTemplate() {
        ChecklistTemplateDTO template = new ChecklistTemplateDTO();
        template.setName("通用入职清单");
        template.setPositionType(PositionType.ADMIN);
        template.setDescription("通用入职清单模板（适用于所有岗位）");
        template.setDefault(true);
        
        List<TemplateItemDTO> items = new ArrayList<>();
        
        TemplateItemDTO item1 = new TemplateItemDTO();
        item1.setName("提前准备设备和工位");
        item1.setDescription("为新员工准备电脑、显示器等设备，并安排好工位");
        item1.setResponsiblePerson("IT管理员");
        item1.setDepartment("IT部门");
        item1.setDaysRelativeToOnboarding(-3);
        item1.setRequired(true);
        item1.setPreOnboarding(true);
        items.add(item1);
        
        TemplateItemDTO item2 = new TemplateItemDTO();
        item2.setName("开邮箱");
        item2.setDescription("为新员工开通企业邮箱");
        item2.setResponsiblePerson("IT管理员");
        item2.setDepartment("IT部门");
        item2.setDaysRelativeToOnboarding(0);
        item2.setRequired(true);
        item2.setOnboardingDayRequired(true);
        items.add(item2);
        
        TemplateItemDTO item3 = new TemplateItemDTO();
        item3.setName("签合同");
        item3.setDescription("与新员工签订劳动合同");
        item3.setResponsiblePerson("HR专员");
        item3.setDepartment("人力资源部");
        item3.setDaysRelativeToOnboarding(0);
        item3.setRequired(true);
        item3.setOnboardingDayRequired(true);
        items.add(item3);
        
        TemplateItemDTO item4 = new TemplateItemDTO();
        item4.setName("安全培训");
        item4.setDescription("为新员工进行信息安全和公司规章制度培训");
        item4.setResponsiblePerson("安全专员");
        item4.setDepartment("安全部");
        item4.setDaysRelativeToOnboarding(0);
        item4.setRequired(true);
        item4.setOnboardingDayRequired(true);
        items.add(item4);
        
        TemplateItemDTO item5 = new TemplateItemDTO();
        item5.setName("领门禁卡");
        item5.setDescription("为新员工办理并发放门禁卡");
        item5.setResponsiblePerson("行政专员");
        item5.setDepartment("行政部");
        item5.setDaysRelativeToOnboarding(0);
        item5.setRequired(true);
        items.add(item5);
        
        template.setItems(items);
        templateRepository.save(template);
    }
}
