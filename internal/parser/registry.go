package parser

import (
	"errors"
	"fmt"
	"sort"
)

const minDetectConfidence = 60

// Registry holds registered issuer parsers.
type Registry struct {
	parsers []IssuerParser
}

func NewRegistry(parsers ...IssuerParser) *Registry {
	r := &Registry{parsers: make([]IssuerParser, 0, len(parsers))}
	for _, p := range parsers {
		r.Register(p)
	}
	return r
}

func (r *Registry) Register(p IssuerParser) {
	r.parsers = append(r.parsers, p)
}

func (r *Registry) List() []IssuerParser {
	out := make([]IssuerParser, len(r.parsers))
	copy(out, r.parsers)
	return out
}

type detectResult struct {
	parser     IssuerParser
	confidence int
}

// Detect picks the highest-confidence parser above threshold.
func (r *Registry) Detect(text string) (IssuerParser, error) {
	if len(r.parsers) == 0 {
		return nil, ErrIssuerUnknown
	}
	results := make([]detectResult, 0, len(r.parsers))
	for _, p := range r.parsers {
		if score := p.DetectConfidence(text); score > 0 {
			results = append(results, detectResult{parser: p, confidence: score})
		}
	}
	if len(results) == 0 {
		return nil, ErrIssuerUnknown
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].confidence > results[j].confidence
	})
	if results[0].confidence < minDetectConfidence {
		return nil, ErrIssuerUnknown
	}
	return results[0].parser, nil
}

// Parse runs detection then parsing.
func (r *Registry) Parse(text string, opts ParseOptions) (*NormalizedStatement, error) {
	p, err := r.Detect(text)
	if err != nil {
		return nil, err
	}
	stmt, err := p.Parse(text, opts)
	if stmt != nil && stmt.Issuer == "" {
		stmt.Issuer = p.IssuerID()
	}
	if err != nil {
		if errors.Is(err, ErrParseIncomplete) {
			return stmt, err
		}
		return nil, err
	}
	if stmt != nil && len(stmt.MissingRequired) > 0 {
		return stmt, fmt.Errorf("%w: %v", ErrParseIncomplete, stmt.MissingRequired)
	}
	return stmt, nil
}
