export interface ParsedVersion {
  major: number;
  minor: number;
  patch: number;
  original: string;
  normalized: string;
}

export function parseVersion(version: string): ParsedVersion | null {
  const parts = version.split('.');
  if (parts.length !== 3) {
    return null;
  }

  const [majorStr, minorStr, patchStr] = parts;

  if (!/^0+$/.test(majorStr) && !/^[1-9]\d*$/.test(majorStr) && !/^0[1-9]\d*$/.test(majorStr)) {
    if (!/^\d+$/.test(majorStr)) return null;
  }
  if (!/^0+$/.test(minorStr) && !/^[1-9]\d*$/.test(minorStr) && !/^0[1-9]\d*$/.test(minorStr)) {
    if (!/^\d+$/.test(minorStr)) return null;
  }
  if (!/^0+$/.test(patchStr) && !/^[1-9]\d*$/.test(patchStr) && !/^0[1-9]\d*$/.test(patchStr)) {
    if (!/^\d+$/.test(patchStr)) return null;
  }

  if (!/^\d+$/.test(majorStr) || !/^\d+$/.test(minorStr) || !/^\d+$/.test(patchStr)) {
    return null;
  }

  const major = parseInt(majorStr, 10);
  const minor = parseInt(minorStr, 10);
  const patch = parseInt(patchStr, 10);

  const MAX_VALUE = 65535;
  if (major > MAX_VALUE || minor > MAX_VALUE || patch > MAX_VALUE) {
    return null;
  }

  return {
    major,
    minor,
    patch,
    original: version,
    normalized: `${major}.${minor}.${patch}`
  };
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

export function isValidVersion(version: string): boolean {
  return parseVersion(version) !== null;
}

export function normalizeVersion(version: string): string | null {
  const parsed = parseVersion(version);
  return parsed ? parsed.normalized : null;
}
