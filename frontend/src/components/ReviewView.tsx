import { useEffect, useMemo, useRef, useState } from 'react';
import {
  AlertTriangle,
  ArrowLeft,
  Camera,
  Captions,
  Check,
  ClipboardCopy,
  Download,
  ExternalLink,
  FileText,
  Loader2,
  Play,
  Search,
  Sparkles,
  X,
} from 'lucide-react';
import { activeJobForPath, boardById, useStudyBoard } from '../store';
import type { MediaInfo, Segment, Transcript } from '../types';
import {
  api,
  fileName,
  formatDuration,
  formatTimestamp,
  formatTimestampPrecise,
  friendlyError,
  mediaUrl,
} from '../api';

interface Props {
  path: string;
}

export default function ReviewView({ path }: Props) {
  const state = useStudyBoard((s) => s.state);
  const jobs = useStudyBoard((s) => s.jobs);
  const mediaVersion = useStudyBoard((s) => s.mediaVersion);
  const closeReview = useStudyBoard((s) => s.closeReview);
  const startTranscription = useStudyBoard((s) => s.startTranscription);
  const showToast = useStudyBoard((s) => s.showToast);
  const board = boardById(state);

  const [transcript, setTranscript] = useState<Transcript | null>(null);
  const [info, setInfo] = useState<MediaInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [query, setQuery] = useState('');
  const [currentTime, setCurrentTime] = useState(0);
  const [follow, setFollow] = useState(false);
  const [videoError, setVideoError] = useState(false);
  const [frame, setFrame] = useState<{ src: string; time: number } | null>(null);
  const [capturing, setCapturing] = useState(false);
  const videoRef = useRef<HTMLVideoElement | null>(null);

  const job = activeJobForPath(jobs, path);
  const active = useMemo(
    () =>
      transcript?.segments.findIndex(
        (s) => currentTime >= s.start && currentTime < Math.max(s.end, s.start + 0.4),
      ) ?? -1,
    [transcript, currentTime],
  );

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError('');
    setVideoError(false);
    setFrame(null);
    setQuery('');
    Promise.all([
      api.getTranscript(path).catch(() => null),
      api.getMediaInfo(path).catch(() => null),
    ])
      .then(([tr, media]) => {
        if (cancelled) return;
        setTranscript(tr);
        setInfo(media);
      })
      .catch((err) => {
        if (!cancelled) setError(friendlyError(err));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [path, mediaVersion]);

  useEffect(() => {
    if (!follow || active < 0) return;
    const el = document.querySelector<HTMLElement>(`.segment-row[data-index="${active}"]`);
    el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
  }, [active, follow]);

  const filtered = useMemo(() => {
    if (!transcript) return [];
    const q = query.trim().toLowerCase();
    return transcript.segments
      .map((s, index) => ({ ...s, index }))
      .filter((s) => !q || s.text.toLowerCase().includes(q));
  }, [transcript, query]);

  const card = board?.cards.find((c) => c.path.toLowerCase() === path.toLowerCase());

  const seek = (seconds: number) => {
    const video = videoRef.current;
    if (video) {
      video.currentTime = Math.max(0, seconds);
      void video.play().catch(() => undefined);
    }
  };

  const transcribe = () => {
    void startTranscription(path, card?.id || '');
  };

  const exportFile = async (format: 'md' | 'srt') => {
    try {
      const saved = await api.exportTranscript(path, format);
      if (saved) showToast(`Guardado en ${saved}`, 'success');
    } catch (err) {
      showToast(friendlyError(err), 'error');
    }
  };

  const copyAll = async () => {
    if (!transcript) return;
    const text = transcript.segments
      .map((s) => `[${formatTimestamp(s.start)}] ${s.text}`)
      .join('\n');
    try {
      await api.copyText(text);
      showToast('Transcripción copiada al portapapeles', 'success');
    } catch (err) {
      showToast(friendlyError(err), 'error');
    }
  };

  const capture = async (segment: Segment) => {
    setCapturing(true);
    try {
      const exactStart = segment.words?.[0]?.start ?? segment.start;
      const src = await api.captureFrame(path, exactStart + 0.2);
      setFrame({ src, time: exactStart });
    } catch (err) {
      showToast(friendlyError(err), 'error');
    } finally {
      setCapturing(false);
    }
  };

  const ready = Boolean(state?.engine.ready);
  const progress = job?.progress ?? 0;
  const status = job?.status ?? info?.transcript.status ?? (transcript ? 'done' : 'none');
  const duration = transcript?.duration || info?.duration || 0;

  return (
    <div className="review-view">
      <header className="review-toolbar">
        <button className="icon-btn" onClick={closeReview} title="Volver al tablero">
          <ArrowLeft size={17} />
        </button>
        <div className="review-title">
          <h1 title={path}>{fileName(path)}</h1>
          <span>
            {formatDuration(duration)}
            {transcript ? ` · ${transcript.segments.length} segmentos` : ''}
            {transcript ? ` · ${transcript.language}` : ''}
          </span>
        </div>
        <div className="review-actions">
          {transcript && (
            <>
              <span className={`review-status ${status}`}>
                {status === 'done' ? <Check size={12} /> : <Loader2 size={12} />}
                {status === 'done' ? 'Transcrito' : status === 'error' ? 'Error' : 'Procesando'}
              </span>
              <button className="btn btn-ghost btn-sm" onClick={() => void exportFile('md')}>
                <FileText size={14} />
                Markdown
              </button>
              <button className="btn btn-ghost btn-sm" onClick={() => void exportFile('srt')}>
                <Download size={14} />
                SRT
              </button>
              <button className="btn btn-ghost btn-sm" onClick={() => void copyAll()}>
                <ClipboardCopy size={14} />
                Copiar
              </button>
            </>
          )}
          <button
            className="icon-btn"
            title="Mostrar en Explorer"
            onClick={() => void api.reveal(path).catch((err) => showToast(friendlyError(err), 'error'))}
          >
            <ExternalLink size={16} />
          </button>
        </div>
      </header>

      <div className="review-grid">
        <section className="review-player">
          <div className="video-stage">
            <video
              ref={videoRef}
              src={state?.mediaBase ? mediaUrl(state.mediaBase, path) : undefined}
              controls
              preload="metadata"
              onTimeUpdate={(e) => setCurrentTime(e.currentTarget.currentTime)}
              onError={() => setVideoError(true)}
            />
            {videoError && (
              <div className="video-error">
                <AlertTriangle size={22} />
                <strong>Este formato no se reproduce dentro de StudyBoard App</strong>
                <p>Prueba con un MP4, o ábrelo con tu reproductor de Windows.</p>
                <button className="btn btn-primary btn-sm" onClick={() => void api.openExternal(path)}>
                  <Play size={14} />
                  Abrir con Windows
                </button>
              </div>
            )}
          </div>

          <div className="player-meta">
            <div>
              <span className="meta-label">Archivo</span>
              <span className="meta-value" title={path}>{path}</span>
            </div>
            <div className="meta-row">
              <div>
                <span className="meta-label">Duración</span>
                <span className="meta-value">{formatDuration(duration)}</span>
              </div>
              <div>
                <span className="meta-label">Modelo</span>
                <span className="meta-value">{transcript?.model || state?.engine.modelName || '—'}</span>
              </div>
              <div>
                <span className="meta-label">Idioma</span>
                <span className="meta-value">{transcript?.language || 'es'}</span>
              </div>
            </div>
          </div>

          {!transcript && !loading && (
            <div className="review-empty-action">
              <div className="empty-action-icon"><Sparkles size={20} /></div>
              <div>
                <strong>{job ? job.stage : 'Este video todavía no tiene transcripción'}</strong>
                <p>
                  {job
                    ? 'Whisper está procesando el audio. Puedes seguir usando el tablero.'
                    : 'Genera texto con timestamps para navegar la clase y exportarla a tu LLM.'}
                </p>
                {job ? (
                  <div className="job-inline wide">
                    <div className="job-inline-top">
                      <span>{job.stage}</span>
                      <span>{Math.round(progress)}%</span>
                    </div>
                    <div className="progress-track">
                      <div className="progress-fill" style={{ width: `${Math.max(2, progress)}%` }} />
                    </div>
                  </div>
                ) : (
                  <button className="btn btn-primary btn-sm" disabled={!ready} onClick={transcribe}>
                    <Captions size={14} />
                    Transcribir ahora
                  </button>
                )}
                {!ready && <span className="inline-warning">El motor no está listo.</span>}
              </div>
            </div>
          )}

          {status === 'error' && (
            <div className="review-error">
              <AlertTriangle size={16} />
              <div>
                <strong>La transcripción falló</strong>
                <p>{info?.transcript.error || 'Revisa el video e inténtalo de nuevo.'}</p>
                <button className="btn btn-ghost btn-sm" onClick={transcribe}>Reintentar</button>
              </div>
            </div>
          )}

          <div className="player-tips">
            <Camera size={14} />
            <span>
              Usa el botón de cámara en cualquier línea para ver el frame de ese momento.
              Más adelante el LLM podrá usar esos frames como contexto visual.
            </span>
          </div>
        </section>

        <section className="transcript-panel">
          <div className="transcript-head">
            <div>
              <h2>Transcripción</h2>
              <span>
                {transcript
                  ? `${transcript.segments.length} líneas · ${formatTimestamp(duration)}`
                  : 'Sin texto todavía'}
              </span>
            </div>
            <div className="transcript-tools">
              {transcript && (
                <label className="follow-toggle" title="Desplazar la transcripción con el video">
                  <input type="checkbox" checked={follow} onChange={(e) => setFollow(e.target.checked)} />
                  Seguir
                </label>
              )}
              <div className="search-box">
                <Search size={14} />
                <input
                  placeholder="Buscar en la clase…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  disabled={!transcript}
                />
                {query && (
                  <button className="icon-btn tiny" onClick={() => setQuery('')} title="Limpiar">
                    <X size={12} />
                  </button>
                )}
              </div>
            </div>
          </div>

          {loading ? (
            <div className="transcript-empty">
              <Loader2 size={20} className="spin" />
              Cargando transcripción…
            </div>
          ) : error ? (
            <div className="transcript-empty error">{error}</div>
          ) : transcript ? (
            <div className="segment-list">
              {filtered.length === 0 ? (
                <div className="transcript-empty">No hay líneas que coincidan con “{query}”.</div>
              ) : (
                filtered.map((segment) => (
                  <div
                    key={`${segment.index}-${segment.start}`}
                    className={`segment-row ${segment.index === active ? 'active' : ''}`}
                    data-index={segment.index}
                  >
                    <button
                      className="segment-time"
                      title={`Ir a ${formatTimestampPrecise(segment.start)}`}
                      onClick={() => seek(segment.start)}
                    >
                      {formatTimestamp(segment.start)}
                    </button>
                    <p>{segment.text}</p>
                    <button
                      className="segment-camera"
                      title="Ver el frame de este momento"
                      disabled={capturing}
                      onClick={() => void capture(segment)}
                    >
                      <Camera size={13} />
                    </button>
                  </div>
                ))
              )}
            </div>
          ) : (
            <div className="transcript-empty">
              <Captions size={24} />
              <strong>Sin transcripción</strong>
              <span>Transcribe el video para navegar la clase por texto.</span>
              {ready && !job && (
                <button className="btn btn-primary btn-sm" onClick={transcribe}>
                  <Captions size={14} />
                  Transcribir
                </button>
              )}
            </div>
          )}

          {transcript && (
            <footer className="transcript-foot">
              <span>Click en un timestamp para saltar</span>
              <span>{transcript.model}</span>
            </footer>
          )}
        </section>
      </div>

      {frame && (
        <div className="frame-backdrop" onMouseDown={() => setFrame(null)}>
          <div className="frame-modal" onMouseDown={(e) => e.stopPropagation()}>
            <div className="frame-head">
              <div>
                <strong>Frame de la clase</strong>
                <span>{formatTimestampPrecise(frame.time)}</span>
              </div>
              <button className="icon-btn" onClick={() => setFrame(null)}>
                <X size={17} />
              </button>
            </div>
            <img src={frame.src} alt={`Frame en ${formatTimestamp(frame.time)}`} />
            <div className="frame-foot">
              <span>Este frame corresponde al momento seleccionado del video.</span>
              <button className="btn btn-ghost btn-sm" onClick={() => seek(frame.time)}>
                <Play size={13} />
                Ir a este momento
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
