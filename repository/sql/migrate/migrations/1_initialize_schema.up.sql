DROP TABLE IF EXISTS courses CASCADE;
DROP TABLE IF EXISTS students;

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
  email VARCHAR(255) NOT NULL,
  course_id BIGINT NOT NULL REFERENCES courses
);

CREATE UNIQUE INDEX students_email_idx
ON students (email);