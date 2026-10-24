import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import {
  PigletDeath,
  PigletDeathCause,
  Turn,
  UpdatePigletDeathRequest,
} from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';
import { PIGLET_DEATH_CAUSE_LABELS } from '../piglet-death-cause';
import { PIGLET_DEATH_TURN_LABELS } from '../piglet-death-turn';

@Component({
  selector: 'app-edit-piglet-death-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker],
  styleUrl: './edit-piglet-death-modal.css',
  templateUrl: './edit-piglet-death-modal.html',
})
export class EditPigletDeathModal {
  readonly open = input(false);
  readonly death = input<PigletDeath | null>(null);
  readonly sowCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly updated = output<void>();

  private readonly pigletDeaths = inject(PigletDeathsService);
  private readonly operators = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly operatorOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());
  protected readonly causeOptions: DropdownOption[] = (
    Object.keys(PIGLET_DEATH_CAUSE_LABELS) as PigletDeathCause[]
  ).map((cause) => ({ value: cause, label: PIGLET_DEATH_CAUSE_LABELS[cause] }));
  protected readonly turnOptions: DropdownOption[] = (
    Object.keys(PIGLET_DEATH_TURN_LABELS) as Turn[]
  ).map((turn) => ({ value: turn, label: PIGLET_DEATH_TURN_LABELS[turn] }));

  protected readonly sowOptions = computed<DropdownOption[]>(() => {
    const current = this.death();
    return current ? [{ value: current.sow_id, label: this.sowCode() ?? 'sin identificar' }] : [];
  });

  protected readonly form = this.formBuilder.nonNullable.group({
    sow_id: ['', [Validators.required]],
    operator_id: ['', [Validators.required]],
    death_date: ['', [Validators.required]],
    quantity: [1, [Validators.required, Validators.min(1), Validators.max(10)]],
    weight: [
      null as number | null,
      [Validators.required, Validators.min(0.01), Validators.max(20)],
    ],
    cause: ['', [Validators.required]],
    turn: ['', [Validators.required]],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    this.lockImmutableFields();
    effect(() => {
      const current = this.death();
      if (current && this.open()) {
        this.prefill(current);
        this.lockImmutableFields();
        void this.loadOperators();
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.death();
    if (!current) {
      return;
    }

    const { operator_id, weight, cause, turn, note } = this.form.getRawValue();
    if (weight === null) {
      this.form.markAllAsTouched();
      return;
    }

    const request: UpdatePigletDeathRequest = {
      operator_id,
      weight,
      cause: cause as PigletDeathCause,
      turn: turn as Turn,
      note,
    };

    this.loading.set(true);
    try {
      await firstValueFrom(this.pigletDeaths.updatePigletDeath(current.id, request));
      this.updated.emit();
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  // Disables the sow, death date and quantity, which are fixed after creation.
  // The disabled values stay visible in the form but are never sent on update.
  private lockImmutableFields(): void {
    for (const name of ['sow_id', 'death_date', 'quantity'] as const) {
      this.form.controls[name].disable({ emitEvent: false });
    }
  }

  private prefill(current: PigletDeath): void {
    this.form.reset({
      sow_id: current.sow_id,
      operator_id: current.operator_id,
      death_date: current.death_date,
      quantity: current.quantity,
      weight: current.weight,
      cause: current.cause,
      turn: current.turn,
      note: current.note ?? '',
    });
  }

  private async loadOperators(): Promise<void> {
    try {
      const operators = await firstValueFrom(this.operators.listOperatorDropdown());
      this.operatorOptions.set(
        operators.map((operator) => ({
          value: operator.id,
          label: operator.active ? operator.name : `${operator.name} (inactivo)`,
        })),
      );
    } catch {
      this.notifications.error('No se pudieron cargar los operadores');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 404) {
        return 'La muerte no existe';
      }
    }
    return 'No se pudo actualizar la muerte';
  }
}
