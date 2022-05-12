INSERT INTO enrollments (course_id, student_id)
VALUES (:course_id, :student_id)
RETURNING *;