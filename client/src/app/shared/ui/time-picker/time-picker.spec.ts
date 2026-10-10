import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { TimePicker } from './time-picker';

describe('TimePicker', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [TimePicker] }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(TimePicker);
    fixture.componentRef.setInput('placeholder', 'Seleccionar hora');
    fixture.detectChanges();
    return fixture;
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const toggle = (fixture: any) => {
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();
    fixture.detectChanges();
  };

  it('shows the placeholder when there is no value', () => {
    const fixture = create();

    expect(fixture.nativeElement.textContent).toContain('Seleccionar hora');
  });

  it('shows the selected value', () => {
    const fixture = create();
    fixture.componentInstance.writeValue('09:30');
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('09:30');
  });

  it('lists every hour and every five-minute step when opened', () => {
    const fixture = create();
    toggle(fixture);

    const columns = fixture.nativeElement.querySelectorAll('.max-h-48');
    expect(columns).toHaveLength(2);
    expect(columns[0].querySelectorAll('button')).toHaveLength(24);
    expect(columns[1].querySelectorAll('button')).toHaveLength(12);
  });

  it('commits HH:MM after selecting an hour and a minute', () => {
    const fixture = create();
    const changed = vi.fn();
    fixture.componentInstance.registerOnChange(changed);
    toggle(fixture);

    (
      fixture.nativeElement
        .querySelectorAll('.max-h-48')[0]
        .querySelectorAll('button')[9] as HTMLButtonElement
    ).click();
    fixture.detectChanges();
    (
      fixture.nativeElement
        .querySelectorAll('.max-h-48')[1]
        .querySelectorAll('button')[6] as HTMLButtonElement
    ).click();
    fixture.detectChanges();

    expect(changed).toHaveBeenLastCalledWith('09:30');
    expect(fixture.nativeElement.textContent).toContain('09:30');
  });

  it('clears the value', () => {
    const fixture = create();
    fixture.componentInstance.writeValue('09:30');
    const changed = vi.fn();
    fixture.componentInstance.registerOnChange(changed);
    toggle(fixture);

    const clear = Array.from(fixture.nativeElement.querySelectorAll('button')).find((button) =>
      (button as HTMLButtonElement).textContent?.includes('Limpiar'),
    ) as HTMLButtonElement | undefined;
    clear?.click();
    fixture.detectChanges();

    expect(changed).toHaveBeenCalledWith('');
    expect(fixture.nativeElement.textContent).toContain('Seleccionar hora');
  });

  it('does not open when disabled', () => {
    const fixture = create();
    fixture.componentInstance.setDisabledState(true);
    fixture.detectChanges();

    toggle(fixture);

    expect(fixture.nativeElement.querySelector('.max-h-48')).toBeNull();
  });
});
