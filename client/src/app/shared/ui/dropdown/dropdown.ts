import {
  Component,
  ElementRef,
  Injector,
  afterNextRender,
  computed,
  forwardRef,
  inject,
  input,
  signal,
  viewChild,
} from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

import { shouldFlipUp } from '../../utils/overlay';

export interface DropdownOption {
  value: string;
  label: string;
}

@Component({
  selector: 'app-dropdown',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => Dropdown),
      multi: true,
    },
  ],
  host: {
    '(document:click)': 'close()',
    '(document:keydown.escape)': 'close()',
    '(window:resize)': 'onResize()',
  },
  template: `
    <div class="relative" (click)="$event.stopPropagation()">
      <button
        #trigger
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
        <ul
          #panel
          role="listbox"
          class="absolute z-20 max-h-60 w-full overflow-auto rounded-md border border-border bg-card p-1 shadow-lg"
          [class.top-full]="!dropUp()"
          [class.mt-1]="!dropUp()"
          [class.bottom-full]="dropUp()"
          [class.mb-1]="dropUp()"
          [class.invisible]="!positioned()"
        >
          @for (option of options(); track option.value) {
            <li role="option" [attr.aria-selected]="option.value === value()">
              <button
                type="button"
                (click)="select(option)"
                class="flex w-full items-center justify-between gap-2 rounded px-2 py-1.5 text-left text-sm transition hover:bg-muted"
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
  `,
})
export class Dropdown implements ControlValueAccessor {
  readonly options = input<DropdownOption[]>([]);
  readonly placeholder = input('Seleccionar');

  protected readonly value = signal('');
  protected readonly open = signal(false);
  protected readonly disabled = signal(false);
  protected readonly dropUp = signal(false);
  protected readonly positioned = signal(false);

  private readonly triggerRef = viewChild<ElementRef<HTMLButtonElement>>('trigger');
  private readonly panelRef = viewChild<ElementRef<HTMLElement>>('panel');
  private readonly injector = inject(Injector);

  protected readonly selectedLabel = computed(
    () => this.options().find((option) => option.value === this.value())?.label ?? '',
  );

  private onChange: (value: string) => void = () => {};
  private onTouched: () => void = () => {};

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
    this.openPanel();
  }

  // openPanel reveals the list and schedules its placement after the next render.
  // It resets the placement so the panel can be measured before it becomes visible.
  private openPanel(): void {
    this.dropUp.set(false);
    this.positioned.set(false);
    this.open.set(true);
    afterNextRender(() => this.positionPanel(), { injector: this.injector });
  }

  // positionPanel measures the trigger and panel to choose the open direction.
  // It is reused when the window is resized while the panel is visible.
  private positionPanel(): void {
    const trigger = this.triggerRef()?.nativeElement;
    const panel = this.panelRef()?.nativeElement;
    if (trigger && panel) {
      this.dropUp.set(shouldFlipUp(trigger, panel));
    }
    this.positioned.set(true);
  }

  protected onResize(): void {
    if (this.open()) {
      this.positionPanel();
    }
  }

  protected close(): void {
    this.open.set(false);
    this.positioned.set(false);
    this.dropUp.set(false);
  }

  protected select(option: DropdownOption): void {
    this.value.set(option.value);
    this.onChange(option.value);
    this.onTouched();
    this.open.set(false);
  }
}
