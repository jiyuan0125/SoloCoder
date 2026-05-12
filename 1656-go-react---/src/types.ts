export interface DeviceInfo {
  browserType: string;
  browserVersion: string;
  operatingSystem: string;
  screenResolution: string;
  timezone: string;
  languagePreference: string;
}

export interface TrustRecord {
  id: number;
  userId: string;
  deviceFingerprint: string;
  firstTrustedAt: string;
  lastUsedAt: string;
  status: TrustStatus;
}

export enum TrustStatus {
  TRUSTED = 'trusted',
  REVOKED = 'revoked',
}

export const DEVICE_INFO_FIELDS: (keyof DeviceInfo)[] = [
  'browserType',
  'browserVersion',
  'operatingSystem',
  'screenResolution',
  'timezone',
  'languagePreference',
];

export const TRUST_DURATION_DAYS = 30;
