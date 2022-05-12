SELECT id, code, title, capacity, description
FROM courses
WHERE code = $1;