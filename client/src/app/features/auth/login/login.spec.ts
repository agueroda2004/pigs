import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router } from '@angular/router';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { Login } from './login';

class AuthStub {
  isAdmin = vi.fn(() => false);
  login = vi.fn(() => Promise.resolve());
}

class RouterStub {
  navigateByUrl = vi.fn(() => Promise.resolve(true));
}

describe('Login', () => {
  let auth: AuthStub;
  let router: RouterStub;
  let returnUrl: string | null;

  beforeEach(async () => {
    auth = new AuthStub();
    router = new RouterStub();
    returnUrl = null;

    await TestBed.configureTestingModule({
      imports: [Login],
      providers: [
        { provide: AuthService, useValue: auth },
        { provide: Router, useValue: router },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { queryParamMap: { get: () => returnUrl } } },
        },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const submit = async () => {
    const component = TestBed.createComponent(Login).componentInstance as any;
    component.form.setValue({ username: 'ana', password: 'secret' });
    await component.submit();
  };

  it('sends administrators to the dashboard by default', async () => {
    auth.isAdmin.mockReturnValue(true);

    await submit();

    expect(router.navigateByUrl).toHaveBeenCalledWith('/');
  });

  it('sends regular users to their home page by default', async () => {
    await submit();

    expect(router.navigateByUrl).toHaveBeenCalledWith('/home');
  });

  it('preserves the requested returnUrl', async () => {
    auth.isAdmin.mockReturnValue(true);
    returnUrl = '/breeds';

    await submit();

    expect(router.navigateByUrl).toHaveBeenCalledWith('/breeds');
  });
});
