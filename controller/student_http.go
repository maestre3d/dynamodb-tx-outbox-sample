package controller

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/maestre3d/dynamodb-tx-outbox/application/appservice"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/persistence"
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	appservice.StudentRepository = persistence.NewStudentDynamoDb(dynamodb.NewFromConfig(cfg))
}

func MapStudentHttpRoutes(r *mux.Router) {
	r.Path("/schools/{school_id}/students").Methods(http.MethodPost).HandlerFunc(createStudent)
	r.Path("/schools/{school_id}/students/{student_id}").Methods(http.MethodGet).HandlerFunc(getStudent)
}

func createStudent(w http.ResponseWriter, r *http.Request) {
	studentID := uuid.NewString()
	if err := appservice.RegisterStudent(r.Context(), appservice.RegisterStudentArgs{
		StudentID: studentID,
		Name:      r.PostFormValue("student_name"),
		SchoolID:  mux.Vars(r)["school_id"],
	}); err != nil {
		respondMessageJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondMessageJSON(w, studentID, 200)
}

func getStudent(w http.ResponseWriter, r *http.Request) {
	student, err := appservice.GetStudentByID(r.Context(), mux.Vars(r)["school_id"], mux.Vars(r)["student_id"])
	if err != nil {
		respondMessageJSON(w, err.Error(), http.StatusInternalServerError)
		return
	} else if student == nil {
		respondMessageJSON(w, "student not found", http.StatusNotFound)
		return
	}
	respondStructJSON(w, student, 200)
}
