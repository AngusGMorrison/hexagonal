INSERT INTO courses (title, code, capacity, description)
VALUES (
  'Structure and Interpretation of Computer Programs',
  'SICP',
  5,
  'The classic introduction to computer programming.'
);

INSERT INTO students (name, birthdate, email)
VALUES
  ('Berthe Archibald', '1987-09-03', 'berthe@archibaldindustries.com'),
  ('Kassandra Madhukar', '1996-07-07', 'km1996@gmail.com'),
  ('Blandinus Branislava', '1991-09-18', 'blandinus@gmail.com');

-- Create enrollments from the Cartesian product of courses and students.
INSERT INTO enrollments (course_id, student_id)
SELECT courses.id, student_ids.id
FROM courses
LEFT JOIN LATERAL (SELECT id FROM students) student_ids
ON true;

-- Create unenrolled students
INSERT INTO students (name, birthdate, email)
VALUES
  ('Ramdas Tifft', '1991-10-03', 'r.tifft@gmail.com'),
  ('Matheo Travieso', '1984-04-11', 'mat@travieso.com'),
  ('Rhodri Murray', '1998-12-01', 'murrayboi98@hotmail.com'),
  ('Dobrila Starr', '1989-08-21', 'dob.starr@googlemail.com'),
  ('Ampelius Fabian', '1990-11-22', 'amp-fab@btinternet.com');