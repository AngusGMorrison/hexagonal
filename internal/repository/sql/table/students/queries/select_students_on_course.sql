SELECT id, name, birthdate, email
FROM students
INNER JOIN enrollments
ON students.id = enrollments.student_id
WHERE enrollments.course_id = $1;
