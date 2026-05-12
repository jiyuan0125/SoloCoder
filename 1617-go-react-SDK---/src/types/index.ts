export interface ReleaseNotes {
  features: string[];
  bugFixes: string[];
  breakingChanges: string[];
}

export interface SDKVersion {
  version: string;
  major: number;
  minor: number;
  patch: number;
  releaseNotes: ReleaseNotes;
  integrationGuide: string;
  createdAt: Date;
  downloadCount: number;
}

export interface SDKLanguage {
  lang: string;
  versions: SDKVersion[];
}

export interface SDK {
  id: string;
  name: string;
  description?: string;
  deprecated: boolean;
  createdAt: Date;
  updatedAt: Date;
  languages: SDKLanguage[];
}

export interface CreateSDKRequest {
  id: string;
  name: string;
  description?: string;
}

export interface UpdateSDKRequest {
  name?: string;
  description?: string;
  deprecated?: boolean;
}

export interface UploadVersionRequest {
  version: string;
  releaseNotes: ReleaseNotes;
  integrationGuide?: string;
}
