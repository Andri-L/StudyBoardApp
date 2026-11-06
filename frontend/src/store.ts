import { create } from 'zustand';
import { api, friendlyError } from './api';
import type { AppState, Job, ModelDownload, View, Viewport, WorkflowInfo, WorkflowJob, WorkflowVideo } from './types';

type ToastKind = 'info' | 'success' | 'error';

interface Toast {
  id: number;
  message: string;
  kind: ToastKind;
}

interface StudyBoardStore {
  state: AppState | null;
  jobs: Job[];
  loading: boolean;
  busy: boolean;
  view: View;
  reviewPath: string | null;
  selectedCardId: string | null;
  folderPath: string | null;
  jobsOpen: boolean;
  toast: Toast | null;
  mediaVersion: number;
  modelDownloads: Record<string, ModelDownload>;
  workflows: WorkflowInfo[];
  workflowVideos: WorkflowVideo[];
  workflowJobs: WorkflowJob[];
  workflowDialogOpen: boolean;
  workflowPreselect: string[];
  workflowPanelOpen: boolean;
  theme: 'light' | 'dark';

  init: () => Promise<void>;
  refresh: () => Promise<void>;
  chooseWorkspace: () => Promise<void>;
  createBoard: (name: string) => Promise<void>;
  renameBoard: (id: string, name: string) => Promise<void>;
  deleteBoard: (id: string) => Promise<void>;
  setActiveBoard: (id: string) => Promise<void>;
  saveViewport: (boardId: string, viewport: Viewport) => void;
  addCard: (path: string, x: number, y: number) => Promise<void>;
  createFolderCard: (parent: string, name: string, x: number, y: number) => Promise<boolean>;
  addNote: (x: number, y: number) => Promise<void>;
  updateNote: (noteId: string, text: string, color?: string) => Promise<void>;
  moveNote: (noteId: string, x: number, y: number) => void;
  removeNote: (noteId: string) => Promise<void>;
  addArrow: (x1: number, y1: number, x2: number, y2: number, type: 'straight' | 's' | 'square', color: string) => Promise<void>;
  updateArrow: (arrowId: string, x1: number, y1: number, x2: number, y2: number, type: 'straight' | 's' | 'square', color: string) => Promise<void>;
  removeArrow: (arrowId: string) => Promise<void>;
  setItemColor: (itemId: string, color: string) => Promise<void>;
  renamePath: (path: string, newName: string) => Promise<boolean>;
  moveCard: (boardId: string, cardId: string, x: number, y: number) => void;
  removeCard: (cardId: string) => Promise<void>;
  startTranscription: (path: string, cardId: string) => Promise<void>;
  cancelJob: (id: string) => Promise<void>;
  jobUpdated: (job: Job) => void;
  transcriptReady: (path: string) => void;
  openReview: (path: string) => void;
  closeReview: () => void;
  selectCard: (id: string | null) => void;
  openFolder: (path: string) => void;
  closeFolder: () => void;
  setJobsOpen: (open: boolean) => void;
  setAppState: (state: AppState) => void;
  modelProgress: (download: ModelDownload) => void;
  selectModel: (file: string) => Promise<void>;
  setLanguage: (language: string) => Promise<void>;
  downloadModel: (file: string) => Promise<void>;
  cancelModelDownload: (file: string) => Promise<void>;
  openModelsFolder: () => Promise<void>;
  openWorkflowDialog: (preselect?: string[]) => Promise<void>;
  closeWorkflowDialog: () => void;
  refreshWorkflowData: () => Promise<void>;
  startWorkflows: (workflowIds: string[], videoPaths: string[], language: 'es' | 'en') => Promise<void>;
  startTranslation: (resultId: string, targetLanguage: 'es' | 'en') => Promise<void>;
  workflowJobUpdated: (job: WorkflowJob) => void;
  cancelWorkflow: (id: string) => Promise<void>;
  moveWorkflowResult: (resultId: string, x: number, y: number) => void;
  removeWorkflowResult: (resultId: string) => Promise<void>;
  setWorkflowPanelOpen: (open: boolean) => void;
  toggleTheme: () => void;
  showToast: (message: string, kind?: ToastKind) => void;
  dismissToast: () => void;
}

let toastTimer: number | undefined;

function activeBoardId(state: AppState | null): string {
  return state?.activeBoardId ?? '';
}

const THEME_KEY = 'studyboard.theme';

function initialTheme(): 'light' | 'dark' {
  try {
    const stored = localStorage.getItem(THEME_KEY);
    if (stored === 'light' || stored === 'dark') return stored;
    return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  } catch {
    return 'light';
  }
}

export const useStudyBoard = create<StudyBoardStore>((set, get) => ({
  state: null,
  jobs: [],
  loading: true,
  busy: false,
  view: 'board',
  reviewPath: null,
  selectedCardId: null,
  folderPath: null,
  jobsOpen: false,
  toast: null,
  mediaVersion: 0,
  modelDownloads: {},
  workflows: [],
  workflowVideos: [],
  workflowJobs: [],
  workflowDialogOpen: false,
  workflowPreselect: [],
  workflowPanelOpen: false,
  theme: initialTheme(),

  init: async () => {
    set({ loading: true });
    try {
      const [state, jobs] = await Promise.all([api.getState(), api.getJobs()]);
      set({ state, jobs, loading: false });
    } catch (err) {
      set({ loading: false });
      get().showToast(friendlyError(err), 'error');
    }
  },

  refresh: async () => {
    try {
      const [state, jobs] = await Promise.all([api.getState(), api.getJobs()]);
      set({ state, jobs, mediaVersion: get().mediaVersion + 1 });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  chooseWorkspace: async () => {
    const current = get().state;
    const previous = current?.workspace ?? '';
    const hadCards = Boolean(current?.boards.some((b) => b.cards.length > 0));
    set({ busy: true });
    try {
      const state = await api.chooseWorkspace();
      set({ state, busy: false, view: 'board', reviewPath: null, folderPath: null });
      if (
        previous &&
        state.workspace.toLowerCase() !== previous.toLowerCase() &&
        hadCards
      ) {
        get().showToast(
          'Se creó un tablero nuevo para la carpeta elegida. Tus archivos no se modificaron.',
          'info',
        );
      }
    } catch (err) {
      set({ busy: false });
      get().showToast(friendlyError(err), 'error');
    }
  },

  createBoard: async (name) => {
    try {
      const state = await api.createBoard(name);
      set({ state, selectedCardId: null });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  renameBoard: async (id, name) => {
    try {
      const state = await api.renameBoard(id, name);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  deleteBoard: async (id) => {
    try {
      const state = await api.deleteBoard(id);
      set({ state, selectedCardId: null });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  setActiveBoard: async (id) => {
    try {
      const state = await api.setActiveBoard(id);
      set({ state, selectedCardId: null, view: 'board' });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  saveViewport: (boardId, viewport) => {
    void api.saveViewport(boardId, viewport).catch(() => undefined);
  },

  addCard: async (path, x, y) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.addCard(boardId, path, x, y);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  createFolderCard: async (parent, name, x, y) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return false;
    try {
      const state = await api.createFolderCard(boardId, parent, name, x, y);
      set({ state });
      get().showToast('Carpeta creada en el disco', 'success');
      return true;
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
      return false;
    }
  },

  addNote: async (x, y) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.addNote(boardId, 'Nueva nota', x, y, 230, 170);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  updateNote: async (noteId, text, color) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.updateNote(boardId, noteId, text, color || '#fff4b8');
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  moveNote: (noteId, x, y) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    void api.moveNote(boardId, noteId, x, y).catch(() => undefined);
  },

  removeNote: async (noteId) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.removeNote(boardId, noteId);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  addArrow: async (x1, y1, x2, y2, type, color) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.addArrow(boardId, x1, y1, x2, y2, type, color);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  updateArrow: async (arrowId, x1, y1, x2, y2, type, color) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    const current = get().state;
    if (current) {
      set({
        state: {
          ...current,
          boards: current.boards.map((board) =>
            board.id === boardId
              ? {
                  ...board,
                  arrows: board.arrows.map((arrow) =>
                    arrow.id === arrowId ? { ...arrow, x1, y1, x2, y2, type, color } : arrow,
                  ),
                }
              : board,
          ),
        },
      });
    }
    try {
      const state = await api.updateArrow(boardId, arrowId, x1, y1, x2, y2, type, color);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
      await get().refresh();
    }
  },

  setItemColor: async (itemId, color) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.setItemColor(boardId, itemId, color);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  renamePath: async (path, newName) => {
    try {
      const state = await api.renamePath(path, newName);
      const old = path.toLowerCase();
      const reviewPath = get().reviewPath;
      set({
        state,
        reviewPath:
          reviewPath && (reviewPath.toLowerCase() === old || reviewPath.toLowerCase().startsWith(old + '\\'))
            ? null
            : reviewPath,
        view:
          reviewPath && (reviewPath.toLowerCase() === old || reviewPath.toLowerCase().startsWith(old + '\\'))
            ? 'board'
            : get().view,
      });
      get().showToast('Elemento renombrado en el disco', 'success');
      return true;
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
      return false;
    }
  },

  removeArrow: async (arrowId) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.removeArrow(boardId, arrowId);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  moveCard: (boardId, cardId, x, y) => {
    void api.moveCard(boardId, cardId, x, y).catch(() => undefined);
  },

  removeCard: async (cardId) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.removeCard(boardId, cardId);
      set({ state, selectedCardId: null });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  startTranscription: async (path, cardId) => {
    try {
      const job = await api.startTranscription(path, cardId);
      set((s) => ({
        jobs: [job, ...s.jobs.filter((j) => j.id !== job.id)],
        jobsOpen: true,
      }));
      get().showToast('Transcripción añadida a la cola', 'info');
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  cancelJob: async (id) => {
    try {
      await api.cancelJob(id);
      set((s) => ({
        jobs: s.jobs.map((j) =>
          j.id === id ? { ...j, status: 'canceled', stage: 'Cancelado' } : j,
        ),
      }));
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  jobUpdated: (job) => {
    const previous = get().jobs.find((j) => j.id === job.id);
    set((s) => ({
      jobs: [job, ...s.jobs.filter((j) => j.id !== job.id)],
      mediaVersion:
        job.status === 'done' || job.status === 'error' || job.status === 'canceled'
          ? s.mediaVersion + 1
          : s.mediaVersion,
    }));
    if (job.status === 'done' && previous?.status !== 'done') {
      get().showToast('Transcripción lista', 'success');
    }
    if (job.status === 'error' && previous?.status !== 'error') {
      get().showToast(job.error || 'La transcripción falló', 'error');
    }
  },

  transcriptReady: (path) => {
    set((s) => ({ mediaVersion: s.mediaVersion + 1 }));
    void path;
  },

  openReview: (path) => set({ view: 'review', reviewPath: path, selectedCardId: null }),
  closeReview: () => set({ view: 'board', reviewPath: null }),
  selectCard: (id) => set({ selectedCardId: id }),
  openFolder: (path) => set({ folderPath: path }),
  closeFolder: () => set({ folderPath: null }),
  setJobsOpen: (open) => set({ jobsOpen: open }),
  setAppState: (state) => set({ state }),

  modelProgress: (download) => {
    set((s) => ({
      modelDownloads: { ...s.modelDownloads, [download.file]: download },
    }));
    if (download.status === 'done') {
      get().showToast('Modelo descargado y seleccionado', 'success');
    } else if (download.status === 'error') {
      get().showToast(download.error || 'No se pudo descargar el modelo', 'error');
    }
  },

  selectModel: async (file) => {
    try {
      const state = await api.selectModel(file);
      set({ state });
      get().showToast('Modelo activo actualizado', 'success');
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  setLanguage: async (language) => {
    try {
      const state = await api.setLanguage(language);
      set({ state });
      get().showToast(
        language === 'auto'
          ? 'Idioma automático activado'
          : language === 'en'
            ? 'Idioma: English'
            : 'Idioma: Español',
        'success',
      );
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  downloadModel: async (file) => {
    try {
      const download = await api.downloadModel(file);
      set((s) => ({
        modelDownloads: { ...s.modelDownloads, [file]: download },
      }));
      if (download.status !== 'done') {
        get().showToast('Descargando modelo…', 'info');
      }
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  cancelModelDownload: async (file) => {
    try {
      await api.cancelModelDownload(file);
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  openModelsFolder: async () => {
    try {
      await api.openModelsFolder();
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  openWorkflowDialog: async (preselect = []) => {
    set({ workflowDialogOpen: true, workflowPreselect: preselect });
    try {
      const [workflows, workflowVideos, workflowJobs] = await Promise.all([
        api.getWorkflows(),
        api.getWorkflowVideos(),
        api.getWorkflowJobs(),
      ]);
      set({ workflows, workflowVideos, workflowJobs });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  closeWorkflowDialog: () => set({ workflowDialogOpen: false, workflowPreselect: [] }),

  refreshWorkflowData: async () => {
    try {
      const [workflows, workflowVideos, workflowJobs] = await Promise.all([
        api.getWorkflows(),
        api.getWorkflowVideos(),
        api.getWorkflowJobs(),
      ]);
      set({ workflows, workflowVideos, workflowJobs });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  startWorkflows: async (workflowIds, videoPaths, language) => {
    const boardId = activeBoardId(get().state);
    if (!boardId || workflowIds.length === 0 || videoPaths.length === 0) return;
    set({ workflowDialogOpen: false, workflowPreselect: [] });
    let started = 0;
    for (const workflowId of workflowIds) {
      for (const videoPath of videoPaths) {
        try {
          const job = await api.startWorkflow(boardId, workflowId, videoPath, language);
          started++;
          set((s) => ({
            workflowJobs: [job, ...s.workflowJobs.filter((j) => j.id !== job.id)],
          }));
        } catch (err) {
          get().showToast(friendlyError(err), 'error');
        }
      }
    }
    if (started > 0) {
      set({ workflowPanelOpen: true });
      get().showToast(
        started === 1 ? 'Workflow añadido a la cola' : `${started} workflows añadidos a la cola`,
        'info',
      );
    }
  },

  startTranslation: async (resultId, targetLanguage) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const job = await api.startTranslation(boardId, resultId, targetLanguage);
      set((s) => ({
        workflowJobs: [job, ...s.workflowJobs.filter((j) => j.id !== job.id)],
        workflowPanelOpen: true,
      }));
      get().showToast('Traducción añadida a la cola', 'info');
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  workflowJobUpdated: (job) => {
    const previous = get().workflowJobs.find((j) => j.id === job.id);
    set((s) => ({
      workflowJobs: [job, ...s.workflowJobs.filter((j) => j.id !== job.id)],
    }));
    if (job.status === 'done' && previous?.status !== 'done') {
      get().showToast(`${job.workflowName}: informe listo`, 'success');
    }
    if (job.status === 'error' && previous?.status !== 'error') {
      get().showToast(job.error || 'El workflow falló', 'error');
    }
  },

  cancelWorkflow: async (id) => {
    try {
      await api.cancelWorkflow(id);
      set((s) => ({
        workflowJobs: s.workflowJobs.map((job) =>
          job.id === id ? { ...job, status: 'canceled', stage: 'Cancelado' } : job,
        ),
      }));
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  moveWorkflowResult: (resultId, x, y) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    void api.moveWorkflowResult(boardId, resultId, x, y).catch(() => undefined);
  },

  removeWorkflowResult: async (resultId) => {
    const boardId = activeBoardId(get().state);
    if (!boardId) return;
    try {
      const state = await api.removeWorkflowResult(boardId, resultId);
      set({ state });
    } catch (err) {
      get().showToast(friendlyError(err), 'error');
    }
  },

  setWorkflowPanelOpen: (open) => set({ workflowPanelOpen: open }),

  toggleTheme: () => {
    const next = get().theme === 'dark' ? 'light' : 'dark';
    try {
      localStorage.setItem(THEME_KEY, next);
    } catch {
      // localStorage may be unavailable in restricted environments.
    }
    set({ theme: next });
  },

  showToast: (message, kind = 'info') => {
    if (toastTimer) window.clearTimeout(toastTimer);
    const toast: Toast = { id: Date.now(), message, kind };
    set({ toast });
    toastTimer = window.setTimeout(() => set({ toast: null }), 5200);
  },

  dismissToast: () => {
    if (toastTimer) window.clearTimeout(toastTimer);
    set({ toast: null });
  },
}));

export function activeJobForPath(jobs: Job[], path: string): Job | undefined {
  const normalized = path.toLowerCase();
  return jobs.find(
    (j) =>
      j.mediaPath.toLowerCase() === normalized &&
      ['queued', 'extracting', 'transcribing', 'saving'].includes(j.status),
  );
}

export function boardById(state: AppState | null) {
  if (!state) return undefined;
  return state.boards.find((b) => b.id === state.activeBoardId) ?? state.boards[0];
}
