import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { CreatePigletFosteringRequest } from '../../../core/piglet-fosterings/piglet-fostering.models';
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

@Component({
  selector: 'app-register-piglet-fostering-modal',
  imports: [ReactiveFormsModule, Modal, SearchDropdown, DatePicker],
  styleUrl: './register-piglet-fostering-modal.css',
  templateUrl: './register-piglet-fostering-modal.html',
})
export class RegisterPigletFosteringModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<CreatedFostering>();

  private readonly pigletFosterings = inject(PigletFosteringsService);
  private readonly sows = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());

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
      quantity: [1, [Validators.required, Validators.min(1)]],
      note: ['', [Validators.maxLength(500)]],
    },
    { validators: this.sameSowValidator },
  );

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

    try {
      await firstValueFrom(this.pigletFosterings.createPigletFostering(request));
      this.reset();
      this.created.emit({
        donor: this.sowCode(donor_sow_id),
        receiver: this.sowCode(receiver_sow_id),
      });
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
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
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
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
    }
    return 'No se pudo registrar el traslado';
  }
}
