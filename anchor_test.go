package uianchor

import (
	"fmt"
	"math/rand"
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

func TestCenterStaysCenteredAcrossResolutions(t *testing.T) {
	layout := NewLayout([]Widget{{ID: "w", AnchorX: Center, AnchorY: Center, W: 20, H: 10}})
	for _, root := range []Rect{{W: 100, H: 100}, {W: 640, H: 480}, {X: 3, Y: 7, W: 1920, H: 1080}} {
		rect := layout.Solve(root)["w"]
		if rect.X+rect.W/2 != root.X+root.W/2 || rect.Y+rect.H/2 != root.Y+root.H/2 {
			t.Fatalf("根容器 %+v 下控件中心应与父中心重合：%+v", root, rect)
		}
	}
}

func TestMarginAppliesToAnchors(t *testing.T) {
	root := Rect{X: 10, Y: 20, W: 100, H: 100}
	layout := NewLayout([]Widget{
		{ID: "start", AnchorX: Start, Margin: 5, W: 20},
		{ID: "end", AnchorX: End, Margin: 5, W: 20},
		{ID: "stretch", AnchorX: Stretch, Margin: 5, W: 20},
	})
	out := layout.Solve(root)
	if out["start"].X != 15 {
		t.Fatalf("Start 的 X 应是 parent.X+Margin=15：%+v", out["start"])
	}
	if out["end"].X != 85 {
		t.Fatalf("End 右对齐再减 Margin 应是 85：%+v", out["end"])
	}
	if rect := out["stretch"]; rect.X != 15 || rect.W != 90 {
		t.Fatalf("Stretch 应撑满父宽并两侧各留 Margin：%+v", rect)
	}
}

func TestParentChainScannedWithinWidgetCount(t *testing.T) {
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
	out := layout.Solve(Rect{W: 100, H: 100})
	if len(out) != count {
		t.Fatalf("应求解出 %d 个控件，得到 %d", count, len(out))
	}
	if scanned := layout.Scanned(); scanned > count {
		t.Fatalf("父链只解析一次，scanned=%d 不应超过控件数 %d", scanned, count)
	}
}

func TestSolveIsDeterministicAcrossRuns(t *testing.T) {
	widgets := []Widget{
		{ID: "a", AnchorX: Center, AnchorY: End, Margin: 3, W: 10, H: 10},
		{ID: "b", Parent: "a", AnchorX: Stretch, AnchorY: Stretch, Margin: 2, W: 5, H: 5},
		{ID: "c", Parent: "b", AnchorX: End, AnchorY: Center, Margin: 1, W: 4, H: 4},
	}
	layout := NewLayout(widgets)
	root := Rect{X: 1, Y: 2, W: 300, H: 200}
	first := layout.Solve(root)
	for run := 0; run < 5; run++ {
		again := layout.Solve(root)
		for id, rect := range first {
			if again[id] != rect {
				t.Fatalf("第 %d 次重复求解 %s 不一致：%+v != %+v", run, id, again[id], rect)
			}
		}
	}
}

// referenceSolve 是「父矩形 + 锚点 + 边距 + 记忆化」的独立参考实现，用于对照自证。
func referenceSolve(widgets []Widget, root Rect) map[string]Rect {
	table := map[string]Widget{}
	for _, widget := range widgets {
		table[widget.ID] = widget
	}
	memo := map[string]Rect{}
	var resolve func(id string) Rect
	resolve = func(id string) Rect {
		if rect, ok := memo[id]; ok {
			return rect
		}
		widget := table[id]
		parent := root
		if _, ok := table[widget.Parent]; ok && widget.Parent != "" {
			parent = resolve(widget.Parent)
		}
		rect := Rect{}
		rect.X, rect.W = referenceAxis(parent.X, parent.W, widget.Margin, widget.AnchorX, widget.W)
		rect.Y, rect.H = referenceAxis(parent.Y, parent.H, widget.Margin, widget.AnchorY, widget.H)
		memo[id] = rect
		return rect
	}
	out := map[string]Rect{}
	for _, widget := range widgets {
		out[widget.ID] = resolve(widget.ID)
	}
	return out
}

func referenceAxis(parentPos, parentSize, margin float64, anchor Anchor, own float64) (float64, float64) {
	switch anchor {
	case Center:
		return parentPos + parentSize/2 - own/2, own
	case End:
		return parentPos + parentSize - own - margin, own
	case Stretch:
		return parentPos + margin, parentSize - margin*2
	default:
		return parentPos + margin, own
	}
}

func TestMatchesReferenceOnRandomScenes(t *testing.T) {
	rng := rand.New(rand.NewSource(20260930))
	for scene := 0; scene < 200; scene++ {
		count := 1 + rng.Intn(40)
		widgets := make([]Widget, 0, count)
		for index := 0; index < count; index++ {
			parent := ""
			if index > 0 && rng.Intn(4) != 0 {
				parent = fmt.Sprintf("w-%d", rng.Intn(index))
			}
			widgets = append(widgets, Widget{
				ID:      fmt.Sprintf("w-%d", index),
				Parent:  parent,
				AnchorX: Anchor(rng.Intn(4)),
				AnchorY: Anchor(rng.Intn(4)),
				Margin:  float64(rng.Intn(9)),
				W:       float64(1 + rng.Intn(50)),
				H:       float64(1 + rng.Intn(50)),
			})
		}
		root := Rect{
			X: float64(rng.Intn(20)),
			Y: float64(rng.Intn(20)),
			W: float64(50 + rng.Intn(2000)),
			H: float64(50 + rng.Intn(2000)),
		}
		got := NewLayout(widgets).Solve(root)
		want := referenceSolve(widgets, root)
		for id, wantRect := range want {
			if got[id] != wantRect {
				t.Fatalf("场景 %d 控件 %s 与参考实现不一致：got %+v want %+v", scene, id, got[id], wantRect)
			}
		}
	}
}
