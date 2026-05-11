package com.hotelbooking.repository;

import com.hotelbooking.model.entity.CleaningTask;
import com.hotelbooking.model.enums.CleaningPriority;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface CleaningTaskRepository extends JpaRepository<CleaningTask, Long> {
    List<CleaningTask> findByRoomIdOrderByCreatedAtDesc(Long roomId);
    List<CleaningTask> findByIsCompletedFalseOrderByPriorityAscCreatedAtAsc();
    List<CleaningTask> findByAssignedTo(Long assignedTo);
    List<CleaningTask> findByPriority(CleaningPriority priority);
}
