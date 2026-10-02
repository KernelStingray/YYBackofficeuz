import { create } from 'zustand';

export type Mode = 'dark' | 'light' | 'system';
export type ColorTheme = 'default' | 'emerald' | 'rose';

interface ThemeState {
  mode: Mode;
  color: ColorTheme;
  setMode: (mode: Mode) => void;
  setColor: (color: ColorTheme) => void;
}

const getInitialMode = (): Mode => {
  const savedMode = localStorage.getItem('theme_mode') as Mode | null;
  if (savedMode && ['dark', 'light', 'system'].includes(savedMode)) {
    return savedMode;
  }
  // Default to system
  return 'system';
};

const getInitialColor = (): ColorTheme => {
  const savedColor = localStorage.getItem('theme_color') as ColorTheme | null;
  if (savedColor && ['default', 'emerald', 'rose'].includes(savedColor)) {
    return savedColor;
  }
  return 'default';
};

const applyTheme = (mode: Mode, color: ColorTheme) => {
  const root = document.documentElement;
  
  // Clean up old classes
  root.classList.remove('light', 'dark', 'theme-emerald', 'theme-rose');

  // Apply Mode
  if (mode === 'system') {
    const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    root.classList.add(systemTheme);
  } else {
    root.classList.add(mode);
  }

  // Apply Color Variant
  if (color !== 'default') {
    root.classList.add(`theme-${color}`);
  }
};

// Initial application
applyTheme(getInitialMode(), getInitialColor());

// Listen for system theme changes if set to system
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  const savedMode = localStorage.getItem('theme_mode') as Mode | null;
  if (!savedMode || savedMode === 'system') {
    applyTheme('system', getInitialColor());
  }
});

export const useThemeStore = create<ThemeState>((set, get) => ({
  mode: getInitialMode(),
  color: getInitialColor(),

  setMode: (mode: Mode) => {
    localStorage.setItem('theme_mode', mode);
    applyTheme(mode, get().color);
    set({ mode });
  },

  setColor: (color: ColorTheme) => {
    localStorage.setItem('theme_color', color);
    applyTheme(get().mode, color);
    set({ color });
  }
}));
