package main

import "fmt"

// Output prefixes (ASCII for compatibility, emoji mapping in CLAUDE.md)
const (
	prefixHost    = "[H]"
	prefixTask    = "[T]"
	prefixStep    = ">"
	prefixOK      = "[ok]"
	prefixDone    = "[OK]"
	prefixError   = "[error]"
	prefixWarning = "[!]"
	prefixOverride = "*"
)

// errorf returns a formatted error with prefix.
func errorf(format string, a ...any) error {
	return fmt.Errorf(prefixError+" "+format, a...)
}

// printHost prints a host entry.
func printHost(name, user, address string, port int, source string, override bool) {
	if override {
		fmt.Printf("  %s %s -> %s@%s:%d  %s %s\n", prefixHost, name, user, address, port, prefixOverride, source)
	} else {
		fmt.Printf("  %s %s -> %s@%s:%d  [%s]\n", prefixHost, name, user, address, port, source)
	}
}

// printTask prints a task entry.
func printTask(name, info, source string, override bool) {
	if override {
		fmt.Printf("  %s %s (%s)  %s %s\n", prefixTask, name, info, prefixOverride, source)
	} else {
		fmt.Printf("  %s %s (%s)  [%s]\n", prefixTask, name, info, source)
	}
}

// printTaskHeader prints a task execution header.
func printTaskHeader(name string) {
	fmt.Printf("%s Running %s...\n", prefixTask, name)
}

// printHostHeader prints a host header during task execution.
func printHostHeader(name string) {
	fmt.Printf("  %s %s\n", prefixHost, name)
}

// printStep prints a step being executed.
func printStep(current, total int, step string, indented bool) {
	indent := "  "
	if indented {
		indent = "    "
	}
	fmt.Printf("%s%s [%d/%d] %s\n", indent, prefixStep, current, total, step)
}

// printStepDone prints a completed host.
func printStepDone(name string) {
	fmt.Printf("    %s %s done\n", prefixOK, name)
}

// printSuccess prints a success message.
func printSuccess(format string, a ...any) {
	fmt.Printf(prefixDone+" "+format+"\n", a...)
}

// printWarning prints a warning message.
func printWarning(format string, a ...any) {
	fmt.Printf(prefixWarning+" "+format+"\n", a...)
}

// printSection prints a section header.
func printSection(name string) {
	fmt.Printf("%s %s:\n", prefixTask, name)
}

// printValid prints a valid item in check-config.
func printValid(format string, a ...any) {
	fmt.Printf("  %s "+format+"\n", append([]any{prefixOK}, a...)...)
}

// printInvalid prints an invalid item in check-config.
func printInvalid(name string) {
	fmt.Printf("  %s %s:\n", prefixError, name)
}

// printIssue prints an issue detail.
func printIssue(issue string) {
	fmt.Printf("      -> %s\n", issue)
}
