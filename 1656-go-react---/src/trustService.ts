import { Database } from './database';
import { TrustRecord, TrustStatus, TRUST_DURATION_DAYS } from './types';

const verificationLocks = new Map<string, Promise<boolean>>();

function isTrustRecordExpired(record: TrustRecord): boolean {
  const firstTrusted = new Date(record.firstTrustedAt);
  const expirationDate = new Date(firstTrusted);
  expirationDate.setDate(expirationDate.getDate() + TRUST_DURATION_DAYS);
  return new Date() > expirationDate;
}

export async function checkDeviceTrust(
  userId: string,
  deviceFingerprint: string,
  performVerification: () => Promise<boolean>
): Promise<{
  trusted: boolean;
  requiresVerification: boolean;
  alreadyVerifying: boolean;
  message?: string;
}> {
  const lockKey = `${userId}:${deviceFingerprint}`;

  const existingRecord = await Database.getTrustRecord(userId, deviceFingerprint);

  if (!existingRecord) {
    if (verificationLocks.has(lockKey)) {
      return { trusted: false, requiresVerification: true, alreadyVerifying: true };
    }

    const verificationPromise = (async () => {
      try {
        const verified = await performVerification();
        if (verified) {
          await Database.createTrustRecord(userId, deviceFingerprint);
        }
        return verified;
      } finally {
        verificationLocks.delete(lockKey);
      }
    })();

    verificationLocks.set(lockKey, verificationPromise);
    const result = await verificationPromise;
    return { trusted: result, requiresVerification: !result, alreadyVerifying: false };
  }

  if (existingRecord.status === TrustStatus.REVOKED) {
    return { trusted: false, requiresVerification: false, alreadyVerifying: false, message: '设备信任已撤销' };
  }

  if (isTrustRecordExpired(existingRecord)) {
    if (verificationLocks.has(lockKey)) {
      return { trusted: false, requiresVerification: true, alreadyVerifying: true };
    }

    const verificationPromise = (async () => {
      try {
        const verified = await performVerification();
        if (verified) {
          await Database.updateTrustRecordStatus(userId, deviceFingerprint, TrustStatus.TRUSTED);
        }
        return verified;
      } finally {
        verificationLocks.delete(lockKey);
      }
    })();

    verificationLocks.set(lockKey, verificationPromise);
    const result = await verificationPromise;
    return { trusted: result, requiresVerification: !result, alreadyVerifying: false };
  }

  await Database.updateTrustRecordStatus(userId, deviceFingerprint, TrustStatus.TRUSTED);
  return { trusted: true, requiresVerification: false, alreadyVerifying: false };
}

export async function getAllTrustRecords(): Promise<TrustRecord[]> {
  return Database.getAllTrustRecords();
}

export async function revokeTrustById(id: number): Promise<boolean> {
  try {
    await Database.revokeTrustRecordById(id);
    return true;
  } catch (error: any) {
    if (error.message === 'RecordNotFound') {
      return false;
    }
    throw error;
  }
}
