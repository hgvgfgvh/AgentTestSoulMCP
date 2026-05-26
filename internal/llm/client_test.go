package llm

import (
	"strings"
	"testing"
	"time"
)

func TestFormatNowContext_30MinSlot(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	early := time.Date(2026, 5, 22, 15, 4, 5, 0, loc)
	late := time.Date(2026, 5, 22, 15, 29, 59, 0, loc)
	blockEarly := FormatNowContext(early)
	blockLate := FormatNowContext(late)
	if blockEarly != blockLate {
		t.Fatalf("same 30m slot should produce identical prefix:\nearly:\n%s\nlate:\n%s", blockEarly, blockLate)
	}
	for _, want := range []string{
		"2026-05-22 15:00 ~ 15:29",
		"2026-05-22T07:00Z",
		"2026-05-22T07:29Z",
		"今日（日历日）: 2026-05-22",
		"昨日: 2026-05-21",
		"前天: 2026-05-20",
	} {
		if !strings.Contains(blockEarly, want) {
			t.Fatalf("missing %q in:\n%s", want, blockEarly)
		}
	}
	if strings.Contains(blockEarly, ":05") || strings.Contains(blockEarly, ":59") {
		t.Fatalf("should not contain seconds-level times:\n%s", blockEarly)
	}
}

func TestFormatNowContext_nextSlotDiffers(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	a := FormatNowContext(time.Date(2026, 5, 22, 15, 29, 0, 0, loc))
	b := FormatNowContext(time.Date(2026, 5, 22, 15, 30, 0, 0, loc))
	if a == b {
		t.Fatal("adjacent 30m slots should differ")
	}
	if !strings.Contains(b, "15:30 ~ 15:59") {
		t.Fatalf("expected next slot in:\n%s", b)
	}
}

func TestFloorTo30Min(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	in := time.Date(2026, 5, 22, 15, 44, 0, 0, loc)
	got := floorTo30Min(in)
	want := time.Date(2026, 5, 22, 15, 30, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
