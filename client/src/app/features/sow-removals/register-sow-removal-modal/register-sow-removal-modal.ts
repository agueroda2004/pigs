import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import {
  CreateSowRemovalRequest,
  RemovalReason,
  RemovalType,
} from '../../../core/sow-removals/sow-removal.models';
import { SowRemovalsService } from '../../../core/sow-removals/sow-removals.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';
import { REMOVAL_REASON_LABELS } from '../removal-reason';
import { REMOVAL_TYPE_LABELS } from '../removal-type';

@Component({
  selector: 'app-register-sow-removal-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker],
  styleUrl: './register-sow-removal-modal.css',
  templateUrl: './register-sow-removal-modal.html',
})
export class RegisterSowRemovalModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly removals = inject(SowRemovalsService);
  private readonly sows = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());
  protected readonly typeOptions: DropdownOption[] = (
    Object.keys(REMOVAL_TYPE_LABELS) as RemovalType[]
  ).map((type) => ({ value: type, label: REMOVAL_TYPE_LABELS[type] }));
  protected readonly reasonOptions: DropdownOption[] = (
    Object.keys(REMOVAL_REASON_LABELS) as RemovalReason[]
  ).map((reason) => ({ value: reason, label: REMOVAL_REASON_LABELS[reason] }));

  protected readonly form = this.formBuilder.nonNullable.group({
    sow_id: ['', [Validators.required]],
    removal_date: [toISODate(new Date()), [Validators.required]],
    type: ['', [Validators.required]],
    reason: ['', [Validators.required]],
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

    const { sow_id, removal_date, type, reason, note } = this.form.getRawValue();
    const request: CreateSowRemovalRequest = {
      sow_id,
      removal_date,
      type: type as RemovalType,
      reason: reason as RemovalReason,
    };
    if (note) {
      request.note = note;
    }

    const sowCode = this.sowOptions().find((option) => option.value === sow_id)?.label ?? '';

    try {
      await firstValueFrom(this.removals.createSowRemoval(request));
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
      removal_date: this.today,
      type: '',
      reason: '',
      note: '',
    });
  }

  private async loadOptions(): Promise<void> {
    try {
      const sows = await firstValueFrom(this.sows.listSowOptions(true));
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
    } catch {
      this.notifications.error('No se pudieron cargar las cerdas disponibles');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 409) {
        return 'La cerda no está en un estado válido para la baja';
      }
    }
    return 'No se pudo registrar la baja';
  }
}
