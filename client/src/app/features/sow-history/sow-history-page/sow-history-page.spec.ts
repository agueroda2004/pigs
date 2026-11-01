import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, Subject, throwError } from 'rxjs';
import { vi } from 'vitest';

import { SowHistory } from '../../../core/sows/sow-history.models';
import { SowsService } from '../../../core/sows/sows.service';
import { SowHistoryPage } from './sow-history-page';

const historyPayload: SowHistory = {
  sow: {
    code: 'C-001', entry_date: '2026-01-10', birth_date: null,
    state: 'Viva', parity: 2, breed_name: 'Large White', age: '1 año',
  },
  events: [
    { type: 'sow_entry', date: '2026-01-10', data: {
      code: 'C-001', entry_date: '2026-01-10', birth_date: null,
      state: 'Viva', parity: 2, breed_name: 'Large White', age: '1 año',
    } },
    { type: 'abortion', date: '2026-02-10', data: {
      sow_code: 'C-001', abortion_date: '2026-02-10', cause: 'Infeccioso', note: null,
    } },
    { type: 'farrowing', date: '2026-03-01', data: {
      farrow_date: '2026-03-01', start_time: '22:00', end_time: '02:30', duration: '04:30',
      location: 'Corral A', live_born: 10, mummified: 1, stillborn: 1,
      litter_weight: 15.5, stillborn_weight: 2.5, is_manipulated: false, note: null,
      is_nurse: false,
      operators: [{ operator_name: 'Op 1' }],
      medications: [{ medication_name: 'Oxitocina', dose: 2, applied_by_name: 'Op 2' }],
    } },
  ],
};

describe('SowHistoryPage', () => {
  const sowService = { getSowHistory: vi.fn() };

  beforeEach(async () => {
    sowService.getSowHistory.mockReset();
    await TestBed.configureTestingModule({
      imports: [SowHistoryPage],
      providers: [{ provide: SowsService, useValue: sowService }],
    }).compileComponents();
  });

  it('searches by code and renders the sow and its timeline', async () => {
    sowService.getSowHistory.mockReturnValue(of(historyPayload));
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = ' C-001 ';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    expect(sowService.getSowHistory).toHaveBeenCalledWith('C-001');
    expect(fixture.nativeElement.textContent).toContain('Cerda C-001');
    expect(fixture.nativeElement.textContent).toContain('Aborto');
    expect(fixture.nativeElement.textContent).toContain('Horario: 22:00 → 02:30');
    expect(fixture.nativeElement.textContent).toContain('Duración: 4 h 30 min');
    expect(fixture.nativeElement.textContent).toContain('Oxitocina');

    const cards = fixture.nativeElement.querySelectorAll('ol > li');
    expect(cards[0].classList).toContain('border-l-gray-400');
    expect(cards[1].classList).toContain('border-l-red-500');
    expect(cards[2].classList).toContain('border-l-blue-500');
  });

  it('shows a loading state and an empty reproductive history', async () => {
    const response = new Subject<SowHistory>();
    sowService.getSowHistory.mockReturnValue(response);
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    fixture.detectChanges();
    expect(fixture.nativeElement.querySelector('[role="status"]')?.textContent).toContain('Cargando');

    response.next({ ...historyPayload, events: [historyPayload.events[0]] });
    response.complete();
    await fixture.whenStable();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('No hay eventos reproductivos registrados');
  });

  it('shows the inactive sow message returned by the server', async () => {
    sowService.getSowHistory.mockReturnValue(throwError(() => new HttpErrorResponse({
      status: 404,
      error: { error: 'Esta cerda ya no se encuentra en la granja' },
    })));
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('[role="alert"]')?.textContent).toContain(
      'Esta cerda ya no se encuentra en la granja',
    );
  });

  it('shows a not-found message for unknown codes', async () => {
    sowService.getSowHistory.mockReturnValue(throwError(() => new HttpErrorResponse({ status: 404 })));
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-404';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('[role="alert"]')?.textContent).toContain(
      'No se encontró una cerda con ese código.',
    );
  });
});
