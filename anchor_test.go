package uianchor

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

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

func TestCenterWidgetStaysCenteredAfterResize(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "w", AnchorX: Center, AnchorY: Center, W: 20, H: 20}})
	for _, size := range []float64{100, 200, 480, 1080} {
		rect := layout.Solve(Rect{W: size, H: size})["w"]
		if rect.X+rect.W/2 != size/2 || rect.Y+rect.H/2 != size/2 {
			t.Fatalf("根容器 %.0f 时控件中心应与父中心重合：%+v", size, rect)
		}
	}
}

func TestMarginOffsetsStartAnchor(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "w", AnchorX: Start, AnchorY: Start, Margin: 5, W: 20, H: 20}})
	rect := layout.Solve(Rect{W: 100, H: 100})["w"]
	if rect.X != 5 || rect.Y != 5 {
		t.Fatalf("Start 锚点的位置应该是 parent 起点加 Margin：%+v", rect)
	}
}

func TestEndAlignsRightMinusMargin(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "w", AnchorX: End, AnchorY: End, Margin: 5, W: 20, H: 20}})
	rect := layout.Solve(Rect{W: 100, H: 100})["w"]
	if rect.X != 75 || rect.Y != 75 {
		t.Fatalf("End 锚点应该右对齐再减 Margin：%+v", rect)
	}
}

func TestStretchFillsParentMinusMargins(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "w", AnchorX: Stretch, AnchorY: Stretch, Margin: 5, W: 20, H: 20}})
	rect := layout.Solve(Rect{W: 100, H: 100})["w"]
	if rect.X != 5 || rect.W != 90 || rect.Y != 5 || rect.H != 90 {
		t.Fatalf("Stretch 应该撑满父矩形并两侧各留 Margin（忽略自身 W/H）：%+v", rect)
	}
}

func TestChildRectFollowsParentRect(t *testing.T) {
	layout := NewLayout([]Widget{
		{ID: "parent", AnchorX: Start, AnchorY: Start, Margin: 10, W: 50, H: 50},
		{ID: "child", Parent: "parent", AnchorX: Start, AnchorY: Start, Margin: 5, W: 10, H: 10},
	})
	rects := layout.Solve(Rect{W: 100, H: 100})
	if rects["parent"].X != 10 || rects["child"].X != 15 || rects["child"].Y != 15 {
		t.Fatalf("子控件应该相对父矩形定位：parent=%+v child=%+v", rects["parent"], rects["child"])
	}
}

func TestDeepChainScansWithinWidgetCount(t *testing.T) {
	const count = 3000
	widgets := make([]Widget, 0, count)
	for index := 0; index < count; index++ {
		parent := ""
		if index > 0 {
			parent = fmt.Sprintf("w-%d", index-1)
		}
		widgets = append(widgets, Widget{ID: fmt.Sprintf("w-%d", index), Parent: parent, W: 10})
	}
	layout := NewLayout(widgets)
	rects := layout.Solve(Rect{W: 100, H: 100})
	if len(rects) != count {
		t.Fatalf("应该求解出 %d 个控件，实际 %d", count, len(rects))
	}
	if scanned := layout.Scanned(); scanned > count {
		t.Fatalf("父链只解析一次，scanned=%d 不应超过控件数 %d", scanned, count)
	}
}

func TestSolveIsDeterministic(t *testing.T) {
	widgets := []Widget{
		{ID: "a", AnchorX: Center, AnchorY: Center, W: 20, H: 20},
		{ID: "b", Parent: "a", AnchorX: End, AnchorY: End, Margin: 3, W: 10, H: 10},
		{ID: "c", Parent: "b", AnchorX: Stretch, AnchorY: Stretch, Margin: 2, W: 5, H: 5},
	}
	layout := NewLayout(widgets)
	first := layout.Solve(Rect{W: 100, H: 100})
	second := layout.Solve(Rect{W: 100, H: 100})
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("同一份控件表重复求解结果应该相同：%+v vs %+v", first, second)
	}
	if layout.Scanned() > len(widgets) {
		t.Fatalf("重复求解后 scanned=%d 不应超过控件数", layout.Scanned())
	}
}

// referenceSolve 是「父矩形 + 锚点 + 边距 + 记忆化」的独立参考实现，用于对照自证。
func referenceSolve(widgets []Widget, root Rect) map[string]Rect {
	table := make(map[string]Widget, len(widgets))
	for _, widget := range widgets {
		table[widget.ID] = widget
	}
	memo := make(map[string]Rect, len(widgets))
	var resolve func(id string) Rect
	resolve = func(id string) Rect {
		if rect, ok := memo[id]; ok {
			return rect
		}
		widget := table[id]
		parent := root
		if up, ok := table[widget.Parent]; ok {
			parent = resolve(up.ID)
		}
		rect := Rect{W: widget.W, H: widget.H}
		switch widget.AnchorX {
		case Center:
			rect.X = parent.X + (parent.W-widget.W)/2
		case End:
			rect.X = parent.X + parent.W - widget.W - widget.Margin
		case Stretch:
			rect.X = parent.X + widget.Margin
			rect.W = parent.W - 2*widget.Margin
		default:
			rect.X = parent.X + widget.Margin
		}
		switch widget.AnchorY {
		case Center:
			rect.Y = parent.Y + (parent.H-widget.H)/2
		case End:
			rect.Y = parent.Y + parent.H - widget.H - widget.Margin
		case Stretch:
			rect.Y = parent.Y + widget.Margin
			rect.H = parent.H - 2*widget.Margin
		default:
			rect.Y = parent.Y + widget.Margin
		}
		memo[id] = rect
		return rect
	}
	out := make(map[string]Rect, len(widgets))
	for _, widget := range widgets {
		out[widget.ID] = resolve(widget.ID)
	}
	return out
}

func TestSolveMatchesReferenceOnRandomTables(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	for trial := 0; trial < 200; trial++ {
		count := 1 + rng.Intn(40)
		widgets := make([]Widget, 0, count)
		for index := 0; index < count; index++ {
			widget := Widget{
				ID:      fmt.Sprintf("w-%d", index),
				AnchorX: Anchor(rng.Intn(4)),
				AnchorY: Anchor(rng.Intn(4)),
				Margin:  float64(rng.Intn(21)),
				W:       1 + float64(rng.Intn(80)),
				H:       1 + float64(rng.Intn(80)),
			}
			switch rng.Intn(4) {
			case 1:
				widget.Parent = fmt.Sprintf("ghost-%d", rng.Intn(count))
			case 2, 3:
				if index > 0 {
					widget.Parent = fmt.Sprintf("w-%d", rng.Intn(index))
				}
			}
			widgets = append(widgets, widget)
		}
		root := Rect{
			X: float64(rng.Intn(201) - 100),
			Y: float64(rng.Intn(201) - 100),
			W: 50 + float64(rng.Intn(1951)),
			H: 50 + float64(rng.Intn(1951)),
		}
		got := NewLayout(widgets).Solve(root)
		want := referenceSolve(widgets, root)
		if len(got) != len(want) {
			t.Fatalf("第 %d 组：控件数不一致 %d != %d", trial, len(got), len(want))
		}
		for id, wantRect := range want {
			if got[id] != wantRect {
				t.Fatalf("第 %d 组控件 %s 矩形不一致：got %+v want %+v", trial, id, got[id], wantRect)
			}
		}
	}
}
