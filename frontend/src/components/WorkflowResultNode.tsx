import type { Node, NodeProps } from '@xyflow/react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import type { Components } from 'react-markdown';
import {
  ClipboardCopy,
  ExternalLink,
  FileText,
  Trash2,
  Workflow,
} from 'lucide-react';
import { useStudyBoard } from '../store';
import type { WorkflowResult } from '../types';
import { api, fileName, friendlyError } from '../api';

type WorkflowResultNodeType = Node<{ result: WorkflowResult }, 'result'>;

function safeWebUrl(raw?: string): string | null {
  if (!raw) return null;
  try {
    const parsed = new URL(raw);
    if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return null;
    if (!parsed.hostname || parsed.username || parsed.password) return null;
    return parsed.toString();
  } catch {
    return null;
  }
}

export default function WorkflowResultNode({ data, selected }: NodeProps<WorkflowResultNodeType>) {
  const result = data.result;
  const removeWorkflowResult = useStudyBoard((s) => s.removeWorkflowResult);
  const showToast = useStudyBoard((s) => s.showToast);

  const copy = async () => {
    try {
      await api.copyText(result.content);
      showToast('Informe copiado al portapapeles', 'success');
    } catch (err) {
      showToast(friendlyError(err), 'error');
    }
  };

  const openLink = async (raw?: string) => {
    const url = safeWebUrl(raw);
    if (!url) {
      showToast('Enlace bloqueado: solo se permiten enlaces http o https', 'error');
      return;
    }
    try {
      await api.openUrl(url);
    } catch (err) {
      showToast(friendlyError(err), 'error');
    }
  };

  const components: Components = {
    a: ({ href, children }) => {
      const url = safeWebUrl(href);
      if (!url) return <span className="md-link-blocked">{children}</span>;
      return (
        <button
          type="button"
          className="md-link"
          title={url}
          onClick={(event) => {
            event.preventDefault();
            event.stopPropagation();
            void openLink(url);
          }}
        >
          {children}
        </button>
      );
    },
    img: ({ src, alt }) => (
      <span className="md-image-blocked" title={typeof src === 'string' ? src : undefined}>
        {alt ? `[imagen bloqueada: ${alt}]` : '[imagen externa bloqueada]'}
      </span>
    ),
  };

  return (
    <div className={`node-card workflow-result-node ${selected ? 'selected' : ''}`}>
      <header className="workflow-result-head">
        <span className="workflow-result-icon"><Workflow size={15} /></span>
        <div className="workflow-result-title">
          <strong>{result.workflowName}</strong>
          <span title={result.videoPath}>{fileName(result.videoName)} · {result.className}</span>
        </div>
        <button
          className="icon-btn tiny node-remove nodrag"
          title="Quitar del tablero (conserva el archivo)"
          onClick={(e) => {
            e.stopPropagation();
            void removeWorkflowResult(result.id);
          }}
        >
          <Trash2 size={13} />
        </button>
      </header>

      <div className="workflow-result-body nodrag nowheel">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          skipHtml
          components={components}
          urlTransform={(url) => safeWebUrl(url) ?? ''}
        >
          {result.content}
        </ReactMarkdown>
      </div>

      <footer className="workflow-result-foot">
        <span className="workflow-result-file" title={result.filePath || result.videoPath}>
          <FileText size={12} />
          {result.filePath ? fileName(result.filePath) : 'Markdown no guardado'}
        </span>
        <div className="workflow-result-actions">
          <button className="icon-btn tiny nodrag" title="Copiar informe" onClick={() => void copy()}>
            <ClipboardCopy size={13} />
          </button>
          <button
            className="icon-btn tiny nodrag"
            title="Mostrar archivo en Explorer"
            onClick={() => {
              const target = result.filePath || result.videoPath;
              void api.reveal(target).catch((err) => showToast(friendlyError(err), 'error'));
            }}
          >
            <ExternalLink size={13} />
          </button>
        </div>
      </footer>
    </div>
  );
}
