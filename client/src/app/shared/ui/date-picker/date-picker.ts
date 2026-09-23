import { Component, computed, forwardRef, input, signal } from '@angular/core';
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from '@angular/forms';

const MONTHS = [
  'Enero',
  'Febrero',
  'Marzo',
  'Abril',
  'Mayo',
  'Junio',
  'Julio',
  'Agosto',
  'Septiembre',
  'Octubre',
  'Noviembre',
  'Diciembre',
];

const DAYS_OF_WEEK = ['Do', 'Lu', 'Ma', 'Mi', 'Ju', 'Vi', 'Sá'];
const YEARS_PER_PAGE = 25;

type ViewMode = 'days' | 'months' | 'years';

// parseISO converts a YYYY-MM-DD (or ISO datetime) value into a local Date.
// It returns null for empty or unparseable values.
function parseISO(value: string | null | undefined): Date | null {
  if (!value) {
    return null;
  }
  const datePart = value.includes('T') ? value.split('T')[0] : value;
  const [year, month, day] = datePart.split('-').map(Number);
  if (!year || !month || !day) {
    return null;
  }
  const date = new Date(year, month - 1, day);
  return Number.isNaN(date.getTime()) ? null : date;
}

// toISO formats a Date as a local YYYY-MM-DD string without timezone shifts.
function toISO(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

// startOfDay strips the time portion from a Date.
function startOfDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

// isSameDay reports whether two dates share the same calendar day.
function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

// daysInMonth returns the number of days of the given month and year.
function daysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate();
}

// firstDayOfMonth returns the weekday index (0 = Sunday) of the first day.
function firstDayOfMonth(year: number, month: number): number {
  return new Date(year, month, 1).getDay();
}

@Component({
  selector: 'app-date-picker',
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => DatePicker),
      multi: true,
    },
  ],
  host: {
    '(document:click)': 'close()',
    '(document:keydown.escape)': 'close()',
  },
  template: `
    <div class="relative" [class.inline]="inline()" (click)="$event.stopPropagation()">
      <button
        type="button"
        [disabled]="disabled()"
        (click)="toggle()"
        class="flex w-full items-center justify-between gap-2 rounded-md border border-border bg-background px-3 py-2 text-left text-sm outline-none transition focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        [class.border-foreground]="open()"
      >
        <span [class.text-muted-foreground]="!value()">
          {{ value() ? displayValue() : placeholder() }}
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
          <path d="M8 2v4M16 2v4M3 10h18" />
          <rect width="18" height="18" x="3" y="4" rx="2" />
        </svg>
      </button>

      @if (open()) {
        <div
          class="z-30 mt-1 rounded-md border border-border bg-card p-3 shadow-lg"
          [class.relative]="inline()"
          [class.absolute]="!inline()"
          [class.left-0]="!inline()"
          [class.right-0]="!inline()"
        >
          @if (viewMode() === 'days') {
            <div class="mb-2 flex items-center justify-between">
              <button
                type="button"
                (click)="prevMonth()"
                aria-label="Mes anterior"
                class="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground"
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
                  <path d="m15 18-6-6 6-6" />
                </svg>
              </button>
              <button
                type="button"
                (click)="viewMode.set('months')"
                class="rounded-md px-3 py-1 text-sm font-semibold transition hover:bg-muted"
              >
                {{ months[viewMonth()] }} {{ viewYear() }}
              </button>
              <button
                type="button"
                (click)="nextMonth()"
                [disabled]="!canGoNext()"
                aria-label="Mes siguiente"
                class="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground disabled:opacity-30 disabled:hover:bg-transparent"
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
                  <path d="m9 18 6-6-6-6" />
                </svg>
              </button>
            </div>

            <div class="mb-1 grid grid-cols-7 gap-0.5">
              @for (day of daysOfWeek; track day) {
                <div
                  class="flex h-8 items-center justify-center text-xs font-medium text-muted-foreground"
                >
                  {{ day }}
                </div>
              }
            </div>

            <div class="grid grid-cols-7 gap-0.5">
              @for (day of calendarDays(); track $index) {
                @if (day === null) {
                  <div></div>
                } @else {
                  <button
                    type="button"
                    (click)="selectDay(day)"
                    [disabled]="isDayDisabled(day)"
                    [class]="dayClass(day)"
                  >
                    {{ day }}
                  </button>
                }
              }
            </div>
          }

          @if (viewMode() === 'months') {
            <div class="mb-2 flex items-center justify-between">
              <button
                type="button"
                (click)="prevYear()"
                aria-label="Año anterior"
                class="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground"
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
                  <path d="m15 18-6-6 6-6" />
                </svg>
              </button>
              <button
                type="button"
                (click)="viewMode.set('years')"
                class="rounded-md px-3 py-1 text-sm font-semibold transition hover:bg-muted"
              >
                {{ viewYear() }}
              </button>
              <button
                type="button"
                (click)="nextYear()"
                [disabled]="!canGoNextYear()"
                aria-label="Año siguiente"
                class="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground disabled:opacity-30 disabled:hover:bg-transparent"
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
                  <path d="m9 18 6-6-6-6" />
                </svg>
              </button>
            </div>

            <div class="grid grid-cols-3 gap-1">
              @for (month of months; track month; let index = $index) {
                <button
                  type="button"
                  (click)="selectMonth(index)"
                  [disabled]="isMonthDisabled(index)"
                  [class]="monthClass(index)"
                >
                  {{ month.slice(0, 3) }}
                </button>
              }
            </div>
          }

          @if (viewMode() === 'years') {
            <div class="mb-2 flex items-center justify-between">
              <button
                type="button"
                (click)="prevYearPage()"
                [disabled]="yearPage() === 0"
                aria-label="Página anterior"
                class="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground disabled:opacity-30 disabled:hover:bg-transparent"
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
                  <path d="m15 18-6-6 6-6" />
                </svg>
              </button>
              <span class="text-sm font-semibold">
                {{ startYear() }}–{{ startYear() + yearsPerPage - 1 }}
              </span>
              <button
                type="button"
                (click)="nextYearPage()"
                [disabled]="!canGoNextYearPage()"
                aria-label="Página siguiente"
                class="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition hover:bg-muted hover:text-foreground disabled:opacity-30 disabled:hover:bg-transparent"
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
                  <path d="m9 18 6-6-6-6" />
                </svg>
              </button>
            </div>

            <div class="grid grid-cols-5 gap-1">
              @for (year of years(); track year) {
                <button
                  type="button"
                  (click)="selectYear(year)"
                  [disabled]="isYearDisabled(year)"
                  [class]="yearClass(year)"
                >
                  {{ year }}
                </button>
              }
            </div>
          }

          <div class="mt-2 flex gap-2">
            <button
              type="button"
              (click)="selectToday()"
              class="flex-1 rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:bg-muted hover:text-primary"
            >
              Hoy
            </button>
            @if (value()) {
              <button
                type="button"
                (click)="selectClear()"
                class="flex-1 rounded-md border border-border px-3 py-1.5 text-xs font-medium text-muted-foreground transition hover:bg-muted hover:text-danger"
              >
                Limpiar
              </button>
            }
          </div>
        </div>
      }
    </div>
  `,
})
export class DatePicker implements ControlValueAccessor {
  readonly placeholder = input('Seleccionar fecha');
  readonly minDate = input<string | null>(null);
  readonly maxDate = input<string | null>(null);
  readonly inline = input(false);

  protected readonly months = MONTHS;
  protected readonly daysOfWeek = DAYS_OF_WEEK;
  protected readonly yearsPerPage = YEARS_PER_PAGE;

  protected readonly value = signal('');
  protected readonly disabled = signal(false);
  protected readonly open = signal(false);
  protected readonly viewMode = signal<ViewMode>('days');
  protected readonly yearPage = signal(0);
  protected readonly viewYear = signal(new Date().getFullYear());
  protected readonly viewMonth = signal(new Date().getMonth());

  private readonly selectedDate = computed(() => parseISO(this.value()));
  private readonly effectiveMin = computed(() => parseISO(this.minDate()));
  private readonly effectiveMax = computed(() => parseISO(this.maxDate()));

  protected readonly displayValue = computed(() => {
    const date = this.selectedDate();
    return date ? `${date.getDate()} de ${MONTHS[date.getMonth()]} ${date.getFullYear()}` : '';
  });

  private readonly maxYear = computed(
    () => this.effectiveMax()?.getFullYear() ?? new Date().getFullYear() + 50,
  );

  protected readonly startYear = computed(
    () => this.maxYear() - 100 + this.yearPage() * YEARS_PER_PAGE,
  );

  protected readonly years = computed(() => {
    const start = this.startYear();
    const max = this.maxYear();
    const list: number[] = [];
    for (let year = start; year < start + YEARS_PER_PAGE && year <= max; year++) {
      list.push(year);
    }
    return list;
  });

  protected readonly calendarDays = computed<(number | null)[]>(() => {
    const total = daysInMonth(this.viewYear(), this.viewMonth());
    const leading = firstDayOfMonth(this.viewYear(), this.viewMonth());
    const cells: (number | null)[] = [];
    for (let i = 0; i < leading; i++) {
      cells.push(null);
    }
    for (let day = 1; day <= total; day++) {
      cells.push(day);
    }
    return cells;
  });

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
    const next = !this.open();
    if (next) {
      this.viewMode.set('days');
      this.yearPage.set(0);
      const base = this.selectedDate() ?? startOfDay(new Date());
      this.viewYear.set(base.getFullYear());
      this.viewMonth.set(base.getMonth());
    }
    this.open.set(next);
  }

  protected close(): void {
    this.open.set(false);
  }

  protected prevMonth(): void {
    if (this.viewMonth() === 0) {
      this.viewMonth.set(11);
      this.viewYear.update((year) => year - 1);
    } else {
      this.viewMonth.update((month) => month - 1);
    }
  }

  protected nextMonth(): void {
    if (this.viewMonth() === 11) {
      this.viewMonth.set(0);
      this.viewYear.update((year) => year + 1);
    } else {
      this.viewMonth.update((month) => month + 1);
    }
  }

  protected prevYear(): void {
    this.viewYear.update((year) => year - 1);
  }

  protected nextYear(): void {
    this.viewYear.update((year) => year + 1);
  }

  protected prevYearPage(): void {
    this.yearPage.update((page) => Math.max(0, page - 1));
  }

  protected nextYearPage(): void {
    this.yearPage.update((page) => page + 1);
  }

  protected canGoNext(): boolean {
    const max = this.effectiveMax();
    if (!max) {
      return true;
    }
    if (this.viewYear() < max.getFullYear()) {
      return true;
    }
    return this.viewYear() === max.getFullYear() && this.viewMonth() < max.getMonth();
  }

  protected canGoNextYear(): boolean {
    const max = this.effectiveMax();
    return max ? this.viewYear() < max.getFullYear() : true;
  }

  protected canGoNextYearPage(): boolean {
    const max = this.effectiveMax();
    return max ? this.startYear() + YEARS_PER_PAGE <= max.getFullYear() : true;
  }

  protected selectDay(day: number): void {
    if (this.isDayDisabled(day)) {
      return;
    }
    this.commit(toISO(new Date(this.viewYear(), this.viewMonth(), day)));
    this.open.set(false);
  }

  protected selectMonth(month: number): void {
    if (this.isMonthDisabled(month)) {
      return;
    }
    this.viewMonth.set(month);
    this.viewMode.set('days');
  }

  protected selectYear(year: number): void {
    if (this.isYearDisabled(year)) {
      return;
    }
    this.viewYear.set(year);
    this.viewMode.set('months');
    const page = Math.floor((year - (this.maxYear() - 100)) / YEARS_PER_PAGE);
    this.yearPage.set(Math.max(0, page));
  }

  protected selectToday(): void {
    const today = startOfDay(new Date());
    this.commit(toISO(today));
    this.viewMonth.set(today.getMonth());
    this.viewYear.set(today.getFullYear());
    this.open.set(false);
  }

  protected selectClear(): void {
    this.commit('');
    this.open.set(false);
  }

  // commit updates the local value and notifies the form control.
  // Angular does not call writeValue after onChange, so the signal is synced here.
  private commit(iso: string): void {
    this.value.set(iso);
    this.onChange(iso);
    this.onTouched();
  }

  protected isDayDisabled(day: number): boolean {
    const date = new Date(this.viewYear(), this.viewMonth(), day);
    const min = this.effectiveMin();
    const max = this.effectiveMax();
    if (max && date > max) {
      return true;
    }
    return Boolean(min && date < min);
  }

  protected isMonthDisabled(month: number): boolean {
    const max = this.effectiveMax();
    const min = this.effectiveMin();
    const afterMax = max
      ? this.viewYear() > max.getFullYear() ||
        (this.viewYear() === max.getFullYear() && month > max.getMonth())
      : false;
    const beforeMin = min
      ? this.viewYear() < min.getFullYear() ||
        (this.viewYear() === min.getFullYear() && month < min.getMonth())
      : false;
    return afterMax || beforeMin;
  }

  protected isYearDisabled(year: number): boolean {
    const max = this.effectiveMax();
    const min = this.effectiveMin();
    if (max && year > max.getFullYear()) {
      return true;
    }
    return Boolean(min && year < min.getFullYear());
  }

  protected dayClass(day: number): string {
    const base = 'flex h-8 w-full items-center justify-center rounded-md text-sm transition';
    const date = new Date(this.viewYear(), this.viewMonth(), day);
    if (this.selectedDate() && isSameDay(date, this.selectedDate() as Date)) {
      return `${base} bg-primary font-semibold text-primary-foreground`;
    }
    if (this.isDayDisabled(day)) {
      return `${base} cursor-not-allowed text-muted-foreground/50`;
    }
    if (isSameDay(date, startOfDay(new Date()))) {
      return `${base} font-semibold text-primary hover:bg-muted`;
    }
    return `${base} text-foreground hover:bg-muted`;
  }

  protected monthClass(month: number): string {
    const base = 'rounded-md px-2 py-2 text-sm transition';
    if (this.viewMonth() === month) {
      return `${base} bg-primary font-semibold text-primary-foreground`;
    }
    if (this.isMonthDisabled(month)) {
      return `${base} cursor-not-allowed text-muted-foreground/50`;
    }
    return `${base} text-foreground hover:bg-muted`;
  }

  protected yearClass(year: number): string {
    const base = 'rounded-md px-2 py-1.5 text-sm transition';
    if (this.viewYear() === year) {
      return `${base} bg-primary font-semibold text-primary-foreground`;
    }
    if (this.isYearDisabled(year)) {
      return `${base} cursor-not-allowed text-muted-foreground/50`;
    }
    return `${base} text-foreground hover:bg-muted`;
  }
}
