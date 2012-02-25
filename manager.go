package main

/* Matric number: U096996N

This program is written in Go, Google's system language.
*/

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"
	//"flag"
	"io/ioutil"
	"container/list"
)

var (
	Ready_List    = list.New()
	Resource_List = list.New()
	IO            = list.New()
)

// current running process
var Curr, Init *PCB

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
	PID             string
	Other_Resources *list.List
	Creation_Tree   CT
	Status          Stat
	Priority        int
}

type RCB struct {
	RID          string
	Status       string
	Waiting_List *list.List
}

type IO_RCB struct {
	Waiting_List *list.List
}

// Operations on processes
func (p *PCB) Init() *PCB {
	p.PID = ""
	p.Other_Resources = list.New()
	p.Creation_Tree = CT{p, list.New()}
	p.Status = Stat{"ready_s", Ready_List}
	p.Priority = 0
	return p
}

// create new process
func (p *PCB) Create(name string, priority int) {
	newP := PCB{name,
		list.New(),
		CT{p, list.New()},
		Stat{"ready_s", Ready_List},
		priority}

	listInsert(&newP, p.Creation_Tree.Child)
	listRLInsert(&newP, Ready_List)
	Scheduler()
}

// destroy processes
func (p *PCB) Destroy(pid string) {
	pcb := getPCB(pid)
	killTree(pcb)
	Scheduler()
}

// suspend process
func (p *PCB) Suspend(pid string) {
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
func (p *PCB) Activate(pid string) {
	pcb := getPCB(pid)
	if pcb.Status.Type == "ready_s" {
		pcb.Status.Type = "ready_a"
		Scheduler()
	} else {
		pcb.Status.Type = "blocked_a"
	}
}

// kill creation_tree for given PCB
func killTree(p *PCB) {
	for e := p.Creation_Tree.Child.Front(); e != nil; e = e.Next() {
		killTree(e.Value.(*PCB))
	}
	listRemove(p, p.Status.List)
}

func (p *PCB) Request(rid string) {
	r := getRCB(rid)
	if r.Status == "free" {
		r.Status = "allocated"
		p.Other_Resources.PushFront(r) // or pushback?
	} else {
		p.Status.Type = "blocked_a"
		p.Status.List.PushFront(r) // warning, watch this
		listRLRemove(p, Ready_List)
		listInsert(p, r.Waiting_List)
	}
	Scheduler()
}

func (p *PCB) Release(rid string) {
	r := getRCB(rid)
	rcbListRemove(r, p.Other_Resources)
	if r.Waiting_List.Len() == 0 {
		r.Status = "free"
	} else {
		r.Waiting_List.Remove(r.Waiting_List.Front())
		p.Status.Type = "ready_a"
		p.Status.List = Ready_List
		listRLInsert(p, Ready_List)
	}
	Scheduler()
}

// scheduler
func Scheduler() {
	p := maxPriorityPCB()
	fmt.Println(p.PID)
	if Curr.Priority < p.Priority || Curr.Status.Type != "running" || Curr == nil {
		preempt(p, Curr)
		fmt.Printf("Process %s is running\n", p.PID)
	}
}

func preempt(p, curr *PCB) {
	if curr == nil {
		Curr = p
	}
	p.Status.Type = "running"
	fmt.Printf("Process %s is running\n", p.PID)
}

// find and return the highest priority PCB
func maxPriorityPCB() *PCB {
	system := Ready_List.Front()
	user := system.Next()
	init := user.Next()

	switch {
	case system.Value.(*list.List).Len() != 0:
		return system.Value.(*list.List).Front().Value.(*PCB)
	case user.Value.(*list.List).Len() != 0:
		return user.Value.(*list.List).Front().Value.(*PCB)
	case init.Value.(*list.List).Len() > 1:
		return init.Value.(*list.List).Front().Value.(*PCB)
	}
	return getPCB("init") // return init
}

func (p *PCB) Request_IO() {
	p.Status.Type = "blocked_a"
	p.Status.List = IO
	listRLRemove(p, Ready_List)

	iowl := IO.Front().Value.(IO_RCB).Waiting_List
	listInsert(p, iowl)
	Scheduler()
}

func (p *PCB) IO_completion() {
	listRemove(p, IO.Front().Value.(IO_RCB).Waiting_List)
	p.Status.Type = "ready"
	p.Status.List = Ready_List
	listRLInsert(p, Ready_List)
	Scheduler()
}

func getRCB(rid string) *RCB {
	for e := Resource_List.Front(); e != nil; e = e.Next() {
		if e.Value.(RCB).RID == rid {
			return e.Value.(*RCB)
		}
	}
	return nil
}

// get PCB based on pid by recursing through
// all children of Init
func getPCB(pid string) *PCB {
	ct := getPCB("init").Creation_Tree.Child
	for e := ct.Front(); e != nil; e = e.Next() {
		if e.Value.(PCB).PID == pid {
			return e.Value.(*PCB)
		}
	}
	return getChildPCB(ct, pid)
}

func getChildPCB(ls *list.List, pid string) *PCB {
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

func listRLInsert(p *PCB, rl *list.List) {
	pr := p.Priority
	var e *list.Element

	switch {
	case pr == 0:
		e = rl.Front()
	case pr == 1:
		e = rl.Front().Next()
	case pr == 2:
		e = rl.Front().Next().Next()
	}
	ls := e.Value.(*list.List)
	ls.PushFront(p)
}

func listRLRemove(p *PCB, rl *list.List) {
	pr := p.Priority
	var e *list.Element

	switch {
	case pr == 0:
		e = rl.Front()
	case pr == 1:
		e = rl.Front().Next()
	case pr == 2:
		e = rl.Front().Next().Next()
	}
	ls := e.Value.(*list.List)
	listRemove(p, ls)
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

// parser for commands
func Manager(cmd string) {
	cmds := strings.Split(cmd, " ")

	switch ins := cmds[0]; {
		case ins == "cr" && len(cmds) == 3:
			x, err := strconv.Atoi(cmds[2])
			if err == nil {
				Curr.Create(cmds[1], x)
			}
		case ins == "de" && len(cmds) == 2:
			Curr.Destroy(cmds[1])
		case ins == "req" && len(cmds) == 2:
			Curr.Request(cmds[1])
		case ins == "rel" && len(cmds) == 2:
			Curr.Release(cmds[1])
		case ins == "rio" && len(cmds) == 1:
			Curr.Request_IO()
		case ins == "ioc" && len(cmds) == 1:
			Curr.IO_completion()
		default:
			fmt.Println("Unknown command")
	}
}

func read(title string) string {
	filename := title
	body, err := ioutil.ReadFile(filename + ".txt")
	if err != nil {
		panic(err)
	}
	return string(body)
}

// setup all the data structures
func initialize() {
	Init = *PCB{
		"init",
		list.New(),
		CT{nil, list.New()},
		Stat{"ready_s", Ready_List},
		0}
	Curr = Init

	Ready_List.PushFront(list.New())
	Ready_List.PushFront(list.New())
	Ready_List.PushFront(list.New())

	listRLInsert(Init, Ready_List)

	Resource_List.PushFront(&RCB{"R1", "free", list.New()})
	Resource_List.PushFront(&RCB{"R2", "free", list.New()})
	Resource_List.PushFront(&RCB{"R3", "free", list.New()})

	IO.PushFront(IO_RCB{list.New()})
}

func main() {
	//flag.Parse()
	i := ""
	var err os.Error
	in := bufio.NewReader(os.Stdin)

	// initialize process
	initialize()

	for {
		i, err = in.ReadString('\n')
		i = strings.TrimSpace(i)
		if err != nil {
			fmt.Println("Read error: ", err)
		}
		if i == "quit" && len(strings.Split(i, " ")) == 1 {
			fmt.Println("Exiting")
			break
		}
		Manager(i)
	}
	/*if *terminal {
		// REPL mode

	} else {
		// read file mode
	}*/
}
