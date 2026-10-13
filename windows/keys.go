package main

// usKey maps a key to its character on the US layout (Qalam's keys are defined on QWERTY).
// Shift decides upper/lower case; Caps Lock is ignored so س never turns into ص by accident.
func usKey(vk uint32, shift bool) (rune, bool) {
	switch {
	case vk >= 'A' && vk <= 'Z':
		if shift {
			return rune(vk), true
		}
		return rune(vk + 32), true
	case vk >= '0' && vk <= '9':
		if shift {
			return rune(")!@#$%^&*("[vk-'0']), true
		}
		return rune(vk), true
	}
	pairs := map[uint32][2]rune{
		0xBA: {';', ':'}, 0xBB: {'=', '+'}, 0xBC: {',', '<'}, 0xBD: {'-', '_'}, 0xBE: {'.', '>'},
		0xBF: {'/', '?'}, 0xC0: {'`', '~'}, 0xDB: {'[', '{'}, 0xDC: {'\\', '|'}, 0xDD: {']', '}'},
		0xDE: {'\'', '"'},
	}
	if p, ok := pairs[vk]; ok {
		if shift {
			return p[1], true
		}
		return p[0], true
	}
	return 0, false
}
