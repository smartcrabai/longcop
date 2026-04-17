package console

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var ErrEmptyFeatureRequest = errors.New("feature request cannot be empty")

type Terminal struct {
	requestReader io.Reader
	promptReader  *bufio.Reader
	out           io.Writer
	closePrompt   func() error
}

func Open(requestInput *os.File, out io.Writer) (*Terminal, error) {
	if requestInput == nil {
		return nil, errors.New("request input is required")
	}

	terminal, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/tty for interactive prompts: %w", err)
	}

	return &Terminal{
		requestReader: requestInput,
		promptReader:  bufio.NewReader(terminal),
		out:           out,
		closePrompt:   terminal.Close,
	}, nil
}

func New(requestReader io.Reader, promptReader io.Reader, out io.Writer) *Terminal {
	return &Terminal{
		requestReader: requestReader,
		promptReader:  bufio.NewReader(promptReader),
		out:           out,
	}
}

func (t *Terminal) Close() error {
	if t.closePrompt == nil {
		return nil
	}
	return t.closePrompt()
}

func (t *Terminal) ReadFeatureRequest() (string, error) {
	if _, err := fmt.Fprintln(t.out, "Describe the feature to implement. Press Ctrl+D when done:"); err != nil {
		return "", fmt.Errorf("write feature request prompt: %w", err)
	}

	body, err := io.ReadAll(t.requestReader)
	if err != nil {
		return "", fmt.Errorf("read feature request: %w", err)
	}

	request := strings.TrimSpace(string(body))
	if request == "" {
		return "", ErrEmptyFeatureRequest
	}

	return request, nil
}

func (t *Terminal) AskYesNo(question string, defaultYes bool) (bool, error) {
	suffix := "[Y/n]"
	if !defaultYes {
		suffix = "[y/N]"
	}

	for {
		if _, err := fmt.Fprintf(t.out, "%s %s ", question, suffix); err != nil {
			return false, fmt.Errorf("write yes/no prompt: %w", err)
		}

		answer, err := t.readLine()
		if err != nil {
			return false, err
		}

		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "":
			return defaultYes, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			if _, err := fmt.Fprintln(t.out, "Please answer with y or n."); err != nil {
				return false, fmt.Errorf("write validation message: %w", err)
			}
		}
	}
}

func (t *Terminal) AskChoice(question string, options []string) (int, error) {
	if _, err := fmt.Fprintln(t.out, question); err != nil {
		return 0, fmt.Errorf("write choice question: %w", err)
	}

	for index, option := range options {
		if _, err := fmt.Fprintf(t.out, "%d. %s\n", index+1, option); err != nil {
			return 0, fmt.Errorf("write choice option: %w", err)
		}
	}

	for {
		if _, err := fmt.Fprintf(t.out, "Select one [1-%d]: ", len(options)); err != nil {
			return 0, fmt.Errorf("write choice prompt: %w", err)
		}

		line, err := t.readLine()
		if err != nil {
			return 0, err
		}

		selected, err := strconv.Atoi(strings.TrimSpace(line))
		if err == nil && selected >= 1 && selected <= len(options) {
			return selected - 1, nil
		}

		if _, err := fmt.Fprintln(t.out, "Please enter one of the listed numbers."); err != nil {
			return 0, fmt.Errorf("write validation message: %w", err)
		}
	}
}

func (t *Terminal) readLine() (string, error) {
	if t.promptReader == nil {
		return "", errors.New("interactive prompt reader is not configured")
	}

	line, err := t.promptReader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return strings.TrimSpace(line), nil
		}
		return "", fmt.Errorf("read interactive input: %w", err)
	}

	return strings.TrimSpace(line), nil
}
