import express from 'express';
import './db';
import { observationService } from './services/observationService';
import { bookingService } from './services/bookingService';
import { reportService } from './services/reportService';

const app = express();
const port = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.get('/api/equipment', (_req, res) => {
  const result = observationService.getEquipmentList();
  res.json({ data: result });
});

app.post('/api/observation/tasks', (req, res) => {
  const { target_object, responsible_person, scheduled_month } = req.body;
  
  if (!target_object || !responsible_person || !scheduled_month) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  const result = observationService.createTask({
    target_object,
    responsible_person,
    scheduled_month,
  });
  
  res.status(201).json({ data: result });
});

app.get('/api/observation/tasks/month/:month', (req, res) => {
  const result = observationService.getTasksByMonth(req.params.month);
  res.json({ data: result || [] });
});

app.get('/api/observation/tasks/:id', (req, res) => {
  const result = observationService.getTask(parseInt(req.params.id, 10));
  if (!result) {
    return res.json({ data: null });
  }
  res.json({ data: result });
});

app.get('/api/observation/calendar', (req, res) => {
  const { start_time, end_time } = req.query as Record<string, string>;
  
  if (!start_time || !end_time) {
    return res.status(400).json({ error: '缺少时间范围参数' });
  }

  const result = observationService.getCalendar(start_time, end_time);
  res.json({ data: result || [] });
});

app.post('/api/observation/tasks/schedule', (req, res) => {
  const { task_id, equipment_id, start_time, end_time } = req.body;
  
  if (!task_id || !equipment_id || !start_time || !end_time) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  const result = observationService.scheduleTask({
    task_id: parseInt(task_id, 10),
    equipment_id: parseInt(equipment_id, 10),
    start_time,
    end_time,
  });

  if (result.error) {
    return res.status(result.status).json({ error: result.error });
  }
  
  res.json({ data: result.data });
});

app.post('/api/observation/tasks/:id/advance', (req, res) => {
  const result = observationService.advanceStatus(parseInt(req.params.id, 10));
  
  if (result.error) {
    return res.status(result.status).json({ error: result.error });
  }
  
  res.json({ data: result.data });
});

app.post('/api/observation/tasks/:id/cancel', (req, res) => {
  const { reason, fault_description } = req.body;
  
  if (!reason) {
    return res.status(400).json({ error: '缺少取消原因' });
  }

  const result = observationService.cancelTask(
    parseInt(req.params.id, 10),
    reason,
    fault_description
  );

  if (result.error) {
    return res.status(result.status).json({ error: result.error });
  }
  
  res.json({ data: result.data });
});

app.post('/api/observation/tasks/:id/adjust-equipment', (req, res) => {
  const { new_equipment_id, new_start_time, new_end_time } = req.body;
  
  if (!new_equipment_id) {
    return res.status(400).json({ error: '缺少新设备ID' });
  }

  const result = observationService.adjustEquipment({
    task_id: parseInt(req.params.id, 10),
    new_equipment_id: parseInt(new_equipment_id, 10),
    new_start_time,
    new_end_time,
  });

  if (result.error) {
    return res.status(result.status).json({ error: result.error });
  }
  
  res.json({ data: result.data });
});

app.post('/api/public/events', (req, res) => {
  const { title, start_time, end_time, is_special, max_capacity } = req.body;
  
  if (!title || !start_time || !end_time) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  const result = bookingService.createEvent({
    title,
    start_time,
    end_time,
    is_special,
    max_capacity,
  });

  if (result.error) {
    return res.status(result.status).json({ error: result.error });
  }
  
  res.status(201).json({ data: result.data });
});

app.get('/api/public/events', (_req, res) => {
  const result = bookingService.getUpcomingEvents();
  res.json({ data: result.data || [] });
});

app.get('/api/public/events/:id', (req, res) => {
  const result = bookingService.getEvent(parseInt(req.params.id, 10));
  res.json({ data: result.data });
});

app.post('/api/public/bookings', (req, res) => {
  const { event_id, phone, seats } = req.body;
  
  if (!event_id || !phone || !seats) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  const result = bookingService.createBooking({
    event_id: parseInt(event_id, 10),
    phone,
    seats: parseInt(seats, 10),
  });

  if (result.error) {
    return res.status(result.status).json({ error: result.error });
  }
  
  res.status(201).json({ data: result.data, message: result.message });
});

app.delete('/api/public/bookings/:id', (req, res) => {
  const result = bookingService.cancelBooking(parseInt(req.params.id, 10));
  
  if (result.error) {
    return res.status(result.status).json({ error: result.error });
  }
  
  res.json({ data: result.data });
});

app.post('/api/public/events/auto-check', (_req, res) => {
  const result = bookingService.autoCheckAndCancel();
  res.json({ data: result.data || [] });
});

app.get('/api/public/events/:id/bookings', (req, res) => {
  const result = bookingService.getBookingsByEvent(parseInt(req.params.id, 10));
  res.json({ data: result.data || [] });
});

app.get('/api/public/bookings/my/:phone', (req, res) => {
  const result = bookingService.getMyBookings(req.params.phone);
  res.json({ data: result.data || [] });
});

app.get('/api/reports/observation/:month', (req, res) => {
  const { equipment_id } = req.query as Record<string, string>;
  
  const result = reportService.getObservationReport(
    req.params.month,
    equipment_id ? parseInt(equipment_id, 10) : undefined
  );
  
  res.json({ data: result.data || [] });
});

app.get('/api/reports/activity/:month', (req, res) => {
  const result = reportService.getActivityReport(req.params.month);
  res.json({ data: result.data || [] });
});

app.get('/api/reports/stats/:month', (req, res) => {
  const result = reportService.getMonthlyStats(req.params.month);
  res.json({ data: result.data });
});

app.use((_req, res) => {
  res.status(404).json({ error: '接口不存在' });
});

app.listen(port, () => {
  console.log(`天文台系统服务运行在端口 ${port}`);
});
