import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import {
  CreatePigletDeathRequest,
  PigletDeathCause,
  Turn,
} from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';
import { PIGLET_DEATH_CAUSE_LABELS } from '../piglet-death-cause';
import { PIGLET_DEATH_TURN_LABELS } from '../piglet-death-turn';

@Component({
  selector: 'app-register-piglet-death-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker],
  styleUrl: './register-piglet-death-modal.css',
  templateUrl: './register-piglet-death-modal.html',
})
export class RegisterPigletDeathModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly pigletDeaths = inject(PigletDeathsService);
  private readonly sows = inject(SowsService);
  private readonly operators = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly operatorOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());
  protected readonly causeOptions: DropdownOption[] = (
    Object.keys(PIGLET_DEATH_CAUSE_LABELS) as PigletDeathCause[]
  ).map((cause) => ({ value: cause, label: PIGLET_DEATH_CAUSE_LABELS[cause] }));
  protected readonly turnOptions: DropdownOption[] = (
    Object.keys(PIGLET_DEATH_TURN_LABELS) as Turn[]
  ).map((turn) => ({ value: turn, label: PIGLET_DEATH_TURN_LABELS[turn] }));

  protected readonly form = this.formBuilder.nonNullable.group({
    sow_id: ['', [Validators.required]],
    operator_id: ['', [Validators.required]],
    death_date: [this.today, [Validators.required]],
    quantity: [1, [Validators.required, Validators.min(1)]],
    weight: [null as number | null, [Validators.min(0.01)]],
    cause: ['', [Validators.required]],
    turn: ['', [Validators.required]],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      if (this.open()) {
        void this.loadOptions();
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    const { sow_id, operator_id, death_date, quantity, weight, cause, turn, note } =
      this.form.getRawValue();
    const request: CreatePigletDeathRequest = {
      sow_id,
      operator_id,
      death_date,
      quantity,
      cause: cause as PigletDeathCause,
      turn: turn as Turn,
    };
    if (weight !== null) {
      request.weight = weight;
    }
    if (note) {
      request.note = note;
    }

    const sowCode = this.sowOptions().find((option) => option.value === sow_id)?.label ?? '';

    try {
      await firstValueFrom(this.pigletDeaths.createPigletDeath(request));
      this.reset();
      this.created.emit(sowCode);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private reset(): void {
    this.form.reset({
      sow_id: '',
      operator_id: '',
      death_date: this.today,
      quantity: 1,
      weight: null,
      cause: '',
      turn: '',
      note: '',
    });
  }

  private async loadOptions(): Promise<void> {
    try {
      const [sows, operators] = await Promise.all([
        firstValueFrom(this.sows.listSows({ state: 'Lactando' })),
        firstValueFrom(this.operators.listOperators()),
      ]);
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
      this.operatorOptions.set(
        operators
          .filter((operator) => operator.active)
          .map((operator) => ({ value: operator.id, label: operator.name })),
      );
    } catch {
      this.notifications.error('No se pudieron cargar los datos de la muerte');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 409) {
        return 'La cerda no está lactando o la cantidad supera los lechones actuales';
      }
    }
    return 'No se pudo registrar la muerte';
  }
}
