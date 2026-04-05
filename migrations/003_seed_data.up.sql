-- Seed data: default admin, analyst, and viewer users
-- Passwords are bcrypt hashes of "password123"

INSERT INTO users (id, email, password, name, role, is_active)
VALUES
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@zorvyn.com',
   '$2a$10$tSl4EQJxzs3GhBTVySZYuuPO8YMKLp1ZzWfjPypVdmowu58mZ3iAG',
   'Admin User', 'admin', true),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'analyst@zorvyn.com',
   '$2a$10$tSl4EQJxzs3GhBTVySZYuuPO8YMKLp1ZzWfjPypVdmowu58mZ3iAG',
   'Analyst User', 'analyst', true),
  ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'viewer@zorvyn.com',
   '$2a$10$tSl4EQJxzs3GhBTVySZYuuPO8YMKLp1ZzWfjPypVdmowu58mZ3iAG',
   'Viewer User', 'viewer', true)
ON CONFLICT (email) DO NOTHING;

-- Sample financial records for the analyst user
INSERT INTO financial_records (user_id, amount, type, category, date, description)
VALUES
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 5000.00, 'income', 'salary', '2026-03-01', 'March salary'),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 1200.00, 'expense', 'rent', '2026-03-05', 'Monthly rent'),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 150.00, 'expense', 'utilities', '2026-03-10', 'Electricity bill'),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 300.00, 'expense', 'groceries', '2026-03-12', NULL),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 2000.00, 'income', 'freelance', '2026-03-15', 'Web dev project'),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 80.00, 'expense', 'transport', '2026-03-18', 'Metro pass'),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 500.00, 'expense', 'insurance', '2026-03-20', 'Health insurance'),
  ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 3500.00, 'income', 'salary', '2026-04-01', 'April salary');
