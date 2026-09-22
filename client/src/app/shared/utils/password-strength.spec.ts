import { passwordStrength } from './password-strength';

describe('passwordStrength', () => {
  it('returns low for empty or weak passwords', () => {
    expect(passwordStrength('')).toBe('low');
    expect(passwordStrength('password')).toBe('low');
  });

  it('returns medium for moderate passwords', () => {
    expect(passwordStrength('Password1')).toBe('medium');
  });

  it('returns high for strong passwords', () => {
    expect(passwordStrength('Password1!')).toBe('high');
    expect(passwordStrength('Str0ng-Password!')).toBe('high');
  });
});
