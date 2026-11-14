package models

import "testing"

func TestNotifyDiagnosisPersistence(t *testing.T) {
	db := setupTestDB(t)
	original := Db
	Db = db
	t.Cleanup(func() { Db = original })
	if err := db.AutoMigrate(&Task{}, &TaskTemplate{}); err != nil {
		t.Fatal(err)
	}
	task := Task{Name: "diagnosis", Spec: "* * * * *", Command: "true", NotifyDiagnosis: 1}
	id, err := task.Create()
	if err != nil {
		t.Fatal(err)
	}
	var saved Task
	if err := db.First(&saved, id).Error; err != nil || saved.NotifyDiagnosis != 1 {
		t.Fatalf("create lost diagnosis: %+v, %v", saved, err)
	}
	for _, value := range []int8{0, 1, 0} {
		task.NotifyDiagnosis = value
		if _, err := task.UpdateBean(id); err != nil {
			t.Fatal(err)
		}
		if err := db.First(&saved, id).Error; err != nil || saved.NotifyDiagnosis != value {
			t.Fatalf("update lost diagnosis %d: %+v, %v", value, saved, err)
		}
	}
	tmpl := TaskTemplate{Name: "diagnosis", NotifyDiagnosis: 1}
	templateID, err := tmpl.Create()
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []int8{0, 1, 0} {
		tmpl.NotifyDiagnosis = value
		if _, err := tmpl.UpdateBean(templateID); err != nil {
			t.Fatal(err)
		}
		saved, err := tmpl.Detail(templateID)
		if err != nil || saved.NotifyDiagnosis != value {
			t.Fatalf("template lost diagnosis %d: %+v, %v", value, saved, err)
		}
	}
}
