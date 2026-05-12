import { v4 as uuidv4 } from 'uuid';
import { db, SHIFT_TYPES, STAFF_ROLES } from '../db';
import { Shift, ShiftAssignment, Staff, ShiftType, StaffRole } from '../types';

const VALID_SHIFT_TYPES: ShiftType[] = ['morning', 'afternoon', 'night'];
const VALID_STAFF_ROLES: StaffRole[] = ['maternal_care', 'pediatric_nurse'];

export const getStaffById = (id: string): Staff | undefined => {
  const row = db.prepare('SELECT * FROM staff WHERE id = ?').get(id);
  return row as Staff | undefined;
};

export const getShiftById = (id: string): Shift | undefined => {
  const row = db.prepare('SELECT * FROM shifts WHERE id = ?').get(id);
  return row as Shift | undefined;
};

export const getShiftByTypeAndDate = (shiftType: string, shiftDate: string): Shift | undefined => {
  const row = db.prepare('SELECT * FROM shifts WHERE shift_type = ? AND shift_date = ?').get(shiftType, shiftDate);
  return row as Shift | undefined;
};

export const createStaff = (name: string, role: string): Staff => {
  if (!VALID_STAFF_ROLES.includes(role as StaffRole)) {
    throw new Error('INVALID_STAFF_ROLE');
  }

  const id = uuidv4();
  db.prepare('INSERT INTO staff (id, name, role) VALUES (?, ?, ?)').run(id, name, role);
  
  return {
    id,
    name,
    role: role as StaffRole
  };
};

export const createShift = (shiftType: string, shiftDate: string): Shift => {
  if (!VALID_SHIFT_TYPES.includes(shiftType as ShiftType)) {
    throw new Error('INVALID_SHIFT_TYPE');
  }

  const existing = getShiftByTypeAndDate(shiftType, shiftDate);
  if (existing) {
    return existing;
  }

  const id = uuidv4();
  db.prepare('INSERT INTO shifts (id, shift_type, shift_date) VALUES (?, ?, ?)').run(id, shiftType, shiftDate);
  
  return {
    id,
    shift_type: shiftType as ShiftType,
    shift_date: shiftDate
  };
};

export const getStaffShiftsInRange = (staffId: string, startDate: string, endDate: string): Shift[] => {
  const rows = db.prepare(`
    SELECT s.* FROM shifts s
    JOIN shift_assignments sa ON s.id = sa.shift_id
    WHERE sa.staff_id = ? AND s.shift_date >= ? AND s.shift_date <= ?
    ORDER BY s.shift_date
  `).all(staffId, startDate, endDate) as Shift[];
  
  return rows;
};

export const checkStaffWorkDays = (staffId: string, shiftDate: string): boolean => {
  const shiftDateObj = new Date(shiftDate);
  const weekStart = new Date(shiftDateObj);
  weekStart.setDate(shiftDateObj.getDate() - shiftDateObj.getDay());
  const weekEnd = new Date(weekStart);
  weekEnd.setDate(weekStart.getDate() + 6);

  const weekStartStr = weekStart.toISOString().split('T')[0];
  const weekEndStr = weekEnd.toISOString().split('T')[0];

  const shifts = getStaffShiftsInRange(staffId, weekStartStr, weekEndStr);
  
  const workDays = new Set(shifts.map(s => s.shift_date));
  
  if (workDays.size >= 7) {
    return false;
  }

  const sortedDates = [...workDays].sort();
  for (let i = 0; i < sortedDates.length; i++) {
    let consecutiveDays = 1;
    const currentDate = new Date(sortedDates[i]);
    
    for (let j = i + 1; j < sortedDates.length; j++) {
      const nextDate = new Date(sortedDates[j]);
      const diffTime = nextDate.getTime() - currentDate.getTime();
      const diffDays = diffTime / (1000 * 60 * 60 * 24);
      
      if (diffDays === 1) {
        consecutiveDays++;
        if (consecutiveDays > 6) {
          return false;
        }
      } else {
        break;
      }
    }
  }

  return true;
};

export const checkShiftHasBothRoles = (shiftId: string): boolean => {
  const assignments = db.prepare(`
    SELECT s.role FROM shift_assignments sa
    JOIN staff s ON sa.staff_id = s.id
    WHERE sa.shift_id = ?
  `).all(shiftId) as { role: string }[];

  const hasMaternal = assignments.some(a => a.role === STAFF_ROLES.MATERNAL_CARE);
  const hasPediatric = assignments.some(a => a.role === STAFF_ROLES.PEDIATRIC_NURSE);

  return hasMaternal && hasPediatric;
};

export const checkStaffShiftConflict = (staffId: string, shiftId: string, excludeAssignmentId?: string): boolean => {
  const shift = getShiftById(shiftId);
  if (!shift) {
    throw new Error('SHIFT_NOT_FOUND');
  }

  const existingAssignments = db.prepare(`
    SELECT sa.*, s.shift_date, s.shift_type FROM shift_assignments sa
    JOIN shifts s ON sa.shift_id = s.id
    WHERE sa.staff_id = ? AND s.shift_date = ?
    ${excludeAssignmentId ? 'AND sa.id != ?' : ''}
  `).all(staffId, shift.shift_date, ...(excludeAssignmentId ? [excludeAssignmentId] : [])) as (ShiftAssignment & Shift)[];

  const hasSameShift = existingAssignments.some(a => 
    a.shift_type === shift.shift_type && (!excludeAssignmentId || a.id !== excludeAssignmentId)
  );

  return hasSameShift;
};

export const assignStaffToShift = (staffId: string, shiftId: string): ShiftAssignment => {
  const staff = getStaffById(staffId);
  if (!staff) {
    throw new Error('STAFF_NOT_FOUND');
  }

  const shift = getShiftById(shiftId);
  if (!shift) {
    throw new Error('SHIFT_NOT_FOUND');
  }

  if (checkStaffShiftConflict(staffId, shiftId)) {
    throw new Error('SHIFT_CONFLICT');
  }

  if (!checkStaffWorkDays(staffId, shift.shift_date)) {
    throw new Error('WORK_DAYS_RULE_VIOLATION');
  }

  const existingAssignment = db.prepare(`
    SELECT * FROM shift_assignments WHERE staff_id = ? AND shift_id = ?
  `).get(staffId, shiftId);

  if (existingAssignment) {
    return existingAssignment as ShiftAssignment;
  }

  const id = uuidv4();
  db.prepare('INSERT INTO shift_assignments (id, staff_id, shift_id) VALUES (?, ?, ?)').run(id, staffId, shiftId);

  if (!checkShiftHasBothRoles(shiftId)) {
    const currentAssignments = db.prepare(`
      SELECT s.role FROM shift_assignments sa
      JOIN staff s ON sa.staff_id = s.id
      WHERE sa.shift_id = ? AND sa.staff_id != ?
    `).all(shiftId, staffId) as { role: string }[];

    const hasMaternal = currentAssignments.some(a => a.role === STAFF_ROLES.MATERNAL_CARE) || 
                       staff.role === STAFF_ROLES.MATERNAL_CARE;
    const hasPediatric = currentAssignments.some(a => a.role === STAFF_ROLES.PEDIATRIC_NURSE) || 
                        staff.role === STAFF_ROLES.PEDIATRIC_NURSE;

    if (!hasMaternal || !hasPediatric) {
      db.prepare('DELETE FROM shift_assignments WHERE id = ?').run(id);
      throw new Error('SHIFT_ROLES_INCOMPLETE');
    }
  }

  return {
    id,
    staff_id: staffId,
    shift_id: shiftId
  };
};

export const swapShifts = (staff1Id: string, shift1Id: string, staff2Id: string, shift2Id: string): {
  assignment1: ShiftAssignment;
  assignment2: ShiftAssignment;
} => {
  const staff1 = getStaffById(staff1Id);
  const staff2 = getStaffById(staff2Id);
  const shift1 = getShiftById(shift1Id);
  const shift2 = getShiftById(shift2Id);

  if (!staff1 || !staff2) {
    throw new Error('STAFF_NOT_FOUND');
  }

  if (!shift1 || !shift2) {
    throw new Error('SHIFT_NOT_FOUND');
  }

  const assignment1 = db.prepare(`
    SELECT * FROM shift_assignments WHERE staff_id = ? AND shift_id = ?
  `).get(staff1Id, shift1Id) as ShiftAssignment | undefined;

  const assignment2 = db.prepare(`
    SELECT * FROM shift_assignments WHERE staff_id = ? AND shift_id = ?
  `).get(staff2Id, shift2Id) as ShiftAssignment | undefined;

  if (!assignment1 || !assignment2) {
    throw new Error('ASSIGNMENT_NOT_FOUND');
  }

  const transaction = db.transaction(() => {
    db.prepare('DELETE FROM shift_assignments WHERE id = ?').run(assignment1.id);
    db.prepare('DELETE FROM shift_assignments WHERE id = ?').run(assignment2.id);

    if (checkStaffShiftConflict(staff1Id, shift2Id, assignment2.id)) {
      throw new Error('SHIFT_CONFLICT');
    }

    if (!checkStaffWorkDays(staff1Id, shift2.shift_date)) {
      throw new Error('WORK_DAYS_RULE_VIOLATION');
    }

    if (checkStaffShiftConflict(staff2Id, shift1Id, assignment1.id)) {
      throw new Error('SHIFT_CONFLICT');
    }

    if (!checkStaffWorkDays(staff2Id, shift1.shift_date)) {
      throw new Error('WORK_DAYS_RULE_VIOLATION');
    }

    const newId1 = uuidv4();
    const newId2 = uuidv4();

    db.prepare('INSERT INTO shift_assignments (id, staff_id, shift_id) VALUES (?, ?, ?)').run(newId1, staff1Id, shift2Id);
    db.prepare('INSERT INTO shift_assignments (id, staff_id, shift_id) VALUES (?, ?, ?)').run(newId2, staff2Id, shift1Id);

    if (!checkShiftHasBothRoles(shift1Id) || !checkShiftHasBothRoles(shift2Id)) {
      throw new Error('SHIFT_ROLES_INCOMPLETE');
    }

    return {
      assignment1: { id: newId1, staff_id: staff1Id, shift_id: shift2Id },
      assignment2: { id: newId2, staff_id: staff2Id, shift_id: shift1Id }
    };
  });

  try {
    return transaction();
  } catch (e) {
    throw e;
  }
};

export const getAllStaff = (): Staff[] => {
  return db.prepare('SELECT * FROM staff').all() as Staff[];
};

export const getAllShifts = (): Shift[] => {
  return db.prepare('SELECT * FROM shifts ORDER BY shift_date, shift_type').all() as Shift[];
};

export const getShiftAssignments = (shiftId: string): (ShiftAssignment & Staff)[] => {
  return db.prepare(`
    SELECT sa.*, s.name, s.role FROM shift_assignments sa
    JOIN staff s ON sa.staff_id = s.id
    WHERE sa.shift_id = ?
  `).all(shiftId) as (ShiftAssignment & Staff)[];
};
