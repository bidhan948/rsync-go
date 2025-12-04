package ui

import (
	"fmt"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

func PrintBanner() {
	fmt.Println()
	fmt.Println(colorCyan + "╔═══════════════════════════════════════╗" + colorReset)
	fmt.Println(colorCyan + "║         rsync-go  🚀  mini CLI        ║" + colorReset)
	fmt.Println(colorCyan + "╚═══════════════════════════════════════╝" + colorReset)
	fmt.Println()
}

func Info(msg string) {
	fmt.Println(colorBlue + "[INFO] " + colorReset + msg)
}

func Success(msg string) {
	fmt.Println(colorGreen + "[OK]   " + colorReset + msg)
}

func Error(msg string) {
	fmt.Println(colorRed + "[ERR]  " + colorReset + msg)
}

func Warn(msg string) {
	fmt.Println(colorYellow + "[WARN] " + colorReset + msg)
}

func SpinnerPrefix(prefix string, done <-chan struct{}) {
	frames := []string{"⠋", "⠙", "⠸", "⠴", "⠦", "⠇"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r%s   \n", prefix) // clear line
			return
		default:
			fmt.Printf("\r%s %s", frames[i%len(frames)], prefix)
			i++
			time.Sleep(120 * time.Millisecond)
		}
	}
}
