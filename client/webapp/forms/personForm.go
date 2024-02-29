package forms

import (
	"benefitsDomain/apiResponse"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

var PersonIdTemplate *template.Template
var PersonHomeTemplate *template.Template

func InitializePerson() {
	PersonIdTemplate = template.Must(template.ParseFiles("./public/personIdForm.html"))
	PersonHomeTemplate = template.Must(template.ParseFiles("./public/personHomePage.html"))
	http.HandleFunc("/person", handlePersonForm)
}
func handlePersonForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		err0 := PersonIdTemplate.Execute(w, nil)
		if err0 != nil {
			log.Fatal(err0)
		}
		return
	}
	personId := r.FormValue("personId")
	fmt.Println(personId)

	url := "http://localhost:8080/api/persons/" + personId + "/view/profile"
	response, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Headers:", response.Header)
	body, _ := io.ReadAll(response.Body)
	fmt.Println("response Body:", string(body))
	ppv := apiResponse.PersonProfileViewResponse{}
	if err := json.Unmarshal(body, &ppv); err != nil {
		log.Fatal(err)
	}
	/*_, err3 := w.Write(body)
	if err3 != nil {
		log.Fatal(err3)
	}
	*/
	err0 := PersonHomeTemplate.Execute(w, ppv)
	if err0 != nil {
		log.Fatal(err0)
	}

}
