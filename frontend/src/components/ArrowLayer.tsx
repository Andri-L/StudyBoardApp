import { useState } from 'react';
import { ViewportPortal, useReactFlow } from '@xyflow/react';
import type { Arrow } from '../types';

interface Props {
  arrows: Arrow[];
  selectedId: string | null;
  onSelect: (id: string | null) => void;
  onCommit: (id: string, x1: number, y1: number, x2: number, y2: number) => void;
  draft: { start: { x: number; y: number } | null; end: { x: number; y: number } | null } | null;
}

interface DragState {
  arrow: Arrow;
  mode: 'move' | 'start' | 'end';
  origin: { x: number; y: number };
  moved: boolean;
  preview: Arrow;
}

function arrowPath(arrow: Arrow, draftType: 'straight' | 's' = 'straight'): { path: string; endAngle: number } {
  const type = arrow.type || draftType;
  if (type === 's') {
    const dx = arrow.x2 - arrow.x1;
    const dy = arrow.y2 - arrow.y1;
    let c1x: number;
    let c1y: number;
    let c2x: number;
    let c2y: number;
    if (Math.abs(dx) >= Math.abs(dy)) {
      c1x = arrow.x1 + dx * 0.55;
      c1y = arrow.y1;
      c2x = arrow.x2 - dx * 0.55;
      c2y = arrow.y2;
    } else {
      c1x = arrow.x1;
      c1y = arrow.y1 + dy * 0.55;
      c2x = arrow.x2;
      c2y = arrow.y2 - dy * 0.55;
    }
    return {
      path: `M ${arrow.x1} ${arrow.y1} C ${c1x} ${c1y}, ${c2x} ${c2y}, ${arrow.x2} ${arrow.y2}`,
      endAngle: Math.atan2(arrow.y2 - c2y, arrow.x2 - c2x),
    };
  }
  return {
    path: `M ${arrow.x1} ${arrow.y1} L ${arrow.x2} ${arrow.y2}`,
    endAngle: Math.atan2(arrow.y2 - arrow.y1, arrow.x2 - arrow.x1),
  };
}

function arrowHead(x: number, y: number, angle: number, size = 9): string {
  const left = angle + Math.PI * 0.82;
  const right = angle - Math.PI * 0.82;
  return [
    `${x},${y}`,
    `${x + Math.cos(left) * size},${y + Math.sin(left) * size}`,
    `${x + Math.cos(right) * size},${y + Math.sin(right) * size}`,
  ].join(' ');
}

function previewArrow(base: Arrow, point: { x: number; y: number }, mode: DragState['mode'], origin: { x: number; y: number }): Arrow {
  if (mode === 'move') {
    const dx = point.x - origin.x;
    const dy = point.y - origin.y;
    return { ...base, x1: base.x1 + dx, y1: base.y1 + dy, x2: base.x2 + dx, y2: base.y2 + dy };
  }
  if (mode === 'start') {
    return { ...base, x1: point.x, y1: point.y };
  }
  return { ...base, x2: point.x, y2: point.y };
}

export default function ArrowLayer({ arrows, selectedId, onSelect, onCommit, draft }: Props) {
  const { screenToFlowPosition } = useReactFlow();
  const [drag, setDrag] = useState<DragState | null>(null);
  const [hover, setHover] = useState<string | null>(null);

  const beginDrag = (event: React.PointerEvent, arrow: Arrow, mode: DragState['mode']) => {
    event.stopPropagation();
    event.preventDefault();
    (event.currentTarget as Element).setPointerCapture(event.pointerId);
    const point = screenToFlowPosition({ x: event.clientX, y: event.clientY });
    setDrag({ arrow, mode, origin: point, moved: false, preview: { ...arrow } });
    onSelect(arrow.id);
  };

  const updateDrag = (event: React.PointerEvent) => {
    if (!drag) return;
    const point = screenToFlowPosition({ x: event.clientX, y: event.clientY });
    const moved =
      drag.moved ||
      Math.abs(point.x - drag.origin.x) > 2 ||
      Math.abs(point.y - drag.origin.y) > 2;
    setDrag({ ...drag, moved, preview: previewArrow(drag.arrow, point, drag.mode, drag.origin) });
  };

  const finishDrag = (event: React.PointerEvent) => {
    if (!drag) return;
    event.stopPropagation();
    const final = drag.preview;
    const moved = drag.moved;
    setDrag(null);
    if (moved) {
      onCommit(final.id, final.x1, final.y1, final.x2, final.y2);
    }
  };

  const draftStart = draft?.start;
  const draftEnd = draft?.end;

  return (
    <ViewportPortal>
      <svg className="board-arrow-layer" aria-hidden="true">
        {arrows.map((base) => {
          const arrow = drag?.arrow.id === base.id ? drag.preview : base;
          const selected = arrow.id === selectedId;
          const hovered = arrow.id === hover;
          const color = selected ? '#c42b1c' : arrow.color || '#0067c0';
          const { path, endAngle } = arrowPath(arrow);
          return (
            <g key={arrow.id} className={`board-arrow ${selected ? 'selected' : ''}`}>
              <path
                d={path}
                stroke="transparent"
                strokeWidth={20}
                fill="none"
                pointerEvents="stroke"
                style={{ cursor: drag ? 'grabbing' : 'grab' }}
                onPointerDown={(event) => beginDrag(event, arrow, 'move')}
                onPointerMove={updateDrag}
                onPointerUp={finishDrag}
                onPointerCancel={finishDrag}
                onPointerEnter={() => setHover(arrow.id)}
                onPointerLeave={() => setHover(null)}
                onClick={(event) => {
                  event.stopPropagation();
                  onSelect(arrow.id);
                }}
              />
              <path
                d={path}
                stroke={color}
                strokeWidth={selected || hovered ? 3.2 : 2.4}
                fill="none"
                strokeLinecap="round"
                pointerEvents="none"
              />
              <polygon points={arrowHead(arrow.x2, arrow.y2, endAngle, selected ? 11 : 9)} fill={color} pointerEvents="none" />
              {selected && (
                <>
                  <circle
                    cx={arrow.x1}
                    cy={arrow.y1}
                    r={6}
                    className="arrow-handle"
                    onPointerDown={(event) => beginDrag(event, arrow, 'start')}
                    onPointerMove={updateDrag}
                    onPointerUp={finishDrag}
                    onPointerCancel={finishDrag}
                  />
                  <circle
                    cx={arrow.x2}
                    cy={arrow.y2}
                    r={6}
                    className="arrow-handle"
                    onPointerDown={(event) => beginDrag(event, arrow, 'end')}
                    onPointerMove={updateDrag}
                    onPointerUp={finishDrag}
                    onPointerCancel={finishDrag}
                  />
                </>
              )}
            </g>
          );
        })}

        {draftStart && draftEnd && (
          <g pointerEvents="none">
            <path
              d={arrowPath(
                {
                  id: 'draft',
                  x1: draftStart.x,
                  y1: draftStart.y,
                  x2: draftEnd.x,
                  y2: draftEnd.y,
                  type: 'straight',
                  color: '#0067c0',
                  createdAt: 0,
                },
              ).path}
              stroke="#0067c0"
              strokeWidth={2}
              strokeDasharray="6 5"
              fill="none"
              strokeLinecap="round"
            />
            <polygon
              points={arrowHead(
                draftEnd.x,
                draftEnd.y,
                Math.atan2(draftEnd.y - draftStart.y, draftEnd.x - draftStart.x),
              )}
              fill="#0067c0"
            />
          </g>
        )}
      </svg>
    </ViewportPortal>
  );
}
