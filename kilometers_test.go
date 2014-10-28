package km

import (
	"testing"
	"time"
)

func TestGetMax(t *testing.T) {
	kiloTests := []Kilometers{
		Kilometers{1, time.Now(), 1, 0, 0, 0, "test"},
		Kilometers{1, time.Now(), 1, 2, 0, 0, "test"},
		Kilometers{1, time.Now(), 1, 2, 3, 0, "test"},
		Kilometers{1, time.Now(), 1, 2, 3, 4, "test"},
	}
	for i, k := range kiloTests {
		if v := k.getMax(); v != i+1 {
			t.Errorf("got %d: want %d", v, i+1)
		}
	}
	if v := new(Kilometers).getMax(); v != 0 {
		t.Errorf("got %d: want %d", v, 0)
	}
}

func TestAddFields(t *testing.T) {
	fields := []Field{
		Field{Km: 1, Name: "Begin"},
		Field{Km: 2, Name: "Eerste"},
		Field{Km: 3, Name: "Laatste"},
		Field{Km: 4, Name: "Terug"},
	}
	ki := new(Kilometers)
	ki.AddFields(fields)
	if ki.Begin != 1 || ki.Eerste != 2 || ki.Laatste != 3 || ki.Terug != 4 {
		t.Error("Kilometers AddFields did not do its job")
	}

}
