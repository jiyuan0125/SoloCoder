export interface Project {
  id: string;
  name: string;
  description: string;
  target_amount: number;
  start_date: string;
  end_date: string;
  has_tax_deductible_qualification: number;
  created_at: string;
}

export interface Donation {
  id: string;
  project_id: string;
  donor_name: string;
  id_card_last4: string;
  amount: number;
  certificate_number: string | null;
  donated_at: string;
  created_at: string;
}

export interface Certificate {
  id: string;
  certificate_number: string;
  donation_id: string;
  donor_name: string;
  id_card_last4: string;
  project_id: string;
  amount: number;
  issued_at: string;
  created_at: string;
}

export interface TaxDeductionVoucher {
  id: string;
  voucher_number: string;
  donation_id: string;
  donor_name: string;
  id_card_last4: string;
  project_id: string;
  deductible_amount: number;
  tax_year: number;
  status: string;
  voided_at: string | null;
  issued_at: string;
  created_at: string;
}

export interface DonorAnnualSummary {
  id: string;
  donor_name: string;
  id_card_last4: string;
  year: number;
  total_donation_amount: number;
  total_tax_deductible_amount: number;
  tax_income_base: number;
  tax_limit: number;
  updated_at: string;
}

export interface ProjectWarning {
  id: string;
  project_id: string;
  warning_type: string;
  message: string;
  created_at: string;
}

export interface CreateProjectRequest {
  name: string;
  description: string;
  target_amount: number;
  start_date: string;
  end_date: string;
  has_tax_deductible_qualification?: boolean;
}

export interface CreateDonationRequest {
  project_id: string;
  donor_name: string;
  id_card_last4: string;
  amount: number;
  donated_at?: string;
}

export interface CreateVoucherRequest {
  donation_id: string;
  tax_income_base: number;
}
