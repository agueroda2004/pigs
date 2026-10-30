import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { Subject, of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../core/auth/auth.service';
import { FarrowingSchedule, FarrowingScheduleEntry } from '../../core/services/service.models';
import { ServicesService } from '../../core/services/services.service';
import { UserHome } from './user-home';

function entry(overrides: Partial<FarrowingScheduleEntry> = {}): FarrowingScheduleEntry {
  return {
    id: '1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    expected_farrowing_date: '2026-06-20',
    average_gestation_days: 110.5,
    average_expected_farrowing_date: '2026-06-23',
    ...overrides,
  };
}

function scheduleOf(
  upcoming: FarrowingScheduleEntry[] = [],
  overdue: FarrowingScheduleEntry[] = [],
): FarrowingSchedule {
  return { upcoming, overdue };
}

class ServicesStub {
  getFarrowingSchedule = vi.fn(() => of(scheduleOf()));
}

describe('UserHome', () => {
  let stub: ServicesStub;
  let user: WritableSignal<{ id: string; username: string; role: string }>;

  beforeEach(async () => {
    stub = new ServicesStub();
    user = signal({ id: '1', username: 'ana', role: 'User' });
    await TestBed.configureTestingModule({
      imports: [UserHome],
      providers: [
        { provide: AuthService, useValue: { user } },
        { provide: ServicesService, useValue: stub },
      ],
    }).compileComponents();
  });

  async function render(): Promise<{ fixture: any; element: HTMLElement }> {
    const fixture = TestBed.createComponent(UserHome);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    return { fixture, element: fixture.nativeElement as HTMLElement };
  }

  it('greets the signed-in user and requests the schedule once', async () => {
    const { element } = await render();

    expect(element.textContent).toContain('Hola, ana');
    expect(stub.getFarrowingSchedule).toHaveBeenCalledTimes(1);
  });

  it('renders both dates and the gestation average on each card', async () => {
    stub.getFarrowingSchedule = vi.fn(() =>
      of(
        scheduleOf(
          [entry({ id: 'u1', sow_code: 'C-001', expected_farrowing_date: '2026-06-20' })],
          [
            entry({
              id: 'o1',
              sow_code: 'D-001',
              expected_farrowing_date: '2026-06-01',
              average_gestation_days: 114,
              average_expected_farrowing_date: '2026-06-03',
            }),
          ],
        ),
      ),
    );

    const { element } = await render();
    const text = element.textContent ?? '';

    expect(text).toContain('Próximos partos (7 días)');
    expect(text).toContain('Partos vencidos');
    expect(text).toContain('Cerda C-001');
    expect(text).toContain('Parto probable (114 días): 2026-06-20');
    expect(text).toContain('Parto según su promedio: 2026-06-23');
    expect(text).toContain('Gestación promedio: 110.5 días');
    expect(text).toContain('Cerda D-001');
    expect(text).toContain('Parto probable (114 días): 2026-06-01');
    expect(text).toContain('Parto según su promedio: 2026-06-03');
    expect(text).toContain('Gestación promedio: 114 días');
  });

  it('keeps the standard date and shows a fallback when history is missing', async () => {
    stub.getFarrowingSchedule = vi.fn(() =>
      of(
        scheduleOf([
          entry({
            average_gestation_days: null,
            average_expected_farrowing_date: null,
          }),
        ]),
      ),
    );

    const { element } = await render();
    const text = element.textContent ?? '';

    expect(text).toContain('Parto probable (114 días): 2026-06-20');
    expect(text).toContain('Parto según su promedio: Sin historial');
    expect(text).toContain('Gestación promedio: Sin historial');
  });

  it('shows a loading state until the request resolves', async () => {
    const pending = new Subject<FarrowingSchedule>();
    stub.getFarrowingSchedule = vi.fn(() => pending.asObservable());
    const fixture = TestBed.createComponent(UserHome);

    fixture.detectChanges();
    expect(fixture.nativeElement.querySelectorAll('.animate-pulse')).toHaveLength(2);

    pending.next(scheduleOf([entry({ sow_code: 'C-003' })]));
    pending.complete();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelectorAll('.animate-pulse')).toHaveLength(0);
    expect(fixture.nativeElement.textContent).toContain('Cerda C-003');
  });

  it('shows a server-empty message for each group', async () => {
    const { element } = await render();
    const text = element.textContent ?? '';

    expect(text).toContain('No hay partos previstos para los próximos 7 días.');
    expect(text).toContain('No hay partos vencidos.');
    expect(text).not.toContain('Ningún parto coincide');
  });

  it('shows the failure state and retries the request', async () => {
    const failing = vi.fn(() => throwError(() => new Error('failed')));
    stub.getFarrowingSchedule = failing;
    const fixture = TestBed.createComponent(UserHome);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    const element = fixture.nativeElement as HTMLElement;
    expect(failing).toHaveBeenCalledTimes(1);
    expect(element.textContent).toContain('No se pudo cargar la programación de partos.');

    const retrying = vi.fn(() => of(scheduleOf([entry({ sow_code: 'C-002' })])));
    stub.getFarrowingSchedule = retrying;
    const retry = Array.from(element.querySelectorAll('button')).find((button) =>
      button.textContent?.includes('Reintentar'),
    );
    retry?.click();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(retrying).toHaveBeenCalledTimes(1);
    expect(element.textContent).toContain('Cerda C-002');
    expect(element.textContent).not.toContain('No se pudo cargar la programación de partos.');
  });

  it('filters entries live by sow code across both sections', async () => {
    stub.getFarrowingSchedule = vi.fn(() =>
      of(
        scheduleOf(
          [entry({ id: 'u1', sow_code: 'C-001' }), entry({ id: 'u2', sow_code: 'C-002' })],
          [entry({ id: 'o1', sow_code: 'D-001' })],
        ),
      ),
    );

    const { fixture, element } = await render();
    const component = fixture.componentInstance as any;

    component.filterCode.set('c-002');
    fixture.detectChanges();

    expect(element.textContent).toContain('Cerda C-002');
    expect(element.textContent).not.toContain('Cerda C-001');
    expect(element.textContent).not.toContain('Cerda D-001');
    expect(element.textContent).toContain('Ningún parto coincide con el código "c-002".');
    expect(element.textContent).toContain('Próximos partos (7 días)');
  });

  it('trims whitespace and ignores case while filtering', async () => {
    stub.getFarrowingSchedule = vi.fn(() =>
      of(scheduleOf([entry({ id: 'u1', sow_code: 'C-001' })])),
    );

    const { fixture, element } = await render();
    const component = fixture.componentInstance as any;

    component.filterCode.set('  c-001  ');
    fixture.detectChanges();

    expect(element.textContent).toContain('Cerda C-001');
    expect(element.textContent).not.toContain('Ningún parto coincide');
  });

  it('clears the filter and restores all entries', async () => {
    stub.getFarrowingSchedule = vi.fn(() =>
      of(
        scheduleOf(
          [entry({ id: 'u1', sow_code: 'C-001' }), entry({ id: 'u2', sow_code: 'C-002' })],
          [],
        ),
      ),
    );

    const { fixture, element } = await render();
    const component = fixture.componentInstance as any;

    component.filterCode.set('c-002');
    fixture.detectChanges();
    expect(element.textContent).not.toContain('Cerda C-001');

    component.clearFilter();
    fixture.detectChanges();

    expect(element.textContent).toContain('Cerda C-001');
    expect(element.textContent).toContain('Cerda C-002');
    expect(
      (
        (fixture.nativeElement as HTMLElement).querySelector(
          '#home-sow-code',
        ) as HTMLInputElement | null
      )?.value,
    ).toBe('');
  });

  it('filters without triggering another schedule request', async () => {
    stub.getFarrowingSchedule = vi.fn(() => of(scheduleOf([entry({ sow_code: 'C-001' })])));

    const { fixture } = await render();
    const component = fixture.componentInstance as any;

    component.filterCode.set('zzz');
    fixture.detectChanges();

    expect(stub.getFarrowingSchedule).toHaveBeenCalledTimes(1);
  });

  it('applies the filter from the input event', async () => {
    stub.getFarrowingSchedule = vi.fn(() =>
      of(scheduleOf([entry({ id: 'u1', sow_code: 'C-001' })])),
    );

    const { fixture, element } = await render();
    const input = element.querySelector('#home-sow-code') as HTMLInputElement;
    input.value = 'C-999';
    input.dispatchEvent(new Event('input'));
    fixture.detectChanges();

    expect(element.textContent).toContain('Ningún parto coincide con el código "C-999".');
  });
});
