import React from 'react';
import { createRoot } from 'react-dom/client';
import './style.css';
import App from './App';

class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  { error: Error | null }
> {
  state: { error: Error | null } = { error: null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  componentDidCatch(error: Error) {
    console.error('StudyBoard UI error:', error);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="splash">
          <div className="brand-mark large">A</div>
          <div className="splash-title">StudyBoard App encontró un problema</div>
          <div className="splash-sub">{this.state.error.message}</div>
          <button
            className="btn btn-primary"
            style={{ marginTop: 14 }}
            onClick={() => window.location.reload()}
          >
            Recargar
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

const container = document.getElementById('root');
const root = createRoot(container!);

try {
  const stored = localStorage.getItem('studyboard.theme');
  const theme =
    stored === 'light' || stored === 'dark'
      ? stored
      : window.matchMedia?.('(prefers-color-scheme: dark)').matches
        ? 'dark'
        : 'light';
  document.documentElement.dataset.theme = theme;
} catch {
  // The app still works if storage or matchMedia is unavailable.
}

root.render(
  <React.StrictMode>
    <ErrorBoundary>
      <App />
    </ErrorBoundary>
  </React.StrictMode>,
);
