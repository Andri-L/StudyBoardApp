import { useEffect, useRef, useState } from 'react';
import {
  AlertTriangle,
  Check,
  ChevronDown,
  Cpu,
  Download,
  FolderOpen,
  Loader2,
  X,
} from 'lucide-react';
import { useStudyBoard } from '../store';
import { formatBytes } from '../api';

export default function QuickSettings() {
  const [open, setOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);
  const engine = useStudyBoard((s) => s.state?.engine);
  const language = useStudyBoard((s) => s.state?.language ?? 'es');
  const downloads = useStudyBoard((s) => s.modelDownloads);
  const selectModel = useStudyBoard((s) => s.selectModel);
  const setLanguage = useStudyBoard((s) => s.setLanguage);
  const downloadModel = useStudyBoard((s) => s.downloadModel);
  const cancelDownload = useStudyBoard((s) => s.cancelModelDownload);
  const openModelsFolder = useStudyBoard((s) => s.openModelsFolder);

  useEffect(() => {
    if (!open) return;
    const onDown = (event: MouseEvent) => {
      if (!wrapRef.current?.contains(event.target as globalThis.Node)) setOpen(false);
    };
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    window.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDown);
      window.removeEventListener('keydown', onKey);
    };
  }, [open]);

  const exactModel = engine?.modelName || 'sin modelo';
  const shortModel = exactModel.replace(/^ggml-/, '').replace(/\.bin$/, '');
  const models = engine?.models ?? [];
  const activeModel = models.find((m) => m.active);

  return (
    <div className="engine-menu-wrap" ref={wrapRef}>
      <button
        className={`engine-pill model-pill ${engine?.ready ? 'ok' : 'warn'} ${open ? 'open' : ''}`}
        title={`Modelo: ${exactModel} · Backend: ${engine?.gpu || 'CPU'}`}
        onClick={() => setOpen((value) => !value)}
      >
        <Cpu size={13} />
        <span className="model-pill-name">{shortModel}</span>
        <span className="model-pill-backend">{engine?.gpu || 'CPU'}</span>
        <ChevronDown size={12} className={`model-pill-chevron ${open ? 'open' : ''}`} />
      </button>

      {open && (
        <div className="quick-menu" role="dialog" aria-label="Ajustes rápidos del motor">
          <header className="quick-head">
            <div className="quick-head-title">
              <span className="quick-head-icon"><Cpu size={16} /></span>
              <div>
                <strong>Modelo de Whisper</strong>
                <span title={exactModel}>{exactModel}</span>
              </div>
            </div>
            <button className="icon-btn tiny" onClick={() => setOpen(false)} title="Cerrar">
              <X size={14} />
            </button>
          </header>

          <div className="quick-meta">
            <span className={`status-dot ${engine?.ready ? 'ok' : 'warn'}`} />
            {engine?.ready ? (
              <span>
                Motor listo · {engine.gpu || 'CPU'}
                {activeModel ? ` · ${formatBytes(activeModel.size)}` : ''}
              </span>
            ) : (
              <span className="quick-warning">
                <AlertTriangle size={12} />
                Falta el motor o el modelo
              </span>
            )}
          </div>

          <div className="quick-note">
            El modelo se usa en las próximas transcripciones. Descargar uno más pequeño
            ayuda a probar velocidad sin tocar tus videos.
          </div>

          <div className="model-list">
            {models.map((model) => {
              const progress = downloads[model.file];
              const downloading = model.downloading || progress?.status === 'downloading';
              const pct = progress?.progress ?? model.progress ?? 0;
              return (
                <div className={`model-row ${model.active ? 'active' : ''}`} key={model.file}>
                  <button
                    className="model-main"
                    disabled={downloading}
                    onClick={() => {
                      if (downloading) return;
                      if (model.downloaded) {
                        if (!model.active) void selectModel(model.file);
                      } else {
                        void downloadModel(model.file);
                      }
                    }}
                    title={model.downloaded ? model.file : `Descargar ${model.file}`}
                  >
                    <span className={`model-radio ${model.active ? 'active' : ''}`}>
                      {model.active && <Check size={11} />}
                    </span>
                    <span className="model-copy">
                      <span className="model-name">{model.name}</span>
                      <span className="model-desc">{model.description}</span>
                    </span>
                  </button>
                  <span className="model-size">{formatBytes(model.size)}</span>

                  {downloading ? (
                    <div className="model-downloading">
                      <span>{Math.round(pct)}%</span>
                      <button
                        className="icon-btn tiny"
                        title="Cancelar descarga"
                        onClick={() => void cancelDownload(model.file)}
                      >
                        <X size={12} />
                      </button>
                    </div>
                  ) : !model.downloaded ? (
                    <button
                      className="icon-btn tiny model-download"
                      title={`Descargar ${model.name}`}
                      onClick={() => void downloadModel(model.file)}
                    >
                      <Download size={14} />
                    </button>
                  ) : model.active ? (
                    <span className="model-active-tag">Activo</span>
                  ) : (
                    <span className="model-ready-tag">Listo</span>
                  )}

                  {downloading && (
                    <div className="model-progress">
                      <div className="progress-track">
                        <div className="progress-fill" style={{ width: `${Math.max(2, pct)}%` }} />
                      </div>
                    </div>
                  )}
                  {progress?.status === 'error' && (
                    <div className="model-error" title={progress.error}>{progress.error}</div>
                  )}
                </div>
              );
            })}
          </div>

          <div className="quick-language">
            <div className="quick-language-head">
              <span>Idioma de la clase</span>
              <small>{language === 'auto' ? 'Detección automática' : language.toUpperCase()}</small>
            </div>
            <div className="language-options">
              {[
                { value: 'es', label: 'Español' },
                { value: 'en', label: 'English' },
                { value: 'auto', label: 'Auto' },
              ].map((option) => (
                <button
                  key={option.value}
                  className={`language-btn ${language === option.value ? 'active' : ''}`}
                  onClick={() => void setLanguage(option.value)}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>

          <footer className="quick-footer">
            <button className="btn btn-ghost btn-sm" onClick={() => void openModelsFolder()}>
              <FolderOpen size={14} />
              Abrir carpeta de modelos
            </button>
            <button className="btn btn-primary btn-sm" onClick={() => setOpen(false)}>
              Listo
            </button>
          </footer>
        </div>
      )}
    </div>
  );
}
