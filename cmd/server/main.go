// Command server runs the statsfm-card HTTP service.
package main

import (
	"fmt"
	"os"
)

func main() {
	if _, err := fmt.Fprintln(os.Stdout, "statsfm-card: not yet implemented"); err != nil {
		os.Exit(1)
	}
}
