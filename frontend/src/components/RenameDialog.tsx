import { useEffect, useRef, useState } from 'react';
import { Pencil, X } from 'lucide-react';
import { useStudyBoard } from '../store';

interface Props {
  path: string;
  initialName: string;
  onClose: () => void;
}

export default function RenameDialog({ path, initialName, onClose }: Props) {
  const [name, setName] = useState(initialName);
  const [busy, setBusy] = useState(false);
  const renamePath = useStudyBoard((s) => s.renamePath);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, []);

  const submit = async () => {
    if (busy || !name.trim() || name === initialName) {
      if (name === initialName) onClose();
      return;
    }
    setBusy(true);
    const ok = await renamePath(path, name);
    setBusy(false);
    if (ok) onClose();
  };

  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <section
        className="rename-dialog"
        onMouseDown={(event) => event.stopPropagation()}
        onKeyDown={(event) => {
          if (event.key === 'Escape') onClose();
          if (event.key === 'Enter') void submit();
        }}
      >
        <header className="rename-dialog-head">
          <span className="rename-dialog-icon"><Pencil size={16} /></span>
          <div>
            <h2>Renombrar en el disco</h2>
            <p title={path}>{path}</p>
          </div>
          <button className="icon-btn tiny" onClick={onClose} title="Cerrar">
            <X size={15} />
          </button>
        </header>
        <div className="rename-dialog-body">
          <label htmlFor="rename-input">Nuevo nombre</label>
          <input
            id="rename-input"
            ref={inputRef}
            value={name}
            onChange={(event) => setName(event.target.value)}
            spellCheck={false}
          />
          <span className="rename-dialog-hint">
            Se renombra el archivo o carpeta real. Si es un video, conserva la extensión.
          </span>
        </div>
        <footer className="rename-dialog-foot">
          <button className="btn btn-ghost" onClick={onClose}>Cancelar</button>
          <button className="btn btn-primary" disabled={busy || !name.trim()} onClick={() => void submit()}>
            {busy ? 'Renombrando…' : 'Renombrar'}
          </button>
        </footer>
      </section>
    </div>
  );
}
