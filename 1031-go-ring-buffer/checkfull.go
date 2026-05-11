package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"ringbuffer/common"
	"ringbuffer/ringbuffer"
)

var pass, fail int

func check(name string, ok bool) {
	if ok {
		fmt.Printf("  PASS: %s\n", name)
		pass++
	} else {
		fmt.Printf("  FAIL: %s\n", name)
		fail++
	}
}

func main() {
	fmt.Println("=== Core Tests ===")
	testCore()
	fmt.Printf("\nCore: %d/%d\n", pass, fail)

	p1, f1 := pass, fail
	pass, fail = 0, 0
	fmt.Println("\n=== HTTP Tests ===")
	testHTTP()
	fmt.Printf("\nHTTP: %d/%d\n", pass, fail)

	fmt.Printf("\nTOTAL: %d passed, %d failed\n", p1+pass, f1+fail)
	if f1+fail > 0 {
		os.Exit(1)
	}
}

func testCore() {
	rb := ringbuffer.New(10)
	check("New cap=10", rb.Capacity() == 10)
	check("Size=0", rb.Size() == 0)

	n, err := rb.Write([]byte("Hello"), -1)
	check("Write 5", n == 5 && err == nil)
	check("Size=5", rb.Size() == 5)

	buf := make([]byte, 5)
	n, err = rb.Read(buf, -1)
	check("Read 5", n == 5 && err == nil)
	check("Data=Hello", string(buf[:n]) == "Hello")
	check("Size=0", rb.Size() == 0)

	rb2 := ringbuffer.New(3)
	rb2.Write([]byte("abc"), -1)
	n, err = rb2.Write([]byte("d"), 0)
	check("Full write t=0 err", n == 0 && err != nil)

	rb3 := ringbuffer.New(10)
	n, err = rb3.Read(make([]byte, 1), 0)
	check("Empty read t=0 err", n == 0 && err != nil)

	rb4 := ringbuffer.New(10)
	rb4.Close()
	n, err = rb4.Write([]byte("x"), -1)
	check("Write closed err", n == 0 && err != nil)

	rb5 := ringbuffer.New(10)
	rb5.Close()
	n, err = rb5.Read(make([]byte, 1), -1)
	check("Read closed=EOF", n == 0 && err == io.EOF)

	rb6 := ringbuffer.New(10)
	rb6.Write([]byte("residual"), -1)
	rb6.Close()
	buf6 := make([]byte, 100)
	n, err = rb6.Read(buf6, -1)
	check("Residual read", n == 8 && err == nil && string(buf6[:n]) == "residual")
	n, err = rb6.Read(buf6, -1)
	check("After residual=EOF", n == 0 && err == io.EOF)

	for i, data := range [][]byte{[]byte("x"), []byte("repeat-repeat-repeat!!")} {
		rbRT := ringbuffer.New(100)
		wn, werr := rbRT.Write(data, -1)
		check(fmt.Sprintf("RT%d write", i), wn == len(data) && werr == nil)
		bufRT := make([]byte, len(data))
		rn, rerr := rbRT.Read(bufRT, -1)
		check(fmt.Sprintf("RT%d read match", i), rn == len(data) && rerr == nil && bytes.Equal(bufRT[:rn], data))
	}

	rb9 := ringbuffer.New(5)
	rb9.Write([]byte("abcde"), -1)
	buf9 := make([]byte, 3)
	n, _ = rb9.Read(buf9, -1)
	check("Wrap r1", string(buf9[:n]) == "abc")
	rb9.Write([]byte("fgh"), -1)
	buf9 = make([]byte, 5)
	n, _ = rb9.Read(buf9, -1)
	check("Wrap r2", string(buf9[:n]) == "defgh")

	rb10 := ringbuffer.New(0)
	check("ZeroCap New", rb10.Capacity() == 0)

	rb11 := ringbuffer.New(0)
	ch := make(chan string)
	go func() {
		wn, werr := rb11.Write([]byte("zero"), -1)
		ch <- fmt.Sprintf("w=%d e=%v", wn, werr)
	}()
	time.Sleep(100 * time.Millisecond)
	buf11 := make([]byte, 10)
	rn, rerr := rb11.Read(buf11, -1)
	check("ZeroCap read", rn == 4 && rerr == nil && string(buf11[:rn]) == "zero")
	r := <-ch
	check("ZeroCap write done", r == "w=4 e=<nil>")

	rb12 := ringbuffer.New(0)
	n, err = rb12.Write([]byte("x"), 0)
	check("ZeroCap write t=0", n == 0 && err != nil)

	rb13 := ringbuffer.New(100)
	ch13 := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			rb13.Write([]byte("ab"), -1)
		}
		close(ch13)
	}()
	total := 0
	buf13 := make([]byte, 2)
	for i := 0; i < 100; i++ {
		n, _ = rb13.Read(buf13, -1)
		total += n
	}
	<-ch13
	check("Concurrent 200B", total == 200)

	rb14 := ringbuffer.New(1000)
	big := bytes.Repeat([]byte("ABCDEFGHIJ"), 100)
	n, err = rb14.Write(big, -1)
	check("Write 1000B", n == 1000 && err == nil)
	buf14 := make([]byte, 1000)
	n, err = rb14.Read(buf14, -1)
	check("Read 1000B", n == 1000 && err == nil && bytes.Equal(buf14[:n], big))

	rb15 := ringbuffer.New(1)
	rb15.Write([]byte("a"), -1)
	start := time.Now()
	rb15.Write([]byte("b"), 100*time.Millisecond)
	el := time.Since(start)
	check("Write timeout ~100ms", el >= 80*time.Millisecond && el < 500*time.Millisecond)

	rb16 := ringbuffer.New(10)
	start = time.Now()
	rb16.Read(make([]byte, 1), 100*time.Millisecond)
	el = time.Since(start)
	check("Read timeout ~100ms", el >= 80*time.Millisecond && el < 500*time.Millisecond)
}

func testHTTP() {
	mux := http.NewServeMux()
	bufs := map[string]*ringbuffer.RingBuffer{}
	nextID := 1

	mux.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		var req common.CreateRequest
		json.NewDecoder(r.Body).Decode(&req)
		id := fmt.Sprintf("t-%d", nextID); nextID++
		bufs[id] = ringbuffer.New(req.Capacity)
		json.NewEncoder(w).Encode(common.CreateResponse{ID: id, Capacity: req.Capacity})
	})
	mux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		var req common.WriteRequest
		json.NewDecoder(r.Body).Decode(&req)
		rb, ok := bufs[req.ID]
		if !ok { json.NewEncoder(w).Encode(common.WriteResponse{Error: "not found"}); return }
		var t time.Duration
		if req.Timeout > 0 { t = time.Duration(req.Timeout) * time.Millisecond } else if req.Timeout == 0 { t = 0 } else { t = -1 }
		wn, werr := rb.Write(req.Data, t)
		resp := common.WriteResponse{Written: wn}
		if werr != nil { resp.Error = werr.Error() }
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		var req common.ReadRequest
		json.NewDecoder(r.Body).Decode(&req)
		rb, ok := bufs[req.ID]
		if !ok { json.NewEncoder(w).Encode(common.ReadResponse{Error: "not found"}); return }
		var t time.Duration
		if req.Timeout > 0 { t = time.Duration(req.Timeout) * time.Millisecond } else if req.Timeout == 0 { t = 0 } else { t = -1 }
		data := make([]byte, req.Length)
		rn, rerr := rb.Read(data, t)
		resp := common.ReadResponse{Data: data[:rn], Read: rn}
		if rerr != nil {
			if rerr == io.EOF { resp.EOF = true } else { resp.Error = rerr.Error() }
		}
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		var req common.StatusRequest
		json.NewDecoder(r.Body).Decode(&req)
		rb, ok := bufs[req.ID]
		if !ok { json.NewEncoder(w).Encode(common.StatusResponse{Error: "not found"}); return }
		json.NewEncoder(w).Encode(common.StatusResponse{ID: req.ID, Capacity: rb.Capacity(), Size: rb.Size()})
	})
	mux.HandleFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		var req common.CloseRequest
		json.NewDecoder(r.Body).Decode(&req)
		rb, ok := bufs[req.ID]
		if !ok { json.NewEncoder(w).Encode(common.CloseResponse{Error: "not found"}); return }
		rb.Close()
		json.NewEncoder(w).Encode(common.CloseResponse{})
	})
	mux.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		list := make([]common.StatusResponse, 0)
		for id, rb := range bufs { list = append(list, common.StatusResponse{ID: id, Capacity: rb.Capacity(), Size: rb.Size()}) }
		json.NewEncoder(w).Encode(common.ListResponse{Buffers: list})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()
	c := ts.Client()

	post := func(path, body string) []byte {
		resp, _ := c.Post(ts.URL+path, "application/json", bytes.NewReader([]byte(body)))
		defer resp.Body.Close()
		d, _ := io.ReadAll(resp.Body)
		return d
	}

	var cr common.CreateResponse
	json.Unmarshal(post("/create", `{"capacity":10}`), &cr)
	check("H create", cr.ID != "" && cr.Capacity == 10)

	var sr common.StatusResponse
	json.Unmarshal(post("/status", fmt.Sprintf(`{"id":"%s"}`, cr.ID)), &sr)
	check("H size=0", sr.Size == 0)

	var wr common.WriteResponse
	json.Unmarshal(post("/write", fmt.Sprintf(`{"id":"%s","data":[72,101,108,108,111],"timeout":-1}`, cr.ID)), &wr)
	check("H write 5", wr.Written == 5 && wr.Error == "")

	json.Unmarshal(post("/status", fmt.Sprintf(`{"id":"%s"}`, cr.ID)), &sr)
	check("H size=5", sr.Size == 5)

	var rr common.ReadResponse
	json.Unmarshal(post("/read", fmt.Sprintf(`{"id":"%s","length":5,"timeout":-1}`, cr.ID)), &rr)
	check("H read Hello", rr.Read == 5 && string(rr.Data) == "Hello")

	var clr common.CloseResponse
	json.Unmarshal(post("/close", fmt.Sprintf(`{"id":"%s"}`, cr.ID)), &clr)
	check("H close", clr.Error == "")

	json.Unmarshal(post("/read", fmt.Sprintf(`{"id":"%s","length":1,"timeout":0}`, cr.ID)), &rr)
	check("H read closed=EOF", rr.EOF && rr.Read == 0)

	json.Unmarshal(post("/write", fmt.Sprintf(`{"id":"%s","data":[88],"timeout":0}`, cr.ID)), &wr)
	check("H write closed err", wr.Error != "")

	json.Unmarshal(post("/create", `{"capacity":10}`), &cr)
	post("/write", fmt.Sprintf(`{"id":"%s","data":[114,101,115],"timeout":-1}`, cr.ID))
	post("/close", fmt.Sprintf(`{"id":"%s"}`, cr.ID))
	json.Unmarshal(post("/read", fmt.Sprintf(`{"id":"%s","length":10,"timeout":-1}`, cr.ID)), &rr)
	check("H residual", rr.Read == 3 && string(rr.Data) == "res")

	var lr common.ListResponse
	json.Unmarshal(post("/list", `{}`), &lr)
	check("H list", len(lr.Buffers) >= 1)

	json.Unmarshal(post("/create", `{"capacity":0}`), &cr)
	check("H zero-cap", cr.Capacity == 0)

	json.Unmarshal(post("/write", fmt.Sprintf(`{"id":"%s","data":[88],"timeout":100}`, cr.ID)), &wr)
	check("H zero-cap timeout", wr.Error != "")
}
