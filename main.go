package main

import (
	"embed"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mimarea/internal/api"
)

//go:embed web/*
var embeddedWebFS embed.FS

func main() {
	portFlag := flag.Int("port", 8080, "Puerto HTTP para el servidor de mareas")
	headlessFlag := flag.Bool("headless", false, "Ejecutar en segundo plano sin ventana gráfica de control")
	noBrowserFlag := flag.Bool("no-browser", false, "No abrir el navegador automáticamente al iniciar")
	flag.Parse()

	port := *portFlag
	if port <= 0 {
		port = 8080
	}

	// 1. Iniciar servidor HTTP con API y archivos web
	apiServer := api.NewServer(port, "./web", embeddedWebFS, true)
	if err := apiServer.Start(); err != nil {
		log.Fatalf("Error iniciando servidor HTTP en puerto %d: %v", port, err)
	}

	webURL := fmt.Sprintf("http://localhost:%d", port)
	localIP := api.GetLocalIP()
	networkURL := fmt.Sprintf("http://%s:%d", localIP, port)

	// 2. Abrir automáticamente en el navegador web predeterminado
	if !*noBrowserFlag {
		go func() {
			time.Sleep(300 * time.Millisecond)
			openBrowser(webURL)
		}()
	}

	// 3. Si se solicita modo headless / consola, esperar señal de parada
	if *headlessFlag {
		fmt.Printf("\n🌊 MiMarea Gijón — Servidor Activo (Modo Consola)\n")
		fmt.Printf("• URL Local:     %s\n", webURL)
		fmt.Printf("• Red Local:     %s\n", networkURL)
		fmt.Printf("• Estación:      Puerto de El Musel (Gijón, Asturias)\n")
		fmt.Printf("• Motor:         Cálculo Armónico Perpetuo\n")
		fmt.Println("Presione Ctrl+C para detener el servicio...")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nDeteniendo servidor...")
		_ = apiServer.Stop()
		fmt.Println("Servidor detenido. ¡Hasta pronto!")
		return
	}

	// 4. Modo predeterminado: Ventana de control ligera Fyne (estilo GoJubila PRO)
	runMiniControlWindow(apiServer, webURL, networkURL, port)
}

func runMiniControlWindow(apiServer *api.Server, webURL string, networkURL string, port int) {
	myApp := app.NewWithID("com.manuel.mimarea")
	win := myApp.NewWindow("MiMarea Gijón — Servidor Web")

	win.SetOnClosed(func() {
		_ = apiServer.Stop()
	})

	// Título centrado con estilo náutico
	lblTitle := widget.NewLabelWithStyle("🌊 MiMarea Gijón — Servidor Activo", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Indicador visual esmeralda de servidor en línea
	greenDot := canvas.NewCircle(color.NRGBA{R: 34, G: 197, B: 94, A: 255})
	dotBox := container.NewCenter(container.NewGridWrap(fyne.NewSize(10, 10), greenDot))
	statusText := widget.NewLabel(fmt.Sprintf("En línea atendiendo en puerto %d", port))
	statusRow := container.NewCenter(container.NewHBox(dotBox, statusText))

	// Enlaces de acceso
	lblLocal := widget.NewLabel(fmt.Sprintf("• Local (Navegador):    %s", webURL))
	lblNetwork := widget.NewLabel(fmt.Sprintf("• Red Local (Móvil):     %s", networkURL))
	lblStation := widget.NewLabel("• Estación:             Puerto de El Musel (Gijón)")
	lblEngine := widget.NewLabel("• Predicción:           Perpetua (Sin tablas anuales)")

	// Botones de acción rápida
	btnOpen := widget.NewButton("Abrir en Navegador", func() {
		openBrowser(webURL)
	})

	btnQuit := widget.NewButton("Detener y Salir", func() {
		_ = apiServer.Stop()
		win.Close()
	})

	buttonsRow := container.NewGridWithColumns(2, btnOpen, btnQuit)

	content := container.NewVBox(
		lblTitle,
		statusRow,
		widget.NewSeparator(),
		lblLocal,
		lblNetwork,
		lblStation,
		lblEngine,
		widget.NewSeparator(),
		buttonsRow,
	)

	win.SetContent(content)
	win.Resize(fyne.NewSize(420, 260))
	win.CenterOnScreen()
	win.ShowAndRun()
}

// openBrowser abre la URL en el navegador predeterminado del sistema
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
