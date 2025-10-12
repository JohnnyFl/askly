import { createClient } from '@supabase/supabase-js';

const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL!;
const supabaseAnonKey = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!;

export const supabase = createClient(supabaseUrl, supabaseAnonKey);

export type Channel = {
  id: string;
  name: string;
  description: string;
  avatar_url: string;
  youtube_url: string;
  created_at: string;
};

export type Question = {
  id: string;
  channel_id: string;
  user_id: string;
  question_text: string;
  answer_text: string;
  created_at: string;
  is_faq: boolean;
  faq_order: number;
};

export type UserProfile = {
  id: string;
  email: string;
  full_name: string;
  created_at: string;
};
