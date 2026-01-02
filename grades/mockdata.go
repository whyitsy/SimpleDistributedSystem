package grades

// 模拟数据对 students变量进行初始化
func init() {
	students = []Student{
		{
			ID:        1,
			FirstName: "Alice",
			LastName:  "Johnson",
			Grades: []Grade{
				{Title: "Math Exam", Type: Exam, Score: 85},
				{Title: "Science Homework", Type: Homework, Score: 90},
				{Title: "History Project", Type: Project, Score: 88},
			},
		},
		{
			ID:        2,
			FirstName: "Bob",
			LastName:  "Smith",
			Grades: []Grade{
				{Title: "Math Exam", Type: Exam, Score: 78},
				{Title: "Science Homework", Type: Homework, Score: 82},
				{Title: "History Project", Type: Project, Score: 80},
			},
		},
		{
			ID:        3,
			FirstName: "Charlie",
			LastName:  "Brown",
			Grades: []Grade{
				{Title: "Math Exam", Type: Exam, Score: 92},
				{Title: "Science Homework", Type: Homework, Score: 88},
				{Title: "History Project", Type: Project, Score: 91},
			},
		},
		{
			ID:        4,
			FirstName: "Diana",
			LastName:  "Prince",
			Grades: []Grade{
				{Title: "Math Exam", Type: Exam, Score: 95},
				{Title: "Science Homework", Type: Homework, Score: 93},
				{Title: "History Project", Type: Project, Score: 97},
			},
		},
		{
			ID:        5,
			FirstName: "Ethan",
			LastName:  "Hunt",
			Grades: []Grade{
				{Title: "Math Exam", Type: Exam, Score: 88},
				{Title: "Science Homework", Type: Homework, Score: 85},
				{Title: "History Project", Type: Project, Score: 87},
			},
		},
	}
}
