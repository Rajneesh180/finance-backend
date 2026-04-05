-- Fix seed user passwords (correct bcrypt hash of "password123")
UPDATE users SET password = '$2a$10$tSl4EQJxzs3GhBTVySZYuuPO8YMKLp1ZzWfjPypVdmowu58mZ3iAG'
WHERE email IN ('admin@zorvyn.com', 'analyst@zorvyn.com', 'viewer@zorvyn.com');
