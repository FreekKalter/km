package km

import (
	"testing"
	"time"
)

var k = Kilometers{1, time.Now(), 1, 2, 3, 4, "test"}
var k2 = Kilometers{1, time.Now(), 1, 2, 3, 0, "test"}

func TestGetMax(t *testing.T) {
	if v := k.getMax(); v != 4 {
		t.Errorf("got %d: want 4", v)
	}
	if v := k2.getMax(); v != 3 {
		t.Errorf("got %d: want 3", v)
	}
}
