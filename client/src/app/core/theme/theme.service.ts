import { Injectable, computed, effect, signal } from '@angular/core';

export type Theme = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'theme';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  private readonly themeSignal = signal<Theme>(this.readStoredTheme());
  private readonly systemDark = signal(this.prefersDark());

  readonly theme = this.themeSignal.asReadonly();
  readonly isDark = computed(() => {
    const theme = this.themeSignal();
    return theme === 'dark' || (theme === 'system' && this.systemDark());
  });

  constructor() {
    effect(() => {
      document.documentElement.classList.toggle('dark', this.isDark());
    });

    if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
      window
        .matchMedia('(prefers-color-scheme: dark)')
        .addEventListener('change', (event) => this.systemDark.set(event.matches));
    }
  }

  setTheme(theme: Theme): void {
    this.themeSignal.set(theme);
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(STORAGE_KEY, theme);
    }
  }

  toggle(): void {
    this.setTheme(this.isDark() ? 'light' : 'dark');
  }

  private readStoredTheme(): Theme {
    if (typeof localStorage === 'undefined') {
      return 'system';
    }

    const stored = localStorage.getItem(STORAGE_KEY);
    return stored === 'light' || stored === 'dark' ? stored : 'system';
  }

  private prefersDark(): boolean {
    return (
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches
    );
  }
}
