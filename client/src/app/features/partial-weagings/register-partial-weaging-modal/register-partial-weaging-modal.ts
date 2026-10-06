import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import {
  CreatePartialWeagingRequest,
  PartialWeagingType,
} from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';
import { PARTIAL_WEAGING_TYPE_LABELS } from '../partial-weaging-type';

@Component({
  selector: 'app-register-partial-weaging-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker],
  styleUrl: './register-partial-weaging-modal.css',
  templateUrl: './register-partial-weaging-modal.html',
})
export class RegisterPartialWeagingModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly partialWeagings = inject(PartialWeagingsService);
  private readonly sows = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());
  protected readonly typeOptions: DropdownOption[] = (
    Object.keys(PARTIAL_WEAGING_TYPE_LABELS) as PartialWeagingType[]
  ).map((type) => ({ value: type, label: PARTIAL_WEAGING_TYPE_LABELS[type] }));

  protected readonly form = this.formBuilder.nonNullable.group({
    sow_id: ['', [Validators.required]],
    weaging_date: [this.today, [Validators.required]],
    quantity: [1, [Validators.required, Validators.min(1)]],
    type: ['', [Validators.required]],
    total_weight: [null as number | null, [Validators.min(0.01)]],
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

    const { sow_id, weaging_date, quantity, type, total_weight, note } = this.form.getRawValue();
    const request: CreatePartialWeagingRequest = {
      sow_id,
      weaging_date,
      quantity,
      type: type as PartialWeagingType,
    };
    if (total_weight !== null) {
      request.total_weight = total_weight;
    }
    if (note) {
      request.note = note;
    }

    const sowCode = this.sowOptions().find((option) => option.value === sow_id)?.label ?? '';

    try {
      await firstValueFrom(this.partialWeagings.createPartialWeaging(request));
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
      type: '',
      total_weight: null,
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
        return 'La cerda no está lactando o la cantidad supera los lechones actuales';
      }
    }
    return 'No se pudo registrar el destete';
  }
}
