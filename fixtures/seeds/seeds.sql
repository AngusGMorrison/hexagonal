INSERT INTO courses (title, code, capacity, description)
VALUES (
  'Structure and Interpretation of Computer Programs',
  'SCIP',
  5,
  'The classic introduction to computer programming.'
);

WITH course_fk AS (
  SELECT id FROM courses
  WHERE code = 'SCIP'
)
INSERT INTO students (name, birthday, email, course_id)
VALUES
  ('Ramdas Tifft', '1991-10-03', 'r.tifft@gmail.com', course_fk),
  ('Matheo Travieso', '1984-04-11', 'mat@travieso.com', course_fk);
