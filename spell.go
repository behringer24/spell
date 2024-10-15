package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/behringer24/argumentative"
	"github.com/bmaupin/go-epub"
	"github.com/russross/blackfriday/v2"
)

const (
	title       = "spell"
	description = "Smart Processing and Enhanced Lightweight Layout"
	version     = "v0.0.1"
)

var (
	inFileName  *string
	outFileName *string
	showHelp    *bool
	showVer     *bool
)

// Befehlstyp
type Command func(args string, baseDir string) (string, error)

// Struktur des Befehlsprozessors
type CommandProcessor struct {
	commands map[string]Command
}

// Initialisiert einen neuen Befehlsprozessor
func NewCommandProcessor() *CommandProcessor {
	return &CommandProcessor{commands: make(map[string]Command)}
}

// Fügt einen neuen Befehl hinzu
func (cp *CommandProcessor) AddCommand(name string, command Command) {
	cp.commands[name] = command
}

// Verarbeitet einen Text mit registrierten Befehlen
func (cp *CommandProcessor) Process(content string, baseDir string) (string, error) {
	commandRegex := regexp.MustCompile(`/([a-zA-Z]+)(.*?)`)

	// Funktion, die den Inhalt für ein gefundenes Kommando ersetzt
	return commandRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extrahiere den Befehl und seine Argumente
		matches := commandRegex.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match // Fallback: Wenn das Muster nicht korrekt ist
		}

		commandName := matches[1]
		commandArgs := matches[2]

		// Suche nach dem passenden Befehl
		if command, exists := cp.commands[commandName]; exists {
			result, err := command(commandArgs, baseDir)
			if err != nil {
				log.Printf("Fehler beim Ausführen des Befehls '%s': %v", commandName, err)
				return fmt.Sprintf("**Fehler beim Ausführen des Befehls '%s'**", commandName)
			}
			return result
		}
		return fmt.Sprintf("**Unbekannter Befehl: %s**", commandName)
	}), nil
}

// Funktion zum Einlesen einer Datei
func readFile(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Modul für Markdown zu HTML Konvertierung
func markdownToHTML(mdContent string) string {
	return string(blackfriday.Run([]byte(mdContent)))
}

// Funktion für das /include Kommando
func includeCommand(args string, baseDir string) (string, error) {
	filePath := filepath.Join(baseDir, args)

	// Lese den Inhalt der Datei ein
	includedContent, err := readFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Lesen der Datei '%s': %v", filePath, err)
	}

	// Verarbeite mögliche Kommandos in der referenzierten Datei
	processedContent, err := commandProcessor.Process(includedContent, baseDir)
	if err != nil {
		return "", fmt.Errorf("Fehler beim Verarbeiten von Befehlen in Datei '%s': %v", filePath, err)
	}

	return processedContent, nil
}

// Befehlsprozessor instanziieren
var commandProcessor = NewCommandProcessor()

// Initialisiert alle verfügbaren Befehle
func initCommands() {
	// Füge das /include Kommando hinzu
	commandProcessor.AddCommand("include", includeCommand)
}

// Funktion zum Hinzufügen eines Kapitels zur EPUB-Datei
func addChapter(e *epub.Epub, title, content string) error {
	// Erstelle eine HTML-Seite für das Kapitel
	htmlContent := markdownToHTML(content)
	chapterFileName := fmt.Sprintf("%s.xhtml", strings.ReplaceAll(title, " ", "_"))

	// Kapitel zur EPUB-Datei hinzufügen
	_, err := e.AddSection(htmlContent, title, chapterFileName, "")
	return err
}

// Funktion zum Verarbeiten einer Markdown-Datei
func processMarkdownFile(e *epub.Epub, filePath string) error {
	// Lese den Inhalt der Markdown-Datei ein
	content, err := readFile(filePath)
	if err != nil {
		return err
	}

	// Verarbeite Kommandos im Markdown-Inhalt
	baseDir := filepath.Dir(filePath)
	processedContent, err := commandProcessor.Process(content, baseDir)
	if err != nil {
		return err
	}

	// Kapitel hinzufügen
	title := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	err = addChapter(e, title, processedContent)
	if err != nil {
		return err
	}

	return nil
}

func parseArgs() {
	flags := &argumentative.Flags{}
	showHelp = flags.Flags().AddBool("help", "h", "Show this help text")
	showVer = flags.Flags().AddBool("version", "v", "Show version information")
	inFileName = flags.Flags().AddPositional("infile", true, "", "File to read from")
	outFileName = flags.Flags().AddPositional("outfile", false, "", "File to write to")

	err := flags.Parse(os.Args)
	if *showHelp {
		flags.Usage(title, description, nil)
		os.Exit(0)
	} else if *showVer {
		fmt.Print(title, " version: ", version)
		os.Exit(0)
	} else if err != nil {
		flags.Usage(title, description, err)
		os.Exit(1)
	}
}

func main() {
	parseArgs()

	if len(os.Args) < 3 {
		fmt.Println("Usage: md2epub <input-files>... <output-epub>")
		return
	}

	outputEpub := *outFileName
	inputFile := *inFileName

	// Befehle initialisieren
	initCommands()

	// Neue EPUB-Datei erstellen
	e := epub.NewEpub("Markdown to EPUB")

	// Verarbeite alle angegebenen Markdown-Dateien
	err := processMarkdownFile(e, inputFile)
	if err != nil {
		log.Fatalf("Fehler beim Verarbeiten der Datei '%s': %v", inputFile, err)
	}

	// EPUB speichern
	err = e.Write(outputEpub)
	if err != nil {
		log.Fatalf("Fehler beim Speichern der EPUB-Datei: %v", err)
	}

	fmt.Printf("EPUB-Datei '%s' erfolgreich erstellt!\n", outputEpub)
}
