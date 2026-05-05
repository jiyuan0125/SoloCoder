export async function asyncWrapper<T>(
  fn: () => Promise<T>,
  errorHandler: (error: Error) => T
): Promise<T> {
  try {
    return await fn();
  } catch (error: unknown) {
    if (error instanceof Error) {
      return errorHandler(error);
    }
    return errorHandler(new Error(String(error)));
  }
}
