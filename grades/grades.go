package grades

import "fmt"

type Student struct {
	ID        int
	FirstName string
	LastName  string
	Grades    []Grade
}

func (s *Student) Average() float32 {
	if len(s.Grades) == 0 {
		return 0
	}
	var total float32
	for _, grade := range s.Grades {
		total += grade.Score
	}
	return total / float32(len(s.Grades))
}

type Students []Student

var students Students

func (ss Students) GetByID(id int) (*Student, error) {
	for i, s := range ss {
		if s.ID == id {
			return &ss[i], nil
		}
	}
	return nil, fmt.Errorf("student with ID: %d not found", id)
}

type GradeType string

const (
	Exam     GradeType = "Exam"
	Homework GradeType = "Homework"
	Project  GradeType = "Project"
)

type Grade struct {
	Title string
	Type  GradeType
	Score float32
}
