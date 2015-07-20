package main
import(
	"net/http"
	"text/template"
	"goquery"
	"log"
	"fmt"
	"strings"
	"strconv"
	"regexp"
	"bytes"
	"encoding/json"
	"io/ioutil"
)

type Context struct {
	FirstName string
	Item string
	Title string
}

type test_struct struct {
	Test string
}

type data_struct struct {
	Prices []float64
	Times []string
}


func myPostFunc(w http.ResponseWriter, req *http.Request){
	decoder := json.NewDecoder(req.Body)
	var t test_struct
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	prices, times := getPrices(t.Test)
	ds := data_struct{prices, times}
	b,err := json.Marshal(ds)
	if err != nil{
		log.Fatal(err)
	}else{
		w.Write(b)
	}
}
func myHandlerFunc(w http.ResponseWriter, req *http.Request){
	bytedoc, err := ioutil.ReadFile("doc.html")
	if err == nil{
		doc := string(bytedoc)
		w.Header().Add("Content Type", "text/html")
		tmpl, err := template.New("anyNameForTemplate").Parse(doc)
		if err == nil{
			context := Context{
				"Todd",
				req.URL.Path[1:], 
				"eBay sold prices",
			}
			tmpl.Execute(w, context)
		}
	}else{
		log.Fatal(err)
	}
}

func priceFinder(firstPrice string) (float64){
	reg, err := regexp.Compile("[^0-9.]")
	if err != nil{
		log.Fatal(err)
	}
	if (strings.Count(firstPrice, "$") > 1){
		price, err := strconv.ParseFloat(reg.ReplaceAllString(strings.Split(firstPrice, "$")[1], ""), 64)
		if err != nil{
			log.Fatal(err)
		}else{
			return price
		}
	}else{
    	price, err := strconv.ParseFloat(reg.ReplaceAllString(firstPrice, ""), 64)
        if err != nil{
			log.Fatal(err)
		}else{
			return price
		}
	}
	return 0.0
}


func getPrices(item string) ([]float64, []string){
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprint("http://www.ebay.com/sch/CPUs-Processors-/164/i.html?_from=R40&_nkw=", item, "&LH_Complete=1&_ex_kw=xeon&LH_Sold=1&LH_ItemCondition=3000&_ipg=200&rt=nc&_trksid=p2045573.m1684"))
    doc, err := goquery.NewDocument(buffer.String())
    if err != nil {
            log.Fatal(err)
    }
    var prices []float64
    var times []string
    doc.Find("li.sresult.lvresult.clearfix").Each(func(i int, s *goquery.Selection) {

    		t := s.Find("span.bold.bidsold").Text()
    		firstPrice := strings.Split(t, "Trending at")[0]

    		price := priceFinder(firstPrice)
            time  := strings.Split(s.Find("span.tme").Text(), " ")

            prices = append(prices, price)
            times = append(times, strings.TrimSpace(time[0]))
    })
    return prices, times
}


func main(){
	http.HandleFunc("/postit", myPostFunc)
    http.HandleFunc("/", myHandlerFunc)
	http.ListenAndServe(":8080", nil)
}

