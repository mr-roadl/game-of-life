package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
	}
	defer screen.Fini()

	if err := screen.Init(); err != nil {
		log.Fatalf("Error initializing screen: %v", err)
	}

	screen.EnableMouse()

	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.Color108)
	screen.SetStyle(style)

	width, height := screen.Size()
	terminalGrid := createGrid(width, height)

	gamePaused := true
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	eventChannel := make(chan tcell.Event)
	go func() {
		for {
			eventChannel <- screen.PollEvent()
		}
	}()

	for {
		if gamePaused {
			drawText(screen, 2, 1, width, width, style, "Paused!")
		} else {
			drawText(screen, 2, 1, width, width, style, "Runnin'")
		}
		
		screen.Show()

		select {
		case event := <-eventChannel:
			switch ev := event.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape {
					return
				}
				if ev.Key() == tcell.KeyEnter {
					gamePaused = !gamePaused
				}
			case *tcell.EventResize:
				screen.Sync()
				width, height = screen.Size()
				newGrid := createGrid(width, height)
				for y := 0; y < min(len(terminalGrid), len(newGrid)); y++ {
					for x := 0; x < min(len(terminalGrid[0]), len(newGrid[0])); x++ {
						newGrid[y][x] = terminalGrid[y][x]
					}
				}
				terminalGrid = newGrid
			case *tcell.EventMouse:
				if !gamePaused {
					break
				}
				if ev.Buttons() == tcell.Button1 {
					x, y := ev.Position()
					if y < len(terminalGrid) && x < len(terminalGrid[0]) {
						terminalGrid[y][x] = 'c'
						screen.SetContent(x, y, 'o', nil, style)
					}
				}
			}
		case <-ticker.C:
			if !gamePaused {
				terminalGrid = updateGrid(terminalGrid)
				drawGrid(screen, terminalGrid, style)
			}
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func createGrid(width, height int) [][]rune {
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	return grid
}

func drawText(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, text string) {
	row := y1
	col := x1
	for _, r := range []rune(text) {
		s.SetContent(col, row, r, nil, style)
		col++
		if col >= x2 {
			row++
			col = x1
		}
		if row > y2 {
			break
		}
	}
}

func drawGrid(s tcell.Screen, grid [][]rune, style tcell.Style) {
	for y := 0; y < len(grid); y++ {
		for x := 0; x < len(grid[0]); x++ {
			if grid[y][x] == 'c' {
				s.SetContent(x, y, '0', nil, style)
			} else {
				s.SetContent(x, y, ' ', nil, style)
			}
		}
	}
}

func updateGrid(grid [][]rune) [][]rune {
	height := len(grid)
	width := len(grid[0])
	newGrid := createGrid(width, height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			neighbors := countNeighbors(grid, x, y)
			isAlive := grid[y][x] == 'c'

			if isAlive && (neighbors == 2 || neighbors == 3) {
				newGrid[y][x] = 'c'
			} else if !isAlive && neighbors == 3 {
				newGrid[y][x] = 'c'
			}
		}
	}
	return newGrid
}

func countNeighbors(grid [][]rune, x, y int) int {
	count := 0
	height := len(grid)
	width := len(grid[0])

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			newY := (y + dy + height) % height
			newX := (x + dx + width) % width

			if grid[newY][newX] == 'c' {
				count++
			}
		}
	}
	return count
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
