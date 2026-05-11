package com.company.meetingroom.init;

import com.company.meetingroom.model.Room;
import com.company.meetingroom.repository.RoomRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

@Component
public class DataInitializer implements CommandLineRunner {
    
    private final RoomRepository roomRepository;
    
    @Autowired
    public DataInitializer(RoomRepository roomRepository) {
        this.roomRepository = roomRepository;
    }
    
    @Override
    public void run(String... args) {
        if (roomRepository.count() == 0) {
            Room r1 = new Room();
            r1.setCode("A101");
            r1.setName("小型会议室A101");
            r1.setLocation("1楼101号");
            r1.setCapacity(6);
            roomRepository.save(r1);
            
            Room r2 = new Room();
            r2.setCode("A201");
            r2.setName("中型会议室A201");
            r2.setLocation("2楼201号");
            r2.setCapacity(12);
            roomRepository.save(r2);
            
            Room r3 = new Room();
            r3.setCode("B301");
            r3.setName("大型会议室B301");
            r3.setLocation("3楼301号");
            r3.setCapacity(30);
            roomRepository.save(r3);
            
            Room r4 = new Room();
            r4.setCode("B302");
            r4.setName("培训室B302");
            r4.setLocation("3楼302号");
            r4.setCapacity(50);
            roomRepository.save(r4);
            
            System.out.println("已初始化4个会议室数据");
        }
    }
}
