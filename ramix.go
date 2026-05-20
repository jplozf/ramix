package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"fmt"
	"image/color"
	"ramix/grummi"
	"regexp"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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
var statusMsg *widget.Label
var statusNames []*widget.Label
var statusP1NameLabel *widget.Label
var statusLabel *widget.Label
var statusDrawLabel *widget.Label
var statusTiles []*widget.Label
var gameState grummi.GameState
var myApp fyne.App
var background *canvas.Rectangle
var aiLogEntry *widget.Label
var humanPool []grummi.Tile // Added for human player's temporary tiles
var aiLogScroll *container.Scroll

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// stripANSI removes ANSI escape sequences (like colors) from a string.
func stripANSI(str string) string {
	str = ansiRegex.ReplaceAllString(str, "")

	// Replace emojis with text equivalents to avoid rendering issues on some platforms
	replacer := strings.NewReplacer(
		"😁", "[Joker]",
		"🔴", "(R)",
		"🔵", "(B)",
		"🟢", "(G)",
		"🟠", "(O)",
		"🤖", "[AI]",
		"🎲", "Roll:",
		"🧩", "Table",
		"📥", "Pool",
		"🖐️", "Hand",
		"✨", "*",
		"⭐", "!",
		"🎉", "!",
		"🏆", "WIN",
		"📊", "Stats",
		"👤", "P",
		"🃏", "Deck",
		"✔", "OK",
		"✖", "X",
		"►", ">",
	)
	return replacer.Replace(str)
}

// uiLogger implements the grummi.Logger interface to direct logs to the UI.
type uiLogger struct{}

// Log sends messages to the UI's status bar and log panel.
func (l *uiLogger) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fyne.Do(func() {
		SetStatus(msg)
		refreshTable()
		refreshRack()
	})
	time.Sleep(600 * time.Millisecond) // Short delay to let the user see the move
}

// ----------------------------------------------------------------------------
// main()
// ----------------------------------------------------------------------------
func main() {

	myApp = app.NewWithID(APP_ID)
	myApp.SetIcon(resourceRamixPng)
	myApp.Settings().SetTheme(&compactTheme{Theme: theme.DefaultTheme()})

	myWindow = myApp.NewWindow(APP_NAME)
	setMenu()
	myWindow.SetCloseIntercept(func() {
		confirmExit()
	})

	boardCellSize = fyne.NewSize(40, 52)
	rackCellSize = fyne.NewSize(30, 39)

	gameState = grummi.InitializeGame(2, &uiLogger{})
	humanPool = []grummi.Tile{} // Initialize the pool for the human player

	// The main game table
	gameTable = container.New(layout.NewGridLayoutWithColumns(24))
	for i := range 192 {
		cell := createCell(boardCellSize, true)
		gameTable.Add(cell)
		registerCell(cell, gameTable, i)
	}
	tableWidth := boardCellSize.Width * 24
	tableHeight := boardCellSize.Height * 8
	totalTableSize := fyne.NewSize(tableWidth, tableHeight)

	fixedTable := container.NewGridWrap(totalTableSize, gameTable)

	// The bottom area
	aiLogEntry = widget.NewLabel("")
	aiLogEntry.Wrapping = fyne.TextWrapWord
	aiLogScroll = container.NewScroll(aiLogEntry)
	aiLogScroll.SetMinSize(fyne.NewSize(250, 100))

	playerRack = container.New(layout.NewGridLayoutWithColumns(20))
	for i := range 80 { // 4 rows * 20 columns
		cell := createCell(rackCellSize, false)
		playerRack.Add(cell)
		registerCell(cell, playerRack, i)
	}
	rackWidth := rackCellSize.Width * 20
	rackHeight := rackCellSize.Height * 4
	totalRackSize := fyne.NewSize(rackWidth, rackHeight)

	fixedRack := container.NewGridWrap(totalRackSize, playerRack)

	buttons := container.NewVBox(
		widget.NewButtonWithIcon("Valider", theme.ConfirmIcon(), func() {
			if syncUItoGameState() {
				gameState.CurrentPlayerID = (gameState.CurrentPlayerID + 1) % len(gameState.Players)
				gameState.TurnNumber++
				playNextTurn()
			}
		}),
		widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
			grummi.SortTiles(gameState.Players[0].Hand)
			refreshRack()
			SetStatus("Tri des tuiles")
		}),
		widget.NewButtonWithIcon("Piocher", theme.ContentAddIcon(), func() {
			gameState.DrawTile()
			refreshRack()
			SetStatus(fmt.Sprintf("Pioché ! Il reste %d tuiles.", len(gameState.Remaining)))
			gameState.CurrentPlayerID = (gameState.CurrentPlayerID + 1) % len(gameState.Players) // End human turn
			playNextTurn()
		}),
		widget.NewButtonWithIcon("Passer", theme.CancelIcon(), func() { // Pass button
			SetStatus("Vous avez passé votre tour.")
			gameState.CurrentPlayerID = (gameState.CurrentPlayerID + 1) % len(gameState.Players)
			playNextTurn()
		}),
	)

	gapBetweenRackAndButtons := canvas.NewRectangle(color.Transparent)
	gapBetweenRackAndButtons.SetMinSize(fyne.NewSize(10, 0))
	rackAndButtonsContainer := container.NewHBox(gapBetweenRackAndButtons, fixedRack, gapBetweenRackAndButtons, container.NewPadded(buttons))

	rackAssembly := container.NewBorder(nil, nil, aiLogScroll, nil, rackAndButtonsContainer)
	statusMsg = widget.NewLabel("")
	centeredBottom := container.NewBorder(nil, statusMsg, nil, nil, rackAssembly)

	// The status area on the right
	statusLabel = widget.NewLabelWithStyle("1", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	statusDrawLabel = widget.NewLabelWithStyle("0", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	statusP1NameLabel = widget.NewLabelWithStyle("P1", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	statusTiles = make([]*widget.Label, 4)
	for i := range 4 {
		statusTiles[i] = widget.NewLabelWithStyle("-", fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})
	}

	// shrink wraps an object in a fixed-height container to force tighter row spacing.
	shrink := func(obj fyne.CanvasObject, width float32) fyne.CanvasObject {
		return container.NewGridWrap(fyne.NewSize(width, 20), obj)
	}

	// Use FormLayout to align labels and values in two consistent columns
	statusDetails := container.New(layout.NewFormLayout(),
		shrink(widget.NewLabel("Tour"), 100), shrink(statusLabel, 40),
		shrink(widget.NewLabel("Pioche"), 100), shrink(statusDrawLabel, 40),
		shrink(statusP1NameLabel, 100), shrink(statusTiles[0], 40),
		shrink(widget.NewLabel("AI#1"), 100), shrink(statusTiles[1], 40),
		shrink(widget.NewLabel("AI#2"), 100), shrink(statusTiles[2], 40),
		shrink(widget.NewLabel("AI#3"), 100), shrink(statusTiles[3], 40),
	)

	statusTitle := widget.NewLabelWithStyle("STATUS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	statusArea := container.NewVBox(
		container.NewCenter(shrink(statusTitle, 140)),
		statusDetails,
	)

	refreshRack()
	readPreferences()
	refreshTable() // Display the initial (empty) table

	// The final layout
	overlay = container.NewWithoutLayout()
	background = canvas.NewRectangle(color.Transparent)

	// Group the table and status area together in an HBox to remove the gap between them,
	// then center the combined assembly in the main window area.
	tableAndStatus := container.NewCenter(container.NewHBox(fixedTable, statusArea))
	mainInterface := container.NewBorder(nil, centeredBottom, nil, nil, tableAndStatus)

	windowContent := container.NewStack(background, mainInterface)
	finalStack := container.NewStack(windowContent, overlay)

	// Let's show the window and run the app
	updateBackgroundColor()
	myWindow.SetContent(finalStack)
	myWindow.Resize(fyne.NewSize(1100, 700))

	SetStatus(grummi.T("status_welcome"))
	showNewGameDialog(myWindow, onNewGame)
	// The game loop will start after the dialog is closed and onNewGame is called.
	myWindow.ShowAndRun()
}

// ----------------------------------------------------------------------------
// updateBackgroundColor()
// ----------------------------------------------------------------------------
func updateBackgroundColor() {
	if myApp.Preferences().StringWithFallback("AppTheme", "light") == "dark" {
		background.FillColor = ColorBackgroundGlobalDark
	} else {
		background.FillColor = ColorBackgroundGlobalLight
	}
	background.Refresh()
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

	// Add tiles from the player's hand using a more compact 4-row layout
	for _, tile := range gameState.Players[0].Hand {
		if tile.Value != 0 {
			// Place regular tiles by color and value
			row := int(tile.Color)
			col := tile.Value - 1
			idx := row*20 + col

			if isCellOccupied(playerRack, idx) {
				// Handle duplicates: place in the 'overflow' area (cols 13-19)
				for c := 13; c < 20; c++ {
					altIdx := row*20 + c
					if !isCellOccupied(playerRack, altIdx) {
						setTileAt(playerRack, altIdx, tile.Value, tile.Color)
						break
					}
				}
			} else {
				setTileAt(playerRack, idx, tile.Value, tile.Color)
			}
		} else {
			// Joker: place in the first available overflow slot (cols 13-19) in any row
			placed := false
			for r := 0; r < 4 && !placed; r++ {
				for c := 13; c < 20; c++ {
					idx := r*20 + c
					if !isCellOccupied(playerRack, idx) {
						setTileAt(playerRack, idx, tile.Value, tile.Color)
						placed = true
						break
					}
				}
			}
		}
	}

	updateStatusTiles()
	playerRack.Refresh()
}

// ----------------------------------------------------------------------------
// refreshTable()
// ----------------------------------------------------------------------------
// refreshTable clears the game table and repopulates it with combinations from the current game state.
func refreshTable() {
	// Clear the current table display
	for i := 0; i < len(gameTable.Objects); i++ {
		wrapper := gameTable.Objects[i].(*fyne.Container)
		cellStack := wrapper.Objects[0].(*fyne.Container)
		if len(cellStack.Objects) > 1 {
			// Remove the old DragTile from the cellMap if it exists
			if dt, ok := cellStack.Objects[1].(*DragTile); ok {
				delete(cellMap, dt)
			}
			cellStack.Objects = cellStack.Objects[:1] // Keep only the background
			cellStack.Refresh()
		}
	}

	const maxCols = 24 // Number of columns in gameTable
	groupCount := 0

	for _, combo := range gameState.Table {
		if grummi.IsValidRun(combo) {
			// Logic for Runs: Fixed positions by color and value on the left (cols 0-12)
			c := getRunColor(combo)
			// We assign 2 rows per color (8 rows total for 4 colors)
			row := int(c) * 2

			// Check if the first designated row for this color is already occupied
			// at the required positions to handle duplicate runs of the same color.
			isRow0Occupied := false
			for i := range combo {
				val := getTileValueInRun(combo, i)
				idx := row*maxCols + (val - 1)
				if idx >= 0 && idx < 192 && isCellOccupied(gameTable, idx) {
					isRow0Occupied = true
					break
				}
			}
			if isRow0Occupied {
				row++ // Use the second row for this color
			}

			for i, tile := range combo {
				val := getTileValueInRun(combo, i)
				if val >= 1 && val <= 13 {
					setTileAt(gameTable, row*maxCols+(val-1), tile.Value, tile.Color)
				}
			}
		} else {
			// Logic for Groups: Place them on the right side (columns 15-23)
			// We fit 2 groups per row: cols 15-18 and 20-23.
			r := groupCount / 2
			cOffset := 15
			if groupCount%2 == 1 {
				cOffset = 20
			}

			if r < 8 {
				for i, tile := range combo {
					if i < 4 {
						setTileAt(gameTable, r*maxCols+cOffset+i, tile.Value, tile.Color)
					}
				}
				groupCount++
			}
		}
	}
	gameTable.Refresh()
}

// syncUItoGameState reads the current state of the UI grids and updates the underlying game logic.
// It returns true if the current table configuration is valid.
func syncUItoGameState() bool {
	// 1. Extract Hand from Rack
	newHand := []grummi.Tile{}
	for i := 0; i < 80; i++ {
		if t := getTileAtCell(playerRack, i); t != nil {
			newHand = append(newHand, *t)
		}
	}

	// 2. Extract Table combinations from GameTable
	// We scan row by row, grouping contiguous tiles into combinations.
	newTable := [][]grummi.Tile{}
	var currentCombo []grummi.Tile
	const cols = 24

	for i := 0; i < 192; i++ {
		t := getTileAtCell(gameTable, i)
		if t != nil {
			currentCombo = append(currentCombo, *t)
		}

		// A combination ends if we hit an empty cell or the end of a row
		isEndOfRow := (i+1)%cols == 0
		if (t == nil || isEndOfRow) && len(currentCombo) > 0 {
			newTable = append(newTable, currentCombo)
			currentCombo = nil
		}
	}

	// 3. Validation Logic
	// Check if all combinations on the table are valid
	for _, combo := range newTable {
		if !grummi.IsValidCombination(combo) {
			SetStatus("Mouvement invalide : vérifiez vos combinaisons sur la table !")
			return false
		}
	}

	// 4. Handle the "Opening" rule (30 points minimum for the first play)
	if !gameState.Players[0].HasPlayedFirst {
		oldVal := calculateTableValue(gameState.Table)
		newVal := calculateTableValue(newTable)
		playedPoints := newVal - oldVal

		// Did the player actually play anything?
		if len(newHand) == len(gameState.Players[0].Hand) {
			SetStatus("Vous n'avez posé aucune tuile. Piochez ou passez votre tour.")
			return false
		}

		if playedPoints < 30 {
			SetStatus(fmt.Sprintf("Ouverture refusée : %d/30 points requis.", playedPoints))
			return false
		}
		gameState.Players[0].HasPlayedFirst = true
		SetStatus("Félicitations ! Ouverture validée.")
	}

	// 5. Update the game state if all checks pass
	gameState.Players[0].Hand = newHand
	gameState.Table = newTable
	return true
}

// getTileAtCell is a helper to retrieve the grummi.Tile pointer from a specific cell in a grid.
func getTileAtCell(grid *fyne.Container, idx int) *grummi.Tile {
	if idx < 0 || idx >= len(grid.Objects) {
		return nil
	}
	// Structure: GridWrap -> Stack -> [HoverCell, DragTile]
	wrapper := grid.Objects[idx].(*fyne.Container)
	cellStack := wrapper.Objects[0].(*fyne.Container)
	if len(cellStack.Objects) > 1 {
		if dt, ok := cellStack.Objects[1].(*DragTile); ok {
			return dt.tile
		}
	}
	return nil
}

// calculateTableValue sums the points of all combinations currently on the table.
func calculateTableValue(table [][]grummi.Tile) int {
	total := 0
	for _, combo := range table {
		// We use the exported GetComboValueWithJoker from the grummi package
		total += grummi.GetComboValueWithJoker(combo, grummi.IsValidRun(combo))
	}
	return total
}

// isCellOccupied checks if a specific cell in the gameTable contains a tile.
func isCellOccupied(grid *fyne.Container, idx int) bool {
	if idx < 0 || idx >= len(grid.Objects) {
		return false
	}
	wrapper := grid.Objects[idx].(*fyne.Container)
	cellStack := wrapper.Objects[0].(*fyne.Container)
	return len(cellStack.Objects) > 1
}

// getRunColor returns the color of a run by finding the first non-joker tile.
func getRunColor(combo []grummi.Tile) grummi.Color {
	for _, t := range combo {
		if t.Value != 0 {
			return t.Color
		}
	}
	return grummi.Red
}

// getTileValueInRun deduces the intended value of a tile (including jokers) within a run.
func getTileValueInRun(combo []grummi.Tile, index int) int {
	t := combo[index]
	if t.Value != 0 {
		return t.Value
	}

	// For a Joker (Value 0), we must determine its logical value in the run.
	// We scan the combination to find internal gaps and then fill ends, matching grummi's scoring logic.
	var realTiles []int
	jokerCount := 0
	for _, tile := range combo {
		if tile.Value == 0 {
			jokerCount++
		} else {
			realTiles = append(realTiles, tile.Value)
		}
	}

	if len(realTiles) == 0 {
		return 0
	}
	sort.Ints(realTiles)

	assignedValues := make(map[int]int)
	used := 0
	// 1. Fill internal gaps (e.g., between 1 and 3)
	for i := 0; i < len(realTiles)-1; i++ {
		for v := realTiles[i] + 1; v < realTiles[i+1]; v++ {
			if used < jokerCount {
				assignedValues[used] = v
				used++
			}
		}
	}
	// 2. Fill high end and then low end
	high := realTiles[len(realTiles)-1]
	for high < 13 && used < jokerCount {
		high++
		assignedValues[used] = high
		used++
	}
	low := realTiles[0]
	for low > 1 && used < jokerCount {
		low--
		assignedValues[used] = low
		used++
	}

	thisJokerIdx := 0
	for i := 0; i < index; i++ {
		if combo[i].Value == 0 {
			thisJokerIdx++
		}
	}
	return assignedValues[thisJokerIdx]
}

// ----------------------------------------------------------------------------
// setMenu()
// ----------------------------------------------------------------------------
func setMenu() {
	newItem := fyne.NewMenuItem(grummi.T("menu_new_game"), func() { showNewGameDialog(myWindow, onNewGame) })
	saveItem := fyne.NewMenuItem(grummi.T("menu_save"), func() { /* Save logic */ })
	quitItem := fyne.NewMenuItem(grummi.T("menu_quit"), func() { confirmExit() })
	appearanceMenu := fyne.NewMenu(grummi.T("menu_display"),
		fyne.NewMenuItem(grummi.T("menu_theme_dark"), func() {
			SetStatus("Application du thème sombre")
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.DarkTheme()})
			myApp.Preferences().SetString("AppTheme", "dark")
			updateBackgroundColor()
			myWindow.Content().Refresh()
		}),
		fyne.NewMenuItem(grummi.T("menu_theme_light"), func() {
			SetStatus("Application du thème clair")
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.LightTheme()})
			myApp.Preferences().SetString("AppTheme", "light")
			updateBackgroundColor()
			myWindow.Content().Refresh()
		}),
	)

	languageMenu := fyne.NewMenu(grummi.T("menu_language"),
		fyne.NewMenuItem("English", func() {
			grummi.SetLanguage("en")
			myApp.Preferences().SetString("AppLanguage", "en")
			SetStatus(grummi.T("status_lang_changed", "English"))
			setMenu() // Refresh menu labels
		}),
		fyne.NewMenuItem("Français", func() {
			grummi.SetLanguage("fr")
			myApp.Preferences().SetString("AppLanguage", "fr")
			SetStatus(grummi.T("status_lang_changed", "Français"))
			setMenu() // Refresh menu labels
		}),
	)

	// Add it to our menu bar
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu(grummi.T("menu_file"), newItem, saveItem, quitItem),
		appearanceMenu, // Our new menu
		languageMenu,
		fyne.NewMenu("Aide", fyne.NewMenuItem("À propos", func() { showAbout(myWindow) })),
	)
	myWindow.SetMainMenu(mainMenu)
}

// ----------------------------------------------------------------------------
// confirmExit()
// ----------------------------------------------------------------------------
func confirmExit() {
	SetStatus("Confirmation pour quitter")
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
	SetStatus("À propos")
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
	SetStatus("Lecture des préférences")
	langPref := myApp.Preferences().StringWithFallback("AppLanguage", "fr")
	grummi.SetLanguage(langPref)
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

// ****************************************************************************
// updateStatusTiles()
// ****************************************************************************
func updateStatusTiles() {
	if len(gameState.Players) > 0 {
		statusLabel.SetText(fmt.Sprintf("%d", gameState.TurnNumber)) // Display turn number
		statusP1NameLabel.SetText(gameState.Players[0].Name)
		statusDrawLabel.SetText(fmt.Sprintf("%d", len(gameState.Remaining)))
	}

	for i := 0; i < 4; i++ {
		if i < len(gameState.Players) {
			p := gameState.Players[i]
			statusTiles[i].SetText(fmt.Sprintf("%d", len(p.Hand)))
		} else {
			statusTiles[i].SetText("-")
		}
	}
}

// ****************************************************************************
// SetStatus()
// ****************************************************************************
func SetStatus(msg string) {
	msg = stripANSI(msg)
	statusMsg.SetText(msg)
	appendAIMessage(msg)
}

// ****************************************************************************
// showNewGameDialog()
// ****************************************************************************
func showNewGameDialog(win fyne.Window, startCallback func(playerName string, aiCount int)) {
	// 1. Champ pour le nom du joueur
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Entrez votre nom...")

	// Optionnel: On peut recharger le dernier nom utilisé depuis les préférences
	nameEntry.SetText(fyne.CurrentApp().Preferences().StringWithFallback("PlayerName", "Humain"))

	// 2. Sélecteur pour le nombre d'adversaires (de 1 à 3)
	aiSelect := widget.NewSelect([]string{"1", "2", "3"}, nil)
	aiSelect.SetSelected("3") // Valeur par défaut

	// 3. Mise en page du formulaire
	form := widget.NewForm(
		widget.NewFormItem(grummi.T("label_your_name"), nameEntry),
		widget.NewFormItem(grummi.T("label_ai_opponents"), aiSelect),
	)

	// 4. Création du dialogue avec boutons Confirmer/Annuler
	dialog.ShowCustomConfirm(
		grummi.T("dialog_new_game_title"), // Titre
		grummi.T("btn_start"),             // Bouton de validation
		grummi.T("btn_cancel"),            // Bouton d'annulation
		form,                              // Le contenu du formulaire
		func(confirmed bool) {
			if confirmed {
				// On convertit le choix de l'IA en entier
				aiCount := 1
				switch aiSelect.Selected {
				case "2":
					aiCount = 2
				case "3":
					aiCount = 3
				}

				nomJoueur := nameEntry.Text
				if nomJoueur == "" {
					nomJoueur = "Humain" // Sécurité si le nom est vide
				}

				// On sauvegarde le nom pour la prochaine fois
				fyne.CurrentApp().Preferences().SetString("PlayerName", nomJoueur)

				// On lance le callback avec les données récupérées
				startCallback(nomJoueur, aiCount)
			}
		},
		win,
	)
}

// ****************************************************************************
// onNewGame()
// ****************************************************************************
func onNewGame(name string, ais int) {
	gameLogger := &uiLogger{}
	gameState = grummi.InitializeGame(ais+1, gameLogger) // Pass the UI logger to the game state
	gameState.Players[0].Name = name

	gameState.CurrentPlayerID = gameState.DetermineFirstPlayer() // This now logs the message
	refreshRack()
	refreshTable() // Refresh table after new game initialization

	// Start the turn sequence after a new game is initialized
	playNextTurn()
}

// ****************************************************************************
// playNextTurn()
// ****************************************************************************
// playNextTurn manages the game flow, handling AI turns automatically
// and setting up for the human player's turn.
func playNextTurn() {
	go func() {
		// Loop through AI turns until it's the human player's turn
		for gameState.Players[gameState.CurrentPlayerID].IsAI {
			currentPlayer := &gameState.Players[gameState.CurrentPlayerID]

			// Initial "thinking" pause
			time.Sleep(1 * time.Second)

			// Execute AI turn
			// Note: The UI now refreshes inside IATurn via the Logger
			gameState.IATurn(currentPlayer)

			// Final refresh after turn
			fyne.Do(func() {
				refreshTable()
				refreshRack()
			})

			// Move to the next player
			gameState.CurrentPlayerID = (gameState.CurrentPlayerID + 1) % len(gameState.Players)
		}

		// It's now the human player's turn
		fyne.Do(func() {
			SetStatus(fmt.Sprintf("C'est à votre tour, %s !", gameState.Players[0].Name))
		})
	}()
}

// ****************************************************************************
// appendAIMessage()
// ****************************************************************************
func appendAIMessage(msg string) {
	currentText := aiLogEntry.Text
	if currentText != "" {
		aiLogEntry.SetText(currentText + "\n> " + msg)
	} else {
		aiLogEntry.SetText("> " + msg)
	}

	aiLogScroll.ScrollToBottom()
}
