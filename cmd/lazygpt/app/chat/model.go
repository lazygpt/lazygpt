package chat

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/lazygpt/lazygpt/plugin/log"
)

const (
	StrChatPlaceholder   = "Welcome to LazyGPT! Type a prompt and press Enter to send."
	StrLogsPlaceholder   = "Logs will appear here."
	StrPromptPlaceholder = "Type a prompt and press Enter to send."

	ViewChatSizeFactor   = 0.55
	ViewLogsSizeFactor   = 0.40
	ViewPromptSizeFactor = 0.05
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

func floorToInt(f float64) int {
	return int(math.Floor(f))
}

type model struct {
	chatViewport viewport.Model
	logsViewport viewport.Model
	promptText   textarea.Model

	senderStyle   lipgloss.Style
	receiverStyle lipgloss.Style

	textRenderer *glamour.TermRenderer

	windowWidth  int
	windowHeight int

	logBuf *bytes.Buffer
	err    error

	promptCtx      *promptContext
	promptExecutor promptExecutor
}

func NewModel(ctx context.Context, promptCtx *promptContext, promptExec promptExecutor) model {
	const height = 40
	const width = 80

	promptTa := textarea.New()
	promptTa.Placeholder = StrPromptPlaceholder
	promptTa.Focus()

	promptTa.Prompt = "┃ "
	promptTa.CharLimit = 280

	promptTa.SetWidth(width)
	promptTa.SetHeight(height * ViewPromptSizeFactor)

	// Remove cursor line styling
	promptTa.FocusedStyle.CursorLine = lipgloss.NewStyle()
	promptTa.ShowLineNumbers = false
	promptTa.KeyMap.InsertNewline.SetEnabled(false)

	chatVp := viewport.New(width, floorToInt(height*ViewChatSizeFactor))
	chatVp.SetContent(StrChatPlaceholder)
	chatVp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86"))

	logsVp := viewport.New(width, floorToInt(height*ViewLogsSizeFactor))
	logsVp.SetContent(StrLogsPlaceholder)
	logsVp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("202"))

	// TODO(seanj): Need to make sure this buffer can't grow unbounded
	// TODO(seanj): Create a BufferSinkAdapter that implements hclog.SinkAdapter
	// the buffer sink adapter should take a pointer to a bytes buffer
	logBuf := bytes.NewBufferString("")
	log.RegisterSink(ctx, log.NewBufferSinkAdapter(logBuf))
	log.ResetOutput(ctx, io.Discard)

	return model{
		chatViewport: chatVp,
		logsViewport: logsVp,
		promptText:   promptTa,

		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		receiverStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),

		textRenderer: textRendererWithWidth(width),

		windowWidth:  width,
		windowHeight: height,

		logBuf: logBuf,
		err:    nil,

		promptCtx:      promptCtx,
		promptExecutor: promptExec,
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
			switch strings.ToLower(promptText) {
			case "quit":
				fallthrough
			case "exit":
				return m, tea.Quit
			case "clear":
				return m, tea.ClearScreen
			}

			m.promptCtx.AddUserMessage(promptText)
			return m, tea.Batch(doPromptUpdated(), m.promptExecutor())
		}

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height

		m.textRenderer = textRendererWithWidth(msg.Width)
		m.promptText.SetWidth(msg.Width)
		m.chatViewport = m.renderChatViewport()
		m.logsViewport = m.renderLogsViewport()

		return m, nil

	case tickMsg:
		// Update logs viewport
		m.logsViewport = m.renderLogsViewport()
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
		"%s\n%s\n%s\n",
		m.logsViewport.View(),
		m.chatViewport.View(),
		m.promptText.View(),
	)
}

func (m model) renderChatViewport() viewport.Model {
	messages := m.renderChatMessages()

	chatViewport := viewport.New(m.windowWidth, floorToInt(float64(m.windowHeight)*ViewChatSizeFactor))
	chatViewport.Style = m.chatViewport.Style.Copy()

	content, err := m.textRenderer.Render(strings.Join(messages, "\n"))
	if err != nil {
		content = fmt.Errorf("error rendering chat: %w", err).Error()
	}
	if content == "" {
		content = StrChatPlaceholder
	}
	chatViewport.SetContent(content)
	chatViewport.GotoBottom()
	return chatViewport
}

func (m model) renderLogsViewport() viewport.Model {
	logsViewport := viewport.New(m.windowWidth, floorToInt(float64(m.windowHeight)*ViewLogsSizeFactor))
	logsViewport.Style = m.logsViewport.Style.Copy()

	content, err := m.textRenderer.Render(m.logBuf.String())
	if err != nil {
		content = fmt.Errorf("error rendering logs: %w", err).Error()
	}
	if content == "" {
		content = StrLogsPlaceholder
	}

	logsViewport.SetContent(content)
	logsViewport.GotoBottom()

	return logsViewport
}

func (m model) renderChatMessages() []string {
	var messages []string = make([]string, 0)

	for _, msg := range *m.promptCtx.messages {
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
