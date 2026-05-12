from datetime import date, time, datetime
from decimal import Decimal
from sqlalchemy.orm import Session
from app.database import SessionLocal, engine, Base
from app.models import (
    Book,
    BookCopy,
    Reader,
    Seat,
    Event,
    ReaderType,
    CopyStatus,
)


def init_sample_data():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    
    try:
        print("初始化示例数据...")
        
        if db.query(Reader).count() == 0:
            readers = [
                Reader(card_number="R001", name="张三", reader_type=ReaderType.NORMAL.value, phone="13800000001"),
                Reader(card_number="R002", name="李四", reader_type=ReaderType.STUDENT.value, phone="13800000002"),
                Reader(card_number="R003", name="王五", reader_type=ReaderType.TEACHER.value, phone="13800000003"),
            ]
            db.add_all(readers)
            db.commit()
            print(f"  已创建 {len(readers)} 个读者")
        
        if db.query(Book).count() == 0:
            books = [
                Book(isbn="9787111123451", title="Python编程从入门到实践", author="Eric Matthes",
                     publisher="机械工业出版社", price=Decimal("89.00"), category="编程"),
                Book(isbn="9787111123452", title="深入理解计算机系统", author="Randal E. Bryant",
                     publisher="机械工业出版社", price=Decimal("139.00"), category="计算机"),
                Book(isbn="9787111123453", title="算法导论", author="Thomas H. Cormen",
                     publisher="机械工业出版社", price=Decimal("128.00"), category="算法"),
            ]
            db.add_all(books)
            db.commit()
            
            copies = []
            for i, book in enumerate(books):
                for j in range(3):
                    copies.append(BookCopy(
                        book_id=book.id,
                        copy_number=f"C{book.id:03d}-{j+1:02d}",
                        location=f"A{1+i:02d}",
                        condition="良好",
                    ))
            db.add_all(copies)
            db.commit()
            print(f"  已创建 {len(books)} 本图书，{len(copies)} 个副本")
        
        if db.query(Seat).count() == 0:
            seats = []
            floors = ["1F", "2F", "3F"]
            for floor in floors:
                for row in "ABC":
                    for num in range(1, 11):
                        seats.append(Seat(
                            seat_number=f"{floor}-{row}{num:02d}",
                            floor=floor,
                            area="阅览区",
                            has_power=num % 2 == 0,
                        ))
            db.add_all(seats)
            db.commit()
            print(f"  已创建 {len(seats)} 个座位")
        
        if db.query(Event).count() == 0:
            events = [
                Event(
                    title="Python编程工作坊",
                    description="学习Python基础编程",
                    event_date=date.today(),
                    start_time=time(14, 0),
                    end_time=time(17, 0),
                    location="会议室A",
                    max_participants=20,
                ),
                Event(
                    title="读书会：经典文学赏析",
                    description="一起讨论经典文学作品",
                    event_date=date.today(),
                    start_time=time(19, 0),
                    end_time=time(21, 0),
                    location="阅读区",
                    max_participants=15,
                ),
            ]
            db.add_all(events)
            db.commit()
            print(f"  已创建 {len(events)} 个活动")
        
        print("\n初始化完成！")
        print("\n示例数据：")
        print("  读者: R001(普通), R002(学生), R003(教师)")
        print("  图书: 3本，每本3个副本")
        print("  座位: 1F-3F，每层30个")
        print("  活动: 2个示例活动")
        
    finally:
        db.close()


if __name__ == "__main__":
    init_sample_data()
