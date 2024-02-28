package datatypes

import (
	"encoding/json"
	"fmt"
	"math/big"

	carbon "github.com/golang-module/carbon/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type YYYYMMDD_Date string

func (y YYYYMMDD_Date) String() string {
	return string(y)
}
func (y YYYYMMDD_Date) FormattedString(pattern string) string {
	yyyy := y[0:4]
	mm := y[4:6]
	dd := y[6:8]
	r := fmt.Sprintf("%s/%s/%s", mm, dd, yyyy)
	return r
}
func (y YYYYMMDD_Date) Equal(aYYYYMMDDString string) bool {
	return string(y) == aYYYYMMDDString
}
func YYYYMMDD_Date_Now() YYYYMMDD_Date {

	s := carbon.Now().String()
	return YYYYMMDD_Date(s)
}

type BigFloat struct {
	aBigMathFloat *big.Float
}

func NewBigFloat(s string) (BigFloat, error) {
	f := new(big.Float)
	f, ok := f.SetString(s)
	if !ok {
		fmt.Println("bad Value")
	}
	return BigFloat{
		aBigMathFloat: f,
	}, nil

}
func (f BigFloat) String() string {
	x := f.aBigMathFloat
	//s := fmt.Sprintf("x = %.10g (%s, prec = %d, acc = %s)\n", x, x.Text('p', 0), x.Prec(), x.Acc())
	if x == nil {
		return "-1.0"
	}

	s := fmt.Sprintf("%.10g", x)
	return s
}
func (f BigFloat) DebugString() string {
	x := f.aBigMathFloat
	s := fmt.Sprintf("x = %.10g (%s, prec = %d, acc = %s)\n", x, x.Text('p', 0), x.Prec(), x.Acc())
	return s
}

const twoDecimal = 32

func (f *BigFloat) Add(x, y BigFloat) BigFloat {

	var c big.Float
	a := x.aBigMathFloat
	b := y.aBigMathFloat
	c.SetPrec(twoDecimal)
	c.Add(a, b)
	z := BigFloat{
		aBigMathFloat: &c,
	}
	return z

}
func (f *BigFloat) Sub(x, y BigFloat) BigFloat {

	var c big.Float
	a := x.aBigMathFloat
	b := y.aBigMathFloat
	c.SetPrec(twoDecimal)
	c.Sub(a, b)
	z := BigFloat{
		aBigMathFloat: &c,
	}
	return z

}
func (f *BigFloat) Mul(x, y BigFloat) BigFloat {

	var c big.Float
	a := x.aBigMathFloat
	b := y.aBigMathFloat
	c.SetPrec(twoDecimal)
	c.Mul(a, b)
	z := BigFloat{
		aBigMathFloat: &c,
	}
	return z

}
func (f *BigFloat) Quo(x, y BigFloat) BigFloat {

	var c big.Float
	a := x.aBigMathFloat
	b := y.aBigMathFloat
	c.SetPrec(twoDecimal)
	c.Quo(a, b)
	z := BigFloat{
		aBigMathFloat: &c,
	}
	return z

}
func (f BigFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.aBigMathFloat)
}

func (f *BigFloat) UnmarshalJSON(data []byte) error {
	var b big.Float
	err := json.Unmarshal(data, &b)
	if err != nil {
		return err
	}
	*f = BigFloat{
		aBigMathFloat: &b,
	}
	return nil
}

func (f BigFloat) MarshalBSONValue() (bsontype.Type, []byte, error) {
	value := f.String()
	return bson.MarshalValue(value)
}

func (f *BigFloat) UnmarshalBSONValue(t bsontype.Type, value []byte) error {
	if t != bson.TypeString {
		return fmt.Errorf("invalid bson value type '%s'", t.String())
	}
	s, _, ok := bsoncore.ReadString(value)
	if !ok {
		return fmt.Errorf("invalid bson string value")
	}
	bf := new(big.Float)
	bf, ok = bf.SetString(s)
	if !ok {
		fmt.Printf("UnmarshalBSONValue: Bad Value: %s", value)
		bf, _ = bf.SetString("-1.0")
	}
	f.aBigMathFloat = bf
	return nil
}
func GetGlobalInternalIdentifier() string {
	uuid, _ := uuid.NewRandom()
	return uuid.String()

}

type EnvironmentVariables struct {
	TemplateDirectory string
	Cors              bool
	IsKafka           bool
}
