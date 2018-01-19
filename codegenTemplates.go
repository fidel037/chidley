package main

import (
	"strconv"
)

type XmlInfo struct {
	BaseXML         *XMLType
	OneLevelDownXML []*XMLType
	Structs         string
	Filenames       []string
	Filename        string
}

type XMLType struct {
	NameType, XMLName, XMLNameUpper, XMLSpace string
}

//     b := [5]int{1, 2, 3, 4, 5}
func (xi *XmlInfo) init() {
	xi.Filename = "[" + strconv.Itoa(len(xi.Filenames)) + "]string{"
	for i, _ := range xi.Filenames {
		if i != 0 {
			xi.Filename = xi.Filename + ","
		}
		xi.Filename = xi.Filename + "\"" + xi.Filenames[i] + "\""
	}
	xi.Filename = xi.Filename + "}"
}

const codeTemplate = `

package main

/////////////////////////////////////////////////////////////////
//Code generated by chidley https://github.com/gnewton/chidley //
/////////////////////////////////////////////////////////////////

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

const (
	JsonOut = iota
	XmlOut
	CountAll
)

var toJson bool = false
var toXml bool = false
var oneLevelDown bool = false
var countAll bool = false
var musage bool = false

var uniqueFlags = []*bool{
	&toJson,
	&toXml,
	&countAll}

var filenames = {{.Filename}}



func init() {

	flag.BoolVar(&toJson, "j", toJson, "Convert to JSON")
	flag.BoolVar(&toXml, "x", toXml, "Convert to XML")
	flag.BoolVar(&countAll, "c", countAll, "Count each instance of XML tags")
	flag.BoolVar(&oneLevelDown, "s", oneLevelDown, "Stream XML by using XML elements one down from the root tag. Good for huge XML files (see http://blog.davidsingleton.org/parsing-huge-xml-files-with-go/")
	flag.BoolVar(&musage, "h", musage, "Usage")
	//flag.StringVar(&filename, "f", filename, "XML file or URL to read in")
}

var out int = -1

var counters map[string]*int

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if musage {
		flag.Usage()
		return
	}

	numSetBools, outFlag := numberOfBoolsSet(uniqueFlags)
	if numSetBools == 0 {
		flag.Usage()
                return
	}

	if numSetBools != 1 {
		flag.Usage()
		log.Fatal("Only one of ", uniqueFlags, " can be set at once")
	}

	counters = make(map[string]*int)
	for i, _ := range filenames {
		filename := filenames[i]
		reader, xmlFile, err := genericReader(filename)
		if err != nil {
			log.Fatal(err)
			return
		}

		decoder := xml.NewDecoder(reader)

		for {
			token, _ := decoder.Token()
			if token == nil {
				break
			}
			switch se := token.(type) {
			case xml.StartElement:
				handleFeed(se, decoder, outFlag)
			}
		}
		if xmlFile != nil {
			defer xmlFile.Close()
		}
	}
	if countAll {
		for k, v := range counters {
			fmt.Println(*v, k)
		}
	}
}

func handleFeed(se xml.StartElement, decoder *xml.Decoder, outFlag *bool) {
	if outFlag == &countAll {
		incrementCounter(se.Name.Space, se.Name.Local)
	} else {
                if !oneLevelDown{
        		if se.Name.Local == "{{.BaseXML.XMLName}}" && se.Name.Space == "{{.BaseXML.XMLSpace}}" {
	        	      var item {{.BaseXML.NameType}}
			      decoder.DecodeElement(&item, &se)
			      switch outFlag {
			      case &toJson:
				      writeJson(item)
			      case &toXml:
				      writeXml(item)
			      }
		      }
                }else{
                   {{ range .OneLevelDownXML }}
        		if se.Name.Local == "{{.XMLName}}" && se.Name.Space == "{{.XMLSpace}}" {
	        	      var item {{.NameType}}
			      decoder.DecodeElement(&item, &se)
			      switch outFlag {
			      case &toJson:
				      writeJson(item)
			      case &toXml:
				      writeXml(item)
			      }
		      }
                   {{ end }}
               }
	}
}

func makeKey(space string, local string) string {
	if space == "" {
		space = "_"
	}
	return space + ":" + local
}

func incrementCounter(space string, local string) {
	key := makeKey(space, local)

	counter, ok := counters[key]
	if !ok {
		n := 1
		counters[key] = &n
	} else {
		newv := *counter + 1
		counters[key] = &newv
	}
}

func writeJson(item interface{}) {
	b, err := json.MarshalIndent(item, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

func writeXml(item interface{}) {
	output, err := xml.MarshalIndent(item, "  ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	os.Stdout.Write(output)
}

func genericReader(filename string) (io.Reader, *os.File, error) {
	if filename == "" {
		return bufio.NewReader(os.Stdin), nil, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	if strings.HasSuffix(filename, "bz2") {
		return bufio.NewReader(bzip2.NewReader(bufio.NewReader(file))), file, err
	}

	if strings.HasSuffix(filename, "gz") {
		reader, err := gzip.NewReader(bufio.NewReader(file))
		if err != nil {
			return nil, nil, err
		}
		return bufio.NewReader(reader), file, err
	}
	return bufio.NewReader(file), file, err
}

func numberOfBoolsSet(a []*bool) (int, *bool) {
	var setBool *bool
	counter := 0
	for i := 0; i < len(a); i++ {
		if *a[i] {
			counter += 1
			setBool = a[i]
		}
	}
	return counter, setBool
}


///////////////////////////
/// structs
///////////////////////////

{{.Structs}}
///////////////////////////

`
