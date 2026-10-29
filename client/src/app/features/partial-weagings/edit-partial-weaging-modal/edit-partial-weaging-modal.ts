import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PartialWeaging,
  UpdatePartialWeagingRequest,
} from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { Modal } from '../../../shared/ui/modal/modal';
import { PARTIAL_WEAGING_TYPE_LABELS } from '../partial-weaging-type';

@Component({
  selector: 'app-edit-partial-weaging-modal',
  imports: [ReactiveFormsModule, Modal],
  styleUrl: './edit-partial-weaging-modal.css',
  templateUrl: './edit-partial-weaging-modal.html',
})
export class EditPartialWeagingModal {
  readonly open = input(false);
  readonly weaging = input<PartialWeaging | null>(null);
  readonly sowCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly updated = output<void>();

  private readonly partialWeagings = inject(PartialWeagingsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly typeLabel = computed(() => {
    const current = this.weaging();
    return current ? PARTIAL_WEAGING_TYPE_LABELS[current.type] : '';
  });

  protected readonly form = this.formBuilder.nonNullable.group({
    total_weight: [null as number | null, [Validators.min(0.01), Validators.max(200)]],
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
      await firstValueFrom(this.partialWeagings.updatePartialWeaging(current.id, request));
      this.updated.emit();
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private buildRequest(current: PartialWeaging): UpdatePartialWeagingRequest {
    const { total_weight, note } = this.form.getRawValue();
    const request: UpdatePartialWeagingRequest = {};

    if (total_weight === null) {
      if (current.total_weight !== null) {
        request.clear_total_weight = true;
      }
    } else if (total_weight !== current.total_weight) {
      request.total_weight = total_weight;
    }

    const trimmedNote = note.trim();
    if (trimmedNote !== (current.note ?? '')) {
      request.note = trimmedNote;
    }

    return request;
  }

  private prefill(current: PartialWeaging): void {
    this.form.reset({
      total_weight: current.total_weight,
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
        return 'El destete parcial no existe';
      }
    }
    return 'No se pudo actualizar el destete parcial';
  }
}
