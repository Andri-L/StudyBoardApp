import { memo, useEffect, useRef, useState } from 'react';
import type { Node, NodeProps } from '@xyflow/react';
import {
  AlertTriangle,
  Captions,
  Check,
  Clock3,
  ExternalLink,
  Film,
  FolderOpen,
  Loader2,
  Play,
  StickyNote,
  Trash2,
} from 'lucide-react';
import { useStudyBoard, activeJobForPath } from '../store';
import type { Card, MediaInfo, Note } from '../types';
import {
  api,
  fileName,
  formatDuration,
  friendlyError,
  parentName,
  thumbUrl,
} from '../api';
import WorkflowResultNode from './WorkflowResultNode';
import ColorPicker from './ColorPicker';
import { PASTEL_COLORS } from '../colors';

type CardNode = Node<{ card: Card }, 'folder' | 'video'>;
type NoteNodeType = Node<{ note: Note }, 'note'>;

export const FolderNode = memo(function FolderNode({ data, selected }: NodeProps<CardNode>) {
  const card = data.card;
  const openFolder = useStudyBoard((s) => s.openFolder);
  const removeCard = useStudyBoard((s) => s.removeCard);
  const showToast = useStudyBoard((s) => s.showToast);

  return (
    <div
      className={`node-card folder-node ${selected ? 'selected' : ''} ${card.color ? 'has-custom-color' : ''}`}
      style={card.color ? { background: card.color } : undefined}
      onDoubleClick={(e) => {
        e.stopPropagation();
        openFolder(card.path);
      }}
    >
      <div className="folder-node-head">
        <div className="folder-glyph">
          <FolderOpen size={20} />
        </div>
        <div className="folder-copy">
          <div className="node-title" title={card.label}>{card.label}</div>
          <div className="node-sub">{card.dateLabel || 'Carpeta de clase'}</div>
        </div>
        <button
          className="icon-btn tiny node-remove nodrag"
          title="Quitar del tablero"
          onClick={(e) => {
            e.stopPropagation();
            void removeCard(card.id);
          }}
        >
          <Trash2 size={13} />
        </button>
      </div>
      <div className="node-path" title={card.path}>{card.path}</div>
      <div className="folder-node-actions">
        <button
          className="btn btn-ghost btn-xs nodrag"
          onClick={(e) => {
            e.stopPropagation();
            openFolder(card.path);
          }}
        >
          <FolderOpen size={13} />
          Explorar
        </button>
        <button
          className="icon-btn tiny nodrag"
          title="Mostrar en Explorer"
          onClick={(e) => {
            e.stopPropagation();
            void api.reveal(card.path).catch((err) => showToast(friendlyError(err), 'error'));
          }}
        >
          <ExternalLink size={13} />
        </button>
      </div>
    </div>
  );
});

export const VideoNode = memo(function VideoNode({ data, selected }: NodeProps<CardNode>) {
  const card = data.card;
  const mediaBase = useStudyBoard((s) => s.state?.mediaBase ?? '');
  const jobs = useStudyBoard((s) => s.jobs);
  const mediaVersion = useStudyBoard((s) => s.mediaVersion);
  const engineReady = useStudyBoard((s) => s.state?.engine.ready ?? false);
  const openReview = useStudyBoard((s) => s.openReview);
  const startTranscription = useStudyBoard((s) => s.startTranscription);
  const removeCard = useStudyBoard((s) => s.removeCard);
  const showToast = useStudyBoard((s) => s.showToast);

  const [info, setInfo] = useState<MediaInfo | null>(null);
  const [infoError, setInfoError] = useState('');
  const [thumbFailed, setThumbFailed] = useState(false);
  const job = activeJobForPath(jobs, card.path);

  useEffect(() => {
    let cancelled = false;
    api
      .getMediaInfo(card.path)
      .then((value) => {
        if (!cancelled) {
          setInfo(value);
          setInfoError('');
        }
      })
      .catch((err) => {
        if (!cancelled) setInfoError(friendlyError(err));
      });
    return () => {
      cancelled = true;
    };
  }, [card.path, mediaVersion]);

  const status = job?.status ?? info?.transcript.status ?? 'none';
  const progress = job?.progress ?? info?.transcript.progress ?? 0;
  const busy = Boolean(job);
  const done = status === 'done';
  const error = status === 'error';
  const title = card.label || fileName(card.path);
  const subtitle = card.dateLabel || parentName(card.path) || 'Video';

  return (
    <div
      className={`node-card video-node ${selected ? 'selected' : ''} ${busy ? 'busy' : ''} ${card.color ? 'has-custom-color' : ''}`}
      style={card.color ? { background: card.color } : undefined}
      onDoubleClick={(e) => {
        e.stopPropagation();
        openReview(card.path);
      }}
    >
      <div className="video-thumb">
        {!thumbFailed && mediaBase ? (
          <img
            src={thumbUrl(mediaBase, card.path)}
            alt=""
            loading="lazy"
            draggable={false}
            onError={() => setThumbFailed(true)}
          />
        ) : (
          <div className="thumb-placeholder">
            <Film size={26} />
          </div>
        )}
        <div className="thumb-shade" />
        <span className="duration-badge">
          <Clock3 size={11} />
          {formatDuration(info?.duration ?? 0)}
        </span>
        <span className={`node-status ${status}`}>
          {busy ? (
            <>
              <Loader2 size={11} className="spin" />
              {Math.round(progress)}%
            </>
          ) : done ? (
            <>
              <Check size={11} />
              Transcrito
            </>
          ) : error ? (
            <>
              <AlertTriangle size={11} />
              Error
            </>
          ) : (
            'Sin transcribir'
          )}
        </span>
        <button
          className="thumb-play nodrag"
          title="Abrir video y transcripción"
          onClick={(e) => {
            e.stopPropagation();
            openReview(card.path);
          }}
        >
          <Play size={18} fill="currentColor" />
        </button>
      </div>

      <div className="video-body">
        <div className="video-head">
          <div>
            <div className="node-title" title={title}>{title}</div>
            <div className="node-sub">{subtitle}</div>
          </div>
          <button
            className="icon-btn tiny node-remove nodrag"
            title="Quitar del tablero"
            onClick={(e) => {
              e.stopPropagation();
              void removeCard(card.id);
            }}
          >
            <Trash2 size={13} />
          </button>
        </div>

        {busy && (
          <div className="job-inline">
            <div className="job-inline-top">
              <span>{job?.stage || 'Procesando'}</span>
              <span>{Math.round(progress)}%</span>
            </div>
            <div className="progress-track">
              <div className="progress-fill" style={{ width: `${Math.max(2, progress)}%` }} />
            </div>
          </div>
        )}

        {error && !busy && (
          <div className="node-error" title={info?.transcript.error || ''}>
            <AlertTriangle size={12} />
            {info?.transcript.error || infoError || 'No se pudo transcribir'}
          </div>
        )}

        <div className="video-actions">
          <button
            className="btn btn-ghost btn-xs nodrag"
            onClick={(e) => {
              e.stopPropagation();
              openReview(card.path);
            }}
          >
            <Play size={12} />
            Reproducir
          </button>
          {done ? (
            <button
              className="btn btn-primary btn-xs nodrag"
              onClick={(e) => {
                e.stopPropagation();
                openReview(card.path);
              }}
            >
              <Captions size={12} />
              Ver texto
            </button>
          ) : (
            <button
              className="btn btn-primary btn-xs nodrag"
              disabled={busy || !engineReady}
              title={!engineReady ? 'El motor de transcripción no está listo' : undefined}
              onClick={(e) => {
                e.stopPropagation();
                void startTranscription(card.path, card.id);
              }}
            >
              <Captions size={12} />
              Transcribir
            </button>
          )}
        </div>
      </div>
    </div>
  );
});

const noteColors = [...PASTEL_COLORS];

export const NoteNode = memo(function NoteNode({ data, selected }: NodeProps<NoteNodeType>) {
  const note = data.note;
  const updateNote = useStudyBoard((s) => s.updateNote);
  const removeNote = useStudyBoard((s) => s.removeNote);
  const [text, setText] = useState(note.text);
  const [color, setColor] = useState(note.color || noteColors[0]);
  const initial = useRef(true);

  useEffect(() => {
    setColor(note.color || noteColors[0]);
  }, [note.color]);

  useEffect(() => {
    if (initial.current) {
      initial.current = false;
      return;
    }
    const timer = window.setTimeout(() => {
      void updateNote(note.id, text, color);
    }, 650);
    return () => window.clearTimeout(timer);
  }, [text, color, note.id, updateNote]);

  const saveNow = () => {
    void updateNote(note.id, text, color);
  };

  return (
    <div
      className={`node-card note-node ${selected ? 'selected' : ''}`}
      style={{ background: color }}
    >
      <div className="note-head">
        <span className="note-label"><StickyNote size={12} /> Nota</span>
        <button
          className="icon-btn tiny note-remove nodrag"
          title="Eliminar nota"
          onClick={(e) => {
            e.stopPropagation();
            void removeNote(note.id);
          }}
        >
          <Trash2 size={12} />
        </button>
      </div>
      <textarea
        className="note-text nodrag nowheel"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onBlur={saveNow}
        placeholder="Escribe una idea, una duda o un paso…"
        spellCheck={false}
      />
      <div className="note-colors nodrag">
        <ColorPicker
          colors={noteColors}
          value={color}
          compact
          label="Color"
          onChange={(value) => {
            setColor(value);
            void updateNote(note.id, text, value);
          }}
        />
      </div>
    </div>
  );
});

export const nodeTypes = {
  folder: FolderNode,
  video: VideoNode,
  note: NoteNode,
  result: WorkflowResultNode,
};
