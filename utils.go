package main

import (
	"fmt"
	"strings"
	"time"
)

// shortenName преобразует "Иванов Иван Владимирович" в "Иванов И. В."
func shortenName(fullName string) string {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return ""
	}
	parts := strings.Fields(fullName)
	if len(parts) >= 3 {
		firstName := []rune(parts[1])
		patronymic := []rune(parts[2])
		return fmt.Sprintf("%s %c. %c.", parts[0], firstName[0], patronymic[0])
	} else if len(parts) == 2 {
		firstName := []rune(parts[1])
		return fmt.Sprintf("%s %c.", parts[0], firstName[0])
	}
	return parts[0]
}

// formatDateRu форматирует дату в русский формат
func formatDateRu(s string) string {
	if s == "" {
		return "«___» ________ 20__ г."
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	months := []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
	return fmt.Sprintf("«%d» %s %d г.", t.Day(), months[t.Month()-1], t.Year())
}

// formatDateTimeRu форматирует дату и время
func formatDateTimeRu(s string) string {
	if s == "" {
		return "____________________"
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("02.01.2006 15:04")
}