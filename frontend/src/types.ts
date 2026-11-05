export interface Viewport {
  x: number;
  y: number;
  zoom: number;
}

export interface Card {
  id: string;
  kind: 'folder' | 'video';
  path: string;
  label: string;
  dateLabel?: string;
  x: number;
  y: number;
  width: number;
  height: number;
  addedAt: number;
}

export interface Note {
  id: string;
  text: string;
  color: string;
  x: number;
  y: number;
  width: number;
  height: number;
  createdAt: number;
}

export interface Arrow {
  id: string;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  color: string;
  createdAt: number;
}

export interface Board {
  id: string;
  name: string;
  cards: Card[];
  notes: Note[];
  arrows: Arrow[];
  viewport: Viewport;
}

export interface ModelInfo {
  file: string;
  name: string;
  description: string;
  url: string;
  size: number;
  downloaded: boolean;
  active: boolean;
  path: string;
  downloading: boolean;
  progress: number;
  error: string;
}

export interface ModelDownload {
  file: string;
  status: 'downloading' | 'done' | 'error' | 'canceled';
  progress: number;
  bytes: number;
  total: number;
  error: string;
}

export interface EngineStatus {
  ready: boolean;
  whisperPath: string;
  ffmpegPath: string;
  ffprobePath: string;
  modelPath: string;
  modelName: string;
  vadPath: string;
  gpu: string;
  missing: string[];
  models: ModelInfo[];
}

export interface AppState {
  version: number;
  workspace: string;
  boards: Board[];
  activeBoardId: string;
  language: string;
  engine: EngineStatus;
  mediaBase: string;
}

export interface DirEntry {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  modTime: number;
  kind: 'folder' | 'video' | 'other';
  dateLabel?: string;
}

export type JobStatus =
  | 'queued'
  | 'extracting'
  | 'transcribing'
  | 'saving'
  | 'done'
  | 'error'
  | 'canceled';

export interface TranscriptMeta {
  status: JobStatus | 'none';
  progress: number;
  stage: string;
  segments: number;
  language: string;
  createdAt: number;
  error: string;
}

export interface MediaInfo {
  path: string;
  name: string;
  size: number;
  modTime: number;
  duration: number;
  kind: string;
  transcript: TranscriptMeta;
}

export interface Word {
  start: number;
  end: number;
  text: string;
  confidence?: number;
}

export interface Segment {
  start: number;
  end: number;
  text: string;
  confidence?: number;
  words?: Word[];
}

export interface Transcript {
  version: number;
  mediaPath: string;
  mediaSize: number;
  mediaModTime: number;
  duration: number;
  language: string;
  model: string;
  createdAt: number;
  segments: Segment[];
}

export interface Job {
  id: string;
  mediaPath: string;
  cardId: string;
  status: JobStatus;
  stage: string;
  progress: number;
  error: string;
  createdAt: number;
  startedAt: number;
  finishedAt: number;
}

export interface ClipInfo {
  path: string;
  duration: number;
}

export type View = 'board' | 'review';
