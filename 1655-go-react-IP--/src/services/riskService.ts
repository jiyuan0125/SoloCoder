import { RiskAssessment, RiskFactor } from '../types';
import { rateLimitService } from './rateLimitService';
import { ruleService } from './ruleService';

const SCORE_THRESHOLD = 80;
const NEAR_LIMIT_THRESHOLD = 0.8;

const SCORE = {
  RATE_NORMAL: 0,
  RATE_NEAR_LIMIT: 30,
  RATE_TRIGGERED: 100,
  BLACKLIST_IP: 50,
  ADMIN_PATH: 20,
  CONFIG_PATH: 20
};

export const riskService = {
  assessRisk(ip: string, path: string): RiskAssessment {
    const factors: RiskFactor[] = [];
    let totalScore = 0;

    const maxRequests = rateLimitService.getMaxRequests();
    const currentCount = rateLimitService.getCurrentCount(ip);
    const usageRatio = currentCount / maxRequests;

    if (usageRatio >= 1) {
      totalScore += SCORE.RATE_TRIGGERED;
      factors.push({
        type: 'rate_limit',
        description: '请求频率已触发限制',
        score: SCORE.RATE_TRIGGERED
      });
    } else if (usageRatio >= NEAR_LIMIT_THRESHOLD) {
      totalScore += SCORE.RATE_NEAR_LIMIT;
      factors.push({
        type: 'rate_near_limit',
        description: `请求频率接近限制 (${currentCount}/${maxRequests})`,
        score: SCORE.RATE_NEAR_LIMIT
      });
    } else {
      factors.push({
        type: 'rate_normal',
        description: '请求频率正常',
        score: SCORE.RATE_NORMAL
      });
    }

    const ipCheck = ruleService.checkIP(ip);
    if (ipCheck.ruleType === 'blacklist') {
      totalScore += SCORE.BLACKLIST_IP;
      factors.push({
        type: 'blacklist_ip',
        description: `IP ${ip} 在黑名单中`,
        score: SCORE.BLACKLIST_IP
      });
    }

    const lowerPath = path.toLowerCase();
    if (lowerPath.includes('admin')) {
      totalScore += SCORE.ADMIN_PATH;
      factors.push({
        type: 'admin_path',
        description: '请求路径包含 admin 关键字',
        score: SCORE.ADMIN_PATH
      });
    }

    if (lowerPath.includes('config')) {
      totalScore += SCORE.CONFIG_PATH;
      factors.push({
        type: 'config_path',
        description: '请求路径包含 config 关键字',
        score: SCORE.CONFIG_PATH
      });
    }

    return {
      score: totalScore,
      factors,
      allowed: totalScore <= SCORE_THRESHOLD
    };
  },

  getThreshold(): number {
    return SCORE_THRESHOLD;
  }
};
