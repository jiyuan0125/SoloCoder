const crypto = require('crypto');

const PBKDF2_ITERATIONS = 100000;
const KEY_LENGTH = 32; // 256 bits
const IV_LENGTH = 16; // 128 bits
const SALT_LENGTH = 32;
const ALGORITHM = 'aes-256-gcm';
const AUTH_TAG_LENGTH = 16;

function generateSalt(length = SALT_LENGTH) {
  return crypto.randomBytes(length);
}

function deriveKey(password, salt, iterations = PBKDF2_ITERATIONS) {
  return new Promise((resolve, reject) => {
    crypto.pbkdf2(
      password,
      salt,
      iterations,
      KEY_LENGTH,
      'sha256',
      (err, derivedKey) => {
        if (err) reject(err);
        else resolve(derivedKey);
      }
    );
  });
}

async function encrypt(data, key) {
  const iv = crypto.randomBytes(IV_LENGTH);
  const cipher = crypto.createCipheriv(ALGORITHM, key, iv);
  
  const jsonString = JSON.stringify(data);
  let encrypted = cipher.update(jsonString, 'utf8', 'hex');
  encrypted += cipher.final('hex');
  
  const authTag = cipher.getAuthTag();
  
  return {
    iv: iv.toString('hex'),
    encryptedData: encrypted,
    authTag: authTag.toString('hex')
  };
}

async function decrypt(encryptedObject, key) {
  const iv = Buffer.from(encryptedObject.iv, 'hex');
  const encryptedData = encryptedObject.encryptedData;
  const authTag = Buffer.from(encryptedObject.authTag, 'hex');
  
  const decipher = crypto.createDecipheriv(ALGORITHM, key, iv);
  decipher.setAuthTag(authTag);
  
  try {
    let decrypted = decipher.update(encryptedData, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    return JSON.parse(decrypted);
  } catch (error) {
    throw new Error('解密失败：密钥无效或数据已损坏');
  }
}

async function hashPasswordForVerification(password, salt) {
  const key = await deriveKey(password, salt);
  return key.toString('hex');
}

function generatePassword(options = {}) {
  const {
    length = 16,
    useUppercase = true,
    useLowercase = true,
    useNumbers = true,
    useSpecial = true
  } = options;
  
  if (length < 8 || length > 64) {
    throw new Error('密码长度必须在 8-64 位之间');
  }
  
  let charset = '';
  if (useUppercase) charset += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
  if (useLowercase) charset += 'abcdefghijklmnopqrstuvwxyz';
  if (useNumbers) charset += '0123456789';
  if (useSpecial) charset += '!@#$%^&*()_+-=[]{}|;:,.<>?';
  
  if (charset === '') {
    throw new Error('必须至少选择一种字符类型');
  }
  
  let password = '';
  const charsetLength = charset.length;
  const randomBytes = crypto.randomBytes(length);
  
  for (let i = 0; i < length; i++) {
    const randomIndex = randomBytes[i] % charsetLength;
    password += charset[randomIndex];
  }
  
  return password;
}

module.exports = {
  generateSalt,
  deriveKey,
  encrypt,
  decrypt,
  hashPasswordForVerification,
  generatePassword,
  PBKDF2_ITERATIONS,
  SALT_LENGTH
};
