import { AlertTriangle, Check, Info, X } from 'lucide-react';
import { useStudyBoard } from '../store';

export default function Toast() {
  const toast = useStudyBoard((s) => s.toast);
  const dismiss = useStudyBoard((s) => s.dismissToast);
  if (!toast) return null;

  return (
    <div className={`toast ${toast.kind}`} role="status">
      <span className="toast-icon">
        {toast.kind === 'success' ? (
          <Check size={15} />
        ) : toast.kind === 'error' ? (
          <AlertTriangle size={15} />
        ) : (
          <Info size={15} />
        )}
      </span>
      <span className="toast-message">{toast.message}</span>
      <button className="icon-btn tiny" onClick={dismiss} title="Cerrar">
        <X size={13} />
      </button>
    </div>
  );
}
