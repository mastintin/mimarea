package api

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"mimarea/internal/tides"
)

// Server maneja el servidor HTTP y la API REST para MiMarea
type Server struct {
	port       int
	webDir     string
	embedFS    embed.FS
	hasEmbed   bool
	httpServer *http.Server
	listener   net.Listener
}

// NewServer crea una nueva instancia del servidor web y API
func NewServer(port int, webDir string, embeddedFS embed.FS, hasEmbed bool) *Server {
	if port <= 0 {
		port = 8080
	}
	if webDir == "" {
		webDir = "./web"
	}
	s := &Server{
		port:     port,
		webDir:   webDir,
		embedFS:  embeddedFS,
		hasEmbed: hasEmbed,
	}
	return s
}

// Start arranca el listener y el servidor HTTP en una goroutine
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("no se pudo abrir puerto %d: %w", s.port, err)
	}
	s.listener = ln

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	handler := s.corsMiddleware(mux)
	s.httpServer = &http.Server{
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := s.httpServer.Serve(s.listener); err != nil && err != http.ErrServerClosed {
			log.Printf("[MiMarea] Error en servidor HTTP: %v", err)
		}
	}()

	return nil
}

// Stop detiene ordenadamente el servidor HTTP
func (s *Server) Stop() error {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// GetPort devuelve el puerto efectivo del servidor
func (s *Server) GetPort() int {
	return s.port
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Endpoints API
	mux.HandleFunc("/api/ping", s.handlePing)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/day", s.handleDay)
	mux.HandleFunc("/api/month", s.handleMonth)

	// Servir archivos estáticos del frontend
	var fileHandler http.Handler
	if _, err := os.Stat(filepath.Join(s.webDir, "index.html")); err == nil {
		// Modo desarrollo / carpeta local en disco
		fileHandler = http.FileServer(http.Dir(s.webDir))
	} else if s.hasEmbed {
		// Modo embebido en el binario
		sub, err := fs.Sub(s.embedFS, "web")
		if err == nil {
			fileHandler = http.FileServer(http.FS(sub))
		}
	}

	if fileHandler != nil {
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			fileHandler.ServeHTTP(w, r)
		}))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprintln(w, "<h2>MiMarea Gijón</h2><p>Servicio activo en puerto "+strconv.Itoa(s.port)+"</p>")
		})
	}
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"service":   "mimarea",
		"port_name": "Gijón (El Musel)",
		"time":      time.Now().Format(time.RFC3339),
	})
}

// GET /api/status - Devuelve el estado en tiempo real (marea ahora, subiendo/bajando, prox evento)
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	datum := tides.DatumType(r.URL.Query().Get("datum"))
	if datum != tides.DatumMSL {
		datum = tides.DatumPort
	}

	loc := tides.GetGijonLocation()
	now := time.Now().In(loc)

	dayRes := tides.CalculateDayTides(now, datum)
	current := tides.GetCurrentStatus(now, datum, dayRes.Events)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"now":          now.Format("2006-01-02 15:04:05"),
		"datum":        datum,
		"datum_name":   dayRes.DatumName,
		"current":      current,
		"coefficient":  dayRes.Coefficient,
		"coeff_level":  dayRes.CoeffLevel,
		"moon_phase":   dayRes.MoonPhase,
		"water_temp_c": dayRes.WaterTempC,
		"today_events": dayRes.Events,
	})
}

// GET /api/day?date=YYYY-MM-DD&datum=port|msl
func (s *Server) handleDay(w http.ResponseWriter, r *http.Request) {
	datum := tides.DatumType(r.URL.Query().Get("datum"))
	if datum != tides.DatumMSL {
		datum = tides.DatumPort
	}

	loc := tides.GetGijonLocation()
	dateStr := r.URL.Query().Get("date")

	var targetDate time.Time
	if dateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", dateStr, loc)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Formato de fecha inválido. Utilice YYYY-MM-DD (ej: 2026-10-20)",
			})
			return
		}
		targetDate = parsed
	} else {
		targetDate = time.Now().In(loc)
	}

	result := tides.CalculateDayTides(targetDate, datum)
	writeJSON(w, http.StatusOK, result)
}

// GET /api/month?year=YYYY&month=MM&datum=port|msl
func (s *Server) handleMonth(w http.ResponseWriter, r *http.Request) {
	datum := tides.DatumType(r.URL.Query().Get("datum"))
	if datum != tides.DatumMSL {
		datum = tides.DatumPort
	}

	loc := tides.GetGijonLocation()
	now := time.Now().In(loc)

	year := now.Year()
	month := int(now.Month())

	if yStr := r.URL.Query().Get("year"); yStr != "" {
		if y, err := strconv.Atoi(yStr); err == nil && y > 1900 && y < 2200 {
			year = y
		}
	}
	if mStr := r.URL.Query().Get("month"); mStr != "" {
		if m, err := strconv.Atoi(mStr); err == nil && m >= 1 && m <= 12 {
			month = m
		}
	}

	result := tides.CalculateMonthTides(year, month, datum)
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// GetLocalIP devuelve la dirección IP principal de la red local
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				if !strings.HasPrefix(ipStr, "169.254.") {
					return ipStr
				}
			}
		}
	}
	return "127.0.0.1"
}
