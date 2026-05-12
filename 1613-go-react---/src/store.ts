import { VersionRecord, ReleaseStatus, GrayScaleConfig } from './types';
import { parseVersion, compareVersionStrings, isValidStatusTransition, truncateChangelog, isValidBuildNumber, isValidVersion } from './utils';

class VersionStore {
  private versions: Map<string, VersionRecord> = new Map();
  private versionList: VersionRecord[] = [];
  private grayScaleConfigs: Map<string, GrayScaleConfig> = new Map();
  private latestPublishedVersion: VersionRecord | null = null;
  private currentMinCompatibleVersion: string | null = null;

  getVersions(): VersionRecord[] {
    return [...this.versionList];
  }

  getVersion(version: string): VersionRecord | undefined {
    return this.versions.get(version);
  }

  hasVersion(version: string): boolean {
    return this.versions.has(version);
  }

  getLatestPublishedVersion(): VersionRecord | null {
    return this.latestPublishedVersion;
  }

  getMinCompatibleVersion(): string | null {
    return this.currentMinCompatibleVersion;
  }

  getMaxBuildNumber(): number {
    if (this.versionList.length === 0) {
      return 0;
    }
    return Math.max(...this.versionList.map(v => v.buildNumber));
  }

  createVersion(
    version: string,
    buildNumber: number,
    changelog: string
  ): { error?: string; statusCode?: number; record?: VersionRecord } {
    if (!isValidVersion(version)) {
      return { error: '版本号格式错误，必须是 MAJOR.MINOR.PATCH 格式，每段为 0-65535 的整数', statusCode: 400 };
    }

    if (this.versions.has(version)) {
      return { 
        error: '版本号已存在', 
        statusCode: 409, 
        record: this.versions.get(version) 
      };
    }

    if (!isValidBuildNumber(buildNumber)) {
      return { error: '构建号必须是正整数', statusCode: 400 };
    }

    const maxBuild = this.getMaxBuildNumber();
    if (buildNumber <= maxBuild) {
      return { 
        error: `构建号必须大于已有的最大构建号 ${maxBuild}`, 
        statusCode: 400 
      };
    }

    const parsed = parseVersion(version)!;
    const now = new Date();
    
    const record: VersionRecord = {
      version,
      major: parsed.major,
      minor: parsed.minor,
      patch: parsed.patch,
      buildNumber,
      changelog: truncateChangelog(changelog),
      status: ReleaseStatus.UNPUBLISHED,
      createdAt: now,
      updatedAt: now
    };

    this.versions.set(version, record);
    this.versionList.push(record);
    this.sortVersionList();

    return { record };
  }

  publishVersion(
    version: string,
    minCompatibleVersion?: string
  ): { error?: string; statusCode?: number; record?: VersionRecord } {
    const record = this.versions.get(version);
    if (!record) {
      return { error: '版本不存在', statusCode: 404 };
    }

    const transition = isValidStatusTransition(record.status, ReleaseStatus.PUBLISHED);
    if (!transition.valid) {
      return { error: transition.message, statusCode: 400 };
    }

    if (minCompatibleVersion !== undefined) {
      if (minCompatibleVersion && !isValidVersion(minCompatibleVersion)) {
        return { error: '最低兼容版本号格式错误', statusCode: 400 };
      }
      
      if (minCompatibleVersion && !this.versions.has(minCompatibleVersion)) {
        return { error: '最低兼容版本不存在', statusCode: 400 };
      }

      if (minCompatibleVersion && compareVersionStrings(minCompatibleVersion, version) > 0) {
        return { error: '最低兼容版本不能高于当前发布版本', statusCode: 400 };
      }

      record.minCompatibleVersion = minCompatibleVersion || undefined;
    }

    record.status = ReleaseStatus.PUBLISHED;
    record.updatedAt = new Date();

    this.updateLatestPublishedVersion();
    
    if (record.minCompatibleVersion) {
      this.currentMinCompatibleVersion = record.minCompatibleVersion;
    }

    return { record };
  }

  withdrawVersion(
    version: string
  ): { 
    error?: string; 
    statusCode?: number; 
    record?: VersionRecord;
    rolledBackMinCompatibleVersion?: string;
  } {
    const record = this.versions.get(version);
    if (!record) {
      return { error: '版本不存在', statusCode: 404 };
    }

    const transition = isValidStatusTransition(record.status, ReleaseStatus.WITHDRAWN);
    if (!transition.valid) {
      return { error: transition.message, statusCode: 400 };
    }

    record.status = ReleaseStatus.WITHDRAWN;
    record.updatedAt = new Date();

    this.resetGrayScaleForVersion(version);

    let rolledBackVersion: string | undefined;
    if (this.latestPublishedVersion && this.latestPublishedVersion.version === version) {
      this.updateLatestPublishedVersion();
      if (this.latestPublishedVersion) {
        this.currentMinCompatibleVersion = this.latestPublishedVersion.version;
        rolledBackVersion = this.latestPublishedVersion.version;
      } else {
        this.currentMinCompatibleVersion = null;
      }
    }

    return { 
      record, 
      rolledBackMinCompatibleVersion: rolledBackVersion
    };
  }

  private updateLatestPublishedVersion(): void {
    const publishedVersions = this.versionList.filter(
      v => v.status === ReleaseStatus.PUBLISHED
    );

    if (publishedVersions.length === 0) {
      this.latestPublishedVersion = null;
      return;
    }

    this.latestPublishedVersion = publishedVersions[publishedVersions.length - 1];
  }

  private sortVersionList(): void {
    this.versionList.sort((a, b) => {
      if (a.major !== b.major) return a.major - b.major;
      if (a.minor !== b.minor) return a.minor - b.minor;
      if (a.patch !== b.patch) return a.patch - b.patch;
      return a.buildNumber - b.buildNumber;
    });
  }

  getGrayScaleConfig(version: string): GrayScaleConfig {
    if (!this.grayScaleConfigs.has(version)) {
      this.grayScaleConfigs.set(version, {
        enabled: false,
        percentage: 0,
        whitelist: new Set()
      });
    }
    return this.grayScaleConfigs.get(version)!;
  }

  setGrayScalePercentage(
    version: string,
    percentage: number
  ): { error?: string; statusCode?: number; config?: GrayScaleConfig } {
    if (!this.versions.has(version)) {
      return { error: '版本不存在', statusCode: 404 };
    }

    if (!Number.isInteger(percentage) || percentage < 0 || percentage > 100) {
      return { error: '灰度百分比必须是 0-100 之间的整数', statusCode: 400 };
    }

    const config = this.getGrayScaleConfig(version);
    config.percentage = percentage;
    config.enabled = percentage > 0 || config.whitelist.size > 0;

    return { config };
  }

  addToWhitelist(
    version: string,
    users: string[]
  ): { 
    error?: string; 
    statusCode?: number; 
    added?: string[];
    duplicate?: string[];
  } {
    if (!this.versions.has(version)) {
      return { error: '版本不存在', statusCode: 404 };
    }

    const config = this.getGrayScaleConfig(version);
    const added: string[] = [];
    const duplicate: string[] = [];

    for (const user of users) {
      if (config.whitelist.has(user)) {
        duplicate.push(user);
      } else {
        config.whitelist.add(user);
        added.push(user);
      }
    }

    if (duplicate.length > 0) {
      return { 
        error: '部分用户已在白名单中', 
        statusCode: 409,
        duplicate,
        added: added.length > 0 ? added : undefined
      };
    }

    config.enabled = config.percentage > 0 || config.whitelist.size > 0;
    return { added };
  }

  private resetGrayScaleForVersion(version: string): void {
    const config = this.grayScaleConfigs.get(version);
    if (config) {
      config.percentage = 0;
    }
  }
}

export const store = new VersionStore();
