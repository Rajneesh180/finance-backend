DELETE FROM financial_records WHERE user_id IN (
  'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22'
);

DELETE FROM users WHERE email IN (
  'admin@zorvyn.com', 'analyst@zorvyn.com', 'viewer@zorvyn.com'
);
