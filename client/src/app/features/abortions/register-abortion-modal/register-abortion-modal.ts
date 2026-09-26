import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AbortionCause, CreateAbortionRequest } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';
import { ABORTION_CAUSE_LABELS } from '../abortion-cause';

@Component({
  selector: 'app-register-abortion-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker],
  styleUrl: './register-abortion-modal.css',
  templateUrl: './register-abortion-modal.html',
})
export class RegisterAbortionModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly abortions = inject(AbortionsService);
  private readonly sows = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());
  protected readonly causeOptions: DropdownOption[] = (
    Object.keys(ABORTION_CAUSE_LABELS) as AbortionCause[]
  ).map((cause) => ({ value: cause, label: ABORTION_CAUSE_LABELS[cause] }));

  protected readonly form = this.formBuilder.nonNullable.group({
    sow_id: ['', [Validators.required]],
    abortion_date: [toISODate(new Date()), [Validators.required]],
    cause: ['', [Validators.required]],
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

    const { sow_id, abortion_date, cause, note } = this.form.getRawValue();
    const request: CreateAbortionRequest = {
      sow_id,
      abortion_date,
      cause: cause as AbortionCause,
    };
    if (note) {
      request.note = note;
    }

    const sowCode = this.sowOptions().find((option) => option.value === sow_id)?.label ?? '';

    try {
      await firstValueFrom(this.abortions.createAbortion(request));
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
      abortion_date: this.today,
      cause: '',
      note: '',
    });
  }

  private async loadOptions(): Promise<void> {
    try {
      const sows = await firstValueFrom(this.sows.listSows({ state: 'Gestando' }));
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
    } catch {
      this.notifications.error('No se pudieron cargar las cerdas gestando');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 409) {
        return 'La cerda no está en un estado válido para el aborto';
      }
    }
    return 'No se pudo registrar el aborto';
  }
}
