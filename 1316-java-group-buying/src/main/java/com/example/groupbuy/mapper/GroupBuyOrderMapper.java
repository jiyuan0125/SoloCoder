package com.example.groupbuy.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.example.groupbuy.entity.GroupBuyOrder;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Update;

import java.time.LocalDateTime;
import java.util.List;

@Mapper
public interface GroupBuyOrderMapper extends BaseMapper<GroupBuyOrder> {
    
    @Update("UPDATE group_buy_order SET actual_people_count = actual_people_count + 1 WHERE id = #{orderId} AND actual_people_count < target_people_count")
    int incrementPeopleCount(@Param("orderId") Long orderId);
    
    @Update("UPDATE group_buy_order SET actual_people_count = actual_people_count - 1 WHERE id = #{orderId} AND actual_people_count > 0")
    int decrementPeopleCount(@Param("orderId") Long orderId);
    
    List<GroupBuyOrder> findExpiredOrders(@Param("now") LocalDateTime now);
}
