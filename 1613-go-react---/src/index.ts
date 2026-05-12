import express, { Request, Response } from 'express';
import { store } from './store';
import { isValidVersion, isValidStatusTransition, isVersionLessThan, isUserInGrayScale } from './utils';
import { ReleaseStatus, CreateVersionRequest, PublishVersionRequest, AddWhitelistRequest, GrayScalePercentageRequest, CheckUpdateResponse } from './types';

const app = express();
const PORT = process.env.PORT || 9103;

app.use(express.json());

app.get('/versions', (req: Request, res: Response) => {
  const versions = store.getVersions();
  res.json(versions);
});

app.get('/versions/check', (req: Request, res: Response) => {
  const current = req.query.current as string;
  const userId = req.query.userId as string | undefined;

  if (!current) {
    res.status(400).json({ error: '缺少 current 参数' });
    return;
  }

  if (!isValidVersion(current)) {
    res.status(400).json({ error: '版本号格式错误' });
    return;
  }

  const latestPublished = store.getLatestPublishedVersion();
  const minCompatibleVersion = store.getMinCompatibleVersion();

  if (!latestPublished) {
    const response: CheckUpdateResponse = {
      forceUpdate: false,
      latestVersion: current,
      currentVersion: current,
      updateAvailable: false
    };
    res.json(response);
    return;
  }

  let effectiveLatest = latestPublished;
  const grayScaleConfig = store.getGrayScaleConfig(latestPublished.version);
  
  if (grayScaleConfig.enabled && userId) {
    if (!isUserInGrayScale(userId, grayScaleConfig.percentage, grayScaleConfig.whitelist)) {
      const publishedVersions = store.getVersions().filter(v => 
        v.status === ReleaseStatus.PUBLISHED && v.version !== latestPublished.version
      );
      if (publishedVersions.length > 0) {
        effectiveLatest = publishedVersions[publishedVersions.length - 1];
      }
    }
  }

  const isCurrentOlderThanLatest = isVersionLessThan(current, effectiveLatest.version);
  const forceUpdate = minCompatibleVersion !== null && isVersionLessThan(current, minCompatibleVersion);

  const response: CheckUpdateResponse = {
    forceUpdate,
    latestVersion: effectiveLatest.version,
    currentVersion: current,
    updateAvailable: isCurrentOlderThanLatest
  };

  if (isCurrentOlderThanLatest) {
    response.changelog = effectiveLatest.changelog;
  }

  res.json(response);
});

app.get('/versions/:version', (req: Request, res: Response) => {
  const version = req.params.version as string;
  const record = store.getVersion(version);
  
  if (!record) {
    res.status(404).json({ error: '版本不存在' });
    return;
  }
  
  res.json(record);
});

app.post('/versions', (req: Request, res: Response) => {
  const body = req.body as CreateVersionRequest;
  const { version, buildNumber, changelog } = body;

  if (!version || buildNumber === undefined || changelog === undefined) {
    res.status(400).json({ error: '缺少必填字段: version, buildNumber, changelog' });
    return;
  }

  const result = store.createVersion(version, buildNumber, changelog);

  if (result.error) {
    if (result.statusCode === 409 && result.record) {
      res.status(409).json({ 
        error: result.error, 
        existingVersion: result.record 
      });
      return;
    }
    res.status(result.statusCode || 400).json({ error: result.error });
    return;
  }

  res.status(201).json(result.record);
});

app.post('/versions/:version/publish', (req: Request, res: Response) => {
  const version = req.params.version as string;
  const body = req.body as PublishVersionRequest;

  const result = store.publishVersion(version, body.minCompatibleVersion);

  if (result.error) {
    res.status(result.statusCode || 400).json({ error: result.error });
    return;
  }

  res.json(result.record);
});

app.post('/versions/:version/withdraw', (req: Request, res: Response) => {
  const version = req.params.version as string;

  const result = store.withdrawVersion(version);

  if (result.error) {
    res.status(result.statusCode || 400).json({ error: result.error });
    return;
  }

  const response: any = { version: result.record };
  if (result.rolledBackMinCompatibleVersion !== undefined) {
    response.rolledBackMinCompatibleVersion = result.rolledBackMinCompatibleVersion;
  }

  res.json(response);
});

app.post('/versions/:version/grayscale/percentage', (req: Request, res: Response) => {
  const version = req.params.version as string;
  const body = req.body as GrayScalePercentageRequest;
  const { percentage } = body;

  if (percentage === undefined) {
    res.status(400).json({ error: '缺少 percentage 字段' });
    return;
  }

  const result = store.setGrayScalePercentage(version, percentage);

  if (result.error) {
    res.status(result.statusCode || 400).json({ error: result.error });
    return;
  }

  res.json({
    enabled: result.config!.enabled,
    percentage: result.config!.percentage
  });
});

app.post('/versions/:version/grayscale/whitelist', (req: Request, res: Response) => {
  const version = req.params.version as string;
  const body = req.body as AddWhitelistRequest;
  const { users } = body;

  if (!users || !Array.isArray(users)) {
    res.status(400).json({ error: 'users 必须是数组' });
    return;
  }

  const result = store.addToWhitelist(version, users);

  if (result.error) {
    const response: any = { error: result.error };
    if (result.duplicate) response.duplicate = result.duplicate;
    if (result.added) response.added = result.added;
    res.status(result.statusCode || 400).json(response);
    return;
  }

  res.json({
    added: result.added,
    totalWhitelist: store.getGrayScaleConfig(version).whitelist.size
  });
});

app.get('/versions/:version/grayscale', (req: Request, res: Response) => {
  const version = req.params.version as string;

  if (!store.hasVersion(version)) {
    res.status(404).json({ error: '版本不存在' });
    return;
  }

  const config = store.getGrayScaleConfig(version);
  res.json({
    enabled: config.enabled,
    percentage: config.percentage,
    whitelistSize: config.whitelist.size,
    whitelist: Array.from(config.whitelist)
  });
});

app.listen(PORT, () => {
  console.log(`App Version Management Service running on port ${PORT}`);
});

export default app;
