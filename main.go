// kggseq project main.go
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"kggseq/controller"
	"kggseq/io"

	gzip "github.com/biogo/hts/bgzf"
	_ "github.com/klauspost/pgzip"
)

func main() {

	const PVERSION = "2.0"       // 3 chars
	const PREL = "KGGSeq"        // space or p (full, or prelease)
	const PDATE = "24/Sep./2016" // 11 chars
	var headInfor string = "@----------------------------------------------------------@\n" + "|        " + PREL + "        |     v" + PVERSION + "     |   " + PDATE + "     |\n" +
		"|----------------------------------------------------------|\n|  (C) 2011 Miaoxin Li,  limx54@yahoo.com                  |\n" +
		"|----------------------------------------------------------|\n|  For documentation, citation & bug-report instructions:  |\n" +
		"|              http://grass.cgs.hku.hk/kggseq              |\n@----------------------------------------------------------@"

	fmt.Println(headInfor)

	start := time.Now()
	logFileName := "kggseq.log"
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", os.Stdout, ":", err)
	}
	multi := io.MultiWriter(logFile, os.Stdout)

	inforLog := log.New(multi, "", log.Ldate|log.Ltime)
	//errLog := log.New(multi, "ERROR:", log.Ldate|log.Ltime)
	//warnLog := log.New(multi, "WARNING:", log.Ldate|log.Ltime)
	var thrNum int = runtime.NumCPU()
	//thrNum = 1
	runtime.GOMAXPROCS(thrNum)
	fmt.Println("CPU number: ", thrNum)
	parseInputStr("1kgafreur.20150813.flt.vcf.gz", thrNum)
	//parseInputStr("1kg.phase3.v5.shapeit2.eas.hg19.chr17.vcf.gz", thrNum)
	end := time.Now()
	inforLog.Println("Elapsed time:", end.Sub(start))
	inforLog.Println("The log information is saved in kggseq.log.\n\n")

}

func parseInputStr(fileName string, thrNum int) {
	var br *bufiots.Reader
	fi, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v, Can't open %s: error: %s\n", os.Args[0], fileName, err)
		os.Exit(1)
	}

	fs, _ := fi.Stat()
	fmt.Printf("The file is %d bytes long\n", fs.Size())
	//fi.Seek(8144, 0)

	//I do not know why the pgzip is slow for bgz format although it is fast for conventional zip format
	//fz, err := gzip.NewReaderN(fi, 500000, 32)
	//So I used a new package hts bgzip
	fz, err := gzip.NewReaderN(fi, 500000, 32)
	if err != nil {
		br = bufiots.NewReader(fi)
	} else {
		br = bufiots.NewReader(fz)
	}

	defer fi.Close()
	readLineNum := make(chan int, thrNum)
	for i := 0; i < thrNum; i++ {
		go vcfparser.ParseLines(br, readLineNum)
	}

	var totalNum int = 0
	for i := 0; i < thrNum; i++ {
		totalNum += <-readLineNum
	}

	fmt.Println("Total line: ", totalNum)

}
