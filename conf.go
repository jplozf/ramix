package main

import "image/color"

// ----------------------------------------------------------------------------
// EXTERNAL VARIABLES
// ----------------------------------------------------------------------------
var GitVersion = "dev" // GitVersion is the number of git commits and the git hash, injected at build time

// ----------------------------------------------------------------------------
// CONSTANTS
// ----------------------------------------------------------------------------
const (
	MAJOR           = "0" // Major version, incremented for significant changes
	APP_NAME        = "Ramix"
	APP_ID          = "fr.ozf.ramix"
	APP_URL         = "https://github.com/jplozf/ramix"
	APP_COPYRIGHT   = "© 2026 JPL. Tous droits réservés."
	APP_DESCRIPTION = "Un jeu développé avec Go et Fyne."
)

// ----------------------------------------------------------------------------
// VARS
// ----------------------------------------------------------------------------
var (
	ColorBackgroundGlobalDark  = color.NRGBA{R: 5, G: 35, B: 20, A: 255}
	ColorBackgroundGlobalLight = color.NRGBA{R: 200, G: 220, B: 205, A: 255}
	ColorBoardBackgroundDark   = color.NRGBA{R: 10, G: 60, B: 30, A: 255}
	ColorBoardBackgroundLight  = color.NRGBA{R: 140, G: 180, B: 150, A: 255}
	ColorRummyRed              = color.NRGBA{R: 210, G: 90, B: 90, A: 255}
	ColorRummyBlue             = color.NRGBA{R: 90, G: 130, B: 180, A: 255}
	ColorRummyYellow           = color.NRGBA{R: 220, G: 170, B: 0, A: 255}
	ColorRummyGreen            = color.NRGBA{R: 60, G: 140, B: 80, A: 255}
	ColorRummyIvory            = color.NRGBA{R: 190, G: 180, B: 160, A: 255}
	ColorIvoryLine             = color.NRGBA{R: 255, G: 255, B: 240, A: 255}
	ColorTileStroke            = color.NRGBA{R: 190, G: 180, B: 160, A: 255}
	ColorRackStrokeLight       = color.NRGBA{R: 150, G: 110, B: 75, A: 255}
	ColorRackStrokeDark        = color.NRGBA{R: 150, G: 110, B: 75, A: 255}
	ColorRackCellLight         = color.NRGBA{R: 245, G: 222, B: 179, A: 255}
	ColorRackCellDark          = color.NRGBA{R: 70, G: 45, B: 30, A: 255}
	ColorRackCellHoverLight    = color.NRGBA{R: 255, G: 245, B: 230, A: 255}
	ColorRackCellHoverDark     = color.NRGBA{R: 105, G: 70, B: 48, A: 255}
	ColorBoardCellHoverLight   = color.NRGBA{R: 130, G: 210, B: 150, A: 255}
	ColorBoardCellHoverDark    = color.NRGBA{R: 20, G: 100, B: 50, A: 255}
	ColorBoardCellStrokeLight  = color.NRGBA{R: 0, G: 0, B: 0, A: 16}
	ColorBoardCellStrokeDark   = color.NRGBA{R: 220, G: 220, B: 220, A: 16}
)
