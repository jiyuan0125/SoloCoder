export interface Volunteer {
  id: number;
  name: string;
  phone: string;
  id_card: string;
  skills: string;
  availability: string;
  total_hours: number;
  is_backbone: number;
  created_at: string;
}

export interface Activity {
  id: number;
  name: string;
  category: string;
  date: string;
  start_time: string;
  end_time: string;
  location: string;
  required_people: number;
  skills_required: string;
  created_by: number;
  created_at: string;
}

export interface ActivityRegistration {
  id: number;
  volunteer_id: number;
  activity_id: number;
  status: string;
  actual_start_time: string | null;
  actual_end_time: string | null;
  calculated_hours: number;
  created_at: string;
}

export interface MeetingRoom {
  id: number;
  name: string;
  capacity: number;
}

export interface RoomBooking {
  id: number;
  room_id: number;
  date: string;
  start_time: string;
  end_time: string;
  team_name: string;
  purpose: string | null;
  created_at: string;
}
