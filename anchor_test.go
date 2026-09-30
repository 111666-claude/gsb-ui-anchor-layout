package uianchor

import "testing"

func TestStartWidgetWithNoMargin(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "w", AnchorX: Start, W: 20, H: 20}})
	rect := layout.Solve(Rect{W: 100, H: 100})["w"]
	if rect.X != 0 || rect.Y != 0 {
		t.Fatalf("贴着左上角、无边距的控件应该在原点：%+v", rect)
	}
}

func TestSolveReturnsEveryWidget(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "a", W: 10}, {ID: "b", W: 10}})
	if len(layout.Solve(Rect{W: 100, H: 100})) != 2 {
		t.Fatal("两个控件应该都求解出来")
	}
}

func TestUnknownParentFallsBackToRoot(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "w", Parent: "ghost", W: 10}})
	if rect := layout.Solve(Rect{X: 5, W: 100})["w"]; rect.X != 5 {
		t.Fatalf("父级不存在时应该退回根容器：%+v", rect)
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if NewLayout(nil).Scanned() != 0 {
		t.Fatal("初始扫描数应该是 0")
	}
}
