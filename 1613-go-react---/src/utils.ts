import { ReleaseStatus, STATUS_TRANSITIONS, STATUS_TRANSITION_MESSAGES } from './types';

export interface ParsedVersion {
  major: number;
  minor: number;
  patch: number;
}

export function parseVersion(version: string): ParsedVersion | null {
  const regex = /^(\d+)\.(\d+)\.(\d+)$/;
  const match = version.match(regex);
  
  if (!match) {
    return null;
  }

  const major = parseInt(match[1], 10);
  const minor = parseInt(match[2], 10);
  const patch = parseInt(match[3], 10);

  if (major > 65535 || minor > 65535 || patch > 65535) {
    return null;
  }

  return { major, minor, patch };
}

export function isValidVersion(version: string): boolean {
  return parseVersion(version) !== null;
}

export function formatVersion(parsed: ParsedVersion): string {
  return `${parsed.major}.${parsed.minor}.${parsed.patch}`;
}

export function compareVersions(a: ParsedVersion, b: ParsedVersion): number {
  if (a.major !== b.major) {
    return a.major - b.major;
  }
  if (a.minor !== b.minor) {
    return a.minor - b.minor;
  }
  return a.patch - b.patch;
}

export function compareVersionStrings(a: string, b: string): number {
  const parsedA = parseVersion(a);
  const parsedB = parseVersion(b);
  
  if (!parsedA || !parsedB) {
    throw new Error('Invalid version string');
  }
  
  return compareVersions(parsedA, parsedB);
}

export function isVersionLessThan(a: string, b: string): boolean {
  return compareVersionStrings(a, b) < 0;
}

export function isVersionGreaterThan(a: string, b: string): boolean {
  return compareVersionStrings(a, b) > 0;
}

export function truncateChangelog(changelog: string): string {
  if (changelog.length <= 500) {
    return changelog;
  }
  return changelog.substring(0, 500);
}

export function isValidStatusTransition(from: ReleaseStatus, to: ReleaseStatus): { valid: boolean; message: string } {
  const allowedTransitions = STATUS_TRANSITIONS.get(from);
  
  if (!allowedTransitions || !allowedTransitions.has(to)) {
    return {
      valid: false,
      message: STATUS_TRANSITION_MESSAGES.get(from) || '无效的状态流转'
    };
  }
  
  return { valid: true, message: '' };
}

export function isValidBuildNumber(buildNumber: number): boolean {
  return Number.isInteger(buildNumber) && buildNumber > 0;
}

export function hashStringToPercentage(key: string): number {
  let hash = 0;
  for (let i = 0; i < key.length; i++) {
    const char = key.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash;
  }
  return (Math.abs(hash) % 100);
}

export function isUserInGrayScale(
  userId: string,
  percentage: number,
  whitelist: Set<string>
): boolean {
  if (whitelist.has(userId)) {
    return true;
  }
  
  if (percentage <= 0) {
    return false;
  }
  
  if (percentage >= 100) {
    return true;
  }
  
  const userHash = hashStringToPercentage(userId);
  return userHash < percentage;
}
