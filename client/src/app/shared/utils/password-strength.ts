export type PasswordStrength = 'low' | 'medium' | 'high';

export const PASSWORD_MIN_LENGTH = 8;

export const PASSWORD_PATTERN = /^(?=.*\d)(?=.*[^A-Za-z0-9]).+$/;

export function passwordStrength(value: string): PasswordStrength {
  if (!value) {
    return 'low';
  }

  let score = 0;

  if (value.length >= PASSWORD_MIN_LENGTH) {
    score += 1;
  }
  if (value.length >= 12) {
    score += 1;
  }
  if (/[a-z]/.test(value) && /[A-Z]/.test(value)) {
    score += 1;
  }
  if (/\d/.test(value)) {
    score += 1;
  }
  if (/[^A-Za-z0-9]/.test(value)) {
    score += 1;
  }

  if (score >= 4) {
    return 'high';
  }
  if (score >= 3) {
    return 'medium';
  }
  return 'low';
}
