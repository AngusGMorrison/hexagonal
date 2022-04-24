DROP TABLE IF EXISTS courses CASCADE;
DROP TABLE IF EXISTS students CASCADE;
DROP TABLE IF EXISTS enrollments;

CREATE TABLE courses (
  id BIGSERIAL PRIMARY KEY,
  code VARCHAR(255) NOT NULL,
  title VARCHAR(255) NOT NULL,
  capacity INT NOT NULL,
  description TEXT
);

CREATE UNIQUE INDEX courses_code_idx
ON courses (code);

CREATE TABLE students (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  birthdate DATE NOT NULL,
  email VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX students_email_idx
ON students (email);

CREATE TABLE enrollments (
  id BIGSERIAL PRIMARY KEY,
  course_id BIGINT REFERENCES courses NOT NULL,
  student_id BIGINT REFERENCES students NOT NULL
);

CREATE INDEX enrollments_course_id_idx
ON enrollments (course_id);

CREATE INDEX enrollments_student_id_idx
ON enrollments (student_id);