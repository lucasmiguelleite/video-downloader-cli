package cli

import "testing"

func TestParseArgs_URLAndResolution(t *testing.T) {
	args := []string{"--url", "http://video", "--resolution", "1080p"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("Não esperava erro, mas recebeu: %v", err)
	}

	if input.URL != "http://video" {
		t.Errorf("URL esperada 'http://video', recebeu '%s'", input.URL)
	}

	if input.Quality != "1080p" {
		t.Errorf("Resolução esperada '1080p', recebeu '%s'", input.Quality)
	}
}

func TestParseArgs_DefaultResolution(t *testing.T) {
	args := []string{"--url", "http://video"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("Não esperava erro, mas recebeu: %v", err)
	}

	if input.Quality != "720p" {
		t.Errorf("Resolução padrão deve ser 720p, recebeu '%s'", input.Quality)
	}
}

func TestParseArgs_Help(t *testing.T) {
	args := []string{"--help"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("Não esperava erro, mas recebeu: %v", err)
	}

	if !input.ShowHelp {
		t.Error("ShowHelp deveria ser true")
	}
}

func TestParseArgs_Version(t *testing.T) {
	args := []string{"--version"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("Não esperava erro, mas recebeu: %v", err)
	}

	if !input.ShowVersion {
		t.Error("ShowVersion deveria ser true")
	}
}

func TestParseArgs_EmptyURL(t *testing.T) {
	args := []string{}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("Não esperava erro no parse, mas recebeu: %v", err)
	}

	if input.URL != "" {
		t.Errorf("URL deveria ser vazia, recebeu '%s'", input.URL)
	}
}

func TestParseArgs_InvalidFlag(t *testing.T) {
	args := []string{"--flag-inexistente"}

	_, err := ParseArgs(args)

	if err == nil {
		t.Error("Esperava erro para flag inválida")
	}
}
