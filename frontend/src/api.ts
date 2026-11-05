import type {
  AppState,
  Card,
  ClipInfo,
  DirEntry,
  Job,
  MediaInfo,
  ModelDownload,
  Transcript,
  Viewport,
} from './types';

declare global {
  interface Window {
    go: {
      main: {
        App: Record<string, (...args: any[]) => Promise<any>>;
      };
    };
    runtime: any;
  }
}

function invoke<T>(method: string, ...args: any[]): Promise<T> {
  const fn = window.go?.main?.App?.[method];
  if (!fn) {
    return Promise.reject(new Error('Wails todavía no está disponible'));
  }
  return fn(...args) as Promise<T>;
}

export const api = {
  getState: () => invoke<AppState>('GetState'),
  chooseWorkspace: () => invoke<AppState>('ChooseWorkspace'),
  createBoard: (name: string) => invoke<AppState>('CreateBoard', name),
  renameBoard: (id: string, name: string) => invoke<AppState>('RenameBoard', id, name),
  deleteBoard: (id: string) => invoke<AppState>('DeleteBoard', id),
  setActiveBoard: (id: string) => invoke<AppState>('SetActiveBoard', id),
  saveViewport: (boardId: string, viewport: Viewport) =>
    invoke<void>('SaveViewport', boardId, viewport),
  addCard: (boardId: string, path: string, x: number, y: number) =>
    invoke<AppState>('AddCard', boardId, path, x, y),
  createFolderCard: (boardId: string, parent: string, name: string, x: number, y: number) =>
    invoke<AppState>('CreateFolderCard', boardId, parent, name, x, y),
  addNote: (boardId: string, text: string, x: number, y: number, width: number, height: number) =>
    invoke<AppState>('AddNote', boardId, text, x, y, width, height),
  updateNote: (boardId: string, noteId: string, text: string, color: string) =>
    invoke<AppState>('UpdateNote', boardId, noteId, text, color),
  moveNote: (boardId: string, noteId: string, x: number, y: number) =>
    invoke<void>('MoveNote', boardId, noteId, x, y),
  removeNote: (boardId: string, noteId: string) =>
    invoke<AppState>('RemoveNote', boardId, noteId),
  addArrow: (boardId: string, x1: number, y1: number, x2: number, y2: number) =>
    invoke<AppState>('AddArrow', boardId, x1, y1, x2, y2),
  removeArrow: (boardId: string, arrowId: string) =>
    invoke<AppState>('RemoveArrow', boardId, arrowId),
  moveCard: (boardId: string, cardId: string, x: number, y: number) =>
    invoke<void>('MoveCard', boardId, cardId, x, y),
  removeCard: (boardId: string, cardId: string) =>
    invoke<AppState>('RemoveCard', boardId, cardId),
  listDir: (path: string) => invoke<DirEntry[]>('ListDir', path),
  getMediaInfo: (path: string) => invoke<MediaInfo>('GetMediaInfo', path),
  getTranscript: (path: string) => invoke<Transcript>('GetTranscript', path),
  startTranscription: (path: string, cardId: string) =>
    invoke<Job>('StartTranscription', path, cardId),
  cancelJob: (id: string) => invoke<void>('CancelJob', id),
  getJobs: () => invoke<Job[]>('GetJobs'),
  selectModel: (file: string) => invoke<AppState>('SelectModel', file),
  setLanguage: (language: string) => invoke<AppState>('SetLanguage', language),
  downloadModel: (file: string) => invoke<ModelDownload>('DownloadModel', file),
  cancelModelDownload: (file: string) => invoke<void>('CancelModelDownload', file),
  openModelsFolder: () => invoke<void>('OpenModelsFolder'),
  exportTranscript: (path: string, format: 'md' | 'srt') =>
    invoke<string>('ExportTranscript', path, format),
  copyText: (text: string) => invoke<void>('CopyText', text),
  reveal: (path: string) => invoke<void>('Reveal', path),
  openExternal: (path: string) => invoke<void>('OpenExternal', path),
  captureFrame: (path: string, second: number) =>
    invoke<string>('CaptureFrame', path, second),
  probeMedia: (path: string) => invoke<ClipInfo>('ProbeMedia', path),
};

export function mediaUrl(base: string, path: string): string {
  return `${base}/file?path=${encodeURIComponent(path)}`;
}

export function onWailsEvent(name: string, callback: (...args: any[]) => void): () => void {
  const runtime = window.runtime;
  if (!runtime?.EventsOn) return () => undefined;
  runtime.EventsOn(name, callback);
  return () => {
    try {
      runtime.EventsOff?.(name);
    } catch {
      // The runtime may already be gone during shutdown.
    }
  };
}

export function thumbUrl(base: string, path: string): string {
  return `${base}/thumb?path=${encodeURIComponent(path)}`;
}

export function friendlyError(err: unknown): string {
  if (typeof err === 'string') return err;
  if (err instanceof Error) return err.message;
  return String(err ?? 'Error desconocido');
}

export function fileName(path: string): string {
  const parts = path.split(/[\\/]/);
  return parts[parts.length - 1] || path;
}

export function parentName(path: string): string {
  const parts = path.split(/[\\/]/).filter(Boolean);
  return parts.length > 1 ? parts[parts.length - 2] : '';
}

export function formatBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes <= 0) return '—';
  const units = ['B', 'KB', 'MB', 'GB'];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  return `${value >= 100 ? Math.round(value) : value.toFixed(1)} ${units[unit]}`;
}

export function formatDuration(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return '—';
  const total = Math.round(seconds);
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  if (h > 0) return `${h}h ${String(m).padStart(2, '0')}m`;
  return `${m}m ${String(s).padStart(2, '0')}s`;
}

export function formatTimestamp(seconds: number): string {
  const total = Math.max(0, Math.floor(seconds));
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

export function formatTimestampPrecise(seconds: number): string {
  const value = Math.max(0, seconds);
  const total = Math.floor(value);
  const ms = Math.round((value - total) * 1000);
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}.${String(ms).padStart(3, '0')}`;
}

export function compactPath(path: string, max = 44): string {
  if (path.length <= max) return path;
  const parts = path.split(/[\\/]/);
  const tail = parts.slice(-2).join('\\');
  return `…\\${tail}`;
}

// Finds the first free slot in a 3-column grid, avoiding existing cards.
export function nextCardPosition(cards: Card[]): { x: number; y: number } {
  for (let i = 0; i < 120; i++) {
    const x = 80 + (i % 3) * 352;
    const y = 90 + Math.floor(i / 3) * 300;
    const occupied = cards.some(
      (card) => Math.abs(card.x - x) < 320 && Math.abs(card.y - y) < 275,
    );
    if (!occupied) return { x, y };
  }
  return { x: 80 + cards.length * 18, y: 90 + cards.length * 18 };
}
