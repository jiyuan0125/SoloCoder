import { db } from '../db';

interface CreateEventParams {
  title: string;
  start_time: string;
  end_time: string;
  is_special?: boolean;
  max_capacity?: number;
}

interface CreateBookingParams {
  event_id: number;
  phone: string;
  seats: number;
}

export class BookingService {
  private getEventById(id: number) {
    return db.prepare('SELECT * FROM public_events WHERE id = ?').get(id);
  }

  private getPhoneBookedSeats(eventId: number, phone: string): number {
    const result = db.prepare(`
      SELECT COALESCE(SUM(seats), 0) as total
      FROM bookings 
      WHERE event_id = ? AND phone = ? AND is_waitlist = 0
    `).get(eventId, phone) as { total: number };
    return result.total;
  }

  private processWaitlist(eventId: number, availableSeats: number) {
    if (availableSeats <= 0) return;

    const waitlist = db.prepare(`
      SELECT * FROM bookings 
      WHERE event_id = ? AND is_waitlist = 1 
      ORDER BY created_at ASC
    `).all(eventId) as any[];

    let remaining = availableSeats;

    for (const wl of waitlist) {
      if (remaining <= 0) break;

      const seatsToPromote = Math.min(wl.seats, remaining);
      
      db.prepare(`
        UPDATE bookings SET is_waitlist = 0 WHERE id = ?
      `).run(wl.id);

      db.prepare(`
        UPDATE public_events SET current_bookings = current_bookings + ? WHERE id = ?
      `).run(seatsToPromote, eventId);

      remaining -= seatsToPromote;
    }
  }

  private validateSaturday(startTime: string): boolean {
    const date = new Date(startTime);
    return date.getDay() === 6;
  }

  createEvent(params: CreateEventParams) {
    const maxCapacity = params.max_capacity || 50;
    
    if (!params.is_special && !this.validateSaturday(params.start_time)) {
      return { error: '普通科普活动只能在周六举行', status: 400 };
    }

    const stmt = db.prepare(`
      INSERT INTO public_events (title, start_time, end_time, max_capacity, is_special)
      VALUES (?, ?, ?, ?, ?)
    `);
    const result = stmt.run(
      params.title,
      params.start_time,
      params.end_time,
      maxCapacity,
      params.is_special ? 1 : 0
    );
    return { data: this.getEventById(result.lastInsertRowid as number) };
  }

  getEvent(id: number) {
    return { data: this.getEventById(id) };
  }

  getUpcomingEvents() {
    const events = db.prepare(`
      SELECT * FROM public_events 
      WHERE start_time >= datetime('now') AND is_canceled = 0
      ORDER BY start_time
    `).all();
    return { data: events };
  }

  createBooking(params: CreateBookingParams) {
    const event = this.getEventById(params.event_id) as any;
    if (!event) {
      return { error: '活动不存在', status: 404 };
    }
    if (event.is_canceled) {
      return { error: '活动已取消', status: 400 };
    }

    const currentBooked = this.getPhoneBookedSeats(params.event_id, params.phone);
    if (currentBooked + params.seats > 4) {
      return { 
        error: `一个手机号最多预约4个名额，当前已预约${currentBooked}个`, 
        status: 400 
      };
    }

    const availableSeats = event.max_capacity - event.current_bookings;
    const isWaitlist = availableSeats < params.seats;

    const stmt = db.prepare(`
      INSERT INTO bookings (event_id, phone, seats, is_waitlist)
      VALUES (?, ?, ?, ?)
    `);

    if (!isWaitlist) {
      db.prepare(`
        UPDATE public_events SET current_bookings = current_bookings + ? WHERE id = ?
      `).run(params.seats, params.event_id);
    }

    const result = stmt.run(params.event_id, params.phone, params.seats, isWaitlist ? 1 : 0);

    const booking = db.prepare('SELECT * FROM bookings WHERE id = ?').get(result.lastInsertRowid as number);
    return { 
      data: booking, 
      message: isWaitlist ? '名额已满，已加入候补列表' : '预约成功'
    };
  }

  cancelBooking(bookingId: number) {
    const booking = db.prepare('SELECT * FROM bookings WHERE id = ?').get(bookingId) as any;
    if (!booking) {
      return { error: '预约不存在', status: 404 };
    }

    const event = this.getEventById(booking.event_id) as any;
    if (!event) {
      return { error: '活动不存在', status: 404 };
    }

    const eventStart = new Date(event.start_time);
    const now = new Date();
    const hoursDiff = (eventStart.getTime() - now.getTime()) / (1000 * 60 * 60);

    if (hoursDiff < 24) {
      return { error: '活动开始前24小时内不能取消', status: 400 };
    }

    db.prepare('DELETE FROM bookings WHERE id = ?').run(bookingId);

    if (!booking.is_waitlist) {
      db.prepare(`
        UPDATE public_events SET current_bookings = current_bookings - ? WHERE id = ?
      `).run(booking.seats, booking.event_id);

      this.processWaitlist(booking.event_id, booking.seats);
    }

    return { data: { message: '取消成功' } };
  }

  autoCheckAndCancel() {
    const events = db.prepare(`
      SELECT * FROM public_events 
      WHERE is_canceled = 0 
        AND start_time > datetime('now')
        AND datetime('now') >= datetime(start_time, '-48 hours')
    `).all() as any[];

    const canceledEvents: any[] = [];

    for (const event of events) {
      if (event.current_bookings < 10) {
        db.prepare(`
          UPDATE public_events SET is_canceled = 1 WHERE id = ?
        `).run(event.id);

        const bookings = db.prepare(`
          SELECT phone FROM bookings WHERE event_id = ? AND is_waitlist = 0
        `).all(event.id) as any[];

        canceledEvents.push({
          event_id: event.id,
          title: event.title,
          start_time: event.start_time,
          notified_phones: bookings.map(b => b.phone)
        });

        db.prepare('DELETE FROM bookings WHERE event_id = ?').run(event.id);
      }
    }

    return { data: canceledEvents };
  }

  getBookingsByEvent(eventId: number) {
    const bookings = db.prepare(`
      SELECT * FROM bookings WHERE event_id = ? ORDER BY is_waitlist, created_at
    `).all(eventId);
    return { data: bookings };
  }

  getMyBookings(phone: string) {
    const bookings = db.prepare(`
      SELECT b.*, pe.title, pe.start_time, pe.end_time, pe.is_canceled as event_canceled
      FROM bookings b
      JOIN public_events pe ON b.event_id = pe.id
      WHERE b.phone = ?
      ORDER BY pe.start_time DESC
    `).all(phone);
    return { data: bookings };
  }
}

export const bookingService = new BookingService();
