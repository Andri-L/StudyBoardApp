import { useEffect, useState } from 'react';
import {
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  Film,
  Folder,
  FolderOpen,
  Loader2,
  Plus,
} from 'lucide-react';
import { useStudyBoard, boardById } from '../store';
import { api, friendlyError, nextCardPosition } from '../api';
import type { DirEntry } from '../types';

interface Props {
  entry: DirEntry;
  depth: number;
  defaultOpen?: boolean;
  refreshKey: number;
}

export default function FolderTree({ entry, depth, defaultOpen = false, refreshKey }: Props) {
  const [open, setOpen] = useState(defaultOpen);
  const [children, setChildren] = useState<DirEntry[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const state = useStudyBoard((s) => s.state);
  const addCard = useStudyBoard((s) => s.addCard);
  const showToast = useStudyBoard((s) => s.showToast);
  const board = boardById(state);

  useEffect(() => {
    if (!open || !entry.isDir) return;
    let cancelled = false;
    setLoading(true);
    setError('');
    api
      .listDir(entry.path)
      .then((items) => {
        if (!cancelled) setChildren(items);
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
  }, [open, entry.path, entry.isDir, refreshKey]);

  const addToBoard = () => {
    const position = nextCardPosition(board?.cards ?? []);
    void addCard(entry.path, position.x, position.y);
    showToast(entry.isDir ? 'Carpeta añadida al tablero' : 'Video añadido al tablero', 'success');
  };

  const onDragStart = (e: React.DragEvent) => {
    e.dataTransfer.setData('application/x-studyboard-path', entry.path);
    e.dataTransfer.setData('application/x-studyboard-kind', entry.kind);
    e.dataTransfer.setData('text/plain', entry.path);
    e.dataTransfer.effectAllowed = 'copy';
  };

  return (
    <div className="tree-node">
      <div
        className={`tree-row ${entry.isDir ? 'is-dir' : 'is-video'}`}
        style={{ paddingLeft: 7 + depth * 14 }}
        draggable
        onDragStart={onDragStart}
        onDoubleClick={addToBoard}
        title={entry.path}
      >
        {entry.isDir ? (
          <button
            className="tree-toggle"
            onClick={(e) => {
              e.stopPropagation();
              setOpen((v) => !v);
            }}
            aria-label={open ? 'Contraer' : 'Expandir'}
          >
            {loading ? (
              <Loader2 size={12} className="spin" />
            ) : open ? (
              <ChevronDown size={12} />
            ) : (
              <ChevronRight size={12} />
            )}
          </button>
        ) : (
          <span className="tree-toggle spacer" />
        )}
        <span className={`tree-icon ${entry.isDir ? 'folder' : 'video'}`}>
          {entry.isDir ? (open ? <FolderOpen size={15} /> : <Folder size={15} />) : <Film size={15} />}
        </span>
        <span className="tree-name">{entry.name}</span>
        {entry.dateLabel && <span className="tree-date">{entry.dateLabel}</span>}
        <button
          className="tree-add icon-btn tiny"
          onClick={(e) => {
            e.stopPropagation();
            addToBoard();
          }}
          title="Añadir al tablero activo"
        >
          <Plus size={13} />
        </button>
      </div>

      {open && entry.isDir && (
        <div className="tree-children">
          {error ? (
            <div className="tree-note error">
              <AlertTriangle size={12} /> {error}
            </div>
          ) : loading && children === null ? (
            <div className="tree-note">Cargando…</div>
          ) : children && children.length === 0 ? (
            <div className="tree-note">Carpeta vacía</div>
          ) : (
            children?.map((child) => (
              <FolderTree
                key={child.path}
                entry={child}
                depth={depth + 1}
                refreshKey={refreshKey}
              />
            ))
          )}
        </div>
      )}
    </div>
  );
}
