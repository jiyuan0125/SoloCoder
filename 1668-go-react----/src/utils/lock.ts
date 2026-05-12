const locks = new Map<string, { promise: Promise<void>; resolve: () => void }>();

export async function acquireLock(key: string, timeout: number = 5000): Promise<() => void> {
  const startTime = Date.now();
  
  while (locks.has(key)) {
    const lock = locks.get(key);
    if (lock) {
      const elapsed = Date.now() - startTime;
      const remaining = timeout - elapsed;
      
      if (remaining <= 0) {
        throw new Error('原发票正在被操作请稍后重试');
      }
      
      const timeoutPromise = new Promise<void>((_, reject) => {
        setTimeout(() => reject(new Error('原发票正在被操作请稍后重试')), remaining);
      });
      
      try {
        await Promise.race([lock.promise, timeoutPromise]);
      } catch {
        throw new Error('原发票正在被操作请稍后重试');
      }
    }
  }
  
  let resolveFn: () => void = () => {};
  const promise = new Promise<void>((resolve) => {
    resolveFn = resolve;
  });
  
  locks.set(key, { promise, resolve: resolveFn });
  
  return () => {
    locks.delete(key);
    resolveFn();
  };
}
