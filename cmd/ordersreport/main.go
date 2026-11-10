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

var re = regexp.MustCompile(`\A(\d+)(?:\.(\d{1,2}))?\z`)

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
	ErrEmptyFilepath           = errors.New("путь к файлу пуст")
	ErrMissingCmdFileArgument  = errors.New("отсутствует флаг -file")
)

type Order struct {
	ID        int
	Customer  string
	Category  string
	Amount    int64
	Status    string // Тут тоже давай использовать наш новый тип Status
	CreatedAt time.Time
}

type Line struct {
	Order
	LineNumber int
}

type Loader interface {
	Load() ([]string, error)
}

func load(l Loader) ([]string, error) {
	return l.Load()
}

type FileLoader struct {
	filepath string
}

type MemoryLoader struct{}

func NewFileLoader(filePath string) *FileLoader {
	return &FileLoader{filepath: filePath}
}

func (f FileLoader) Load() ([]string, error) {
	var lines []string

	// Open file
	file, err := os.Open(f.filepath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл %q: %w", f.filepath, err)
	}
	// На уроке мы с тобой обсуждали, что лучше никогда не игнорировать ошибки. Делать это можно с полным понимаем того, почему и зачем мы так делаем
	// Давай тут проверять ошибку у file.Close и если есть ошибка, то выводить ее в консоль с помощью log.Print(err). Само логирование еще позже обсудим
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "неудачное закрытие файла "+f.filepath, err)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		lines = append(lines, line)
	}
	err = scanner.Err()
	if err != nil {
		return lines, errors.New("")
	}
	return lines, nil
}

func (MemoryLoader) Load() ([]string, error) {
	arr := make([]string, 0, 30)
	arr = append(arr,
		"id,customer,category,amount,status,created_at",
		"1001,Анна,books,1250.50,paid,2026-09-01",
		"1002,Борис,electronics,15990.00,paid,2026-09-01",
		"1003,Анна,books,430,paid,2026-09-02",
		"1004,Виктор,toys,899.90,pending,2026-09-02",
		"1005,Галина,books,1200.00,cancelled,2026-09-03",
		"1006,Борис,toys,2100.00,paid,2026-09-03",
		"1007,Дмитрий,electronics,7350.25,paid,2026-09-04",
		"1008, Галина , garden ,560.00,paid,2026-09-04",
		"1009,Анна,garden,abc,paid,2026-09-05",
		"1010,Елена,books,-50.00,paid,2026-09-05",
		"1011,Виктор,toys,310.5,paid,2026-09-06",
		"1001,Борис,books,99.00,paid,2026-09-06",
		"1012,Дмитрий,garden,1800.00,paid,2026-02-30",
		"1013,,books,100.00,paid,2026-09-07",
		"1014,Елена,toys,12.345,paid,2026-09-07",
		"1015,Елена,electronics,3200.00,refunded,2026-09-08",
		"1016,Галина,books,0,paid,2026-09-08",
		"1017,Елена,garden,640.00,paid,2026-09-08",
		"1018,Борис,electronics,2500.00,pending,2026-09-09",
		"1019,Дмитрий,books,1000.00,paid,2026-09-09",
		"1020,Анна,toys,215.75,paid,2026-09-10,extra",
		"1021,Виктор,books,75.00,paid",
		"x22,Галина,garden,100.00,paid,2026-09-10",
		"1023,Елена,books,1560.00,paid,2026-09-11",
		"1024,Анна,garden,1000.00,cancelled,2026-09-11",
		"1025,Борис,books,1000.00,paid,2026-09-12",
		"1026,Анна,garden,519.50,paid,2026-09-12",
		"1027,Ирина,music,19.99,paid,2026-09-13",
		"1028,Ирина,music,4.35,paid,2026-09-13",
	)
	return arr, nil
}

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

type Status struct {
	Status string
}

func (s Status) IsValid() bool {
	switch s.Status {
	case "paid", "pending", "cancelled":
		return true
	default:
		return false
	}
}

func printReport(acceptedLinesCount int, rejectedLinesCount int) {
	totalLines := acceptedLinesCount + rejectedLinesCount
	fmt.Printf("Всего строк данных: %d\n", totalLines) // Если нам нужно вывести данные в stdout, то лучше использовать fmt.Printf. Суть таже самая, только он сам под капотом пишет в stdout
	fmt.Printf("Принято: %d\n", acceptedLinesCount)
	fmt.Printf("Отклонено: %d\n", rejectedLinesCount)
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

	// На уроке мы с тобой обсуждали, как нам создать свой тип Status и вынести данную проверку туда
	// type Status string
	//
	// func (s Status) Valid() bool {...}
	//
	// Попробуй сам вынести данную проверку в метод и заменить им код ниже. Для этого нам нужно будет на входе твоей программы
	// присвоить строковый статус к типу Status
	statusStruct := Status{Status: status}
	if !statusStruct.IsValid() {
		return Line{}, &ValidationError{Err: ErrUnknownStatus, LineNumber: lineIdx + 1, Field: "Status"}
	}

	amountInt, err = processAmount(amount, lineIdx)
	if err != nil {
		return Line{}, err
	}

	if amountInt <= 0 {
		return Line{}, &ValidationError{Err: ErrNonPositiveAmount, LineNumber: lineIdx + 1, Field: "Amount"}
	}

	parsedCreatedAt, err := processCreatedAt(createdAt, lineIdx)
	if err != nil {
		return Line{}, err
	}

	seenIds[idInt] = true

	order := Order{
		ID:        idInt,
		Customer:  customer,
		Category:  category,
		Status:    status,
		Amount:    amountInt, // тут приводить к int64 нет смысла, тк переменная уже int64
		CreatedAt: parsedCreatedAt}

	return Line{Order: order, LineNumber: lineIdx + 1}, nil
}

// Данная функция должна возвращать ошибку, а какой код ошибки вызывать пусть решает main
func run() error {
	// в go нет необходимости все переменные нужные для работы функции выносить вверх. Объявляй их по мере необходимости
	// Так же переменные через var имеет смысл объявлять когда явно нужно дефолтное значение, или когда мы знаем что значение переменной может поменять внутри блока if/for/switch

	// Теперь у нас есть своя структура FileLoader и мы можем хранить путь к файлу как поле структуры
	// Давай создадим функцию
	// func NewFileLoader(filePath string) (*FileLoader) {}
	// которая будет создавать FileLoader и внутри функции сохранять путь к файлу внутри поля
	filepathPtr := flag.String("file", "", "Путь к CSV файлу")
	flag.Parse()
	filepath := *filepathPtr

	if filepath == "" {
		return fmt.Errorf("run: %w", ErrMissingCmdFileArgument)
	}

	seenIds := make(map[int]bool)

	fileLoader := NewFileLoader(filepath)
	lines, err := load(fileLoader)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	var (
		acceptedLines              []Line
		rejectedLinesErrorMessages []error
	)

	for idx, rawLine := range lines {
		if idx == 0 {
			continue
		}
		line, err := processLine(rawLine, idx, seenIds) // тут нет необходимости в новой переменной err
		if err != nil {
			rejectedLinesErrorMessages = append(rejectedLinesErrorMessages, err)
			continue
		}
		acceptedLines = append(acceptedLines, line)
	}
	printReport(len(acceptedLines), len(rejectedLinesErrorMessages))
	printErrors(rejectedLinesErrorMessages)
	return nil
}

func main() {
	err := run() // возвращает ошибку и в main ее обрабатываем. В зависимости от ошибки вызываем интересующий нас exit err

	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, err)

	switch {
	case errors.Is(err, ErrEmptyFilepath) || errors.Is(err, os.ErrNotExist):
		os.Exit(1)
	case errors.Is(err, ErrMissingCmdFileArgument):
		os.Exit(2)
	default:
		os.Exit(1)
	}
}
