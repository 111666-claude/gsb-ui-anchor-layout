// Package uianchor 是 UI 锚点布局：按锚点与边距定位控件，分辨率变化后相对位置不变。
// 缺陷：忽略边距、所有锚点都按左上角处理（居中/右对齐/撑满全失效）、
// 每个控件都向上递归父链（代价随层级增长）。
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
// 缺陷：所有控件都直接贴父矩形左上角（Margin 被忽略），Center/End/Stretch 全按 Start 处理，
// 而且每个控件都重新向上递归一遍父链。
func (l *Layout) Solve(root Rect) map[string]Rect {
	out := map[string]Rect{}
	for id, widget := range l.widgets {
		for name := widget.Parent; name != ""; {
			l.scanned++
			up, ok := l.widgets[name]
			if !ok {
				break
			}
			name = up.Parent
		}
		out[id] = Rect{X: root.X, Y: root.Y, W: widget.W, H: widget.H}
	}
	return out
}

// Scanned 是累计遍历的父链节点数（规模观测）。
func (l *Layout) Scanned() int { return l.scanned }
