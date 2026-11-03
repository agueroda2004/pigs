import { HttpErrorResponse } from '@angular/common/http';
import { DatePipe, NgClass } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { PartialWeagingType } from '../../../core/partial-weagings/partial-weaging.models';
import { PigletDeathCause, Turn } from '../../../core/piglet-deaths/piglet-death.models';
import {
  SowHistory,
  SowHistoryEvent,
  SowHistoryFarrowing,
  SowHistoryPigletFostering,
} from '../../../core/sows/sow-history.models';
import { SowsService } from '../../../core/sows/sows.service';
import { PARTIAL_WEAGING_TYPE_LABELS } from '../../partial-weagings/partial-weaging-type';
import { PIGLET_DEATH_CAUSE_LABELS } from '../../piglet-deaths/piglet-death-cause';
import { PIGLET_DEATH_TURN_LABELS } from '../../piglet-deaths/piglet-death-turn';

const FOSTERING_ROLE_LABELS: Record<SowHistoryPigletFostering['role'], string> = {
  donor: 'Dona',
  receiver: 'Adopta',
};

@Component({
  selector: 'app-sow-history-page',
  imports: [DatePipe, NgClass, ReactiveFormsModule],
  styleUrl: './sow-history-page.css',
  templateUrl: './sow-history-page.html',
})
export class SowHistoryPage {
  private readonly sowsService = inject(SowsService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly history = signal<SowHistory | null>(null);
  protected readonly loading = signal(false);
  protected readonly error = signal('');
  protected readonly searchForm = this.formBuilder.nonNullable.group({
    code: ['', [Validators.required, Validators.maxLength(50)]],
  });

  protected async search(): Promise<void> {
    if (this.searchForm.invalid) {
      this.searchForm.controls.code.markAsTouched();
      return;
    }

    const code = this.searchForm.controls.code.value.trim();
    if (!code) return;

    this.loading.set(true);
    this.error.set('');
    this.history.set(null);
    try {
      this.history.set(await firstValueFrom(this.sowsService.getSowHistory(code)));
    } catch (cause) {
      const response = cause as HttpErrorResponse;
      const message = response.error?.error;
      if (response.status === 404 && message !== 'Esta cerda ya no se encuentra en la granja') {
        this.error.set('No se encontró una cerda con ese código.');
      } else {
        this.error.set(message || 'No se pudo cargar el historial. Inténtelo de nuevo.');
      }
    } finally {
      this.loading.set(false);
    }
  }

  protected eventTitle(event: SowHistoryEvent): string {
    const titles: Record<SowHistoryEvent['type'], string> = {
      sow_entry: 'Ingreso a la granja',
      service: 'Servicio',
      abortion: 'Aborto',
      farrowing: 'Parto',
      piglet_death: 'Muerte de lechones',
      piglet_fostering: 'Traslado de lechones',
      partial_weaging: 'Destete parcial',
      weaging: 'Destete',
    };
    return titles[event.type];
  }

  protected eventBorderClass(type: SowHistoryEvent['type']): string {
    switch (type) {
      case 'sow_entry':
        return 'border-l-4 border-l-gray-400';
      case 'service':
        return 'border-l-4 border-l-green-500';
      case 'abortion':
        return 'border-l-4 border-l-red-500';
      case 'farrowing':
        return 'border-l-4 border-l-blue-500';
      case 'piglet_death':
        return 'border-l-4 border-l-warning';
      case 'piglet_fostering':
        return 'border-l-4 border-l-purple-500';
      case 'partial_weaging':
        return 'border-l-4 border-l-teal-500';
      case 'weaging':
        return 'border-l-4 border-l-emerald-500';
      default:
        return 'border-l-4 border-l-gray-300';
    }
  }

  // fosteringRoleLabel translates the server role into its display label.
  // It falls back to the raw role value when the role is unknown.
  protected fosteringRoleLabel(role: SowHistoryPigletFostering['role']): string {
    return FOSTERING_ROLE_LABELS[role] ?? role;
  }

  // pigletDeathCauseLabel translates a server cause into its display label.
  // It falls back to the raw value when the cause is unknown.
  protected pigletDeathCauseLabel(cause: string): string {
    return PIGLET_DEATH_CAUSE_LABELS[cause as PigletDeathCause] ?? cause;
  }

  // pigletDeathTurnLabel translates a server turn into its display label.
  // It falls back to the raw value when the turn is unknown.
  protected pigletDeathTurnLabel(turn: string): string {
    return PIGLET_DEATH_TURN_LABELS[turn as Turn] ?? turn;
  }

  // partialWeagingTypeLabel translates a server type into its display label.
  // It falls back to the raw value when the type is unknown.
  protected partialWeagingTypeLabel(type: string): string {
    return PARTIAL_WEAGING_TYPE_LABELS[type as PartialWeagingType] ?? type;
  }

  protected farrowingTimeRange(farrowing: SowHistoryFarrowing): string | null {
    const { start_time, end_time } = farrowing;
    if (start_time && end_time) {
      return `${start_time} → ${end_time}`;
    }
    return start_time ?? end_time ?? null;
  }

  protected farrowingDuration(farrowing: SowHistoryFarrowing): string | null {
    const { start_time, end_time } = farrowing;
    if (!start_time || !end_time) {
      return null;
    }
    const start = toMinutes(start_time);
    const end = toMinutes(end_time);
    if (start === null || end === null) {
      return null;
    }
    const diff = end >= start ? end - start : end + 24 * 60 - start;
    const hours = Math.floor(diff / 60);
    const minutes = diff % 60;
    return minutes === 0 ? `${hours} h` : `${hours} h ${minutes} min`;
  }
}

// toMinutes converts an HH:MM string into its total minutes.
// It returns null when the value does not match the HH:MM format.
function toMinutes(value: string): number | null {
  const match = /^([01]\d|2[0-3]):([0-5]\d)$/.exec(value);
  if (!match) {
    return null;
  }
  return Number(match[1]) * 60 + Number(match[2]);
}
