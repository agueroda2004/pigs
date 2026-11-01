import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, Subject, throwError } from 'rxjs';
import { vi } from 'vitest';

import { SowHistory } from '../../../core/sows/sow-history.models';
import { SowsService } from '../../../core/sows/sows.service';
import { SowHistoryPage } from './sow-history-page';

const historyPayload: SowHistory = {
  sow: {
    id: 'sow-1', code: 'C-001', location: 'Corral A', active: true, entry_date: '2026-01-10',
    birth_date: null, note: null, state: 'Viva', origin: 'Propio', parity: 2, breed_id: 'breed-1',
  },
  events: [
    { type: 'sow_entry', date: '2026-01-10T00:00:00Z', data: {
      id: 'sow-1', code: 'C-001', location: 'Corral A', active: true, entry_date: '2026-01-10',
      birth_date: null, note: null, state: 'Viva', origin: 'Propio', parity: 2, breed_id: 'breed-1',
    } },
    { type: 'abortion', date: '2026-02-10T00:00:00Z', data: {
      id: 'abortion-1', sow_id: 'sow-1', sow_code: 'C-001', service_id: 'service-1',
      abortion_date: '2026-02-10', cause: 'Infeccioso', note: null,
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
