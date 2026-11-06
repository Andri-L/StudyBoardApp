import { Moon, Sun } from 'lucide-react';
import { useStudyBoard } from '../store';

export default function ThemeToggle({ compact = false }: { compact?: boolean }) {
  const theme = useStudyBoard((s) => s.theme);
  const toggleTheme = useStudyBoard((s) => s.toggleTheme);
  const dark = theme === 'dark';
  return (
    <button
      className={compact ? 'icon-btn' : 'btn btn-ghost'}
      onClick={toggleTheme}
      title={dark ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}
      aria-label={dark ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro'}
    >
      {dark ? <Sun size={compact ? 16 : 15} /> : <Moon size={compact ? 16 : 15} />}
      {!compact && (dark ? 'Modo claro' : 'Modo oscuro')}
    </button>
  );
}
