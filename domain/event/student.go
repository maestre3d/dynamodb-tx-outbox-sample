package event

type StudentRegistered struct {
	StudentID string `json:"student_id" avro:"student_id"`
	Name      string `json:"student_name" avro:"student_name"`
	SchoolID  string `json:"school_id" avro:"school_id"`
}
