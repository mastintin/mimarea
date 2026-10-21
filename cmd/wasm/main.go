package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"
	_ "time/tzdata"

	"mimarea/internal/tides"
)

func main() {
	loc := tides.GetGijonLocation()

	// 1. Registrar función JS: calcDayTides(dateStr, datum) -> JSON string
	js.Global().Set("calcDayTides", js.FuncOf(func(this js.Value, args []js.Value) any {
		dateStr := ""
		datumStr := "port"

		if len(args) > 0 && args[0].Type() == js.TypeString {
			dateStr = args[0].String()
		}
		if len(args) > 1 && args[1].Type() == js.TypeString {
			datumStr = args[1].String()
		}

		datum := tides.DatumType(datumStr)
		if datum != tides.DatumMSL {
			datum = tides.DatumPort
		}

		var targetDate time.Time
		if dateStr != "" {
			parsed, err := time.ParseInLocation("2006-01-02", dateStr, loc)
			if err != nil {
				return createErrorJSON(fmt.Sprintf("Fecha inválida: %v", err))
			}
			targetDate = parsed
		} else {
			targetDate = time.Now().In(loc)
		}

		res := tides.CalculateDayTides(targetDate, datum)
		bytes, err := json.Marshal(res)
		if err != nil {
			return createErrorJSON(fmt.Sprintf("Error generando JSON: %v", err))
		}
		return string(bytes)
	}))

	// 2. Registrar función JS: calcMonthTides(year, month, datum) -> JSON string
	js.Global().Set("calcMonthTides", js.FuncOf(func(this js.Value, args []js.Value) any {
		now := time.Now().In(loc)
		year := now.Year()
		month := int(now.Month())
		datumStr := "port"

		if len(args) > 0 && args[0].Type() == js.TypeNumber {
			year = args[0].Int()
		}
		if len(args) > 1 && args[1].Type() == js.TypeNumber {
			month = args[1].Int()
		}
		if len(args) > 2 && args[2].Type() == js.TypeString {
			datumStr = args[2].String()
		}

		datum := tides.DatumType(datumStr)
		if datum != tides.DatumMSL {
			datum = tides.DatumPort
		}

		res := tides.CalculateMonthTides(year, month, datum)
		bytes, err := json.Marshal(res)
		if err != nil {
			return createErrorJSON(fmt.Sprintf("Error generando JSON: %v", err))
		}
		return string(bytes)
	}))

	// 3. Registrar función JS: calcNext30DaysTides(dateStr, datum) -> JSON string
	js.Global().Set("calcNext30DaysTides", js.FuncOf(func(this js.Value, args []js.Value) any {
		dateStr := ""
		datumStr := "port"

		if len(args) > 0 && args[0].Type() == js.TypeString {
			dateStr = args[0].String()
		}
		if len(args) > 1 && args[1].Type() == js.TypeString {
			datumStr = args[1].String()
		}

		datum := tides.DatumType(datumStr)
		if datum != tides.DatumMSL {
			datum = tides.DatumPort
		}

		var targetDate time.Time
		if dateStr != "" {
			parsed, err := time.ParseInLocation("2006-01-02", dateStr, loc)
			if err != nil {
				return createErrorJSON(fmt.Sprintf("Fecha inválida: %v", err))
			}
			targetDate = parsed
		} else {
			targetDate = time.Now().In(loc)
		}

		res := tides.CalculateNext30DaysTides(targetDate, datum)
		bytes, err := json.Marshal(res)
		if err != nil {
			return createErrorJSON(fmt.Sprintf("Error generando JSON: %v", err))
		}
		return string(bytes)
	}))

	// 4. Registrar función JS: calcWaterLevelNow(datum) -> float64
	js.Global().Set("calcWaterLevelNow", js.FuncOf(func(this js.Value, args []js.Value) any {
		datumStr := "port"
		if len(args) > 0 && args[0].Type() == js.TypeString {
			datumStr = args[0].String()
		}
		datum := tides.DatumType(datumStr)
		if datum != tides.DatumMSL {
			datum = tides.DatumPort
		}
		now := time.Now().In(loc)
		return tides.CalculateWaterLevel(now, datum)
	}))

	fmt.Println("🌊 [MiMarea Wasm] Motor armónico perpetuo iniciado en el navegador para Gijón (El Musel)")

	// Mantener el runtime de WebAssembly vivo
	select {}
}

func createErrorJSON(msg string) string {
	m := map[string]string{"error": msg}
	b, _ := json.Marshal(m)
	return string(b)
}
