import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import {
  BoarRemoval,
  RemovalReason,
  RemovalType,
  UpdateBoarRemovalRequest,
} from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { toISODate } from '../../../shared/utils/date';
import { REMOVAL_REASON_LABELS } from '../removal-reason';
import { REMOVAL_TYPE_LABELS } from '../removal-type';

@Component({
  selector: 'app-edit-boar-removal-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, DatePicker],
  styleUrl: './edit-boar-removal-modal.css',
  templateUrl: './edit-boar-removal-modal.html',
})
export class EditBoarRemovalModal {
  readonly open = input(false);
  readonly removal = input<BoarRemoval | null>(null);
  readonly boarCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly updated = output<BoarRemoval>();

  private readonly removals = inject(BoarRemovalsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly today = toISODate(new Date());
  protected readonly typeOptions: DropdownOption[] = (
    Object.keys(REMOVAL_TYPE_LABELS) as RemovalType[]
  ).map((type) => ({ value: type, label: REMOVAL_TYPE_LABELS[type] }));
  protected readonly reasonOptions: DropdownOption[] = (
    Object.keys(REMOVAL_REASON_LABELS) as RemovalReason[]
  ).map((reason) => ({ value: reason, label: REMOVAL_REASON_LABELS[reason] }));

  protected readonly form = this.formBuilder.nonNullable.group({
    removal_date: ['', [Validators.required]],
    type: ['', [Validators.required]],
    reason: ['', [Validators.required]],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      const current = this.removal();
      if (current && this.open()) {
        this.form.reset({
          removal_date: current.removal_date,
          type: current.type,
          reason: current.reason,
          note: current.note ?? '',
        });
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.removal();
    if (!current) {
      return;
    }

    const { removal_date, type, reason, note } = this.form.getRawValue();
    const request: UpdateBoarRemovalRequest = {};
    if (removal_date !== current.removal_date) {
      request.removal_date = removal_date;
    }
    if (type !== current.type) {
      request.type = type as RemovalType;
    }
    if (reason !== current.reason) {
      request.reason = reason as RemovalReason;
    }
    if (note !== (current.note ?? '')) {
      request.note = note;
    }

    if (Object.keys(request).length === 0) {
      this.updated.emit(current);
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.removals.updateBoarRemoval(current.id, request));
      this.updated.emit(current);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 404) {
        return 'La baja no existe';
      }
    }
    return 'No se pudo actualizar la baja';
  }
}
