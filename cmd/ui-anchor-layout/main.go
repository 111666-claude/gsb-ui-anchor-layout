// Command ui-anchor-layout 跑 UI 锚点样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/uianchor"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("ui-anchor-layout", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "center", "center / margin / stretch / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	root := uianchor.Rect{W: 100, H: 100}
	switch *sample {
	case "center":
		layout := uianchor.NewLayout([]uianchor.Widget{
			{ID: "w", AnchorX: uianchor.Center, AnchorY: uianchor.Center, W: 20, H: 20},
		})
		rect := layout.Solve(root)["w"]
		centered := rect.X+rect.W/2 == root.W/2 && rect.Y+rect.H/2 == root.H/2
		fmt.Fprintf(stdout, "x=%.0f centered=%v\n", rect.X, centered)
	case "margin":
		layout := uianchor.NewLayout([]uianchor.Widget{
			{ID: "w", AnchorX: uianchor.Start, Margin: 5, W: 20, H: 20},
		})
		fmt.Fprintf(stdout, "x=%.0f\n", layout.Solve(root)["w"].X)
	case "stretch":
		layout := uianchor.NewLayout([]uianchor.Widget{
			{ID: "w", AnchorX: uianchor.Stretch, Margin: 5, W: 20, H: 20},
		})
		fmt.Fprintf(stdout, "w=%.0f\n", layout.Solve(root)["w"].W)
	case "scan":
		widgets := make([]uianchor.Widget, 0, 3000)
		for index := 0; index < 3000; index++ {
			parent := ""
			if index > 0 {
				parent = fmt.Sprintf("w-%d", index-1)
			}
			widgets = append(widgets, uianchor.Widget{ID: fmt.Sprintf("w-%d", index), Parent: parent, W: 10})
		}
		layout := uianchor.NewLayout(widgets)
		layout.Solve(root)
		fmt.Fprintf(stdout, "scanned=%d\n", layout.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample center|margin|stretch|scan")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
