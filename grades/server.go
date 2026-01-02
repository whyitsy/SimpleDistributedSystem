package grades

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandler() {
	sh := &studentHandler{}
	http.Handle("/students", sh)  // 请求集合数据
	http.Handle("/students/", sh) // 请求单个数据的操作
}

type studentHandler struct{}

// /students  GET 获取所有学生
// /students/{id}  GET 获取单个学生的信息
// /students/{id}/grades  POST 添加学生成绩
func (sh *studentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// web框架提供了路由解析功能, 这里就简单按照 / 分割路径处理三个不同长度的path
	pathSegments := strings.Split(r.URL.Path, "/")
	if len(pathSegments) < 2 || pathSegments[1] != "students" {
		http.NotFound(w, r)
		return
	}
	switch len(pathSegments) {
	case 2:
		// /students
		if r.Method == http.MethodGet {
			handleGetAllStudents(w, r)
			return
		}
	case 3:
		// /students/{id}
		if r.Method == http.MethodGet {
			handleGetStudentByID(w, r, pathSegments[2])
			return
		}
	case 4:
		// /students/{id}/grades
		if r.Method == http.MethodPost && pathSegments[3] == "grades" {
			handlePostStudentGrades(w, r, pathSegments[2])
			return
		}
	}

}

func handlePostStudentGrades(w http.ResponseWriter, r *http.Request, s string) {
	id, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		log.Println(err)
		return
	}
	student, err := students.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	var grade Grade
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&grade)
	if err != nil {
		http.Error(w, "Invalid grade data", http.StatusBadRequest)
		log.Println(err)
		return
	}
	student.Grades = append(student.Grades, grade)
	w.WriteHeader(http.StatusCreated)
	// 返回更新后的学生信息
	sj, err := json.Marshal(student)
	if err != nil {
		http.Error(w, "Failed to marshal student", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(sj)
	if err != nil {
		log.Println(err)
		return
	}

}

func handleGetStudentByID(w http.ResponseWriter, r *http.Request, s string) {
	id, err := strconv.Atoi(s)
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		log.Println(err)
		return
	}
	student, err := students.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Println(err)
		return
	}
	sj, err := json.Marshal(student)
	if err != nil {
		http.Error(w, "Failed to marshal student", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(sj)
	if err != nil {
		log.Println(err)
		return
	}
}

func handleGetAllStudents(w http.ResponseWriter, r *http.Request) {
	sj, err := json.Marshal(students)
	if err != nil {
		http.Error(w, "Failed to marshal students", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(sj)
	if err != nil {
		log.Println(err)
		return
	}

}
