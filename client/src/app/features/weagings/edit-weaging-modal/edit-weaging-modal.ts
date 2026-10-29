import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { UpdateWeagingRequest, Weaging } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-edit-weaging-modal',
  imports: [ReactiveFormsModule, Modal],
  styleUrl: './edit-weaging-modal.css',
  templateUrl: './edit-weaging-modal.html',
})
export class EditWeagingModal {
  readonly open = input(false);
  readonly weaging = input<Weaging | null>(null);
  readonly sowCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly updated = output<void>();

  private readonly weagings = inject(WeagingsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);

  protected readonly form = this.formBuilder.nonNullable.group({
    total_weight: [null as number | null, [Validators.min(0.01), Validators.max(200)]],
    destination: ['', [Validators.maxLength(100)]],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      const current = this.weaging();
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

    const current = this.weaging();
    if (!current) {
      return;
    }

    const request = this.buildRequest(current);
    if (Object.keys(request).length === 0) {
      this.closed.emit();
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.weagings.updateWeaging(current.id, request));
      this.updated.emit();
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private buildRequest(current: Weaging): UpdateWeagingRequest {
    const { total_weight, destination, note } = this.form.getRawValue();
    const request: UpdateWeagingRequest = {};

    if (total_weight === null) {
      if (current.total_weight !== null) {
        request.clear_total_weight = true;
      }
    } else if (total_weight !== current.total_weight) {
      request.total_weight = total_weight;
    }

    const trimmedDestination = destination.trim();
    if (trimmedDestination !== (current.destination ?? '')) {
      request.destination = trimmedDestination;
    }

    const trimmedNote = note.trim();
    if (trimmedNote !== (current.note ?? '')) {
      request.note = trimmedNote;
    }

    return request;
  }

  private prefill(current: Weaging): void {
    this.form.reset({
      total_weight: current.total_weight,
      destination: current.destination ?? '',
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
        return 'El destete no existe';
      }
    }
    return 'No se pudo actualizar el destete';
  }
}
