import { HttpErrorResponse } from '@angular/common/http';
import { Component, computed, effect, inject, input, output, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { firstValueFrom, map } from 'rxjs';

import { UserRole } from '../../../core/auth/auth.models';
import { NotificationService } from '../../../core/notifications/notification.service';
import { UpdateUserRequest, User } from '../../../core/users/user.models';
import { UsersService } from '../../../core/users/users.service';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import {
  PASSWORD_MIN_LENGTH,
  PASSWORD_PATTERN,
  PasswordStrength,
  passwordStrength,
} from '../../../shared/utils/password-strength';

function passwordsMatch(group: AbstractControl): ValidationErrors | null {
  const password = group.get('password')?.value;
  if (!password) {
    return null;
  }
  const confirmPassword = group.get('confirmPassword')?.value;
  return password === confirmPassword ? null : { passwordMismatch: true };
}

@Component({
  selector: 'app-edit-user-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown],
  styleUrl: './edit-user-modal.css',
  templateUrl: './edit-user-modal.html',
})
export class EditUserModal {
  readonly open = input(false);
  readonly user = input<User | null>(null);
  readonly closed = output<void>();
  readonly updated = output<User>();

  private readonly users = inject(UsersService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly minLength = PASSWORD_MIN_LENGTH;
  protected readonly roleOptions: DropdownOption[] = [
    { value: 'User', label: 'Usuario' },
    { value: 'Admin', label: 'Administrador' },
  ];

  protected readonly showPassword = signal(false);
  protected readonly showConfirmPassword = signal(false);
  protected readonly loading = signal(false);

  protected readonly form = this.formBuilder.nonNullable.group(
    {
      name: ['', [Validators.required, Validators.maxLength(50)]],
      username: ['', [Validators.required, Validators.maxLength(50)]],
      password: [
        '',
        [Validators.minLength(PASSWORD_MIN_LENGTH), Validators.pattern(PASSWORD_PATTERN)],
      ],
      confirmPassword: [''],
      role: ['User' as UserRole, [Validators.required]],
      active: [true],
    },
    { validators: passwordsMatch },
  );

  protected readonly strength = toSignal(
    this.form.controls.password.valueChanges.pipe(map((value) => passwordStrength(value ?? ''))),
    { initialValue: passwordStrength('') as PasswordStrength },
  );

  protected readonly strengthLabel = computed(() => {
    switch (this.strength()) {
      case 'high':
        return 'Alta';
      case 'medium':
        return 'Media';
      default:
        return 'Baja';
    }
  });

  constructor() {
    effect(() => {
      const current = this.user();
      if (current && this.open()) {
        this.form.reset({
          name: current.name,
          username: current.username,
          password: '',
          confirmPassword: '',
          role: current.role,
          active: current.active,
        });
        this.showPassword.set(false);
        this.showConfirmPassword.set(false);
      }
    });
  }

  protected segmentClass(index: number): string {
    const filled = index <= this.filledSegments();
    if (!filled) {
      return 'bg-muted';
    }
    switch (this.strength()) {
      case 'high':
        return 'bg-success';
      case 'medium':
        return 'bg-warning';
      default:
        return 'bg-danger';
    }
  }

  protected strengthTextClass(): string {
    switch (this.strength()) {
      case 'high':
        return 'text-success';
      case 'medium':
        return 'text-warning';
      default:
        return 'text-danger';
    }
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.user();
    if (!current) {
      return;
    }

    this.loading.set(true);

    const { name, username, password, role, active } = this.form.getRawValue();
    const request: UpdateUserRequest = { name, username, role, active };
    if (password) {
      request.password = password;
    }

    try {
      const updated = await firstValueFrom(this.users.updateUser(current.id, request));
      this.updated.emit(updated);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private filledSegments(): number {
    switch (this.strength()) {
      case 'high':
        return 3;
      case 'medium':
        return 2;
      default:
        return 1;
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El nombre de usuario ya está en uso';
      }
      if (error.status === 404) {
        return 'El usuario no existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo actualizar el usuario';
  }
}
