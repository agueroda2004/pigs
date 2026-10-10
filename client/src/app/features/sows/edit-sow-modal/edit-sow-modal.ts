import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { Sow, SowOrigin, UpdateSowRequest } from '../../../core/sows/sow.models';
import { SowsService } from '../../../core/sows/sows.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { toISODate } from '../../../shared/utils/date';
import { SOW_STATE_CLASSES, SOW_STATE_LABELS } from '../sow-state';

@Component({
  selector: 'app-edit-sow-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, DatePicker],
  styleUrl: './edit-sow-modal.css',
  templateUrl: './edit-sow-modal.html',
})
export class EditSowModal {
  readonly open = input(false);
  readonly sow = input<Sow | null>(null);
  readonly closed = output<void>();
  readonly updated = output<string>();

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

  protected readonly stateLabel = computed(() => {
    const current = this.sow();
    return current ? SOW_STATE_LABELS[current.state] : '';
  });

  protected readonly stateClass = computed(() => {
    const current = this.sow();
    return current ? SOW_STATE_CLASSES[current.state] : '';
  });

  protected readonly canEditDates = computed(() => this.sow()?.active ?? false);
  protected readonly canEditParity = computed(() => this.sow()?.state === 'Viva');

  protected readonly form = this.formBuilder.nonNullable.group({
    code: ['', [Validators.required, Validators.maxLength(50)]],
    breed_id: ['', [Validators.required]],
    origin: ['', [Validators.required]],
    parity: [0, [Validators.required, Validators.min(0)]],
    location: ['', [Validators.maxLength(100)]],
    entry_date: ['', [Validators.required]],
    birth_date: [''],
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      const current = this.sow();
      if (current && this.open()) {
        this.form.reset({
          code: current.code,
          breed_id: current.breed_id,
          origin: current.origin,
          parity: current.parity,
          location: current.location ?? '',
          entry_date: current.entry_date,
          birth_date: current.birth_date ?? '',
          note: current.note ?? '',
        });
        this.applyDateEditability(current.active);
        this.applyParityEditability(current.state === 'Viva');
        void this.loadBreeds();
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.sow();
    if (!current) {
      return;
    }

    const { code, breed_id, origin, parity, location, entry_date, birth_date, note } =
      this.form.getRawValue();
    const request: UpdateSowRequest = {};

    if (code !== current.code) {
      request.code = code;
    }
    if (breed_id !== current.breed_id) {
      request.breed_id = breed_id;
    }
    if (origin !== current.origin) {
      request.origin = origin as SowOrigin;
    }
    if (this.canEditParity() && parity !== current.parity) {
      request.parity = parity;
    }
    if (this.canEditDates() && entry_date !== current.entry_date) {
      request.entry_date = entry_date;
    }
    if (this.canEditDates() && birth_date !== (current.birth_date ?? '')) {
      request.birth_date = birth_date;
    }
    if (location !== (current.location ?? '')) {
      request.location = location;
    }
    if (note !== (current.note ?? '')) {
      request.note = note;
    }

    if (Object.keys(request).length === 0) {
      this.updated.emit(current.code);
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.sows.updateSow(current.id, request));
      this.updated.emit(request.code ?? current.code);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private applyDateEditability(editable: boolean): void {
    const action = editable ? 'enable' : 'disable';
    this.form.controls.entry_date[action]({ emitEvent: false });
    this.form.controls.birth_date[action]({ emitEvent: false });
  }

  private applyParityEditability(editable: boolean): void {
    const action = editable ? 'enable' : 'disable';
    this.form.controls.parity[action]({ emitEvent: false });
  }

  private async loadBreeds(): Promise<void> {
    try {
      const breeds = await firstValueFrom(this.breeds.listBreedDropdown(false));
      this.breedOptions.set(
        breeds.map((breed) => ({
          value: breed.id,
          label: breed.active ? breed.name : `${breed.name} (inactiva)`,
        })),
      );
    } catch {
      this.notifications.error('No se pudieron cargar las razas');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El código de la cerda ya existe';
      }
      if (error.status === 404) {
        return 'La cerda no existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo actualizar la cerda';
  }
}
