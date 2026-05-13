package handlers

import (
	"course-schedule/models"
	"course-schedule/services"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateCourse(c *gin.Context) {
	var course models.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := services.CreateCourse(&course)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	course.ID = id
	c.JSON(http.StatusCreated, course)
}

func GetCourse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	course, err := services.GetCourseByID(id)
	if errors.Is(err, services.ErrCourseNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, course)
}

func ListCourses(c *gin.Context) {
	courses, err := services.GetAllCourses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func UpdateCourse(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	var course models.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateCourse(id, &course); err != nil {
		if errors.Is(err, services.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "course updated"})
}

func CreateSchedule(c *gin.Context) {
	var input models.ScheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := services.CreateSchedule(&input)
	if errors.Is(err, services.ErrCourseNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}
	if errors.Is(err, services.ErrSeriesOrderInvalid) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "series time not in order"})
		return
	}
	if errors.Is(err, services.ErrClassroomConflict) || errors.Is(err, services.ErrInstructorConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func ListSchedules(c *gin.Context) {
	schedules, err := services.GetAllSchedules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

func ListSchedulesByCourse(c *gin.Context) {
	courseID, err := strconv.ParseInt(c.Param("courseId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}

	schedules, err := services.GetSchedulesByCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schedules)
}

func EnrollCourse(c *gin.Context) {
	var input models.EnrollInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.EnrollCourse(&input); err != nil {
		if errors.Is(err, services.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, services.ErrCourseFull) {
			c.JSON(http.StatusConflict, gin.H{"error": "课程已满"})
			return
		}
		if errors.Is(err, services.ErrAlreadyEnrolled) {
			c.JSON(http.StatusConflict, gin.H{"error": "already enrolled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "enrolled successfully"})
}

func DropCourse(c *gin.Context) {
	var input models.DropInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.DropCourse(&input); err != nil {
		if errors.Is(err, services.ErrCourseNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
			return
		}
		if errors.Is(err, services.ErrEnrollmentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "enrollment not found"})
			return
		}
		if errors.Is(err, services.ErrDropTooLate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dropped successfully"})
}

func ListEnrollments(c *gin.Context) {
	if courseIDStr := c.Query("course_id"); courseIDStr != "" {
		courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
			return
		}
		enrollments, err := services.GetEnrollmentsByCourse(courseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, enrollments)
		return
	}

	if studentName := c.Query("student_name"); studentName != "" {
		enrollments, err := services.GetEnrollmentsByStudent(studentName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, enrollments)
		return
	}

	enrollments, err := services.GetAllEnrollments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, enrollments)
}

func ListNotifications(c *gin.Context) {
	if courseIDStr := c.Query("course_id"); courseIDStr != "" {
		courseID, err := strconv.ParseInt(courseIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
			return
		}
		notifications, err := services.GetNotificationsByCourse(courseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, notifications)
		return
	}

	if studentName := c.Query("student_name"); studentName != "" {
		notifications, err := services.GetNotificationsByStudent(studentName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, notifications)
		return
	}

	notifications, err := services.GetAllNotifications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}
