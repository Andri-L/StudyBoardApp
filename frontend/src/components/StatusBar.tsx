import { Cpu, ListVideo, Loader2 } from 'lucide-react';
import { useStudyBoard } from '../store';
import { compactPath } from '../api';

export default function StatusBar() {
  const state = useStudyBoard((s) => s.state);
  const jobs = useStudyBoard((s) => s.jobs);
  const jobsOpen = useStudyBoard((s) => s.jobsOpen);
  const setJobsOpen = useStudyBoard((s) => s.setJobsOpen);

  const activeJobs = jobs.filter((j) =>
    ['queued', 'extracting', 'transcribing', 'saving'].includes(j.status),
  );
  const running = activeJobs[0];
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
        <span className="status-version">StudyBoard App 1.0.0</span>
      </div>
    </footer>
  );
}
