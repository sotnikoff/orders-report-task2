package csvfile

import (
	"bufio"
	"fmt"
	"os"
)

type Source struct {
	filepath string
}

func NewSource(filepath string) *Source {
	return &Source{filepath: filepath}
}

func (l Source) Load() ([]string, error) {
	var lines []string

	// Open file
	file, err := os.Open(l.filepath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл %q: %w", l.filepath, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "неудачное закрытие файла "+l.filepath, err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		lines = append(lines, line)
	}
	err = scanner.Err()
	if err != nil {
		return lines, fmt.Errorf("не удалось прочитать файл %q: %w", l.filepath, err)
	}
	return lines, nil
}
