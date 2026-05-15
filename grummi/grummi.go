package grummi

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// TYPES
// ----------------------------------------------------------------------------
type Color int

type Tile struct {
	Value int
	Color Color
}

type Player struct {
	ID             int
	Hand           []Tile
	HasPlayedFirst bool
	Name           string
	IsAI           bool
}

type Combination []Tile

type GameState struct {
	Players           []Player
	Remaining         []Tile
	CurrentPlayerID   int
	Table             [][]Tile
	Hand              []Tile
	ConsecutivePasses int
}

// ----------------------------------------------------------------------------
// CONSTS
// ----------------------------------------------------------------------------
const (
	Red Color = iota
	Blue
	Green
	Orange
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorBlue   = "\033[34m"
	ColorOrange = "\033[33m"
)

// MAJOR Version number, injected at build time
const MAJOR = "0"

// ----------------------------------------------------------------------------
// GLOBALS
// ----------------------------------------------------------------------------
// GitVersion is the number of git commits and the git hash, injected at build time
var GitVersion = "dev"
var nTurn = 1

// ----------------------------------------------------------------------------
// main()
// ----------------------------------------------------------------------------
func main() {
	args := os.Args
	if len(args) != 2 {
		fmt.Printf("Usage: %s <nombre_de_joueurs>\n", args[0])
		return
	}

	numPlayers := 0
	_, err := fmt.Sscanf(args[1], "%d", &numPlayers)
	if err != nil || numPlayers < 2 || numPlayers > 4 {
		fmt.Println("Le nombre de joueurs doit être entre 2 et 4.")
		return
	}
	game := InitializeGame(numPlayers)
	game.CurrentPlayerID = game.DetermineFirstPlayer()

	for {
		player := &game.Players[game.CurrentPlayerID]

		if player.IsAI {
			game.IATurn(player)
		} else {
			game.HumanTurn(player)
		}

		// Check if the player won
		if len(player.Hand) == 0 {
			fmt.Printf("\n🏆 FÉLICITATIONS ! %s a vidé sa main et gagne la partie !\n", player.Name)
			game.PrintFinalScores(player.ID)
			game.PrintFinalHands()
			break
		}

		// Check for stalemate (no one can play and draw pile is empty)
		if game.ConsecutivePasses >= len(game.Players) && len(game.Remaining) == 0 {
			game.PrintFinalScores(-1) // -1 indicates no specific winner (stalemate)
			fmt.Println("\n🤝 MATCH NUL ! La pioche est vide et plus personne ne peut jouer.")
			game.PrintFinalHands()
			break
		}

		// Move to the next player
		game.CurrentPlayerID = (game.CurrentPlayerID + 1) % len(game.Players)
	}
}

// ----------------------------------------------------------------------------
// IsValidGroup()
// ----------------------------------------------------------------------------
func IsValidGroup(combo Combination) bool {
	if len(combo) < 3 || len(combo) > 4 {
		return false
	}

	firstValue := -1
	colorsSeen := make(map[Color]bool)

	for _, tile := range combo {
		// Joker management (value 0)
		if tile.Value == 0 {
			continue
		}

		// Check for unique value (e.g., all are 11)
		if firstValue == -1 {
			firstValue = tile.Value
		} else if tile.Value != firstValue {
			return false // Different values in a group!
		}

		// Check for unique colors
		if colorsSeen[tile.Color] {
			return false // Duplicate color detected!
		}
		colorsSeen[tile.Color] = true
	}

	return true
}

// ----------------------------------------------------------------------------
// IsValidRun()
// ----------------------------------------------------------------------------
func IsValidRun(combo Combination) bool {
	if len(combo) < 3 {
		return false
	}

	// 1. Extract real tiles and count jokers
	var realTiles []Tile
	jokerCount := 0
	var color Color
	colorSet := false

	for _, tile := range combo {
		if tile.Value == 0 {
			jokerCount++
		} else {
			if !colorSet {
				color = tile.Color
				colorSet = true
			} else if tile.Color != color {
				return false // All real tiles must be the same color
			}
			realTiles = append(realTiles, tile)
		}
	}

	// If all are Jokers (rare but possible), it's valid
	if len(realTiles) == 0 {
		return true
	}

	// 2. Sort real tiles
	sort.Slice(realTiles, func(i, j int) bool {
		return realTiles[i].Value < realTiles[j].Value
	})

	// 3. Check for duplicates and gaps
	for i := 0; i < len(realTiles)-1; i++ {
		diff := realTiles[i+1].Value - realTiles[i].Value

		if diff == 0 {
			return false // Two identical tiles (e.g., two red 7s)
		}

		// If the gap is > 1, consume jokers to fill it
		// e.g., between 5 and 8, 2 jokers are needed (6 and 7)
		neededJokers := diff - 1
		jokerCount -= neededJokers
	}

	// 4. Check Rummikub limits (values from 1 to 13)
	// Check if the remaining jokers can be placed without exceeding limits
	// (This part is simplified here, mainly checking if jokerCount is not negative)
	if jokerCount < 0 {
		return false
	}

	// Optional: Check that the total run (real + used jokers)
	// does not exceed 13 tiles and remains between 1 and 13.
	return len(combo) <= 13
}

// ----------------------------------------------------------------------------
// IsValidCombination()
// ----------------------------------------------------------------------------
func IsValidCombination(combo Combination) bool {
	return IsValidGroup(combo) || IsValidRun(combo)
}

// ----------------------------------------------------------------------------
// initializeAllTiles()
// ----------------------------------------------------------------------------
func initializeAllTiles() []Tile {
	var tiles []Tile
	// 2 sets of 52 tiles (4 colors x 13 values)
	for i := 0; i < 2; i++ {
		for value := 1; value <= 13; value++ {
			for color := Red; color <= Orange; color++ {
				tiles = append(tiles, Tile{Value: value, Color: color})
			}
		}
	}
	// Add 2 jokers
	tiles = append(tiles, Tile{Value: 0, Color: -1})
	tiles = append(tiles, Tile{Value: 0, Color: -1})
	return tiles
}

// ----------------------------------------------------------------------------
// shuffleTiles()
// ----------------------------------------------------------------------------
func shuffleTiles(tiles []Tile) []Tile {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tiles), func(i, j int) {
		tiles[i], tiles[j] = tiles[j], tiles[i]
	})
	return tiles
}

// ----------------------------------------------------------------------------
// dealTiles()
// ----------------------------------------------------------------------------
func dealTiles(tiles []Tile, numPlayers int) ([]Player, []Tile) {
	players := make([]Player, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = Player{ID: i + 1}
		var name string
		isAI := false

		if i == 0 {
			name = "Humain" // The first player is you
			isAI = false
		} else {
			// The following are AI (AI#1, AI#2...)
			name = fmt.Sprintf("AI#%d", i)
			isAI = true
		}
		players[i].Name = name
		players[i].IsAI = isAI
		players[i].HasPlayedFirst = false
	}

	// Each player gets 14 tiles
	for i := 0; i < numPlayers*14; i++ {
		playerIndex := i / 14
		players[playerIndex].Hand = append(players[playerIndex].Hand, tiles[i])
	}

	// The rest forms the draw pile
	remaining := tiles[numPlayers*14:]
	return players, remaining
}

// ----------------------------------------------------------------------------
// InitializeGame()
// ----------------------------------------------------------------------------
func InitializeGame(numPlayers int) GameState {
	if numPlayers < 2 || numPlayers > 4 {
		panic("Le nombre de joueurs doit être entre 2 et 4.")
	}

	allTiles := initializeAllTiles()
	shuffledTiles := shuffleTiles(allTiles)
	players, remaining := dealTiles(shuffledTiles, numPlayers)

	return GameState{
		Players:         players,
		Remaining:       remaining,
		Table:           [][]Tile{},
		CurrentPlayerID: 0,
	}
}

// ----------------------------------------------------------------------------
// SortTiles()
// ----------------------------------------------------------------------------
func SortTiles(tiles []Tile) {
	sort.Slice(tiles, func(i, j int) bool {
		// 1. Compare colors
		if tiles[i].Color != tiles[j].Color {
			return tiles[i].Color < tiles[j].Color
		}
		// 2. If colors are identical, compare values
		return tiles[i].Value < tiles[j].Value
	})
}

// ----------------------------------------------------------------------------
// removeTiles()
// ----------------------------------------------------------------------------
func removeTiles(hand []Tile, combo Combination) []Tile {
	newHand := make([]Tile, len(hand))
	copy(newHand, hand)

	for _, comboTile := range combo {
		for i, handTile := range newHand {
			// Check value and color (Joker has Value: 0)
			if handTile.Value == comboTile.Value && handTile.Color == comboTile.Color {
				newHand = append(newHand[:i], newHand[i+1:]...)
				break
			}
		}
	}
	return newHand
}

// ----------------------------------------------------------------------------
// FindBestHandLayout()
// ----------------------------------------------------------------------------
func FindBestHandLayout(hand []Tile, hoardJokers bool) ([][]Tile, []Tile) {
	var bestTable [][]Tile
	minRealLeft := 999
	maxJokerLeft := -1

	bestTable = [][]Tile{}
	remainingHand := hand

	// Internal recursive function
	var backtrack func(currentHand []Tile, currentTable [][]Tile)
	backtrack = func(currentHand []Tile, currentTable [][]Tile) {
		realLeft := 0
		jokerLeft := 0
		for _, t := range currentHand {
			if t.Value == 0 {
				jokerLeft++
			} else {
				realLeft++
			}
		}

		// Search for all possible combinations in the current hand
		possibilities := GetAllPossibleCombos(currentHand)

		// If no combination is possible, compare with our best score
		if len(possibilities) == 0 {
			// If hoarding, we prioritize keeping jokers.
			// Otherwise, we just minimize the total count (classic solver).
			isBetter := false
			if hoardJokers {
				isBetter = (realLeft < minRealLeft) || (realLeft == minRealLeft && jokerLeft > maxJokerLeft)
			} else {
				isBetter = (realLeft + jokerLeft) < (minRealLeft + (func() int {
					if maxJokerLeft < 0 {
						return 0
					}
					return maxJokerLeft
				}()))
			}

			if isBetter {
				minRealLeft = realLeft
				maxJokerLeft = jokerLeft
				bestTable = make([][]Tile, len(currentTable))
				copy(bestTable, currentTable)
				remainingHand = currentHand
			}
			return
		}

		// Try each possibility
		for _, combo := range possibilities {
			newHand := removeTiles(currentHand, combo)
			newTable := append(currentTable, combo)
			backtrack(newHand, newTable)
		}
	}

	backtrack(hand, [][]Tile{})
	return bestTable, remainingHand
}

// ----------------------------------------------------------------------------
// DrawTile()
// ----------------------------------------------------------------------------
func (state *GameState) DrawTile() {
	if len(state.Remaining) == 0 {
		fmt.Println("La pioche est vide !")
		return
	}

	// Take the first tile from the draw pile
	tile := state.Remaining[0]
	state.Remaining = state.Remaining[1:]

	// Add it to the current player's hand
	currentPlayer := &state.Players[state.CurrentPlayerID]
	currentPlayer.Hand = append(currentPlayer.Hand, tile)

	fmt.Printf("Joueur %d pioche une tuile.\n", currentPlayer.ID)
}

// ----------------------------------------------------------------------------
// PrintTable()
// ----------------------------------------------------------------------------
func PrintTable(table [][]Tile) {
	fmt.Println(strings.Repeat("═", 80))
	if len(table) == 0 {
		fmt.Println("    [ Vide ]")
	} else {
		for i, combo := range table {
			fmt.Printf(" [%2d] : ", i+1)
			for _, tile := range combo {
				fmt.Print(FormatTile(tile))
			}
			fmt.Println()
		}
	}
	fmt.Println(strings.Repeat("═", 80))
}

// ----------------------------------------------------------------------------
// PrintFinalHands()
// ----------------------------------------------------------------------------
func (state *GameState) PrintFinalHands() {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("📊 ÉTAT FINAL DES MAINS :")
	for _, p := range state.Players {
		SortTiles(p.Hand)
		fmt.Printf(" 👤 %-10s : ", p.Name)
		for _, tile := range p.Hand {
			fmt.Print(FormatTile(tile))
		}
		fmt.Println()
	}
	fmt.Println(strings.Repeat("═", 80))
}

// ----------------------------------------------------------------------------
// FormatTile()
// ----------------------------------------------------------------------------
func FormatTile(tile Tile) string {
	if tile.Value == 0 {
		return "  😁 "
	}
	colorIcon := ""
	colorCode := ""
	switch tile.Color {
	case Red:
		colorIcon = "🔴"
		colorCode = ColorRed
	case Blue:
		colorIcon = "🔵"
		colorCode = ColorBlue
	case Green:
		colorIcon = "🟢"
		colorCode = ColorGreen
	case Orange:
		colorIcon = "🟠"
		colorCode = ColorOrange
	}
	return fmt.Sprintf("%s%2d%s%s ", colorCode, tile.Value, colorIcon, ColorReset)
}

// ----------------------------------------------------------------------------
// TotalValue()
// ----------------------------------------------------------------------------
func (combo Combination) TotalValue() int {
	sum := 0
	// For a simple version: sum the real values.
	// Note: A Joker (0) should normally take the value
	// of the tile it replaces in the run or group.
	for _, tile := range combo {
		sum += tile.Value
	}
	return sum
}

// ----------------------------------------------------------------------------
// CalculateHandPoints()
// ----------------------------------------------------------------------------
func CalculateHandPoints(hand []Tile) int {
	total := 0
	for _, tile := range hand {
		if tile.Value == 0 { // Joker
			total += 30
		} else {
			total += tile.Value
		}
	}
	return total
}

// ----------------------------------------------------------------------------
// HumanTurn()
// ----------------------------------------------------------------------------
func (state *GameState) HumanTurn(p *Player) {
	// 1. We backup the current state of the table and the player's hand to allow for cancellation if needed
	backupTable := cloneTable(state.Table)
	backupHand := cloneHand(p.Hand)

	// 2. We create an empty pool to hold the tiles the player is currently manipulating
	pool := []Tile{} // Let's start with an empty pool, and the player will add tiles from their hand to it as they prepare their move

	for {
		// We clear the console at the beginning of each loop to keep the interface clean
		fmt.Print("\033[H\033[2J")

		currentValueOnTable := calculateTotalValue(state.Table)
		backupValueOnTable := calculateTotalValue(backupTable)
		pointsFromHand := currentValueOnTable - backupValueOnTable

		// We display the current state of the game and the action menu
		state.PrintUserMenu(p, pool, pointsFromHand)

		// 2. Read user action
		var action string
		fmt.Scanln(&action)
		action = strings.ToLower(action)

		// 3. Process action
		switch action {
		case "h":
			// We move a tile from the hand to the pool (to be placed on the table)
			idx := getIndex()
			if idx >= 0 && idx < len(p.Hand) {
				pool = append(pool, p.Hand[idx])
				p.Hand = append(p.Hand[:idx], p.Hand[idx+1:]...)
			} else {
				fmt.Println("❌ Index invalide dans votre main.")
				time.Sleep(1 * time.Second)
			}

		case "n":
			// We use tiles from the pool to create a new combination on the table
			indices := getMultipleIndices()
			if len(indices) == 0 {
				continue
			}
			var newCombo Combination
			validIndices := true
			for _, i := range indices {
				if i < 0 || i >= len(pool) {
					fmt.Printf("❌ Index invalide dans la réserve : %d\n", i)
					validIndices = false
					break
				}
				newCombo = append(newCombo, pool[i])
			}

			if validIndices && IsValidCombination(newCombo) {
				SortTiles(newCombo)
				state.Table = append(state.Table, newCombo)
				pool = removeTilesFromPool(pool, indices)
				fmt.Println("✅ Nouvelle combinaison posée !")
			} else if validIndices {
				fmt.Println("❌ Cette combinaison n'est pas valide.")
			}
			time.Sleep(1 * time.Second)

		case "p":
			// Cancellation: restore the initial state
			state.Table = backupTable
			p.Hand = backupHand

			// Then perform the mandatory draw
			state.DrawTile()
			state.ConsecutivePasses++
			fmt.Println("Tour annulé, vous avez pioché une tuile.")
			nTurn++
			return

		case "m":
			idx := getIndex()
			if idx < 0 || idx >= len(pool) {
				fmt.Println("❌ Index invalide.")
				continue
			}

			tuileAChecker := pool[idx]

			// Verification: does the tile come from the original hand?
			isGenuine := false
			for _, t := range backupHand {
				// Compare Value and Color (and possibly a unique ID if you have one)
				if t.Value == tuileAChecker.Value && t.Color == tuileAChecker.Color {
					isGenuine = true
					break
				}
			}

			if isGenuine {
				p.Hand = append(p.Hand, tuileAChecker)
				pool = removeTilesFromPool(pool, []int{idx})
				fmt.Println("✅ Tuile remise en main.")
			} else {
				fmt.Println("❌ Interdit ! Cette tuile provient de la table, vous ne pouvez pas la prendre dans votre main.")
			}

		case "t":
			if len(state.Table) == 0 {
				fmt.Println("❌ La table est vide.")
				continue
			}
			fmt.Printf("Quelle combinaison voulez-vous ramasser (1 à %d) ? ", len(state.Table))
			var idx int
			fmt.Scanln(&idx)
			idx -= 1 // Ajustement pour l'index 0

			if idx >= 0 && idx < len(state.Table) {
				// 1. Get the tiles
				comboARamasser := state.Table[idx]

				// 2. Add them to the pool
				pool = append(pool, comboARamasser...)

				// 3. Remove them from the table
				state.Table = append(state.Table[:idx], state.Table[idx+1:]...)

				fmt.Println("✅ Combinaison envoyée dans la réserve.")
			} else {
				fmt.Println("❌ Index invalide.")
			}

		case "s":
			// 1. Check for orphan tiles
			if len(pool) > 0 {
				fmt.Printf("❌ Action impossible : il reste %d tuile(s) dans la réserve !\n", len(pool))
				fmt.Println("Vous devez toutes les replacer sur la table.")
				continue // Return to the beginning of the loop so the player continues
			}

			// 1bis. Check: the player must have placed at least one tile from their hand
			if len(p.Hand) == len(backupHand) {
				fmt.Println("❌ Vous n'avez posé aucune tuile de votre main. Tour annulé, vous piochez une tuile.")
				state.Table = backupTable
				p.Hand = backupHand
				state.DrawTile()
				state.ConsecutivePasses++
				nTurn++
				return
			}

			// 2. Score calculation for the first play (if necessary)
			if !p.HasPlayedFirst {
				if pointsFromHand < 30 {
					fmt.Printf("❌ Premier coup invalide : vous avez posé %d points (minimum 30).\n", pointsFromHand)
					fmt.Println("Souhaitez-vous [c]ontinuer ou [a]nnuler et piocher ?")

					var choice string
					fmt.Scanln(&choice)
					if choice == "a" {
						// Restore everything and draw (Action "p")
						state.Table = backupTable
						p.Hand = backupHand
						state.DrawTile()
						state.ConsecutivePasses++
						nTurn++
						return
					}
					continue // Return to the turn so they add tiles
				}

				// If we reach this point, the score is >= 30
				p.HasPlayedFirst = true
				fmt.Println("🎉 Félicitations ! Vous avez validé votre ouverture.")
			}

			// 3. Final validation
			state.ConsecutivePasses = 0
			fmt.Println("✅ Tour validé. Fin du tour.")
			nTurn++
			return // Exit HumanTurn, state.Table already contains the new modifications

		case "q":
			fmt.Print("Êtes-vous sûr de vouloir quitter ? (o/n) : ")
			var confirm string
			fmt.Scanln(&confirm)
			if strings.ToLower(confirm) == "o" {
				fmt.Println("Fermeture du jeu...")
				os.Exit(0)
			}
		}
	}
}

// ----------------------------------------------------------------------------
// DetermineFirstPlayer()
// ----------------------------------------------------------------------------
func (state *GameState) DetermineFirstPlayer() int {
	// Initialize random seed (to avoid always having the same result)
	rand.Seed(time.Now().UnixNano())

	// Draw a number between 0 and number of players - 1
	firstPlayerIndex := rand.Intn(len(state.Players))

	fmt.Printf("\n🎲 Tirage au sort... C'est %s qui commence !\n", state.Players[firstPlayerIndex].Name)
	// Leave a short pause for suspense
	time.Sleep(2 * time.Second)

	return firstPlayerIndex
}

// ----------------------------------------------------------------------------
// calculateTotalValue()
// ----------------------------------------------------------------------------
func calculateTotalValue(table [][]Tile) int {
	total := 0
	for _, combo := range table {
		isRun := IsValidRun(combo)
		total += GetComboValueWithJoker(combo, isRun)
	}
	return total
}

// ----------------------------------------------------------------------------
// PrintHandWithIndices()
// ----------------------------------------------------------------------------
func (p Player) PrintHandWithIndices() {
	SortTiles(p.Hand)

	// fmt.Printf("Votre main :\n")
	var lastColor Color = -2 // Dummy initial value to detect the first tile

	for i, tile := range p.Hand {
		// If color changes (and it's not a Joker), add a small space
		if i > 0 && tile.Color != lastColor && tile.Value != 0 {
			fmt.Print("\n")
		}
		fmt.Printf("[%2d]:%s ", i, FormatTile(tile))
		lastColor = tile.Color
	}
	fmt.Println()
}

// ----------------------------------------------------------------------------
// parseIndices()
// ----------------------------------------------------------------------------
func parseIndices(input string) []int {
	var indices []int
	parts := strings.Split(input, ",")
	for _, p := range parts {
		var idx int
		_, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &idx)
		if err == nil {
			indices = append(indices, idx)
		}
	}
	return indices
}

// ----------------------------------------------------------------------------
// IATurn()
// ----------------------------------------------------------------------------
func (state *GameState) IATurn(currentPlayer *Player) {
	fmt.Printf("\n🤖 %s réfléchit...\n", currentPlayer.Name)
	initialHandSize := len(currentPlayer.Hand)

	// Save current state for potential rollback if the move is invalid or doesn't improve the hand
	backupTable := cloneTable(state.Table)
	backupHand := cloneHand(currentPlayer.Hand)
	backupHasPlayedFirst := currentPlayer.HasPlayedFirst

	// 1. If already opened, try to liberate jokers and append single tiles
	if currentPlayer.HasPlayedFirst {
		changed := true
		for changed {
			changed = false
			if state.LiberateJokers(currentPlayer) {
				changed = true
			}
			if state.TrySplitLongCombos() {
				changed = true
			}
			if state.TryAppendToTable(currentPlayer, false) {
				changed = true
			}
			if state.TrySplitAndInsert(currentPlayer) {
				changed = true
			}
		}
	}

	// 2. The AI analyzes its hand to find new complete combinations
	// Pass 1: Try to play while hoarding Jokers
	bestLayout, remainingHand := FindBestHandLayout(currentPlayer.Hand, true)

	calculatePoints := func(layout [][]Tile) int {
		pts := 0
		for _, combo := range layout {
			pts += GetComboValueWithJoker(combo, IsValidRun(combo))
		}
		return pts
	}
	totalProposed := calculatePoints(bestLayout)

	canPlayNew := false
	if !currentPlayer.HasPlayedFirst {
		if totalProposed >= 30 {
			canPlayNew = true
			currentPlayer.HasPlayedFirst = true
			fmt.Printf("⭐ %s : Première pose validée avec %d points !\n", currentPlayer.Name, totalProposed)
		} else {
			// Pass 2: If we can't open by hoarding, try using Jokers aggressively
			aggLayout, aggRemaining := FindBestHandLayout(currentPlayer.Hand, false)
			if calculatePoints(aggLayout) >= 30 {
				bestLayout = aggLayout
				remainingHand = aggRemaining
				canPlayNew = true
				currentPlayer.HasPlayedFirst = true
				fmt.Printf("⭐ %s : Ouverture validée avec Joker (%d points) !\n", currentPlayer.Name, calculatePoints(aggLayout))
			}
		}
	} else {
		if len(bestLayout) > 0 {
			canPlayNew = true
		}
	}

	if canPlayNew {
		state.Table = append(state.Table, bestLayout...)
		currentPlayer.Hand = remainingHand
	}

	// Final cleanup: try appending any remaining tiles, now including Jokers
	if currentPlayer.HasPlayedFirst {
		state.TryAppendToTable(currentPlayer, true)
	}

	// 3. Final check: did the hand size decrease?
	if len(currentPlayer.Hand) < initialHandSize {
		state.ConsecutivePasses = 0
		time.Sleep(1 * time.Second) // Pause so the player can see the AI's moves
	} else {
		// Rollback if no progress was made (ensures AI doesn't keep "stolen" jokers without playing)
		state.Table = backupTable
		currentPlayer.Hand = backupHand
		currentPlayer.HasPlayedFirst = backupHasPlayedFirst
		state.DrawTile()
		state.ConsecutivePasses++
	}
}

// ****************************************************************************
// TryAppendToTable()
// ****************************************************************************
// TryAppendToTable attempts to add individual tiles from the player's hand to existing table combinations.
func (state *GameState) TryAppendToTable(p *Player, allowJokers bool) bool {
	playedAtLeastOne := false
	modified := true

	for modified {
		modified = false
		for i := 0; i < len(p.Hand); i++ {
			tile := p.Hand[i]

			// Skip Jokers for simple greedy appending.
			// This prevents infinite loops where the AI frees a Joker and immediately appends it back.
			if tile.Value == 0 && !allowJokers {
				continue
			}

			for j := 0; j < len(state.Table); j++ {
				// Create a temporary combination to test the addition
				newCombo := append(Combination(nil), state.Table[j]...)
				newCombo = append(newCombo, tile)
				SortTiles(newCombo)

				if IsValidCombination(newCombo) {
					state.Table[j] = newCombo
					p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
					playedAtLeastOne = true
					modified = true
					fmt.Printf("🤖 %s ajoute %s à une combinaison sur la table.\n", p.Name, FormatTile(tile))
					break
				}
			}
			if modified {
				break
			}
		}
	}
	return playedAtLeastOne
}

// ****************************************************************************
// TrySplitLongCombos()
// ****************************************************************************
// TrySplitLongCombos looks for runs of 6+ tiles and splits them into two combinations.
// This creates more "ends" on the table for future tiles to be attached to.
func (state *GameState) TrySplitLongCombos() bool {
	for i, combo := range state.Table {
		// Groups are max 4, so only runs can be 6+.
		if len(combo) >= 6 {
			// Try splitting at points that leave at least 3 tiles on each side.
			for k := 3; k <= len(combo)-3; k++ {
				part1 := append(Combination(nil), combo[:k]...)
				part2 := append(Combination(nil), combo[k:]...)

				if IsValidCombination(part1) && IsValidCombination(part2) {
					state.Table = append(state.Table[:i], state.Table[i+1:]...)
					state.Table = append(state.Table, part1, part2)
					fmt.Printf("🤖 Scission d'une longue combinaison (%d tuiles) en deux.\n", len(combo))
					return true
				}
			}
		}
	}
	return false
}

// ****************************************************************************
// TrySplitAndInsert()
// ****************************************************************************
// TrySplitAndInsert attempts to split a combo of 5+ tiles by inserting a tile from the hand
// to make both resulting parts valid (length 3+).
func (state *GameState) TrySplitAndInsert(p *Player) bool {
	for i, combo := range state.Table {
		if len(combo) < 5 {
			continue
		}

		for hIdx, handTile := range p.Hand {
			// Skip placeholder/empty tiles
			if handTile.Color == -1 && handTile.Value == 0 {
				continue
			}

			// Try every possible split point k
			for k := 1; k < len(combo); k++ {
				part1 := append(Combination(nil), combo[:k]...)
				part2 := append(Combination(nil), combo[k:]...)

				// Case A: Hand tile joins part 1, and part 2 is valid as-is
				if len(part2) >= 3 {
					test1 := append(Combination(nil), part1...)
					test1 = append(test1, handTile)
					SortTiles(test1)
					if IsValidCombination(test1) && IsValidCombination(part2) {
						state.Table = append(state.Table[:i], state.Table[i+1:]...)
						state.Table = append(state.Table, test1, part2)
						// Remove tile from hand
						p.Hand = append(p.Hand[:hIdx], p.Hand[hIdx+1:]...)
						fmt.Printf("🤖 %s scinde une combinaison pour insérer %s.\n", p.Name, FormatTile(handTile))
						return true
					}
				}

				// Case B: Hand tile joins part 2, and part 1 is valid as-is
				if len(part1) >= 3 {
					test2 := append(Combination(nil), part2...)
					test2 = append(test2, handTile)
					SortTiles(test2)
					if IsValidCombination(test2) && IsValidCombination(part1) {
						state.Table = append(state.Table[:i], state.Table[i+1:]...)
						state.Table = append(state.Table, part1, test2)
						// Remove tile from hand
						p.Hand = append(p.Hand[:hIdx], p.Hand[hIdx+1:]...)
						fmt.Printf("🤖 %s scinde une combinaison pour insérer %s.\n", p.Name, FormatTile(handTile))
						return true
					}
				}
			}
		}
	}
	return false
}

// ****************************************************************************
// LiberateJokers()
// ****************************************************************************
// LiberateJokers attempts to recover Jokers from the table by either replacing them
// with a valid tile from the hand or removing them from a combination of 4+ tiles.
func (state *GameState) LiberateJokers(p *Player) bool {
	if !p.HasPlayedFirst {
		return false
	}

	for i, combo := range state.Table {
		jokerIdx := -1
		for k, t := range combo {
			if t.Value == 0 {
				jokerIdx = k
				break
			}
		}

		if jokerIdx == -1 {
			continue
		}

		// 1. Try replacing with a tile from hand
		for hIdx, handTile := range p.Hand {
			if handTile.Value == 0 {
				continue
			}

			newCombo := make(Combination, len(combo))
			copy(newCombo, combo)
			newCombo[jokerIdx] = handTile

			if IsValidCombination(newCombo) {
				SortTiles(newCombo)
				state.Table[i] = newCombo
				p.Hand[hIdx] = Tile{Value: 0, Color: -1}
				fmt.Printf("🤖 %s remplace un Joker sur la table par %s.\n", p.Name, FormatTile(handTile))
				return true
			}
		}

		// 2. Try simple removal if the combination remains valid (length > 3)
		if len(combo) > 3 {
			newCombo := append(Combination(nil), combo[:jokerIdx]...)
			newCombo = append(newCombo, combo[jokerIdx+1:]...)

			if IsValidCombination(newCombo) {
				state.Table[i] = newCombo
				p.Hand = append(p.Hand, Tile{Value: 0, Color: -1})
				fmt.Printf("🤖 %s libère un Joker (combinaison de %d tuiles).\n", p.Name, len(combo))
				return true
			}
		}
	}
	return false
}

// ----------------------------------------------------------------------------
// GetComboValueWithJoker()
// ----------------------------------------------------------------------------
func GetComboValueWithJoker(combo Combination, isRun bool) int {
	if len(combo) == 0 {
		return 0
	}

	// 1. Extract real values and count Jokers
	var realValues []int
	jokers := 0
	for _, t := range combo {
		if t.Value == 0 {
			jokers++
		} else {
			realValues = append(realValues, t.Value)
		}
	}

	if len(realValues) == 0 {
		return 0
	}
	sort.Ints(realValues)

	if !isRun {
		// Group: All tiles are worth the same as the real tile
		return realValues[0] * len(combo)
	}

	// Run: Calculate the sum considering the gaps filled by Jokers
	total := 0
	for _, v := range realValues {
		total += v
	}

	// Fill the gaps between real tiles
	for i := 0; i < len(realValues)-1; i++ {
		for v := realValues[i] + 1; v < realValues[i+1]; v++ {
			total += v
			jokers--
		}
	}

	// Place remaining jokers at the ends (priority to the high end, max 13)
	low := realValues[0]
	high := realValues[len(realValues)-1]
	for jokers > 0 {
		if high < 13 {
			high++
			total += high
		} else {
			low--
			total += low
		}
		jokers--
	}
	return total
}

// ----------------------------------------------------------------------------
// FindAllGroupsWithJokers()
// ----------------------------------------------------------------------------
func FindAllGroupsWithJokers(hand []Tile) []Combination {
	var groups []Combination
	byValue := make(map[int]map[Color]Tile)
	jokersCount := 0

	// 1. Sort by value and filter colors at the same time
	for _, t := range hand {
		if t.Value == 0 {
			jokersCount++
		} else {
			if byValue[t.Value] == nil {
				byValue[t.Value] = make(map[Color]Tile)
			}
			byValue[t.Value][t.Color] = t
		}
	}

	// 2. Analyze each value group
	for _, colorMap := range byValue {
		// Transform the color map into a slice to manipulate tiles
		var tilesInValue []Tile
		for _, t := range colorMap {
			tilesInValue = append(tilesInValue, t)
		}

		count := len(tilesInValue)

		// Case A: 3 or 4 tiles of different colors
		if count >= 3 {
			groups = append(groups, Combination(tilesInValue))
		}

		// Case B: 2 or 3 real tiles + 1 Joker (to make a group of 3 or 4)
		if (count == 2 || count == 3) && jokersCount >= 1 {
			combo := append(Combination{}, tilesInValue...)
			combo = append(combo, Tile{Value: 0, Color: -1}) // Adding the Joker
			groups = append(groups, combo)
		}

		// Note: We could also handle 1 tile + 2 Jokers,
		// but it is often less strategic for the AI at the start.
	}
	return groups
}

// ----------------------------------------------------------------------------
// FindAllRunsWithJokers()
// ----------------------------------------------------------------------------
func FindAllRunsWithJokers(hand []Tile) []Combination {
	var allRuns []Combination

	// 1. Separate by color and count jokers
	byColor := make(map[Color][]Tile)
	jokersInHand := 0
	for _, t := range hand {
		if t.Value == 0 {
			jokersInHand++
		} else {
			byColor[t.Color] = append(byColor[t.Color], t)
		}
	}

	// 2. For each color, look for runs
	for _, tiles := range byColor {
		// Sort and remove duplicates to facilitate run search
		sortedTiles := uniqueSortedTiles(tiles)

		// Test all possible starting windows (value from 1 to 13)
		for startVal := 1; startVal <= 11; startVal++ {
			// And all possible lengths (3 to 13)
			for length := 3; length <= 13; length++ {
				if startVal+length-1 > 13 {
					break
				}

				currentCombo := Combination{}
				jokersNeeded := 0

				// Try to build the run [startVal ... startVal+length-1]
				for v := startVal; v < startVal+length; v++ {
					found := false
					for _, t := range sortedTiles {
						if t.Value == v {
							currentCombo = append(currentCombo, t)
							found = true
							break
						}
					}
					if !found {
						// The tile is missing, a joker will be needed
						currentCombo = append(currentCombo, Tile{Value: 0, Color: -1})
						jokersNeeded++
					}
				}

				// If enough jokers are available and it's not a "jokers-only" run
				if jokersNeeded <= jokersInHand && jokersNeeded < length {
					// Make a clean copy to add to the results
					runCopy := make(Combination, len(currentCombo))
					copy(runCopy, currentCombo)
					allRuns = append(allRuns, runCopy)
				}
			}
		}
	}
	return allRuns
}

// ----------------------------------------------------------------------------
// uniqueSortedTiles()
// ----------------------------------------------------------------------------
// Utility function to sort and ignore duplicates (e.g., two red 7s)
func uniqueSortedTiles(tiles []Tile) []Tile {
	if len(tiles) == 0 {
		return tiles
	}
	sort.Slice(tiles, func(i, j int) bool { return tiles[i].Value < tiles[j].Value })
	unique := []Tile{tiles[0]}
	for i := 1; i < len(tiles); i++ {
		if tiles[i].Value != tiles[i-1].Value {
			unique = append(unique, tiles[i])
		}
	}
	return unique
}

// ----------------------------------------------------------------------------
// GetAllPossibleCombos()
// ----------------------------------------------------------------------------
func GetAllPossibleCombos(hand []Tile) []Combination {
	var all []Combination
	all = append(all, FindAllGroupsWithJokers(hand)...)
	all = append(all, FindAllRunsWithJokers(hand)...)
	return all
}

// ----------------------------------------------------------------------------
// cloneHand()
// ----------------------------------------------------------------------------
func cloneHand(hand []Tile) []Tile {
	newHand := make([]Tile, len(hand))
	copy(newHand, hand)
	return newHand
}

// ----------------------------------------------------------------------------
// cloneTable()
// ----------------------------------------------------------------------------
func cloneTable(table [][]Tile) [][]Tile {
	newTable := make([][]Tile, len(table))
	for i, combo := range table {
		newTable[i] = make([]Tile, len(combo))
		copy(newTable[i], combo)
	}
	return newTable
}

// ----------------------------------------------------------------------------
// PrintTilePool()
// ----------------------------------------------------------------------------
func PrintTilePool(pool []Tile) {
	if len(pool) == 0 {
		fmt.Println("    [ Vide ]")
		return
	}
	fmt.Print("RÉSERVE : ")
	for i, t := range pool {
		fmt.Printf("[%2d]:%s ", i, FormatTile(t))
	}
	fmt.Println()
}

// ----------------------------------------------------------------------------
// getIndex()
// ----------------------------------------------------------------------------
func getIndex() int {
	var idx int
	fmt.Print("Entrez l'index : ")
	fmt.Scanln(&idx)
	return idx
}

// ----------------------------------------------------------------------------
// getMultipleIndices()
// ----------------------------------------------------------------------------
func getMultipleIndices() []int {
	fmt.Print("Entrez les index séparés par des virgules (ex: 0,1,2) : ")
	var input string
	fmt.Scanln(&input)

	// We use a helper function to parse the indices from the input string
	return parseIndices(input)
}

// ----------------------------------------------------------------------------
// removeTilesFromPool()
// ----------------------------------------------------------------------------
func removeTilesFromPool(pool []Tile, indices []int) []Tile {
	// 1. Sort indices from largest to smallest
	sort.Sort(sort.Reverse(sort.IntSlice(indices)))

	for _, idx := range indices {
		if idx >= 0 && idx < len(pool) {
			// Standard Go technique for removing an element from a slice
			pool = append(pool[:idx], pool[idx+1:]...)
		}
	}
	return pool
}

// ----------------------------------------------------------------------------
// PrintUserMenu()
// ----------------------------------------------------------------------------
func (state *GameState) PrintUserMenu(p *Player, pool []Tile, points int) {
	// Calculate the number of remaining tiles
	remainingTiles := len(state.Remaining)

	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("                                                        _ ")
	fmt.Println("                    __ _ _ __ _   _ _ __ ___  _ __ ___ (_)")
	fmt.Println("                   / _` | '__| | | | '_ ` _ \\| '_ ` _ \\| |")
	fmt.Println("                  | (_| | |  | |_| | | | | | | | | | | | |")
	fmt.Println("                   \\__, |_|   \\__,_|_| |_| |_|_| |_| |_|_|")
	fmt.Printf("                   |___/        v%s © JPL 2026\n", getFullVersion())
	fmt.Println("\n" + strings.Repeat("═", 80))
	// Add the draw pile to the banner
	fmt.Printf(" 👤 JOUEUR : %-8s | 🃏 PIOCHE : %-3d | 🏆 OUVERTURE : %s | 🃏 TOUR : %d \n",
		p.Name, remainingTiles, formatStatus(p.HasPlayedFirst), nTurn)
	fmt.Print(" 👥 TUILES : ")
	for _, other := range state.Players {
		if other.HasPlayedFirst {
			fmt.Printf("%s[✔ %-7s: %2d]%s  ", ColorGreen, other.Name, len(other.Hand), ColorReset)
		} else {
			fmt.Printf("%s[✖ %-7s: %2d]%s  ", ColorRed, other.Name, len(other.Hand), ColorReset)
		}
	}
	fmt.Println("\n" + strings.Repeat("─", 80))

	// 1. Table State (What is validated)
	fmt.Println(" 🧩 TABLE ACTUELLE :")
	if len(state.Table) == 0 {
		fmt.Println("    [ Vide ]")
	} else {
		PrintTable(state.Table)
	}

	// 2. The Reserve (Bulk to process)
	fmt.Println("\n 📥 RÉSERVE (Tuiles à replacer) :")
	PrintTilePool(pool)

	// 3. The Player's Hand
	fmt.Println("\n 🖐️  VOTRE MAIN :")
	p.PrintHandWithIndices()

	// 4. Action menu
	fmt.Println("\n" + strings.Repeat("─", 80))
	if !p.HasPlayedFirst {
		fmt.Printf(" ✨ Points de ce tour : %d / 30\n", points)
	} else {
		fmt.Printf(" ✨ Points de ce tour : %d\n", points)
	}
	fmt.Println("  [N] Nouveau combo        (Réserve -> Table)")
	fmt.Println("  [H] Jouer de la main     (Main    -> Réserve)")
	fmt.Println("  [T] Ramasser de la table (Table   -> Réserve)")
	fmt.Println("  [M] Reprendre en main    (Réserve -> Main)")
	fmt.Println("  [S] Valider le tour")
	fmt.Println("  [P] Piocher et/ou annuler le tour")
	fmt.Println("  [Q] Quitter le jeu")
	fmt.Println(strings.Repeat("═", 80))
	fmt.Print("👉 Votre choix : ")
}

// ----------------------------------------------------------------------------
// formatStatus()
// ----------------------------------------------------------------------------
func formatStatus(b bool) string {
	if b {
		return "✅ OUI"
	}
	return "❌ NON"
}

// ----------------------------------------------------------------------------
// PrintFinalScores()
// ----------------------------------------------------------------------------
func (state *GameState) PrintFinalScores(winnerID int) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("📊 RÉSULTATS FINAUX :")

	playerPoints := make(map[int]int)
	totalOpponentPoints := 0

	// Calculate points for each player's hand
	for _, p := range state.Players {
		handPoints := CalculateHandPoints(p.Hand)
		if p.ID == winnerID {
			// Winner's score is calculated later by summing opponents' points
			playerPoints[p.ID] = 0
		} else {
			playerPoints[p.ID] = -handPoints
			totalOpponentPoints += handPoints
		}
	}

	// Assign winner's score if there is a winner
	if winnerID != -1 {
		playerPoints[winnerID] = totalOpponentPoints
	}

	for _, p := range state.Players {
		fmt.Printf(" 👤 %-10s : %d points\n", p.Name, playerPoints[p.ID])
	}
	fmt.Println(strings.Repeat("═", 80))
}

// ****************************************************************************
// getFullVersion()
// ****************************************************************************
func getFullVersion() string {
	return fmt.Sprintf("%s.%s", MAJOR, GitVersion)
}
