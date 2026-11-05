import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Background,
  BackgroundVariant,
  Controls,
  ReactFlow,
  ReactFlowProvider,
  useNodesState,
  useReactFlow,
  type Node,
  type NodeTypes,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import {
  Cpu,
  FolderOpen,
  FolderPlus,
  Layers,
  Maximize2,
  MousePointer2,
  MoveUpRight,
  Plus,
  StickyNote,
  Workflow,
  X,
} from 'lucide-react';
import { boardById, useStudyBoard } from '../store';
import type { Card, Note, WorkflowResult } from '../types';
import { compactPath } from '../api';
import { nodeTypes } from './nodes';
import ArrowLayer from './ArrowLayer';
import FolderDialog from './FolderDialog';
import QuickSettings from './QuickSettings';

type BoardNode = Node<{ card?: Card; note?: Note; result?: WorkflowResult }, 'folder' | 'video' | 'note' | 'result'>;

function cardToNode(card: Card): BoardNode {
  return {
    id: card.id,
    type: card.kind,
    position: { x: card.x, y: card.y },
    width: card.width,
    height: card.height,
    data: { card },
    style: { width: card.width, height: card.height },
  };
}

function noteToNode(note: Note): BoardNode {
  return {
    id: note.id,
    type: 'note',
    position: { x: note.x, y: note.y },
    width: note.width,
    height: note.height,
    data: { note },
    style: { width: note.width, height: note.height },
  };
}

function resultToNode(result: WorkflowResult): BoardNode {
  return {
    id: result.id,
    type: 'result',
    position: { x: result.x, y: result.y },
    width: result.width,
    height: result.height,
    data: { result },
    style: { width: result.width, height: result.height },
  };
}

interface ContextMenuState {
  x: number;
  y: number;
  flowX: number;
  flowY: number;
}

interface ArrowDraft {
  start: { x: number; y: number } | null;
  end: { x: number; y: number } | null;
}

export default function BoardView() {
  return (
    <ReactFlowProvider>
      <BoardInner />
    </ReactFlowProvider>
  );
}

function BoardInner() {
  const state = useStudyBoard((s) => s.state);
  const addCard = useStudyBoard((s) => s.addCard);
  const moveCard = useStudyBoard((s) => s.moveCard);
  const removeCard = useStudyBoard((s) => s.removeCard);
  const saveViewport = useStudyBoard((s) => s.saveViewport);
  const openFolder = useStudyBoard((s) => s.openFolder);
  const selectCard = useStudyBoard((s) => s.selectCard);
  const selectedCardId = useStudyBoard((s) => s.selectedCardId);
  const addNote = useStudyBoard((s) => s.addNote);
  const moveNote = useStudyBoard((s) => s.moveNote);
  const removeNote = useStudyBoard((s) => s.removeNote);
  const addArrow = useStudyBoard((s) => s.addArrow);
  const removeArrow = useStudyBoard((s) => s.removeArrow);
  const openWorkflowDialog = useStudyBoard((s) => s.openWorkflowDialog);
  const moveWorkflowResult = useStudyBoard((s) => s.moveWorkflowResult);
  const removeWorkflowResult = useStudyBoard((s) => s.removeWorkflowResult);
  const board = boardById(state);
  const { screenToFlowPosition, fitView } = useReactFlow();

  const [menu, setMenu] = useState<ContextMenuState | null>(null);
  const [folderDialog, setFolderDialog] = useState<{ x: number; y: number } | null>(null);
  const [arrowDraft, setArrowDraft] = useState<ArrowDraft | null>(null);
  const [selectedArrowId, setSelectedArrowId] = useState<string | null>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  const signature = useMemo(
    () =>
      board
        ? [
            ...(board.cards ?? []).map((card) => card.id),
            ...(board.notes ?? []).map((note) => note.id),
            ...(board.results ?? []).map((result) => result.id),
          ].join('|')
        : '',
    [board],
  );
  const [nodes, setNodes, onNodesChange] = useNodesState<BoardNode>([]);

  useEffect(() => {
    if (!board) {
      setNodes([]);
      return;
    }
    setNodes([
      ...(board.cards ?? []).map(cardToNode),
      ...(board.notes ?? []).map(noteToNode),
      ...(board.results ?? []).map(resultToNode),
    ]);
    // The signature only changes when cards or notes are added/removed.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [board?.id, signature]);

  useEffect(() => {
    if (!menu) return;
    const close = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as globalThis.Node)) setMenu(null);
    };
    window.addEventListener('mousedown', close);
    return () => window.removeEventListener('mousedown', close);
  }, [menu]);

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const target = event.target as HTMLElement | null;
      const editing =
        target?.tagName === 'INPUT' ||
        target?.tagName === 'TEXTAREA' ||
        Boolean(target?.isContentEditable);
      if (event.key === 'Escape') {
        setMenu(null);
        setArrowDraft(null);
        setSelectedArrowId(null);
      }
      if (!editing && (event.key === 'Delete' || event.key === 'Backspace') && selectedArrowId && board) {
        event.preventDefault();
        void removeArrow(selectedArrowId);
        setSelectedArrowId(null);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [board, selectedArrowId, removeArrow]);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      const path = event.dataTransfer.getData('application/x-studyboard-path');
      if (!path || !board) return;
      const position = screenToFlowPosition({ x: event.clientX, y: event.clientY });
      void addCard(path, position.x - 140, position.y - 60);
    },
    [addCard, board, screenToFlowPosition],
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }, []);

  if (!board) {
    return (
      <div className="board-view">
        <div className="empty-board">
          <Layers size={32} />
          <h2>No hay tableros</h2>
          <p>Crea uno desde la barra lateral.</p>
        </div>
      </div>
    );
  }

  const ready = Boolean(state?.engine.ready);
  const workspaceName = (state?.workspace || '').split(/[\\/]/).filter(Boolean).pop() || '';
  const selectedCard = (board.cards ?? []).find((card) => card.id === selectedCardId);
  const folderParent =
    selectedCard?.kind === 'folder' ? selectedCard.path : state?.workspace || '';
  const noteCount = board.notes?.length ?? 0;
  const arrowCount = board.arrows?.length ?? 0;
  const resultCount = board.results?.length ?? 0;

  const handlePaneClick = (event: React.MouseEvent) => {
    setMenu(null);
    if (arrowDraft) {
      const point = screenToFlowPosition({ x: event.clientX, y: event.clientY });
      if (!arrowDraft.start) {
        setArrowDraft({ start: point, end: point });
      } else {
        void addArrow(arrowDraft.start.x, arrowDraft.start.y, point.x, point.y);
        setArrowDraft(null);
      }
      return;
    }
    setSelectedArrowId(null);
    selectCard(null);
  };

  const handlePaneMouseMove = (event: React.MouseEvent) => {
    if (!arrowDraft?.start) return;
    const point = screenToFlowPosition({ x: event.clientX, y: event.clientY });
    setArrowDraft({ ...arrowDraft, end: point });
  };

  const menuPosition = (x: number, y: number) => {
    const width = 220;
    const height = 150;
    return {
      left: Math.min(x, window.innerWidth - width - 14),
      top: Math.min(y, window.innerHeight - height - 14),
    };
  };

  return (
    <div className="board-view">
      <header className="board-toolbar">
        <div className="board-heading">
          <span className="board-heading-icon"><Layers size={17} /></span>
          <div>
            <h1>{board.name}</h1>
            <span>
              {board.cards.length} {board.cards.length === 1 ? 'tarjeta' : 'tarjetas'}
              {noteCount > 0 ? ` · ${noteCount} ${noteCount === 1 ? 'nota' : 'notas'}` : ''}
              {arrowCount > 0 ? ` · ${arrowCount} ${arrowCount === 1 ? 'flecha' : 'flechas'}` : ''}
              {resultCount > 0 ? ` · ${resultCount} ${resultCount === 1 ? 'informe' : 'informes'}` : ''}
              {workspaceName ? ` · ${workspaceName}` : ''}
            </span>
          </div>
        </div>
        <div className="toolbar-actions">
          <QuickSettings />
          <button className="btn btn-ghost" onClick={() => void openWorkflowDialog()}>
            <Workflow size={15} />
            Ejecutar workflow
          </button>
          <button className="btn btn-ghost" onClick={() => openFolder(state?.workspace || '')}>
            <FolderOpen size={15} />
            Añadir carpeta
          </button>
          <button className="btn btn-ghost" onClick={() => void fitView({ padding: 0.22, duration: 220 })}>
            <Maximize2 size={15} />
            Ajustar
          </button>
        </div>
      </header>

      {!ready && (
        <div className="engine-banner">
          <Cpu size={14} />
          El motor de transcripción no está completo.
          {state?.engine.missing?.length ? ` Falta: ${state.engine.missing.join(', ')}.` : ''}
        </div>
      )}

      <div className="flow-wrap" onDrop={onDrop} onDragOver={onDragOver}>
        <ReactFlow
          key={board.id}
          nodes={nodes}
          edges={[]}
          nodeTypes={nodeTypes as unknown as NodeTypes}
          onNodesChange={onNodesChange}
          onNodeDragStop={(_, node) => {
            if (node.type === 'note') {
              const noteId = node.data?.note?.id ?? node.id;
              moveNote(noteId, node.position.x, node.position.y);
            } else if (node.type === 'result') {
              const resultId = node.data?.result?.id ?? node.id;
              moveWorkflowResult(resultId, node.position.x, node.position.y);
            } else {
              moveCard(board.id, node.id, node.position.x, node.position.y);
            }
          }}
          onNodesDelete={(deleted) =>
            deleted.forEach((node) => {
              if (node.type === 'note') {
                const noteId = node.data?.note?.id ?? node.id;
                void removeNote(noteId);
              } else if (node.type === 'result') {
                const resultId = node.data?.result?.id ?? node.id;
                void removeWorkflowResult(resultId);
              } else {
                void removeCard(node.id);
              }
            })
          }
          onNodeClick={(_, node) => {
            setSelectedArrowId(null);
            if (node.type === 'note' || node.type === 'result') {
              selectCard(null);
            } else {
              selectCard(node.id);
            }
          }}
          onNodeContextMenu={(event) => event.preventDefault()}
          onPaneClick={handlePaneClick}
          onPaneMouseMove={handlePaneMouseMove}
          onPaneContextMenu={(event) => {
            event.preventDefault();
            if (arrowDraft) return;
            const point = screenToFlowPosition({ x: event.clientX, y: event.clientY });
            setMenu({
              x: event.clientX,
              y: event.clientY,
              flowX: point.x,
              flowY: point.y,
            });
          }}
          onMoveEnd={(_, viewport) => saveViewport(board.id, viewport)}
          defaultViewport={board.viewport}
          minZoom={0.2}
          maxZoom={2.2}
          deleteKeyCode={['Delete']}
          multiSelectionKeyCode="Shift"
          selectionOnDrag={!arrowDraft}
          panOnDrag={!arrowDraft}
          nodesDraggable={!arrowDraft}
          elementsSelectable={!arrowDraft}
          fitViewOptions={{ padding: 0.25, maxZoom: 1.1 }}
          proOptions={{ hideAttribution: true }}
          className="studyboard-flow"
        >
          <ArrowLayer
            arrows={board.arrows ?? []}
            selectedId={selectedArrowId}
            onSelect={setSelectedArrowId}
            draft={arrowDraft}
          />
          <Background variant={BackgroundVariant.Dots} gap={18} size={1.2} color="#c9c9c9" />
          <Controls showInteractive={false} position="bottom-right" />
        </ReactFlow>

        {arrowDraft && (
          <div className="arrow-hint">
            <MoveUpRight size={15} />
            <span>
              {arrowDraft.start
                ? 'Haz clic en el punto final de la flecha'
                : 'Haz clic en el punto inicial de la flecha'}
            </span>
            <button className="icon-btn tiny" onClick={() => setArrowDraft(null)} title="Cancelar">
              <X size={13} />
            </button>
          </div>
        )}

        {board.cards.length === 0 && noteCount === 0 && resultCount === 0 && (
          <div className="empty-board-overlay">
            <div className="empty-board-card">
              <div className="empty-board-icon"><MousePointer2 size={22} /></div>
              <h2>Tu tablero está vacío</h2>
              <p>
                Arrastra una carpeta o un video desde la izquierda, o usa clic derecho
                en el tablero para crear una carpeta, una nota o una flecha.
              </p>
              <div className="empty-board-actions">
                <button className="btn btn-primary" onClick={() => openFolder(state?.workspace || '')}>
                  <Plus size={14} />
                  Explorar carpetas
                </button>
                <span className="empty-tip">Clic derecho para crear elementos</span>
              </div>
            </div>
          </div>
        )}

        {menu && (
          <div
            ref={menuRef}
            className="board-context-menu"
            style={menuPosition(menu.x, menu.y)}
            onMouseDown={(event) => event.stopPropagation()}
          >
            <button
              onClick={() => {
                setFolderDialog({ x: menu.flowX, y: menu.flowY });
                setMenu(null);
              }}
            >
              <FolderPlus size={15} />
              Nueva carpeta
            </button>
            <button
              onClick={() => {
                void addNote(menu.flowX, menu.flowY);
                setMenu(null);
              }}
            >
              <StickyNote size={15} />
              Nueva nota
            </button>
            <button
              onClick={() => {
                setArrowDraft({ start: null, end: null });
                setSelectedArrowId(null);
                setMenu(null);
              }}
            >
              <MoveUpRight size={15} />
              Nueva flecha
            </button>
          </div>
        )}
      </div>

      <div className="board-foot">
        <span title={state?.workspace}>{compactPath(state?.workspace || '', 64)}</span>
        <span>Clic derecho para crear · Arrastra para mover · Delete para quitar</span>
      </div>

      {folderDialog && (
        <FolderDialog
          parent={folderParent}
          x={folderDialog.x}
          y={folderDialog.y}
          onClose={() => setFolderDialog(null)}
        />
      )}
    </div>
  );
}
