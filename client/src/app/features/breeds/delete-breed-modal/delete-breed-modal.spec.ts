import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Breed } from '../../../core/breeds/breed.models';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteBreedModal } from './delete-breed-modal';

function buildBreed(): Breed {
  return { id: '42', name: 'Duroc', active: true };
}

class BreedsStub {
  deleteBreed = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteBreedModal', () => {
  let stub: BreedsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BreedsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteBreedModal],
      providers: [
        { provide: BreedsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteBreedModal);
    fixture.componentRef.setInput('breed', buildBreed());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no breed', async () => {
    const fixture = TestBed.createComponent(DeleteBreedModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteBreed).not.toHaveBeenCalled();
  });

  it('deletes the breed and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteBreed).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildBreed());
  });

  it('shows the server message when the breed has linked records', async () => {
    stub.deleteBreed = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error:
                'No se puede eliminar esta raza porque tiene registros enlazados. Desactívala en su lugar.',
            },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar esta raza porque tiene registros enlazados. Desactívala en su lugar.',
    );
  });

  it('shows a fallback when the breed is missing', async () => {
    stub.deleteBreed = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('La raza no existe');
  });
});
