import {
  Component,
  DestroyRef,
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

import { OverlayPlacement, computeOverlayPlacement } from '../../utils/overlay';

const HOURS = Array.from({ length: 24 }, (_, hour) => String(hour).padStart(2, '0'));

@Component({
  selector: 'app-time-picker',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => TimePicker),
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
        [disabled]="disabled()"
        (click)="toggle()"
        class="flex w-full items-center justify-between gap-2 rounded-md border border-border bg-background px-3 py-2 text-left text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        [class.border-foreground]="open()"
      >
        <span [class.text-muted-foreground]="!value()">
          {{ value() || placeholder() }}
        </span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="h-4 w-4 shrink-0 text-muted-foreground"
        >
          <circle cx="12" cy="12" r="10" />
          <path d="M12 6v6l4 2" />
        </svg>
      </button>

      @if (open()) {
        <div
          #panel
          class="fixed z-30 rounded-md border border-border bg-card p-3 shadow-lg"
          [style.top.px]="placement()?.top"
          [style.left.px]="placement()?.left"
          [style.width.px]="placement()?.width"
          [class.invisible]="!positioned()"
        >
          <div class="flex gap-2">
            <div class="flex-1">
              <p class="mb-1 text-center text-xs font-medium text-muted-foreground">Hora</p>
              <div class="max-h-48 overflow-auto rounded-md border border-border p-1">
                @for (hour of hours; track hour) {
                  <button
                    type="button"
                    (click)="selectHour(hour)"
                    class="w-full rounded px-2 py-1 text-center text-sm transition hover:bg-muted"
                    [class.bg-primary]="hour === selectedHour()"
                    [class.font-semibold]="hour === selectedHour()"
                    [class.text-primary-foreground]="hour === selectedHour()"
                  >
                    {{ hour }}
                  </button>
                }
              </div>
            </div>

            <div class="flex-1">
              <p class="mb-1 text-center text-xs font-medium text-muted-foreground">Minuto</p>
              <div class="max-h-48 overflow-auto rounded-md border border-border p-1">
                @for (minute of minutes(); track minute) {
                  <button
                    type="button"
                    (click)="selectMinute(minute)"
                    class="w-full rounded px-2 py-1 text-center text-sm transition hover:bg-muted"
                    [class.bg-primary]="minute === selectedMinute()"
                    [class.font-semibold]="minute === selectedMinute()"
                    [class.text-primary-foreground]="minute === selectedMinute()"
                  >
                    {{ minute }}
                  </button>
                }
              </div>
            </div>
          </div>

          @if (value()) {
            <button
              type="button"
              (click)="selectClear()"
              class="mt-2 w-full rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:bg-muted hover:text-danger"
            >
              Limpiar
            </button>
          }
        </div>
      }
    </div>
  `,
})
export class TimePicker implements ControlValueAccessor {
  readonly placeholder = input('Seleccionar hora');
  readonly minuteStep = input(5);

  protected readonly hours = HOURS;
  protected readonly value = signal('');
  protected readonly disabled = signal(false);
  protected readonly open = signal(false);
  protected readonly positioned = signal(false);
  protected readonly placement = signal<OverlayPlacement | null>(null);

  protected readonly minutes = computed(() => {
    const step = Math.min(Math.max(this.minuteStep(), 1), 60);
    const list: string[] = [];
    for (let minute = 0; minute < 60; minute += step) {
      list.push(String(minute).padStart(2, '0'));
    }
    return list;
  });

  protected readonly selectedHour = computed(() => this.value().split(':')[0] ?? '');
  protected readonly selectedMinute = computed(() => this.value().split(':')[1] ?? '');

  private readonly triggerRef = viewChild<ElementRef<HTMLButtonElement>>('trigger');
  private readonly panelRef = viewChild<ElementRef<HTMLElement>>('panel');
  private readonly injector = inject(Injector);

  private onChange: (value: string) => void = () => {};
  private onTouched: () => void = () => {};

  constructor() {
    const destroyRef = inject(DestroyRef);
    const onScroll = () => this.reposition();
    document.addEventListener('scroll', onScroll, true);
    destroyRef.onDestroy(() => document.removeEventListener('scroll', onScroll, true));
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
    this.openPanel();
  }

  // openPanel reveals the picker and schedules its placement after the next render.
  // It resets the placement so the panel can be measured before it becomes visible.
  private openPanel(): void {
    this.placement.set(null);
    this.positioned.set(false);
    this.open.set(true);
    afterNextRender(() => this.positionPanel(), { injector: this.injector });
  }

  // positionPanel measures the trigger and panel to place the fixed picker.
  // It is reused on window resize and while an ancestor scrolls.
  private positionPanel(): void {
    const trigger = this.triggerRef()?.nativeElement;
    const panel = this.panelRef()?.nativeElement;
    if (trigger && panel) {
      this.placement.set(computeOverlayPlacement(trigger, panel));
    }
    this.positioned.set(true);
  }

  // reposition updates the fixed placement while the panel is visible.
  // It keeps the picker anchored to its trigger when an ancestor scrolls.
  private reposition(): void {
    if (this.open()) {
      this.positionPanel();
    }
  }

  protected onResize(): void {
    this.reposition();
  }

  protected close(): void {
    this.open.set(false);
    this.positioned.set(false);
    this.placement.set(null);
  }

  protected selectHour(hour: string): void {
    const minute = this.selectedMinute() || this.minutes()[0];
    this.commit(`${hour}:${minute}`);
  }

  protected selectMinute(minute: string): void {
    const hour = this.selectedHour() || this.hours[0];
    this.commit(`${hour}:${minute}`);
  }

  protected selectClear(): void {
    this.commit('');
  }

  // commit updates the local value and notifies the form control.
  // Angular does not call writeValue after onChange, so the signal is synced here.
  private commit(value: string): void {
    this.value.set(value);
    this.onChange(value);
    this.onTouched();
  }
}
