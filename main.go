package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/sys/unix"
)

//go:embed icon/ceferino.ico
var iconData []byte

const configFile = ".config/newpipi_project_path"

type ProjectType int

const (
	Python ProjectType = iota
	Go
	Rust
	JavaScript
	TypeScript
	CPlusPlus
	CSharp
	Java
)

type ProjectSetup struct {
	window         fyne.Window
	parentPath     string
	projectName    string
	projectType    ProjectType
	statusMessage  string
	messageTimeout *time.Timer
	createBtn      *widget.Button
}

type Template struct {
	Name        string
	Description string
	Type        ProjectType
	Files       map[string]string
	Packages    []string
}

var templates = []Template{
	{
		Name:        "CLI App",
		Description: "Kommandozeilen-Anwendung",
		Type:        Python,
		Files: map[string]string{
			"src/cli.py": `import click
@click.command()
def main():
    click.echo("Hello CLI!")

if __name__ == "__main__":
    main()`,
		},
		Packages: []string{"click"},
	},
	// Weitere Templates...
}

func NewProjectSetup() *ProjectSetup {
	ps := &ProjectSetup{}
	if err := ps.loadProjectPath(); err != nil {
		log.Printf("Fehler beim Laden des Projektpfads: %v", err)
	}
	return ps
}

func (ps *ProjectSetup) loadProjectPath() error {
	log.Println("Lade Projektpfad...")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir nicht gefunden: %v", err)
	}

	configPath := filepath.Join(homeDir, configFile)
	content, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("config lesen fehlgeschlagen: %v", err)
		}
		return nil
	}

	path := strings.TrimSpace(string(content))
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		ps.parentPath = path
		log.Printf("Projektpfad geladen: %s", path)
	}
	return nil
}

func (ps *ProjectSetup) saveProjectPath() error {
	log.Printf("Speichere Projektpfad: %s", ps.parentPath)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir nicht gefunden: %v", err)
	}

	configPath := filepath.Join(homeDir, configFile)
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("config dir erstellen fehlgeschlagen: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(ps.parentPath), 0644); err != nil {
		return fmt.Errorf("config schreiben fehlgeschlagen: %v", err)
	}

	log.Println("Projektpfad erfolgreich gespeichert")
	return nil
}

func (ps *ProjectSetup) showMessage(msg string) {
	log.Printf("Zeige Nachricht: %s", msg)
	ps.statusMessage = msg
	if ps.messageTimeout != nil {
		ps.messageTimeout.Stop()
	}
	ps.messageTimeout = time.AfterFunc(2*time.Second, func() {
		ps.statusMessage = ""
		ps.window.Canvas().Refresh(ps.window.Content())
	})
}

func (ps *ProjectSetup) createProject() error {
	log.Println("Starte Projekterstellung...")

	// Validierungen
	if ps.parentPath == "" {
		return fmt.Errorf("elternpfad darf nicht leer sein")
	}
	if ps.projectName == "" {
		return fmt.Errorf("projektname darf nicht leer sein")
	}
	if strings.ContainsAny(ps.projectName, " \t\n") {
		return fmt.Errorf("projektname darf keine Leerzeichen enthalten")
	}

	projectDir := filepath.Join(ps.parentPath, ps.projectName)
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("projektverzeichnis existiert bereits: %s", projectDir)
	}

	// Prüfe zuerst die Installation
	var err error
	switch ps.projectType {
	case Go:
		err = ps.checkGoInstallation()
	case Python:
		// Python-Check implementieren
	case Rust:
		err = ps.checkRustInstallation()
	case JavaScript:
		err = ps.checkJavaScriptInstallation()
	case TypeScript:
		// Prüfe sowohl Node.js als auch TypeScript
		if err = ps.checkJavaScriptInstallation(); err == nil {
			err = ps.checkTypeScriptInstallation()
		}
	case CPlusPlus:
		err = ps.checkCPlusPlusInstallation()
	case CSharp:
		err = ps.checkCSharpInstallation()
	case Java:
		err = ps.checkJavaInstallation()
	}

	if err != nil {
		return fmt.Errorf("installation prüfung fehlgeschlagen: %v", err)
	}

	// Wenn die Installation-Prüfung erfolgreich war, erstelle das Projekt
	switch ps.projectType {
	case Python:
		return ps.createPythonProject()
	case Go:
		return ps.createGoProject()
	case Rust:
		return ps.createRustProject()
	case JavaScript:
		return ps.createJavaScriptProject()
	case TypeScript:
		return ps.createTypeScriptProject()
	case CPlusPlus:
		return ps.createCPlusPlusProject()
	case CSharp:
		return ps.createCSharpProject()
	case Java:
		return ps.createJavaProject()
	}

	return nil
}

func (ps *ProjectSetup) createPythonProject() error {
	log.Println("Erstelle Python-Projekt...")
	projectDir := filepath.Join(ps.parentPath, ps.projectName)

	// Erstelle Projektverzeichnis
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("projektverzeichnis erstellen fehlgeschlagen: %v", err)
	}

	// Wechsel ins Projektverzeichnis
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("verzeichniswechsel fehlgeschlagen: %v", err)
	}

	// Erstelle virtuelle Umgebung
	log.Println("Erstelle virtuelle Umgebung...")
	cmd := exec.Command("python3", "-m", "venv", "venv")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("venv erstellen fehlgeschlagen: %v", err)
	}

	// Aktualisiere pip und installiere Pakete
	log.Println("Installiere Pakete...")
	cmd = exec.Command("sh", "-c", "source venv/bin/activate && pip install --upgrade pip && pip install numpy PyQt5")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("paketinstallation fehlgeschlagen: %v", err)
	}

	// Erstelle Projektstruktur
	log.Println("Erstelle Projektstruktur...")
	dirs := []string{"src", "tests"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("verzeichnis %s erstellen fehlgeschlagen: %v", dir, err)
		}
	}

	// Erstelle Dateien
	files := map[string]string{
		"src/__init__.py": "",
		"src/main.py": `import sys
from PyQt5.QtWidgets import QApplication, QMainWindow, QPushButton, QVBoxLayout, QWidget

class MainWindow(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setWindowTitle("PyQt5 Boilerplate")
        self.setGeometry(100, 100, 300, 200)

        layout = QVBoxLayout()
        
        button = QPushButton("Click me!")
        button.clicked.connect(self.button_clicked)
        layout.addWidget(button)

        central_widget = QWidget()
        central_widget.setLayout(layout)
        self.setCentralWidget(central_widget)

    def button_clicked(self):
        print("Button clicked!")

def main():
    app = QApplication(sys.argv)
    window = MainWindow()
    window.show()
    sys.exit(app.exec_())

if __name__ == "__main__":
    main()`,
		"tests/__init__.py": "",
		"README.md":         "",
		"requirements.txt":  "numpy\nPyQt5\n",
		".gitignore":        "/venv\n__pycache__\n*.pyc\n",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("datei %s erstellen fehlgeschlagen: %v", path, err)
		}
	}

	// Öffne Terminal
	log.Println("Öffne Terminal...")
	terminalCmd := fmt.Sprintf("wezterm start --cwd '%s' --always-new-process -- bash -c 'source venv/bin/activate && echo \"python src/main.py\" && bash'", projectDir)
	cmd = exec.Command("sh", "-c", terminalCmd)
	if err := cmd.Start(); err != nil {
		log.Printf("Terminal öffnen fehlgeschlagen: %v", err)
	}

	// Warte kurz, damit das Terminal Zeit hat zu starten
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ps *ProjectSetup) createGoProject() error {
	log.Println("Erstelle Go-Projekt...")
	projectDir := filepath.Join(ps.parentPath, ps.projectName)

	// Erstelle Projektverzeichnis
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("projektverzeichnis erstellen fehlgeschlagen: %v", err)
	}

	// Wechsel ins Projektverzeichnis
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("verzeichniswechsel fehlgeschlagen: %v", err)
	}

	// Initialisiere Go-Modul
	log.Println("Initialisiere Go-Modul...")
	cmd := exec.Command("go", "mod", "init", ps.projectName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod init fehlgeschlagen: %v", err)
	}

	// Installiere Fyne
	log.Println("Installiere Fyne...")
	cmd = exec.Command("go", "get", "fyne.io/fyne/v2")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fyne installation fehlgeschlagen: %v", err)
	}

	// Erstelle main.go
	mainContent := `package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Hello")
	myWindow.Resize(fyne.NewSize(600, 400))

	hello := widget.NewLabel("Hello Fyne!")
	content := container.New(layout.NewVBoxLayout(),
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	)
	leftAligned := container.New(layout.NewHBoxLayout(), content, layout.NewSpacer())

	myWindow.SetContent(leftAligned)
	myWindow.ShowAndRun()
}`

	if err := os.WriteFile("main.go", []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("main.go erstellen fehlgeschlagen: %v", err)
	}

	// Führe go mod tidy aus
	log.Println("Führe go mod tidy aus...")
	cmd = exec.Command("go", "mod", "tidy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy fehlgeschlagen: %v", err)
	}

	// Öffne Terminal
	log.Println("Öffne Terminal...")
	terminalCmd := fmt.Sprintf("wezterm start --cwd '%s' --always-new-process -- bash -c 'echo \"go run .\" && bash'", projectDir)
	cmd = exec.Command("sh", "-c", terminalCmd)
	if err := cmd.Start(); err != nil {
		log.Printf("Terminal öffnen fehlgeschlagen: %v", err)
	}

	// Warte kurz, damit das Terminal Zeit hat zu starten
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ps *ProjectSetup) createRustProject() error {
	log.Println("Erstelle Rust-Projekt...")

	// Wechsel ins Elternverzeichnis
	if err := os.Chdir(ps.parentPath); err != nil {
		return fmt.Errorf("verzeichniswechsel fehlgeschlagen: %v", err)
	}

	// Erstelle neues Cargo-Projekt
	log.Println("Erstelle Cargo-Projekt...")
	cmd := exec.Command("cargo", "new", ps.projectName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cargo new fehlgeschlagen: %v", err)
	}

	projectDir := filepath.Join(ps.parentPath, ps.projectName)
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("verzeichniswechsel fehlgeschlagen: %v", err)
	}

	// Füge Druid hinzu
	log.Println("Füge Druid hinzu...")
	cmd = exec.Command("cargo", "add", "druid")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("druid installation fehlgeschlagen: %v", err)
	}

	// Erstelle main.rs
	mainContent := `use druid::widget::{Button, Flex, Label};
use druid::{AppLauncher, LocalizedString, PlatformError, Widget, WidgetExt, WindowDesc};

fn main() -> Result<(), PlatformError> {
	let main_window = WindowDesc::new(ui_builder())
		.title(LocalizedString::new("Rust Druid App"))
		.window_size((300.0, 200.0));

	let data = 0_u32;

	AppLauncher::with_window(main_window)
		.log_to_console()
		.launch(data)
}

fn ui_builder() -> impl Widget<u32> {
	let label = Label::new(|data: &u32, _env: &_| format!("Count: {}", data))
		.padding(5.0)
		.center();

	let button = Button::new("Increment")
		.on_click(|_ctx, data: &mut u32, _env| *data += 1)
		.padding(5.0);

	Flex::column().with_child(label).with_child(button)
}`

	if err := os.WriteFile("src/main.rs", []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("main.rs erstellen fehlgeschlagen: %v", err)
	}

	// Öffne Terminal
	log.Println("Öffne Terminal...")
	terminalCmd := fmt.Sprintf("wezterm start --cwd '%s' --always-new-process -- bash -c 'echo \"cargo run\" && bash'", projectDir)
	cmd = exec.Command("sh", "-c", terminalCmd)
	if err := cmd.Start(); err != nil {
		log.Printf("Terminal öffnen fehlgeschlagen: %v", err)
	}

	// Warte kurz, damit das Terminal Zeit hat zu starten
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ps *ProjectSetup) createJavaScriptProject() error {
	log.Println("Erstelle JavaScript-Projekt...")
	projectDir := filepath.Join(ps.parentPath, ps.projectName)

	// Erstelle Projektverzeichnis
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("projektverzeichnis erstellen fehlgeschlagen: %v", err)
	}

	// Wechsel ins Projektverzeichnis
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("verzeichniswechsel fehlgeschlagen: %v", err)
	}

	// Initialisiere npm
	cmd := exec.Command("npm", "init", "-y")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm init fehlgeschlagen: %v", err)
	}

	// Installiere Express
	cmd = exec.Command("npm", "install", "express")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("express installation fehlgeschlagen: %v", err)
	}

	// Erstelle app.js
	appContent := `const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
    res.send('Hello World!');
});

app.listen(port, () => {
    console.log('Server running at http://localhost:' + port);
});`

	if err := os.WriteFile("app.js", []byte(appContent), 0644); err != nil {
		return fmt.Errorf("app.js erstellen fehlgeschlagen: %v", err)
	}

	return ps.openTerminal(projectDir, "node app.js")
}

func (ps *ProjectSetup) createTypeScriptProject() error {
	log.Println("Erstelle TypeScript-Projekt...")
	projectDir := filepath.Join(ps.parentPath, ps.projectName)

	// Erstelle Projektverzeichnis
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("projektverzeichnis erstellen fehlgeschlagen: %v", err)
	}

	// Wechsel ins Projektverzeichnis
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("verzeichniswechsel fehlgeschlagen: %v", err)
	}

	// Initialisiere npm und installiere TypeScript
	commands := [][]string{
		{"npm", "init", "-y"},
		{"npm", "install", "typescript", "@types/node", "--save-dev"},
		{"npx", "tsc", "--init"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("befehl fehlgeschlagen %v: %v", args, err)
		}
	}

	// Erstelle src/index.ts
	if err := os.MkdirAll("src", 0755); err != nil {
		return fmt.Errorf("src verzeichnis erstellen fehlgeschlagen: %v", err)
	}

	indexContent := `class Greeter {
    constructor(private name: string) {}

    greet(): void {
        console.log("Hello, " + this.name + "!");
    }
}

const greeter = new Greeter("World");
greeter.greet();`

	if err := os.WriteFile("src/index.ts", []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("index.ts erstellen fehlgeschlagen: %v", err)
	}

	return ps.openTerminal(projectDir, "npx tsc && node dist/index.js")
}

func (ps *ProjectSetup) createCPlusPlusProject() error {
	log.Println("Erstelle C++-Projekt...")
	projectDir := filepath.Join(ps.parentPath, ps.projectName)

	// Erstelle Projektstruktur
	dirs := []string{
		"src",
		"include",
		"build",
	}

	// Erstelle Projektverzeichnis und Unterverzeichnisse
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("projektverzeichnis erstellen fehlgeschlagen: %v", err)
	}
	for _, dir := range dirs {
		path := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("verzeichnis %s erstellen fehlgeschlagen: %v", dir, err)
		}
	}

	// Erstelle CMakeLists.txt
	cmakeContent := fmt.Sprintf(`cmake_minimum_required(VERSION 3.10)
project(%s)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_executable(${PROJECT_NAME} src/main.cpp)
target_include_directories(${PROJECT_NAME} PRIVATE include)`, ps.projectName)

	if err := os.WriteFile(filepath.Join(projectDir, "CMakeLists.txt"), []byte(cmakeContent), 0644); err != nil {
		return fmt.Errorf("CMakeLists.txt erstellen fehlgeschlagen: %v", err)
	}

	// Erstelle main.cpp
	mainContent := `#include <iostream>

int main() {
    std::cout << "Hello, C++17!" << std::endl;
    return 0;
}`

	if err := os.WriteFile(filepath.Join(projectDir, "src", "main.cpp"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("main.cpp erstellen fehlgeschlagen: %v", err)
	}

	// Öffne Terminal
	terminalCmd := fmt.Sprintf("wezterm start --cwd '%s' --always-new-process -- bash -c 'cd build && cmake .. && make && echo \"Build abgeschlossen.\" && bash'", projectDir)
	cmd := exec.Command("sh", "-c", terminalCmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("terminal öffnen fehlgeschlagen: %v", err)
	}

	return nil
}

func (ps *ProjectSetup) createCSharpProject() error {
	log.Println("Erstelle C#-Projekt...")
	projectDir := filepath.Join(ps.parentPath, ps.projectName)

	// Erstelle neues .NET Projekt
	cmd := exec.Command("dotnet", "new", "console", "-n", ps.projectName)
	cmd.Dir = ps.parentPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dotnet new fehlgeschlagen: %v", err)
	}

	return ps.openTerminal(projectDir, "dotnet run")
}

func (ps *ProjectSetup) createJavaProject() error {
	log.Println("Erstelle Java-Projekt...")
	projectDir := filepath.Join(ps.parentPath, ps.projectName)

	// Erstelle Projektstruktur
	srcPath := filepath.Join(projectDir, "src", "main", "java")
	if err := os.MkdirAll(srcPath, 0755); err != nil {
		return fmt.Errorf("verzeichnis struktur erstellen fehlgeschlagen: %v", err)
	}

	// Erstelle Main.java
	mainContent := `public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, Java!");
    }
}`

	if err := os.WriteFile(filepath.Join(srcPath, "Main.java"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("Main.java erstellen fehlgeschlagen: %v", err)
	}

	return ps.openTerminal(projectDir, "javac src/main/java/Main.java && java -cp src/main/java Main")
}

// Hilfsfunktion für das Öffnen des Terminals
func (ps *ProjectSetup) openTerminal(dir string, command string) error {
	// Füge eine kleine Verzögerung hinzu
	time.Sleep(100 * time.Millisecond)

	terminalCmd := fmt.Sprintf("wezterm start --cwd '%s' --always-new-process -- bash -c 'echo \"Running: %s\"; %s; exec bash'",
		dir, command, command)

	cmd := exec.Command("sh", "-c", terminalCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("terminal öffnen fehlgeschlagen: %v", err)
	}

	// Warte kurz nach dem Öffnen
	time.Sleep(100 * time.Millisecond)
	return nil
}

func isValidProjectName(name string) (bool, string) {
	if name == "" {
		return false, "Projektname darf nicht leer sein"
	}
	if len(name) > 255 {
		return false, "Projektname ist zu lang (max 255 Zeichen)"
	}
	if strings.ContainsAny(name, "/\\:*?\"<>|$%&# \t\n") {
		return false, "Projektname darf nur Buchstaben, Zahlen, Unterstriche und Bindestriche enthalten"
	}
	return true, ""
}

func main() {
	log.Println("Starte Anwendung...")
	myApp := app.New()

	// Korrigierte Icon-Setzung
	icon := fyne.NewStaticResource("icon", iconData)
	myApp.SetIcon(icon)

	window := myApp.NewWindow("Project Setup")

	ps := NewProjectSetup()
	ps.window = window

	// UI-Komponenten erstellen
	projectTypeRadio := widget.NewRadioGroup([]string{
		"Python",
		"Go",
		"Rust",
		"JavaScript",
		"TypeScript",
		"C++",
		"C#",
		"Java",
	}, func(value string) {
		switch value {
		case "Python":
			ps.projectType = Python
		case "Go":
			ps.projectType = Go
		case "Rust":
			ps.projectType = Rust
		case "JavaScript":
			ps.projectType = JavaScript
		case "TypeScript":
			ps.projectType = TypeScript
		case "C++":
			ps.projectType = CPlusPlus
		case "C#":
			ps.projectType = CSharp
		case "Java":
			ps.projectType = Java
		}
		log.Printf("Projekttyp gewählt: %s", value)
	})
	projectTypeRadio.SetSelected("Python")

	var parentPathBtn *widget.Button
	parentPathBtn = widget.NewButton(ps.parentPath, func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				log.Printf("Fehler bei Ordnerauswahl: %v", err)
				return
			}
			if uri == nil {
				return
			}
			ps.parentPath = uri.Path()
			parentPathBtn.SetText(ps.parentPath)
			if err := ps.saveProjectPath(); err != nil {
				log.Printf("Fehler beim Speichern des Pfads: %v", err)
			}
		}, window)
	})

	var createBtn *widget.Button
	projectNameEntry := widget.NewEntry()

	// Status-Label mit fester Breite
	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Container für Status mit fester Breite
	statusContainer := container.NewHBox(
		layout.NewSpacer(),
		container.NewPadded(statusLabel),
		layout.NewSpacer(),
	)

	// Funktion zum Aktualisieren der Statusmeldung
	updateStatus := func(msg string) {
		log.Printf("Zeige Nachricht: %s", msg)
		// Kürze die Nachricht auf maximal 50 Zeichen
		if len(msg) > 50 {
			msg = msg[:47] + "..."
		}
		statusLabel.SetText(msg)
	}

	// OnChanged-Handler für projectNameEntry
	projectNameEntry.OnChanged = func(value string) {
		ps.projectName = value
		valid, msg := isValidProjectName(value)
		if !valid {
			projectNameEntry.SetText(strings.Map(func(r rune) rune {
				if strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-", r) {
					return r
				}
				return -1
			}, value))
			updateStatus(msg)
			createBtn.Disable()
		} else {
			createBtn.Enable()
		}
	}

	// Setze feste Fensterbreite
	window.Resize(fyne.NewSize(500, 300))
	window.SetFixedSize(true)

	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	// Initialisiere createBtn
	createBtn = widget.NewButton("Create Project", func() {
		// Deaktiviere UI-Elemente
		createBtn.Disable()
		projectNameEntry.Disable()
		parentPathBtn.Disable()
		projectTypeRadio.Disable()
		progress.Show()

		// Starte Projekterstellung
		updateStatus("Erstelle Projekt...")
		go func() {
			if err := ps.createProject(); err != nil {
				log.Printf("Fehler bei Projekterstellung: %v", err)
				updateStatus("Fehler: " + err.Error())
				createBtn.Enable()
				projectNameEntry.Enable()
				parentPathBtn.Enable()
				projectTypeRadio.Enable()
				progress.Hide()
			} else {
				updateStatus("Projekt erfolgreich erstellt")
				os.Exit(0)
			}
		}()
	})

	// Layout erstellen
	content := container.NewVBox(
		widget.NewLabel("Project Setup"),
		container.NewHBox(layout.NewSpacer(), projectTypeRadio, layout.NewSpacer()),
		container.NewGridWithColumns(2,
			widget.NewLabel("Parent Path:"),
			parentPathBtn,
			widget.NewLabel("Project Name:"),
			projectNameEntry,
		),
		createBtn,
		progress,
		statusContainer, // Verwende den Container mit fester Höhe
	)

	window.SetContent(content)

	window.ShowAndRun()
}

// Hilfsfunktionen für Installationsprüfungen
func (ps *ProjectSetup) checkGoInstallation() error {
	cmd := exec.Command("go", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go ist nicht installiert: %v", err)
	}
	return nil
}

func (ps *ProjectSetup) checkRustInstallation() error {
	cmd := exec.Command("rustc", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rust ist nicht installiert: %v", err)
	}
	return nil
}

func (ps *ProjectSetup) validate() error {
	if valid, msg := isValidProjectName(ps.projectName); !valid {
		return fmt.Errorf(msg)
	}
	if _, err := os.Stat(ps.parentPath); os.IsNotExist(err) {
		return fmt.Errorf("ausgewählter pfad existiert nicht")
	}
	return nil
}

func (ps *ProjectSetup) showProjectPreview() string {
	return fmt.Sprintf(
		"Projektübersicht:\n"+
			"- Name: %s\n"+
			"- Typ: %s\n"+
			"- Pfad: %s\n"+
			"- Geschätzte Größe: ~%dMB",
		ps.projectName,
		[]string{"Python", "Go", "Rust"}[ps.projectType],
		filepath.Join(ps.parentPath, ps.projectName),
		ps.estimateProjectSize(),
	)
}

func (ps *ProjectSetup) estimateProjectSize() int {
	switch ps.projectType {
	case Python:
		return 50 // Python mit venv und Paketen
	case Go:
		return 30 // Go mit Fyne
	case Rust:
		return 100 // Rust mit Cargo und Druid
	}
	return 0
}

func (ps *ProjectSetup) checkDiskSpace() error {
	var stat unix.Statfs_t
	if err := unix.Statfs(ps.parentPath, &stat); err != nil {
		return fmt.Errorf("konnte speicherplatz nicht prüfen: %v", err)
	}

	// Verfügbarer Speicher in MB
	available := (stat.Bavail * uint64(stat.Bsize)) / 1024 / 1024
	needed := uint64(ps.estimateProjectSize())

	if available < needed {
		return fmt.Errorf("nicht genug speicherplatz. benötigt: %dMB, verfügbar: %dMB", needed, available)
	}
	return nil
}

func (ps *ProjectSetup) initGit() error {
	projectDir := filepath.Join(ps.parentPath, ps.projectName)
	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git-initialisierung fehlgeschlagen: %v", err)
	}

	// Erstelle initial commit
	commands := [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "Initial commit"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = projectDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git-befehl fehlgeschlagen: %v", err)
		}
	}
	return nil
}

func (ps *ProjectSetup) checkJavaScriptInstallation() error {
	cmd := exec.Command("node", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("node.js ist nicht installiert: %v", err)
	}
	return nil
}

func (ps *ProjectSetup) checkTypeScriptInstallation() error {
	cmd := exec.Command("tsc", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("typescript ist nicht installiert: %v", err)
	}
	return nil
}

func (ps *ProjectSetup) checkCPlusPlusInstallation() error {
	cmd := exec.Command("g++", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("g++ ist nicht installiert: %v", err)
	}
	return nil
}

func (ps *ProjectSetup) checkCSharpInstallation() error {
	cmd := exec.Command("dotnet", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(".NET SDK ist nicht installiert: %v", err)
	}
	return nil
}

func (ps *ProjectSetup) checkJavaInstallation() error {
	cmd := exec.Command("javac", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Java Development Kit ist nicht installiert: %v", err)
	}
	return nil
}
