package com.hotelbooking.repository;

import com.hotelbooking.model.entity.BookingDailyDetail;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface BookingDailyDetailRepository extends JpaRepository<BookingDailyDetail, Long> {
    List<BookingDailyDetail> findByBookingId(Long bookingId);
    void deleteByBookingId(Long bookingId);
}
