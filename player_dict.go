package main

import (
	"sort"
	"strings"
)

type Dict struct {
	ident  mrows
	binary mrows
}

func (d *Dict) Clone() *Dict {
	ident := make(mrows, len(d.ident))
	binary := make(mrows, len(d.binary))
	copy(ident, d.ident)
	copy(binary, d.binary)
	return &Dict{ident, binary}
}

func (d *Dict) Sort() {
	d.ident = d.ident.Sort()
	d.binary = d.binary.Sort()
}

type mrows []mrow

func (mr mrows) Sort() mrows {
	// Clean dups first
	var nr mrows
	exists := func(m2 mrow) bool {
	loop:
		for _, m := range nr {
			if len(m.match) == len(m2.match) {
				for j := range m.match {
					if m.match[j] != m2.match[j] {
						continue loop
					}
				}
				return true
			}
		}

		return false
	}
	for _, m := range mr {
		if exists(m) {
			continue
		}
		nr = append(nr, m)
	}

	// Sort
	f := func(i, j int) bool {
		if len(nr[i].match) != len(nr[j].match) {
			return len(nr[i].match) > len(nr[j].match)
		}
		for k := 0; k < len(nr[i].match); k++ {
			if nr[i].match[k] != "@" && nr[j].match[k] == "@" {
				return true
			}
			if nr[i].match[k] == "@" && nr[j].match[k] != "@" {
				return false
			}
		}

		return false
	}
	sort.Slice(nr, f)

	return nr
}

type mrow struct {
	target interface{}
	match  []string
}

func (d *Dict) AddBinary(target, v string) {
	s := strings.Split(v, " ")
	if len(s) > 0 && s[0] == "is" {
		d.binary = append(d.binary, mrow{
			target: &PExprBinary{Op: target, Neg: true},
			match:  append([]string{"is", "not"}, s[1:]...),
		})
		if len(s) > 1 {
			d.binary = append(d.binary, mrow{
				target: &PExprBinary{Op: target},
				match:  s[1:],
			})
		}
	}
	d.binary = append(d.binary, mrow{
		target: &PExprBinary{Op: target},
		match:  s,
	})
}

func (d *Dict) AddFuncStub(v ...string) {
	d.ident = append(d.ident, mrow{
		target: &PFunc{Name: strings.Join(v, " ")},
		match:  v,
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

	d.ident = append(d.ident, mrow{
		target: target,
		match:  m,
	})
}

func (d *Dict) Add(target interface{}, v string) {
	d.ident = append(d.ident, mrow{
		target: target,
		match:  strings.Split(v, " "),
	})
}

func (d *Dict) Add2(target interface{}, v ...string) {
	d.ident = append(d.ident, mrow{
		target: target,
		match:  v,
	})
}
