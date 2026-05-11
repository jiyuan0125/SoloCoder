package com.hotelbooking.dto;

import com.hotelbooking.model.enums.RoomType;
import jakarta.validation.constraints.Future;
import jakarta.validation.constraints.FutureOrPresent;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.time.LocalDate;

@Data
public class BookingRequest {
    
    @NotNull(message = "房型不能为空")
    private RoomType roomType;
    
    @NotNull(message = "入住日期不能为空")
    @FutureOrPresent(message = "入住日期不能是过去的日期")
    private LocalDate checkInDate;
    
    @NotNull(message = "退房日期不能为空")
    @Future(message = "退房日期必须是未来的日期")
    private LocalDate checkOutDate;
    
    @Min(value = 1, message = "至少有1位客人")
    private Integer numberOfGuests = 1;
    
    private String customerName;
    
    @NotNull(message = "客户手机号不能为空")
    private String customerPhone;
    
    private String customerEmail;
    
    private String customerIdCard;
    
    private String specialRequests;
}
