import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Breed } from '../../../core/breeds/breed.models';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { EditBreedModal } from './edit-breed-modal';

function buildBreed(): Breed {
  return {
    id: '42',
    name: 'Duroc',
    active: true,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class BreedsStub {
  updateBreed = vi.fn(() => of(buildBreed()));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditBreedModal', () => {
  let stub: BreedsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BreedsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditBreedModal],
      providers: [
        { provide: BreedsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => {
    const fixture = TestBed.createComponent(EditBreedModal);
    fixture.componentRef.setInput('breed', buildBreed());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    return fixture.componentInstance as any;
  };

  it('prefills the form from the selected breed', () => {
    const component = create();

    expect(component.form.getRawValue()).toEqual({ name: 'Duroc', active: true });
  });

  it('updates the breed with changed fields and emits the result', async () => {
    const component = create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.controls.name.setValue('Landrace');
    component.form.controls.active.setValue(false);

    await component.submit();

    expect(stub.updateBreed).toHaveBeenCalledWith('42', { name: 'Landrace', active: false });
    expect(updated).toHaveBeenCalled();
  });

  it('does not call the API when nothing changed', async () => {
    const component = create();
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateBreed).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalled();
  });

  it('shows an error toast when the name is taken', async () => {
    stub.updateBreed = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.controls.name.setValue('Landrace');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre de la raza ya existe');
  });
});
