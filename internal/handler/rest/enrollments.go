package rest

import (
	"net/http"

	"github.com/angusgmorrison/hexagonal/internal/primitive"
	"github.com/angusgmorrison/hexagonal/internal/service/classservice"
	"github.com/gin-gonic/gin"
)

type enrollmentRequest struct {
	CourseTitle string   `json:"course_title"`
	CourseCode  string   `json:"course_code"`
	Students    students `json:"students"`
}

func (er enrollmentRequest) toDomain() classservice.EnrollmentRequest {
	return classservice.EnrollmentRequest{
		CourseCode: er.CourseCode,
		Students:   er.Students.toDomain(),
	}
}

type students []student

func (s students) toDomain() classservice.Students {
	domainStudents := make(classservice.Students, 0, len(s))

	for _, student := range s {
		domainStudents = append(domainStudents, student.toDomain())
	}

	return domainStudents
}

type student struct {
	Name      string                 `json:"name"`
	Birthdate primitive.Birthdate    `json:"birthdate"`
	Email     primitive.EmailAddress `json:"email"`
}

func (s student) toDomain() classservice.Student {
	return classservice.Student{
		Name:      s.Name,
		Birthdate: s.Birthdate,
		Email:     s.Email,
	}
}

// handleCreateEnrollments receives enrollment requests over HTTP and executes
// them.
func (s *Server) handleCreateEnrollments() gin.HandlerFunc {
	return func(c *gin.Context) {
		var enReq enrollmentRequest
		if err := c.ShouldBind(&enReq); err != nil {
			s.logger.Printf("Failed to parse enrollment request: %s", err)
			c.AbortWithStatus(http.StatusBadRequest)

			return
		}

		if err := s.classService.Enroll(c, enReq.toDomain()); err != nil {
			s.logger.Printf("Enrollment failed: %s", err)
			c.AbortWithStatus(http.StatusUnprocessableEntity)

			return
		}

		c.Status(http.StatusCreated)
	}
}
