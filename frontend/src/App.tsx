import { useEffect } from 'react';
import { useStudyBoard } from './store';
import { onWailsEvent } from './api';
import Sidebar from './components/Sidebar';
import BoardView from './components/BoardView';
import ReviewView from './components/ReviewView';
import StatusBar from './components/StatusBar';
import Welcome from './components/Welcome';
import FolderDrawer from './components/FolderDrawer';
import JobsPanel from './components/JobsPanel';
import Toast from './components/Toast';
import WorkflowDialog from './components/WorkflowDialog';
import WorkflowJobsPanel from './components/WorkflowJobsPanel';

export default function App() {
  const loading = useStudyBoard((s) => s.loading);
  const state = useStudyBoard((s) => s.state);
  const view = useStudyBoard((s) => s.view);
  const reviewPath = useStudyBoard((s) => s.reviewPath);
  const init = useStudyBoard((s) => s.init);
  const jobUpdated = useStudyBoard((s) => s.jobUpdated);
  const transcriptReady = useStudyBoard((s) => s.transcriptReady);
  const setAppState = useStudyBoard((s) => s.setAppState);
  const modelProgress = useStudyBoard((s) => s.modelProgress);
  const workflowJobUpdated = useStudyBoard((s) => s.workflowJobUpdated);
  const closeFolder = useStudyBoard((s) => s.closeFolder);
  const closeReview = useStudyBoard((s) => s.closeReview);
  const jobsOpen = useStudyBoard((s) => s.jobsOpen);
  const setJobsOpen = useStudyBoard((s) => s.setJobsOpen);
  const workflowPanelOpen = useStudyBoard((s) => s.workflowPanelOpen);
  const theme = useStudyBoard((s) => s.theme);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
  }, [theme]);

  useEffect(() => {
    void init();
    const offJob = onWailsEvent('job:update', (job) => jobUpdated(job));
    const offReady = onWailsEvent('transcript:ready', (path) => transcriptReady(String(path)));
    const offEngine = onWailsEvent('engine:update', (state) => setAppState(state));
    const offModel = onWailsEvent('model:progress', (download) => modelProgress(download));
    const offWorkflow = onWailsEvent('workflow:update', (job) => workflowJobUpdated(job));
    const offWorkflowResult = onWailsEvent('workflow:result', (state) => setAppState(state));
    return () => {
      offJob();
      offReady();
      offEngine();
      offModel();
      offWorkflow();
      offWorkflowResult();
    };
  }, [init, jobUpdated, transcriptReady, setAppState, modelProgress, workflowJobUpdated]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return;
      if (jobsOpen) {
        setJobsOpen(false);
      } else if (view === 'review') {
        closeReview();
      } else {
        closeFolder();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [closeFolder, closeReview, jobsOpen, setJobsOpen, view]);

  if (loading) {
    return (
      <div className="splash">
        <div className="brand-mark large">A</div>
        <div className="splash-title">StudyBoard App</div>
        <div className="splash-sub">Preparando tu espacio de clases…</div>
      </div>
    );
  }

  if (!state?.workspace) {
    return (
      <>
        <Welcome />
        <Toast />
      </>
    );
  }

  return (
    <div className="app-shell">
      <Sidebar />
      <div className="workspace">
        {view === 'review' && reviewPath ? (
          <ReviewView path={reviewPath} />
        ) : (
          <BoardView />
        )}
        <StatusBar />
      </div>
      <FolderDrawer />
      <WorkflowDialog />
      {jobsOpen && <JobsPanel />}
      {workflowPanelOpen && <WorkflowJobsPanel />}
      <Toast />
    </div>
  );
}
