import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
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
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { PIGLET_DEATH_CAUSE_LABELS } from '../piglet-death-cause';
import { PIGLET_DEATH_TURN_LABELS } from '../piglet-death-turn';

@Component({
  selector: 'app-edit-piglet-death-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown],
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
  protected readonly causeOptions: DropdownOption[] = (
    Object.keys(PIGLET_DEATH_CAUSE_LABELS) as PigletDeathCause[]
  ).map((cause) => ({ value: cause, label: PIGLET_DEATH_CAUSE_LABELS[cause] }));
  protected readonly turnOptions: DropdownOption[] = (
    Object.keys(PIGLET_DEATH_TURN_LABELS) as Turn[]
  ).map((turn) => ({ value: turn, label: PIGLET_DEATH_TURN_LABELS[turn] }));

  protected readonly form = this.formBuilder.nonNullable.group({
    operator_id: ['', [Validators.required]],
    weight: [null as number | null, [Validators.min(0.01), Validators.max(20)]],
    cause: ['', [Validators.required]],
    turn: ['', [Validators.required]],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      const current = this.death();
      if (current && this.open()) {
        this.prefill(current);
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

    const request: UpdatePigletDeathRequest = {
      operator_id,
      cause: cause as PigletDeathCause,
      turn: turn as Turn,
      note,
    };
    if (weight === null) {
      if (current.weight !== null) {
        request.clear_weight = true;
      }
    } else if (weight !== current.weight) {
      request.weight = weight;
    }

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

  // Prefills only the editable fields; the sow, date and quantity are read from
  // the death record and shown as read-only values in the template.
  private prefill(current: PigletDeath): void {
    this.form.reset({
      operator_id: current.operator_id,
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
