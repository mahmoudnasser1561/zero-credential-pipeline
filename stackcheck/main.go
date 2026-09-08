package main

import (
	"os"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
