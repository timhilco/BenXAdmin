// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type MyDate time.Time

type MyData struct {
	FieldS string
	FieldI int64
	FieldD *MyDate
}

func (v MyDate) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(time.Time(v))
}

func (v *MyDate) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	rv := bson.RawValue{
		Type:  t,
		Value: b,
	}

	var res time.Time
	if err := rv.Unmarshal(&res); err != nil {
		return err
	}
	*v = MyDate(res)

	return nil
}

func (v MyDate) String() string {
	return time.Time(v).String()
}

func main() {
	// Marshal as BSON.
	d := MyDate(time.Now())
	fmt.Println("Input MyDate:", d)
	md := MyData{
		FieldS: "my string",
		FieldI: 12345,
		FieldD: &d,
	}

	b, err := bson.Marshal(md)
	if err != nil {
		panic(err)
	}
	fmt.Println("As Extended JSON:", bson.Raw(b))

	// Unmarshal from BSON.
	var res MyData
	err = bson.Unmarshal(b, &res)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Unmarshalled: %+v\n", res)
	fmt.Println("Output MyDate:", res.FieldD)
}
