export enum ReleaseStatus {
  UNPUBLISHED = 'unpublished',
  PUBLISHED = 'published',
  WITHDRAWN = 'withdrawn'
}

export interface VersionRecord {
  version: string;
  major: number;
  minor: number;
  patch: number;
  buildNumber: number;
  changelog: string;
  status: ReleaseStatus;
  minCompatibleVersion?: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface GrayScaleConfig {
  enabled: boolean;
  percentage: number;
  whitelist: Set<string>;
}

export interface CheckUpdateResponse {
  forceUpdate: boolean;
  latestVersion: string;
  currentVersion: string;
  updateAvailable: boolean;
  changelog?: string;
}

export interface CreateVersionRequest {
  version: string;
  buildNumber: number;
  changelog: string;
}

export interface PublishVersionRequest {
  minCompatibleVersion?: string;
}

export interface AddWhitelistRequest {
  users: string[];
}

export interface GrayScalePercentageRequest {
  percentage: number;
}

export function getStatusTransitions(): Map<ReleaseStatus, Set<ReleaseStatus>> {
  const transitions = new Map<ReleaseStatus, Set<ReleaseStatus>>();
  transitions.set(ReleaseStatus.UNPUBLISHED, new Set([ReleaseStatus.PUBLISHED]));
  transitions.set(ReleaseStatus.PUBLISHED, new Set([ReleaseStatus.WITHDRAWN]));
  transitions.set(ReleaseStatus.WITHDRAWN, new Set([ReleaseStatus.PUBLISHED]));
  return transitions;
}

export const STATUS_TRANSITIONS = getStatusTransitions();

export function getStatusTransitionMessages(): Map<ReleaseStatus, string> {
  const messages = new Map<ReleaseStatus, string>();
  messages.set(ReleaseStatus.UNPUBLISHED, '未发布状态只能流转到已发布');
  messages.set(ReleaseStatus.PUBLISHED, '已发布状态只能流转到已撤回');
  messages.set(ReleaseStatus.WITHDRAWN, '已撤回状态只能流转到已发布');
  return messages;
}

export const STATUS_TRANSITION_MESSAGES = getStatusTransitionMessages();
