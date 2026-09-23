import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { Breed, UpdateBreedRequest } from '../../../core/breeds/breed.models';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-edit-breed-modal',
  imports: [ReactiveFormsModule, Modal],
  styleUrl: './edit-breed-modal.css',
  templateUrl: './edit-breed-modal.html',
})
export class EditBreedModal {
  readonly open = input(false);
  readonly breed = input<Breed | null>(null);
  readonly closed = output<void>();
  readonly updated = output<Breed>();

  private readonly breeds = inject(BreedsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);

  protected readonly form = this.formBuilder.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(100)]],
    active: [true],
  });

  constructor() {
    effect(() => {
      const current = this.breed();
      if (current && this.open()) {
        this.form.reset({ name: current.name, active: current.active });
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.breed();
    if (!current) {
      return;
    }

    const { name, active } = this.form.getRawValue();
    const request: UpdateBreedRequest = {};
    if (name !== current.name) {
      request.name = name;
    }
    if (active !== current.active) {
      request.active = active;
    }

    if (request.name === undefined && request.active === undefined) {
      this.updated.emit(current);
      return;
    }

    this.loading.set(true);

    try {
      const updated = await firstValueFrom(this.breeds.updateBreed(current.id, request));
      this.updated.emit(updated);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El nombre de la raza ya existe';
      }
      if (error.status === 404) {
        return 'La raza no existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo actualizar la raza';
  }
}
