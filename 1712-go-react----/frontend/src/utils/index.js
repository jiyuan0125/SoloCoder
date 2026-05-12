export const getBMICategory = (bmi) => {
  if (bmi < 18.5) return { category: '偏瘦', color: '#3498db' };
  if (bmi < 24) return { category: '正常', color: '#2ecc71' };
  if (bmi < 28) return { category: '超重', color: '#f39c12' };
  return { category: '肥胖', color: '#e74c3c' };
};

export const isHypertension = (systolic, diastolic) => {
  return systolic >= 140 || diastolic >= 90;
};

export const formatDate = (dateStr) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
};

export const calculateAge = (birthDate) => {
  const today = new Date();
  const birth = new Date(birthDate);
  let age = today.getFullYear() - birth.getFullYear();
  const monthDiff = today.getMonth() - birth.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    age--;
  }
  return age;
};

export const diseaseNames = {
  hypertension: '高血压',
  diabetes: '糖尿病',
  coronary: '冠心病',
  stroke: '脑卒中',
  copd: '慢性阻塞性肺疾病',
};

export const diseaseFollowupIntervals = {
  hypertension: 3,
  diabetes: 1,
  coronary: 6,
  stroke: 3,
  copd: 3,
};
