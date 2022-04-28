INSERT INTO courses (title, code, capacity, description)
VALUES
  (:title, :code, :capacity, :description)
RETURNING *;