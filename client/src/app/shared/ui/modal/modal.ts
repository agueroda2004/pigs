import { Component, computed, input, output } from '@angular/core';

export type ModalSize = 'sm' | 'md' | 'lg' | 'xl';

const SIZE_CLASSES: Record<ModalSize, string> = {
  sm: 'sm:max-w-sm',
  md: 'sm:max-w-md',
  lg: 'sm:max-w-lg',
  xl: 'sm:max-w-2xl',
};

@Component({
  selector: 'app-modal',
  host: {
    '(document:keydown.escape)': 'onEscape()',
  },
  template: `
    @if (open()) {
      <div class="fixed inset-0 z-50 flex items-end justify-center">
        <div
          class="absolute inset-0 animate-fade-in bg-black/50"
          (click)="closed.emit()"
          aria-hidden="true"
        ></div>

        <div
          role="dialog"
          aria-modal="true"
          class="relative z-10 flex max-h-[90vh] w-full animate-sheet-up flex-col rounded-t-2xl border border-b-0 border-border bg-card p-6 pb-[calc(1.5rem+env(safe-area-inset-bottom))] shadow-lg"
          [class]="sizeClass()"
        >
          <div class="mb-4 flex shrink-0 justify-center">
            <span class="h-1.5 w-12 rounded-full bg-border" aria-hidden="true"></span>
          </div>

          <div class="mb-5 flex shrink-0 items-center justify-between gap-4">
            <h2 class="text-lg font-semibold tracking-tight">{{ title() }}</h2>
            <button
              type="button"
              (click)="closed.emit()"
              aria-label="Cerrar"
              class="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="h-4 w-4"
              >
                <path d="M18 6 6 18M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto pr-1">
            <ng-content />
          </div>
        </div>
      </div>
    }
  `,
})
export class Modal {
  readonly open = input(false);
  readonly title = input('');
  readonly size = input<ModalSize>('md');
  readonly closed = output<void>();

  protected readonly sizeClass = computed(() => SIZE_CLASSES[this.size()]);

  protected onEscape(): void {
    if (this.open()) {
      this.closed.emit();
    }
  }
}
