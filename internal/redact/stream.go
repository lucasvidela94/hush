package redact

import (
	"bytes"
	"io"
)

// pendingLimit acota la línea incompleta retenida; pasado el límite se
// reenvía igual, guardando una cola que alcance para no cortar un secreto.
const pendingLimit = 1 << 20

// recentLimit es cuánto de lo ya reenviado se mira junto a cada línea nueva:
// alcanza para un secreto de windowMaxLen repartido de a un carácter por línea.
const recentLimit = 4 * windowMaxLen

// Stream redacta salida que llega por partes y la reenvía por línea, sin
// esperar al final. Cada línea se mira junto con las anteriores, así un
// secreto repartido entre líneas se tapa antes de que salga más que un
// fragmento menor a una ventana. Llamá a Flush al terminar. No es seguro
// para uso concurrente.
type Stream struct {
	w        io.Writer
	secrets  []string
	patterns [][]byte
	hold     int
	limit    int
	pending  []byte
	recent   []byte
}

func NewStream(w io.Writer, secrets []string) *Stream {
	long := longSecrets(secrets)
	patterns := literalPatterns(long)
	hold := 0
	for _, p := range patterns {
		hold = max(hold, len(p))
	}
	return &Stream{w: w, secrets: long, patterns: patterns, hold: hold, limit: pendingLimit}
}

func (s *Stream) Write(p []byte) (int, error) {
	if len(s.secrets) == 0 {
		return s.w.Write(p)
	}
	s.pending = append(s.pending, p...)
	for {
		i := bytes.IndexByte(s.pending, '\n')
		if i < 0 {
			break
		}
		if err := s.emit(s.pending[:i+1]); err != nil {
			return 0, err
		}
		s.pending = s.pending[i+1:]
	}
	if len(s.pending) > s.limit {
		cut := len(s.pending) - s.hold
		if err := s.emit(s.pending[:cut]); err != nil {
			return 0, err
		}
		s.pending = s.pending[cut:]
	}
	return len(p), nil
}

// Flush reenvía la última línea si quedó sin salto.
func (s *Stream) Flush() error {
	if len(s.pending) == 0 {
		return nil
	}
	err := s.emit(s.pending)
	s.pending = nil
	return err
}

func (s *Stream) emit(chunk []byte) error {
	window := append(bytes.Clone(s.recent), chunk...)
	out := apply(chunk, s.secrets, s.patterns)
	redacted := applyPatterns(window, s.patterns)
	if transformed(window, redacted, s.secrets) {
		out = []byte(nuke + "\n")
	}
	s.recent = bytes.Clone(window[max(0, len(window)-recentLimit):])
	_, err := s.w.Write(out)
	return err
}
