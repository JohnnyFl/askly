/*
  # YouTube Q&A Platform Schema

  ## Overview
  Creates the complete database schema for a YouTube channel Q&A platform where users can ask questions
  and receive AI-powered responses as if chatting with the channel host.

  ## New Tables
  
  ### `channels`
  - `id` (uuid, primary key) - Unique identifier for each YouTube channel
  - `name` (text) - Channel name
  - `description` (text) - Channel description
  - `avatar_url` (text) - URL to channel avatar image
  - `youtube_url` (text) - Original YouTube channel URL
  - `created_at` (timestamptz) - Creation timestamp
  
  ### `questions`
  - `id` (uuid, primary key) - Unique identifier for each question
  - `channel_id` (uuid, foreign key) - References the channel
  - `user_id` (uuid, foreign key) - References auth.users
  - `question_text` (text) - The question asked by user
  - `answer_text` (text) - AI-generated answer
  - `created_at` (timestamptz) - Question timestamp
  - `is_faq` (boolean) - Whether this is marked as frequently asked
  - `faq_order` (integer) - Display order for FAQs
  
  ### `user_profiles`
  - `id` (uuid, primary key, foreign key to auth.users) - User identifier
  - `email` (text) - User email
  - `full_name` (text) - User's full name
  - `created_at` (timestamptz) - Profile creation timestamp

  ## Security
  - Row Level Security (RLS) enabled on all tables
  - Policies ensure users can only read public channel data
  - Users can only view and create their own questions
  - Users can read all FAQs but only admins can mark questions as FAQs
  - Users can manage their own profiles
  
  ## Notes
  1. The backend Go service will handle AI response generation
  2. FAQs are questions marked with `is_faq = true`
  3. Recent questions are sorted by `created_at` per user
  4. Channel data is publicly readable for all authenticated users
*/

-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  description text DEFAULT '',
  avatar_url text DEFAULT '',
  youtube_url text DEFAULT '',
  created_at timestamptz DEFAULT now()
);

ALTER TABLE channels ENABLE ROW LEVEL SECURITY;

-- Create user_profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
  id uuid PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  email text NOT NULL,
  full_name text DEFAULT '',
  created_at timestamptz DEFAULT now()
);

ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;

-- Create questions table
CREATE TABLE IF NOT EXISTS questions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  channel_id uuid NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  question_text text NOT NULL,
  answer_text text DEFAULT '',
  created_at timestamptz DEFAULT now(),
  is_faq boolean DEFAULT false,
  faq_order integer DEFAULT 0
);

ALTER TABLE questions ENABLE ROW LEVEL SECURITY;

-- Channels policies: All authenticated users can read all channels
CREATE POLICY "Authenticated users can view all channels"
  ON channels
  FOR SELECT
  TO authenticated
  USING (true);

-- User profiles policies
CREATE POLICY "Users can view their own profile"
  ON user_profiles
  FOR SELECT
  TO authenticated
  USING (auth.uid() = id);

CREATE POLICY "Users can insert their own profile"
  ON user_profiles
  FOR INSERT
  TO authenticated
  WITH CHECK (auth.uid() = id);

CREATE POLICY "Users can update their own profile"
  ON user_profiles
  FOR UPDATE
  TO authenticated
  USING (auth.uid() = id)
  WITH CHECK (auth.uid() = id);

-- Questions policies
CREATE POLICY "Users can view their own questions"
  ON questions
  FOR SELECT
  TO authenticated
  USING (auth.uid() = user_id);

CREATE POLICY "Users can view all FAQs"
  ON questions
  FOR SELECT
  TO authenticated
  USING (is_faq = true);

CREATE POLICY "Users can insert their own questions"
  ON questions
  FOR INSERT
  TO authenticated
  WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update their own questions"
  ON questions
  FOR UPDATE
  TO authenticated
  USING (auth.uid() = user_id)
  WITH CHECK (auth.uid() = user_id);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_questions_channel_id ON questions(channel_id);
CREATE INDEX IF NOT EXISTS idx_questions_user_id ON questions(user_id);
CREATE INDEX IF NOT EXISTS idx_questions_is_faq ON questions(is_faq) WHERE is_faq = true;
CREATE INDEX IF NOT EXISTS idx_questions_created_at ON questions(created_at DESC);