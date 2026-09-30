// Package uianchor 是 UI 锚点布局：按锚点与边距定位控件，分辨率变化后相对位置不变。
package uianchor

// Anchor 是水平或垂直锚点。
type Anchor int

// 锚点取值。
const (
	Start Anchor = iota
	Center
	End
	Stretch
)

// Widget 是一个控件。
type Widget struct {
	ID      string
	Parent  string
	AnchorX Anchor
	AnchorY Anchor
	Margin  float64
	W, H    float64
}

// Rect 是一个矩形。
type Rect struct {
	X, Y, W, H float64
}

// Layout 是布局求解器。
type Layout struct {
	widgets map[string]Widget
	scanned int
}

// NewLayout 建布局。
func NewLayout(widgets []Widget) *Layout {
	table := make(map[string]Widget, len(widgets))
	for _, widget := range widgets {
		table[widget.ID] = widget
	}
	return &Layout{widgets: table}
}

// Solve 求出所有控件的矩形。
// 父链只解析一次：每个控件的父矩形记忆化，整表单次 O(n) 遍历；
// 父级不存在（或父链成环）时退回根容器。
func (l *Layout) Solve(root Rect) map[string]Rect {
	l.scanned = 0
	rects := make(map[string]Rect, len(l.widgets))
	resolving := make(map[string]bool, len(l.widgets))
	var resolve func(id string) Rect
	resolve = func(id string) Rect {
		if rect, ok := rects[id]; ok {
			return rect
		}
		resolving[id] = true
		widget := l.widgets[id]
		parent := root
		if widget.Parent != "" {
			l.scanned++
			if _, ok := l.widgets[widget.Parent]; ok && !resolving[widget.Parent] {
				parent = resolve(widget.Parent)
			}
		}
		delete(resolving, id)
		rect := place(parent, widget)
		rects[id] = rect
		return rect
	}
	for id := range l.widgets {
		resolve(id)
	}
	return rects
}

// place 按锚点与边距把控件放进父矩形。
func place(parent Rect, widget Widget) Rect {
	rect := Rect{}
	rect.X, rect.W = anchor(parent.X, parent.W, widget.W, widget.Margin, widget.AnchorX)
	rect.Y, rect.H = anchor(parent.Y, parent.H, widget.H, widget.Margin, widget.AnchorY)
	return rect
}

// anchor 沿一个轴求控件的位置与尺寸。
func anchor(parentPos, parentSize, size, margin float64, a Anchor) (pos, length float64) {
	switch a {
	case Center:
		return parentPos + (parentSize-size)/2, size
	case End:
		return parentPos + parentSize - size - margin, size
	case Stretch:
		return parentPos + margin, parentSize - 2*margin
	default:
		return parentPos + margin, size
	}
}

// Scanned 是最近一次 Solve 遍历的父链节点数（规模观测）。
func (l *Layout) Scanned() int { return l.scanned }
