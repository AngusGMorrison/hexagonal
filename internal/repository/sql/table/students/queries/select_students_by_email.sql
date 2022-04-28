SELECT id, name, birthdate, email
FROM students
WHERE email IN (?);
