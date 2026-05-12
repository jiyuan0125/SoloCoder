import { Router, Request, Response } from 'express';
import { CreateSDKRequest, UpdateSDKRequest, UploadVersionRequest } from '../types';
import * as store from '../store/store';
import { parseVersion, compareVersions } from '../utils/version';

const router = Router();

function getResponseVersion(version: any): any {
  return {
    version: version.version,
    releaseNotes: version.releaseNotes,
    integrationGuide: version.integrationGuide,
    createdAt: version.createdAt,
    downloadCount: version.downloadCount
  };
}

function getResponseLanguage(language: any): any {
  return {
    lang: language.lang,
    versions: language.versions.map(getResponseVersion)
  };
}

function getResponseSDK(sdk: any): any {
  return {
    id: sdk.id,
    name: sdk.name,
    description: sdk.description,
    deprecated: sdk.deprecated,
    createdAt: sdk.createdAt,
    updatedAt: sdk.updatedAt,
    languages: sdk.languages.map(getResponseLanguage)
  };
}

router.post('/', (req: Request, res: Response) => {
  const body: CreateSDKRequest = req.body;
  
  if (!body.id || typeof body.id !== 'string' || body.id.trim() === '') {
    return res.status(400).json({ error: 'SDK ID is required' });
  }
  if (!body.name || typeof body.name !== 'string' || body.name.trim() === '') {
    return res.status(400).json({ error: 'SDK name is required' });
  }

  const existingSDK = store.getSDKById(body.id);
  if (existingSDK) {
    return res.status(409).json({ error: 'SDK with this ID already exists' });
  }

  const sdk = store.createSDK(body.id, body.name, body.description);
  return res.status(201).json(getResponseSDK(sdk));
});

router.get('/', (req: Request, res: Response) => {
  const includeDeprecated = req.query.include_deprecated === 'true';
  const sdks = store.getAllSDKs(includeDeprecated);
  return res.json(sdks.map(getResponseSDK));
});

router.get('/:sdkId', (req: Request, res: Response) => {
  const sdkId = req.params.sdkId;
  const sdk = store.getSDKById(sdkId);
  
  if (!sdk) {
    return res.status(404).json({ error: 'SDK not found' });
  }

  if (sdk.deprecated && req.query.include_deprecated !== 'true') {
    return res.status(404).json({ error: 'SDK not found' });
  }

  return res.json(getResponseSDK(sdk));
});

router.put('/:sdkId', (req: Request, res: Response) => {
  const sdkId = req.params.sdkId;
  const updates: UpdateSDKRequest = req.body;

  const existingSDK = store.getSDKById(sdkId);
  if (!existingSDK) {
    return res.status(404).json({ error: 'SDK not found' });
  }

  const updateData: any = {};
  if (updates.name !== undefined) {
    if (typeof updates.name !== 'string' || updates.name.trim() === '') {
      return res.status(400).json({ error: 'Invalid name' });
    }
    updateData.name = updates.name;
  }
  if (updates.description !== undefined) {
    updateData.description = updates.description;
  }
  if (updates.deprecated !== undefined) {
    if (typeof updates.deprecated !== 'boolean') {
      return res.status(400).json({ error: 'Invalid deprecated value' });
    }
    updateData.deprecated = updates.deprecated;
  }

  const updatedSDK = store.updateSDK(sdkId, updateData);
  return res.json(getResponseSDK(updatedSDK!));
});

router.delete('/:sdkId', (req: Request, res: Response) => {
  const sdkId = req.params.sdkId;

  const existingSDK = store.getSDKById(sdkId);
  if (!existingSDK) {
    return res.status(404).json({ error: 'SDK not found' });
  }

  if (store.hasPublishedVersions(sdkId)) {
    return res.status(409).json({ error: 'Cannot delete SDK with published versions' });
  }

  const success = store.deleteSDK(sdkId);
  if (success) {
    return res.status(204).send();
  }
  return res.status(500).json({ error: 'Failed to delete SDK' });
});

router.get('/:sdkId/languages', (req: Request, res: Response) => {
  const sdkId = req.params.sdkId;
  const languages = store.getSDKLanguages(sdkId);
  
  if (languages === undefined) {
    return res.status(404).json({ error: 'SDK not found' });
  }

  return res.json(languages.map(getResponseLanguage));
});

router.post('/:sdkId/languages/:lang/versions', (req: Request, res: Response) => {
  const sdkId = req.params.sdkId;
  const lang = req.params.lang;
  const body: UploadVersionRequest = req.body;

  const sdk = store.getSDKById(sdkId);
  if (!sdk) {
    return res.status(404).json({ error: 'SDK not found' });
  }

  if (!body.version || typeof body.version !== 'string') {
    return res.status(400).json({ error: 'Version is required' });
  }

  if (!body.releaseNotes) {
    return res.status(400).json({ error: 'Release notes are required' });
  }

  const parsedVersion = parseVersion(body.version);
  if (!parsedVersion) {
    return res.status(400).json({ error: 'Invalid version format. Expected MAJOR.MINOR.PATCH where each part is a non-negative integer not exceeding 65535' });
  }

  let language = store.getSDKLanguage(sdkId, lang);
  if (language) {
    const existingVersion = store.getVersionByNormalized(language, parsedVersion.normalized);
    if (existingVersion) {
      return res.status(409).json({
        error: 'Version already exists',
        existingVersion: getResponseVersion(existingVersion)
      });
    }

    const latest = store.getLatestVersion(language);
    if (latest) {
      const latestParsed = {
        major: latest.major,
        minor: latest.minor,
        patch: latest.patch,
        original: latest.version,
        normalized: latest.version
      };
      if (compareVersions(parsedVersion, latestParsed) <= 0) {
        return res.status(400).json({
          error: 'New version must be greater than the latest version',
          latestVersion: latest.version
        });
      }
    }
  }

  const newVersion = store.addVersion(sdkId, lang, body, parsedVersion);
  return res.status(201).json(getResponseVersion(newVersion));
});

router.get('/:sdkId/languages/:lang/versions/:version/download', (req: Request, res: Response) => {
  const sdkId = req.params.sdkId;
  const lang = req.params.lang;
  const version = req.params.version;

  const sdk = store.getSDKById(sdkId);
  if (!sdk) {
    return res.status(404).json({ error: 'SDK not found' });
  }

  const language = store.getSDKLanguage(sdkId, lang);
  if (!language || language.versions.length === 0) {
    return res.status(404).json({ error: '该SDK该语言无已发布版本' });
  }

  const parsedVersion = parseVersion(version);
  if (!parsedVersion) {
    return res.status(400).json({ error: 'Invalid version format' });
  }

  const existingVersion = store.getVersionByNormalized(language, parsedVersion.normalized);
  if (!existingVersion) {
    return res.status(404).json({ error: 'Version not found' });
  }

  store.incrementDownloadCount(sdkId, lang, parsedVersion.normalized);

  return res.json({
    downloadUrl: `/sdks/${sdkId}/languages/${lang}/versions/${parsedVersion.normalized}/download`,
    version: parsedVersion.normalized,
    message: 'Download started'
  });
});

router.get('/:sdkId/languages/:lang/versions', (req: Request, res: Response) => {
  const sdkId = req.params.sdkId;
  const lang = req.params.lang;

  const sdk = store.getSDKById(sdkId);
  if (!sdk) {
    return res.status(404).json({ error: 'SDK not found' });
  }

  const language = store.getSDKLanguage(sdkId, lang);
  if (!language) {
    return res.json([]);
  }

  return res.json(language.versions.map(getResponseVersion));
});

export default router;
