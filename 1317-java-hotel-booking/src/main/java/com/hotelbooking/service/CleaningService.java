package com.hotelbooking.service;

import com.hotelbooking.exception.ResourceNotFoundException;
import com.hotelbooking.model.entity.CleaningTask;
import com.hotelbooking.model.entity.Room;
import com.hotelbooking.model.enums.CleaningPriority;
import com.hotelbooking.model.enums.RoomStatus;
import com.hotelbooking.repository.CleaningTaskRepository;
import com.hotelbooking.repository.RoomRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Service
@RequiredArgsConstructor
public class CleaningService {

    private final CleaningTaskRepository cleaningTaskRepository;
    private final RoomRepository roomRepository;

    @Transactional
    public CleaningTask createCleaningTask(Long roomId, CleaningPriority priority, Long assignedTo) {
        Room room = roomRepository.findById(roomId)
                .orElseThrow(() -> new ResourceNotFoundException("房间不存在"));

        if (room.getStatus() != RoomStatus.CLEANING) {
            room.setStatus(RoomStatus.CLEANING);
        }

        LocalDateTime estimatedCompletion = calculateEstimatedCompletion(priority);
        room.setCleaningAvailableAt(estimatedCompletion);
        roomRepository.save(room);

        CleaningTask task = CleaningTask.builder()
                .room(room)
                .priority(priority)
                .assignedTo(assignedTo)
                .estimatedCompletionTime(estimatedCompletion)
                .build();

        return cleaningTaskRepository.save(task);
    }

    @Transactional
    public CleaningTask completeCleaningTask(Long taskId) {
        CleaningTask task = cleaningTaskRepository.findById(taskId)
                .orElseThrow(() -> new ResourceNotFoundException("清洁任务不存在"));

        task.setIsCompleted(true);
        task.setActualCompletionTime(LocalDateTime.now());
        cleaningTaskRepository.save(task);

        Room room = task.getRoom();
        room.setStatus(RoomStatus.AVAILABLE);
        room.setCleaningAvailableAt(null);
        roomRepository.save(room);

        return task;
    }

    @Transactional
    public CleaningTask updateCleaningEstimate(Long taskId, LocalDateTime newEstimatedCompletion) {
        CleaningTask task = cleaningTaskRepository.findById(taskId)
                .orElseThrow(() -> new ResourceNotFoundException("清洁任务不存在"));

        task.setEstimatedCompletionTime(newEstimatedCompletion);
        cleaningTaskRepository.save(task);

        Room room = task.getRoom();
        room.setCleaningAvailableAt(newEstimatedCompletion);
        roomRepository.save(room);

        return task;
    }

    @Transactional
    public CleaningTask rushCleaning(Long taskId) {
        CleaningTask task = cleaningTaskRepository.findById(taskId)
                .orElseThrow(() -> new ResourceNotFoundException("清洁任务不存在"));

        if (task.getPriority() != CleaningPriority.RUSH) {
            task.setPriority(CleaningPriority.RUSH);
            LocalDateTime newEstimate = calculateEstimatedCompletion(CleaningPriority.RUSH);
            task.setEstimatedCompletionTime(newEstimate);
            cleaningTaskRepository.save(task);

            Room room = task.getRoom();
            room.setCleaningAvailableAt(newEstimate);
            roomRepository.save(room);
        }

        return task;
    }

    public List<CleaningTask> getPendingTasks() {
        return cleaningTaskRepository.findByIsCompletedFalseOrderByPriorityAscCreatedAtAsc();
    }

    public List<CleaningTask> getTasksByAssignee(Long assigneeId) {
        return cleaningTaskRepository.findByAssignedTo(assigneeId);
    }

    public Optional<CleaningTask> findById(Long taskId) {
        return cleaningTaskRepository.findById(taskId);
    }

    @Scheduled(fixedRate = 60000)
    @Transactional
    public void checkAndCompleteExpiredCleaningTasks() {
        LocalDateTime now = LocalDateTime.now();
        List<CleaningTask> pendingTasks = cleaningTaskRepository.findByIsCompletedFalseOrderByPriorityAscCreatedAtAsc();

        for (CleaningTask task : pendingTasks) {
            if (task.getEstimatedCompletionTime().isBefore(now)) {
                completeCleaningTask(task.getId());
            }
        }
    }

    private LocalDateTime calculateEstimatedCompletion(CleaningPriority priority) {
        int minHours = priority.getMinHours();
        int maxHours = priority.getMaxHours();
        int hours = minHours + (int) (Math.random() * (maxHours - minHours + 1));
        return LocalDateTime.now().plusHours(hours);
    }
}
