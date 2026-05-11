package com.hospital.config;

import com.hospital.entity.Department;
import com.hospital.entity.Doctor;
import com.hospital.entity.DoctorSchedule;
import com.hospital.enums.DayOfWeek;
import com.hospital.enums.DoctorTitle;
import com.hospital.enums.TimeSlot;
import com.hospital.repository.DepartmentRepository;
import com.hospital.repository.DoctorRepository;
import com.hospital.repository.DoctorScheduleRepository;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.util.Arrays;
import java.util.List;

@Component
public class DataInitializer implements CommandLineRunner {

    private final DepartmentRepository departmentRepository;
    private final DoctorRepository doctorRepository;
    private final DoctorScheduleRepository scheduleRepository;

    public DataInitializer(DepartmentRepository departmentRepository,
                           DoctorRepository doctorRepository,
                           DoctorScheduleRepository scheduleRepository) {
        this.departmentRepository = departmentRepository;
        this.doctorRepository = doctorRepository;
        this.scheduleRepository = scheduleRepository;
    }

    @Override
    @Transactional
    public void run(String... args) {
        if (departmentRepository.count() > 0) {
            return;
        }

        Department internalMedicine = createDepartment("内科", "内科是一个集医疗、教学、科研为一体的综合性科室");
        Department surgery = createDepartment("外科", "外科以手术治疗为主，治疗各种外科疾病");
        Department dentistry = createDepartment("口腔科", "口腔科提供口腔疾病的预防、诊断和治疗服务");
        Department ophthalmology = createDepartment("眼科", "眼科专注于眼部疾病的诊断和治疗");

        List<DayOfWeek> workDays = Arrays.asList(
                DayOfWeek.MONDAY, DayOfWeek.TUESDAY, DayOfWeek.WEDNESDAY,
                DayOfWeek.THURSDAY, DayOfWeek.FRIDAY
        );

        Doctor doctor1 = createDoctor("张医生", DoctorTitle.EXPERT, internalMedicine);
        createSchedules(doctor1, workDays, Arrays.asList(TimeSlot.MORNING, TimeSlot.AFTERNOON));

        Doctor doctor2 = createDoctor("李医生", DoctorTitle.ORDINARY, internalMedicine);
        createSchedules(doctor2, workDays, Arrays.asList(TimeSlot.MORNING, TimeSlot.AFTERNOON));

        Doctor doctor3 = createDoctor("王医生", DoctorTitle.EXPERT, surgery);
        createSchedules(doctor3, workDays, Arrays.asList(TimeSlot.MORNING));

        Doctor doctor4 = createDoctor("赵医生", DoctorTitle.ORDINARY, surgery);
        createSchedules(doctor4, workDays, Arrays.asList(TimeSlot.AFTERNOON));

        Doctor doctor5 = createDoctor("刘医生", DoctorTitle.ORDINARY, dentistry);
        createSchedules(doctor5, workDays, Arrays.asList(TimeSlot.MORNING, TimeSlot.AFTERNOON));

        Doctor doctor6 = createDoctor("陈医生", DoctorTitle.EXPERT, ophthalmology);
        createSchedules(doctor6, workDays, Arrays.asList(TimeSlot.MORNING, TimeSlot.AFTERNOON));
    }

    private Department createDepartment(String name, String description) {
        Department department = new Department();
        department.setName(name);
        department.setDescription(description);
        return departmentRepository.save(department);
    }

    private Doctor createDoctor(String name, DoctorTitle title, Department department) {
        Doctor doctor = new Doctor();
        doctor.setName(name);
        doctor.setTitle(title);
        doctor.setDepartment(department);
        return doctorRepository.save(doctor);
    }

    private void createSchedules(Doctor doctor, List<DayOfWeek> days, List<TimeSlot> slots) {
        for (DayOfWeek day : days) {
            for (TimeSlot slot : slots) {
                DoctorSchedule schedule = new DoctorSchedule();
                schedule.setDoctor(doctor);
                schedule.setDayOfWeek(day);
                schedule.setTimeSlot(slot);
                schedule.setIsAvailable(true);
                scheduleRepository.save(schedule);
            }
        }
    }
}
