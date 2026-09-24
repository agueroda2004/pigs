import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { OperatorsService } from '../../../core/operators/operators.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-create-operator-modal',
  imports: [ReactiveFormsModule, Modal],
  styleUrl: './create-operator-modal.css',
  templateUrl: './create-operator-modal.html',
})
export class CreateOperatorModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly operators = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);

  protected readonly form = this.formBuilder.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(100)]],
  });

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    const { name } = this.form.getRawValue();

    try {
      await firstValueFrom(this.operators.createOperator({ name }));
      this.form.reset();
      this.created.emit(name);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El nombre del operador ya existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo crear el operador';
  }
}
