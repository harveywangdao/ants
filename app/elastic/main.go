package main

import (
	"context"
	//"encoding/json"
	"fmt"
	"github.com/olivere/elastic"
	//"gopkg.in/olivere/elastic.v5" //这里使用的是版本5，最新的是6，有改动
	//"log"
	//"os"
	"reflect"
)

var client *elastic.Client
var host = "http://192.168.1.10:9200/"

type Employee struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age"`
	About     string   `json:"about"`
	Interests []string `json:"interests"`
}

func init() {
	// errorlog := log.New(os.Stdout, "APP", log.LstdFlags)
	// elastic.SetErrorLog(errorlog)

	var err error
	client, err = elastic.NewClient(elastic.SetURL(host))
	if err != nil {
		panic(err)
	}

	info, code, err := client.Ping(host).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	esversion, err := client.ElasticsearchVersion(host)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)
}

func create() {
	e1 := Employee{"Jane", "Smith", 32, "I like to collect rock albums", []string{"music"}}
	put1, err := client.Index().
		Index("megacorp").
		Type("employee").
		Id("1").
		BodyJson(e1).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index s%s, type %s\n", put1.Id, put1.Index, put1.Type)

	//使用字符串
	e2 := `{"first_name":"John","last_name":"Smith","age":25,"about":"I love to go rock climbing","interests":["sports","music"]}`
	put2, err := client.Index().
		Index("megacorp").
		Type("employee").
		Id("2").
		BodyJson(e2).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index s%s, type %s\n", put2.Id, put2.Index, put2.Type)

	e3 := `{"first_name":"Douglas","last_name":"Fir","age":35,"about":"I like to build cabinets","interests":["forestry"]}`
	put3, err := client.Index().
		Index("megacorp").
		Type("employee").
		Id("3").
		BodyJson(e3).
		Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index s%s, type %s\n", put3.Id, put3.Index, put3.Type)
}

func delete() {
	res, err := client.Delete().Index("megacorp").
		Type("employee").
		Id("1").
		Do(context.Background())
	if err != nil {
		println(err)
		return
	}
	fmt.Printf("delete result %s\n", res.Result)

	gets("1")
	gets("2")
	gets("3")
}

func update() {
	res, err := client.Update().
		Index("megacorp").
		Type("employee").
		Id("2").
		Doc(map[string]interface{}{"age": 88}).
		Do(context.Background())
	if err != nil {
		println(err.Error())
	}

	fmt.Printf("update age %s\n", res.Result)
	gets("2")
}

func gets(id string) {
	doc, err := client.Get().Index("megacorp").Type("employee").Id(id).Do(context.Background())
	if err != nil {
		fmt.Println("id:", id, err)
		return
	}

	if doc.Found {
		fmt.Printf("Got document %s in version %d from index %s, type %s\n", doc.Id, doc.Version, doc.Index, doc.Type)
	}
}

func query() {
	var res *elastic.SearchResult
	var err error
	//取所有
	res, err = client.Search("megacorp").Type("employee").Do(context.Background())
	printEmployee(res, err)

	//字段相等
	q := elastic.NewQueryStringQuery("last_name:Smith")
	res, err = client.Search("megacorp").Type("employee").Query(q).Do(context.Background())
	if err != nil {
		println(err.Error())
	}
	printEmployee(res, err)

	/*if res.Hits.TotalHits.Value > 0 {
		fmt.Printf("Found a total of %d Employee \n", res.Hits.TotalHits)

		for _, hit := range res.Hits.Hits {
			var t Employee
			err := json.Unmarshal(*hit.Source, &t) //另外一种取数据的方法
			if err != nil {
				fmt.Println("Deserialization failed")
			}
			fmt.Printf("Employee name %s : %s\n", t.FirstName, t.LastName)
		}
	} else {
		fmt.Printf("Found no Employee \n")
	}*/

	//条件查询
	//年龄大于30岁的
	boolQ := elastic.NewBoolQuery()
	boolQ.Must(elastic.NewMatchQuery("last_name", "smith"))
	boolQ.Filter(elastic.NewRangeQuery("age").Gt(30))
	res, err = client.Search("megacorp").Type("employee").Query(q).Do(context.Background())
	printEmployee(res, err)

	//短语搜索 搜索about字段中有 rock climbing
	matchPhraseQuery := elastic.NewMatchPhraseQuery("about", "rock climbing")
	res, err = client.Search("megacorp").Type("employee").Query(matchPhraseQuery).Do(context.Background())
	printEmployee(res, err)

	//分析 interests
	aggs := elastic.NewTermsAggregation().Field("interests")
	res, err = client.Search("megacorp").Type("employee").Aggregation("all_interests", aggs).Do(context.Background())
	printEmployee(res, err)

}

func list(size, page int) {
	if size < 0 || page < 1 {
		fmt.Println("param error")
		return
	}

	res, err := client.Search("megacorp").
		Type("employee").
		Size(size).
		From((page - 1) * size).
		Do(context.Background())
	printEmployee(res, err)
}

func printEmployee(res *elastic.SearchResult, err error) {
	if err != nil {
		println(err.Error())
		return
	}
	var typ Employee
	for _, item := range res.Each(reflect.TypeOf(typ)) { //从搜索结果中取数据的方法
		t := item.(Employee)
		fmt.Printf("%#v\n", t)
	}
}

func main() {
	create()
	delete()
	update()
	query()
	list(10, 1)
}
