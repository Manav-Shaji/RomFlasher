package platform

import (
	"testing"
)

func TestFetchFastbootDetails_Parsing(t *testing.T) {
	// Using the actual regexs from scanner.go
	outputStr := `(bootloader) product: beryllium
(bootloader) battery-voltage: 4100
(bootloader) secure: yes
(bootloader) unlocked: yes
all: done`

	if m := reProduct.FindStringSubmatch(outputStr); len(m) > 1 {
		if m[1] != "beryllium" {
			t.Errorf("expected beryllium, got %s", m[1])
		}
	}

	if m := reBattery.FindStringSubmatch(outputStr); len(m) > 1 {
		if m[1] != "4100" {
			t.Errorf("expected 4100, got %s", m[1])
		}
	}
}
