import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, Subject, throwError } from 'rxjs';
import { vi } from 'vitest';

import { SowHistory, SowHistoryEvent } from '../../../core/sows/sow-history.models';
import { SowsService } from '../../../core/sows/sows.service';
import { SowHistoryPage } from './sow-history-page';

const sowEntryEvent: SowHistoryEvent = {
  type: 'sow_entry',
  date: '2026-01-10',
  data: {
    code: 'C-001',
    entry_date: '2026-01-10',
    birth_date: null,
    state: 'Viva',
    parity: 2,
    breed_name: 'Large White',
    age: '1 año',
  },
};

const serviceEvent: SowHistoryEvent = {
  type: 'service',
  date: '2026-01-20',
  data: {
    sow_code: 'C-001',
    note: null,
    state: 'Confirmado',
    location: 'Corral A',
    mounts: [
      {
        boar_code: 'B-001',
        operator_name: 'Op 1',
        mount_date: '2026-01-20',
        type: 'Natural',
      },
    ],
  },
};

const abortionEvent: SowHistoryEvent = {
  type: 'abortion',
  date: '2026-02-10',
  data: {
    sow_code: 'C-001',
    abortion_date: '2026-02-10',
    cause: 'Infeccioso',
    note: null,
  },
};

const farrowingEvent: SowHistoryEvent = {
  type: 'farrowing',
  date: '2026-03-01',
  data: {
    farrow_date: '2026-03-01',
    start_time: '22:00',
    end_time: '02:30',
    duration: '04:30',
    location: 'Corral A',
    live_born: 10,
    mummified: 1,
    stillborn: 1,
    litter_weight: 15.5,
    stillborn_weight: 2.5,
    is_manipulated: false,
    note: null,
    is_nurse: false,
    operators: [{ operator_name: 'Op 1' }],
    medications: [{ medication_name: 'Oxitocina', dose: 2, applied_by_name: 'Op 2' }],
  },
};

const pigletDeathEvent: SowHistoryEvent = {
  type: 'piglet_death',
  date: '2026-04-05',
  data: {
    operator_name: 'Op 3',
    death_date: '2026-04-05',
    quantity: 3,
    weight: 2.5,
    cause: 'Pata_abierta',
    turn: 'Madrugada_no_asistida',
    note: 'Muerte durante la noche',
  },
};

const fosteringReceiverEvent: SowHistoryEvent = {
  type: 'piglet_fostering',
  date: '2026-04-10',
  data: {
    sow_id: '11111111-1111-1111-1111-111111111111',
    role: 'receiver',
    movement_date: '2026-04-10',
    quantity: 3,
    note: 'Traslado recibido',
  },
};

const fosteringDonorEvent: SowHistoryEvent = {
  type: 'piglet_fostering',
  date: '2026-03-15',
  data: {
    sow_id: '11111111-1111-1111-1111-111111111111',
    role: 'donor',
    movement_date: '2026-03-15',
    quantity: -2,
    note: null,
  },
};

const partialWeagingEvent: SowHistoryEvent = {
  type: 'partial_weaging',
  date: '2026-04-15',
  data: {
    weaging_date: '2026-04-15',
    quantity: 4,
    total_weight: 42.5,
    type: 'Nodriza',
    note: 'Destete parcial de la camada',
  },
};

const weagingEvent: SowHistoryEvent = {
  type: 'weaging',
  date: '2026-04-20',
  data: {
    weaging_date: '2026-04-20',
    quantity: 5,
    total_weight: 55.5,
    note: 'Destete completo',
  },
};

const historyPayload: SowHistory = {
  sow: {
    code: 'C-001',
    entry_date: '2026-01-10',
    birth_date: null,
    state: 'Viva',
    parity: 2,
    breed_name: 'Large White',
    age: '1 año',
  },
  events: [pigletDeathEvent, farrowingEvent, abortionEvent, serviceEvent, sowEntryEvent],
  averages: {
    average_live_born: 10,
    average_stillborn: 1,
    average_mummified: 1,
    average_live_born_weight: 1.55,
    average_litter_weight: 15.5,
    average_stillborn_weight: 2.5,
    mortality_rate_percentage: 30,
    average_deaths_per_farrowing: 3,
    average_dead_piglet_weight: 2.5,
    average_weaned_piglets: 3,
    average_weaned_piglet_weight: 19.5,
    average_gestation_days: 114,
    average_weaning_age_days: 21,
    manipulated_farrowings: 0,
    total_farrowings: 1,
    manipulated_text: 'Esta cerda ha sido manipulada 0 partos de 1 totales',
    average_farrowing_duration: '04:30',
    average_farrowing_duration_formatted: '4 h 30 min',
  },
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
    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Cerda C-001');
    expect(text).toContain('Aborto');
    expect(text).toContain('Horario: 22:00 → 02:30');
    expect(text).toContain('Duración: 4 h 30 min');
    expect(text).toContain('Oxitocina');
    expect(text).toContain('Muerte de lechones');
    expect(text).toContain('Cantidad: 3');
    expect(text).toContain('Peso: 2.5 kg');
    expect(text).toContain('Pata abierta');
    expect(text).toContain('Madrugada no asistida');
    expect(text).toContain('Operador: Op 3');
    expect(text).toContain('Fecha: 2026-04-05');
    expect(text).toContain('Nota: Muerte durante la noche');

    const cards = fixture.nativeElement.querySelectorAll('ol > li') as NodeListOf<HTMLElement>;
    expect(cards[0].classList).toContain('border-l-warning');
    expect(cards[1].classList).toContain('border-l-blue-500');
    expect(cards[2].classList).toContain('border-l-red-500');
    expect(cards[3].classList).toContain('border-l-green-500');
    expect(cards[4].classList).toContain('border-l-gray-400');

    const titles = Array.from(cards).map((card: HTMLElement) =>
      card.querySelector('h3')?.textContent?.trim(),
    );
    expect(titles).toEqual([
      'Muerte de lechones',
      'Parto',
      'Aborto',
      'Servicio',
      'Ingreso a la granja',
    ]);
  });

  it('renders piglet deaths as a non-empty history and omits null note and weight', async () => {
    const deathWithoutNoteOrWeight: SowHistoryEvent = {
      type: 'piglet_death',
      date: '2026-04-06',
      data: {
        operator_name: 'Op 4',
        death_date: '2026-04-06',
        quantity: 1,
        weight: null,
        cause: 'Aplastado',
        turn: 'Mañana',
        note: null,
      },
    };
    sowService.getSowHistory.mockReturnValue(
      of({ ...historyPayload, events: [deathWithoutNoteOrWeight, sowEntryEvent] }),
    );
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Muerte de lechones');
    expect(text).toContain('Cantidad: 1');
    expect(text).not.toContain('No hay eventos reproductivos registrados');
    expect(text).not.toContain('Nota:');
    expect(text).not.toContain('Peso:');
  });

  it('renders donor and receiver fostering events with signed quantities', async () => {
    sowService.getSowHistory.mockReturnValue(
      of({
        ...historyPayload,
        events: [fosteringReceiverEvent, pigletDeathEvent, fosteringDonorEvent, sowEntryEvent],
      }),
    );
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Traslado de lechones');
    expect(text).toContain('Dona');
    expect(text).toContain('Adopta');
    expect(text).toContain('Cantidad: -2');
    expect(text).toContain('Cantidad: 3');
    expect(text).toContain('Fecha: 2026-04-10');
    expect(text).toContain('Fecha: 2026-03-15');
    expect(text).toContain('Nota: Traslado recibido');

    const cards = fixture.nativeElement.querySelectorAll('ol > li') as NodeListOf<HTMLElement>;
    expect(cards[0].classList).toContain('border-l-purple-500');
    expect(cards[1].classList).toContain('border-l-warning');
    expect(cards[2].classList).toContain('border-l-purple-500');
    expect(cards[3].classList).toContain('border-l-gray-400');

    const receiverCard = cards[0];
    const donorCard = cards[2];
    expect(receiverCard.textContent).toContain('Adopta');
    expect(receiverCard.textContent).toContain('Cantidad: 3');
    expect(receiverCard.textContent).toContain('Nota: Traslado recibido');
    expect(donorCard.textContent).toContain('Dona');
    expect(donorCard.textContent).toContain('Cantidad: -2');
    expect(donorCard.textContent).not.toContain('Nota:');

    const titles = Array.from(cards).map((card: HTMLElement) =>
      card.querySelector('h3')?.textContent?.trim(),
    );
    expect(titles).toEqual([
      'Traslado de lechones',
      'Muerte de lechones',
      'Traslado de lechones',
      'Ingreso a la granja',
    ]);
  });

  it('renders partial weaging events with type, quantity and optional weight', async () => {
    sowService.getSowHistory.mockReturnValue(
      of({
        ...historyPayload,
        events: [partialWeagingEvent, pigletDeathEvent, sowEntryEvent],
      }),
    );
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Destete parcial');
    expect(text).toContain('Tipo: Nodriza');
    expect(text).toContain('Cantidad: 4');
    expect(text).toContain('Fecha: 2026-04-15');
    expect(text).toContain('Peso total: 42.5 kg');
    expect(text).toContain('Nota: Destete parcial de la camada');
    expect(text).not.toContain('No hay eventos reproductivos registrados');

    const cards = fixture.nativeElement.querySelectorAll('ol > li') as NodeListOf<HTMLElement>;
    expect(cards[0].classList).toContain('border-l-teal-500');
    expect(cards[1].classList).toContain('border-l-warning');
    expect(cards[2].classList).toContain('border-l-gray-400');

    const titles = Array.from(cards).map((card: HTMLElement) =>
      card.querySelector('h3')?.textContent?.trim(),
    );
    expect(titles).toEqual(['Destete parcial', 'Muerte de lechones', 'Ingreso a la granja']);
  });

  it('omits the partial weaging weight and note when they are null', async () => {
    const weagingWithoutOptionalFields: SowHistoryEvent = {
      type: 'partial_weaging',
      date: '2026-04-16',
      data: {
        weaging_date: '2026-04-16',
        quantity: 2,
        total_weight: null,
        type: 'Baja_Viabilidad',
        note: null,
      },
    };
    sowService.getSowHistory.mockReturnValue(
      of({ ...historyPayload, events: [weagingWithoutOptionalFields, sowEntryEvent] }),
    );
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Tipo: Baja viabilidad');
    expect(text).not.toContain('Peso total:');
    expect(text).not.toContain('Nota:');
    expect(text).not.toContain('No hay eventos reproductivos registrados');
  });

  it('renders full weaging events with quantity and optional weight', async () => {
    sowService.getSowHistory.mockReturnValue(
      of({
        ...historyPayload,
        events: [weagingEvent, partialWeagingEvent, sowEntryEvent],
      }),
    );
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Destete');
    expect(text).toContain('Cantidad: 5');
    expect(text).toContain('Fecha: 2026-04-20');
    expect(text).toContain('Peso total: 55.5 kg');
    expect(text).toContain('Nota: Destete completo');
    expect(text).not.toContain('No hay eventos reproductivos registrados');

    const cards = fixture.nativeElement.querySelectorAll('ol > li') as NodeListOf<HTMLElement>;
    expect(cards[0].classList).toContain('border-l-emerald-500');
    expect(cards[1].classList).toContain('border-l-teal-500');
    expect(cards[2].classList).toContain('border-l-gray-400');

    const titles = Array.from(cards).map((card: HTMLElement) =>
      card.querySelector('h3')?.textContent?.trim(),
    );
    expect(titles).toEqual(['Destete', 'Destete parcial', 'Ingreso a la granja']);
  });

  it('omits the full weaging weight and note when they are null', async () => {
    const weagingWithoutOptionalFields: SowHistoryEvent = {
      type: 'weaging',
      date: '2026-04-21',
      data: {
        weaging_date: '2026-04-21',
        quantity: 2,
        total_weight: null,
        note: null,
      },
    };
    sowService.getSowHistory.mockReturnValue(
      of({ ...historyPayload, events: [weagingWithoutOptionalFields, sowEntryEvent] }),
    );
    const fixture = TestBed.createComponent(SowHistoryPage);
    fixture.detectChanges();
    const input = fixture.nativeElement.querySelector('#sow-history-code') as HTMLInputElement;
    input.value = 'C-001';
    input.dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('form').dispatchEvent(new Event('submit'));
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain('Peso total:');
    expect(text).not.toContain('Nota:');
    expect(text).not.toContain('No hay eventos reproductivos registrados');
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
    expect(fixture.nativeElement.querySelector('[role="status"]')?.textContent).toContain(
      'Cargando',
    );

    response.next({ ...historyPayload, events: [sowEntryEvent] });
    response.complete();
    await fixture.whenStable();
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('No hay eventos reproductivos registrados');
  });

  it('shows the inactive sow message returned by the server', async () => {
    sowService.getSowHistory.mockReturnValue(
      throwError(
        () =>
          new HttpErrorResponse({
            status: 404,
            error: { error: 'Esta cerda ya no se encuentra en la granja' },
          }),
      ),
    );
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
    sowService.getSowHistory.mockReturnValue(
      throwError(() => new HttpErrorResponse({ status: 404 })),
    );
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
