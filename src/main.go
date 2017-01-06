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

	"github.com/biogo/hts/bgzf"
	//	gzip "github.com/klauspost/compress/gzip"

	gzip "github.com/klauspost/pgzip"
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
	//parseInputStr("1kgafreur.20150813.flt.vcf.gz", thrNum)
	//parseInputStr("1kg.phase3.v5.shapeit2.eas.hg19.chr17.vcf.gz", thrNum)
	//parseInputStr("D:\\01WORK\\GO\\1kgeasasa.flt.vcf.b.gz", thrNum)
	//parseInputStr("D:\\01WORK\\KGGseq\\testdata\\assoc.hg19.vcf.gz", thrNum)
	//parseInputStr("D:\\01WORK\\KGGseq\\testdata\\hg19.crh.env.vcf.gz", thrNum)
	//testBGZF("D:\\01WORK\\KGGseq\\testdata\\hg19.crh.env.vcf.b.gz")
	testBGZF("D:\\01WORK\\KGGseq\\testdata\\ALL.chr22.phase3_shapeit2_mvncall_integrated_v5a.20130502.genotypes.vcf.b.gz")
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
		fmt.Println("Errors in gzip.NewReaderN!")
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
		//		totalNum += <-readLineNum
		var temp = <-readLineNum
		fmt.Println(temp)
		totalNum += temp
	}

	fmt.Println("Total line: ", totalNum)

}

func testBGZF(fileName string) {
	fi, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v, Can't open %s: error: %s\n", os.Args[0], fileName, err)
		os.Exit(1)
	}

	fs, _ := fi.Stat()
	fmt.Printf("The file is %d bytes long\n", fs.Size())

	bgzfReader, err := bgzf.NewReader(fi, 20)
	var lj [1024 * 1024]byte

	var breakLine vcfparser.Splinter
	end := false
	var line int64
	line = 0
	var currPos int
	currPos = 0
	for !end {
		n, err := bgzfReader.Read(lj[:])
		//fmt.Printf("%d bytes are read!\n", n)
		//		n, err := fi.Read(lj[:])
		currPos = 0
		if err != nil {
			if n > 0 {
				fmt.Printf("The value of n for last block is: %d\n", n)
				for currPos != n {
					_, curr := vcfparser.ReadBGZFLines(lj[:], n, currPos, &breakLine)
					currPos = curr
					//fmt.Println(string(tim))
					//fmt.Printf("%d\n", currPos)
					//break
				}
			}
			//fmt.Println(n)
			//fmt.Println(string(lj[:100]))
			fmt.Printf("\nThis is the end of file: %s\n", err)
			end = true
		} else {
			//fmt.Printf("%d,", n)
			line += int64(n)
			//			fmt.Println(string(lj[1024*1024-50:]))
			//for test the speed of READBGZFLines
			for currPos != n {
				_, curr := vcfparser.ReadBGZFLines(lj[:], n, currPos, &breakLine)
				currPos = curr
				//fmt.Println(string(tim))
				//fmt.Println(currPos)
			}

			//break
		}
	}
	fmt.Printf("Total lines is: %d\n", line)
	bgzfReader.Close()
}
