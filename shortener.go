package main

import (
	"fmt"
	"strings"
	"unicode"
)

// Known standard module mappings
var knownModules = map[string]string{
	"SYS":                 "SYS",
	"OS":                  "SYS",
	"HST":                 "HST",
	"HOST":                "HST",
	"KRN":                 "KRN",
	"KERNEL":              "KRN",
	"UPT":                 "UPT",
	"UPTIME":              "UPT",
	"PKG":                 "PKG",
	"PACKAGES":            "PKG",
	"SHL":                 "SHL",
	"SHELL":               "SHL",
	"DSP":                 "DSP",
	"DISPLAY":             "DSP",
	"ENV":                 "ENV",
	"DE":                  "ENV",
	"DESKTOP":             "ENV",
	"DESKTOP ENVIRONMENT": "ENV",
	"DESKTOPENVIRONMENT":  "ENV",
	"MONITOR":             "DSP",
	"WND":                 "WND",
	"WM":                  "WND",
	"WINDOW MANAGER":      "WND",
	"WINDOWMANAGER":       "WND",
	"BRD":                 "BRD",
	"BOARD":               "BRD",
	"BIO":                 "BIO",
	"BIOS":                "BIO",
	"WM THEME":            "WMT",
	"WMTHEME":             "WMT",
	"THEME":               "THM",
	"ICONS":               "ICN",
	"FONT":                "FNT",
	"CURSOR":              "CUR",
	"TRM":                 "TRM",
	"TERMINAL":            "TRM",
	"TERMINAL FONT":       "TFT",
	"CPU":                 "CPU",
	"GPU":                 "GPU",
	"RAM":                 "RAM",
	"MEMORY":              "RAM",
	"SWP":                 "SWP",
	"SWAP":                "SWP",
	"DSK":                 "DSK",
	"DISK":                "DSK",
	"BAT":                 "BAT",
	"BATTERY":             "BAT",
	"PWR":                 "PWR",
	"POWER ADAPTER":       "PWR",
	"POWERADAPTER":        "PWR",
	"WIF":                 "WIF",
	"WIFI":                "WIF",
	"LIP":                 "LIP",
	"LOCALIP":             "LIP",
	"LOCAL IP":            "LIP",
	"PIP":                 "PIP",
	"PUBLIC IP":           "PIP",
	"PUBLICIP":            "PIP",
	"LOC":                 "LOC",
	"LOCALE":              "LOC",
	"BLT":                 "BLT",
	"BLUETOOTH":           "BLT",
	"SND":                 "SND",
	"SOUND":               "SND",
	"MDA":                 "MDA",
	"MEDIA":               "MDA",
	"FLT":                 "FLT",
}

// Prefixes for multiple instances of same type (Method 2: Base-36)
var multiPrefixes = map[string]string{
	"DSK":  "DK",
	"DISK": "DK",
	"GPU":  "GP",
	"DSP":  "DP",
	"BAT":  "BT",
	"LIP":  "LP",
}

// cleanRawKey strips parenthetical notes e.g. "Disk (/)" -> "Disk", "Display (Virtual-1)" -> "Display"
func cleanRawKey(raw string) string {
	idx := strings.Index(raw, "(")
	if idx != -1 {
		raw = raw[:idx]
	}
	return strings.TrimSpace(raw)
}

// base36Char converts an integer index (1-based: 1..36) to a single character '1'..'9', 'A'..'Z'.
func base36Char(idx int) byte {
	if idx >= 1 && idx <= 9 {
		return byte('0' + idx)
	}
	if idx >= 10 && idx <= 36 {
		return byte('A' + (idx - 10))
	}
	return 'Z' // Fallback cap
}

// isVowel checks if a character is an English vowel.
func isVowel(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

// GenerateShortKey generates a strictly 3-character uppercase abbreviation for any key.
func GenerateShortKey(raw string) string {
	cleaned := cleanRawKey(raw)
	rawUpper := strings.ToUpper(cleaned)

	// 1. Direct dictionary match
	if mapped, exists := knownModules[rawUpper]; exists {
		return mapped
	}

	// 2. Multi-word acronym check (e.g. "Some New Thing" -> "SNT", "Desktop Environment" -> "DE" -> "ENV")
	words := strings.Fields(cleaned)
	if len(words) >= 2 {
		var acronym []rune
		for _, w := range words {
			runes := []rune(w)
			if len(runes) > 0 && unicode.IsLetter(runes[0]) {
				acronym = append(acronym, unicode.ToUpper(runes[0]))
			}
		}
		if len(acronym) == 2 {
			if mapped, exists := knownModules[string(acronym)]; exists {
				return mapped
			}
			return string(acronym) + "_"
		} else if len(acronym) >= 3 {
			return string(acronym[:3])
		}
	}

	// 3. Clean out non-letter characters
	var letters []rune
	for _, r := range cleaned {
		if unicode.IsLetter(r) {
			letters = append(letters, unicode.ToUpper(r))
		}
	}

	if len(letters) == 0 {
		return "UNK"
	}
	if len(letters) <= 3 {
		return fmt.Sprintf("%-3.3s", string(letters))
	}

	// 4. Drop vowels after the first letter (consonant compression)
	first := letters[0]
	var consonants []rune
	for _, r := range letters[1:] {
		if !isVowel(r) {
			consonants = append(consonants, r)
		}
	}

	var res string
	if len(consonants) >= 2 {
		res = string(first) + string(consonants[:2])
	} else if len(consonants) == 1 {
		res = string(first) + string(consonants[0]) + string(letters[1])
	} else {
		res = string(letters[:3])
	}

	return fmt.Sprintf("%-3.3s", strings.ToUpper(res))
}

// KeyManager handles key resolution, tracking duplicates, and assigning Base-36 numbers.
type KeyManager struct {
	counts    map[string]int
	allocated map[string]int
}

// NewKeyManager initializes key tracking.
func NewKeyManager() *KeyManager {
	return &KeyManager{
		counts:    make(map[string]int),
		allocated: make(map[string]int),
	}
}

// RegisterRawKey counts occurrences before formatting.
func (km *KeyManager) RegisterRawKey(raw string) {
	baseKey := GenerateShortKey(raw)
	km.counts[baseKey]++
}

// ResolveKey returns the final 3-character key, adding Base-36 suffix if multiple instances exist.
func (km *KeyManager) ResolveKey(raw string) string {
	baseKey := GenerateShortKey(raw)
	total := km.counts[baseKey]

	if total <= 1 {
		return baseKey
	}

	// Multiple instances detected (e.g. multiple Disks)
	km.allocated[baseKey]++
	idx := km.allocated[baseKey]

	// Determine 2-letter prefix
	prefix, hasPrefix := multiPrefixes[baseKey]
	if !hasPrefix {
		if len(baseKey) >= 2 {
			prefix = baseKey[:2]
		} else {
			prefix = "M"
		}
	}

	// Append Base-36 single character (1..9, A..Z)
	suffix := string(base36Char(idx))
	return fmt.Sprintf("%-2.2s%s", prefix, suffix)
}
