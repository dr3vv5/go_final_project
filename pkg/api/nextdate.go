package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func AfterNow(date, now time.Time) bool {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return date.After(now)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if dstart == "" || repeat == "" {
		return "", errors.New("пустые параметры dstart или repeat")
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректный формат даты dstart: %w", err)
	}

	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", errors.New("неверный формат правила повторения")
	}

	ruleType := parts[0]

	switch ruleType {
	case "d":
		return handleDays(date, now, parts)
	case "y":
		return handleYear(date, now, parts)
	case "w":
		return handleWeek(date, now, parts)
	case "m":
		return handleMonth(date, now, parts)
	default:
		return "", fmt.Errorf("неподдерживаемый формат правила: %s", ruleType)
	}
}

func handleYear(date time.Time, now time.Time, parts []string) (string, error) {

	if len(parts) > 1 {
		return "", errors.New("правило 'y' не принимает дополнительных аргументов")
	}

	for {
		date = date.AddDate(1, 0, 0)
		if AfterNow(date, now) {
			return date.Format(dateFormat), nil
		}
	}
}

func handleDays(date time.Time, now time.Time, parts []string) (string, error) {

	if len(parts) < 2 {
		return "", errors.New("для правила 'd' требуется указать количество дней")
	}

	days, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", errors.New("некорректное число дней")
	}

	if days < 1 || days > 400 {
		return "", errors.New("количество дней должно быть от 1 до 400")
	}

	for {
		date = date.AddDate(0, 0, days)
		if AfterNow(date, now) {
			return date.Format(dateFormat), nil
		}
	}
}

func handleWeek(date time.Time, now time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", errors.New("для правила 'w' требуется список дней недели (например, '1,4,5')")
	}

	validDays := make(map[int]bool)
	dayStrs := strings.Split(parts[1], ",")

	for _, s := range dayStrs {
		dayNum, err := strconv.Atoi(s)
		if err != nil || dayNum < 1 || dayNum > 7 {
			return "", errors.New("некорректный день недели (должно быть от 1 до 7)")
		}
		validDays[dayNum] = true
	}

	// Идем вперед по дням, пока не найдем подходящий
	for {
		date = date.AddDate(0, 0, 1) // Шаг на 1 день вперед

		wd := date.Weekday()
		taskDay := int(wd)
		if taskDay == 0 {
			taskDay = 7 // Конвертация: Go Sunday(0) -> Задача Sunday(7)
		}

		if validDays[taskDay] {
			if AfterNow(date, now) {
				return date.Format(dateFormat), nil
			}
		}
	}
}
func handleMonth(date time.Time, now time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", errors.New("для правила 'm' требуется список дней месяца")
	}

	daysStr := parts[1]
	monthsStr := ""
	if len(parts) >= 3 {
		monthsStr = parts[2]
	}

	dayParts := strings.Split(daysStr, ",")

	positiveDays := make(map[int]bool)
	negativeDays := []int{}

	for _, s := range dayParts {
		d, err := strconv.Atoi(s)
		if err != nil {
			return "", errors.New("некорректный формат дня месяца")
		}

		if d == 0 {
			return "", errors.New("день месяца не может быть равен 0")
		}

		if d < 0 {
			if d != -1 && d != -2 {
				return "", errors.New("некорректный формат правила: отрицательные смещения могут быть только -1 или -2")
			}
			negativeDays = append(negativeDays, d)
		} else {
			if d > 31 {
				return "", errors.New("день месяца не может быть больше 31")
			}
			positiveDays[d] = true
		}
	}

	validMonths := make(map[int]bool)
	if monthsStr != "" {
		monthParts := strings.Split(monthsStr, ",")
		for _, s := range monthParts {
			m, err := strconv.Atoi(s)
			if err != nil || m < 1 || m > 12 {
				return "", errors.New("некорректный месяц (должно быть от 1 до 12)")
			}
			validMonths[m] = true
		}
	}

	for {
		date = date.AddDate(0, 0, 1)

		y, m, d := date.Date()
		monthInt := int(m)

		if len(validMonths) > 0 && !validMonths[monthInt] {
			continue
		}

		if positiveDays[d] {
			if AfterNow(date, now) {
				return date.Format(dateFormat), nil
			}
			continue
		}

		if len(negativeDays) > 0 {
			lastDay := time.Date(y, m+1, 1, 0, 0, 0, 0, date.Location()).AddDate(0, 0, -1).Day()

			for _, neg := range negativeDays {
				target := lastDay + neg + 1
				if target == d {
					if AfterNow(date, now) {
						return date.Format(dateFormat), nil
					}
					break
				}
			}
		}
	}
}
