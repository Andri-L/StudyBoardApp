import { ViewportPortal } from '@xyflow/react';
import type { Arrow } from '../types';

interface Props {
  arrows: Arrow[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  draft: { start: { x: number; y: number } | null; end: { x: number; y: number } | null } | null;
}

export default function ArrowLayer({ arrows, selectedId, onSelect, draft }: Props) {
  const start = draft?.start;
  const end = draft?.end;

  return (
    <ViewportPortal>
      <svg className="board-arrow-layer" aria-hidden="true">
        <defs>
          <marker id="studyboard-arrowhead" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M 0 0 L 10 5 L 0 10 z" fill="#0067c0" />
          </marker>
          <marker id="studyboard-arrowhead-selected" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
            <path d="M 0 0 L 10 5 L 0 10 z" fill="#c42b1c" />
          </marker>
        </defs>

        {arrows.map((arrow) => {
          const selected = arrow.id === selectedId;
          return (
            <g key={arrow.id} className={`board-arrow ${selected ? 'selected' : ''}`}>
              <line
                x1={arrow.x1}
                y1={arrow.y1}
                x2={arrow.x2}
                y2={arrow.y2}
                stroke="transparent"
                strokeWidth={18}
                pointerEvents="stroke"
                style={{ cursor: 'pointer' }}
                onClick={(e) => {
                  e.stopPropagation();
                  onSelect(arrow.id);
                }}
              />
              <line
                x1={arrow.x1}
                y1={arrow.y1}
                x2={arrow.x2}
                y2={arrow.y2}
                stroke={selected ? '#c42b1c' : arrow.color || '#0067c0'}
                strokeWidth={selected ? 3 : 2.4}
                strokeLinecap="round"
                markerEnd={selected ? 'url(#studyboard-arrowhead-selected)' : 'url(#studyboard-arrowhead)'}
                pointerEvents="none"
              />
            </g>
          );
        })}

        {start && end && (
          <line
            x1={start.x}
            y1={start.y}
            x2={end.x}
            y2={end.y}
            stroke="#0067c0"
            strokeWidth={2}
            strokeDasharray="6 5"
            strokeLinecap="round"
            markerEnd="url(#studyboard-arrowhead)"
            pointerEvents="none"
          />
        )}
      </svg>
    </ViewportPortal>
  );
}
