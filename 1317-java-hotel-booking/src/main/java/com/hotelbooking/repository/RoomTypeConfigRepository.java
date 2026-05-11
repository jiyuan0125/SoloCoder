package com.hotelbooking.repository;

import com.hotelbooking.model.entity.RoomTypeConfig;
import com.hotelbooking.model.enums.RoomType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface RoomTypeConfigRepository extends JpaRepository<RoomTypeConfig, Long> {
    Optional<RoomTypeConfig> findByRoomType(RoomType roomType);
}
