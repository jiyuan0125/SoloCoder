package com.company.meetingroom.dto;

import java.time.LocalDateTime;

public class ReservationRequest {
    private Long roomId;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private String booker;
    private String topic;

    public Long getRoomId() { return roomId; }
    public void setRoomId(Long roomId) { this.roomId = roomId; }
    public LocalDateTime getStartTime() { return startTime; }
    public void setStartTime(LocalDateTime startTime) { this.startTime = startTime; }
    public LocalDateTime getEndTime() { return endTime; }
    public void setEndTime(LocalDateTime endTime) { this.endTime = endTime; }
    public String getBooker() { return booker; }
    public void setBooker(String booker) { this.booker = booker; }
    public String getTopic() { return topic; }
    public void setTopic(String topic) { this.topic = topic; }
}
