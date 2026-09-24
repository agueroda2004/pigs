import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { CreateSowRequest, SowOrigin } from '../../../core/sows/sow.models';
import { SowsService } from '../../../core/sows/sows.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { toISODate } from '../../../shared/utils/date';

@Component({
  selector: 'app-create-sow-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, DatePicker],
  styleUrl: './create-sow-modal.css',
  templateUrl: './create-sow-modal.html',
})
export class CreateSowModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly sows = inject(SowsService);
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
    parity: [0, [Validators.min(0)]],
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

    const { code, breed_id, origin, parity, location, entry_date, birth_date, note } =
      this.form.getRawValue();
    const request: CreateSowRequest = {
      code,
      entry_date,
      breed_id,
      origin: origin as SowOrigin,
      parity,
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
      await firstValueFrom(this.sows.createSow(request));
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
      const breeds = await firstValueFrom(this.breeds.listBreeds());
      this.breedOptions.set(breeds.map((breed) => ({ value: breed.id, label: breed.name })));
    } catch {
      this.notifications.error('No se pudieron cargar las razas');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El código de la cerda ya existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo crear la cerda';
  }
}
