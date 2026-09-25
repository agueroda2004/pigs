import { ServiceState } from '../../core/services/service.models';

export const SERVICE_STATE_LABELS: Record<ServiceState, string> = {
  Confirmado: 'Confirmado',
  Fallido: 'Fallido',
  Aborto: 'Aborto',
  Terminado: 'Terminado',
};

export const SERVICE_STATE_CLASSES: Record<ServiceState, string> = {
  Confirmado: 'bg-info/10 text-info',
  Fallido: 'bg-danger/10 text-danger',
  Aborto: 'bg-warning/10 text-warning',
  Terminado: 'bg-success/10 text-success',
};
