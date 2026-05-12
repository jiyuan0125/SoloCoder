import { Medication, Contraindication, RefillReminder } from '../models/types';

const contraindications: Map<string, Contraindication[]> = new Map();
const currentMedications: Map<string, Medication[]> = new Map();
const refillReminders: Map<string, RefillReminder> = new Map();

const REFILL_REMINDER_DAYS = 7;

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
}

function normalizeDrugName(name: string): string {
  return name.toLowerCase().trim();
}

export function initBuiltinContraindications(): void {
  const builtin: Contraindication[] = [
    {
      drugA: '阿司匹林',
      drugB: '布洛芬',
      description: '布洛芬可能降低阿司匹林的心血管保护作用',
      severity: 'moderate',
    },
    {
      drugA: '华法林',
      drugB: '阿司匹林',
      description: '增加出血风险',
      severity: 'severe',
    },
    {
      drugA: '华法林',
      drugB: '布洛芬',
      description: '增加出血风险',
      severity: 'severe',
    },
    {
      drugA: '二甲双胍',
      drugB: '格列本脲',
      description: '联合使用可能增加低血糖风险',
      severity: 'moderate',
    },
    {
      drugA: '硝苯地平',
      drugB: '西地那非',
      description: '可能导致严重低血压',
      severity: 'severe',
    },
    {
      drugA: '缬沙坦',
      drugB: '氨苯蝶啶',
      description: '增加高血钾风险',
      severity: 'moderate',
    },
    {
      drugA: '辛伐他汀',
      drugB: '红霉素',
      description: '增加横纹肌溶解风险',
      severity: 'severe',
    },
    {
      drugA: '氯吡格雷',
      drugB: '奥美拉唑',
      description: '可能降低氯吡格雷的疗效',
      severity: 'moderate',
    },
  ];

  builtin.forEach(c => addContraindication(c));
}

export function addContraindication(contraindication: Contraindication): void {
  const keyA = normalizeDrugName(contraindication.drugA);
  const keyB = normalizeDrugName(contraindication.drugB);

  if (!contraindications.has(keyA)) {
    contraindications.set(keyA, []);
  }
  if (!contraindications.has(keyB)) {
    contraindications.set(keyB, []);
  }

  contraindications.get(keyA)!.push(contraindication);
  contraindications.get(keyB)!.push(contraindication);
}

export function checkContraindications(
  newMedication: Medication,
  existingMedications: Medication[]
): Contraindication[] {
  const conflicts: Contraindication[] = [];
  const newDrugName = normalizeDrugName(newMedication.name);
  const newGenericName = newMedication.genericName 
    ? normalizeDrugName(newMedication.genericName) 
    : null;

  existingMedications.forEach(med => {
    const existingName = normalizeDrugName(med.name);
    const existingGeneric = med.genericName 
      ? normalizeDrugName(med.genericName) 
      : null;

    const drugContraindications = contraindications.get(newDrugName) || [];
    const genericContraindications = newGenericName 
      ? contraindications.get(newGenericName) || [] 
      : [];

    const allContraindications = [...drugContraindications, ...genericContraindications];

    allContraindications.forEach(c => {
      const conflictName = normalizeDrugName(c.drugA) === newDrugName 
        ? normalizeDrugName(c.drugB) 
        : normalizeDrugName(c.drugA);
      
      if (conflictName === existingName || 
          (existingGeneric && conflictName === existingGeneric) ||
          (newGenericName && conflictName === newGenericName)) {
        if (!conflicts.includes(c)) {
          conflicts.push(c);
        }
      }
    });
  });

  return conflicts;
}

export function getAllContraindications(): Contraindication[] {
  const all: Contraindication[] = [];
  const seen = new Set<string>();

  contraindications.forEach(list => {
    list.forEach(c => {
      const key = `${normalizeDrugName(c.drugA)}|${normalizeDrugName(c.drugB)}`;
      if (!seen.has(key)) {
        seen.add(key);
        all.push(c);
      }
    });
  });

  return all;
}

export function addMedication(
  patientId: string,
  medication: Omit<Medication, 'medicationId'>
): {
  medication: Medication;
  contraindications: Contraindication[];
} {
  const medicationId = generateId();
  const newMedication: Medication = {
    ...medication,
    medicationId,
  };

  const existing = currentMedications.get(patientId) || [];
  const conflicts = checkContraindications(newMedication, existing);

  if (conflicts.length > 0) {
    return {
      medication: newMedication,
      contraindications: conflicts,
    };
  }

  if (!currentMedications.has(patientId)) {
    currentMedications.set(patientId, []);
  }
  currentMedications.get(patientId)!.push(newMedication);

  return {
    medication: newMedication,
    contraindications: [],
  };
}

export function getPatientMedications(patientId: string): Medication[] {
  const meds = currentMedications.get(patientId) || [];
  return meds.map(med => {
    let remainingDays: number | undefined;
    if (med.endDate) {
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      const endDate = new Date(med.endDate);
      endDate.setHours(0, 0, 0, 0);
      const diffDays = Math.ceil((endDate.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));
      remainingDays = diffDays > 0 ? diffDays : 0;
    }
    return { ...med, remainingDays };
  });
}

export function removeMedication(patientId: string, medicationId: string): boolean {
  const meds = currentMedications.get(patientId);
  if (!meds) return false;

  const index = meds.findIndex(m => m.medicationId === medicationId);
  if (index === -1) return false;

  meds.splice(index, 1);
  return true;
}

export function generateRefillReminders(): RefillReminder[] {
  const today = new Date();
  today.setHours(0, 0, 0, 0);

  const newReminders: RefillReminder[] = [];
  const reminderDate = new Date(today);
  reminderDate.setDate(reminderDate.getDate() + REFILL_REMINDER_DAYS);

  currentMedications.forEach((meds, patientId) => {
    meds.forEach(med => {
      if (!med.endDate) return;

      const endDate = new Date(med.endDate);
      endDate.setHours(0, 0, 0, 0);

      const daysUntilEnd = Math.ceil(
        (endDate.getTime() - today.getTime()) / (1000 * 60 * 60 * 24)
      );

      if (daysUntilEnd <= REFILL_REMINDER_DAYS && daysUntilEnd >= 0) {
        const existingReminder = Array.from(refillReminders.values()).find(
          r => r.patientId === patientId && 
               r.medicationId === med.medicationId &&
               !r.sent
        );

        if (!existingReminder) {
          const reminder: RefillReminder = {
            id: generateId(),
            patientId,
            medicationId: med.medicationId,
            medicationName: med.name,
            expirationDate: med.endDate,
            reminderDate: today,
            sent: false,
            createdAt: new Date(),
          };
          refillReminders.set(reminder.id, reminder);
          newReminders.push(reminder);
        }
      }
    });
  });

  return newReminders;
}

export interface SendReminderResult {
  success: boolean;
  reminder: RefillReminder;
  message: string;
}

export function sendRefillReminder(reminderId: string): SendReminderResult {
  const reminder = refillReminders.get(reminderId);
  if (!reminder) {
    return {
      success: false,
      reminder: {
        id: reminderId,
        patientId: '',
        medicationId: '',
        medicationName: '',
        expirationDate: new Date(),
        reminderDate: new Date(),
        sent: false,
        createdAt: new Date(),
      },
      message: '提醒不存在',
    };
  }

  if (reminder.sent) {
    return {
      success: false,
      reminder,
      message: '提醒已经发送过',
    };
  }

  const sentSuccessfully = Math.random() > 0.1;

  if (sentSuccessfully) {
    reminder.sent = true;
    reminder.sentAt = new Date();
    refillReminders.set(reminderId, reminder);
    return {
      success: true,
      reminder,
      message: '续药提醒已成功发送',
    };
  } else {
    return {
      success: false,
      reminder,
      message: '提醒发送失败，将在下次调度周期重试',
    };
  }
}

export function processRefillReminders(): {
  generated: RefillReminder[];
  sentResults: SendReminderResult[];
} {
  const generated = generateRefillReminders();
  const sentResults: SendReminderResult[] = [];

  const pendingReminders = Array.from(refillReminders.values()).filter(r => !r.sent);
  pendingReminders.forEach(reminder => {
    const result = sendRefillReminder(reminder.id);
    sentResults.push(result);
  });

  return {
    generated,
    sentResults,
  };
}

export function getPendingReminders(): RefillReminder[] {
  return Array.from(refillReminders.values()).filter(r => !r.sent);
}

export function getAllReminders(): RefillReminder[] {
  return Array.from(refillReminders.values());
}

export function updateContraindications(newContraindications: Contraindication[]): void {
  contraindications.clear();
  newContraindications.forEach(c => addContraindication(c));
}
