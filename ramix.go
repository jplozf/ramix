package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"fmt"
	"ramix/grummi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ----------------------------------------------------------------------------
// TYPES
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// CONSTANTS
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
var statusMsg *widget.Label // Make statusMsg globally accessible
var gameState grummi.GameState
var myApp fyne.App

// ----------------------------------------------------------------------------
// main()
// ----------------------------------------------------------------------------
func main() {
	myApp = app.NewWithID(APP_ID)
	myApp.Settings().SetTheme(&compactTheme{Theme: theme.DefaultTheme()})
	myWindow = myApp.NewWindow(APP_NAME)
	setMenu()
	myWindow.SetCloseIntercept(func() {
		confirmExit()
	})
	readPreferences()

	boardCellSize = fyne.NewSize(40, 52)
	rackCellSize = fyne.NewSize(30, 39)

	gameState = grummi.InitializeGame(2)
	gameState.CurrentPlayerID = 0 // Assuming player 0 is the human player

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
		widget.NewButton("Valider", func() {
		}),
		widget.NewButton("Trier", func() { // Connect Sort button to grummi.SortTiles
			grummi.SortTiles(gameState.Players[0].Hand)
			refreshRack()
		}),
		widget.NewButton("Piocher", func() {
			gameState.DrawTile()
			refreshRack()
			statusMsg.SetText(fmt.Sprintf("Pioché ! Il reste %d tuiles.", len(gameState.Remaining)))
		}),
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

	// 1. Create the label for messages (the bottom status bar) // Make statusMsg global
	statusMsg = widget.NewLabel("Prêt pour la partie !")

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

	// --- PLACE TILES --- // Populate the rack from the game state
	refreshRack()
	// The final stack: the game underneath, the flying tiles above // Already done in previous step
	finalStack := container.NewStack(content, overlay) // Already done in previous step
	myWindow.SetContent(finalStack)                    // Already done in previous step
	myWindow.Resize(fyne.NewSize(1100, 700))           // Already done in previous step
	myWindow.ShowAndRun()                              // Already done in previous step
}

// ----------------------------------------------------------------------------
// refreshRack()
// ----------------------------------------------------------------------------
// refreshRack clears the player's rack and repopulates it with tiles from the current game state.
func refreshRack() {
	// Clear the current rack display
	for i := 0; i < len(playerRack.Objects); i++ {
		wrapper := playerRack.Objects[i].(*fyne.Container)
		cellStack := wrapper.Objects[0].(*fyne.Container)
		if len(cellStack.Objects) > 1 {
			delete(cellMap, cellStack.Objects[1]) // Remove the old tile from the cellMap
			cellStack.Objects = cellStack.Objects[:1]
			cellStack.Refresh()
		}
	}

	// Add tiles from the player's hand
	for i, tile := range gameState.Players[0].Hand {
		setTileAt(playerRack, i, tile.Value, tile.Color)
	}
	playerRack.Refresh()
}

// ----------------------------------------------------------------------------
// setMenu()
// ----------------------------------------------------------------------------
func setMenu() {
	newItem := fyne.NewMenuItem("Nouvelle Partie", func() { /* Reset logic */ })
	saveItem := fyne.NewMenuItem("Sauvegarder", func() { /* Save logic */ })
	quitItem := fyne.NewMenuItem("Quitter", func() { confirmExit() })
	appearanceMenu := fyne.NewMenu("Affichage",
		fyne.NewMenuItem("Thème Sombre", func() {
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.DarkTheme()})
			myApp.Preferences().SetString("AppTheme", "dark")
			myWindow.Content().Refresh()
		}),
		fyne.NewMenuItem("Thème Clair", func() {
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.LightTheme()})
			myApp.Preferences().SetString("AppTheme", "light")
			myWindow.Content().Refresh()
		}),
	)

	// Add it to our menu bar
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Fichier", newItem, saveItem, quitItem),
		appearanceMenu, // Our new menu
		fyne.NewMenu("Aide", fyne.NewMenuItem("À propos", func() { showAbout(myWindow) })),
	)
	myWindow.SetMainMenu(mainMenu)
}

// ----------------------------------------------------------------------------
// confirmExit()
// ----------------------------------------------------------------------------
func confirmExit() {
	d := dialog.NewConfirm("Confirmation", "Quitter la partie en cours ?", func(confirm bool) {
		if confirm {
			myApp.Quit()
		}
	}, myWindow)
	d.Show()
}

// ----------------------------------------------------------------------------
// showAbout()
// ----------------------------------------------------------------------------
func showAbout(win fyne.Window) {
	info := APP_NAME + "\n" +
		"Version " + getFullVersion() + "\n\n" +
		APP_DESCRIPTION + "\n\n" +
		APP_URL + "\n\n" +
		APP_COPYRIGHT
	dialog.ShowInformation("À propos", info, win)
}

// ----------------------------------------------------------------------------
// readPreferences()
// ----------------------------------------------------------------------------
func readPreferences() {
	themePref := myApp.Preferences().StringWithFallback("AppTheme", "light")
	if themePref == "dark" {
		myApp.Settings().SetTheme(&compactTheme{Theme: theme.DarkTheme()})
	} else {
		myApp.Settings().SetTheme(&compactTheme{Theme: theme.LightTheme()})
	}
}

// ****************************************************************************
// getFullVersion()
// ****************************************************************************
func getFullVersion() string {
	return fmt.Sprintf("%s.%s", MAJOR, GitVersion)
}
