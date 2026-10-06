import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { CreateWeagingRequest } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';

@Component({
  selector: 'app-register-weaging-modal',
  imports: [ReactiveFormsModule, Modal, SearchDropdown, DatePicker],
  styleUrl: './register-weaging-modal.css',
  templateUrl: './register-weaging-modal.html',
})
export class RegisterWeagingModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly weagings = inject(WeagingsService);
  private readonly sows = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());

  protected readonly form = this.formBuilder.nonNullable.group({
    sow_id: ['', [Validators.required]],
    weaging_date: [this.today, [Validators.required]],
    quantity: [1, [Validators.required, Validators.min(1)]],
    total_weight: [null as number | null, [Validators.min(0.01)]],
    destination: ['', [Validators.maxLength(100)]],
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

    const { sow_id, weaging_date, quantity, total_weight, destination, note } =
      this.form.getRawValue();
    const request: CreateWeagingRequest = {
      sow_id,
      weaging_date,
      quantity,
    };
    if (total_weight !== null) {
      request.total_weight = total_weight;
    }
    if (destination) {
      request.destination = destination;
    }
    if (note) {
      request.note = note;
    }

    const sowCode = this.sowOptions().find((option) => option.value === sow_id)?.label ?? '';

    try {
      await firstValueFrom(this.weagings.createWeaging(request));
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
      weaging_date: this.today,
      quantity: 1,
      total_weight: null,
      destination: '',
      note: '',
    });
  }

  private async loadOptions(): Promise<void> {
    try {
      const sows = await firstValueFrom(this.sows.listSows({ state: 'Lactando' }));
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
        return 'La cerda no está lactando, la cantidad no coincide con los lechones actuales o el parto ya tiene un destete';
      }
    }
    return 'No se pudo registrar el destete';
  }
}
