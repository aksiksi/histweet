package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	histweet "github.com/aksiksi/histweet/lib"
)

func ruleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req := struct {
		Rule     string `json:"rule"`
		IsDaemon bool   `json:"is_daemon"`
		Interval int    `json:"interval"`
	}{}

	resp := struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}{
		Success: true,
	}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&req)
	if err != nil {
		resp.Success = false
		resp.Msg = fmt.Sprintf("Invalid JSON request body: %s", err)
		resp, _ := json.Marshal(&resp)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)

		return
	}

	// Parse the Rule
	parser := histweet.NewParser(req.Rule)
	rule, err := parser.Parse()
	if err != nil {
		resp.Success = false
		resp.Msg = fmt.Sprintf("Invalid rule string provided: %s", err)
		resp, _ := json.Marshal(&resp)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)

		return
	}

	fmt.Println(rule)

	bytes := []byte(rule.ToString())
	w.Write(bytes)
}

func main() {
	http.HandleFunc("/rule", ruleHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
