import { Event, Funnel } from '../types';
import { v4 as uuidv4 } from 'uuid';

class MemoryStore {
  private events: Event[] = [];
  private funnels: Funnel[] = [];
  private eventTypes: Set<string> = new Set();

  addEvent(event: Omit<Event, 'id'>): Event {
    const newEvent: Event = {
      ...event,
      id: uuidv4(),
    };
    this.events.push(newEvent);
    this.eventTypes.add(event.eventName);
    return newEvent;
  }

  getEvents(filters?: {
    userId?: string;
    eventName?: string;
    startTime?: Date;
    endTime?: Date;
  }): Event[] {
    let result = [...this.events];
    
    if (filters?.userId) {
      result = result.filter(e => e.userId === filters.userId);
    }
    if (filters?.eventName) {
      result = result.filter(e => e.eventName === filters.eventName);
    }
    if (filters?.startTime) {
      result = result.filter(e => e.occurredAt >= filters.startTime!);
    }
    if (filters?.endTime) {
      result = result.filter(e => e.occurredAt <= filters.endTime!);
    }
    
    return result.sort((a, b) => a.occurredAt.getTime() - b.occurredAt.getTime());
  }

  addFunnel(funnel: Omit<Funnel, 'id' | 'createdAt' | 'updatedAt' | 'version'>): Funnel {
    const now = new Date();
    const newFunnel: Funnel = {
      ...funnel,
      id: uuidv4(),
      createdAt: now,
      updatedAt: now,
      version: 1,
    };
    this.funnels.push(newFunnel);
    return newFunnel;
  }

  getFunnel(id: string): Funnel | undefined {
    return this.funnels.find(f => f.id === id);
  }

  getFunnels(): Funnel[] {
    return [...this.funnels];
  }

  updateFunnel(
    id: string,
    updates: Partial<Pick<Funnel, 'name' | 'steps'>>
  ): Funnel | undefined {
    const index = this.funnels.findIndex(f => f.id === id);
    if (index === -1) return undefined;
    
    const updated: Funnel = {
      ...this.funnels[index],
      ...updates,
      updatedAt: new Date(),
      version: this.funnels[index].version + 1,
    };
    this.funnels[index] = updated;
    return updated;
  }

  deleteFunnel(id: string): boolean {
    const index = this.funnels.findIndex(f => f.id === id);
    if (index === -1) return false;
    this.funnels.splice(index, 1);
    return true;
  }

  getEventTypes(): string[] {
    return Array.from(this.eventTypes);
  }

  addEventType(eventName: string): void {
    this.eventTypes.add(eventName);
  }

  deleteEventType(eventName: string): boolean {
    return this.eventTypes.delete(eventName);
  }

  isEventTypeUsedInFunnel(eventName: string): boolean {
    return this.funnels.some(f => 
      f.steps.some(s => s.eventName === eventName)
    );
  }
}

export const memoryStore = new MemoryStore();
