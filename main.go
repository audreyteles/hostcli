package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"log"
	"os"
	"regexp"
)

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	inputs []textinput.Model
	cursor int
	err    error
}

// Define constants to each input
const (
	ip = iota // iota -> define a sequence of numbers 0, 1, 2...
	host
)

// Define colors
const (
	blue     = lipgloss.Color("#4287f5")
	darkGray = lipgloss.Color("#767676")
)

// Store labels colors variables
var (
	inputStyle    = lipgloss.NewStyle().Foreground(blue)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

func addHost(ip string, name string) {
	// Open local hosts file
	dat, err := os.ReadFile("/etc/hostcli/data.txt")

	if err != nil {
		log.Fatal(err)
	}

	// Convert to bytes the ip and the name
	newHost := []byte(fmt.Sprintf("\n%s %s", ip, name))

	// Append new host in old file content
	newContent := append(dat, newHost...)

	err = os.WriteFile("/etc/hostcli/data.txt", newContent, 0644)

	if err != nil {
		log.Fatal(err)
	}

	etcHosts, err := os.ReadFile("/etc/hosts")

	if err != nil {
		log.Fatal(err)
	}
	// Append hosts in actual /etc/hosts file content
	newEtcHosts := append(etcHosts, newContent...)

	err = os.WriteFile("/etc/hosts", newEtcHosts, 0644)

	if err != nil {
		log.Fatal(err)
	}

	// Success message
	fmt.Printf("New host (%s) has been added to the IP (%s)\n", name, ip)
}

func setup() {

	if os.Geteuid() != 0 {
		log.Fatalf("You may need to run with root privileges.")
	}

	_, err := os.Stat("/etc/hostcli/data.txt")

	if err != nil {
		log.Fatal(err)
	}

	_ = os.Mkdir("/etc/hostcli", 0644)

	localFile, err := os.Create("/etc/hostcli/data.txt")
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}

	_ = localFile.Close()

}

// IPv4 addres validator
func ipValidator(s string) error {
	match, _ := regexp.MatchString("\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b", s)

	if match == false {
		return fmt.Errorf("invalid IP address: %s", s)
	}
	return nil
}

func hostValidator(s string) error {
	match, _ := regexp.MatchString("[^\\w\\s\\.]", s)

	if match == false {
		return fmt.Errorf("your host can't be an URL: %s", s)
	}

	return nil
}

func initialModel() model {
	setup()
	var inputs []textinput.Model = make([]textinput.Model, 2)

	inputs[ip] = textinput.New()

	inputs[ip].Focus()
	inputs[ip].Width = 30
	inputs[ip].Prompt = ""
	inputs[ip].CharLimit = 20
	inputs[ip].Validate = ipValidator
	inputs[ip].Placeholder = "127.0.0.1"

	inputs[host] = textinput.New()

	inputs[host].Width = 30
	inputs[host].Prompt = ""
	inputs[host].Placeholder = "host.name"
	inputs[host].Validate = hostValidator

	return model{
		inputs: inputs,
		cursor: 0,
		err:    nil,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.inputs[m.cursor].Err == nil {
				m.nextInput()
			} else if m.cursor == len(m.inputs)-1 {
				// Add new host
				addHost(m.inputs[ip].Value(), m.inputs[host].Value())

				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab:
			m.prevInput()
		case tea.KeyTab:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.cursor].Focus()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	// Set inputs
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m model) Init() tea.Cmd {

	return textinput.Blink
}

func (m model) View() string {
	return fmt.Sprintf(
		`Add a new local host:

 %s
 %s

 %s
 %s

 %s
`,
		// Line 01
		inputStyle.Width(30).Render("IPv4 Address"),
		m.inputs[ip].View(),

		// Line 02
		inputStyle.Width(30).Render("Host"),
		m.inputs[host].View(),

		continueStyle.Render("Press [Enter] to add"),
	) + "\n"
}

// nextInput focuses the next input field
func (m *model) nextInput() {
	m.cursor = (m.cursor + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *model) prevInput() {
	m.cursor--
	// Wrap around
	if m.cursor < 0 {
		m.cursor = len(m.inputs) - 1
	}
}
