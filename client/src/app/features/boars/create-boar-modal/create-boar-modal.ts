import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { BoarOrigin, CreateBoarRequest } from '../../../core/boars/boar.models';
import { BoarsService } from '../../../core/boars/boars.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { toISODate } from '../../../shared/utils/date';

@Component({
  selector: 'app-create-boar-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, DatePicker],
  styleUrl: './create-boar-modal.css',
  templateUrl: './create-boar-modal.html',
})
export class CreateBoarModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

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

  protected readonly form = this.formBuilder.nonNullable.group({
    code: ['', [Validators.required, Validators.maxLength(50)]],
    breed_id: ['', [Validators.required]],
    origin: ['', [Validators.required]],
    location: ['', [Validators.maxLength(100)]],
    entry_date: ['', [Validators.required]],
    birth_date: [''],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      if (this.open()) {
        void this.loadBreeds();
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    const { code, breed_id, origin, location, entry_date, birth_date, note } =
      this.form.getRawValue();
    const request: CreateBoarRequest = {
      code,
      entry_date,
      breed_id,
      origin: origin as BoarOrigin,
    };
    if (location) {
      request.location = location;
    }
    if (birth_date) {
      request.birth_date = birth_date;
    }
    if (note) {
      request.note = note;
    }

    try {
      await firstValueFrom(this.boars.createBoar(request));
      this.form.reset();
      this.created.emit(code);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private async loadBreeds(): Promise<void> {
    try {
      const options = await firstValueFrom(this.breeds.listBreedOptions());
      this.breedOptions.set(options.map((option) => ({ value: option.id, label: option.name })));
    } catch {
      this.notifications.error('No se pudieron cargar las razas');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El código del verraco ya existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo crear el verraco';
  }
}
