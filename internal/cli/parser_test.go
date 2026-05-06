package cli

import "testing"

func TestParseArgs_Success(t *testing.T) {
	args := []string{"http://video", "1080p"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("Não esperava erro, mas recebeu: %v", err)
	}

	if input.URL != "http://video" {
		t.Errorf("URL inválida")
	}

	if input.Quality != "1080p" {
		t.Errorf("Qualidade inválida")
	}
}

func TestParseArgs_DefaultQuality(t *testing.T) {
	args := []string{"http://video"}

	input, _ := ParseArgs(args)

	if input.Quality != "720p" {
		t.Errorf("Padrão da qualidade deve ser 720p")
	}
}

func TestParseArgs_Error(t *testing.T) {
	_, err := ParseArgs([]string{})

	if err == nil {
		t.Errorf("Esperava um erro para argumentos vazios")
	}
}
