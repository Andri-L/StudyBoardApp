import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { ChevronDown } from 'lucide-react';

interface Props {
  colors: readonly string[];
  value?: string;
  onChange: (color: string) => void;
  label?: string;
  compact?: boolean;
  title?: string;
}

interface PanelPosition {
  left: number;
  top: number;
}

export default function ColorPicker({ colors, value, onChange, label, compact = false, title }: Props) {
  const [open, setOpen] = useState(false);
  const [position, setPosition] = useState<PanelPosition>({ left: 0, top: 0 });
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  const place = () => {
    const rect = triggerRef.current?.getBoundingClientRect();
    if (!rect) return;
    const panelWidth = 146;
    const panelHeight = 116;
    const left = Math.max(8, Math.min(rect.left, window.innerWidth - panelWidth - 8));
    const top = rect.bottom + panelHeight + 12 > window.innerHeight
      ? Math.max(8, rect.top - panelHeight - 6)
      : rect.bottom + 6;
    setPosition({ left, top });
  };

  useEffect(() => {
    if (!open) return;
    place();
    const close = (event: MouseEvent) => {
      const target = event.target as Node;
      if (!triggerRef.current?.contains(target) && !panelRef.current?.contains(target)) {
        setOpen(false);
      }
    };
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpen(false);
    };
    const reposition = () => place();
    window.addEventListener('mousedown', close);
    window.addEventListener('keydown', onKey);
    window.addEventListener('resize', reposition);
    window.addEventListener('scroll', reposition, true);
    return () => {
      window.removeEventListener('mousedown', close);
      window.removeEventListener('keydown', onKey);
      window.removeEventListener('resize', reposition);
      window.removeEventListener('scroll', reposition, true);
    };
  }, [open]);

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        className={`color-picker-trigger ${compact ? 'compact' : ''}`}
        title={title || label || 'Elegir color'}
        onClick={(event) => {
          event.stopPropagation();
          setOpen((current) => !current);
        }}
      >
        <span className="color-picker-dot" style={{ background: value || colors[0] }} />
        {!compact && <span className="color-picker-label">{label || 'Color'}</span>}
        <ChevronDown size={12} className={open ? 'open' : ''} />
      </button>
      {open &&
        createPortal(
          <div
            ref={panelRef}
            className="color-picker-panel"
            style={{ left: position.left, top: position.top }}
            onMouseDown={(event) => event.stopPropagation()}
          >
            <span className="color-picker-title">{title || label || 'Color'}</span>
            <div className="color-picker-grid">
              {colors.map((color) => (
                <button
                  key={color}
                  type="button"
                  className={`color-swatch ${value === color ? 'active' : ''}`}
                  style={{ background: color }}
                  title={color}
                  onClick={(event) => {
                    event.stopPropagation();
                    onChange(color);
                    setOpen(false);
                  }}
                />
              ))}
            </div>
          </div>,
          document.body,
        )}
    </>
  );
}
