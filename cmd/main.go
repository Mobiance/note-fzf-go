package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	notesDir = "E:/Notes" // Change to your preferred notes directory
	editor   = "nvim"     // Assumes 'nvim' is in your PATH
	fzf      = "fzf"      // Assumes 'fzf' is in your PATH
)

// Ensure the notes directory exists
func ensureNotesDir() {
	if _, err := os.Stat(notesDir); os.IsNotExist(err) {
		err := os.MkdirAll(notesDir, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating notes directory:", err)
			os.Exit(1)
		}
		fmt.Println("Created notes directory:", notesDir)
	}
}

// Create a new note
func newNote() {
	fmt.Print("Enter the note title: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	title := scanner.Text()

	date := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("%s_%s.md", date, title)
	filePath := filepath.Join(notesDir, fileName)

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		fmt.Println("Note already exists:", fileName)
	} else {
		content := fmt.Sprintf("# %s\n\nCreated: %s\n", title, time.Now().Format("2006-01-02 15:04:05"))
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			fmt.Println("Error creating new note:", err)
			return
		}
		fmt.Println("Created new note:", fileName)
	}

	openInEditor(filePath)
}

// Open a file in the editor
func openInEditor(filePath string) {
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error opening note:", err)
	}
}

// List all notes in the directory
func getNotes() ([]string, error) {
	files, err := os.ReadDir(notesDir)
	if err != nil {
		return nil, err
	}
	var notes []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			notes = append(notes, file.Name())
		}
	}
	return notes, nil
}

// Search and open notes using fzf
func searchNotesWithFzf() {
	notes, err := getNotes()
	if err != nil {
		fmt.Println("Error retrieving notes:", err)
		return
	}

	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return
	}

	cmd := exec.Command(fzf, "--multi", "--preview", fmt.Sprintf("bat --style=plain --color=always %s/{}", notesDir), "--preview-window=right:70%")
	cmd.Stdin = strings.NewReader(strings.Join(notes, "\n"))
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error with fzf:", err)
		return
	}

	selectedNote := strings.TrimSpace(string(output))
	if selectedNote != "" {
		filePath := filepath.Join(notesDir, selectedNote)
		openInEditor(filePath)
	}
}

// Delete a note
func deleteNote() {
	notes, err := getNotes()
	if err != nil {
		fmt.Println("Error retrieving notes:", err)
		return
	}

	if len(notes) == 0 {
		fmt.Println("No notes found.")
		return
	}

	cmd := exec.Command(fzf, "--prompt=Select a note to delete: ", "--multi")
	cmd.Stdin = strings.NewReader(strings.Join(notes, "\n"))
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error with fzf:", err)
		return
	}

	selectedNote := strings.TrimSpace(string(output))
	if selectedNote != "" {
		filePath := filepath.Join(notesDir, selectedNote)

		// Confirm deletion
		fmt.Printf("Are you sure you want to delete %s? (y/n): ", selectedNote)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		confirm := scanner.Text()

		if confirm == "y" || confirm == "Y" {
			err := os.Remove(filePath)
			if err != nil {
				fmt.Println("Error deleting note:", err)
				return
			}
			fmt.Println("Deleted note:", selectedNote)
		} else {
			fmt.Println("Deletion cancelled.")
		}
	}
}

// Display the main menu using fzf
func showFzfMenu() (bool, error) {
	options := []string{
		"Create a new note",
		"Search and open notes",
		"Delete a note",
		"Exit",
	}

	cmd := exec.Command(fzf, "--prompt=Select an option: ", "--height=10")
	cmd.Stdin = strings.NewReader(strings.Join(options, "\n"))
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	choice := strings.TrimSpace(string(output))
	switch choice {
	case "Create a new note":
		newNote()
	case "Search and open notes":
		searchNotesWithFzf()
	case "Delete a note":
		deleteNote()
	case "Exit":
		return false, nil
	default:
		fmt.Println("Invalid selection")
	}

	return true, nil
}

func main() {
	ensureNotesDir()

	for {
		continueLoop, err := showFzfMenu()
		if err != nil {
			fmt.Println("Error displaying menu:", err)
			break
		}
		if !continueLoop {
			break
		}
	}
}
