import {
  Component,
  ElementRef,
  computed,
  effect,
  forwardRef,
  input,
  signal,
  viewChild,
} from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

import { DropdownOption } from '../dropdown/dropdown';

@Component({
  selector: 'app-search-dropdown',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => SearchDropdown),
      multi: true,
    },
  ],
  host: {
    '(document:click)': 'close()',
    '(document:keydown.escape)': 'close()',
  },
  template: `
    <div class="relative" (click)="$event.stopPropagation()">
      <button
        type="button"
        [attr.aria-expanded]="open()"
        aria-haspopup="listbox"
        [disabled]="disabled()"
        (click)="toggle()"
        class="flex w-full items-center justify-between gap-2 rounded-md border border-border bg-background px-3 py-2 text-left text-sm transition hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
      >
        <span [class.text-muted-foreground]="!selectedLabel()">
          {{ selectedLabel() || placeholder() }}
        </span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="h-4 w-4 shrink-0 transition-transform"
          [class.rotate-180]="open()"
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      @if (open()) {
        <div class="absolute z-20 mt-1 w-full rounded-md border border-border bg-card shadow-lg">
          <div class="border-b border-border p-2">
            <input
              #searchInput
              type="text"
              [value]="query()"
              (input)="onQueryInput($event)"
              (keydown)="onKeydown($event)"
              autocomplete="off"
              placeholder="Buscar…"
              class="w-full rounded-md border border-border bg-background px-2.5 py-1.5 text-sm outline-none transition focus:border-foreground focus:ring-2 focus:ring-ring"
            />
          </div>

          @if (filteredOptions().length === 0) {
            <p class="px-3 py-3 text-sm text-muted-foreground">Sin resultados</p>
          } @else {
            <ul role="listbox" class="max-h-60 overflow-auto p-1">
              @for (option of filteredOptions(); track option.value; let index = $index) {
                <li role="option" [attr.aria-selected]="option.value === value()">
                  <button
                    type="button"
                    (click)="select(option)"
                    (mouseenter)="highlighted.set(index)"
                    class="flex w-full items-center justify-between gap-2 rounded px-2 py-1.5 text-left text-sm transition"
                    [class.bg-muted]="isHighlighted(index)"
                    [class.font-medium]="option.value === value()"
                  >
                    <span>{{ option.label }}</span>
                    @if (option.value === value()) {
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
                        <path d="M20 6 9 17l-5-5" />
                      </svg>
                    }
                  </button>
                </li>
              }
            </ul>
          }
        </div>
      }
    </div>
  `,
})
export class SearchDropdown implements ControlValueAccessor {
  readonly options = input<DropdownOption[]>([]);
  readonly placeholder = input('Seleccionar');

  protected readonly value = signal('');
  protected readonly open = signal(false);
  protected readonly disabled = signal(false);
  protected readonly query = signal('');
  protected readonly highlighted = signal(0);

  private readonly searchInput = viewChild<ElementRef<HTMLInputElement>>('searchInput');

  protected readonly selectedLabel = computed(
    () => this.options().find((option) => option.value === this.value())?.label ?? '',
  );

  protected readonly filteredOptions = computed(() => {
    const term = this.query().trim().toLowerCase();
    const options = this.options();
    if (!term) {
      return options;
    }
    return options.filter((option) => option.label.toLowerCase().includes(term));
  });

  private onChange: (value: string) => void = () => {};
  private onTouched: () => void = () => {};

  constructor() {
    effect(() => {
      if (this.open()) {
        this.searchInput()?.nativeElement.focus();
      }
    });
  }

  writeValue(value: string | null): void {
    this.value.set(value ?? '');
  }

  registerOnChange(fn: (value: string) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: () => void): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    this.disabled.set(isDisabled);
  }

  protected toggle(): void {
    if (this.disabled()) {
      return;
    }
    if (this.open()) {
      this.close();
      return;
    }
    this.query.set('');
    const index = this.filteredOptions().findIndex((option) => option.value === this.value());
    this.highlighted.set(index >= 0 ? index : 0);
    this.open.set(true);
  }

  protected close(): void {
    this.open.set(false);
  }

  protected isHighlighted(index: number): boolean {
    return index === this.highlighted();
  }

  protected onQueryInput(event: Event): void {
    this.query.set((event.target as HTMLInputElement).value);
    this.highlighted.set(0);
  }

  protected onKeydown(event: KeyboardEvent): void {
    const options = this.filteredOptions();
    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        if (options.length > 0) {
          this.highlighted.update((index) => (index + 1) % options.length);
        }
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (options.length > 0) {
          this.highlighted.update((index) => (index - 1 + options.length) % options.length);
        }
        break;
      case 'Enter':
        event.preventDefault();
        if (options.length > 0) {
          this.select(options[Math.min(this.highlighted(), options.length - 1)]);
        }
        break;
      default:
        break;
    }
  }

  protected select(option: DropdownOption): void {
    this.value.set(option.value);
    this.onChange(option.value);
    this.onTouched();
    this.query.set('');
    this.open.set(false);
  }
}
