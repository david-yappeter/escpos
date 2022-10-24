package main

import (
	"bufio"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/david-yappeter/escpos"
)

func main() {
	f, err := os.OpenFile("/dev/usb/lp0", os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	write := bufio.NewWriter(f)
	reade := bufio.NewReader(f)

	w := bufio.NewReadWriter(reade, write)
	p := escpos.New(w)

	p.Init()
	p.SetAlign("center")

	p.Barcode("1234A", escpos.BarcodeFormatCode128)

	p.Linefeed()
	p.Linefeed()

	w.Flush()

	p.QRCode("ABCDE", true, 10, escpos.QRCodeErrorCorrectionLevelM)

	p.Linefeed()
	p.Linefeed()
	p.Linefeed()
	p.Linefeed()
	p.Linefeed()
	err = w.Flush()

	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	// p.Cut()
	p.End()
	if err = w.Flush(); err != nil {
		panic(err)
	}

}
