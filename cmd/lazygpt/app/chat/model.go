package chat

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/lazygpt/lazygpt/plugin/log"
)

type errMsg error

type tickMsg time.Time

func tickEvery() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func textRendererWithWidth(width int, otherOptions ...glamour.TermRendererOption) *glamour.TermRenderer {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
	)
	if err != nil {
		panic(err)
	}

	return renderer
}

type model struct {
	chatViewport viewport.Model
	logsViewport viewport.Model
	promptText   textarea.Model

	senderStyle   lipgloss.Style
	receiverStyle lipgloss.Style

	textRenderer *glamour.TermRenderer

	logBuf *bytes.Buffer
	err    error

	pctx *promptContext
}

func NewModel(pctx *promptContext) model {
	const width = 30

	promptTa := textarea.New()
	promptTa.Placeholder = "Send a message..."
	promptTa.Focus()

	promptTa.Prompt = "┃ "
	promptTa.CharLimit = 280

	promptTa.SetWidth(30)
	promptTa.SetHeight(3)

	// Remove cursor line styling
	promptTa.FocusedStyle.CursorLine = lipgloss.NewStyle()
	promptTa.ShowLineNumbers = false
	promptTa.KeyMap.InsertNewline.SetEnabled(false)

	chatVp := viewport.New(width, 10)
	chatVp.SetContent("Welcome to LazyGPT! Type a prompt and press Enter to send.")

	logsVp := viewport.New(width, 7)
	logsVp.SetContent("Logs will appear here.")

	// TODO(seanj): Need to make sure this buffer can't grow unbounded
	logBuf := bytes.NewBufferString("")
	log.ReplaceOutput(pctx.ctx, logBuf)

	return model{
		chatViewport: chatVp,
		logsViewport: logsVp,
		promptText:   promptTa,

		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		receiverStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),

		textRenderer: textRendererWithWidth(width),

		logBuf: logBuf,
		err:    nil,

		pctx: pctx,
	}
}

func (m model) Init() tea.Cmd {
	return tickEvery()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd  tea.Cmd
		cvpCmd tea.Cmd
		lvpCmd tea.Cmd
	)

	m.promptText, tiCmd = m.promptText.Update(msg)
	m.chatViewport, cvpCmd = m.chatViewport.Update(msg)
	m.logsViewport, lvpCmd = m.logsViewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.promptText.Value())

			return m, tea.Quit
		case tea.KeyEnter:
			promptText := m.promptText.Value()
			m.promptText.Reset()
			promptText = strings.TrimSpace(promptText)
			if strings.ToLower(promptText) == "quit" || strings.ToLower(promptText) == "exit" {
				return m, tea.Quit
			}

			message := m.pctx.AddUserMessage(promptText)
			return m, tea.Batch(doPromptUpdated(), m.pctx.Executor(message))
		}

	case tea.WindowSizeMsg:
		m.promptText.SetWidth(msg.Width)
		m.chatViewport.Width = msg.Width
		m.logsViewport.Width = msg.Width
		m.textRenderer = textRendererWithWidth(msg.Width)
		return m, nil

	case tickMsg:
		// Update logs viewport
		if m.logBuf.Len() > 0 {
			m.logsViewport = m.renderLogsViewport()
			m.logBuf.Reset()
		}

		return m, tickEvery()

	case promptUpdated:
		m.chatViewport = m.renderChatViewport()
		return m, nil

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, cvpCmd, lvpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n%s\n\n%s",
		m.logsViewport.View(),
		m.chatViewport.View(),
		m.promptText.View(),
	) + "\n\n"
}

func (m model) renderChatViewport() viewport.Model {
	messages := m.renderChatMessages()
	chatViewport := viewport.New(m.chatViewport.Width, m.chatViewport.Height)
	content, err := m.textRenderer.Render(strings.Join(messages, "\n"))
	if err != nil {
		content = fmt.Errorf("error rendering chat: %w", err).Error()
	}
	chatViewport.SetContent(content)
	chatViewport.GotoBottom()
	return chatViewport
}

func (m model) renderLogsViewport() viewport.Model {
	logsViewport := viewport.New(m.logsViewport.Width, m.logsViewport.Height)
	content, err := m.textRenderer.Render(m.logBuf.String())
	if err != nil {
		content = fmt.Errorf("error rendering logs: %w", err).Error()
	}
	logsViewport.SetContent(content)
	logsViewport.GotoBottom()
	return logsViewport
}

func (m model) renderChatMessages() []string {
	var messages []string = make([]string, 0)

	for _, msg := range *m.pctx.messages {
		var content string
		var err error

		switch msg.Role {
		case "user":
			content, err = m.textRenderer.Render(m.senderStyle.Render("You: ") + msg.Content)
			if err != nil {
				content = fmt.Errorf("error rendering You message: %w", err).Error()
			}
		case "assistant":
			content, err = m.textRenderer.Render(m.receiverStyle.Render("Assistant: ") + msg.Content)
			if err != nil {
				content = fmt.Errorf("error rendering Assistant message: %w", err).Error()
			}
		}

		messages = append(messages, content)
	}

	return messages
}
