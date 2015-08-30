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
    "os"
    "bufio"
)

type Context struct {
    Terms string
    Category string
    Condition string
}

type data_struct struct {
    Prices []float64
    Times []string
    Shipping []string
    Titles []string
    URLs []string
}


func myPostFunc(w http.ResponseWriter, req *http.Request){
    decoder := json.NewDecoder(req.Body)
    var t Context
    err := decoder.Decode(&t)
    if err != nil {
        panic(err)
    }
    prices, times, shipping, titles, urls := getPrices(t.Terms, t.Category, t.Condition)
    ds := data_struct{prices, times, shipping, titles, urls}
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
                "hey",
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


func getPrices(item string, category string, condition string) ([]float64, []string, []string, []string, []string){
    var buffer bytes.Buffer
    buffer.WriteString(fmt.Sprint("http://www.ebay.com/sch/",category,"/i.html?_from=R40&LH_Complete=1&LH_Sold=1&_ipg=200&_sop=16&_nkw=",strings.Replace(item, " ", "%20", -1) ,condition))
    fmt.Println(buffer.String())
    doc, err := goquery.NewDocument(buffer.String())
    if err != nil {
        log.Fatal(err)
    }

    var prices []float64
    var times []string
    var shipping []string
    var titles []string
    var urls []string

    doc.Find("li.sresult.lvresult.clearfix").Each(func(i int, s *goquery.Selection) {
        t := s.Find("span.bold.bidsold").Text()
        firstPrice := strings.Split(t, "Trending at")[0]

        price := priceFinder(firstPrice)
        time  := strings.Trim(s.Find("span.tme").Text(), " ")

        prices = append(prices, price)
        times = append(times, time)
    })

    doc.Find("span.ship").Each(func(i int, s *goquery.Selection){
        t := s.Find("span").Text()
        txt := strings.Split(t, " shipping")[0]
        shipping = append(shipping, strings.Trim(txt, " "))
    })

    doc.Find("a.img.imgWr2").Each(func(i int, s *goquery.Selection){
        if val, ok := s.Find("img").Attr("alt"); !ok {
            fmt.Println("Expected a title attribute")
        } else {
            titles = append(titles, val)
        }
        if val, ok := s.Attr("href"); !ok{
            fmt.Println("Expected a value for the href attribute")
        }else {
            urls = append(urls, val)
        }
    })

    if (len(prices) == 0){
        doc.Find("span.amt").Each(func(i int, s *goquery.Selection){
            t := s.Find("span").Text()
            fmt.Println(t)
        });
    }

    return prices, times, shipping, titles, urls
}


type MyHandler struct{
}

func (this *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path[1:]
    log.Println(path)
    f, err := os.Open(path)


    if err == nil {
        bufferedReader := bufio.NewReader(f)

        var contentType string

        if strings.HasSuffix(path, ".css"){
            contentType = "text/css"
        }else if strings.HasSuffix(path, ".html"){
            contentType = "text/html"
        }else if strings.HasSuffix(path, ".js"){
            contentType = "application/javascript"
        }else if strings.HasSuffix(path, ".png"){
            contentType = "image/png"
        }else if strings.HasSuffix(path, ".svg"){
            contentType = "image/svg+xml"
        }else if strings.HasSuffix(path, ".mp4"){
            contentType = "video/mp4"
        }else{
            contentType = "text/plain"
        }

        w.Header().Add("Content Type", contentType)
        bufferedReader.WriteTo(w)
    } else {
        w.WriteHeader(404)
        w.Write([]byte("404 My Friend - " + http.StatusText(404)))
    }
}

func main(){
    http.HandleFunc("/postit", myPostFunc)
    http.Handle("/index.html", new(MyHandler))
    http.Handle("/atad.csv", new(MyHandler))
    http.HandleFunc("/", myHandlerFunc)
    http.ListenAndServe(":8080", nil)
}