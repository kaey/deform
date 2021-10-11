package main

import (
	"sort"
	"strings"

	"go4.org/intern"
)

type Dict struct {
	ident  patterns
	binary patterns
}

func (d *Dict) Clone() *Dict {
	ident := make(patterns, len(d.ident))
	binary := make(patterns, len(d.binary))
	copy(ident, d.ident)
	copy(binary, d.binary)
	return &Dict{ident, binary}
}

func (d *Dict) Sort() {
	d.ident = d.ident.Sort()
	d.binary = d.binary.Sort()
}

type patterns []pattern

func (p patterns) Sort() patterns {
	// Clean dups first
	var np patterns
	exists := func(m2 pattern) bool {
	loop:
		for _, m := range np {
			if len(m.parts) == len(m2.parts) {
				for j := range m.parts {
					if m.parts[j] != m2.parts[j] {
						continue loop
					}
				}
				return true
			}
		}

		return false
	}
	for _, m := range p {
		if exists(m) {
			continue
		}
		np = append(np, m)
	}

	// Sort
	f := func(i, j int) bool {
		if len(np[i].parts) != len(np[j].parts) {
			return len(np[i].parts) > len(np[j].parts)
		}
		for k := 0; k < len(np[i].parts); k++ {
			if np[i].parts[k] != "@" && np[j].parts[k] == "@" {
				return true
			}
			if np[i].parts[k] == "@" && np[j].parts[k] != "@" {
				return false
			}
		}

		return false
	}
	sort.Slice(np, f)

	return np
}

type pattern struct {
	target interface{}
	parts  []string
	iparts []*intern.Value
}

func (d *Dict) AddBinary(target, v string) {
	s := strings.Split(v, " ")
	if len(s) > 0 && s[0] == "is" {
		ss := append([]string{"is", "not"}, s[1:]...)
		d.binary = append(d.binary, pattern{
			target: &PExprBinary{Op: target, Neg: true},
			parts:  ss,
			iparts: iparts(ss),
		})
		if len(s) > 1 {
			ss := s[1:]
			d.binary = append(d.binary, pattern{
				target: &PExprBinary{Op: target},
				parts:  ss,
				iparts: iparts(ss),
			})
		}
	}
	d.binary = append(d.binary, pattern{
		target: &PExprBinary{Op: target},
		parts:  s,
		iparts: iparts(s),
	})
}

func (d *Dict) AddFuncStub(v ...string) {
	d.ident = append(d.ident, pattern{
		target: &PFunc{Name: strings.Join(v, " ")},
		parts:  v,
		iparts: iparts(v),
	})
}

func (d *Dict) AddFunc(target interface{}, parts []FuncPart) {
	m := make([]string, len(parts))
	for i, p := range parts {
		if p.Word != "" {
			m[i] = p.Word
		} else {
			m[i] = "@"
		}
	}

	d.addFunc(target, m)
}

func (d *Dict) addFunc(target interface{}, m []string) {
	for i, p := range m {
		if !strings.Contains(p, "/") {
			continue
		}
		for _, v := range strings.Split(p, "/") {
			m2 := make([]string, len(m))
			copy(m2, m)
			m2[i] = v
			d.addFunc(target, m2)
		}
		return
	}

	d.ident = append(d.ident, pattern{
		target: target,
		parts:  m,
		iparts: iparts(m),
	})
}

func (d *Dict) Add(target interface{}, v string) {
	s := strings.Split(v, " ")
	d.ident = append(d.ident, pattern{
		target: target,
		parts:  s,
		iparts: iparts(s),
	})
}

func (d *Dict) Add2(target interface{}, v ...string) {
	d.ident = append(d.ident, pattern{
		target: target,
		parts:  v,
		iparts: iparts(v),
	})
}

func iparts(s []string) []*intern.Value {
	isl := make([]*intern.Value, len(s))
	for i := range s {
		isl[i] = intern.GetByString(strings.ToLower(s[i]))
	}
	return isl
}
