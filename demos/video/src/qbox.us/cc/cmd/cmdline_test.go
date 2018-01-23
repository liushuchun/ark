package cmd

import (
	"testing"
)

type parseCmdlineCase struct {
	line string
	args []string
}

func TestParseCmdline(t *testing.T) {
	cases := []parseCmdlineCase{
		{"foo \tbar", []string{"foo", "bar"}},
		{"\t \tfoo \tbar\t \ttest", []string{"foo", "bar", "test"}},
		{"\t \tfoo \tbar\t \ttest \t", []string{"foo", "bar", "test"}},
	}
	for _, v := range cases {
		args := ParseCmdline(v.line)
		if len(args) != len(v.args) {
			t.Error("ParseCmdline:", v.line, args, v.args)
		} else {
			for i, arg := range args {
				if v.args[i] != arg {
					t.Error(i, "ParseCmdline:", v.line, args, v.args)
					break
				}
			}
		}
	}
}

/*
func TestStartProcess(t *testing.T) {
	p, err := os.StartProcess("/bin/bash", []string{"ls -l"}, &os.ProcAttr{})
	if err != nil {
		t.Fatal("StartProcess:", err)
	}
	fmt.Println("StartProcess:", p)
	w, err := p.Wait(0)
	if err != nil {
		t.Fatal("Process.Wait:", err)
	}
	fmt.Println(w)
}
*/
