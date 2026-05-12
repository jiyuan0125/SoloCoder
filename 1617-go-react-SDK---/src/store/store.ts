import { SDK, SDKVersion, SDKLanguage, UploadVersionRequest } from '../types';
import { parseVersion, compareVersions, ParsedVersion } from '../utils/version';

const sdkStore: Map<string, SDK> = new Map();

export function getAllSDKs(includeDeprecated: boolean = false): SDK[] {
  const result: SDK[] = [];
  for (const sdk of sdkStore.values()) {
    if (!sdk.deprecated || includeDeprecated) {
      result.push(sdk);
    }
  }
  return result;
}

export function getSDKById(id: string): SDK | undefined {
  return sdkStore.get(id);
}

export function createSDK(id: string, name: string, description?: string): SDK {
  const now = new Date();
  const sdk: SDK = {
    id,
    name,
    description,
    deprecated: false,
    createdAt: now,
    updatedAt: now,
    languages: []
  };
  sdkStore.set(id, sdk);
  return sdk;
}

export function updateSDK(id: string, updates: Partial<SDK>): SDK | undefined {
  const sdk = sdkStore.get(id);
  if (!sdk) return undefined;

  Object.assign(sdk, updates, { updatedAt: new Date() });
  sdkStore.set(id, sdk);
  return sdk;
}

export function deleteSDK(id: string): boolean {
  const sdk = sdkStore.get(id);
  if (!sdk) return false;

  const hasPublishedVersions = sdk.languages.some(lang => lang.versions.length > 0);
  if (hasPublishedVersions) {
    return false;
  }

  return sdkStore.delete(id);
}

export function hasPublishedVersions(sdkId: string): boolean {
  const sdk = sdkStore.get(sdkId);
  if (!sdk) return false;
  return sdk.languages.some(lang => lang.versions.length > 0);
}

export function getSDKLanguages(sdkId: string): SDKLanguage[] | undefined {
  const sdk = sdkStore.get(sdkId);
  if (!sdk) return undefined;
  return sdk.languages;
}

export function getSDKLanguage(sdkId: string, lang: string): SDKLanguage | undefined {
  const sdk = sdkStore.get(sdkId);
  if (!sdk) return undefined;
  return sdk.languages.find(l => l.lang === lang);
}

export function getLatestVersion(language: SDKLanguage): SDKVersion | null {
  if (language.versions.length === 0) return null;
  
  let latest = language.versions[0];
  for (const version of language.versions) {
    if (compareVersions(
      { major: version.major, minor: version.minor, patch: version.patch, original: '', normalized: '' },
      { major: latest.major, minor: latest.minor, patch: latest.patch, original: '', normalized: '' }
    ) > 0) {
      latest = version;
    }
  }
  return latest;
}

export function getVersionByNormalized(language: SDKLanguage, normalizedVersion: string): SDKVersion | undefined {
  return language.versions.find(v => 
    `${v.major}.${v.minor}.${v.patch}` === normalizedVersion
  );
}

export function addVersion(
  sdkId: string, 
  lang: string, 
  request: UploadVersionRequest,
  parsedVersion: ParsedVersion
): SDKVersion {
  let sdk = sdkStore.get(sdkId)!;
  let language = sdk.languages.find(l => l.lang === lang);
  
  if (!language) {
    language = {
      lang,
      versions: []
    };
    sdk.languages.push(language);
  }

  let integrationGuide = request.integrationGuide;
  if (!integrationGuide) {
    const latest = getLatestVersion(language);
    if (latest) {
      integrationGuide = latest.integrationGuide;
    }
  }

  const version: SDKVersion = {
    version: parsedVersion.normalized,
    major: parsedVersion.major,
    minor: parsedVersion.minor,
    patch: parsedVersion.patch,
    releaseNotes: request.releaseNotes,
    integrationGuide: integrationGuide || '',
    createdAt: new Date(),
    downloadCount: 0
  };

  language.versions.push(version);
  language.versions.sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
  
  sdk.updatedAt = new Date();
  sdkStore.set(sdkId, sdk);

  return version;
}

export function incrementDownloadCount(sdkId: string, lang: string, normalizedVersion: string): void {
  const sdk = sdkStore.get(sdkId);
  if (!sdk) return;

  const language = sdk.languages.find(l => l.lang === lang);
  if (!language) return;

  const version = language.versions.find(v => 
    `${v.major}.${v.minor}.${v.patch}` === normalizedVersion
  );
  if (version) {
    version.downloadCount++;
    sdk.updatedAt = new Date();
    sdkStore.set(sdkId, sdk);
  }
}
