package main

import "strings"

type Dict struct {
	rows []mrow
}

type mrow struct {
	target interface{}
	n      []mnode
}

type mnode struct {
	match string
	kind  bool
}

func (d *Dict) AddFuncStub(v string) {
	parts := strings.Split(v, " ")
	n := make([]mnode, len(parts))
	for i, p := range parts {
		n[i] = mnode{match: p}
	}

	d.rows = append(d.rows, mrow{
		target: &PFunc{Name: v},
		n:      n,
	})
}

func (d *Dict) AddFunc(target interface{}, parts []FuncPart) {
	n := make([]mnode, len(parts))
	for i, p := range parts {
		if p.Word != "" {
			n[i] = mnode{match: p.Word}
		} else {
			n[i] = mnode{match: p.ArgKind, kind: true}
		}
	}

	d.rows = append(d.rows, mrow{
		target: target,
		n:      n,
	})
}

func (d *Dict) Add(target interface{}, v string) {
	parts := strings.Split(v, " ")
	n := make([]mnode, len(parts))
	for i, p := range parts {
		n[i] = mnode{match: p}
	}

	d.rows = append(d.rows, mrow{
		target: target,
		n:      n,
	})
}

// TODO: sort
