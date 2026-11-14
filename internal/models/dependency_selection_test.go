package models

import (
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Pin the existing dependency selection contract without changing dispatch
// semantics while investigating older parent/child execution reports.
func TestDependencyTaskSelectionContract(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true, TablePrefix: TablePrefix},
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	original := Db
	Db = db
	t.Cleanup(func() { Db = original })
	if err := db.AutoMigrate(&Task{}, &TaskHost{}, &Host{}); err != nil {
		t.Fatal(err)
	}
	for _, task := range []Task{
		{Id: 1, Name: "parent", Level: TaskLevelParent},
		{Id: 2, Name: "child", Level: TaskLevelChild},
		{Id: 3, Name: "another-parent", Level: TaskLevelParent},
	} {
		if err := db.Create(&task).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		ids  string
		want int
	}{
		{"", 0}, {"1,2,3", 1}, {"2,2", 1}, {"999", 0},
	} {
		tasks, err := new(Task).GetDependencyTaskList(tc.ids)
		if err != nil || len(tasks) != tc.want {
			t.Fatalf("ids=%q count=%d err=%v", tc.ids, len(tasks), err)
		}
		if len(tasks) > 0 && tasks[0].Id != 2 {
			t.Fatalf("selected non-child: %+v", tasks)
		}
	}
}
