import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import {
  CreatePigletFosteringRequest,
  PigletFostering,
  UpdatePigletFosteringRequest,
} from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';

export interface CreatedFostering {
  donor: string;
  receiver: string;
}

const MAX_QUANTITY = 30;

@Component({
  selector: 'app-register-piglet-fostering-modal',
  imports: [ReactiveFormsModule, Modal, SearchDropdown, DatePicker],
  styleUrl: './register-piglet-fostering-modal.css',
  templateUrl: './register-piglet-fostering-modal.html',
})
export class RegisterPigletFosteringModal {
  readonly open = input(false);
  readonly fostering = input<PigletFostering | null>(null);
  readonly donorSowCode = input<string | null>(null);
  readonly receiverSowCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly created = output<CreatedFostering>();
  readonly updated = output<void>();

  private readonly pigletFosterings = inject(PigletFosteringsService);
  private readonly sows = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly loadedSowOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());
  protected readonly maxQuantity = MAX_QUANTITY;
  protected readonly editing = computed(() => this.fostering() !== null);

  // sowOptions uses the selected record on edit mode and the lactating sows on create.
  // The edit mode only shows the two fixed sows so they can be displayed disabled.
  protected readonly sowOptions = computed<DropdownOption[]>(() => {
    const current = this.fostering();
    if (!current) {
      return this.loadedSowOptions();
    }
    return [
      { value: current.donor_sow_id, label: this.donorSowCode() ?? 'sin identificar' },
      { value: current.receiver_sow_id, label: this.receiverSowCode() ?? 'sin identificar' },
    ];
  });

  // sameSowValidator rejects a form whose donor and receiver match.
  // It only applies once both sides have a value.
  private readonly sameSowValidator = (group: AbstractControl): ValidationErrors | null => {
    const donor = group.get('donor_sow_id')?.value;
    const receiver = group.get('receiver_sow_id')?.value;
    if (donor && receiver && donor === receiver) {
      return { sameSow: true };
    }
    return null;
  };

  protected readonly form = this.formBuilder.nonNullable.group(
    {
      donor_sow_id: ['', [Validators.required]],
      receiver_sow_id: ['', [Validators.required]],
      movement_date: [this.today, [Validators.required]],
      quantity: [1, [Validators.required, Validators.min(1), Validators.max(MAX_QUANTITY)]],
      note: ['', [Validators.maxLength(500)]],
    },
    { validators: this.sameSowValidator },
  );

  constructor() {
    effect(() => {
      if (!this.open()) {
        return;
      }
      const current = this.fostering();
      if (current) {
        this.prefill(current);
        this.lockSowFields();
        return;
      }
      this.unlockSowFields();
      this.reset();
      void this.loadOptions();
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);
    try {
      const current = this.fostering();
      if (current) {
        await this.update(current);
      } else {
        await this.create();
      }
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  // create sends the selected sows with the date, quantity and optional note.
  // It emits the code of both sows so the page can build the success message.
  private async create(): Promise<void> {
    const { donor_sow_id, receiver_sow_id, movement_date, quantity, note } =
      this.form.getRawValue();
    const request: CreatePigletFosteringRequest = {
      donor_sow_id,
      receiver_sow_id,
      movement_date,
      quantity,
    };
    if (note) {
      request.note = note;
    }

    await firstValueFrom(this.pigletFosterings.createPigletFostering(request));
    this.reset();
    this.created.emit({
      donor: this.sowCode(donor_sow_id),
      receiver: this.sowCode(receiver_sow_id),
    });
  }

  // update sends only the editable fields, since the sows can never change.
  // An empty note is sent so the server can clear the stored one.
  private async update(current: PigletFostering): Promise<void> {
    const { movement_date, quantity, note } = this.form.getRawValue();
    const request: UpdatePigletFosteringRequest = {
      movement_date,
      quantity,
      note,
    };

    await firstValueFrom(this.pigletFosterings.updatePigletFostering(current.id, request));
    this.updated.emit();
  }

  private prefill(current: PigletFostering): void {
    this.form.reset({
      donor_sow_id: current.donor_sow_id,
      receiver_sow_id: current.receiver_sow_id,
      movement_date: current.movement_date,
      quantity: current.quantity,
      note: current.note ?? '',
    });
  }

  // lockSowFields disables both sow controls, which are fixed after creation.
  // The disabled values stay visible but are still read when submitting.
  private lockSowFields(): void {
    for (const name of ['donor_sow_id', 'receiver_sow_id'] as const) {
      this.form.controls[name].disable({ emitEvent: false });
    }
  }

  // unlockSowFields re-enables both sow controls for the create flow.
  private unlockSowFields(): void {
    for (const name of ['donor_sow_id', 'receiver_sow_id'] as const) {
      this.form.controls[name].enable({ emitEvent: false });
    }
  }

  private reset(): void {
    this.form.reset({
      donor_sow_id: '',
      receiver_sow_id: '',
      movement_date: this.today,
      quantity: 1,
      note: '',
    });
  }

  private sowCode(id: string): string {
    return this.sowOptions().find((option) => option.value === id)?.label ?? '';
  }

  private async loadOptions(): Promise<void> {
    try {
      const sows = await firstValueFrom(this.sows.listSowDropdown(true, ['Lactando']));
      this.loadedSowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
    } catch {
      this.notifications.error('No se pudieron cargar las cerdas');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 409) {
        return 'Alguna de las cerdas no está lactando o la cantidad supera los lechones actuales';
      }
      if (error.status === 404) {
        return 'El traslado no existe';
      }
    }
    return this.editing()
      ? 'No se pudo actualizar el traslado'
      : 'No se pudo registrar el traslado';
  }
}
