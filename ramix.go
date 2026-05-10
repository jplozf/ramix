package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Rummikub Pro - Go")

	// --- 1. CONFIGURATION DU MENU ---
	newItem := fyne.NewMenuItem("Nouvelle Partie", func() { /* Logique reset */ })
	saveItem := fyne.NewMenuItem("Sauvegarder", func() { /* Logique save */ })
	settingsItem := fyne.NewMenuItem("Paramètres", func() { /* Logique config */ })

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Fichier", newItem, saveItem),
		fyne.NewMenu("Édition", settingsItem),
		fyne.NewMenu("Aide", fyne.NewMenuItem("À propos", func() {})),
	)
	myWindow.SetMainMenu(mainMenu)

	// --- 2. COMPOSANTS CENTRAUX ---
	gameTable := container.New(layout.NewGridLayoutWithColumns(25))
	for i := 0; i < 10*25; i++ {
		gameTable.Add(widget.NewButton("", nil))
	}

	stats := container.NewVBox(
		widget.NewLabelWithStyle("Statistiques", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Tour : Joueur 1"),
		widget.NewLabel("Tuiles IA : 14"),
	)

	playerRack := container.NewHBox(
		widget.NewButtonWithIcon("T1", theme.InfoIcon(), nil), // Exemple de tuile
	)

	// --- 3. BARRE DE STATUT ---
	statusLabel := widget.NewLabel("Prêt - En attente du joueur...")
	statusBar := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		statusLabel,
	)

	// --- 4. ASSEMBLAGE FINAL ---
	// On empile la réglette et la barre de statut en bas (VBox)
	bottomArea := container.NewVBox(
		widget.NewSeparator(),
		container.NewHScroll(playerRack),
		statusBar,
	)

	content := container.NewBorder(
		nil,        // Top
		bottomArea, // Bottom (Réglette + Status)
		nil,        // Left
		stats,      // Right
		gameTable,  // Center
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1100, 700))
	myWindow.ShowAndRun()
}

func createColorRow(colorName string) *fyne.Container {
	label := widget.NewLabelWithStyle(colorName, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	label.Resize(fyne.NewSize(50, 30)) // Largeur fixe pour aligner les débuts de lignes

	rowContent := container.NewHBox()
	// Exemple de tuiles ajoutées à cette ligne
	rowContent.Add(widget.NewButton("1", nil))
	rowContent.Add(widget.NewButton("2", nil))

	// On rend chaque ligne scrollable individuellement si elle dépasse la largeur
	return container.NewHBox(label, container.NewHScroll(rowContent))
}
