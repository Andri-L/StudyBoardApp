import { useEffect, useMemo, useState } from 'react';
import {
  ArrowLeft,
  ExternalLink,
  FileVideo,
  Folder,
  FolderOpen,
  Loader2,
  Play,
  Plus,
  X,
} from 'lucide-react';
import { boardById, useStudyBoard } from '../store';
import { api, compactPath, fileName, friendlyError, nextCardPosition } from '../api';
import type { DirEntry } from '../types';

export default function FolderDrawer() {
  const folderPath = useStudyBoard((s) => s.folderPath);
  const closeFolder = useStudyBoard((s) => s.closeFolder);
  const openFolder = useStudyBoard((s) => s.openFolder);
  const state = useStudyBoard((s) => s.state);
  const addCard = useStudyBoard((s) => s.addCard);
  const openReview = useStudyBoard((s) => s.openReview);
  const showToast = useStudyBoard((s) => s.showToast);
  const board = boardById(state);

  const current = folderPath || state?.workspace || '';
  const [entries, setEntries] = useState<DirEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [tick, setTick] = useState(0);

  useEffect(() => {
    if (!folderPath || !state?.workspace) return;
    let cancelled = false;
    setLoading(true);
    setError('');
    api
      .listDir(folderPath)
      .then((items) => {
        if (!cancelled) setEntries(items);
      })
      .catch((err) => {
        if (!cancelled) setError(friendlyError(err));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [folderPath, state?.workspace, tick]);

  const canGoUp = useMemo(() => {
    if (!folderPath || !state?.workspace) return false;
    return folderPath.toLowerCase() !== state.workspace.toLowerCase();
  }, [folderPath, state?.workspace]);

  if (!folderPath) return null;

  const add = (entry: DirEntry) => {
    const position = nextCardPosition(board?.cards ?? []);
    void addCard(entry.path, position.x, position.y);
    showToast(entry.isDir ? 'Carpeta añadida al tablero' : 'Video añadido al tablero', 'success');
  };

  const addCurrentFolder = () => {
    if (!folderPath) return;
    const position = nextCardPosition(board?.cards ?? []);
    void addCard(folderPath, position.x, position.y);
    showToast('Carpeta añadida al tablero', 'success');
  };

  const goUp = () => {
    if (!canGoUp || !state?.workspace) return;
    const parent = folderPath.replace(/[\\/][^\\/]+[\\/]?$/, '');
    if (parent && parent.length >= state.workspace.length) openFolder(parent);
    else openFolder(state.workspace);
  };

  return (
    <div className="drawer-backdrop" onMouseDown={closeFolder}>
      <section className="folder-drawer" onMouseDown={(e) => e.stopPropagation()}>
        <header className="drawer-head">
          <div className="drawer-head-copy">
            <span className="drawer-icon"><FolderOpen size={18} /></span>
            <div>
              <h2>{fileName(current)}</h2>
              <span className="drawer-path" title={current}>{compactPath(current, 56)}</span>
            </div>
          </div>
          <div className="drawer-head-actions">
            <button className="btn btn-primary btn-sm" onClick={addCurrentFolder}>
              <Plus size={14} />
              Añadir carpeta
            </button>
            <button className="icon-btn" title="Cerrar" onClick={closeFolder}>
              <X size={17} />
            </button>
          </div>
        </header>

        <div className="drawer-nav">
          <button className="icon-btn tiny" onClick={goUp} disabled={!canGoUp} title="Subir un nivel">
            <ArrowLeft size={14} />
          </button>
          <span className="drawer-crumb">
            {state?.workspace && current.startsWith(state.workspace)
              ? compactPath(current.slice(state.workspace.length).replace(/^[\\/]/, '') || fileName(state.workspace), 56)
              : compactPath(current, 56)}
          </span>
          <button className="icon-btn tiny" title="Actualizar" onClick={() => setTick((v) => v + 1)}>
            <Loader2 size={14} className={loading ? 'spin' : ''} />
          </button>
        </div>

        <div className="drawer-body">
          {loading && entries.length === 0 ? (
            <div className="drawer-empty"><Loader2 size={18} className="spin" /> Cargando carpeta…</div>
          ) : error ? (
            <div className="drawer-empty error">{error}</div>
          ) : entries.length === 0 ? (
            <div className="drawer-empty">Esta carpeta está vacía.</div>
          ) : (
            entries.map((entry) => (
              <div key={entry.path} className="drawer-row" onDoubleClick={() => (entry.isDir ? openFolder(entry.path) : openReview(entry.path))}>
                <span className={`drawer-row-icon ${entry.isDir ? 'folder' : 'video'}`}>
                  {entry.isDir ? <Folder size={17} /> : <FileVideo size={17} />}
                </span>
                <div className="drawer-row-copy">
                  <div className="drawer-row-name" title={entry.path}>{entry.name}</div>
                  <div className="drawer-row-sub">
                    {entry.isDir ? 'Carpeta' : 'Video'}
                    {entry.dateLabel ? ` · ${entry.dateLabel}` : ''}
                  </div>
                </div>
                <div className="drawer-row-actions">
                  {!entry.isDir && (
                    <button className="icon-btn tiny" title="Reproducir y ver transcripción" onClick={() => openReview(entry.path)}>
                      <Play size={13} />
                    </button>
                  )}
                  {!entry.isDir && (
                    <button
                      className="icon-btn tiny"
                      title="Abrir con el reproductor de Windows"
                      onClick={() => void api.openExternal(entry.path).catch((err) => showToast(friendlyError(err), 'error'))}
                    >
                      <ExternalLink size={13} />
                    </button>
                  )}
                  <button className="icon-btn tiny" title="Añadir al tablero" onClick={() => add(entry)}>
                    <Plus size={14} />
                  </button>
                </div>
              </div>
            ))
          )}
        </div>

        <footer className="drawer-foot">
          <span>Doble clic para abrir · Los archivos no se mueven</span>
          {entries.some((e) => e.isDir) && <span>Explora las subcarpetas dentro del panel</span>}
        </footer>
      </section>
    </div>
  );
}
