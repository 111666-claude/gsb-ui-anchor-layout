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
	table := map[string]Widget{}
	for _, widget := range widgets {
		table[widget.ID] = widget
	}
	return &Layout{widgets: table}
}

// Solve 求出所有控件的矩形。
// 每个控件的父链只解析一次（记忆化），整帧 O(n)、内存 O(n)；
// 父级不存在或成环时退回根容器。
func (l *Layout) Solve(root Rect) map[string]Rect {
	l.scanned = 0
	out := map[string]Rect{}
	visiting := map[string]bool{}

	var resolve func(id string) Rect
	resolve = func(id string) Rect {
		if rect, ok := out[id]; ok {
			return rect
		}
		widget := l.widgets[id]
		parentRect := root
		if widget.Parent != "" {
			if _, ok := l.widgets[widget.Parent]; ok {
				l.scanned++
				if visiting[widget.Parent] {
					parentRect = root
				} else {
					visiting[id] = true
					parentRect = resolve(widget.Parent)
					delete(visiting, id)
				}
			}
		}
		rect := place(widget, parentRect)
		out[id] = rect
		return rect
	}

	for id := range l.widgets {
		resolve(id)
	}
	return out
}

// place 按「父矩形 + 锚点 + 边距」放置控件。
func place(widget Widget, parent Rect) Rect {
	x, w := placeAxis(parent.X, parent.W, widget.Margin, widget.AnchorX, widget.W)
	y, h := placeAxis(parent.Y, parent.H, widget.Margin, widget.AnchorY, widget.H)
	return Rect{X: x, Y: y, W: w, H: h}
}

// placeAxis 求单轴上的位置与尺寸。
func placeAxis(parentPos, parentSize, margin float64, anchor Anchor, ownSize float64) (pos, size float64) {
	switch anchor {
	case Center:
		return parentPos + (parentSize-ownSize)/2, ownSize
	case End:
		return parentPos + parentSize - margin - ownSize, ownSize
	case Stretch:
		return parentPos + margin, parentSize - 2*margin
	default:
		return parentPos + margin, ownSize
	}
}

// Scanned 是本次求解沿父链访问的父级节点数（每控件至多一次，不超过控件数）。
func (l *Layout) Scanned() int { return l.scanned }
