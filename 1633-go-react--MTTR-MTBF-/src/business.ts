import { IncidentSeverity, ImpactScope, IncidentStatus } from './types';

const THIRTY_MINUTES = 30 * 60 * 1000;
const ONE_HOUR = 60 * 60 * 1000;

export function calculateSeverity(
  impactScope: ImpactScope,
  startTime: Date,
  recoveredTime: Date | null
): IncidentSeverity {
  const duration = recoveredTime
    ? recoveredTime.getTime() - startTime.getTime()
    : 0;

  if (impactScope === 'core') {
    if (duration > THIRTY_MINUTES) {
      return 'P1';
    }
    return 'P2';
  } else {
    if (duration > ONE_HOUR) {
      return 'P2';
    }
    if (duration > THIRTY_MINUTES && duration <= ONE_HOUR) {
      return 'P3';
    }
    return 'P4';
  }
}

const VALID_TRANSITIONS: Record<IncidentStatus, IncidentStatus[]> = {
  discovered: ['processing'],
  processing: ['recovered'],
  recovered: ['reviewed', 'processing'],
  reviewed: [],
};

export function isValidStatusTransition(
  from: IncidentStatus,
  to: IncidentStatus
): boolean {
  const allowed = VALID_TRANSITIONS[from];
  return allowed.includes(to);
}

export function canBeReopened(status: IncidentStatus): boolean {
  return status !== 'reviewed';
}
