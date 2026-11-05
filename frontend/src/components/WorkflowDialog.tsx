import { useEffect, useMemo, useState } from 'react';
import {
  Check,
  ClipboardList,
  Languages,
  Loader2,
  Play,
  Target,
  Workflow,
  X,
} from 'lucide-react';
import { useStudyBoard } from '../store';
import { formatDuration } from '../api';

const OUTPUT_LANGUAGE_KEY = 'studyboard.workflowLanguage';

export default function WorkflowDialog() {
  const open = useStudyBoard((s) => s.workflowDialogOpen);
  const workflows = useStudyBoard((s) => s.workflows);
  const videos = useStudyBoard((s) => s.workflowVideos);
  const preselect = useStudyBoard((s) => s.workflowPreselect);
  const close = useStudyBoard((s) => s.closeWorkflowDialog);
  const startWorkflows = useStudyBoard((s) => s.startWorkflows);
  const refresh = useStudyBoard((s) => s.refreshWorkflowData);
  const [selectedWorkflows, setSelectedWorkflows] = useState<Set<string>>(new Set());
  const [selectedVideos, setSelectedVideos] = useState<Set<string>>(new Set());
  const [language, setLanguage] = useState<'es' | 'en'>(
    (localStorage.getItem(OUTPUT_LANGUAGE_KEY) as 'es' | 'en') || 'es',
  );

  useEffect(() => {
    if (!open) return;
    setSelectedWorkflows(new Set());
    setSelectedVideos(new Set(preselect.filter(Boolean)));
    void refresh();
  }, [open, preselect.join('|'), refresh]);

  const transcribed = useMemo(() => videos.filter((video) => video.hasTranscript), [videos]);

  if (!open) return null;

  const toggleWorkflow = (id: string) => {
    setSelectedWorkflows((current) => {
      const next = new Set(current);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const toggleVideo = (path: string) => {
    setSelectedVideos((current) => {
      const next = new Set(current);
      next.has(path) ? next.delete(path) : next.add(path);
      return next;
    });
  };

  const setOutputLanguage = (value: 'es' | 'en') => {
    setLanguage(value);
    localStorage.setItem(OUTPUT_LANGUAGE_KEY, value);
  };

  const run = () => {
    if (selectedWorkflows.size === 0 || selectedVideos.size === 0) return;
    void startWorkflows([...selectedWorkflows], [...selectedVideos], language);
  };

  return (
    <div className="modal-backdrop" onMouseDown={close}>
      <section className="workflow-dialog" onMouseDown={(event) => event.stopPropagation()}>
        <header className="workflow-dialog-head">
          <div className="workflow-head-title">
            <span className="workflow-head-icon"><Workflow size={19} /></span>
            <div>
              <h2>Ejecutar workflow</h2>
              <p>
                OpenCode se ejecutará automáticamente con tu sesión autenticada.
                Cada ejecución usa una sesión nueva.
              </p>
            </div>
          </div>
          <button className="icon-btn tiny" onClick={close} title="Cerrar">
            <X size={16} />
          </button>
        </header>

        <div className="workflow-dialog-body">
          <section className="workflow-pick-section">
            <div className="workflow-section-head">
              <div>
                <h3>1. Workflows</h3>
                <span>Elige uno o varios informes.</span>
              </div>
              <span className="workflow-count">{selectedWorkflows.size} seleccionados</span>
            </div>
            <div className="workflow-options">
              {workflows.map((workflow) => {
                const checked = selectedWorkflows.has(workflow.id);
                const Icon = workflow.id === 'exam-focus' ? Target : ClipboardList;
                return (
                  <button
                    key={workflow.id}
                    className={`workflow-option ${checked ? 'checked' : ''}`}
                    onClick={() => toggleWorkflow(workflow.id)}
                  >
                    <span className="workflow-check">{checked && <Check size={12} />}</span>
                    <span className="workflow-option-icon"><Icon size={16} /></span>
                    <span className="workflow-option-copy">
                      <strong>{workflow.name}</strong>
                      <span>{workflow.description}</span>
                    </span>
                  </button>
                );
              })}
              {workflows.length === 0 && (
                <div className="workflow-empty"><Loader2 size={15} className="spin" /> Cargando workflows…</div>
              )}
            </div>
          </section>

          <section className="workflow-pick-section">
            <div className="workflow-section-head">
              <div>
                <h3>2. Clases</h3>
                <span>Solo se pueden ejecutar sobre videos transcritos.</span>
              </div>
              <span className="workflow-count">{selectedVideos.size} seleccionados</span>
            </div>
            <div className="workflow-video-list">
              {videos.map((video) => {
                const checked = selectedVideos.has(video.path);
                return (
                  <button
                    key={video.path}
                    className={`workflow-video-option ${checked ? 'checked' : ''} ${!video.hasTranscript ? 'disabled' : ''}`}
                    disabled={!video.hasTranscript}
                    onClick={() => toggleVideo(video.path)}
                    title={video.path}
                  >
                    <span className="workflow-check">{checked && <Check size={12} />}</span>
                    <span className="workflow-video-copy">
                      <strong>{video.name}</strong>
                      <span>
                        {video.class} · {formatDuration(video.duration)} ·{' '}
                        {video.hasTranscript ? `${video.segments} segmentos` : 'sin transcripción'}
                      </span>
                    </span>
                  </button>
                );
              })}
              {videos.length === 0 && (
                <div className="workflow-empty">
                  <ClipboardList size={17} />
                  No hay videos en el workspace. Añade y transcribe una clase primero.
                </div>
              )}
              {videos.length > 0 && transcribed.length === 0 && (
                <div className="workflow-empty">
                  Hay videos, pero ninguno tiene transcripción todavía.
                </div>
              )}
            </div>
          </section>

          <section className="workflow-language-row">
            <div>
              <h3><Languages size={15} /> Idioma de salida</h3>
              <span>Los prompts están en inglés; el informe se genera en este idioma.</span>
            </div>
            <div className="language-options workflow-language-options">
              <button className={`language-btn ${language === 'es' ? 'active' : ''}`} onClick={() => setOutputLanguage('es')}>
                Español
              </button>
              <button className={`language-btn ${language === 'en' ? 'active' : ''}`} onClick={() => setOutputLanguage('en')}>
                English
              </button>
            </div>
          </section>
        </div>

        <footer className="workflow-dialog-foot">
          <span className="workflow-foot-note">
            Se guardará un Markdown junto al video y una tarjeta en el tablero.
          </span>
          <div className="workflow-foot-actions">
            <button className="btn btn-ghost" onClick={close}>Cancelar</button>
            <button
              className="btn btn-primary"
              disabled={selectedWorkflows.size === 0 || selectedVideos.size === 0}
              onClick={run}
            >
              <Play size={14} />
              Ejecutar {selectedWorkflows.size * selectedVideos.size > 1
                ? `${selectedWorkflows.size * selectedVideos.size} workflows`
                : 'workflow'}
            </button>
          </div>
        </footer>
      </section>
    </div>
  );
}
