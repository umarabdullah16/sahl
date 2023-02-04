package main

import (
	"fmt"
	"math/rand"
	"os"
)

const ADD = 0
const SUB = 1
const MUL = 2
const DIV = 3
const MOD = 4
const NEG = 5
const NOT = 6
const AND = 7
const OR = 8
const EQUAL = 9
const NOT_EQUAL = 10
const LESS = 11
const LESS_EQUAL = 12
const GREATER = 13
const GREATER_EQUAL = 14
const TRUE = 15
const FALSE = 16
const JUMP = 17
const JUMP_IF_FALSE = 18
const STORE = 19
const INDEX = 20
const APPEND = 21
const LENGTH = 22
const LIST = 23
const CONST_U64 = 24
const CONST_U32 = 25
const CONST_U8 = 26
const STRING = 27
const DEF_LOCAL = 28
const GET_LOCAL = 29
const ASSIGN = 30
const CALL = 31
const RETURN = 32
const PRINT = 33
const POP = 34
const MAKE_TUPLE = 36

type Code struct {
	Strings   []string
	Functions []Function
	Start     int
}

type Function struct {
	Instructions []uint8
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ReadInt32(buffer []uint8, offset int) int {
	return int(buffer[offset]) | int(buffer[offset+1])<<8 | int(buffer[offset+2])<<16 | int(buffer[offset+3])<<24
}

func ReadInt64(buffer []uint8, offset int) int64 {
	return int64(buffer[offset]) | int64(buffer[offset+1])<<8 | int64(buffer[offset+2])<<16 | int64(buffer[offset+3])<<24 | int64(buffer[offset+4])<<32 | int64(buffer[offset+5])<<40 | int64(buffer[offset+6])<<48 | int64(buffer[offset+7])<<56
}

func ReadCode(file string) *Code {
	contents, err := os.ReadFile(file)

	if err != nil {
		println("Error: Could not open file " + file)
		os.Exit(1)
	}

	functions := make([]Function, 0)
	strings := make([]string, 0)

	// first 4 bytes are the start function
	start := ReadInt32(contents, 0)
	// next 4 bytes are the number of strings
	string_count := ReadInt32(contents, 4)

	offset := 8
	for i := 0; i < string_count; i++ {
		strlength := ReadInt32(contents, offset)
		offset += 4
		str := string(contents[offset : offset+strlength])
		offset += strlength
		strings = append(strings, str)
	}

	func_count := ReadInt32(contents, offset)
	offset += 4

	for i := 0; i < func_count; i++ {
		func_length := ReadInt32(contents, offset)
		offset += 4
		functions = append(functions, Function{Instructions: contents[offset : offset+func_length]})
		offset += func_length
	}

	return &Code{Strings: strings, Functions: functions, Start: start}
}

func PrintOpcode(code []byte, i int) int {
	switch code[i] {
	case ADD:
		fmt.Println("Add")
	case SUB:
		fmt.Println("Sub")
	case MUL:
		fmt.Println("Mul")
	case DIV:
		fmt.Println("Div")
	case MOD:
		fmt.Println("Mod")
	case NEG:
		fmt.Println("Neg")
	case NOT:
		fmt.Println("Not")
	case AND:
		fmt.Println("And")
	case OR:
		fmt.Println("Or")
	case EQUAL:
		fmt.Println("Equal")
	case NOT_EQUAL:
		fmt.Println("NotEqual")
	case LESS:
		fmt.Println("Less")
	case LESS_EQUAL:
		fmt.Println("LessEqual")
	case GREATER:
		fmt.Println("Greater")
	case GREATER_EQUAL:
		fmt.Println("GreaterEqual")
	case TRUE:
		fmt.Println("True")
	case FALSE:
		fmt.Println("False")
	case JUMP:
		fmt.Printf("Jump %d\n", ReadInt32(code, i+1))
		i += 4
	case JUMP_IF_FALSE:
		fmt.Printf("JumpIfFalse %d\n", ReadInt32(code, i+1))
		i += 4
	case STORE:
		fmt.Println("Store")
	case INDEX:
		fmt.Println("Index")
	case CONST_U32:
		fmt.Printf("ConstU32 %d\n", ReadInt32(code, i+1))
		i += 4
	case CONST_U64:
		fmt.Printf("ConstU64 %d\n", ReadInt64(code, i+1))
		i += 8
	case CONST_U8:
		fmt.Printf("ConstU8 %d\n", code[i+1])
		i += 1
	case LIST:
		fmt.Printf("List %d\n", ReadInt32(code, i+1))
		i += 4
	case STRING:
		stridx := ReadInt32(code, i+1)
		i += 5
		fmt.Printf("string at index %d\n", stridx)
	case APPEND:
		fmt.Println("Append")
	case DEF_LOCAL:
		fmt.Printf("DefLocal %d\n", ReadInt32(code, i+1))
		i += 4
	case GET_LOCAL:
		fmt.Printf("GetLocal %d\n", ReadInt32(code, i+1))
		i += 4
	case ASSIGN:
		fmt.Printf("Assign %d\n", ReadInt32(code, i+1))
		i += 4
	case LENGTH:
		fmt.Println("Length")
	case CALL:
		fmt.Printf("Call \t fn: %d \t arg count: %d \n", ReadInt32(code, i+1), ReadInt32(code, i+5))
		i += 8
	case RETURN:
		fmt.Println("Return")
	case PRINT:
		fmt.Println("Print")
	case POP:
		fmt.Println("Pop")
	default:
		fmt.Printf("Unknown opcode %d\n", code[i])
	}
	return i
}

// also returns jump targets
func disassemble(code []byte) (map[int]string, map[int]int) {
	i := 0
	jumps := make(map[int]string)
	func_argc := make(map[int]int)
	for i < len(code) {
		fmt.Printf("%5d\t", i)
		if code[i] == JUMP || code[i] == JUMP_IF_FALSE {
			idx := ReadInt32(code, i+1)
			jumps[idx] = rand_str()
		}
		if code[i] == CALL {
			idx := ReadInt32(code, i+1)
			argc := ReadInt32(code, i+5)
			func_argc[idx] = int(argc)
		}
		i = PrintOpcode(code, i)
		i++
	}
	return jumps, func_argc
}

// ------------------------------
// ---- Compiler
// ------------------------------

type Compiler struct {
	lines      []string
	code       *Code
	jmp_labels []map[int]string
	fn_labels  map[int]string
	func_argc  map[int]int
}

func (c *Compiler) WriteData(strings []string) {
	// add strings
	c.AddLine("section .data", false)
	for i := 0; i < len(strings); i++ {
		c.AddLine(fmt.Sprintf("string%d: db \"%s\", 0", i, strings[i]), true)
	}
	c.AddLine(fmt.Sprintf("string%d: db \"%%d\", 10, 0", len(strings)), true)
}

// x86-64 registers
// rdi and rsi are not included
var Registers = []string{"rax", "rbx", "rcx", "rdx", "r8", "r9", "r10", "r11", "r12", "r13", "r14", "r15"}

type ValueLoc uint8

type ValueType uint8

const (
	VALUE_REG ValueLoc = iota
	VALUE_CONST
	VALUE_STACK
)

const (
	TYPE_U8 ValueType = iota
	TYPE_U32
	TYPE_U64
	TYPE_BOOL
	TYPE_STRING
	TYPE_INT_ARR
)

type Value struct {
	Loc   ValueLoc
	Value string
	Type  ValueType
}

func rand_str() string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	b[0] = '_'
	for i := 1; i < 12; i++ {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func (c *Compiler) AddLine(line string, indent bool) {
	if indent {
		c.lines = append(c.lines, "\t")
	} else {
		c.lines = append(c.lines, "")
	}
	c.lines[len(c.lines)-1] += line
	fmt.Println(line)
}

var ArgsReg = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}

// convert bytecodes to assembly
func (c *Compiler) CompileFunc(fn *Function, name string, func_idx int) {
	stack := make([]Value, 0)
	stack_size := 0

	_unused := func() string {
		for _, reg := range Registers {
			found := false
			for _, r := range stack {
				if r.Value == reg {
					found = true
					break
				}
			}
			if !found {
				return reg
			}
		}
		return ""
	}

	c.AddLine(name+":", false)
	c.AddLine("push rbp", true)
	c.AddLine("mov rbp, rsp", true)
	c.AddLine("", false) // space for stack size
	idx_sub := len(c.lines) - 1

	locals := make(map[int]string)

	// move args from registers to stack
	for i := 0; i < c.func_argc[func_idx]; i++ {
		stack_size += 8
		stack_space := fmt.Sprintf("qword [rbp-%d]", stack_size)
		c.AddLine(fmt.Sprintf("mov %s, %s", stack_space, ArgsReg[i]), true)
		locals[i] = stack_space
	}

	binop1 := func(op string) {
		r1 := stack[len(stack)-1]
		r2 := stack[len(stack)-2]
		if r1.Loc == VALUE_CONST && r2.Loc == VALUE_CONST {
			// both are constants
			r := _unused()
			c.AddLine(fmt.Sprintf("mov %s, %s", r, r1.Value), true)
			c.AddLine(fmt.Sprintf("%s %s, %s", op, r, r2.Value), true)
			stack = stack[:len(stack)-2]
			stack = append(stack, Value{VALUE_REG, r, TYPE_U64})
		} else if r1.Loc == VALUE_CONST || r2.Loc == VALUE_CONST {
			// one is constant
			var cnst Value
			var other Value
			if r1.Loc == VALUE_CONST {
				cnst = r1
				other = r2
			} else {
				cnst = r2
				other = r1
			}
			c.AddLine(fmt.Sprintf("%s %s, %s", op, other.Value, cnst.Value), true)
			stack = stack[:len(stack)-2]
			stack = append(stack, other)
		} else if r1.Loc == VALUE_REG || r2.Loc == VALUE_REG {
			// one is register
			var reg Value
			var st Value
			if r1.Loc == VALUE_REG {
				reg = r1
				st = r2
			} else {
				reg = r2
				st = r1
			}
			c.AddLine(fmt.Sprintf("%s %s, %s", op, reg.Value, st.Value), true)
			stack = stack[:len(stack)-2]
			stack = append(stack, reg)
		} else {
			// both are stack
			r := _unused()
			c.AddLine(fmt.Sprintf("mov %s, %s", r, r1.Value), true)
			c.AddLine(fmt.Sprintf("%s %s, %s", op, r, r2.Value), true)
			stack = stack[:len(stack)-2]
			stack = append(stack, Value{VALUE_REG, r, TYPE_U64})
		}
	}

	cmpops := func(op string) {
		r1 := stack[len(stack)-2]
		r2 := stack[len(stack)-1]
		if r1.Loc == VALUE_CONST && r2.Loc == VALUE_CONST {
			// both are constants
			r := _unused()
			c.AddLine(fmt.Sprintf("mov %s, %s", r, r1.Value), true)
			c.AddLine(fmt.Sprintf("%s %s, %s", op, r, r2.Value), true)
			stack = stack[:len(stack)-2]
		} else if r1.Loc == VALUE_CONST || r2.Loc == VALUE_CONST {
			// one is constant
			var cnst Value
			var other Value
			if r1.Loc == VALUE_CONST {
				cnst = r1
				other = r2
			} else {
				cnst = r2
				other = r1
			}
			c.AddLine(fmt.Sprintf("%s %s, %s", op, other.Value, cnst.Value), true)
			stack = stack[:len(stack)-2]
		} else if r1.Loc == VALUE_REG || r2.Loc == VALUE_REG {
			// one is register
			if r1.Loc != VALUE_REG {
				// move r1 to register
				r := _unused()
				c.AddLine(fmt.Sprintf("mov %s, %s", r, r1.Value), true)
				c.AddLine(fmt.Sprintf("%s %s, %s", op, r, r2.Value), true)
			} else {
				c.AddLine(fmt.Sprintf("%s %s, %s", op, r1.Value, r2.Value), true)
			}
		} else {
			// both are stack
			r := _unused()
			c.AddLine(fmt.Sprintf("mov %s, %s", r, r1.Value), true)
			c.AddLine(fmt.Sprintf("%s %s, %s", op, r, r2.Value), true)
			stack = stack[:len(stack)-2]
		}
	}

	for i := 0; i < len(fn.Instructions); i++ {
		instr := fn.Instructions[i]

		PrintOpcode(fn.Instructions, i)

		if _, ok := c.jmp_labels[func_idx][i]; ok {
			c.AddLine(c.jmp_labels[func_idx][i]+":", false)
		}

		switch instr {
		case ADD:
			binop1("add")
		case SUB:
			binop1("sub")
		case MUL:
			binop1("imul")
		case NOT:
			r := stack[len(stack)-1]
			if r.Loc == VALUE_CONST {
				r2 := _unused()
				c.AddLine(fmt.Sprintf("mov %s, %s", r2, r.Value), true)
				c.AddLine(fmt.Sprintf("not %s", r2), true)
			} else if r.Loc == VALUE_REG {
				c.AddLine(fmt.Sprintf("not %s", r.Value), true)
			} else {
				r := _unused()
				c.AddLine(fmt.Sprintf("mov %s, %s", r, stack[len(stack)-1].Value), true)
				c.AddLine(fmt.Sprintf("not %s", r), true)
				stack = stack[:len(stack)-1]
				stack = append(stack, Value{VALUE_REG, r, TYPE_U64})
			}
		case EQUAL:
			cmpops("cmp")
			r := _unused()
			eqlabel := rand_str()
			merge := rand_str()
			c.AddLine(fmt.Sprintf("je %s", eqlabel), true)
			c.AddLine(fmt.Sprintf("mov %s, 0", r), true)
			c.AddLine(fmt.Sprintf("jmp %s", merge), true)
			c.AddLine(eqlabel+":", false)
			c.AddLine(fmt.Sprintf("mov %s, 1", r), true)
			c.AddLine(merge+":", false)
			stack = append(stack, Value{VALUE_REG, r, TYPE_BOOL})
		case NOT_EQUAL:
			cmpops("cmp")
			r := _unused()
			eqlabel := rand_str()
			merge := rand_str()
			c.AddLine(fmt.Sprintf("jne %s", eqlabel), true)
			c.AddLine(fmt.Sprintf("mov %s, 0", r), true)
			c.AddLine(fmt.Sprintf("jmp %s", merge), true)
			c.AddLine(eqlabel+":", false)
			c.AddLine(fmt.Sprintf("mov %s, 1", r), true)
			c.AddLine(merge+":", false)
			stack = append(stack, Value{VALUE_REG, r, TYPE_BOOL})
		case LESS:
			cmpops("cmp")
			r := _unused()
			eqlabel := rand_str()
			merge := rand_str()
			c.AddLine(fmt.Sprintf("jl %s", eqlabel), true)
			c.AddLine(fmt.Sprintf("mov %s, 0", r), true)
			c.AddLine(fmt.Sprintf("jmp %s", merge), true)
			c.AddLine(eqlabel+":", false)
			c.AddLine(fmt.Sprintf("mov %s, 1", r), true)
			c.AddLine(merge+":", false)
			stack = append(stack, Value{VALUE_REG, r, TYPE_BOOL})
		case LESS_EQUAL:
			cmpops("cmp")
			r := _unused()
			eqlabel := rand_str()
			merge := rand_str()
			c.AddLine(fmt.Sprintf("jle %s", eqlabel), true)
			c.AddLine(fmt.Sprintf("mov %s, 0", r), true)
			c.AddLine(fmt.Sprintf("jmp %s", merge), true)
			c.AddLine(eqlabel+":", false)
			c.AddLine(fmt.Sprintf("mov %s, 1", r), true)
			c.AddLine(merge+":", false)
			stack = append(stack, Value{VALUE_REG, r, TYPE_BOOL})
		case GREATER:
			cmpops("cmp")
			r := _unused()
			eqlabel := rand_str()
			merge := rand_str()
			c.AddLine(fmt.Sprintf("jg %s", eqlabel), true)
			c.AddLine(fmt.Sprintf("mov %s, 0", r), true)
			c.AddLine(fmt.Sprintf("jmp %s", merge), true)
			c.AddLine(eqlabel+":", false)
			c.AddLine(fmt.Sprintf("mov %s, 1", r), true)
			c.AddLine(merge+":", false)
			stack = append(stack, Value{VALUE_REG, r, TYPE_BOOL})
		case GREATER_EQUAL:
			cmpops("cmp")
			r := _unused()
			eqlabel := rand_str()
			merge := rand_str()
			c.AddLine(fmt.Sprintf("jge %s", eqlabel), true)
			c.AddLine(fmt.Sprintf("mov %s, 0", r), true)
			c.AddLine(fmt.Sprintf("jmp %s", merge), true)
			c.AddLine(eqlabel+":", false)
			c.AddLine(fmt.Sprintf("mov %s, 1", r), true)
			c.AddLine(merge+":", false)
			stack = append(stack, Value{VALUE_REG, r, TYPE_BOOL})
		case TRUE:
			stack = append(stack, Value{VALUE_CONST, "1", TYPE_BOOL})
		case FALSE:
			stack = append(stack, Value{VALUE_CONST, "0", TYPE_BOOL})
		case AND:
			binop1("and")
			stack[len(stack)-1].Type = TYPE_BOOL
		case OR:
			binop1("or")
			stack[len(stack)-1].Type = TYPE_BOOL
		case CONST_U64:
			val := ReadInt64(fn.Instructions, i+1)
			i += 8
			stack = append(stack, Value{VALUE_CONST, fmt.Sprintf("%d", val), TYPE_U64})
		case CONST_U8:
			val := fn.Instructions[i+1]
			i += 1
			stack = append(stack, Value{VALUE_CONST, fmt.Sprintf("%d", val), TYPE_U8})
		case JUMP:
			instr_idx := ReadInt32(fn.Instructions, i+1)
			i += 4
			c.AddLine(fmt.Sprintf("jmp %s", c.jmp_labels[func_idx][instr_idx]), true)
		case JUMP_IF_FALSE:
			instr_idx := ReadInt32(fn.Instructions, i+1)
			i += 4
			r := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			c.AddLine(fmt.Sprintf("cmp %s, 0", r.Value), true)
			c.AddLine(fmt.Sprintf("je %s", c.jmp_labels[func_idx][instr_idx]), true)
		case PRINT:
			r := stack[len(stack)-1]
			c.AddLine(fmt.Sprintf("mov rdi, %s", r.Value), true)
			if r.Type == TYPE_U64 {
				c.AddLine("call print_int", true)
			} else if r.Type == TYPE_U8 {
				c.AddLine("call print_char", true)
			} else if r.Type == TYPE_BOOL {
				c.AddLine("call print_bool", true)
			} else {
				fmt.Println("cannot print type", r.Type)
				os.Exit(1)
			}
		case DEF_LOCAL:
			idx := ReadInt32(fn.Instructions, i+1)
			val := stack[len(stack)-1]
			stack_space := (idx + 1) * 8
			stack_size = max(stack_size, stack_space)
			dest := fmt.Sprintf("qword [rbp-%d]", stack_space)
			c.AddLine(fmt.Sprintf("mov %s, %s", dest, val.Value), true)
			locals[idx] = dest
			i += 4
		case GET_LOCAL:
			idx := ReadInt32(fn.Instructions, i+1)
			r := _unused()
			var loc ValueLoc
			if r == "" {
				r = locals[idx]
				loc = VALUE_STACK
			} else {
				c.AddLine(fmt.Sprintf("mov %s, %s", r, locals[idx]), true)
				loc = VALUE_REG
			}
			stack = append(stack, Value{loc, r, TYPE_U64})
			i += 4
		case ASSIGN:
			idx := ReadInt32(fn.Instructions, i+1)
			val := stack[len(stack)-1]
			c.AddLine(fmt.Sprintf("mov %s, %s", locals[idx], val.Value), true)
			stack = stack[:len(stack)-1]
			i += 4
		case CALL:
			idx := ReadInt32(fn.Instructions, i+1)
			i += 4
			argc := ReadInt32(fn.Instructions, i+1)
			i += 4
			if argc > len(ArgsReg) {
				fmt.Println("too many arguments")
				os.Exit(1)
			}
			args := stack[len(stack)-int(argc):]
			stack = stack[:len(stack)-int(argc)]
			for i, _ := range args {
				arg := ArgsReg[i]
				c.AddLine(fmt.Sprintf("mov %s, %s", arg, args[i].Value), true)
			}
			if idx == func_idx {
				// recursive call
				// save all registers
				new_stack := make([]Value, 0)
				for _, r := range stack {
					if r.Loc == VALUE_STACK {
						new_stack = append(new_stack, r)
					} else {
						stack_space := (stack_size + 8)
						stack_size = max(stack_size, stack_space)
						dest := fmt.Sprintf("qword [rbp-%d]", stack_space)
						c.AddLine(fmt.Sprintf("mov %s, %s", dest, r.Value), true)
						new_stack = append(new_stack, Value{VALUE_STACK, dest, TYPE_U64})
					}
				}
				stack = new_stack
			}
			c.AddLine(fmt.Sprintf("call %s", c.fn_labels[idx]), true)
			stack = append(stack, Value{VALUE_REG, "rax", TYPE_U64})
		case RETURN:
			if len(stack) > 0 {
				r := stack[len(stack)-1]
				if r.Loc == VALUE_REG {
					c.AddLine(fmt.Sprintf("mov rax, %s", r.Value), true)
				}
			}
			c.AddLine(fmt.Sprintf("jmp %s_ret", name), true)
		default:
			fmt.Printf("Cannot compile opcode \t")
			PrintOpcode(fn.Instructions, i)
			fmt.Println()
			os.Exit(1)
		}
	}

	c.lines[idx_sub] = fmt.Sprintf("\tsub rsp, %d", stack_size)

	c.AddLine(fmt.Sprintf("%s_ret:", name), false)
	c.AddLine("mov rsp, rbp", true)
	c.AddLine("pop rbp", true)
	c.AddLine("ret", true)
}

func (c *Compiler) Compile(code *Code) {
	c.WriteData(code.Strings)
	c.AddLine("section .text", false)
	c.AddLine("extern print_int", false)
	c.AddLine("extern print_char", false)
	c.AddLine("extern print_bool", false)
	c.AddLine("global main", false)
	c.AddLine("\n", false)

	for i, fn := range code.Functions {
		c.AddLine("\n", false)
		c.CompileFunc(&fn, c.fn_labels[i], i)
	}
	c.AddLine("\n", false)

	c.AddLine("main:", false)
	c.AddLine(fmt.Sprintf("call %s", c.fn_labels[c.code.Start]), true)
	// exit
	c.AddLine("mov rax, 60", true)
	c.AddLine("mov rdi, 0", true)
	c.AddLine("syscall", true)
}

func (c *Compiler) Write(file string) {
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, line := range c.lines {
		f.WriteString(line + "\n")
	}
}

func main() {
	args := os.Args
	if len(args) < 2 {
		println("usage: ./sahl_aot <file>")
		return
	}
	file := args[1]
	code := ReadCode(file)
	jmp_labels := make([]map[int]string, len(code.Functions))
	fn_labels := make(map[int]string)
	func_argc := make(map[int]int)
	for i := 0; i < len(code.Functions); i++ {
		if i == code.Start {
			fmt.Println("Start function")
		} else {
			fmt.Printf("Function %d\n", i)
		}
		jmps, funargc := disassemble(code.Functions[i].Instructions)
		jmp_labels[i] = jmps
		fn_labels[i] = rand_str()
		for idx, argc := range funargc {
			func_argc[idx] = argc
		}
	}
	compiler := Compiler{
		code:       code,
		lines:      make([]string, 0),
		jmp_labels: jmp_labels,
		fn_labels:  fn_labels,
		func_argc:  func_argc,
	}
	compiler.Compile(code)
	compiler.Write("exe.asm")
}
