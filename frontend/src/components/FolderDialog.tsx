import { useEffect, useRef, useState } from 'react';
import { FolderPlus, X } from 'lucide-react';
import { useStudyBoard } from '../store';
import { compactPath } from '../api';

interface Props {
  parent: string;
  x: number;
  y: number;
  onClose: () => void;
}

export default function FolderDialog({ parent, x, y, onClose }: Props) {
  const createFolderCard = useStudyBoard((s) => s.createFolderCard);
  const [name, setName] = useState('');
  const [saving, setSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const submit = async () => {
    const value = name.trim();
    if (!value || saving) return;
    setSaving(true);
    const ok = await createFolderCard(parent, value, x, y);
    setSaving(false);
    if (ok) onClose();
  };

  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <div className="folder-dialog" onMouseDown={(e) => e.stopPropagation()}>
        <header className="folder-dialog-head">
          <span className="folder-dialog-icon"><FolderPlus size={19} /></span>
          <div>
            <h2>Nueva carpeta</h2>
            <p>Se creará en el disco y se añadirá al tablero.</p>
          </div>
          <button className="icon-btn tiny" onClick={onClose} title="Cerrar">
            <X size={15} />
          </button>
        </header>

        <div className="folder-dialog-body">
          <label className="field-label" htmlFor="new-folder-name">Nombre</label>
          <input
            id="new-folder-name"
            ref={inputRef}
            className="text-input"
            value={name}
            placeholder="Ej. 05Nov26"
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void submit();
              if (e.key === 'Escape') onClose();
            }}
          />
          <div className="folder-parent-label">Se creará dentro de</div>
          <div className="folder-parent-path" title={parent}>{compactPath(parent, 70)}</div>
          <div className="folder-dialog-note">
            Si quieres crearla dentro de otra carpeta, selecciona primero su tarjeta
            en el tablero y vuelve a usar “Nueva carpeta”.
          </div>
        </div>

        <footer className="folder-dialog-foot">
          <button className="btn btn-ghost" onClick={onClose}>Cancelar</button>
          <button className="btn btn-primary" disabled={!name.trim() || saving} onClick={() => void submit()}>
            {saving ? 'Creando…' : 'Crear carpeta'}
          </button>
        </footer>
      </div>
    </div>
  );
}
