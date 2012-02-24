package main

/* Matric number: U096996N

This program is written in Go, Google's system language.
*/

import (
	"fmt"
	//"flag"
	"io/ioutil"
	"container/list"
)

var (
	GPID          = 0
	Init          = PCB{0, list.New(), CT{nil, list.New()}, Stat{"ready_s", Ready_List}, 0}
	Ready_List    = list.New()
	Resource_List = list.New()
	IO            = list.New()
)

/*// Command flag
var terminal = flag.Bool("t", false, "use terminal mode for input")
*/

// Structs
type Stat struct {
	Type string
	List *list.List
}

type CT struct {
	Parent *PCB
	Child  *list.List
}

type PCB struct {
	PID             int
	Other_Resources *list.List
	Creation_Tree   CT
	Status          Stat
	Priority        int
}

type RCB struct {
	RID          int
	Status       string
	Waiting_List *list.List
}

type IO_RCB struct {
	Waiting_List *list.List
}

// Operations on processes

// create new process
func (p *PCB) Create(priority int) {
	newP := PCB{newPID(),
		list.New(),
		CT{p, list.New()},
		Stat{"ready_s", Ready_List},
		priority}

	listInsert(&newP, p.Creation_Tree.Child)
	listInsert(&newP, Ready_List)
	Scheduler()
}

// suspend process
func (p *PCB) Suspend(pid int) {
	pcb := getPCB(pid)
	s := pcb.Status.Type
	if s == "blocked_a" || s == "blocked_s" {
		pcb.Status.Type = "blocked_s"
	} else {
		pcb.Status.Type = "ready_s"
	}
	Scheduler()
}

// activate process
func (p *PCB) Activate(pid int) {
	pcb := getPCB(pid)
	if pcb.Status.Type == "ready_s" {
		pcb.Status.Type = "ready_a"
		Scheduler()
	} else {
		pcb.Status.Type = "blocked_a"
	}
}

// destroy processes
func (p *PCB) Destroy(pid int) {
	pcb := getPCB(pid)
	killTree(pcb)
	Scheduler()
}

// kill creation_tree for given PCB
func killTree(p *PCB) {
	for e := p.Creation_Tree.Child.Front(); e != nil; e = e.Next() {
		killTree(e.Value.(*PCB))
	}
	listRemove(p, p.Status.List)
}

func (p *PCB) Request(rid int) {
	r := getRCB(rid)
	if r.Status == "free" {
		r.Status = "allocated"
		p.Other_Resources.PushFront(r) // or pushback?
	} else {
		p.Status.Type = "blocked_a"
		p.Status.List.PushFront(r) // warning, watch this
		listRemove(p, Ready_List)
		listInsert(p, r.Waiting_List)
	}
	Scheduler()
}

func (p *PCB) Release(rid int) {
	r := getRCB(rid)
	rcbListRemove(r, p.Other_Resources)
	if r.Waiting_List.Len() == 0 {
		r.Status = "free"
	} else {
		r.Waiting_List.Remove(r.Waiting_List.Front())
		p.Status.Type = "ready_a"
		p.Status.List = Ready_List
	}
	Scheduler()
}

// scheduler
func Scheduler() {

}

func (p *PCB) Request_IO() {
	p.Status.Type = "blocked_a"
	p.Status.List = IO
	listRemove(p, Ready_List)

	iowl := IO.Front().Value.(IO_RCB).Waiting_List
	listInsert(p, iowl)
	Scheduler()
}

// returns a new PID from the global var GPID
func newPID() int {
	GPID += 1
	return GPID
}

func getRCB(rid int) *RCB {
	for e := Resource_List.Front(); e != nil; e = e.Next() {
		if e.Value.(RCB).RID == rid {
			return e.Value.(*RCB)
		}
	}
	return nil
}

// get PCB based on pid by recursing through
// all children of Init
func getPCB(pid int) *PCB {
	ct := Init.Creation_Tree.Child
	for e := ct.Front(); e != nil; e = e.Next() {
		if e.Value.(PCB).PID == pid {
			return e.Value.(*PCB)
		}
	}
	return getChildPCB(ct, pid)
}

// helper function for getPCB
func getChildPCB(ls *list.List, pid int) *PCB {
	if ls == nil {
		return nil
	}

	for e := ls.Front(); e != nil; e = e.Next() {
		if e.Value.(PCB).PID == pid {
			return e.Value.(*PCB)
		} else {
			res := getChildPCB(e.Value.(*PCB).Creation_Tree.Child, pid)
			if res != nil {
				return res
			}
		}
	}
	return nil
}

// removes PCB element from list
func listRemove(p *PCB, ls *list.List) {
	for e := ls.Front(); e != nil; e = e.Next() {
		if e.Value.(PCB).PID == p.PID {
			ls.Remove(e)
		}
	}
}

// removes RCB element
func rcbListRemove(r *RCB, ls *list.List) {
	for e := ls.Front(); e != nil; e = e.Next() {
		if e.Value.(RCB).RID == r.RID {
			ls.Remove(e)
		}
	}
}

// inserts process into list
func listInsert(p *PCB, ls *list.List) {
	ls.PushFront(p)
}

func read(title string) string {
	filename := title
	body, err := ioutil.ReadFile(filename + ".txt")
	if err != nil {
		panic(err)
	}
	return string(body)
}

func main() {
	//flag.Parse()
	in := ""

	/*if *terminal {
		// REPL mode

	} else {
		// read file mode
	}*/

	fmt.Println("hello", in)
}
