package datatypes

import (
	"encoding/json"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func Test_BigFloat_Add(t *testing.T) {
	x, _ := NewBigFloat("12.34")
	y, _ := NewBigFloat("2.345")
	a := BigFloat{}
	b := a.Add(x, y)
	s := b.String()
	if s != "14.685" {
		t.Errorf("Result was incorrect, got: %s, want: %s.", s, "14.685")
	}

}
func Test_BigFloat_Sub(t *testing.T) {
	x, _ := NewBigFloat("12.34")
	y, _ := NewBigFloat("2.345")
	a := BigFloat{}
	b := a.Sub(x, y)
	s := b.String()
	if s != "9.995" {
		t.Errorf("Result was incorrect, got: %s, want: %s.", s, "9.995")
	}

}

type Worker struct {
	Name   string
	Salary BigFloat
}

func Test_BigFloat_BSON(t *testing.T) {
	x, _ := NewBigFloat("120000.34")
	p := Worker{
		Name:   "John Doe",
		Salary: x,
	}

	jsonData, err := bson.Marshal(p)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(jsonData))
	var decodedWorker Worker
	err = bson.Unmarshal(jsonData, &decodedWorker)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("%+v\n", decodedWorker)

}
func Test_BigFloat_JSON(t *testing.T) {
	x, _ := NewBigFloat("120000.34")
	p := Worker{
		Name:   "John Doe",
		Salary: x,
	}

	jsonData, err := json.Marshal(p)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(jsonData))
	var decodedWorker Worker
	err = json.Unmarshal(jsonData, &decodedWorker)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("%+v\n", decodedWorker)

}
