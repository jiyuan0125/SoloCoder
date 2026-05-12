const locks = new Map<number, boolean>();

export function acquireLock(tenantId: number): boolean {
  if (locks.get(tenantId)) {
    return false;
  }
  locks.set(tenantId, true);
  return true;
}

export function releaseLock(tenantId: number): void {
  locks.delete(tenantId);
}
