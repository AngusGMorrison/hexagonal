SELECT s.id, s.name, s.birthdate, s.email
FROM students s
INNER JOIN enrollments e
ON s.id = e.student_id
WHERE e.course_id = $1;
