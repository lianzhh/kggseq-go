// vcfparser project vcfparser.go
package vcfparser

import (
	"bufiots"
	"math"
	"bytes"
)

const(
	UNKNOWN_CHROM_NAME0 = []byte("Un")
	UNKNOWN_CHROM_NAME1 = []byte("GL")	
)

type vcfInfo struct{
	totalVarNum int32
	ignoredInproperChromNum int32
	hasChrLabel bool
	vcf.chromNameIndexMap:=map[string]int{
		"1": 1
		"2": 2
		"23" 23
		"24" 24
		"X": 23
		"Y": 24
		"M": 25
		//to be complete...
	}
}

type variant struct {
	chromID int
	refStartPosition int32
	label            string
	refGeneAnnot     string
	gEncodeAnnot     string
	knownGeneAnnot   string
	ensemblGeneAnnot string
	refAllele        []string
	altAlleles       []string

	//for dbnsfp
	scores1 []float32
	//for non-coding variants
	scores2  []float32
	geneSymb string
	isIndel  bool

	//-1 denotes this SNP does not exist in db NA means db has this variant but no frequency information
	altAF             float32
	localAltAF        float32
	featureValues     []string
	isIBS             bool
	smallestFeatureID byte //by default

	affectedRefHomGtyNum   int32
	affectedHetGtyNum      int32
	affectedAltHomGtyNum   int32
	unaffectedRefHomGtyNum int32
	unaffectedHetGtyNum    int32
	unaffectedAltHomGtyNum int32
	missingtyNum           int32

	compressedGtyLabel int32
	compressedGty      []int32
	encodedGty         []byte

	readInfor      []byte
	chrID          byte
	hasBeenAcced   bool
	consecutiveVar bool
}

func parseInt(s []byte, start int, end int) int {
	// Check for a sign.
	var num int = 0
	var sign int = -1
	var i = start
	//ACSII
	//'0' : 48
	//'9': 57
	//'-' 45
	//'.' 46
	//'e':101
	//'E':69
	//' ': 32
	for s[i] == 32 {
		i++
	}

	ch := s[i]
	i++
	if ch == 45 {
		sign = 1
	} else {
		num = int(48 - ch)
	}

	// Build the number.
	for i < end {
		if s[i] == 46 {
			return sign * num
		} else if s[i] < 48 || s[i] > 57 {
			i++
		} else {
			num = (num*10 + int(48-s[i]))
			i++
		}
	}
	return sign * num
}

func parseFloat(f []byte, start int, end int) float64 {
	var ret float64 = 0  // return value
	var pos int = start  // read pointer position
	var part int = 0     // the current part (int, float and sci parts of the number)
	var neg bool = false // true if part is a negative number
	// the max long is 2147483647
	const MAX_INT_BIT = 9

	//ACSII
	//'0' : 48
	//'9': 57
	//'-' 45
	//'.' 46
	//'e':101
	//'E':69
	for f[pos] == ' ' {
		pos++
	}
	// find start
	for pos < end && (f[pos] < 48 || f[pos] > 57) && f[pos] != 45 && f[pos] != 46 {
		pos++
	}

	// sign
	if f[pos] == 45 {
		neg = true
		pos++
	}

	// integer part
	for pos < end && !(f[pos] > 57 || f[pos] < 48) {
		part = part*10 + int(f[pos]-48)
		pos++
	}

	if neg {
		ret = float64(part * -1)
	} else {
		ret = float64(part)
	}

	// float part
	if pos < end && f[pos] == 46 {
		pos++
		var mul float64 = 1
		part = 0
		var num int = 0
		for pos < end && !(f[pos] > 57 || f[pos] < 48) {
			num++
			if num <= MAX_INT_BIT {
				part = part*10 + int(f[pos]-48)
				mul *= 10
			}
			pos++
		}
		if neg {
			ret = ret - float64(float64(part)/mul)
		} else {
			ret = ret + float64(float64(part)/mul)
		}
	}

	// scientific part
	if pos < end && (f[pos] == 101 || f[pos] == 69) {
		pos++
		neg = (f[pos] == 45)
		pos++
		part = 0
		for pos < end && !(f[pos] > 57 || f[pos] < 48) {
			part = part*10 + int(f[pos]-48)
			pos++
		}
		if neg {
			ret = ret / math.Pow10(part)
		} else {
			ret = ret * math.Pow10(part)
		}
	}
	return ret
}

func empty() int {
	//just for demo. 
	return 0
}

func parseLines(br *bufiots.Reader, readLineNum chan int) {
	tabIndexes := make([]int, 0)
	//var hasNoIndexes bool = true
	var count int = 0
	var lineNum int = 0
	var maxLen int = 0
	var vcf vcfInfo

	
	for {
		line, err := br.ReadBytes('\n')

		if err != nil {
			//fmt.Println("Done reading file", err)
			break
		}
		
		if line[0]=="#" {
			continue
		}
		
		vcf.totalVarNum++
		cntFlag := 0
		currPos, postPos := 0
		var currV variant
		
		for currPos:=range line {
			if line[currPos]!='\t' || line[currPos]!='\n' {
				continue
			}
			
			
			switch cntFlag++ {
				case 1:	//Step1: Chromosome ID
					if bytes.HasPrefix(line[postPos:currPos],UNKNOWN_CHROM_NAME0) || bytes.HasPrefix(line[postPos:currPos],UNKNOWN_CHROM_NAME1) {
						vcf.ignoredInproperChromNum++
						break
					}
			
					if line[postPos]=='c' || line[postPos]=='C' {
						if !vcf.hasChrLabel {
							vcf.hasChrLabel=true
						}
						postPos=postPos+3
					}
			
					if line[postPos]=='m' || line[postPos]=='M' {
						postPos++
					}
					
					currV.chromID=vcfInfo.chromNameIndexMap[string(line[postPos:currPos])]
					if currV.chromID==nil {
						break
					}
					
				case 2:	//Step2: Position
					currV.refStartPosition:= parseInt(line,postPos,currPos-1)
					empty()//check isRegionInModel
					empty()//check isRegionOutModel
					
					
				case 3:	//Step3: rsID
					if line[postPos]=='r' {
						
					}
				case 4: //Step4: REF
					currV.refAllele:=parseDelimeter2String(line,postPos,currPos-1)
					
					
			}
			

			
			
			
			
			
			postPos=currPos+1
			
		//		count = 0
		//		for i := range line {
		//			if line[i] == '\t' {
		//				//fmt.Println(count)
		//				if maxLen <= count+1 {
		//					tabIndexes = append(tabIndexes, i)
		//					maxLen++
		//				} else {
		//					tabIndexes[count] = i
		//				}
		//				count++
		//			}
		//		}
		//		lineNum++
		//		//fmt.Println(lineNum)
	}
	readLineNum <- lineNum

}
