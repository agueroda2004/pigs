export type ServiceState = 'Confirmado' | 'Fallido' | 'Aborto' | 'Terminado';

export type MountType = 'Natural' | 'Artificial';

export interface Mount {
  id: string;
  service_id: string;
  boar_id: string;
  operator_id: string;
  mount_number: number;
  mount_date: string;
  type: MountType;
  note: string | null;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface Service {
  id: string;
  sow_id: string;
  expected_farrowing_date: string | null;
  note: string | null;
  state: ServiceState;
  location: string | null;
  mounts: Mount[];
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreateMountRequest {
  boar_id: string;
  operator_id: string;
  mount_date: string;
  type: MountType;
  note?: string;
}

export interface CreateServiceRequest {
  sow_id: string;
  location?: string;
  note?: string;
  mounts: CreateMountRequest[];
}

export interface ServiceFilters {
  sow_id?: string;
  state?: ServiceState;
}
