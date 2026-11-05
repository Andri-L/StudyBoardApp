import {
  AlertTriangle,
  Captions,
  Check,
  ExternalLink,
  Loader2,
  Play,
  X,
} from 'lucide-react';
import { useStudyBoard } from '../store';
import { api, fileName, friendlyError } from '../api';
import type { Job } from '../types';

const activeStatuses = ['queued', 'extracting', 'transcribing', 'saving'];

function statusLabel(job: Job): string {
  if (job.status === 'queued') return 'En cola';
  if (job.status === 'extracting') return 'Extrayendo audio';
  if (job.status === 'transcribing') return 'Transcribiendo';
  if (job.status === 'saving') return 'Guardando';
  if (job.status === 'done') return 'Listo';
  if (job.status === 'canceled') return 'Cancelado';
  return 'Error';
}

export default function JobsPanel() {
  const jobs = useStudyBoard((s) => s.jobs);
  const setJobsOpen = useStudyBoard((s) => s.setJobsOpen);
  const cancelJob = useStudyBoard((s) => s.cancelJob);
  const openReview = useStudyBoard((s) => s.openReview);
  const showToast = useStudyBoard((s) => s.showToast);

  return (
    <aside className="jobs-panel">
      <header className="jobs-head">
        <div>
          <h2>Tareas de transcripción</h2>
          <span>{jobs.length ? `${jobs.length} en esta sesión` : 'Sin actividad todavía'}</span>
        </div>
        <button className="icon-btn tiny" onClick={() => setJobsOpen(false)} title="Cerrar">
          <X size={15} />
        </button>
      </header>

      <div className="jobs-list">
        {jobs.length === 0 ? (
          <div className="jobs-empty">
            <Captions size={22} />
            <strong>No hay tareas</strong>
            <span>Transcribe un video desde el tablero y aparecerá aquí.</span>
          </div>
        ) : (
          jobs.map((job) => {
            const active = activeStatuses.includes(job.status);
            return (
              <div key={job.id} className={`job-row ${job.status}`}>
                <div className="job-row-head">
                  <span className={`job-icon ${job.status}`}>
                    {active ? (
                      <Loader2 size={13} className="spin" />
                    ) : job.status === 'done' ? (
                      <Check size={13} />
                    ) : job.status === 'error' ? (
                      <AlertTriangle size={13} />
                    ) : (
                      <Captions size={13} />
                    )}
                  </span>
                  <div className="job-copy">
                    <div className="job-name" title={job.mediaPath}>{fileName(job.mediaPath)}</div>
                    <div className="job-sub">
                      {statusLabel(job)}
                      {active ? ` · ${Math.round(job.progress)}%` : ''}
                    </div>
                  </div>
                  {active ? (
                    <button
                      className="icon-btn tiny"
                      title="Cancelar"
                      onClick={() => void cancelJob(job.id)}
                    >
                      <X size={13} />
                    </button>
                  ) : job.status === 'done' ? (
                    <button
                      className="icon-btn tiny"
                      title="Abrir transcripción"
                      onClick={() => {
                        openReview(job.mediaPath);
                        setJobsOpen(false);
                      }}
                    >
                      <Play size={13} />
                    </button>
                  ) : (
                    <button
                      className="icon-btn tiny"
                      title="Mostrar archivo en Explorer"
                      onClick={() => {
                        void api.reveal(job.mediaPath).catch((err) => showToast(friendlyError(err), 'error'));
                      }}
                    >
                      <ExternalLink size={13} />
                    </button>
                  )}
                </div>
                {active && (
                  <div className="progress-track">
                    <div className="progress-fill" style={{ width: `${Math.max(2, job.progress)}%` }} />
                  </div>
                )}
                {job.status === 'error' && job.error && (
                  <div className="job-error" title={job.error}>{job.error}</div>
                )}
              </div>
            );
          })
        )}
      </div>
    </aside>
  );
}
