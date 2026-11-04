package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Order struct {
	ID        int
	Customer  string
	Category  string
	Amount    int64
	Status    string
	CreatedAt time.Time
}

type Line struct {
	Order
	LineNumber int
}

func printReport(acceptedLinesCount int, rejectedLinesCount int) {
	totalLines := acceptedLinesCount + rejectedLinesCount
	fmt.Fprintf(os.Stdout, "Всего строк данных: %d\n", totalLines)
	fmt.Fprintf(os.Stdout, "Принято: %d\n", acceptedLinesCount)
	fmt.Fprintf(os.Stdout, "Отклонено: %d\n", rejectedLinesCount)
}

func printErrors(rejectedLinesErrorMessages []error) {
	for _, err := range rejectedLinesErrorMessages {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

func processCreatedAt(createdAt string, lineIdx int) (time.Time, error) {
	parsedTime, err := time.Parse("2006-01-02", createdAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("строка %d: ошибка разбора времени", lineIdx+1)
	}
	return parsedTime, nil
}

var re = regexp.MustCompile(`\A(\d+)(?:\.(\d{1,2}))?\z`)

func processAmount(amount string, lineIdx int) (int64, error) {
	var (
		wholeInt          int64
		fractionalPartInt int64
		amountInt         int64
		err               error
	)

	m := re.FindStringSubmatch(amount)

	if m == nil {
		// error bad amount regexp failed
		return 0, fmt.Errorf("строка %d: неверный формат суммы", lineIdx+1)
	}

	// whole amount
	wholeInt, err = strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		// error parsing int
		return 0, fmt.Errorf("строка %d: Ошибка преобразования суммы", lineIdx+1)
	}

	amountInt = wholeInt

	if m[2] != "" {
		fractionalPartInt, err = strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			// error parsing fractional part
			return 0, fmt.Errorf("строка %d: Ошибка преобразования дробной части суммы", lineIdx+1)
		}
		if len(m[2]) == 1 {
			fractionalPartInt = fractionalPartInt * 10
		}
		amountInt = (amountInt * 100) + fractionalPartInt
	} else {
		amountInt = amountInt * 100
	}
	return amountInt, nil
}

func processLine(stringLine string, lineIdx int, seenIds map[int]bool) (Line, error) {
	data := strings.Split(stringLine, ",")
	if len(data) != 6 {
		// .... error != 6
		return Line{}, fmt.Errorf("строка %d: Число колонок не равно 6", lineIdx+1)
	}

	id := data[0]
	customer := strings.TrimSpace(data[1])
	category := strings.TrimSpace(data[2])
	amount := strings.TrimSpace(data[3])
	status := strings.TrimSpace(data[4])
	createdAt := strings.TrimSpace(data[5])
	parsedCreatedAt := time.Time{}

	var (
		err       error
		idInt     int
		amountInt int64
	)

	idInt, err = strconv.Atoi(id)
	if err != nil {
		return Line{}, fmt.Errorf("строка %d: ошибка разбора ID", lineIdx+1)
	}

	if idInt < 0 {
		return Line{}, fmt.Errorf("строка %d: ID отрицательный", lineIdx+1)
	}

	if seenIds[idInt] {
		return Line{}, fmt.Errorf("строка %d: ID повторяющийся", lineIdx+1)
	}

	if len(customer) == 0 {
		return Line{}, fmt.Errorf("строка %d: клиент - пустое значение", lineIdx+1)
	}

	if len(category) == 0 {
		return Line{}, fmt.Errorf("строка %d: категория - пустое значение", lineIdx+1)
	}
	if len(status) == 0 {
		return Line{}, fmt.Errorf("строка %d: статус - пустое значение", lineIdx+1)
	}

	switch status {
	case "paid", "pending", "cancelled":
		// ok
	default:
		return Line{}, fmt.Errorf("строка %d: неизвестный статус", lineIdx+1)
	}

	amountInt, err = processAmount(amount, lineIdx)
	if err != nil {
		return Line{}, err
	}

	if amountInt <= 0 {
		return Line{}, fmt.Errorf("строка %d: Сумма должна быть строго больше нуля", lineIdx+1)
	}

	parsedCreatedAt, err = processCreatedAt(createdAt, lineIdx)
	if err != nil {
		return Line{}, err
	}

	seenIds[idInt] = true

	order := Order{
		ID:        idInt,
		Customer:  customer,
		Category:  category,
		Status:    status,
		Amount:    int64(amountInt),
		CreatedAt: parsedCreatedAt}

	return Line{Order: order, LineNumber: lineIdx + 1}, nil
}

func run() int {
	filepathPtr := flag.String("file", "", "Ошибка: Путь к CSV файлу обязателен")
	flag.Parse()
	filepath := *filepathPtr

	if filepath == "" {
		fmt.Fprintln(os.Stderr, "Ошибка: не указан обязательный флаг -file")
		return 2
	}

	// Open file
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка: не удалось открыть файл "+filepath, err)
		return 1
	}
	defer file.Close()

	var lines []string
	var acceptedLines []Line
	var rejectedLinesErrorMessages []error

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		lines = append(lines, line)
	}

	err = scanner.Err()
	if err != nil {
		return 1
	}

	seenIds := make(map[int]bool)

	for idx, rawLine := range lines {
		if idx == 0 {
			continue
		}
		line, err := processLine(rawLine, idx, seenIds)
		if err != nil {
			rejectedLinesErrorMessages = append(rejectedLinesErrorMessages, err)
			continue
		}
		acceptedLines = append(acceptedLines, line)
	}
	printReport(len(acceptedLines), len(rejectedLinesErrorMessages))
	printErrors(rejectedLinesErrorMessages)
	return 0
}

func main() {
	code := run()
	os.Exit(code)
}
