export type ProfilingType = 'cpu' | 'memory';

export interface ProfilingRecord {
  id: string;
  appId: string;
  type: ProfilingType;
  startTime: number;
  duration: number;
  status: 'running' | 'completed';
}

export interface ProfilingSample {
  id: string;
  recordId: string;
  stackId: string;
  sampleCount: number;
  selfTime: number;
}

export interface StackFrame {
  id: string;
  stackId: string;
  depth: number;
  functionName: string;
  fileName: string;
  lineNumber: number;
}

export interface FlamegraphNode {
  functionName: string;
  fileName: string;
  lineNumber: number;
  totalTime: number;
  selfTime: number;
  isHotspot: boolean;
  children: FlamegraphNode[];
}

export interface CompareItem {
  functionName: string;
  fileName: string;
  lineNumber: number;
  selfTimeBefore: number;
  selfTimeAfter: number;
  selfTimeChange: number;
  selfTimeChangePercent: number;
}

export interface StartRequest {
  type: ProfilingType;
}

export interface StopRequest {
  samples: StopSample[];
}

export interface StopSample {
  stack: StackFrameItem[];
  sampleCount: number;
  selfTime: number;
}

export interface StackFrameItem {
  functionName: string;
  fileName: string;
  lineNumber: number;
}

export interface CompareRequest {
  recordId1: string;
  recordId2: string;
}
