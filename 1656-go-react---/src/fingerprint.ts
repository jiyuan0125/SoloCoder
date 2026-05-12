import crypto from 'crypto';
import { DeviceInfo, DEVICE_INFO_FIELDS } from './types';

export function validateDeviceInfo(deviceInfo: Partial<DeviceInfo>): deviceInfo is DeviceInfo {
  return DEVICE_INFO_FIELDS.every(
    (field) => deviceInfo[field] !== undefined && deviceInfo[field] !== null && deviceInfo[field] !== ''
  );
}

export function generateDeviceFingerprint(deviceInfo: DeviceInfo): string {
  const sortedValues = DEVICE_INFO_FIELDS.map((field) => deviceInfo[field]);
  const concatenated = sortedValues.join('|');
  return crypto.createHash('md5').update(concatenated).digest('hex');
}
