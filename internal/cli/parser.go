// Package cli is responsible for parsing command-line arguments and providing a structured input for the application.
package cli

import "errors"

type Input struct {
	URL     string
	Quality string
}

func ParseArgs(args []string) (Input, error) {
	if len(args) == 0 {
		return Input{}, errors.New("URL é obrigatório")
	}

	input := Input{
		URL:     args[0],
		Quality: "720p",
	}

	if len(args) > 1 {
		input.Quality = args[1]
	}

	return input, nil
}
