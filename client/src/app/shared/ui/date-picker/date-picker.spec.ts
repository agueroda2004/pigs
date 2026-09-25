import { ComponentFixture, TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { DatePicker } from './date-picker';

function mockRects(triggerTop: number, triggerBottom: number, panelHeight: number, viewport = 768) {
  vi.spyOn(window, 'innerHeight', 'get').mockReturnValue(viewport);
  vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(function (this: Element) {
    if (this.tagName === 'BUTTON') {
      return {
        top: triggerTop,
        bottom: triggerBottom,
        height: triggerBottom - triggerTop,
      } as DOMRect;
    }
    return { top: 0, bottom: 0, height: panelHeight } as DOMRect;
  });
}

describe('DatePicker', () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let component: any;
  let fixture: ComponentFixture<DatePicker>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [DatePicker] }).compileComponents();
  });

  afterEach(() => vi.restoreAllMocks());

  const create = (): ComponentFixture<DatePicker> => {
    fixture = TestBed.createComponent(DatePicker);
    component = fixture.componentInstance;
    return fixture;
  };

  it('formats the selected value for display', () => {
    create();
    component.writeValue('2026-01-05');

    expect(component.displayValue()).toBe('5 de Enero 2026');
  });

  it('emits the ISO date when a day is selected and closes', () => {
    create();
    let emitted = '';
    component.registerOnChange((value: string) => (emitted = value));
    component.writeValue('2026-01-15');

    component.toggle();
    component.selectDay(20);

    expect(emitted).toBe('2026-01-20');
    expect(component.value()).toBe('2026-01-20');
    expect(component.displayValue()).toBe('20 de Enero 2026');
    expect(component.open()).toBe(false);
  });

  it('does not open when disabled', () => {
    create();
    component.setDisabledState(true);

    component.toggle();

    expect(component.open()).toBe(false);
  });

  it('disables days outside the max date', () => {
    const currentFixture = create();
    currentFixture.componentRef.setInput('maxDate', '2026-01-10');
    component.writeValue('2026-01-05');
    component.toggle();

    expect(component.isDayDisabled(5)).toBe(false);
    expect(component.isDayDisabled(20)).toBe(true);
  });

  it('selects today through the shortcut', () => {
    create();
    let emitted = '';
    component.registerOnChange((value: string) => (emitted = value));

    component.selectToday();

    const today = new Date();
    const expected = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(
      today.getDate(),
    ).padStart(2, '0')}`;
    expect(emitted).toBe(expected);
    expect(component.value()).toBe(expected);
    expect(component.open()).toBe(false);
  });

  it('clears the value and emits an empty string', () => {
    create();
    let emitted = 'initial';
    component.registerOnChange((value: string) => (emitted = value));
    component.writeValue('2026-01-05');

    component.selectClear();

    expect(emitted).toBe('');
    expect(component.value()).toBe('');
    expect(component.displayValue()).toBe('');
    expect(component.open()).toBe(false);
  });

  it('opens downwards when the calendar fits below the trigger', async () => {
    mockRects(100, 140, 200);
    const currentFixture = create();
    currentFixture.detectChanges();
    component.toggle();
    currentFixture.detectChanges();
    await currentFixture.whenStable();
    currentFixture.detectChanges();

    const panel = currentFixture.nativeElement.querySelector('div.absolute') as HTMLElement;
    expect(panel.classList.contains('top-full')).toBe(true);
    expect(panel.classList.contains('bottom-full')).toBe(false);
  });

  it('opens upwards when the calendar does not fit below the trigger', async () => {
    mockRects(700, 740, 340);
    const currentFixture = create();
    currentFixture.detectChanges();
    component.toggle();
    currentFixture.detectChanges();
    await currentFixture.whenStable();
    currentFixture.detectChanges();

    const panel = currentFixture.nativeElement.querySelector('div.absolute') as HTMLElement;
    expect(panel.classList.contains('bottom-full')).toBe(true);
    expect(panel.classList.contains('top-full')).toBe(false);
  });
});
