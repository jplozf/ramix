package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"fmt"
	"image/color"
	"ramix/grummi"

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
var aiLogEntry *widget.TextGrid
var aiLogScroll *container.Scroll

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

	gameState = grummi.InitializeGame(2)
	gameState.CurrentPlayerID = 0

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
	aiLogEntry = widget.NewTextGrid()
	aiLogScroll = container.NewScroll(aiLogEntry)
	aiLogScroll.SetMinSize(fyne.NewSize(250, 100))

	playerRack = container.New(layout.NewGridLayoutWithColumns(20))
	for i := range 80 {
		cell := createCell(rackCellSize, false)
		playerRack.Add(cell)
		registerCell(cell, playerRack, i)
	}
	rackWidth := rackCellSize.Width * 20
	rackHeight := rackCellSize.Height * 4
	totalRackSize := fyne.NewSize(rackWidth, rackHeight)

	fixedRack := container.NewGridWrap(totalRackSize, playerRack)

	buttons := container.NewVBox(
		widget.NewButtonWithIcon("", theme.ConfirmIcon(), func() { /* ... */ }),
		widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
			grummi.SortTiles(gameState.Players[0].Hand)
			refreshRack()
			SetStatus("Tri des tuiles")
		}),
		widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			gameState.DrawTile()
			refreshRack()
			SetStatus(fmt.Sprintf("Pioché ! Il reste %d tuiles.", len(gameState.Remaining)))
		}),
		widget.NewButtonWithIcon("", theme.CancelIcon(), func() { /* ... */ }),
	)

	gapBetweenRackAndButtons := canvas.NewRectangle(color.Transparent)
	gapBetweenRackAndButtons.SetMinSize(fyne.NewSize(10, 0))
	rackAndButtonsContainer := container.NewHBox(fixedRack, gapBetweenRackAndButtons, container.NewPadded(buttons))

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

	// Add tiles from the player's hand
	for i, tile := range gameState.Players[0].Hand {
		setTileAt(playerRack, i, tile.Value, tile.Color)
	}
	updateStatusTiles()
	playerRack.Refresh()
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
	gameState = grummi.InitializeGame(ais + 1)
	gameState.Players[0].Name = name
	gameState.CurrentPlayerID = 0

	refreshRack()
	SetStatus(fmt.Sprintf("Nouvelle partie : %s vs %d AI", name, ais))
}

// ****************************************************************************
// appendAIMessage()
// ****************************************************************************
func appendAIMessage(msg string) {
	currentText := aiLogEntry.Text()
	if currentText != "" {
		aiLogEntry.SetText(currentText + "\n► " + msg)
	} else {
		aiLogEntry.SetText("► " + msg)
	}

	aiLogScroll.ScrollToBottom()
}
