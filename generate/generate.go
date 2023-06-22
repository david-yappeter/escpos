package generate

import (
	"fmt"
	"image"
	"math"

	"github.com/david-yappeter/escpos"
	"github.com/david-yappeter/escpos/raster"
)

const (
	esc byte = 0x1b
	// ASCII dle (DataLinkEscape)
	dle byte = 0x10

	// ASCII eot (EndOfTransmission)
	eot byte = 0x04

	// ASCII gs (Group Separator)
	gs byte = 0x1D
)

func Init() []byte {
	return []byte("\x1B@")
}

func End() []byte {
	return []byte("\xFA")
}

func Cut() []byte {
	return []byte("\x1DVA0")
}

func CutPartial() []byte {
	return []byte{gs, 0x56, 1}
}

func Cash() []byte {
	return []byte("\x1B\x70\x00\x0A\xFF")
}

func Linefeed() []byte {
	return []byte("\n")
}

func FormfeedN(n int) []byte {
	return []byte(fmt.Sprintf("\x1Bd%c", n))
}

func Formfeed() []byte {
	return FormfeedN(1)
}

func SetFont(font string) []byte {
	f := 0

	switch font {
	case "A":
		f = 0
	case "B":
		f = 1
	case "C":
		f = 2
	default:
		f = 0
	}

	return []byte(fmt.Sprintf("\x1BM%c", f))
}

func SetFontSize(width, height uint8) []byte {
	if !(width > 0 && height > 0 && width <= 8 && height <= 8) {
		width = 1
		height = 1
	}
	return []byte(fmt.Sprintf("\x1D!%c", ((width-1)<<4)|(height-1)))
}

func SetUnderline(v uint8) []byte {
	return []byte(fmt.Sprintf("\x1B-%c", v))
}

func SetEmphasize(v uint8) []byte {
	return []byte(fmt.Sprintf("\x1BG%c", v))
}

func SetUpsidedown(v uint8) []byte {
	return []byte(fmt.Sprintf("\x1B{%c", v))
}

func SetRotate(v uint8) []byte {
	return []byte(fmt.Sprintf("\x1BR%c", v))
}

func SetReverse(v uint8) []byte {
	return []byte(fmt.Sprintf("\x1DB%c", v))
}

func SetSmooth(v uint8) []byte {
	return []byte(fmt.Sprintf("\x1Db%c", v))
}

func SetMoveX(x uint16) []byte {
	return []byte{0x1b, 0x24, byte(x % 256), byte(x / 256)}
}

func SetMoveY(y uint16) []byte {
	return []byte{0x1d, 0x24, byte(y % 256), byte(y / 256)}
}

func SetAlign(align string) []byte {
	a := 0
	switch align {
	case "left":
		a = 0
	case "center":
		a = 1
	case "right":
		a = 2
	default:
		a = 0
	}
	return []byte(fmt.Sprintf("\x1Ba%c", a))
}

func Barcode(barcode string, format escpos.BarcodeFormat) []byte {
	var code byte
	switch format {
	case escpos.BarcodeFormatUPC_A:
		code = 0x00
	case escpos.BarcodeFormatUPC_E:
		code = 0x01
	case escpos.BarcodeFormatEAN13:
		code = 0x02
	case escpos.BarcodeFormatEAN8:
		code = 0x03
	case escpos.BarcodeFormatCode39:
		code = 0x04
	case escpos.BarcodeFormatCode128:
		code = 0x49
	}

	// write barcode
	if format > 69 {
		return append([]byte{gs, 'k', code, byte(len(barcode))}, []byte(barcode)...)
	} else if format < 69 {
		return append(append([]byte{gs, 'k', code}, []byte(barcode)...), 0x00)
	}

	return []byte{}
}

func QRCode(code string, model bool, size uint8, correctionLevel escpos.QRCodeErrorCorrectionLevel) ([]byte, error) {
	datas := []byte{}
	if len(code) > 7089 {
		return nil, fmt.Errorf("the code is too long, it's length should be smaller than 7090")
	}
	if size < 1 {
		size = 1
	}
	if size > 16 {
		size = 16
	}
	var m byte = 49
	// set the qr code model
	if model {
		m = 50
	}
	datas = append(datas, []byte{gs, '(', 'k', 4, 0, 49, 65, m, 0}...)

	// set the qr code size
	datas = append(datas, []byte{gs, '(', 'k', 3, 0, 49, 67, size}...)

	// set the qr code error correction level
	if correctionLevel < 48 {
		correctionLevel = 48
	}
	if correctionLevel > 51 {
		correctionLevel = 51
	}
	datas = append(datas, []byte{gs, '(', 'k', 3, 0, 49, 69, size}...)

	// store the data in the buffer
	// we now write stuff to the printer, so lets save it for returning

	// pL and pH define the size of the data. Data ranges from 1 to (pL + pH*256)-3
	// 3 < pL + pH*256 < 7093
	var codeLength = len(code) + 3
	var pL, pH byte
	pH = byte(int(math.Floor(float64(codeLength) / 256)))
	pL = byte(codeLength - 256*int(pH))

	datas = append(datas, append([]byte{gs, '(', 'k', pL, pH, 49, 80, 48}, []byte(code)...)...)

	// finally print the buffer
	datas = append(datas, []byte{gs, '(', 'k', 3, 0, 49, 81, 48}...)

	return datas, nil
}

func SetMarginLeft(marginLeft int) []byte {
	return []byte{gs, 76, byte(marginLeft % 256), byte(marginLeft / 256)}
}

func PrintRasterImage(img image.Image, incrementation int, startXPos, startYPos, endXPos, endYPos int) []byte {
	datas := []byte{}
	printWidth, printHeight, data := raster.PrintRasterImageProcess(img)

	var yPos byte = byte(startYPos % 256)
	var yPosH byte = byte(startYPos / 256)

	for i := 0; i < printHeight/8/3; i++ {
		datas = append(datas, []byte{esc, 87, byte(startXPos % 256), byte(startXPos / 256), yPos, yPosH, byte(endXPos % 256), byte(endXPos / 256), byte(endYPos % 256), byte(endYPos / 256)}...)

		datas = append(datas, append([]byte{esc, 42, 33, byte(printWidth), byte(printWidth / 256)}, data[i*printWidth*3:((i+1)*printWidth*3)]...)...)

		if int(yPos)+incrementation > 255 {
			yPosH += byte(int(yPos) + incrementation)
		}
		yPos = byte(int(yPos) + incrementation)
	}

	return datas
}

func SetPageMode() []byte {
	return []byte{esc, 76}
}

func SetStandardMode() []byte {
	return []byte{esc, 83}
}

func SetPrintArea(startXPos, startYPos, endXPos, endYPos int) []byte {
	return []byte{esc, 87, byte(startXPos % 256), byte(startXPos / 256), byte(startYPos % 256), byte(startYPos / 256), byte(endXPos % 256), byte(endXPos / 256), byte(endYPos % 256), byte(endYPos / 256)}
}

func SetPrintDirection(direction uint8) []byte {
	switch direction {
	case 0, 1, 2, 3, 48, 49, 50, 51:
		// ignore
	default:
		direction = 0
	}
	return []byte{gs, 84, direction}
}

func PrintPageModeBufferData() []byte {
	return []byte{esc, 12}
}
