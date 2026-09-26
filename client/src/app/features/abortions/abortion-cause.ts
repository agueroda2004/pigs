import { AbortionCause } from '../../core/abortions/abortion.models';

export const ABORTION_CAUSE_LABELS: Record<AbortionCause, string> = {
  Desconocido: 'Desconocido',
  Infeccioso: 'Infeccioso',
  Traumatismo: 'Traumatismo',
  Manejo: 'Manejo',
  Nutricional: 'Nutricional',
  Otro: 'Otro',
};

export const ABORTION_CAUSE_CLASSES: Record<AbortionCause, string> = {
  Desconocido: 'bg-muted text-muted-foreground',
  Infeccioso: 'bg-danger/10 text-danger',
  Traumatismo: 'bg-warning/10 text-warning',
  Manejo: 'bg-info/10 text-info',
  Nutricional: 'bg-success/10 text-success',
  Otro: 'bg-muted text-muted-foreground',
};
