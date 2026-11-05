import {
  AlertTriangle,
  Check,
  Cpu,
  FolderOpen,
  GraduationCap,
  HardDrive,
  Sparkles,
} from 'lucide-react';
import { useStudyBoard } from '../store';
import { compactPath } from '../api';

export default function Welcome() {
  const state = useStudyBoard((s) => s.state);
  const busy = useStudyBoard((s) => s.busy);
  const chooseWorkspace = useStudyBoard((s) => s.chooseWorkspace);

  const engine = state?.engine;
  const ready = Boolean(engine?.ready);

  return (
    <div className="welcome">
      <div className="welcome-glow" />
      <header className="welcome-header">
        <div className="brand-lockup">
          <div className="brand-mark">A</div>
          <div>
            <div className="brand-name">StudyBoard App</div>
            <div className="brand-tag">Espacio de clases</div>
          </div>
        </div>
        <div className="welcome-local">
          <HardDrive size={14} />
          Todo se procesa en tu PC
        </div>
      </header>

      <main className="welcome-main">
        <div className="welcome-copy">
          <div className="eyebrow">
            <GraduationCap size={15} />
            Clases virtuales, sin ruido
          </div>
          <h1>
            Tus clases, ordenadas.
            <br />
            <span>Tu tiempo, de vuelta.</span>
          </h1>
          <p>
            Un tablero visual para tus carpetas y un transcriptor local que convierte
            cada grabación en texto con timestamps. Después podrás llevar ese texto a
            tu LLM favorito.
          </p>

          <div className="welcome-actions">
            <button className="btn btn-primary btn-lg" onClick={() => void chooseWorkspace()} disabled={busy}>
              <FolderOpen size={17} />
              {busy ? 'Abriendo…' : 'Elegir carpeta de clases'}
            </button>
            <span className="welcome-hint">Puedes cambiarla después</span>
          </div>

          <div className="feature-row">
            <div className="feature">
              <span className="feature-icon"><FolderOpen size={16} /></span>
              <div>
                <strong>Tablero visual</strong>
                <span>Arrastra carpetas y videos, sin mover archivos.</span>
              </div>
            </div>
            <div className="feature">
              <span className="feature-icon"><Sparkles size={16} /></span>
              <div>
                <strong>Whisper local</strong>
                <span>Español, timestamps y GPU AMD con Vulkan.</span>
              </div>
            </div>
          </div>
        </div>

        <aside className="engine-card">
          <div className="engine-card-head">
            <div>
              <div className="engine-kicker">Motor de transcripción</div>
              <div className="engine-title">Whisper en tu PC</div>
            </div>
            <span className={`status-dot ${ready ? 'ok' : 'warn'}`} />
          </div>

          <div className="engine-checks">
            <CheckRow label="Whisper.cpp" value={engine?.whisperPath ? 'Detectado' : 'No encontrado'} ok={Boolean(engine?.whisperPath)} />
            <CheckRow label="FFmpeg" value={engine?.ffmpegPath ? 'Detectado' : 'No encontrado'} ok={Boolean(engine?.ffmpegPath)} />
            <CheckRow label="Modelo de voz" value={engine?.modelName || 'No encontrado'} ok={Boolean(engine?.modelPath)} />
            <CheckRow label="Aceleración" value={engine?.gpu || 'CPU'} ok={Boolean(engine?.gpu)} neutral />
          </div>

          {ready ? (
            <div className="engine-ready">
              <Check size={15} />
              Motor listo para transcribir
            </div>
          ) : (
            <div className="engine-missing">
              <AlertTriangle size={16} />
              <div>
                <strong>Falta el motor</strong>
                <p>
                  Coloca las herramientas en la carpeta <code>bin</code> junto a la app.
                  {engine?.missing?.length ? ` Faltan: ${engine.missing.join(', ')}.` : ''}
                </p>
              </div>
            </div>
          )}

          <div className="engine-paths">
            <div>
              <span><Cpu size={12} /> Whisper</span>
              <code title={engine?.whisperPath}>{compactPath(engine?.whisperPath || '—', 38)}</code>
            </div>
            <div>
              <span>Modelo</span>
              <code title={engine?.modelPath}>{compactPath(engine?.modelPath || '—', 38)}</code>
            </div>
          </div>
        </aside>
      </main>

      <footer className="welcome-footer">
        <span>Hecho para estudiar mejor, no para ver más videos.</span>
        <span>v1.0.0 · Windows</span>
      </footer>
    </div>
  );
}

function CheckRow({
  label,
  value,
  ok,
  neutral = false,
}: {
  label: string;
  value: string;
  ok: boolean;
  neutral?: boolean;
}) {
  return (
    <div className="check-row">
      <span className={`check-badge ${ok ? 'ok' : 'warn'}`}>
        {ok ? <Check size={12} /> : <AlertTriangle size={12} />}
      </span>
      <span className="check-label">{label}</span>
      <span className={`check-value ${neutral ? 'neutral' : ''}`}>{value}</span>
    </div>
  );
}
