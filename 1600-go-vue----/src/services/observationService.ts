import { db } from '../db';
import { TaskStatus, CancelReason } from '../types';

const STATUS_FLOW: TaskStatus[] = [
  TaskStatus.PENDING,
  TaskStatus.SCHEDULED,
  TaskStatus.OBSERVING,
  TaskStatus.COMPLETED,
  TaskStatus.ARCHIVED,
];

interface CreateTaskParams {
  target_object: string;
  responsible_person: string;
  scheduled_month: string;
}

interface ScheduleTaskParams {
  task_id: number;
  equipment_id: number;
  start_time: string;
  end_time: string;
}

interface AdjustEquipmentParams {
  task_id: number;
  new_equipment_id: number;
  new_start_time?: string;
  new_end_time?: string;
}

export class ObservationService {
  private checkTimeOverlap(
    equipmentId: number,
    startTime: string,
    endTime: string,
    excludeTaskId?: number
  ): any[] {
    const baseQuery = `
      SELECT eb.*, ot.target_object, ot.responsible_person
      FROM equipment_bookings eb
      JOIN observation_tasks ot ON eb.task_id = ot.id
      WHERE eb.equipment_id = ? 
        AND NOT (eb.end_time <= ? OR eb.start_time >= ?)
        AND ot.is_canceled = 0
    `;
    
    let query = baseQuery;
    const params: any[] = [equipmentId, startTime, endTime];
    
    if (excludeTaskId) {
      query += ' AND eb.task_id != ?';
      params.push(excludeTaskId);
    }
    
    return db.prepare(query).all(...params) as any[];
  }

  createTask(params: CreateTaskParams) {
    const stmt = db.prepare(`
      INSERT INTO observation_tasks (target_object, responsible_person, scheduled_month)
      VALUES (?, ?, ?)
    `);
    const result = stmt.run(params.target_object, params.responsible_person, params.scheduled_month);
    return this.getTask(result.lastInsertRowid as number);
  }

  getTask(id: number) {
    return db.prepare('SELECT * FROM observation_tasks WHERE id = ?').get(id);
  }

  getTasksByMonth(month: string) {
    return db.prepare('SELECT * FROM observation_tasks WHERE scheduled_month = ? ORDER BY start_time').all(month);
  }

  getCalendar(startTime: string, endTime: string) {
    return db.prepare(`
      SELECT ot.*, eb.equipment_id
      FROM observation_tasks ot
      LEFT JOIN equipment_bookings eb ON ot.id = eb.task_id
      WHERE ot.status = '已排期' 
        AND ot.is_canceled = 0
        AND NOT (ot.end_time <= ? OR ot.start_time >= ?)
      ORDER BY ot.start_time
    `).all(startTime, endTime);
  }

  scheduleTask(params: ScheduleTaskParams) {
    const task = this.getTask(params.task_id) as any;
    if (!task) {
      return { error: '任务不存在', status: 404 };
    }
    if (task.status !== TaskStatus.PENDING) {
      return { 
        error: `任务当前状态为「${task.status}」，只能从「待排期」状态进行排期`, 
        status: 400 
      };
    }

    const equipment = db.prepare('SELECT * FROM equipment WHERE id = ?').get(params.equipment_id) as any;
    if (!equipment) {
      return { error: '设备不存在', status: 404 };
    }

    const conflicts = this.checkTimeOverlap(params.equipment_id, params.start_time, params.end_time);
    if (conflicts.length > 0) {
      const conflictInfo = conflicts.map(c => 
        `任务「${c.target_object}」(负责人: ${c.responsible_person}) 时段: ${c.start_time} ~ ${c.end_time}`
      ).join('; ');
      return { 
        error: `设备「${equipment.name}」在该时段已被占用: ${conflictInfo}`, 
        status: 409 
      };
    }

    const tx = db.transaction(() => {
      db.prepare(`
        UPDATE observation_tasks 
        SET status = ?, equipment_id = ?, equipment_name = ?, start_time = ?, end_time = ?, updated_at = datetime('now')
        WHERE id = ?
      `).run(TaskStatus.SCHEDULED, params.equipment_id, equipment.name, params.start_time, params.end_time, params.task_id);

      db.prepare(`
        INSERT INTO equipment_bookings (equipment_id, task_id, start_time, end_time)
        VALUES (?, ?, ?, ?)
      `).run(params.equipment_id, params.task_id, params.start_time, params.end_time);
    });

    tx();
    return { data: this.getTask(params.task_id) };
  }

  advanceStatus(taskId: number) {
    const task = this.getTask(taskId) as any;
    if (!task) {
      return { error: '任务不存在', status: 404 };
    }
    if (task.is_canceled) {
      return { error: '任务已取消，无法继续状态流转', status: 400 };
    }

    const currentIndex = STATUS_FLOW.indexOf(task.status);
    if (currentIndex === -1 || currentIndex >= STATUS_FLOW.length - 1) {
      return { 
        error: `任务当前状态为「${task.status}」，已是最终状态`, 
        status: 400 
      };
    }

    const nextStatus = STATUS_FLOW[currentIndex + 1];
    db.prepare(`
      UPDATE observation_tasks 
      SET status = ?, updated_at = datetime('now')
      WHERE id = ?
    `).run(nextStatus, taskId);

    return { data: this.getTask(taskId) };
  }

  cancelTask(taskId: number, reason: CancelReason, faultDescription?: string) {
    const task = this.getTask(taskId) as any;
    if (!task) {
      return { error: '任务不存在', status: 404 };
    }
    if (task.status !== TaskStatus.SCHEDULED) {
      return { 
        error: `任务当前状态为「${task.status}」，只能从「已排期」状态进行取消`, 
        status: 400 
      };
    }
    if (reason === CancelReason.EQUIPMENT_FAULT && !faultDescription) {
      return { error: '设备故障取消需要提供故障原因', status: 400 };
    }

    const tx = db.transaction(() => {
      db.prepare(`
        UPDATE observation_tasks 
        SET is_canceled = 1, cancel_reason = ?, fault_description = ?, updated_at = datetime('now')
        WHERE id = ?
      `).run(reason, faultDescription || null, taskId);

      db.prepare('DELETE FROM equipment_bookings WHERE task_id = ?').run(taskId);
    });

    tx();
    return { data: this.getTask(taskId) };
  }

  adjustEquipment(params: AdjustEquipmentParams) {
    const task = this.getTask(params.task_id) as any;
    if (!task) {
      return { error: '任务不存在', status: 404 };
    }
    if (task.status !== TaskStatus.SCHEDULED) {
      return { 
        error: `任务当前状态为「${task.status}」，只能在「已排期」状态调整设备`, 
        status: 400 
      };
    }
    if (task.is_canceled) {
      return { error: '任务已取消', status: 400 };
    }

    const newEquipment = db.prepare('SELECT * FROM equipment WHERE id = ?').get(params.new_equipment_id) as any;
    if (!newEquipment) {
      return { error: '新设备不存在', status: 404 };
    }

    const newStartTime = params.new_start_time || task.start_time;
    const newEndTime = params.new_end_time || task.end_time;

    const conflicts = this.checkTimeOverlap(params.new_equipment_id, newStartTime, newEndTime, params.task_id);
    if (conflicts.length > 0) {
      const conflictInfo = conflicts.map(c => 
        `任务「${c.target_object}」(负责人: ${c.responsible_person}) 时段: ${c.start_time} ~ ${c.end_time}`
      ).join('; ');
      return { 
        error: `新设备「${newEquipment.name}」在该时段已被占用: ${conflictInfo}`, 
        status: 409 
      };
    }

    const tx = db.transaction(() => {
      db.prepare('DELETE FROM equipment_bookings WHERE task_id = ?').run(params.task_id);

      db.prepare(`
        INSERT INTO equipment_bookings (equipment_id, task_id, start_time, end_time)
        VALUES (?, ?, ?, ?)
      `).run(params.new_equipment_id, params.task_id, newStartTime, newEndTime);

      db.prepare(`
        UPDATE observation_tasks 
        SET equipment_id = ?, equipment_name = ?, start_time = ?, end_time = ?, updated_at = datetime('now')
        WHERE id = ?
      `).run(params.new_equipment_id, newEquipment.name, newStartTime, newEndTime, params.task_id);
    });

    tx();
    return { data: this.getTask(params.task_id) };
  }

  getEquipmentList() {
    return db.prepare('SELECT * FROM equipment ORDER BY id').all();
  }
}

export const observationService = new ObservationService();
