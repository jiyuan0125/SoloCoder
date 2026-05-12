import { Scale, Answer, NormEntry } from '../types';

const FORTY_EIGHT_HOURS = 48 * 60 * 60 * 1000;
const THIRTY_DAYS = 30 * 24 * 60 * 60 * 1000;

export const isAssessmentExpired = (lastActiveTime: number): boolean => {
  return Date.now() - lastActiveTime > FORTY_EIGHT_HOURS;
};

export const isWithinRepeatPeriod = (lastCompletedTime: number): boolean => {
  return Date.now() - lastCompletedTime < THIRTY_DAYS;
};

export const calculateRawScore = (scale: Scale, answers: Answer[]): { rawScore: number; dimensionScores: { [key: string]: number } } => {
  const dimensionScores: { [key: string]: number } = {};
  let totalRawScore = 0;

  for (const question of scale.questions) {
    const answer = answers.find(a => a.questionId === question.id);
    if (!answer) continue;

    let score = answer.score;

    if (question.isReverseScored) {
      const maxScore = Math.max(...question.options.map(o => o.score));
      score = maxScore - score;
    }

    totalRawScore += score;

    if (question.dimension) {
      dimensionScores[question.dimension] = (dimensionScores[question.dimension] || 0) + score;
    }
  }

  if (scale.scoringType === 'weighted' && scale.dimensions) {
    let weightedTotal = 0;
    for (const dim of scale.dimensions) {
      const dimScore = dimensionScores[dim.name] || 0;
      weightedTotal += dimScore * dim.weight;
    }
    return { rawScore: Math.round(weightedTotal), dimensionScores };
  }

  return { rawScore: totalRawScore, dimensionScores };
};

export const getNormEntry = (normTable: NormEntry[], rawScore: number): NormEntry => {
  let bestMatch = normTable[0];
  for (const entry of normTable) {
    if (rawScore >= entry.rawScore) {
      bestMatch = entry;
    } else {
      break;
    }
  }
  return bestMatch;
};

export const generateReport = (
  scale: Scale,
  rawScore: number,
  standardScore: number,
  level: string,
  dimensionScores: { [key: string]: number }
): string => {
  const levelText: { [key: string]: string } = {
    normal: '正常',
    mild: '轻度异常',
    moderate: '中度异常',
    severe: '重度异常'
  };

  let report = `测评报告 - ${scale.name}\n\n`;
  report += `测评对象：${scale.targetGroup}\n`;
  report += `原始分数：${rawScore}\n`;
  report += `标准分数：${standardScore}\n`;
  report += `结果等级：${levelText[level] || level}\n\n`;

  if (Object.keys(dimensionScores).length > 0) {
    report += `维度得分：\n`;
    for (const [dim, score] of Object.entries(dimensionScores)) {
      report += `  - ${dim}：${score}\n`;
    }
    report += `\n`;
  }

  report += `结果说明：\n`;
  switch (level) {
    case 'normal':
      report += `您的测评结果在正常范围内，请继续保持良好的心理状态。`;
      break;
    case 'mild':
      report += `您的测评结果显示存在轻度异常，建议关注自身心理状态，必要时可进行心理咨询。`;
      break;
    case 'moderate':
      report += `您的测评结果显示存在中度异常，建议及时寻求专业心理咨询帮助。`;
      break;
    case 'severe':
      report += `您的测评结果显示存在重度异常，强烈建议立即寻求专业心理医生的帮助。`;
      break;
    default:
      report += `请结合专业人士的建议进行解读。`;
  }

  return report;
};
