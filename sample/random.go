package sample

import (
	"math/rand"

	"example.com/pcbook/pb"
	"github.com/google/uuid"
)

func randomKeyboardLayout() pb.Keyboard_Layout {
	switch rand.Intn(3) {
	case 1:
		return pb.Keyboard_QWERTY
	case 2:
		return pb.Keyboard_QWERTZ
	default:
		return pb.Keyboard_AZERTY
	}
}

func randomBrand() string {
	return randomstringFromSet("Intel", "AMD")
}
func randomGPUBrand() string {
	return randomstringFromSet("NVIDIA", "AMD")
}
func randomGPUName(brand string) string {
	if brand == "NVIDIA" {
		return randomstringFromSet("GTX 1060", "RTX 2060", "RTX 2070", "RTX 2080", "RTX 3080", "RTX 3060")
	}
	return randomstringFromSet(
		"RX 590",
		"RX 580",
		"RX 5700-XT",
		"RX Vega-56",
	)
}
func randomCPUName(brand string) string {
	if brand == "Intel" {
		return randomstringFromSet("Core i9 13900K", "Core i9 13900KF", "Core i7 13700K", "Core i7 13700KF", "Core i5 13600K", "Core i5 13600KF")
	}
	return randomstringFromSet(
		"Ryzen 7 Pro 2700U",
		"Ryzen 5 Pro 3500U",
		"Ryzen 3 Pro 3200GE",
	)
}

func randomScreenResolution() *pb.Screen_Resolution {
	height := randomInt(1080, 4320)
	width := height * 16 / 9

	resolution := &pb.Screen_Resolution{
		Width:  uint32(width),
		Height: uint32(height),
	}
	return resolution
}

func randomScreenPanel() pb.Screen_Panel {
	if rand.Intn(2) == 1 {
		return pb.Screen_IPS
	}
	return pb.Screen_OLED
}

func randomLaptopName(brand string) string {
	switch brand {
	case "Apple":
		return randomstringFromSet("Macbook Air", "Macbook Pro")
	case "Dell":
		return randomstringFromSet("Latitude", "Vostro", "XPS")
	default:
		return randomstringFromSet("Thinkpad X1", "Thinkpad P1", "Thinkpad P53")
	}
}

func randomLaptopBrand() string {
	return randomstringFromSet("Apple", "Dell", "Lenovo")
}

func randomID() string {
	return uuid.New().String()
}
func randomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func randomFloat64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
func randomFloat32(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func randomstringFromSet(a ...string) string {
	n := len(a)
	if n == 0 {
		return ""
	}
	return a[rand.Intn(n)]
}
func randomBool() bool {
	return rand.Intn(2) == 1
}
