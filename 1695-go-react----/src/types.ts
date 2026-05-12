export enum ShelterStatus {
  AVAILABLE = 'AVAILABLE',
  MAINTENANCE = 'MAINTENANCE',
  FULL = 'FULL',
  UNAVAILABLE = 'UNAVAILABLE'
}

export enum TransferStatus {
  PENDING = 'PENDING',
  APPROVED = 'APPROVED',
  IN_TRANSIT = 'IN_TRANSIT',
  ARRIVED = 'ARRIVED',
  STORED = 'STORED'
}

export interface Shelter {
  id: number;
  name: string;
  location: string;
  type: string;
  area: number;
  capacity: number;
  occupied: number;
  facilities: string;
  manager: string;
  status: ShelterStatus;
  latitude?: number;
  longitude?: number;
  created_at: string;
  updated_at: string;
}

export interface Material {
  id: number;
  name: string;
  shelter_id: number;
  quantity: number;
  expiry_date: string;
  min_stock: number;
  created_at: string;
  updated_at: string;
}

export interface Transfer {
  id: number;
  from_shelter_id: number;
  to_shelter_id: number;
  material_id: number;
  material_name: string;
  quantity: number;
  status: TransferStatus;
  created_at: string;
  updated_at: string;
}

export interface Assignment {
  id: number;
  shelter_id: number;
  people_count: number;
  reason?: string;
  created_at: string;
}

export interface ReplenishmentTodo {
  id: number;
  material_id: number;
  shelter_id: number;
  material_name: string;
  current_stock: number;
  min_stock: number;
  required: number;
  status: string;
  created_at: string;
}
