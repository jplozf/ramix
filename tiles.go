package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ----------------------------------------------------------------------------
// TYPES
// ----------------------------------------------------------------------------
type cellInfo struct {
	grid  *fyne.Container
	index int
}

type DragTile struct {
	widget.BaseWidget
	tile    *Tile
	root    *fyne.Container
	parent  *fyne.Container
	phantom *fyne.Container
	grid    *fyne.Container // Original grid
	index   int             // Original index
}

// Tile représente une tuile graphique
type Tile struct {
	Value int    // 1-13, 0 for Joker
	Color string // "Red", "Blue", "Yellow", "Green", "Ivory"
}

// ----------------------------------------------------------------------------
// VARS
// ----------------------------------------------------------------------------
var (
	ColorBoardBackground = color.NRGBA{R: 10, G: 60, B: 30, A: 255}
	ColorRummyRed        = color.NRGBA{R: 210, G: 90, B: 90, A: 255}
	ColorRummyBlue       = color.NRGBA{R: 90, G: 130, B: 180, A: 255}
	ColorRummyYellow     = color.NRGBA{R: 220, G: 170, B: 0, A: 255}
	ColorRummyGreen      = color.NRGBA{R: 60, G: 140, B: 80, A: 255}
	ColorRummyIvory      = color.NRGBA{R: 190, G: 180, B: 160, A: 255}
	ColorIvoryLine       = color.NRGBA{R: 255, G: 255, B: 240, A: 255} // Liseré Ivoire
	ColorTileStroke      = color.NRGBA{R: 190, G: 180, B: 160, A: 255} // Bordure des tuiles
)

var cellMap = make(map[fyne.CanvasObject]cellInfo)

// ----------------------------------------------------------------------------
// Render()
// ----------------------------------------------------------------------------
// Render crée l'objet visuel de la tuile
func (t *Tile) Render() fyne.CanvasObject {
	var bgColor color.Color
	switch t.Color {
	case "Red":
		bgColor = ColorRummyRed
	case "Blue":
		bgColor = ColorRummyBlue
	case "Yellow":
		bgColor = ColorRummyYellow
	case "Ivory":
		bgColor = ColorRummyIvory
	default:
		bgColor = ColorRummyGreen
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
	// A rectangle with the board color for uniformity
	bg := canvas.NewRectangle(ColorBoardBackground)

	// Add a very thin and discrete border to see the grid
	bg.StrokeColor = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	bg.StrokeWidth = 1

	return container.NewStack(bg)
}

// ----------------------------------------------------------------------------
// NewEmptyRackSlot()
// ----------------------------------------------------------------------------
func NewEmptyRackSlot() fyne.CanvasObject {
	// bg := canvas.NewRectangle(color.NRGBA{R: 245, G: 222, B: 179, A: 255}) // Beech
	// bg := canvas.NewRectangle(color.NRGBA{R: 222, G: 184, B: 135, A: 255}) // Pine
	bg := canvas.NewRectangle(color.NRGBA{R: 210, G: 180, B: 140, A: 255}) // Oak
	bg.StrokeColor = color.NRGBA{R: 180, G: 140, B: 100, A: 80}
	bg.StrokeWidth = 2
	bg.CornerRadius = 4

	return container.NewStack(bg)
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
			if s2, ok := s1.Objects[0].(*fyne.Container); ok {
				if rect, ok := s2.Objects[0].(*canvas.Rectangle); ok {
					cellMap[rect] = cellInfo{grid: grid, index: index}
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// setTileAt()
// ----------------------------------------------------------------------------
func setTileAt(grid *fyne.Container, index int, val int, col string) {
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

	tile := &Tile{Value: val, Color: col}
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
func NewDragTile(t *Tile, root *fyne.Container, parent *fyne.Container, grid *fyne.Container, index int) *DragTile {
	dt := &DragTile{tile: t, root: root, parent: parent, grid: grid, index: index}
	dt.ExtendBaseWidget(dt)
	return dt
}

func (dt *DragTile) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(dt.tile.Render())
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
		dt.phantom = container.NewStack(dt.tile.Render())
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
		var targetTile *Tile
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
