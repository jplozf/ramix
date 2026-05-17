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
var statusLabel *widget.Label
var statusDrawLabel *widget.Label
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
	centeredTable := container.NewCenter(fixedTable)

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
	statusLabel = widget.NewLabel("TOUR : Joueur 1\nScore : 0")
	statusDrawLabel = widget.NewLabel("PIOCHE : 0")
	statusArea := container.NewVBox(
		widget.NewLabelWithStyle("STATUS", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		statusLabel,
		statusDrawLabel,
	)

	refreshRack()
	readPreferences()

	// The final layout
	overlay = container.NewWithoutLayout()
	background = canvas.NewRectangle(color.Transparent)

	// Hierarchy fix: By nesting the borders, we ensure centeredBottom (the bottom bar)
	// stays to the left of the statusArea (the sidebar) instead of spanning the full window width.
	contentArea := container.NewBorder(nil, centeredBottom, nil, nil, centeredTable)
	mainInterface := container.NewBorder(nil, nil, nil, statusArea, contentArea)

	windowContent := container.NewStack(background, mainInterface)
	finalStack := container.NewStack(windowContent, overlay)

	// Let's show the window and run the app
	updateBackgroundColor()
	myWindow.SetContent(finalStack)
	myWindow.Resize(fyne.NewSize(1100, 700))

	SetStatus("Bienvenue")
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
	statusDrawLabel.SetText(fmt.Sprintf("PIOCHE : %d", len(gameState.Remaining)))
	playerRack.Refresh()
}

// ----------------------------------------------------------------------------
// setMenu()
// ----------------------------------------------------------------------------
func setMenu() {
	newItem := fyne.NewMenuItem("Nouvelle Partie", func() { showNewGameDialog(myWindow, onNewGame) })
	saveItem := fyne.NewMenuItem("Sauvegarder", func() { /* Save logic */ })
	quitItem := fyne.NewMenuItem("Quitter", func() { confirmExit() })
	appearanceMenu := fyne.NewMenu("Affichage",
		fyne.NewMenuItem("Thème Sombre", func() {
			SetStatus("Application du thème sombre")
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.DarkTheme()})
			myApp.Preferences().SetString("AppTheme", "dark")
			updateBackgroundColor()
			myWindow.Content().Refresh()
		}),
		fyne.NewMenuItem("Thème Clair", func() {
			SetStatus("Application du thème clair")
			myApp.Settings().SetTheme(&compactTheme{Theme: theme.LightTheme()})
			myApp.Preferences().SetString("AppTheme", "light")
			updateBackgroundColor()
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
	nameEntry.SetText(fyne.CurrentApp().Preferences().StringWithFallback("PlayerName", "Joueur 1"))

	// 2. Sélecteur pour le nombre d'adversaires (de 1 à 3)
	aiSelect := widget.NewSelect([]string{"1", "2", "3"}, nil)
	aiSelect.SetSelected("3") // Valeur par défaut

	// 3. Mise en page du formulaire
	form := widget.NewForm(
		widget.NewFormItem("Votre Nom :", nameEntry),
		widget.NewFormItem("Adversaires IA :", aiSelect),
	)

	// 4. Création du dialogue avec boutons Confirmer/Annuler
	dialog.ShowCustomConfirm(
		"Nouvelle Partie", // Titre
		"Démarrer",        // Bouton de validation
		"Annuler",         // Bouton d'annulation
		form,              // Le contenu du formulaire
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
					nomJoueur = "Joueur 1" // Sécurité si le nom est vide
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
	fmt.Printf("Lancement du jeu pour %s contre %d IA\n", name, ais)

	// 1. Initialiser le deck de tuiles
	// 2. Distribuer les tuiles au joueur (en utilisant son 'name')
	// 3. Créer les mains cachées pour les 'ais' adversaires
	// 4. Mettre à jour l'affichage de la réglette et du statut

	// Exemple : updateStatusBar(fmt.Sprintf("Tour de %s", name))
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
