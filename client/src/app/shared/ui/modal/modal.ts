import { Component, input, output } from '@angular/core';

@Component({
  selector: 'app-modal',
  host: {
    '(document:keydown.escape)': 'onEscape()',
  },
  template: `
    @if (open()) {
      <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50" (click)="closed.emit()" aria-hidden="true"></div>

        <div
          role="dialog"
          aria-modal="true"
          class="relative z-10 w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-lg"
        >
          <div class="mb-5 flex items-center justify-between gap-4">
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

          <ng-content />
        </div>
      </div>
    }
  `,
})
export class Modal {
  readonly open = input(false);
  readonly title = input('');
  readonly closed = output<void>();

  protected onEscape(): void {
    if (this.open()) {
      this.closed.emit();
    }
  }
}
