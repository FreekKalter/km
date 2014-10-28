package km

import (
	"fmt"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	ti := Times{}
	if ti.Id != 0 {
		t.Fatal("id should be 0 on times init")
	}
}

func TestUpdate(t *testing.T) {
	testDate := time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC)

	fields := []Field{Field{Km: 123456, Time: "13:00", Name: "Begin"}}
	cmpDate := int64(1257854400) // 10-11-2009 13:00 uur (in unix formaat)

	dateStr := fmt.Sprintf("%d-%d-%d", testDate.Month(), testDate.Day(), testDate.Year())
	ti := Times{}
	err := ti.UpdateObject(dateStr, fields)
	if err != nil {
		t.Fatal(err)
	}

	if ti.Begin != cmpDate {
		t.Fatalf("updating Begin field: expected %d got %d", cmpDate, ti.Begin)
	}

	if ti.CheckIn != 0 || ti.CheckOut != 0 || ti.Laatste != 0 {
		t.Fatalf("only 1 field should change %+v", ti)
	}

	ti = Times{}
	fields = []Field{Field{Km: 123456, Time: "jemoeder", Name: "Begin"}}
	if ti.UpdateObject(dateStr, fields) == nil {
		t.Fatal("updateObjects should fail on invalid time field")
	}
}
