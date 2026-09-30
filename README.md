# ui-anchor-layout

UI 锚点布局：按锚点与边距定位控件，分辨率变化后相对位置不变，只用 Go 标准库。

```
go test ./...
go run ./cmd/ui-anchor-layout --sample center
go run ./cmd/ui-anchor-layout --sample margin
go run ./cmd/ui-anchor-layout --sample stretch
```

## 口径（README 为准）

- **父矩形 + 锚点 + 边距**：`Start` 的 X 是 `parent.X + Margin`；`Center` 让控件中心与父中心重合；
  `End` 右对齐再减 `Margin`；`Stretch` 撑满父宽并两侧各留 `Margin`（忽略控件的 `W`）。
- **分辨率无关**：父容器尺寸变化后，居中控件仍与父中心重合，贴合锚点的相对位置不变。
- **一次遍历**：父链只解析一次，`Scanned` 不超过控件数，不许每个控件各自递归父链。
- **不变量**：同一份控件表重复求解结果相同；父级不存在时退回根容器。
- **规模**：单场景 3000 个控件（含深链）、每帧求解一次，单帧 O(n)，内存 O(n)。

## 输出契约（不改格式）

```
x=40 centered=true
x=5
w=90
scanned<=3000
```

## 目录

```
anchor.go               锚点布局求解
cmd/ui-anchor-layout    命令行入口
anchor_test.go          go test 用例
```
