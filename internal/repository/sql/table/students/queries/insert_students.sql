INSERT INTO students (name, birthdate, email)
VALUES
  (:name, :birthdate, :email)
RETURNING *;