import { AlertTriangle, Check, Loader2, Workflow, X } from 'lucide-react';
import { useStudyBoard } from '../store';
import { fileName } from '../api';

const active = ['queued', 'running'];

export default function WorkflowJobsPanel() {
  const jobs = useStudyBoard((s) => s.workflowJobs);
  const setOpen = useStudyBoard((s) => s.setWorkflowPanelOpen);
  const cancelWorkflow = useStudyBoard((s) => s.cancelWorkflow);

  return (
    <aside className="jobs-panel workflow-jobs-panel">
      <header className="jobs-head">
        <div>
          <h2>Workflows de OpenCode</h2>
          <span>{jobs.length ? `${jobs.length} en esta sesión` : 'Sin ejecuciones todavía'}</span>
        </div>
        <button className="icon-btn tiny" onClick={() => setOpen(false)} title="Cerrar">
          <X size={15} />
        </button>
      </header>

      <div className="jobs-list">
        {jobs.length === 0 ? (
          <div className="jobs-empty">
            <Workflow size={22} />
            <strong>No hay workflows</strong>
            <span>Ejecuta uno desde el tablero y aparecerá aquí.</span>
          </div>
        ) : (
          jobs.map((job) => {
            const running = active.includes(job.status);
            return (
              <div key={job.id} className={`job-row workflow-job ${job.status}`}>
                <div className="job-row-head">
                  <span className={`job-icon ${job.status}`}>
                    {running ? (
                      <Loader2 size={13} className="spin" />
                    ) : job.status === 'done' ? (
                      <Check size={13} />
                    ) : job.status === 'error' ? (
                      <AlertTriangle size={13} />
                    ) : (
                      <Workflow size={13} />
                    )}
                  </span>
                  <div className="job-copy">
                    <div className="job-name" title={job.videoPath}>{job.workflowName}</div>
                    <div className="job-sub">
                      {fileName(job.videoName)} · {job.stage}
                    </div>
                  </div>
                  {running ? (
                    <button
                      className="icon-btn tiny"
                      title="Cancelar workflow"
                      onClick={() => void cancelWorkflow(job.id)}
                    >
                      <X size={13} />
                    </button>
                  ) : job.status === 'done' ? (
                    <button
                      className="icon-btn tiny"
                      title="Cerrar panel"
                      onClick={() => setOpen(false)}
                    >
                      <Check size={13} />
                    </button>
                  ) : null}
                </div>
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
