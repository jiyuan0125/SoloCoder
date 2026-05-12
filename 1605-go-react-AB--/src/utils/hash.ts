import * as crypto from 'crypto';

export const hashToPercentage = (key: string, salt: string = ''): number => {
  const hash = crypto
    .createHash('sha256')
    .update(key + salt)
    .digest('hex');
  
  const intValue = parseInt(hash.substring(0, 16), 16);
  return (intValue % 10000) / 100;
};
