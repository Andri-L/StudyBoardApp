import { useMemo, useState } from 'react';
import {
  FolderOpen,
  HardDrive,
  MoreHorizontal,
  Plus,
  RefreshCw,
  Trash2,
} from 'lucide-react';
import { useStudyBoard, boardById } from '../store';
import { compactPath } from '../api';
import FolderTree from './FolderTree';

export default function Sidebar() {
  const state = useStudyBoard((s) => s.state);
  const busy = useStudyBoard((s) => s.busy);
  const chooseWorkspace = useStudyBoard((s) => s.chooseWorkspace);
  const createBoard = useStudyBoard((s) => s.createBoard);
  const renameBoard = useStudyBoard((s) => s.renameBoard);
  const deleteBoard = useStudyBoard((s) => s.deleteBoard);
  const setActiveBoard = useStudyBoard((s) => s.setActiveBoard);

  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [refreshKey, setRefreshKey] = useState(0);

  const board = boardById(state);
  const workspaceName = useMemo(() => {
    const parts = (state?.workspace || '').split(/[\\/]/).filter(Boolean);
    return parts[parts.length - 1] || 'Clases';
  }, [state?.workspace]);

  if (!state) return null;

  const submitNew = () => {
    const name = newName.trim();
    if (name) void createBoard(name);
    setNewName('');
    setCreating(false);
  };

  const submitRename = (id: string) => {
    const name = editName.trim();
    if (name) void renameBoard(id, name);
    setEditingId(null);
    setEditName('');
  };

  return (
    <aside className="sidebar">
      <div className="sidebar-brand">
        <div className="brand-mark small">A</div>
        <div>
          <div className="brand-name">StudyBoard App</div>
          <div className="brand-tag">Espacio de clases</div>
        </div>
      </div>

      <div className="sidebar-section boards-section">
        <div className="section-head">
          <span>Tableros</span>
          <button
            className="icon-btn tiny"
            title="Nuevo tablero"
            onClick={() => {
              setCreating(true);
              setNewName('');
            }}
          >
            <Plus size={14} />
          </button>
        </div>

        <div className="board-list">
          {state.boards.map((b) => {
            const active = b.id === state.activeBoardId;
            if (editingId === b.id) {
              return (
                <input
                  key={b.id}
                  className="inline-input board-input"
                  autoFocus
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') submitRename(b.id);
                    if (e.key === 'Escape') setEditingId(null);
                  }}
                  onBlur={() => submitRename(b.id)}
                />
              );
            }
            return (
              <div
                key={b.id}
                className={`board-row ${active ? 'active' : ''}`}
                onClick={() => void setActiveBoard(b.id)}
                onDoubleClick={() => {
                  setEditingId(b.id);
                  setEditName(b.name);
                }}
                title={`${b.name} · ${b.cards.length} tarjetas`}
              >
                <span className={`board-pip ${active ? 'active' : ''}`} />
                <span className="board-name">{b.name}</span>
                <span className="board-count">{b.cards.length}</span>
                {state.boards.length > 1 && (
                  <button
                    className="icon-btn tiny board-delete"
                    title="Eliminar tablero (no borra archivos)"
                    onClick={(e) => {
                      e.stopPropagation();
                      void deleteBoard(b.id);
                    }}
                  >
                    <Trash2 size={13} />
                  </button>
                )}
              </div>
            );
          })}

          {creating && (
            <input
              className="inline-input board-input"
              autoFocus
              placeholder="Nombre del tablero"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') submitNew();
                if (e.key === 'Escape') setCreating(false);
              }}
              onBlur={submitNew}
            />
          )}
        </div>
      </div>

      <div className="sidebar-section tree-section">
        <div className="section-head">
          <span>Carpetas</span>
          <div className="section-actions">
            <button
              className="icon-btn tiny"
              title="Actualizar carpetas"
              onClick={() => setRefreshKey((k) => k + 1)}
            >
              <RefreshCw size={13} />
            </button>
          </div>
        </div>
        <div className="tree-scroll">
          {state.workspace ? (
            <FolderTree
              key={state.workspace}
              entry={{
                name: workspaceName,
                path: state.workspace,
                isDir: true,
                size: 0,
                modTime: 0,
                kind: 'folder',
              }}
              depth={0}
              defaultOpen
              refreshKey={refreshKey}
            />
          ) : (
            <div className="tree-empty">Sin carpeta seleccionada</div>
          )}
        </div>
      </div>

      <div className="sidebar-footer">
        <div className="workspace-current" title={state.workspace}>
          <HardDrive size={14} />
          <div>
            <span>{workspaceName}</span>
            <small>{compactPath(state.workspace, 36)}</small>
          </div>
        </div>
        <button
          className="btn btn-ghost btn-compact"
          onClick={() => void chooseWorkspace()}
          disabled={busy}
        >
          <FolderOpen size={14} />
          Cambiar carpeta
        </button>
        <div className="sidebar-hint">
          <MoreHorizontal size={13} />
          Las tarjetas solo referencian archivos
        </div>
      </div>
    </aside>
  );
}
