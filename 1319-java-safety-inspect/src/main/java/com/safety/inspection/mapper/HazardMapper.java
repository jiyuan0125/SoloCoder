package com.safety.inspection.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.safety.inspection.entity.Hazard;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.time.LocalDateTime;
import java.util.List;

@Mapper
public interface HazardMapper extends BaseMapper<Hazard> {

    @Select("SELECT * FROM hazard WHERE rectification_deadline < #{now} AND status IN ('PENDING_RECTIFICATION', 'RECTIFYING') AND deleted = 0")
    List<Hazard> selectOverdueHazards(@Param("now") LocalDateTime now);

    @Select("SELECT * FROM hazard WHERE area_id = #{areaId} AND hazard_type = #{hazardType} AND status != 'CLOSED' AND deleted = 0 ORDER BY created_at DESC")
    List<Hazard> selectSimilarHazards(@Param("areaId") Long areaId, @Param("hazardType") String hazardType);
}
