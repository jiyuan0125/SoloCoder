package com.safety.inspection.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.safety.inspection.entity.InspectionTask;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.time.LocalDate;
import java.util.List;

@Mapper
public interface InspectionTaskMapper extends BaseMapper<InspectionTask> {

    @Select("SELECT * FROM inspection_task WHERE task_date = #{taskDate} AND deleted = 0")
    List<InspectionTask> selectTasksByDate(@Param("taskDate") LocalDate taskDate);

    @Select("SELECT * FROM inspection_task WHERE task_date < #{date} AND task_status = 'PENDING' AND deleted = 0")
    List<InspectionTask> selectOverduePendingTasks(@Param("date") LocalDate date);
}
