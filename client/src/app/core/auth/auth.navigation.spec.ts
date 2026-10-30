import { ADMIN_HOME, USER_HOME, landingUrl } from './auth.navigation';

describe('landingUrl', () => {
  it('sends administrators to the dashboard', () => {
    expect(landingUrl(true)).toBe(ADMIN_HOME);
    expect(ADMIN_HOME).toBe('/');
  });

  it('sends regular users to their home page', () => {
    expect(landingUrl(false)).toBe(USER_HOME);
    expect(USER_HOME).toBe('/home');
  });
});
