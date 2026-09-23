import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { Boar, BoarOrigin, UpdateBoarRequest } from '../../../core/boars/boar.models';
import { BoarsService } from '../../../core/boars/boars.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { toISODate } from '../../../shared/utils/date';
import { BOAR_STATE_CLASSES, BOAR_STATE_LABELS } from '../boar-state';

@Component({
  selector: 'app-edit-boar-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, DatePicker],
  styleUrl: './edit-boar-modal.css',
  templateUrl: './edit-boar-modal.html',
})
export class EditBoarModal {
  readonly open = input(false);
  readonly boar = input<Boar | null>(null);
  readonly closed = output<void>();
  readonly updated = output<Boar>();

  private readonly boars = inject(BoarsService);
  private readonly breeds = inject(BreedsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly breedOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());
  protected readonly originOptions: DropdownOption[] = [
    { value: 'Propio', label: 'Propio' },
    { value: 'Externo', label: 'Externo' },
  ];

  protected readonly stateLabel = computed(() => {
    const current = this.boar();
    return current ? BOAR_STATE_LABELS[current.state] : '';
  });

  protected readonly stateClass = computed(() => {
    const current = this.boar();
    return current ? BOAR_STATE_CLASSES[current.state] : '';
  });

  protected readonly form = this.formBuilder.nonNullable.group({
    code: ['', [Validators.required, Validators.maxLength(50)]],
    breed_id: ['', [Validators.required]],
    origin: ['', [Validators.required]],
    location: ['', [Validators.maxLength(100)]],
    active: [true],
    entry_date: ['', [Validators.required]],
    birth_date: [''],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      const current = this.boar();
      if (current && this.open()) {
        this.form.reset({
          code: current.code,
          breed_id: current.breed_id,
          origin: current.origin,
          location: current.location ?? '',
          active: current.active,
          entry_date: current.entry_date,
          birth_date: current.birth_date ?? '',
          note: current.note ?? '',
        });
        void this.loadBreeds();
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.boar();
    if (!current) {
      return;
    }

    const { code, breed_id, origin, location, active, entry_date, birth_date, note } =
      this.form.getRawValue();
    const request: UpdateBoarRequest = {};

    if (code !== current.code) {
      request.code = code;
    }
    if (breed_id !== current.breed_id) {
      request.breed_id = breed_id;
    }
    if (origin !== current.origin) {
      request.origin = origin as BoarOrigin;
    }
    if (entry_date !== current.entry_date) {
      request.entry_date = entry_date;
    }
    if (active !== current.active) {
      request.active = active;
    }
    if (location !== (current.location ?? '')) {
      request.location = location;
    }
    if (note !== (current.note ?? '')) {
      request.note = note;
    }
    if (birth_date !== (current.birth_date ?? '')) {
      request.birth_date = birth_date;
    }

    if (Object.keys(request).length === 0) {
      this.updated.emit(current);
      return;
    }

    this.loading.set(true);

    try {
      const updated = await firstValueFrom(this.boars.updateBoar(current.id, request));
      this.updated.emit(updated);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private async loadBreeds(): Promise<void> {
    try {
      const breeds = await firstValueFrom(this.breeds.listBreeds());
      this.breedOptions.set(breeds.map((breed) => ({ value: breed.id, label: breed.name })));
    } catch {
      this.notifications.error('No se pudieron cargar las razas');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El código del verraco ya existe';
      }
      if (error.status === 404) {
        return 'El verraco no existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo actualizar el verraco';
  }
}
