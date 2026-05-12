import ipaddr from 'ipaddr.js';
import { Request } from 'express';
import { ParsedIP } from '../types';

export interface CIDRInfo {
  isValid: boolean;
  networkAddress?: string;
  prefixLength?: number;
  isCIDR: boolean;
}

export function parseIP(ipStr: string): ipaddr.IPv4 | ipaddr.IPv6 | null {
  try {
    return ipaddr.parse(ipStr);
  } catch {
    return null;
  }
}

export function validateCIDR(ipPattern: string): CIDRInfo {
  const parts = ipPattern.split('/');
  
  if (parts.length === 1) {
    const parsed = parseIP(parts[0]);
    return {
      isValid: !!parsed,
      isCIDR: false,
      networkAddress: parsed?.toString()
    };
  }

  if (parts.length !== 2) {
    return { isValid: false, isCIDR: true };
  }

  const [ipPart, prefixPart] = parts;
  const prefix = parseInt(prefixPart, 10);
  
  if (isNaN(prefix)) {
    return { isValid: false, isCIDR: true };
  }

  try {
    const ip = ipaddr.parse(ipPart);
    const kind = ip.kind();
    const maxPrefix = kind === 'ipv4' ? 32 : 128;
    
    if (prefix < 0 || prefix > maxPrefix) {
      return { isValid: false, isCIDR: true };
    }

    const cidr = ipaddr.parseCIDR(`${ipPart}/${prefix}`);
    return {
      isValid: true,
      isCIDR: true,
      networkAddress: cidr[0].toString(),
      prefixLength: prefix
    };
  } catch {
    return { isValid: false, isCIDR: true };
  }
}

export function ipMatchesPattern(ip: string, pattern: string): boolean {
  try {
    const targetIp = ipaddr.parse(ip);
    
    if (pattern.includes('/')) {
      const cidr = ipaddr.parseCIDR(pattern);
      return targetIp.match(cidr);
    }
    
    return targetIp.toString() === ipaddr.parse(pattern).toString();
  } catch {
    return false;
  }
}

export function extractRealIP(req: Request): ParsedIP {
  const xff = req.headers['x-forwarded-for'];
  if (xff) {
    const ips = Array.isArray(xff) ? xff[0] : xff;
    const firstIp = ips.split(',')[0].trim();
    if (parseIP(firstIp)) {
      return { ip: firstIp, isTrustedProxy: true };
    }
  }

  const xri = req.headers['x-real-ip'];
  if (xri) {
    const realIp = Array.isArray(xri) ? xri[0] : xri;
    if (parseIP(realIp)) {
      return { ip: realIp, isTrustedProxy: true };
    }
  }

  const remoteAddr = req.ip || req.connection?.remoteAddress || '0.0.0.0';
  return { ip: remoteAddr, isTrustedProxy: false };
}
