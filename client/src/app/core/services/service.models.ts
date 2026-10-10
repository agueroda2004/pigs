export type ServiceState = 'Confirmado' | 'Fallido' | 'Aborto' | 'Terminado';

export type MountType = 'Natural' | 'Artificial';

export interface Mount {
  id: string;
  service_id: string;
  boar_id: string;
  boar_code: string;
  operator_id: string;
  operator_name: string;
  mount_number: number;
  mount_date: string;
  type: MountType;
  note: string | null;
}

export interface Service {
  id: string;
  sow_id: string;
  sow_code: string;
  expected_farrowing_date: string | null;
  note: string | null;
  state: ServiceState;
  location: string | null;
  mounts: Mount[];
}

export interface ServicePage {
  items: Service[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
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

export interface UpdateMountRequest {
  id: string;
  boar_id: string;
  operator_id: string;
  mount_date: string;
  type: MountType;
  note?: string;
}

export interface UpdateServiceMountsRequest {
  create: CreateMountRequest[];
  update: UpdateMountRequest[];
  delete: string[];
}

export interface UpdateServiceRequest {
  location?: string;
  note?: string;
  mounts: UpdateServiceMountsRequest;
}

export interface ServiceFilters {
  sow_code?: string;
  state?: ServiceState;
}
