package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ----------------------------------------------------------------------------
// TYPES
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// VARS
// ----------------------------------------------------------------------------
var (
	boardCellSize fyne.Size
	rackCellSize  fyne.Size
)
var boardSlots []fyne.CanvasObject
var rackSlots []fyne.CanvasObject
var myWindow fyne.Window
var gameTable *fyne.Container
var playerRack *fyne.Container
var overlay *fyne.Container

// ----------------------------------------------------------------------------
// main()
// ----------------------------------------------------------------------------
func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(&compactTheme{Theme: theme.DefaultTheme()})
	myWindow = myApp.NewWindow("Ramix")
	boardCellSize = fyne.NewSize(40, 52)
	rackCellSize = fyne.NewSize(30, 39)

	// --- 1. MENU CONFIGURATION ---
	newItem := fyne.NewMenuItem("Nouvelle Partie", func() { /* Reset logic */ })
	saveItem := fyne.NewMenuItem("Sauvegarder", func() { /* Save logic */ })
	quitItem := fyne.NewMenuItem("Quitter", func() { /* Quit logic */ })

	appearanceMenu := fyne.NewMenu("Affichage",
		fyne.NewMenuItem("Thème Sombre", func() {
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.DarkTheme()})
		}),
		fyne.NewMenuItem("Thème Clair", func() {
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.LightTheme()})
		}),
	)

	// Add it to your menu bar
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Fichier", newItem, saveItem, quitItem),
		appearanceMenu, // Our new menu
		fyne.NewMenu("Aide", fyne.NewMenuItem("À propos", func() {})),
	)
	myWindow.SetMainMenu(mainMenu)

	// Creation of simple grids
	gameTable = container.New(layout.NewGridLayoutWithColumns(24))
	for i := 0; i < 192; i++ {
		cell := createCell(boardCellSize, true)
		gameTable.Add(cell)
		registerCell(cell, gameTable, i)
	}
	// 1. Calculate the total size of the board (ex: 24 columns * 40px, 8 rows * 52px)
	tableWidth := boardCellSize.Width * 24
	tableHeight := boardCellSize.Height * 8
	totalTableSize := fyne.NewSize(tableWidth, tableHeight)

	// 2. Force the grid into this size with a global GridWrap
	// This prevents cells from spreading apart
	fixedTable := container.NewGridWrap(totalTableSize, gameTable)

	// 3. Center this rigid block
	// NewCenter will surround the block with empty space, but the tiles inside will stay grouped
	centeredTable := container.NewCenter(fixedTable)

	playerRack = container.New(layout.NewGridLayoutWithColumns(20))
	for i := 0; i < 80; i++ {
		cell := createCell(rackCellSize, false)
		playerRack.Add(cell)
		registerCell(cell, playerRack, i)
	}
	// 1. Calculate the size of the rack (20 columns * width, 4 rows * height)
	rackWidth := rackCellSize.Width * 20
	rackHeight := rackCellSize.Height * 4
	totalRackSize := fyne.NewSize(rackWidth, rackHeight)

	// 2. Freeze the size with a GridWrap
	fixedRack := container.NewGridWrap(totalRackSize, playerRack)

	// --- BOTTOM AREA (Rack + Buttons) ---
	// Put the buttons in a vertical column to the right of the rack
	buttons := container.NewVBox(
		widget.NewButton("Trier", func() { /* ... */ }),
		widget.NewButton("Piocher", func() { /* ... */ }),
		widget.NewButton("Annuler", func() { /* ... */ }),
	)

	// 4. Stick the rack and buttons together
	// The HBox ensures they touch horizontally
	paddedButtons := container.NewPadded(buttons) // Center buttons vertically
	rackWithButtons := container.NewHBox(fixedRack, paddedButtons)

	// 5. Center the assembly so it is in the middle of the bottom of the screen
	centeredBottom := container.NewCenter(rackWithButtons) // Assemble the rack and its buttons
	// HBox places elements side by side without extra space
	// bottomArea := container.NewHBox(playerRack, buttons)

	// 1. Create the label for messages (the bottom status bar)
	statusMsg := widget.NewLabel("Prêt pour la partie !")

	// 2. Create a vertical container for the South area
	// Use VBox to stack the elements
	southArea := container.NewVBox(
		centeredBottom, // Your centered rack and buttons
		statusMsg,      // The small text bar at the very bottom
	)

	// --- RIGHT ZONE (Status) ---
	// Ensure the status has a title or background to be visible
	statusLabel := widget.NewLabel("TOUR : Joueur 1\nScore : 0")
	statusArea := container.NewVBox(
		widget.NewLabelWithStyle("STATUS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		statusLabel,
	)
	// The classic BorderLayout assembly

	// 2. Final assembly with BorderLayout
	content := container.NewBorder(
		nil,           // Top
		southArea,     // Bottom (Rack + Buttons)
		nil,           // Left
		statusArea,    // Right (Status)
		centeredTable, // Center (The game table takes the rest)
	)

	// The overlay is an empty container that will cover the entire window
	overlay = container.NewWithoutLayout()

	// --- PLACE TILES ---

	// Place a "Green 13" on the first row, 5th slot of the board
	// Index = (row * columns) + column -> (0 * 24) + 4 = 4
	setTileAt(gameTable, 4, 13, "Green")
	setTileAt(gameTable, 5, 13, "Red")
	setTileAt(gameTable, 6, 13, "Blue")
	setTileAt(gameTable, 7, 13, "Yellow")

	// Place a "Joker" on the rack, row 2, slot 1
	// Index = (1 * 20) + 0 = 20
	setTileAt(playerRack, 7, 0, "Ivory")

	// The final stack: the game underneath, the flying tiles above
	finalStack := container.NewStack(content, overlay)
	myWindow.SetContent(finalStack)
	myWindow.Resize(fyne.NewSize(1100, 700))
	myWindow.ShowAndRun()
}
