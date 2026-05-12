import { v4 as uuidv4 } from 'uuid';
import { db, ROOM_PRICES } from '../db';
import { Mother, RoomType } from '../types';

const VALID_ROOM_TYPES: RoomType[] = ['standard', 'deluxe', 'suite'];

export const isValidRoomType = (roomType: string): roomType is RoomType => {
  return VALID_ROOM_TYPES.includes(roomType as RoomType);
};

export const isValidDate = (dateStr: string): boolean => {
  const date = new Date(dateStr);
  return !isNaN(date.getTime());
};

export const formatDate = (date: Date): string => {
  return date.toISOString().split('T')[0];
};

export const getDateDifference = (start: Date, end: Date): number => {
  const startDate = new Date(start.toDateString());
  const endDate = new Date(end.toDateString());
  const diffTime = endDate.getTime() - startDate.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  return Math.max(diffDays, 1);
};

export const getMotherById = (id: string): Mother | undefined => {
  const row = db.prepare('SELECT * FROM mothers WHERE id = ?').get(id);
  return row as Mother | undefined;
};

export const createBooking = (
  name: string,
  expectedDueDate: string,
  roomType: string,
  emergencyContactName: string,
  emergencyContactPhone: string,
  expectedStayDays: number
): Mother => {
  if (!isValidRoomType(roomType)) {
    throw new Error('INVALID_ROOM_TYPE');
  }

  if (!isValidDate(expectedDueDate)) {
    throw new Error('INVALID_DATE');
  }

  const dueDate = new Date(expectedDueDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  if (dueDate < today) {
    throw new Error('DUE_DATE_IN_PAST');
  }

  const id = uuidv4();
  const roomPrice = ROOM_PRICES[roomType];
  const totalDeposit = roomPrice * expectedStayDays;
  const todayStr = formatDate(new Date());

  const insertStmt = db.prepare(`
    INSERT INTO mothers (id, name, expected_due_date, room_type, emergency_contact_name, 
                          emergency_contact_phone, booking_date, status, total_deposit)
    VALUES (?, ?, ?, ?, ?, ?, ?, 'booked', ?)
  `);

  insertStmt.run(id, name, expectedDueDate, roomType, emergencyContactName, emergencyContactPhone, todayStr, totalDeposit);

  return {
    id,
    name,
    expected_due_date: expectedDueDate,
    room_type: roomType as RoomType,
    emergency_contact_name: emergencyContactName,
    emergency_contact_phone: emergencyContactPhone,
    booking_date: todayStr,
    status: 'booked',
    total_deposit: totalDeposit
  };
};

export const checkIn = (motherId: string, checkInDate?: string): Mother => {
  const mother = getMotherById(motherId);
  if (!mother) {
    throw new Error('MOTHER_NOT_FOUND');
  }

  if (mother.status !== 'booked') {
    throw new Error('INVALID_STATUS');
  }

  const actualCheckIn = checkInDate ? new Date(checkInDate) : new Date();
  if (!isValidDate(formatDate(actualCheckIn))) {
    throw new Error('INVALID_DATE');
  }

  const dueDate = new Date(mother.expected_due_date);
  const checkInDay = new Date(actualCheckIn.toDateString());
  const dueDateDay = new Date(dueDate.toDateString());
  const diffTime = dueDateDay.getTime() - checkInDay.getTime();
  const diffDays = diffTime / (1000 * 60 * 60 * 24);

  if (diffDays > 15) {
    throw new Error('CHECK_IN_TOO_EARLY');
  }

  const updateStmt = db.prepare(`
    UPDATE mothers SET check_in_date = ?, status = 'checked_in' WHERE id = ?
  `);

  const checkInStr = formatDate(actualCheckIn);
  updateStmt.run(checkInStr, motherId);

  const updated = getMotherById(motherId);
  if (!updated) {
    throw new Error('MOTHER_NOT_FOUND');
  }

  return updated;
};

export const checkOut = (motherId: string, checkOutDate?: string): { mother: Mother; finalAmount: number; refund: number } => {
  const mother = getMotherById(motherId);
  if (!mother) {
    throw new Error('MOTHER_NOT_FOUND');
  }

  if (mother.status !== 'checked_in') {
    throw new Error('INVALID_STATUS');
  }

  if (!mother.check_in_date) {
    throw new Error('NOT_CHECKED_IN');
  }

  const actualCheckOut = checkOutDate ? new Date(checkOutDate) : new Date();
  const checkInDate = new Date(mother.check_in_date);

  if (actualCheckOut < checkInDate) {
    throw new Error('CHECKOUT_BEFORE_CHECKIN');
  }

  const stayDays = getDateDifference(checkInDate, actualCheckOut);
  const dailyRate = ROOM_PRICES[mother.room_type];
  const finalAmount = dailyRate * stayDays;
  const refund = mother.total_deposit - finalAmount;

  const updateStmt = db.prepare(`
    UPDATE mothers SET check_out_date = ?, status = 'checked_out' WHERE id = ?
  `);

  const checkOutStr = formatDate(actualCheckOut);
  updateStmt.run(checkOutStr, motherId);

  const updated = getMotherById(motherId);
  if (!updated) {
    throw new Error('MOTHER_NOT_FOUND');
  }

  return {
    mother: updated,
    finalAmount,
    refund
  };
};

export const getAllMothers = (): Mother[] => {
  const rows = db.prepare('SELECT * FROM mothers').all();
  return rows as Mother[];
};
