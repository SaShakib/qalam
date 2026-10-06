package engine

type KeyKind int

const (
	KeyChar KeyKind = iota
	KeyBackspace
	KeyEscape
	KeyEnter
	KeySpace
	KeyOther // tab, arrows…: finish the word and let the key through
)

type Key struct {
	Kind KeyKind
	Char rune
}

// Action: Commit=false means "replace the preview with Text", Commit=true means "insert Text".
type Action struct {
	Commit bool
	Text   string
}

// Composer turns key presses into preview / commit actions, like the macOS one.
type Composer struct {
	Buffer string
}

// Handle returns whether the key was used up, plus what to do.
func (c *Composer) Handle(key Key, o Options) (bool, []Action) {
	switch key.Kind {
	case KeyBackspace:
		if c.Buffer == "" {
			return false, nil
		}
		r := []rune(c.Buffer)
		c.Buffer = string(r[:len(r)-1])
		return true, []Action{{Text: c.preview(o)}}
	case KeyEscape:
		if c.Buffer == "" {
			return false, nil
		}
		raw := c.Buffer
		c.Buffer = ""
		return true, []Action{{Commit: true, Text: raw}}
	case KeyEnter:
		if c.Buffer == "" {
			return false, nil
		}
		return true, c.Flush(o)
	case KeySpace, KeyOther:
		return false, c.Flush(o)
	}
	if Accepts(key.Char, c.Buffer) {
		c.Buffer += string(key.Char)
		return true, []Action{{Text: c.preview(o)}}
	}
	actions := c.Flush(o)
	if p := Punctuation(key.Char, o); p != "" {
		return true, append(actions, Action{Commit: true, Text: p})
	}
	return false, actions
}

// Flush commits the word being typed, if any.
func (c *Composer) Flush(o Options) []Action {
	if c.Buffer == "" {
		return nil
	}
	out := Word(c.Buffer, o)
	c.Buffer = ""
	return []Action{{Commit: true, Text: out}}
}

func (c *Composer) preview(o Options) string {
	if c.Buffer == "" {
		return ""
	}
	if a := Word(c.Buffer, o); a != "" {
		return a
	}
	return c.Buffer
}
