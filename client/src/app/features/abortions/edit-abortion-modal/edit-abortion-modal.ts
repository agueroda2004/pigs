import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import {
  Abortion,
  AbortionCause,
  UpdateAbortionRequest,
} from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { toISODate } from '../../../shared/utils/date';
import { ABORTION_CAUSE_LABELS } from '../abortion-cause';

@Component({
  selector: 'app-edit-abortion-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, DatePicker],
  styleUrl: './edit-abortion-modal.css',
  templateUrl: './edit-abortion-modal.html',
})
export class EditAbortionModal {
  readonly open = input(false);
  readonly abortion = input<Abortion | null>(null);
  readonly closed = output<void>();
  readonly updated = output<void>();

  private readonly abortions = inject(AbortionsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly today = toISODate(new Date());
  protected readonly causeOptions: DropdownOption[] = (
    Object.keys(ABORTION_CAUSE_LABELS) as AbortionCause[]
  ).map((cause) => ({ value: cause, label: ABORTION_CAUSE_LABELS[cause] }));

  protected readonly form = this.formBuilder.nonNullable.group({
    abortion_date: ['', [Validators.required]],
    cause: ['', [Validators.required]],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      const current = this.abortion();
      if (current && this.open()) {
        this.prefill(current);
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.abortion();
    if (!current) {
      return;
    }

    const { abortion_date, cause, note } = this.form.getRawValue();
    const request: UpdateAbortionRequest = {
      abortion_date,
      cause: cause as AbortionCause,
      note,
    };

    this.loading.set(true);

    try {
      await firstValueFrom(this.abortions.updateAbortion(current.id, request));
      this.updated.emit();
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private prefill(current: Abortion): void {
    this.form.reset({
      abortion_date: current.abortion_date,
      cause: current.cause,
      note: current.note ?? '',
    });
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 404) {
        return 'El aborto no existe';
      }
    }
    return 'No se pudo actualizar el aborto';
  }
}
