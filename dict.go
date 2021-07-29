package main

import "strings"

type Dict struct {
	m *mnode
}

type mnode struct {
	match    string
	mkind    bool
	target   interface{}
	children []*mnode
}

func (m *mnode) addNode(match string, mkind bool) *mnode {
	for _, c := range m.children {
		if c.match == match && c.mkind == mkind {
			return c
		}
	}

	c := &mnode{
		match: match,
		mkind: mkind,
	}
	m.children = append(m.children, c)
	return c
}

func (m *mnode) getNode(match string) *mnode {
	for _, c := range m.children {
		if c.match == match { // TODO: match kind
			return c
		}
	}

	return nil
}

func (d *Dict) AddStub(v string) {
	parts := strings.Split(v, " ")
	c := d.m
	for _, p := range parts {
		c = c.addNode(p, false)
	}
}

func (d *Dict) AddFunc(target interface{}, v []FuncPart) {
}
