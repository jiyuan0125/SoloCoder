import * as crypto from 'crypto';
import * as db from '../db';
import {
  ProfilingType,
  ProfilingRecord,
  FlamegraphNode,
  CompareItem,
  StopRequest,
  StackFrameItem
} from '../types';

const MAX_RECORDS_PER_APP = 100;

export class ProfilingService {
  startProfiling(appId: string, type: ProfilingType): { recordId: string } {
    const running = db.getRunningRecord(appId);
    if (running) {
      throw new Error('PROFILING_IN_PROGRESS');
    }

    const recordId = crypto.randomUUID();
    const record: ProfilingRecord = {
      id: recordId,
      appId,
      type,
      startTime: Date.now(),
      duration: 0,
      status: 'running'
    };

    db.insertRecord(record);
    this.cleanupOldRecords(appId);

    return { recordId };
  }

  stopProfiling(appId: string, request: StopRequest): { recordId: string } {
    const running = db.getRunningRecord(appId);
    if (!running) {
      throw new Error('NO_RUNNING_PROFILING');
    }

    const duration = Date.now() - running.startTime;
    db.updateRecordStatus(running.id, duration);

    for (const sample of request.samples) {
      const stackId = crypto.randomUUID();
      const sampleId = crypto.randomUUID();

      db.insertSample({
        id: sampleId,
        recordId: running.id,
        stackId,
        sampleCount: sample.sampleCount,
        selfTime: sample.selfTime
      });

      for (let depth = 0; depth < sample.stack.length; depth++) {
        db.insertStackFrame({
          id: crypto.randomUUID(),
          stackId,
          depth,
          functionName: sample.stack[depth].functionName,
          fileName: sample.stack[depth].fileName,
          lineNumber: sample.stack[depth].lineNumber
        });
      }
    }

    return { recordId: running.id };
  }

  getRecords(appId: string): ProfilingRecord[] {
    return db.getRecordsByApp(appId);
  }

  getRecord(appId: string, recordId: string): ProfilingRecord | undefined {
    const record = db.getRecordById(recordId);
    if (!record || record.appId !== appId) {
      return undefined;
    }
    return record;
  }

  getFlamegraph(appId: string, recordId: string): FlamegraphNode[] {
    const record = db.getRecordById(recordId);
    if (!record || record.appId !== appId) {
      throw new Error('RECORD_NOT_FOUND');
    }

    const sampleWithFrames = db.getSamplesWithFrames(recordId);
    
    const stacks: StackFrameItem[][] = [];
    const sampleInfo: { sampleCount: number; selfTime: number }[] = [];
    
    let currentStack: StackFrameItem[] = [];
    for (const row of sampleWithFrames) {
      if (row.depth === 0) {
        if (currentStack.length > 0) {
          stacks.push(currentStack);
        }
        currentStack = [];
      }
      currentStack.push({
        functionName: row.function_name,
        fileName: row.file_name,
        lineNumber: row.line_number
      });
      if (row.depth === 0) {
        sampleInfo.push({ sampleCount: row.sample_count, selfTime: row.self_time });
      }
    }
    if (currentStack.length > 0) {
      stacks.push(currentStack);
    }

    const root: FlamegraphNode = {
      functionName: '',
      fileName: '',
      lineNumber: 0,
      totalTime: 0,
      selfTime: 0,
      isHotspot: false,
      children: []
    };

    for (let i = 0; i < stacks.length; i++) {
      const stack = stacks[i];
      const info = sampleInfo[i];
      this.addStackToTree(root, stack.reverse(), info.selfTime);
    }

    this.calculateTimes(root);
    
    const totalTime = root.totalTime;
    this.markHotspots(root, totalTime);

    return root.children;
  }

  private addStackToTree(
    node: FlamegraphNode, stack: StackFrameItem[], selfTime: number): void {
    if (stack.length === 0) {
      node.selfTime += selfTime;
      return;
    }

    const frame = stack[0];
    let child = node.children.find(
      c => c.functionName === frame.functionName && 
           c.fileName === frame.fileName && 
           c.lineNumber === frame.lineNumber
    );

    if (!child) {
      child = {
        functionName: frame.functionName,
        fileName: frame.fileName,
        lineNumber: frame.lineNumber,
        totalTime: 0,
        selfTime: 0,
        isHotspot: false,
        children: []
      };
      node.children.push(child);
    }

    this.addStackToTree(child, stack.slice(1), selfTime);
  }

  private calculateTimes(node: FlamegraphNode): number {
    let total = node.selfTime;
    for (const child of node.children) {
      total += this.calculateTimes(child);
    }
    node.totalTime = total;
    return total;
  }

  private markHotspots(node: FlamegraphNode, totalTime: number): void {
    if (totalTime > 0) {
      node.isHotspot = (node.totalTime / totalTime) > 0.1;
    }
    for (const child of node.children) {
      this.markHotspots(child, totalTime);
    }
  }

  compare(recordId1: string, recordId2: string): CompareItem[] {
    const record1 = db.getRecordById(recordId1);
    const record2 = db.getRecordById(recordId2);

    if (!record1 || !record2) {
      throw new Error('RECORD_NOT_FOUND');
    }

    const selfTimes1 = this.getFunctionSelfTimes(recordId1);
    const selfTimes2 = this.getFunctionSelfTimes(recordId2);

    const allKeys = new Set([...selfTimes1.keys(), ...selfTimes2.keys()]);
    const items: CompareItem[] = [];

    for (const key of allKeys) {
      const time1 = selfTimes1.get(key) || 0;
      const time2 = selfTimes2.get(key) || 0;
      const change = time2 - time1;
      const changePercent = time1 > 0 ? (change / time1) : (time2 > 0 ? 1 : 0);

      items.push({
        functionName: this.getFunctionNameFromKey(key),
        fileName: this.getFileNameFromKey(key),
        lineNumber: this.getLineNumberFromKey(key),
        selfTimeBefore: time1,
        selfTimeAfter: time2,
        selfTimeChange: change,
        selfTimeChangePercent: changePercent
      });
    }

    items.sort((a, b) => Math.abs(b.selfTimeChange) - Math.abs(a.selfTimeChange));
    return items;
  }

  private getFunctionSelfTimes(recordId: string): Map<string, number> {
    const samplesWithFrames = db.getSamplesWithFrames(recordId);
    const map = new Map<string, number>();

    for (const row of samplesWithFrames) {
      if (row.depth === 0) {
        const key = `${row.function_name}|${row.file_name}|${row.line_number}`;
        map.set(key, (map.get(key) || 0) + row.self_time);
      }
    }

    return map;
  }

  private getFunctionNameFromKey(key: string): string {
    return key.split('|')[0];
  }

  private getFileNameFromKey(key: string): string {
    return key.split('|')[1];
  }

  private getLineNumberFromKey(key: string): number {
    return parseInt(key.split('|')[2], 10);
  }

  private cleanupOldRecords(appId: string): void {
    const count = db.getRecordCount(appId);
    if (count > MAX_RECORDS_PER_APP) {
      db.deleteOldestRecords(appId, count - MAX_RECORDS_PER_APP);
    }
  }
}

export const profilingService = new ProfilingService();
