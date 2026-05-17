package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"image/color"
	"strconv"

	"ramix/grummi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ----------------------------------------------------------------------------
// TYPES
// ----------------------------------------------------------------------------
type cellInfo struct {
	grid  *fyne.Container
	index int
}

type HoverCell struct {
	widget.BaseWidget
	bg           *canvas.Rectangle
	hoverColor   func(fyne.ThemeVariant) color.Color
	defaultColor func(fyne.ThemeVariant) color.Color
	strokeColor  func(fyne.ThemeVariant) color.Color
	isHovered    bool
}

type DragTile struct {
	widget.BaseWidget
	tile    *grummi.Tile
	root    *fyne.Container
	parent  *fyne.Container
	phantom *fyne.Container
	grid    *fyne.Container // Original grid
	index   int             // Original index
}

// ----------------------------------------------------------------------------
// VARS
// ----------------------------------------------------------------------------
var cellMap = make(map[fyne.CanvasObject]cellInfo)

// ----------------------------------------------------------------------------
// Render()
// ----------------------------------------------------------------------------
// renderTile creates the visual object of the tile
func renderTile(t *grummi.Tile) fyne.CanvasObject {
	var bgColor color.Color
	if t.Value == 0 {
		bgColor = ColorRummyIvory
	} else {
		switch t.Color {
		case grummi.Red:
			bgColor = ColorRummyRed
		case grummi.Blue:
			bgColor = ColorRummyBlue
		case grummi.Green:
			bgColor = ColorRummyGreen
		case grummi.Orange:
			bgColor = ColorRummyYellow
		default:
			bgColor = ColorRummyGreen
		}
	}

	// The rectangle with contrasting edging and rounded corners
	rect := canvas.NewRectangle(bgColor)

	// rect.SetMinSize(fyne.NewSize(30, 40))
	rect.StrokeColor = ColorTileStroke // The new sand/dark color
	rect.StrokeWidth = 4               // Clearly visible border
	rect.CornerRadius = 8

	// Clearly readable black text
	txt := canvas.NewText(strconv.Itoa(t.Value), color.Black)
	if t.Value == 0 {
		txt.Text = "✩" // Joker symbol
	}
	txt.TextStyle = fyne.TextStyle{Bold: true}
	txt.TextSize = 22

	return container.NewStack(
		rect,
		container.NewCenter(txt),
	)
}

// ----------------------------------------------------------------------------
// NewEmptySlot()
// ----------------------------------------------------------------------------
func NewEmptySlot() fyne.CanvasObject {
	bg := canvas.NewRectangle(ColorBoardBackgroundDark)

	bg.StrokeWidth = 1

	defaultCol := func(v fyne.ThemeVariant) color.Color {
		if v == theme.VariantLight {
			return ColorBoardBackgroundLight
		}
		return ColorBoardBackgroundDark
	}
	hoverCol := func(v fyne.ThemeVariant) color.Color {
		if v == theme.VariantLight {
			return ColorBoardCellHoverLight
		}
		return ColorBoardCellHoverDark
	}
	strokeCol := func(v fyne.ThemeVariant) color.Color {
		if v == theme.VariantLight {
			return ColorBoardCellStrokeLight
		}
		return ColorBoardCellStrokeDark
	}
	return NewHoverCell(bg, hoverCol, defaultCol, strokeCol)
}

// ----------------------------------------------------------------------------
// NewEmptyRackSlot()
// ----------------------------------------------------------------------------
func NewEmptyRackSlot() fyne.CanvasObject {
	bg := canvas.NewRectangle(color.Transparent)
	bg.StrokeWidth = 2
	bg.CornerRadius = 4

	defaultCol := func(v fyne.ThemeVariant) color.Color {
		if v == theme.VariantLight {
			return ColorRackCellLight
		}
		return ColorRackCellDark
	}
	hoverCol := func(v fyne.ThemeVariant) color.Color {
		if v == theme.VariantLight {
			return ColorRackCellHoverLight // Cream
		}
		return ColorRackCellHoverDark
	}
	strokeCol := func(v fyne.ThemeVariant) color.Color {
		if v == theme.VariantLight {
			return ColorRackStrokeLight
		}
		return ColorRackStrokeDark
	}
	return NewHoverCell(bg, hoverCol, defaultCol, strokeCol)
}

// ----------------------------------------------------------------------------
// createCell()
// ----------------------------------------------------------------------------
func createCell(size fyne.Size, isBoard bool) fyne.CanvasObject {
	var bg fyne.CanvasObject
	if isBoard {
		bg = NewEmptySlot()
	} else {
		bg = NewEmptyRackSlot()
	}

	// The Stack allows overlaying the background and the future tile
	cellStack := container.NewStack(bg)

	// The GridWrap forces this Stack to maintain the 'size' (aspect ratio)
	// Return this object to be stored in the grid
	return container.NewGridWrap(size, cellStack)
}

// ----------------------------------------------------------------------------
// registerCell()
// ----------------------------------------------------------------------------
func registerCell(wrapper fyne.CanvasObject, grid *fyne.Container, index int) {
	// Navigate the hierarchy to find the background rectangle
	if w, ok := wrapper.(*fyne.Container); ok {
		if s1, ok := w.Objects[0].(*fyne.Container); ok {
			if hc, ok := s1.Objects[0].(*HoverCell); ok {
				cellMap[hc] = cellInfo{grid: grid, index: index}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// setTileAt()
// ----------------------------------------------------------------------------
func setTileAt(grid *fyne.Container, index int, val int, col grummi.Color) { // Changed col type to grummi.Color
	if index < 0 || index >= len(grid.Objects) {
		return
	}

	// 1. Get the GridWrap at the given index
	wrapper, ok := grid.Objects[index].(*fyne.Container)
	if !ok {
		return
	}

	// 2. Get the Stack which is the first (and only) child of the GridWrap
	cellStack, ok := wrapper.Objects[0].(*fyne.Container)
	if !ok {
		return
	}

	// 3. Create the tile visual
	// If a tile already existed, remove it from the cellMap
	if len(cellStack.Objects) > 1 {
		delete(cellMap, cellStack.Objects[1])
	}

	tile := &grummi.Tile{Value: val, Color: col} // Use grummi.Tile
	tileVisual := NewDragTile(tile, overlay, cellStack, grid, index)

	// 4. Also register the tile itself in the cellMap
	// so that handleDrop recognizes it as a valid target (occupied slot)
	cellMap[tileVisual] = cellInfo{grid: grid, index: index}

	// 4. Add it to the Stack (index 0 = background, index 1 = tile)
	if len(cellStack.Objects) > 1 {
		cellStack.Objects[1] = tileVisual
	} else {
		cellStack.Add(tileVisual)
	}

	// Targeted refresh
	cellStack.Refresh()
}

// ----------------------------------------------------------------------------
// NewDragTile()
// ----------------------------------------------------------------------------
func NewDragTile(t *grummi.Tile, root *fyne.Container, parent *fyne.Container, grid *fyne.Container, index int) *DragTile { // Changed t type to grummi.Tile
	dt := &DragTile{tile: t, root: root, parent: parent, grid: grid, index: index}
	dt.ExtendBaseWidget(dt)
	return dt
}

// ----------------------------------------------------------------------------
// CreateRenderer()
// ----------------------------------------------------------------------------
func (dt *DragTile) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(renderTile(dt.tile))
}

// ----------------------------------------------------------------------------
// Dragged()
// ----------------------------------------------------------------------------
// Triggered when dragging starts
func (dt *DragTile) Dragged(e *fyne.DragEvent) {
	if dt.phantom == nil {
		// Hide the original tile during movement
		dt.Hide()
		dt.parent.Refresh()

		// Create the phantom on the first move
		dt.phantom = container.NewStack(renderTile(dt.tile))
		dt.phantom.Resize(dt.Size())
		dt.root.Add(dt.phantom)
	}

	// Absolute mouse position to move the phantom
	// Center the phantom on the cursor
	dt.phantom.Move(e.AbsolutePosition.Subtract(fyne.NewPos(dt.Size().Width/2, dt.Size().Height/2)))
	dt.root.Refresh()
}

// ----------------------------------------------------------------------------
// DragEnd()
// ----------------------------------------------------------------------------
// Triggered when released
func (dt *DragTile) DragEnd() {
	if dt.phantom != nil {
		pos := dt.phantom.Position()
		dt.root.Remove(dt.phantom)
		dt.phantom = nil
		dt.root.Refresh()

		// Detect the drop zone (using the center of the moved tile)
		dropPoint := pos.Add(fyne.NewPos(dt.Size().Width/2, dt.Size().Height/2))

		if !handleDrop(dropPoint, dt) {
			// Drop failed: show the tile back in its place
			dt.Show()
			dt.parent.Refresh()
		}
	}
}

// ----------------------------------------------------------------------------
// handleDrop()
// ----------------------------------------------------------------------------
func handleDrop(absPos fyne.Position, src *DragTile) bool {
	// Get the main stack (finalStack) defined in ramix.go
	stack, ok := myWindow.Content().(*fyne.Container)
	if !ok || len(stack.Objects) < 1 {
		return false
	}

	// Search only in the first object of the stack (the game 'content')
	// to prevent the overlay (the second object) from blocking click detection.
	// Since content is at position (0,0) in the Stack, absPos remains valid.
	obj := findObjectAt(stack.Objects[0], absPos)
	if obj == nil {
		return false
	}

	// If the found object is in the cellMap (either an empty background or an existing tile)
	if target, ok := cellMap[obj]; ok {
		// 1. Save the target tile data if it exists (for the swap)
		var targetTile *grummi.Tile // Use grummi.Tile
		wrapper := target.grid.Objects[target.index].(*fyne.Container)
		cellStack := wrapper.Objects[0].(*fyne.Container)
		if len(cellStack.Objects) > 1 {
			if dt, ok := cellStack.Objects[1].(*DragTile); ok {
				targetTile = dt.tile
			}
		}

		// 2. Remove the source tile from its original location
		if len(src.parent.Objects) > 1 {
			delete(cellMap, src.parent.Objects[1])
			src.parent.Objects = src.parent.Objects[:1]
			src.parent.Refresh()
		}

		// 3. Place the source tile at the destination
		setTileAt(target.grid, target.index, src.tile.Value, src.tile.Color)

		// 4. If the destination was occupied, move the old tile to the source (Swap)
		if targetTile != nil {
			setTileAt(src.grid, src.index, targetTile.Value, targetTile.Color)
		}

		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// findObjectAt()
// ----------------------------------------------------------------------------
// findObjectAt recursively traverses the tree from 'obj' to find the object at 'pos'.
// 'pos' is the position relative to the parent of 'obj'.
func findObjectAt(obj fyne.CanvasObject, pos fyne.Position) fyne.CanvasObject {
	if obj == nil || !obj.Visible() {
		return nil
	}

	// OPTIMIZATION: If we hit a DragTile, we stop there.
	// We don't want to go down to the text or the internal rectangle.
	if _, ok := obj.(*DragTile); ok {
		return obj
	}
	if _, ok := obj.(*HoverCell); ok {
		return obj
	}

	p, s := obj.Position(), obj.Size()
	if pos.X < p.X || pos.Y < p.Y || pos.X > p.X+s.Width || pos.Y > p.Y+s.Height {
		return nil
	}

	localPos := pos.Subtract(p)
	if c, ok := obj.(*fyne.Container); ok {
		for i := len(c.Objects) - 1; i >= 0; i-- {
			if res := findObjectAt(c.Objects[i], localPos); res != nil {
				return res
			}
		}
	}
	return obj
}

// ----------------------------------------------------------------------------
// NewHoverCell()
// ----------------------------------------------------------------------------
func NewHoverCell(bg *canvas.Rectangle, hover, def, stroke func(fyne.ThemeVariant) color.Color) *HoverCell {
	hc := &HoverCell{bg: bg, hoverColor: hover, defaultColor: def, strokeColor: stroke}
	hc.ExtendBaseWidget(hc)

	// Initialize color
	hc.updateColor()
	return hc
}

// ----------------------------------------------------------------------------
// CreateRenderer()
// ----------------------------------------------------------------------------
func (h *HoverCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(h.bg)
}

// ----------------------------------------------------------------------------
// Refresh()
// ----------------------------------------------------------------------------
func (h *HoverCell) Refresh() {
	h.updateColor()
}

// ----------------------------------------------------------------------------
// MouseIn()
// ----------------------------------------------------------------------------
func (h *HoverCell) updateColor() {
	// To handle manual theme switching correctly, we detect if the current
	// theme is light or dark by checking the background color luminance.
	th := fyne.CurrentApp().Settings().Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	// We probe the background color of the current theme using the current variant.
	// This ensures we detect the "effective" brightness of the theme.
	bgCol := th.Color(theme.ColorNameBackground, v)
	r, g, b, _ := bgCol.RGBA()

	variant := theme.VariantDark
	if r+g+b > 0xffff*3/2 {
		variant = theme.VariantLight
	}

	if h.isHovered {
		h.bg.FillColor = h.hoverColor(variant)
	} else {
		h.bg.FillColor = h.defaultColor(variant)
	}
	if h.strokeColor != nil {
		h.bg.StrokeColor = h.strokeColor(variant)
	}
	h.bg.Refresh()
}

// ----------------------------------------------------------------------------
// MouseIn()
// ----------------------------------------------------------------------------
func (h *HoverCell) MouseIn(e *desktop.MouseEvent) {
	h.isHovered = true
	h.Refresh()
}

// ----------------------------------------------------------------------------
// MouseMoved()
// ----------------------------------------------------------------------------
func (h *HoverCell) MouseMoved(e *desktop.MouseEvent) {}

// ----------------------------------------------------------------------------
// MouseOut()
// ----------------------------------------------------------------------------
func (h *HoverCell) MouseOut() {
	h.isHovered = false
	h.Refresh()
}

// ----------------------------------------------------------------------------
// ThemeChanged()
// ----------------------------------------------------------------------------
func (h *HoverCell) ThemeChanged() {
	h.Refresh()
}
