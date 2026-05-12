export class ReviewQueue {
  private available: boolean = true;
  private queue: string[] = [];

  setAvailable(available: boolean): void {
    this.available = available;
  }

  isAvailable(): boolean {
    return this.available;
  }

  push(contentId: string): boolean {
    if (!this.available) {
      return false;
    }
    this.queue.push(contentId);
    return true;
  }

  pop(): string | undefined {
    return this.queue.shift();
  }

  getAll(): string[] {
    return [...this.queue];
  }
}

export const reviewQueue = new ReviewQueue();
