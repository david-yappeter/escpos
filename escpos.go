package escpos

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
)

type BarcodeFormat int
type QRCodeErrorCorrectionLevel uint8

const (
	BarcodeFormatUPC_A BarcodeFormat = iota
	BarcodeFormatUPC_E
	BarcodeFormatEAN13
	BarcodeFormatEAN8
	BarcodeFormatCode39
	BarcodeFormatCode128
)

const (
	QRCodeErrorCorrectionLevelL QRCodeErrorCorrectionLevel = iota + 48
	QRCodeErrorCorrectionLevelM
	QRCodeErrorCorrectionLevelQ
	QRCodeErrorCorrectionLevelH
)

const (
	// ASCII DLE (DataLinkEscape)
	DLE byte = 0x10

	// ASCII EOT (EndOfTransmission)
	EOT byte = 0x04

	// ASCII GS (Group Separator)
	GS byte = 0x1D
)

// text replacement map
var textReplaceMap = map[string]string{
	// horizontal tab
	"&#9;":  "\x09",
	"&#x9;": "\x09",

	// linefeed
	"&#10;": "\n",
	"&#xA;": "\n",

	// xml stuff
	"&apos;": "'",
	"&quot;": `"`,
	"&gt;":   ">",
	"&lt;":   "<",

	// ampersand must be last to avoid double decoding
	"&amp;": "&",
}

// replace text from the above map
func textReplace(data string) string {
	for k, v := range textReplaceMap {
		data = strings.Replace(data, k, v, -1)
	}
	return data
}

type Escpos struct {
	// destination
	dst io.ReadWriter

	// font metrics
	width, height uint8

	// state toggles ESC[char]
	underline  uint8
	emphasize  uint8
	upsidedown uint8
	rotate     uint8

	// state toggles GS[char]
	reverse, smooth uint8
}

// reset toggles
func (e *Escpos) reset() {
	e.width = 1
	e.height = 1

	e.underline = 0
	e.emphasize = 0
	e.upsidedown = 0
	e.rotate = 0

	e.reverse = 0
	e.smooth = 0
}

// create Escpos printer
func New(dst io.ReadWriter) (e *Escpos) {
	e = &Escpos{dst: dst}
	e.reset()
	return
}

// write raw bytes to printer
func (e *Escpos) WriteRaw(data []byte) (n int, err error) {
	if len(data) > 0 {
		log.Printf("Writing %d bytes\n", len(data))
		e.dst.Write(data)
	} else {
		log.Printf("Wrote NO bytes\n")
	}

	return 0, nil
}

// read raw bytes from printer
func (e *Escpos) ReadRaw(data []byte) (n int, err error) {
	return e.dst.Read(data)
}

// write a string to the printer
func (e *Escpos) Write(data string) (int, error) {
	return e.WriteRaw([]byte(data))
}

// init/reset printer settings
func (e *Escpos) Init() {
	e.reset()
	e.Write("\x1B@")
}

// end output
func (e *Escpos) End() {
	e.Write("\xFA")
}

// send cut
func (e *Escpos) Cut() {
	e.Write("\x1DVA0")
}

// send cut minus one point (partial cut)
func (e *Escpos) CutPartial() {
	e.WriteRaw([]byte{GS, 0x56, 1})
}

// send cash
func (e *Escpos) Cash() {
	e.Write("\x1B\x70\x00\x0A\xFF")
}

// send linefeed
func (e *Escpos) Linefeed() {
	e.Write("\n")
}

// send N formfeeds
func (e *Escpos) FormfeedN(n int) {
	e.Write(fmt.Sprintf("\x1Bd%c", n))
}

// send formfeed
func (e *Escpos) Formfeed() {
	e.FormfeedN(1)
}

// set font
func (e *Escpos) SetFont(font string) {
	f := 0

	switch font {
	case "A":
		f = 0
	case "B":
		f = 1
	case "C":
		f = 2
	default:
		log.Fatalf("Invalid font: '%s', defaulting to 'A'", font)
		f = 0
	}

	e.Write(fmt.Sprintf("\x1BM%c", f))
}

func (e *Escpos) SendFontSize() {
	e.Write(fmt.Sprintf("\x1D!%c", ((e.width-1)<<4)|(e.height-1)))
}

// set font size
func (e *Escpos) SetFontSize(width, height uint8) {
	if width > 0 && height > 0 && width <= 8 && height <= 8 {
		e.width = width
		e.height = height
		e.SendFontSize()
	} else {
		log.Fatalf("Invalid font size passed: %d x %d", width, height)
	}
}

// send underline
func (e *Escpos) SendUnderline() {
	e.Write(fmt.Sprintf("\x1B-%c", e.underline))
}

// send emphasize / doublestrike
func (e *Escpos) SendEmphasize() {
	e.Write(fmt.Sprintf("\x1BG%c", e.emphasize))
}

// send upsidedown
func (e *Escpos) SendUpsidedown() {
	e.Write(fmt.Sprintf("\x1B{%c", e.upsidedown))
}

// send rotate
func (e *Escpos) SendRotate() {
	e.Write(fmt.Sprintf("\x1BR%c", e.rotate))
}

// send reverse
func (e *Escpos) SendReverse() {
	e.Write(fmt.Sprintf("\x1DB%c", e.reverse))
}

// send smooth
func (e *Escpos) SendSmooth() {
	e.Write(fmt.Sprintf("\x1Db%c", e.smooth))
}

// send move x
func (e *Escpos) SendMoveX(x uint16) {
	e.Write(string([]byte{0x1b, 0x24, byte(x % 256), byte(x / 256)}))
}

// send move y
func (e *Escpos) SendMoveY(y uint16) {
	e.Write(string([]byte{0x1d, 0x24, byte(y % 256), byte(y / 256)}))
}

// set underline
func (e *Escpos) SetUnderline(v uint8) {
	e.underline = v
	e.SendUnderline()
}

// set emphasize
func (e *Escpos) SetEmphasize(u uint8) {
	e.emphasize = u
	e.SendEmphasize()
}

// set upsidedown
func (e *Escpos) SetUpsidedown(v uint8) {
	e.upsidedown = v
	e.SendUpsidedown()
}

// set rotate
func (e *Escpos) SetRotate(v uint8) {
	e.rotate = v
	e.SendRotate()
}

// set reverse
func (e *Escpos) SetReverse(v uint8) {
	e.reverse = v
	e.SendReverse()
}

// set smooth
func (e *Escpos) SetSmooth(v uint8) {
	e.smooth = v
	e.SendSmooth()
}

// pulse (open the drawer)
func (e *Escpos) Pulse() {
	// with t=2 -- meaning 2*2msec
	e.Write("\x1Bp\x02")
}

// set alignment
func (e *Escpos) SetAlign(align string) {
	a := 0
	switch align {
	case "left":
		a = 0
	case "center":
		a = 1
	case "right":
		a = 2
	default:
		log.Fatalf("Invalid alignment: %s", align)
	}
	e.Write(fmt.Sprintf("\x1Ba%c", a))
}

// set language -- ESC R
func (e *Escpos) SetLang(lang string) {
	l := 0

	switch lang {
	case "en":
		l = 0
	case "fr":
		l = 1
	case "de":
		l = 2
	case "uk":
		l = 3
	case "da":
		l = 4
	case "sv":
		l = 5
	case "it":
		l = 6
	case "es":
		l = 7
	case "ja":
		l = 8
	case "no":
		l = 9
	default:
		log.Fatalf("Invalid language: %s", lang)
	}
	e.Write(fmt.Sprintf("\x1BR%c", l))
}

// do a block of text
func (e *Escpos) Text(params map[string]string, data string) {

	// send alignment to printer
	if align, ok := params["align"]; ok {
		e.SetAlign(align)
	}

	// set lang
	if lang, ok := params["lang"]; ok {
		e.SetLang(lang)
	}

	// set smooth
	if smooth, ok := params["smooth"]; ok && (smooth == "true" || smooth == "1") {
		e.SetSmooth(1)
	}

	// set emphasize
	if em, ok := params["em"]; ok && (em == "true" || em == "1") {
		e.SetEmphasize(1)
	}

	// set underline
	if ul, ok := params["ul"]; ok && (ul == "true" || ul == "1") {
		e.SetUnderline(1)
	}

	// set reverse
	if reverse, ok := params["reverse"]; ok && (reverse == "true" || reverse == "1") {
		e.SetReverse(1)
	}

	// set rotate
	if rotate, ok := params["rotate"]; ok && (rotate == "true" || rotate == "1") {
		e.SetRotate(1)
	}

	// set font
	if font, ok := params["font"]; ok {
		e.SetFont(strings.ToUpper(font[5:6]))
	}

	// do dw (double font width)
	if dw, ok := params["dw"]; ok && (dw == "true" || dw == "1") {
		e.SetFontSize(2, e.height)
	}

	// do dh (double font height)
	if dh, ok := params["dh"]; ok && (dh == "true" || dh == "1") {
		e.SetFontSize(e.width, 2)
	}

	// do font width
	if width, ok := params["width"]; ok {
		if i, err := strconv.Atoi(width); err == nil {
			e.SetFontSize(uint8(i), e.height)
		} else {
			log.Fatalf("Invalid font width: %s", width)
		}
	}

	// do font height
	if height, ok := params["height"]; ok {
		if i, err := strconv.Atoi(height); err == nil {
			e.SetFontSize(e.width, uint8(i))
		} else {
			log.Fatalf("Invalid font height: %s", height)
		}
	}

	// do y positioning
	if x, ok := params["x"]; ok {
		if i, err := strconv.Atoi(x); err == nil {
			e.SendMoveX(uint16(i))
		} else {
			log.Fatalf("Invalid x param %s", x)
		}
	}

	// do y positioning
	if y, ok := params["y"]; ok {
		if i, err := strconv.Atoi(y); err == nil {
			e.SendMoveY(uint16(i))
		} else {
			log.Fatalf("Invalid y param %s", y)
		}
	}

	// do text replace, then write data
	data = textReplace(data)
	if len(data) > 0 {
		e.Write(data)
	}
}

// feed the printer
func (e *Escpos) Feed(params map[string]string) {
	// handle lines (form feed X lines)
	if l, ok := params["line"]; ok {
		if i, err := strconv.Atoi(l); err == nil {
			e.FormfeedN(i)
		} else {
			log.Fatalf("Invalid line number %s", l)
		}
	}

	// handle units (dots)
	if u, ok := params["unit"]; ok {
		if i, err := strconv.Atoi(u); err == nil {
			e.SendMoveY(uint16(i))
		} else {
			log.Fatalf("Invalid unit number %s", u)
		}
	}

	// send linefeed
	e.Linefeed()

	// reset variables
	e.reset()

	// reset printer
	e.SendEmphasize()
	e.SendRotate()
	e.SendSmooth()
	e.SendReverse()
	e.SendUnderline()
	e.SendUpsidedown()
	e.SendFontSize()
	e.SendUnderline()
}

// feed and cut based on parameters
func (e *Escpos) FeedAndCut(params map[string]string) {
	if t, ok := params["type"]; ok && t == "feed" {
		e.Formfeed()
	}

	e.Cut()
}

// Barcode sends a barcode to the printer.
func (e *Escpos) Barcode(barcode string, format BarcodeFormat) {
	var code byte
	switch format {
	case BarcodeFormatUPC_A:
		code = 0x00
	case BarcodeFormatUPC_E:
		code = 0x01
	case BarcodeFormatEAN13:
		code = 0x02
	case BarcodeFormatEAN8:
		code = 0x03
	case BarcodeFormatCode39:
		code = 0x04
	case BarcodeFormatCode128:
		code = 0x49
	}

	// reset settings
	e.reset()

	// set align
	e.SetAlign("center")

	// write barcode
	if format > 69 {
		e.Write(string(append([]byte{GS, 'k', code, byte(len(barcode))}, []byte(barcode)...)))
	} else if format < 69 {
		e.Write(string(append(append([]byte{GS, 'k', code}, []byte(barcode)...), 0x00)))
	}
	// e.Write(fmt.Sprintf("%v", barcode))
}

func (e *Escpos) QRCode(code string, model bool, size uint8, correctionLevel QRCodeErrorCorrectionLevel) (int, error) {
	if len(code) > 7089 {
		return 0, fmt.Errorf("the code is too long, it's length should be smaller than 7090")
	}
	if size < 1 {
		size = 1
	}
	if size > 16 {
		size = 16
	}
	var m byte = 49
	var err error
	// set the qr code model
	if model {
		m = 50
	}
	_, err = e.WriteRaw([]byte{GS, '(', 'k', 4, 0, 49, 65, m, 0})
	if err != nil {
		return 0, err
	}

	// set the qr code size
	_, err = e.WriteRaw([]byte{GS, '(', 'k', 3, 0, 49, 67, size})
	if err != nil {
		return 0, err
	}

	// set the qr code error correction level
	if correctionLevel < 48 {
		correctionLevel = 48
	}
	if correctionLevel > 51 {
		correctionLevel = 51
	}
	_, err = e.WriteRaw([]byte{GS, '(', 'k', 3, 0, 49, 69, size})
	if err != nil {
		return 0, err
	}

	// store the data in the buffer
	// we now write stuff to the printer, so lets save it for returning

	// pL and pH define the size of the data. Data ranges from 1 to (pL + pH*256)-3
	// 3 < pL + pH*256 < 7093
	var codeLength = len(code) + 3
	var pL, pH byte
	pH = byte(int(math.Floor(float64(codeLength) / 256)))
	pL = byte(codeLength - 256*int(pH))

	written, err := e.WriteRaw(append([]byte{GS, '(', 'k', pL, pH, 49, 80, 48}, []byte(code)...))
	if err != nil {
		return written, err
	}

	// finally print the buffer
	_, err = e.WriteRaw([]byte{GS, '(', 'k', 3, 0, 49, 81, 48})
	if err != nil {
		return written, err
	}

	return written, nil
}

// used to send graphics headers
func (e *Escpos) gSend(m byte, fn byte, data []byte) {
	l := len(data) + 2

	e.Write("\x1b(L")
	e.WriteRaw([]byte{byte(l % 256), byte(l / 256), m, fn})
	e.WriteRaw(data)
}

// write an image
func (e *Escpos) Image(params map[string]string, data string) {
	// send alignment to printer
	if align, ok := params["align"]; ok {
		e.SetAlign(align)
	}

	// get width
	wstr, ok := params["width"]
	if !ok {
		log.Fatal("No width specified on image")
	}

	// get height
	hstr, ok := params["height"]
	if !ok {
		log.Fatal("No height specified on image")
	}

	// convert width
	width, err := strconv.Atoi(wstr)
	if err != nil {
		log.Fatalf("Invalid image width %s", wstr)
	}

	// convert height
	height, err := strconv.Atoi(hstr)
	if err != nil {
		log.Fatalf("Invalid image height %s", hstr)
	}

	// decode data frome b64 string
	dec, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Image len:%d w: %d h: %d\n", len(dec), width, height)

	// $imgHeader = self::dataHeader(array($img -> getWidth(), $img -> getHeight()), true);
	// $tone = '0';
	// $colors = '1';
	// $xm = (($size & self::IMG_DOUBLE_WIDTH) == self::IMG_DOUBLE_WIDTH) ? chr(2) : chr(1);
	// $ym = (($size & self::IMG_DOUBLE_HEIGHT) == self::IMG_DOUBLE_HEIGHT) ? chr(2) : chr(1);
	//
	// $header = $tone . $xm . $ym . $colors . $imgHeader;
	// $this -> graphicsSendData('0', 'p', $header . $img -> toRasterFormat());
	// $this -> graphicsSendData('0', '2');

	header := []byte{
		byte('0'), 0x01, 0x01, byte('1'),
	}

	a := append(header, dec...)

	e.gSend(byte('0'), byte('p'), a)
	e.gSend(byte('0'), byte('2'), []byte{})

}

// write a "node" to the printer
func (e *Escpos) WriteNode(name string, params map[string]string, data string) {
	cstr := ""
	if data != "" {
		str := data[:]
		if len(data) > 40 {
			str = fmt.Sprintf("%s ...", data[0:40])
		}
		cstr = fmt.Sprintf(" => '%s'", str)
	}
	log.Printf("Write: %s => %+v%s\n", name, params, cstr)

	switch name {
	case "text":
		e.Text(params, data)
	case "feed":
		e.Feed(params)
	case "cut":
		e.FeedAndCut(params)
	case "pulse":
		e.Pulse()
	case "image":
		e.Image(params, data)
	}
}

// ReadStatus Read the status n from the printer
func (e *Escpos) ReadStatus(n byte) (byte, error) {
	e.WriteRaw([]byte{DLE, EOT, n})
	data := make([]byte, 1)
	_, err := e.ReadRaw(data)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}
