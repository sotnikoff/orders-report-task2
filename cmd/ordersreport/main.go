package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math"
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

var (
	ErrWrongNumberOfColumns    = errors.New("число колонок не равно 6")
	ErrInvalidDate             = errors.New("ошибка разбора времени")
	ErrInvalidAmountFormat     = errors.New("неверный формат суммы")
	ErrAmountParsing           = errors.New("ошибка преобразования суммы")
	ErrAmountTooLarge          = errors.New("сумма превышает максимально допустимое значение")
	ErrFractionalAmountParsing = errors.New("ошибка преобразования дробной части суммы")
	ErrInvalidID               = errors.New("ошибка разбора ID")
	ErrNonPositiveID           = errors.New("ID должен быть больше 0")
	ErrDuplicateID             = errors.New("ID повторяющийся")
	ErrEmptyCustomer           = errors.New("клиент - пустое значение")
	ErrEmptyCategory           = errors.New("категория - пустое значение")
	ErrEmptyStatus             = errors.New("статус - пустое значение")
	ErrUnknownStatus           = errors.New("неизвестный статус")
	ErrNonPositiveAmount       = errors.New("сумма должна быть строго больше нуля")
)

type ValidationError struct {
	Err        error
	LineNumber int
	Field      string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("строка %d, поле %s: %v", e.LineNumber, e.Field, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
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
		return time.Time{}, &ValidationError{Err: ErrInvalidDate, LineNumber: lineIdx + 1, Field: "CreatedAt"}
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
		return 0, &ValidationError{Err: ErrInvalidAmountFormat, LineNumber: lineIdx + 1, Field: "Amount"}
	}

	// whole amount
	wholeInt, err = strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		// error parsing int
		return 0, &ValidationError{Err: ErrAmountParsing, LineNumber: lineIdx + 1, Field: "Amount"}
	}

	amountInt = wholeInt

	if m[2] != "" {
		fractionalPartInt, err = strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			// error parsing fractional part
			return 0, &ValidationError{Err: ErrFractionalAmountParsing, LineNumber: lineIdx + 1, Field: "Amount"}
		}
		if len(m[2]) == 1 {
			fractionalPartInt = fractionalPartInt * 10
		}
	}
	if wholeInt > (math.MaxInt64-fractionalPartInt)/100 {
		return 0, &ValidationError{Err: ErrAmountTooLarge, LineNumber: lineIdx + 1, Field: "Amount"}
	}
	amountInt = (amountInt * 100) + fractionalPartInt
	return amountInt, nil
}

func processLine(stringLine string, lineIdx int, seenIds map[int]bool) (Line, error) {
	data := strings.Split(stringLine, ",")
	if len(data) != 6 {
		// .... error != 6
		return Line{}, &ValidationError{Err: ErrWrongNumberOfColumns, LineNumber: lineIdx + 1, Field: "Columns"}
	}

	id := strings.TrimSpace(data[0])
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
		return Line{}, &ValidationError{Err: ErrInvalidID, LineNumber: lineIdx + 1, Field: "ID"}
	}

	if idInt <= 0 {
		return Line{}, &ValidationError{Err: ErrNonPositiveID, LineNumber: lineIdx + 1, Field: "ID"}
	}

	if seenIds[idInt] {
		return Line{}, &ValidationError{Err: ErrDuplicateID, LineNumber: lineIdx + 1, Field: "ID"}
	}

	if len(customer) == 0 {
		return Line{}, &ValidationError{Err: ErrEmptyCustomer, LineNumber: lineIdx + 1, Field: "Customer"}
	}

	if len(category) == 0 {
		return Line{}, &ValidationError{Err: ErrEmptyCategory, LineNumber: lineIdx + 1, Field: "Category"}
	}
	if len(status) == 0 {
		return Line{}, &ValidationError{Err: ErrEmptyStatus, LineNumber: lineIdx + 1, Field: "Status"}
	}

	switch status {
	case "paid", "pending", "cancelled":
		// ok
	default:
		return Line{}, &ValidationError{Err: ErrUnknownStatus, LineNumber: lineIdx + 1, Field: "Status"}
	}

	amountInt, err = processAmount(amount, lineIdx)
	if err != nil {
		return Line{}, err
	}

	if amountInt <= 0 {
		return Line{}, &ValidationError{Err: ErrNonPositiveAmount, LineNumber: lineIdx + 1, Field: "Amount"}
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
