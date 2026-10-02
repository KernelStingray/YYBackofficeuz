import { useThemeStore } from '../../store/useThemeStore';
import type { Mode, ColorTheme } from '../../store/useThemeStore';
import { Sun, Moon, Laptop, Palette } from 'lucide-react';

export const ThemeToggle = () => {
  const { mode, color, setMode, setColor } = useThemeStore();

  const modes: { value: Mode; label: string; icon: typeof Sun }[] = [
    { value: 'light', label: 'Светлая', icon: Sun },
    { value: 'dark', label: 'Тёмная', icon: Moon },
    { value: 'system', label: 'Система', icon: Laptop },
  ];

  const colors: { value: ColorTheme; label: string; className: string }[] = [
    { value: 'default', label: 'Default', className: 'bg-blue-600' },
    { value: 'emerald', label: 'Emerald', className: 'bg-emerald-600' },
    { value: 'rose', label: 'Rose', className: 'bg-rose-600' },
  ];

  return (
    <div className="flex items-center gap-4">
      {/* Mode Selector */}
      <div className="flex items-center gap-1 p-1 bg-muted/50 rounded-lg border border-border shadow-inner">
        {modes.map(({ value, label, icon: Icon }) => {
          const isActive = mode === value;
          return (
            <button
              key={value}
              onClick={() => setMode(value)}
              title={`Режим: ${label}`}
              className={`flex items-center justify-center p-1.5 rounded-md transition-all duration-300 ${
                isActive
                  ? 'bg-background text-foreground shadow-sm scale-100'
                  : 'text-muted-foreground hover:text-foreground hover:bg-muted/80 scale-95 hover:scale-100'
              }`}
            >
              <Icon className="h-4 w-4" />
              <span className="sr-only">{label}</span>
            </button>
          );
        })}
      </div>

      {/* Color Selector */}
      <div className="flex items-center gap-2">
        <Palette className="h-4 w-4 text-muted-foreground" />
        <div className="flex items-center gap-1.5">
          {colors.map(({ value, label, className }) => {
            const isActive = color === value;
            return (
              <button
                key={value}
                onClick={() => setColor(value)}
                title={`Цвет: ${label}`}
                className={`h-5 w-5 rounded-full transition-all duration-300 ${className} ${
                  isActive ? 'ring-2 ring-offset-2 ring-offset-background ring-primary scale-110' : 'hover:scale-110 opacity-70 hover:opacity-100'
                }`}
              >
                <span className="sr-only">{label}</span>
              </button>
            );
          })}
        </div>
      </div>
    </div>
  );
};
