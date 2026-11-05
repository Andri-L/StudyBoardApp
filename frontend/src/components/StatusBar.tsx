import { Cpu, ListVideo, Loader2, Workflow } from 'lucide-react';
import { useStudyBoard } from '../store';
import { compactPath } from '../api';

export default function StatusBar() {
  const state = useStudyBoard((s) => s.state);
  const jobs = useStudyBoard((s) => s.jobs);
  const jobsOpen = useStudyBoard((s) => s.jobsOpen);
  const setJobsOpen = useStudyBoard((s) => s.setJobsOpen);
  const workflowJobs = useStudyBoard((s) => s.workflowJobs);
  const workflowPanelOpen = useStudyBoard((s) => s.workflowPanelOpen);
  const setWorkflowPanelOpen = useStudyBoard((s) => s.setWorkflowPanelOpen);

  const activeJobs = jobs.filter((j) =>
    ['queued', 'extracting', 'transcribing', 'saving'].includes(j.status),
  );
  const running = activeJobs[0];
  const activeWorkflows = workflowJobs.filter((j) =>
    ['queued', 'running'].includes(j.status),
  );
  const runningWorkflow = activeWorkflows[0];
  const ready = state?.engine.ready;

  return (
    <footer className="status-bar">
      <div className="status-left">
        <span className={`status-dot ${ready ? 'ok' : 'warn'}`} />
        <span className="status-engine">
          <Cpu size={12} />
          {ready
            ? `${(state?.engine.modelName || 'Whisper').replace(/^ggml-/, '').replace(/\.bin$/, '')} · ${state?.engine.gpu || 'CPU'}`
            : 'Motor incompleto'}
        </span>
        <span className="status-divider" />
        <span className="status-path" title={state?.workspace}>
          {compactPath(state?.workspace || '', 58)}
        </span>
      </div>
      <div className="status-right">
        {workflowJobs.length > 0 && (
          <button
            className={`status-jobs ${workflowPanelOpen ? 'open' : ''}`}
            onClick={() => setWorkflowPanelOpen(!workflowPanelOpen)}
            title="Ver workflows de OpenCode"
          >
            {runningWorkflow ? (
              <>
                <Loader2 size={12} className="spin" />
                <span>{runningWorkflow.stage}</span>
              </>
            ) : (
              <>
                <Workflow size={12} />
                <span>{workflowJobs.length === 1 ? '1 workflow' : `${workflowJobs.length} workflows`}</span>
              </>
            )}
          </button>
        )}
        <button
          className={`status-jobs ${jobsOpen ? 'open' : ''}`}
          onClick={() => setJobsOpen(!jobsOpen)}
          title="Ver tareas de transcripción"
        >
          {running ? (
            <>
              <Loader2 size={12} className="spin" />
              <span>{running.stage}</span>
              <span className="status-progress">{Math.round(running.progress)}%</span>
            </>
          ) : (
            <>
              <ListVideo size={12} />
              <span>{jobs.length ? `${jobs.length} tareas` : 'Sin tareas'}</span>
            </>
          )}
        </button>
        <span className="status-version">StudyBoard App 1.1.1</span>
      </div>
    </footer>
  );
}
